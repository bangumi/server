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

## Server API migration status (`/v0`)

Implemented in Rust (`crates/app/src/server`):

- Search:
  - `POST /v0/search/subjects`
  - `POST /v0/search/characters`
  - `POST /v0/search/persons`
- Subject read APIs:
  - `GET /v0/subjects/{subject_id}`
  - `GET /v0/subjects/{subject_id}/image`
  - `GET /v0/subjects/{subject_id}/subjects`
  - `GET /v0/subjects/{subject_id}/persons`
  - `GET /v0/subjects/{subject_id}/characters`
- Character read/write APIs:
  - `GET /v0/characters/{character_id}`
  - `GET /v0/characters/{character_id}/image`
  - `GET /v0/characters/{character_id}/subjects`
  - `GET /v0/characters/{character_id}/persons`
  - `POST /v0/characters/{character_id}/collect`
  - `DELETE /v0/characters/{character_id}/collect`
- Person read APIs:
  - `GET /v0/persons/{person_id}`
  - `GET /v0/persons/{person_id}/image`
  - `GET /v0/persons/{person_id}/subjects`
  - `GET /v0/persons/{person_id}/characters`

## Behavior and test parity notes

- Request-scoped auth is resolved once in middleware and injected through request extensions (`RequestAuth`).
- OAuth token lookup no longer relies on SQL `CAST`; `user_id` is read as string and validated/parsed in Rust before member lookup.
- Route behavior parity tests are in place (`server_smoke`, `server_real_deps`) and currently passing with `cargo test -p app`.
