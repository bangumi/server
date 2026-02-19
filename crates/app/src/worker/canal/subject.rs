use anyhow::Context;
use serde::Deserialize;

use super::search_event::SearchDispatcher;
use super::search_event::{self};

#[derive(Debug, Deserialize)]
struct SubjectKey {
  subject_id: u32,
}

#[derive(Debug, Deserialize)]
struct SubjectFieldKey {
  field_sid: u32,
}

pub async fn on_subject(search: &SearchDispatcher, key: &[u8], op: &str) -> anyhow::Result<()> {
  let key: SubjectKey = serde_json::from_slice(key).context("parse subject key")?;
  on_subject_change(search, key.subject_id, op).await?;
  Ok(())
}

pub async fn on_subject_field(
  search: &SearchDispatcher,
  key: &[u8],
  op: &str,
) -> anyhow::Result<()> {
  let key: SubjectFieldKey = serde_json::from_slice(key).context("parse subject field key")?;
  on_subject_change(search, key.field_sid, op).await?;
  Ok(())
}

async fn on_subject_change(
  search: &SearchDispatcher,
  subject_id: u32,
  op: &str,
) -> anyhow::Result<()> {
  search.dispatch(search_event::TARGET_SUBJECT, subject_id, op).await
}
