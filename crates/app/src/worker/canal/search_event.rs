use anyhow::{anyhow, Context};
use bangumi_wiki_parser::{parse_omit_error, FieldValue, Wiki};
use common::config::AppConfig;
use meilisearch_sdk::client::Client as MeiliSdkClient;
use php_serialize::from_str as parse_php_serialize;
use serde::{Deserialize, Serialize};

use super::types::{OP_CREATE, OP_DELETE, OP_SNAPSHOT, OP_UPDATE};

pub const TARGET_SUBJECT: &str = "subject";
pub const TARGET_CHARACTER: &str = "character";
pub const TARGET_PERSON: &str = "person";

pub struct SearchDispatcher {
  pool: sqlx::MySqlPool,
  meili: Option<MeiliClient>,
}

struct MeiliClient {
  client: MeiliSdkClient,
}

#[derive(Serialize)]
struct SubjectDoc {
  id: u32,
  tag: Vec<String>,
  #[serde(rename = "meta_tag")]
  meta_tag: Vec<String>,
  name: String,
  aliases: Vec<String>,
  date: i32,
  score: f64,
  rating_count: u32,
  page_rank: f64,
  heat: u32,
  rank: u32,
  platform: u16,
  #[serde(rename = "type")]
  type_id: u8,
  nsfw: bool,
}

#[derive(Serialize)]
struct CharacterDoc {
  id: u32,
  name: String,
  aliases: Vec<String>,
  comment: u32,
  collect: u32,
  nsfw: bool,
}

#[derive(Serialize)]
struct PersonDoc {
  id: u32,
  name: String,
  aliases: Vec<String>,
  comment: u32,
  collect: u32,
  career: Vec<String>,
}

#[derive(Deserialize)]
struct SubjectTagItem {
  tag_name: Option<String>,
}

#[derive(sqlx::FromRow)]
struct SubjectRow {
  subject_id: u32,
  subject_name: String,
  subject_name_cn: String,
  field_infobox: String,
  subject_type_id: u8,
  subject_nsfw: bool,
  subject_ban: u8,
  subject_platform: u16,
  field_meta_tags: String,
  field_tags: String,
  subject_wish: u32,
  subject_collect: u32,
  subject_doing: u32,
  subject_on_hold: u32,
  subject_dropped: u32,
  field_rank: u32,
  date: String,
  field_rate_1: u32,
  field_rate_2: u32,
  field_rate_3: u32,
  field_rate_4: u32,
  field_rate_5: u32,
  field_rate_6: u32,
  field_rate_7: u32,
  field_rate_8: u32,
  field_rate_9: u32,
  field_rate_10: u32,
  field_redirect: u32,
}

#[derive(sqlx::FromRow)]
struct CharacterRow {
  crt_id: u32,
  crt_name: String,
  crt_infobox: String,
  crt_comment: u32,
  crt_collects: u32,
  crt_nsfw: bool,
  crt_redirect: u32,
}

#[derive(sqlx::FromRow)]
struct PersonRow {
  prsn_id: u32,
  prsn_name: String,
  prsn_infobox: String,
  prsn_comment: u32,
  prsn_collects: u32,
  prsn_redirect: u32,
  prsn_producer: bool,
  prsn_mangaka: bool,
  prsn_artist: bool,
  prsn_seiyu: bool,
  prsn_writer: bool,
  prsn_illustrator: bool,
  prsn_actor: bool,
}

impl SearchDispatcher {
  pub fn new(cfg: &AppConfig, pool: sqlx::MySqlPool) -> anyhow::Result<Self> {
    let meili = if cfg.meilisearch_url.is_empty() {
      None
    } else {
      let api_key = if cfg.meilisearch_key.is_empty() {
        None
      } else {
        Some(cfg.meilisearch_key.clone())
      };
      Some(MeiliClient {
        client: MeiliSdkClient::new(cfg.meilisearch_url.trim_end_matches('/'), api_key)
          .context("create meilisearch client")?,
      })
    };

    Ok(Self { pool, meili })
  }

