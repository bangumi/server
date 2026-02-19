use anyhow::Context;
use common::config::AppConfig;
use opendal::services;
use opendal::Operator;
use serde::Deserialize;
use serde_json::Value;
use std::collections::HashSet;

use super::types::OP_UPDATE;

#[derive(Debug, Deserialize)]
struct UserKey {
  uid: u32,
}

pub struct UserDispatcher {
  pool: sqlx::MySqlPool,
  redis: redis::Client,
  s3: Option<Operator>,
}

#[derive(Debug, serde::Serialize)]
struct RedisUserChannel {
  user_id: u32,
  new_notify: u16,
}

impl UserDispatcher {
  pub fn new(cfg: &AppConfig, pool: sqlx::MySqlPool) -> anyhow::Result<Self> {
    let redis = redis::Client::open(cfg.redis_url.as_str()).context("create redis client")?;
    let s3 = build_s3_operator(cfg)?;
    Ok(Self { pool, redis, s3 })
  }

  pub async fn on_user(
    &self,
    key: &[u8],
    op: &str,
    before: Option<Value>,
    after: Option<Value>,
  ) -> anyhow::Result<()> {
    let key: UserKey = serde_json::from_slice(key).context("parse user key")?;
    self.on_user_change(key.uid, op, before, after).await
  }

  async fn on_user_change(
    &self,
    user_id: u32,
    op: &str,
    before: Option<Value>,
    after: Option<Value>,
  ) -> anyhow::Result<()> {
    if op != OP_UPDATE {
      return Ok(());
    }

    let (Some(before), Some(after)) = (before, after) else {
      return Ok(());
    };

    let old_password = before.get("password_crypt").and_then(Value::as_str);
    let new_password = after.get("password_crypt").and_then(Value::as_str);

    if old_password != new_password {
      self.revoke_user_sessions(user_id).await?;
      tracing::info!(user_id, "password changed, sessions revoked");
    }

    let old_notify = before.get("new_notify").and_then(Value::as_u64);
    let new_notify = after.get("new_notify").and_then(Value::as_u64);
    if old_notify != new_notify {
      let notify = new_notify.unwrap_or_default() as u16;
      self.publish_notify_change(user_id, notify).await?;
      tracing::info!(user_id, new_notify = notify, "new notify changed, published redis event");
    }

    let old_avatar = before.get("avatar").and_then(Value::as_str);
    let new_avatar = after.get("avatar").and_then(Value::as_str);
    if old_avatar != new_avatar {
      if let (Some(s3), Some(avatar)) = (&self.s3, new_avatar) {
        tracing::debug!(user_id, avatar, "avatar changed, clear image cache in background");
        let s3 = s3.clone();
        let avatar = avatar.to_owned();
        tokio::spawn(async move {
          if let Err(err) = clear_image_cache(s3, avatar).await {
            tracing::error!(error = ?err, error_chain = %format!("{err:#}"), "failed to clear s3 cached image");
          }
        });
      }
    }

    Ok(())
  }

  async fn revoke_user_sessions(&self, user_id: u32) -> anyhow::Result<()> {
    let now = std::time::SystemTime::now()
      .duration_since(std::time::UNIX_EPOCH)
      .context("system time before unix epoch")?
      .as_secs() as i64;

    let rows = sqlx::query_as::<_, (String,)>(
      r#"SELECT `key` FROM chii_os_web_sessions WHERE user_id = ?"#,
    )
    .bind(user_id)
    .fetch_all(&self.pool)
    .await
    .context("load user sessions")?;

    sqlx::query(
      r#"UPDATE chii_os_web_sessions SET expired_at = ? WHERE user_id = ?"#,
    )
    .bind(now)
    .bind(user_id)
    .execute(&self.pool)
    .await
    .context("revoke user sessions in mysql")?;

    if rows.is_empty() {
      return Ok(());
    }

    let mut conn = self
      .redis
      .get_multiplexed_async_connection()
      .await
      .context("get redis connection")?;

    let mut cmd = redis::cmd("DEL");
    for (key,) in rows {
      cmd.arg(format!("chii:web:session:{key}"));
    }
    let _: i64 = cmd.query_async(&mut conn).await.context("delete redis sessions")?;

    Ok(())
  }

