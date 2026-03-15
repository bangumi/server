use std::env;

use anyhow::Context;

#[derive(Debug, Clone)]
pub struct AppConfig {
  pub mysql_dsn: String,
  pub redis_url: String,
  pub meilisearch_url: String,
  pub meilisearch_key: String,
  pub s3_entry_point: String,
  pub s3_access_key: String,
  pub s3_secret_key: String,
  pub s3_image_resize_bucket: String,
  pub s3_region: Option<String>,
  pub kafka_brokers: String,
  pub kafka_group_id: String,
  pub kafka_topic: String,
  pub kafka_topics: Vec<String>,
}

impl AppConfig {
  pub fn from_env(service: &str) -> anyhow::Result<Self> {
    let mysql_dsn = env::var("RUST_MYSQL_DSN").context("missing env RUST_MYSQL_DSN")?;

    let redis_url = env::var("RUST_REDIS_URL")
      .or_else(|_| env::var("REDIS_URI"))
      .unwrap_or_else(|_| "redis://127.0.0.1:6379/0".to_string());

    let meilisearch_url = env::var("RUST_MEILISEARCH_URL")
      .or_else(|_| env::var("MEILISEARCH_URL"))
      .unwrap_or_default();

    let meilisearch_key = env::var("RUST_MEILISEARCH_KEY")
      .or_else(|_| env::var("MEILISEARCH_KEY"))
      .unwrap_or_default();

    let s3_entry_point = env::var("RUST_S3_ENTRY_POINT")
      .or_else(|_| env::var("S3_ENTRY_POINT"))
      .unwrap_or_default();

    let s3_access_key = env::var("RUST_S3_ACCESS_KEY")
      .or_else(|_| env::var("S3_ACCESS_KEY"))
      .unwrap_or_default();

    let s3_secret_key = env::var("RUST_S3_SECRET_KEY")
      .or_else(|_| env::var("S3_SECRET_KEY"))
      .unwrap_or_default();

    let s3_image_resize_bucket = env::var("RUST_S3_IMAGE_RESIZE_BUCKET")
      .or_else(|_| env::var("S3_IMAGE_RESIZE_BUCKET"))
      .unwrap_or_else(|_| "img-resize".to_string());

    let s3_region = env::var("RUST_S3_REGION")
      .ok()
      .or_else(|| env::var("AWS_REGION").ok())
      .filter(|x| !x.trim().is_empty());

    let kafka_brokers = env::var("RUST_KAFKA_BROKERS")
      .or_else(|_| env::var("KAFKA_BROKER"))
      .context("missing env RUST_KAFKA_BROKERS or KAFKA_BROKER")?;

    let kafka_group_id =
      env::var("RUST_KAFKA_GROUP_ID").unwrap_or_else(|_| format!("go-{service}"));

    let kafka_topic =
      env::var("RUST_KAFKA_TOPIC").unwrap_or_else(|_| "timeline".to_string());

    let kafka_topics = parse_topics(
      env::var("RUST_KAFKA_TOPICS")
        .ok()
        .or_else(|| env::var("KAFKA_TOPICS").ok()),
    );

    Ok(Self {
      mysql_dsn,
      redis_url,
      meilisearch_url,
      meilisearch_key,
      s3_entry_point,
      s3_access_key,
      s3_secret_key,
      s3_image_resize_bucket,
      s3_region,
      kafka_brokers,
      kafka_group_id,
      kafka_topic,
      kafka_topics,
    })
  }
}

fn parse_topics(raw: Option<String>) -> Vec<String> {
  raw
    .unwrap_or_default()
    .split(',')
    .map(str::trim)
    .filter(|x| !x.is_empty())
    .map(ToOwned::to_owned)
    .collect()
}

pub async fn build_mysql_pool(cfg: &AppConfig) -> anyhow::Result<sqlx::MySqlPool> {
  let pool = sqlx::mysql::MySqlPoolOptions::new()
    .max_connections(5)
    .connect(&cfg.mysql_dsn)
    .await
    .context("failed to connect mysql")?;

  Ok(pool)
}

pub fn build_kafka_client_config(cfg: &AppConfig) -> rdkafka::ClientConfig {
  let mut c = rdkafka::ClientConfig::new();
  c.set("bootstrap.servers", &cfg.kafka_brokers);
  c
}
