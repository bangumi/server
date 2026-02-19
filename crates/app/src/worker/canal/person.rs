use anyhow::Context;
use serde::Deserialize;

use super::search_event::SearchDispatcher;
use super::search_event::{self};

#[derive(Debug, Deserialize)]
struct PersonKey {
  prsn_id: u32,
}

pub async fn on_person(search: &SearchDispatcher, key: &[u8], op: &str) -> anyhow::Result<()> {
  let key: PersonKey = serde_json::from_slice(key).context("parse person key")?;
  on_person_change(search, key.prsn_id, op).await?;
  Ok(())
}

async fn on_person_change(
  search: &SearchDispatcher,
  person_id: u32,
  op: &str,
) -> anyhow::Result<()> {
  search.dispatch(search_event::TARGET_PERSON, person_id, op).await
}
