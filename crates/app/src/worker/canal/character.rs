use anyhow::Context;
use serde::Deserialize;

use super::search_event::SearchDispatcher;
use super::search_event::{self};

#[derive(Debug, Deserialize)]
struct CharacterKey {
  crt_id: u32,
}

pub async fn on_character(
  search: &SearchDispatcher,
  key: &[u8],
  op: &str,
) -> anyhow::Result<()> {
  let key: CharacterKey = serde_json::from_slice(key).context("parse character key")?;
  on_character_change(search, key.crt_id, op).await?;
  Ok(())
}

async fn on_character_change(
  search: &SearchDispatcher,
  character_id: u32,
  op: &str,
) -> anyhow::Result<()> {
  search
    .dispatch(search_event::TARGET_CHARACTER, character_id, op)
    .await
}