  pub async fn dispatch(&self, target: &str, entity_id: u32, op: &str) -> anyhow::Result<()> {
    match op {
      OP_CREATE | OP_UPDATE | OP_SNAPSHOT => self.upsert(target, entity_id, op).await,
      OP_DELETE => self.delete(target, entity_id, op).await,
      _ => Err(anyhow!("unexpected operator: {op}")),
    }
  }

  async fn upsert(&self, target: &str, entity_id: u32, op: &str) -> anyhow::Result<()> {
    let Some(meili) = &self.meili else {
      tracing::debug!(target, entity_id, op, "skip search upsert: meilisearch disabled");
      return Ok(());
    };

    match target {
      TARGET_SUBJECT => {
        let row = sqlx::query_as::<_, SubjectRow>(
          r#"SELECT s.subject_id, s.subject_name, s.subject_name_cn, s.field_infobox,
                    s.subject_type_id, s.subject_nsfw, s.subject_ban,
                    s.subject_platform, s.field_meta_tags, f.field_tags,
                    s.subject_wish, s.subject_collect, s.subject_doing,
                    s.subject_on_hold, s.subject_dropped, f.field_rank, DATE_FORMAT(f.field_date, '%Y-%m-%d'),
                    f.field_rate_1, f.field_rate_2, f.field_rate_3, f.field_rate_4, f.field_rate_5,
                    f.field_rate_6, f.field_rate_7, f.field_rate_8, f.field_rate_9, f.field_rate_10,
                    f.field_redirect
             FROM chii_subjects s
             JOIN chii_subject_fields f ON f.field_sid = s.subject_id
             WHERE s.subject_id = ?
             LIMIT 1"#,
        )
        .bind(entity_id)
        .fetch_optional(&self.pool)
        .await
        .context("load subject")?;

        let Some(row) = row
        else {
          return self.delete(target, entity_id, op).await;
        };

        if row.subject_ban != 0 || row.field_redirect != 0 {
          return self.delete(target, entity_id, op).await;
        }

        let mut aliases = Vec::new();
        if !row.subject_name_cn.is_empty() {
          aliases.push(row.subject_name_cn.clone());
        }
        let wiki = parse_omit_error(&row.field_infobox);
        aliases.extend(extract_values_by_key(&wiki, "别名"));

        let meta_tag = split_space_values(&row.field_meta_tags);
        let tag = parse_subject_tags(&row.field_tags);
        let heat = row
          .subject_wish
          .saturating_add(row.subject_collect)
          .saturating_add(row.subject_doing)
          .saturating_add(row.subject_on_hold)
          .saturating_add(row.subject_dropped);
        let rating_total = row
          .field_rate_1
          .saturating_add(row.field_rate_2)
          .saturating_add(row.field_rate_3)
          .saturating_add(row.field_rate_4)
          .saturating_add(row.field_rate_5)
          .saturating_add(row.field_rate_6)
          .saturating_add(row.field_rate_7)
          .saturating_add(row.field_rate_8)
          .saturating_add(row.field_rate_9)
          .saturating_add(row.field_rate_10);
        let score_sum = (row.field_rate_1 as f64) * 1.0
          + (row.field_rate_2 as f64) * 2.0
          + (row.field_rate_3 as f64) * 3.0
          + (row.field_rate_4 as f64) * 4.0
          + (row.field_rate_5 as f64) * 5.0
          + (row.field_rate_6 as f64) * 6.0
          + (row.field_rate_7 as f64) * 7.0
          + (row.field_rate_8 as f64) * 8.0
          + (row.field_rate_9 as f64) * 9.0
          + (row.field_rate_10 as f64) * 10.0;
        let score = if rating_total == 0 {
          0.0
        } else {
          ((score_sum / (rating_total as f64)) * 10.0).round() / 10.0
        };

        let doc = SubjectDoc {
          id: row.subject_id,
          tag,
          meta_tag,
          name: row.subject_name,
          aliases,
          date: parse_date_val(&row.date),
          score,
          rating_count: rating_total,
          page_rank: rating_total as f64,
          heat,
          rank: row.field_rank,
          platform: row.subject_platform,
          type_id: row.subject_type_id,
          nsfw: row.subject_nsfw,
        };

        meili.update_doc("subjects", &[doc]).await?;
      }
      TARGET_CHARACTER => {
        let row = sqlx::query_as::<_, CharacterRow>(
          r#"SELECT crt_id, crt_name, crt_infobox, crt_comment, crt_collects, crt_nsfw, crt_redirect
             FROM chii_characters WHERE crt_id = ? LIMIT 1"#,
        )
        .bind(entity_id)
        .fetch_optional(&self.pool)
        .await
        .context("load character")?;

        let Some(row) = row else {
          return self.delete(target, entity_id, op).await;
        };

        if row.crt_redirect != 0 {
          return self.delete(target, entity_id, op).await;
        }

        let doc = CharacterDoc {
          id: row.crt_id,
          name: row.crt_name,
          aliases: extract_aliases(&parse_omit_error(&row.crt_infobox)),
          comment: row.crt_comment,
          collect: row.crt_collects,
          nsfw: row.crt_nsfw,
        };

        meili.update_doc("characters", &[doc]).await?;
      }
      TARGET_PERSON => {
        let row = sqlx::query_as::<_, PersonRow>(
          r#"SELECT prsn_id, prsn_name, prsn_infobox, prsn_comment, prsn_collects, prsn_redirect,
                    prsn_producer, prsn_mangaka, prsn_artist, prsn_seiyu,
                    prsn_writer, prsn_illustrator, prsn_actor
             FROM chii_persons WHERE prsn_id = ? LIMIT 1"#,
        )
        .bind(entity_id)
        .fetch_optional(&self.pool)
        .await
        .context("load person")?;

        let Some(row) = row else {
          return self.delete(target, entity_id, op).await;
        };

        if row.prsn_redirect != 0 {
          return self.delete(target, entity_id, op).await;
        }

        let doc = PersonDoc {
          id: row.prsn_id,
          name: row.prsn_name,
          aliases: extract_aliases(&parse_omit_error(&row.prsn_infobox)),
          comment: row.prsn_comment,
          collect: row.prsn_collects,
          career: collect_careers(
            row.prsn_producer,
            row.prsn_mangaka,
            row.prsn_artist,
            row.prsn_seiyu,
            row.prsn_writer,
            row.prsn_illustrator,
            row.prsn_actor,
          ),
        };

        meili.update_doc("persons", &[doc]).await?;
      }
      _ => return Err(anyhow!("unknown search target: {target}")),
    }

    tracing::info!(target, entity_id, op, action = "event_upsert", "search event handled");
    Ok(())
  }

