use anyhow::Context;
use common::config::{build_kafka_client_config, build_mysql_pool, AppConfig};
use rdkafka::producer::FutureProducer;

const TIMELINE_TOPIC: &str = "timeline";

pub async fn run() -> anyhow::Result<()> {
  let cfg = AppConfig::from_env("timeline-worker")?;
  let _mysql = build_mysql_pool(&cfg).await?;

  let _producer: FutureProducer = build_kafka_client_config(&cfg)
    .create()
    .context("create kafka producer")?;

  tracing::info!(topic = TIMELINE_TOPIC, "timeline worker started");

  tokio::signal::ctrl_c().await.context("wait ctrl-c")?;

  tracing::info!("timeline worker shutdown");

  Ok(())
}
