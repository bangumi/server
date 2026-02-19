# Rust Migration Workspace

Rust migration code is integrated into repository root.

## Crates

- `common`: shared config/bootstrap/helpers
- `app`: single executable with top-level subcommands (`worker`, `server`)

## Environment variables

- `RUST_MYSQL_DSN` (required)
- `RUST_KAFKA_BROKERS` (required, fallback: `KAFKA_BROKER`)
- `RUST_KAFKA_TOPICS` (required for `worker canal`, comma-separated)
- `RUST_KAFKA_GROUP_ID` (optional)
- `RUST_KAFKA_TOPIC` (optional, default: `timeline`)
- `RUST_LOG` (optional, default: `info`)

## Run

```bash
cargo run -p app -- worker canal
cargo run -p app -- worker timeline
cargo run -p app -- server
```

## Current migration status

- `worker canal`: real Kafka consume loop with Debezium payload parsing, table-based dispatch, and commit-after-success behavior.
- `worker timeline`: producer bootstrap and reusable timeline producer module are ready.
