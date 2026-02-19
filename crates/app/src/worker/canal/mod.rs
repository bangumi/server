use anyhow::{anyhow, Context};
use common::config::{build_kafka_client_config, build_mysql_pool, AppConfig};
use common::locate_error;
use rdkafka::consumer::{CommitMode, Consumer};
use rdkafka::Message;

mod character;
mod person;
mod search_event;
mod subject;
mod types;
mod user;

use types::DebeziumPayload;

pub async fn run() -> anyhow::Result<()> {
  let cfg = AppConfig::from_env("canal")?;

  if cfg.kafka_topics.is_empty() {
    return Err(anyhow!(
      "empty topics: set RUST_KAFKA_TOPICS (comma-separated) or KAFKA_TOPICS"
    ));
  }

  let mysql_pool = build_mysql_pool(&cfg).await?;
  let search = search_event::SearchDispatcher::new(&cfg, mysql_pool.clone())?;
  let user = user::UserDispatcher::new(&cfg, mysql_pool)?;

  let mut client = build_kafka_client_config(&cfg);
  client.set("group.id", &cfg.kafka_group_id);
  client.set("enable.auto.commit", "false");

  let consumer: rdkafka::consumer::StreamConsumer =
    client.create().context("create kafka consumer")?;

  let topics: Vec<&str> = cfg.kafka_topics.iter().map(String::as_str).collect();
  consumer
    .subscribe(&topics)
    .context("subscribe kafka topics")?;

  tracing::info!(
    group_id = %cfg.kafka_group_id,
    topics = ?cfg.kafka_topics,
    "canal worker started"
  );

  loop {
    tokio::select! {
      _ = tokio::signal::ctrl_c() => {
        tracing::info!("received shutdown signal");
        return Ok(());
      }
      msg = consumer.recv() => {
        let msg = match msg {
          Ok(item) => item,
          Err(err) => {
            tracing::error!(error = ?err, "failed to fetch kafka message");
            continue;
          }
        };

        let key = msg.key().unwrap_or_default();
        let value = msg.payload().unwrap_or_default();

        if let Err(err) = handle_message(&search, &user, key, value).await {
          if let Some(at) = locate_error(&err) {
            tracing::error!(
              error.file = at.file,
              error.lino = at.line,
              error.message = at.message,
              error_chain = %format!("{err:#}"),
              "failed to process kafka message"
            );
          } else {
            tracing::error!(error = ?err, error_chain = %format!("{err:#}"), "failed to process kafka message");
          }
          continue;
        }

        if let Err(err) = consumer.commit_message(&msg, CommitMode::Sync) {
          tracing::error!(error = ?err, "failed to commit kafka message");
        }
      }
    }
  }
}

async fn handle_message(
  search: &search_event::SearchDispatcher,
  user: &user::UserDispatcher,
  key: &[u8],
  value: &[u8],
) -> anyhow::Result<()> {
  if value.is_empty() {
    return Ok(());
  }

  let payload: DebeziumPayload = match serde_json::from_slice(value) {
    Ok(payload) => payload,
    Err(err) => {
      tracing::warn!(error = ?err, "failed to parse kafka value, skip message");
      return Ok(());
    }
  };

  match payload.source.table.as_str() {
    "chii_subject_fields" => subject::on_subject_field(search, key, &payload.op).await?,
    "chii_subjects" => subject::on_subject(search, key, &payload.op).await?,
    "chii_characters" => character::on_character(search, key, &payload.op).await?,
    "chii_persons" => person::on_person(search, key, &payload.op).await?,
    "chii_members" => user
      .on_user(key, &payload.op, payload.before, payload.after)
      .await?,
    _ => tracing::debug!(table = %payload.source.table, "ignored table event"),
  }

  Ok(())
}

#[cfg(test)]
mod tests {
  use common::ResultExt;

  #[test]
  fn demo_error_logging_output() {
    common::init_tracing();

    let err = demo_error().expect_err("demo error should fail");
    if let Some(at) = common::locate_error(&err) {
      tracing::error!(
        error.file = at.file,
        error.lino = at.line,
        error.message = at.message,
        error_chain = %format!("{err:#}"),
        "demo error output"
      );
    } else {
      tracing::error!(
        error = ?err,
        error_chain = %format!("{err:#}"),
        "demo error output"
      );
    }
  }

  fn demo_error() -> anyhow::Result<()> {
    let _: serde_json::Value =
      serde_json::from_slice(b"not-json").context_loc("parse debezium payload")?;
    Ok(())
  }
}