  async fn delete(&self, target: &str, entity_id: u32, op: &str) -> anyhow::Result<()> {
    let Some(meili) = &self.meili else {
      tracing::debug!(target, entity_id, op, "skip search delete: meilisearch disabled");
      return Ok(());
    };

    let index = match target {
      TARGET_SUBJECT => "subjects",
      TARGET_CHARACTER => "characters",
      TARGET_PERSON => "persons",
      _ => return Err(anyhow!("unknown search target: {target}")),
    };

    meili.delete_doc(index, entity_id).await?;
    tracing::info!(target, entity_id, op, action = "event_delete", "search event handled");
    Ok(())
  }
}

fn split_space_values(input: &str) -> Vec<String> {
  input
    .split(' ')
    .map(ToOwned::to_owned)
    .collect()
}

fn wiki_values(v: &FieldValue) -> Vec<String> {
  match v {
    FieldValue::Scalar(text) => vec![text.clone()],
    FieldValue::Array(items) => items.iter().map(|x| x.value.clone()).collect(),
    FieldValue::Null => Vec::new(),
  }
}

fn extract_values_by_key(wiki: &Wiki, target_key: &str) -> Vec<String> {
  let mut out = Vec::new();
  for field in &wiki.fields {
    if field.key == target_key {
      out.extend(wiki_values(&field.value));
    }
  }
  out
}

