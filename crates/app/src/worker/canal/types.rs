use serde::Deserialize;
use serde_json::Value;

pub const OP_CREATE: &str = "c";
pub const OP_DELETE: &str = "d";
pub const OP_UPDATE: &str = "u";
pub const OP_SNAPSHOT: &str = "r";

#[derive(Debug, Deserialize)]
pub struct DebeziumPayload {
  pub before: Option<Value>,
  pub after: Option<Value>,
  pub source: DebeziumSource,
  pub op: String,
}

#[derive(Debug, Deserialize)]
pub struct DebeziumSource {
  pub table: String,
}