  async fn publish_notify_change(&self, user_id: u32, new_notify: u16) -> anyhow::Result<()> {
    let message = serde_json::to_string(&RedisUserChannel {
      user_id,
      new_notify,
    })
    .context("encode redis user notify message")?;

    let mut conn = self
      .redis
      .get_multiplexed_async_connection()
      .await
      .context("get redis connection")?;

    let channel = format!("event-user-notify-{user_id}");
    let _: i64 = redis::cmd("PUBLISH")
      .arg(channel)
      .arg(message)
      .query_async(&mut conn)
      .await
      .context("publish user notify event")?;

    Ok(())
  }
}

fn build_s3_operator(cfg: &AppConfig) -> anyhow::Result<Option<Operator>> {
  if cfg.s3_entry_point.is_empty() || cfg.s3_access_key.is_empty() || cfg.s3_secret_key.is_empty()
  {
    return Ok(None);
  }

  let builder = services::S3::default()
    .root("/")
    .endpoint(&cfg.s3_entry_point)
    .access_key_id(&cfg.s3_access_key)
    .secret_access_key(&cfg.s3_secret_key)
    .bucket(&cfg.s3_image_resize_bucket);

  let builder = if let Some(region) = &cfg.s3_region {
    builder.region(region)
  } else {
    builder
  };

  let operator = Operator::new(builder)
    .context("create s3 operator")?
    .finish();

  Ok(Some(operator))
}

async fn clear_image_cache(s3: Operator, avatar: String) -> anyhow::Result<()> {
  let (path, query) = avatar
    .split_once('?')
    .map_or((avatar.as_str(), ""), |(path, query)| (path, query));

  let mut prefix = format!("/pic/user/l/{path}");
  if query.contains("hd=1") {
    prefix = format!("/hd{prefix}");
  }

  tracing::debug!(avatar, prefix, "clear image cache by prefix");

  let mut keys: HashSet<String> = HashSet::new();

  for candidate_prefix in prefix_candidates(&prefix) {
    let dir = prefix_dirname(&candidate_prefix);

    let entries = match tokio::time::timeout(
      std::time::Duration::from_secs(10),
      s3.list(&dir),
    )
    .await
    {
      Ok(Ok(entries)) => entries,
      Ok(Err(err)) => {
        tracing::warn!(
          prefix = candidate_prefix,
          dir,
          error = ?err,
          "failed to list cached avatar objects by dirname"
        );
        continue;
      }
      Err(_) => {
        tracing::warn!(
          prefix = candidate_prefix,
          dir,
          "timeout while listing cached avatar objects by dirname"
        );
        continue;
      }
    };

    for entry in entries {
      let path = entry.path();
      if path.starts_with(candidate_prefix.as_str()) {
        keys.insert(path.to_string());
      }
    }

    if !keys.is_empty() {
      break;
    }
  }

  for key in keys {
    if let Err(err) = s3.delete(&key).await {
      tracing::error!(
        key,
        error = ?err,
        "failed to delete cached avatar object"
      );
    }
  }

  Ok(())
}

fn prefix_candidates(prefix: &str) -> Vec<String> {
  if let Some(stripped) = prefix.strip_prefix('/') {
    vec![prefix.to_owned(), stripped.to_owned()]
  } else {
    vec![prefix.to_owned()]
  }
}

fn prefix_dirname(prefix: &str) -> String {
  match prefix.rsplit_once('/') {
    Some((dir, _)) if !dir.is_empty() => format!("{dir}/"),
    Some(_) => "/".to_string(),
    None => "/".to_string(),
  }
}

#[cfg(test)]
mod tests {
  use super::{prefix_candidates, prefix_dirname};

  #[test]
  fn test_prefix_dirname() {
    assert_eq!(prefix_dirname("/pic/user/l/a.jpg"), "/pic/user/l/");
    assert_eq!(prefix_dirname("pic/user/l/a.jpg"), "pic/user/l/");
    assert_eq!(prefix_dirname("/a.jpg"), "/");
    assert_eq!(prefix_dirname("a.jpg"), "/");
  }

  #[test]
  fn test_prefix_candidates() {
    assert_eq!(
      prefix_candidates("/pic/user/l/a.jpg"),
      vec!["/pic/user/l/a.jpg".to_string(), "pic/user/l/a.jpg".to_string()]
    );
    assert_eq!(prefix_candidates("pic/user/l/a.jpg"), vec!["pic/user/l/a.jpg".to_string()]);
  }

}
