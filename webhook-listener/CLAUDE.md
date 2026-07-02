# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this service does

`webhook-listener` is a Go worker that consumes events from a Kafka topic (`webhook-events`), resolves the destination webhook config, delivers the event via HTTP POST, and records the outcome in PostgreSQL, retrying failed deliveries with backoff. There is no HTTP server, CLI, or API in this repo — it is a headless consumer process (`cmd/main.go` is the only entrypoint).

This repo does **not** own the `webhooks` / `webhook_events` tables — an external service ("webhook-event-generator") creates and writes to them and is the producer of the Kafka messages this service consumes (see comments in `sql/schema.sql` and `internal/model/event.go`). `sql/schema.sql` only adds `webhook_event_logs` plus a `next_retry_at` column/index on top of that pre-existing schema — treat those tables' existing columns as a fixed external contract, not something to redesign here.

## Commands

```
go build -o webhook-listener.exe ./cmd   # build
go run ./cmd                             # run locally (loads .env via godotenv)

Before running, apply `sql/schema.sql` once against the target Postgres database. The service also needs a reachable Kafka broker and Redis instance — see `.env.example` for all configurable env vars (Kafka brokers/topic/group, `DATABASE_URL`, Redis addr/password/db, `WORKER_COUNT`, `HTTP_TIMEOUT`, `RETRY_BASE_DELAY`, `RETRY_CHECK_INTERVAL`, `CACHE_TTL`). Config loading/defaults live in `internal/config/config.go`.

## Architecture

Wiring happens in `cmd/main.go`: Postgres pool → Redis cache → repositories → services → Kafka producer → retry scheduler (goroutine) → Kafka consumer group (goroutine), then blocks on SIGINT/SIGTERM and drains in-flight work before exiting.

**Delivery pipeline** (`internal/consumer/handler.go` → `internal/service/dispatcher.go`):
- `groupHandler.ConsumeClaim` runs once per assigned Kafka partition. It marks the Kafka offset immediately (at-least-once semantics), then hands the payload to a goroutine bounded by a single shared semaphore (`WORKER_COUNT`) — this caps total concurrent HTTP deliveries *across all partitions*, not per partition.
- `Dispatcher.Process` unmarshals the Kafka message into `model.KafkaEvent`, resolves the webhook config via `WebhookService.GetConfig`, then does the actual delivery: HTTP POST → update `webhook_events` status (`MarkSuccess`/`MarkRetry`) → insert a row into `webhook_event_logs`.
- `WebhookService.GetConfig` is Redis-first, Postgres-fallback with write-through caching (`internal/cache/redis.go`, key prefix `webhook:config:`). A Redis outage logs a warning and falls through to the DB — cache is never on the critical failure path.

**Retry model** has two independent layers:
1. Inline backoff on failure: `EventRepo.MarkRetry` sets `status = PENDING_RETRY` and a short `next_retry_at` (5s/10s/30s escalating by DB `attempt_count`), directly from the dispatch path.
2. `RetryScheduler` (`internal/consumer/retry_scheduler.go`) polls every `RETRY_CHECK_INTERVAL` for `PENDING_RETRY` rows past due, claims a batch with `FOR UPDATE SKIP LOCKED`, and leases them (pushes `next_retry_at` forward ~60s) inside the same transaction so a crash mid-batch can't cause double-claims. For each claimed row: if `attempt_count >= max_attempts` it's marked `FAILED_PERMANENTLY` and the owning webhook is set inactive (`WebhookService.MarkInactive`, which also busts the Redis cache); otherwise the raw payload is re-published to the Kafka topic via `KafkaProducer`.

Delivery status codes (`internal/model/event.go`): `0=PENDING, 1=SUCCESS, 2=PENDING_RETRY, 3=FAILED_PERMANENTLY`. `MarkSuccess`/`MarkRetry` guard their `UPDATE` with `status NOT IN (SUCCESS, FAILED_PERMANENTLY)` so a delayed/duplicate Kafka redelivery can't override an already-terminal row.