fn extract_aliases(wiki: &Wiki) -> Vec<String> {
  let mut aliases = Vec::new();

  for field in &wiki.fields {
    if field.key == "中文名" {
      aliases.extend(wiki_values(&field.value));
    }
    if field.key == "简体中文名" {
      aliases.extend(wiki_values(&field.value));
    }
  }

  for field in &wiki.fields {
    if field.key == "别名" {
      aliases.extend(wiki_values(&field.value));
    }
  }

  aliases
}

fn parse_subject_tags(input: &str) -> Vec<String> {
  let parsed: Vec<SubjectTagItem> = match parse_php_serialize(input) {
    Ok(v) => v,
    Err(_) => return Vec::new(),
  };

  parsed
    .into_iter()
    .filter_map(|item| item.tag_name)
    .filter(|x| !x.is_empty())
    .collect()
}

fn parse_date_val(date: &str) -> i32 {
  if date.len() < 10 {
    return 0;
  }

  let year = date[0..4].parse::<i32>().ok();
  let month = date[5..7].parse::<i32>().ok();
  let day = date[8..10].parse::<i32>().ok();

  match (year, month, day) {
    (Some(y), Some(m), Some(d)) => y * 10000 + m * 100 + d,
    _ => 0,
  }
}

fn collect_careers(
  producer: bool,
  mangaka: bool,
  artist: bool,
  seiyu: bool,
  writer: bool,
  illustrator: bool,
  actor: bool,
) -> Vec<String> {
  let mut out = Vec::new();

  if writer {
    out.push("writer".to_string());
  }

  if producer {
    out.push("producer".to_string());
  }
  if mangaka {
    out.push("mangaka".to_string());
  }
  if artist {
    out.push("artist".to_string());
  }
  if seiyu {
    out.push("seiyu".to_string());
  }
  if illustrator {
    out.push("illustrator".to_string());
  }
  if actor {
    out.push("actor".to_string());
  }

  out
}

impl MeiliClient {
  async fn update_doc<T: Serialize + Send + Sync>(
    &self,
    index: &str,
    docs: &[T],
  ) -> anyhow::Result<()> {
    self
      .client
      .index(index)
      .add_documents(docs, Some("id"))
      .await
      .context("meilisearch update")?;
    Ok(())
  }

  async fn delete_doc(&self, index: &str, id: u32) -> anyhow::Result<()> {
    self
      .client
      .index(index)
      .delete_document(id)
      .await
      .context("meilisearch delete")?;
    Ok(())
  }
}

#[cfg(test)]
mod tests {
  use super::{collect_careers, extract_aliases, parse_subject_tags};
  use bangumi_wiki_parser::parse_omit_error;

  #[test]
  fn parse_subject_tags_php_serialized() {
    let raw = "a:3:{i:0;a:2:{s:8:\"tag_name\";s:6:\"动画\";s:6:\"result\";s:1:\"2\";}i:1;a:2:{s:8:\"tag_name\";N;s:6:\"result\";s:1:\"1\";}i:2;a:2:{s:8:\"tag_name\";s:2:\"TV\";s:6:\"result\";s:1:\"1\";}}";

    let tags = parse_subject_tags(raw);
    assert_eq!(tags, vec!["动画".to_string(), "TV".to_string()]);
  }

  #[test]
  fn extract_aliases_for_person_character() {
    let infobox = "{{Infobox\n|中文名=某角色\n|简体中文名=某角色简中\n|别名={\n[Alpha]\n[Beta]\n}\n|生日=2000-01-01\n}}";

    let aliases = extract_aliases(&parse_omit_error(infobox));
    assert_eq!(
      aliases,
      vec![
        "某角色".to_string(),
        "某角色简中".to_string(),
        "Alpha".to_string(),
        "Beta".to_string(),
      ]
    );
  }

  #[test]
  fn collect_career_order_matches_go() {
    let careers = collect_careers(true, true, true, true, true, true, true);
    assert_eq!(
      careers,
      vec![
        "writer".to_string(),
        "producer".to_string(),
        "mangaka".to_string(),
        "artist".to_string(),
        "seiyu".to_string(),
        "illustrator".to_string(),
        "actor".to_string(),
      ]
    );
  }
}
