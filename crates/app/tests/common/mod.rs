use app::server::{build_router, state_from_env, AppState};
use axum::Router;
use meilisearch_sdk::client::Client as MeiliClient;
use sqlx::{mysql::MySqlPoolOptions, MySql, Transaction};

pub async fn test_router() -> anyhow::Result<Router> {
  let state = if use_real_dependencies() {
    state_from_env().await?
  } else {
    offline_state()?
  };

  Ok(build_router(state))
}

#[allow(dead_code)]
pub async fn test_state() -> anyhow::Result<AppState> {
  if use_real_dependencies() {
    state_from_env().await
  } else {
    offline_state()
  }
}

pub fn use_real_dependencies() -> bool {
  matches!(
    std::env::var("RUST_TEST_REAL_DEPS").as_deref(),
    Ok("1") | Ok("true") | Ok("TRUE") | Ok("yes") | Ok("YES")
  )
}

#[allow(dead_code)]
pub fn allow_write_tests() -> bool {
  matches!(
    std::env::var("RUST_TEST_ALLOW_WRITE").as_deref(),
    Ok("1") | Ok("true") | Ok("TRUE") | Ok("yes") | Ok("YES")
  )
}

#[allow(dead_code)]
pub async fn begin_write_transaction(
  state: &AppState,
) -> anyhow::Result<Transaction<'_, MySql>> {
  if !allow_write_tests() {
    anyhow::bail!("write tests are disabled, set RUST_TEST_ALLOW_WRITE=1 to enable");
  }

  state.pool().begin().await.map_err(|e| anyhow::anyhow!(e))
}

#[allow(dead_code)]
pub fn env_var(name: &str) -> Option<String> {
  std::env::var(name).ok().filter(|v| !v.trim().is_empty())
}

fn offline_state() -> anyhow::Result<AppState> {
  let meili = MeiliClient::new("http://127.0.0.1:17700", Some("test-key".to_string()))
    .map_err(|e| anyhow::anyhow!(e))?;

  let pool = MySqlPoolOptions::new()
    .max_connections(1)
    .connect_lazy("mysql://root:root@127.0.0.1:13306/test")
    .map_err(|e| anyhow::anyhow!(e))?;

  Ok(AppState::new(meili, pool))
}
