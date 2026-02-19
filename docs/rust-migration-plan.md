# Rust Migration Implementation Plan

Date: 2026-02-19
Status: in-progress
Target stack: `tokio` + `rdkafka` + `sqlx`

## Scope (Phase 0 -> Phase 1)

This repository keeps Go as the primary runtime while introducing a Rust workspace for gradual migration.

Initial implementation goals:

1. Add a Rust workspace integrated at repository root.
2. Add a single app executable with top-level subcommands:
   - `worker` (contains `canal` / `timeline` placeholder runtime loops)
   - `server` (placeholder runtime loop)
3. Add shared config loading and connection bootstrap for MySQL/Kafka.
4. Keep all changes non-invasive to existing Go startup and deployment.

## Delivery milestones

### M0: Bootstrap
- Rust workspace compiles.
- `cargo run -p app -- worker canal` starts and exits gracefully.
- `cargo run -p app -- worker timeline` starts and exits gracefully.
- `cargo run -p app -- server` starts and exits gracefully.

### M1: Infra baseline
- Shared config supports environment variables.
- Kafka/MySQL clients can be initialized from config.
- Structured logging and basic shutdown signal handling are in place.

### M2: Contract baseline
- Event and payload schemas are defined in Rust types for timeline/canal.
- Golden fixtures can be added for parity tests.

## Out of scope (for this commit)

- Production traffic switching
- Full endpoint migration
- Replacing Go DAL with sqlx queries

## Next implementation tasks

1. Implement real search/session/redis side effects in canal handlers (consume loop and commit semantics are already in place).
2. Implement timeline publish API and payload parity tests.
3. Add sqlx pool and first read-only repository for subject read-path.
4. Add CI jobs to build Rust workspace and run tests.
