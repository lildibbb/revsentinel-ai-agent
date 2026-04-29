# RevSentinel AI - Async Workflow, Retries, and DLQ Design

## 1. Purpose

Define the next enterprise-grade backend milestone: move event processing from synchronous request path to a queue-backed asynchronous workflow with retries, DLQ, and idempotent processing.

This milestone must preserve current behavior while improving reliability, traceability, and scalability.

## 2. Scope

In scope:

- Asynq + Redis as primary workflow backbone.
- Event enqueue from ingestion-service.
- Go worker processing pipeline for rules + case persistence + reasoning trigger.
- Retry/backoff + dead-letter queue behavior.
- Idempotency to prevent duplicate case creation.
- Queue stats and DLQ replay operational endpoints.

Out of scope:

- Migration to Temporal in this milestone.
- Kafka/NATS stream architecture.
- New frontend UX for queue operations.

## 3. Architectural decision

Selected approach: **Asynq now, Temporal-ready architecture**.

Why:

- Fastest to deliver in current stack (Redis already present).
- Strong enough reliability primitives (retries, scheduled retries, dead queue).
- Lower operational complexity than Temporal/Kafka at current scope.
- Domain handlers are isolated so workflow engine can be swapped later.

## 4. High-level architecture

### Components

- `api-gateway` (Go): unchanged external API entrypoint.
- `ingestion-service` (Go): persists events and enqueues workflow tasks.
- `worker` (Go, within ingestion-service or separate app): consumes Asynq tasks.
- `rules-service` (Go): deterministic checks.
- `case-service` (Go): case/evidence/reasoning persistence.
- `anomaly-service` (Python): reasoning/model endpoint only.
- Redis: queue + retry state + dead queue.

### Separation of concerns

- Go: orchestration, queue semantics, retries, idempotency, deterministic controls.
- Python: model/reasoning interface and validation.
- Queue payload contracts are versioned and explicit in shared schemas.

## 5. Data flow

1. `POST /api/ingest` receives event.
2. ingestion-service stores `ingested_event`.
3. ingestion-service enqueues `process_event` task in Asynq.
4. worker consumes task and loads event context.
5. worker calls rules-service.
6. worker persists findings as cases in case-service.
7. worker optionally triggers reasoning generation based on severity/type gates.
8. worker persists/links reasoning artifact.
9. task success acknowledged.

On failure:

- Retry with exponential backoff up to max attempts.
- If exhausted, task moves to DLQ with structured failure metadata.

## 6. Queue and task contract

Task type:

- `process_event.v1`

Payload:

- `event_id` (uuid)
- `tenant_id` (string)
- `event_type` (string)
- `occurred_at` (RFC3339 timestamp)
- `trace_id` (string)
- `schema_version` (`v1`)

Queue classes:

- `critical`: high severity paths.
- `default`: normal event processing.
- `low`: backfill/non-urgent tasks.

## 7. Idempotency and consistency

Idempotency key:

- `tenant_id:event_id:event_type`

Rules:

- duplicate enqueue is allowed, duplicate processing is not.
- case creation/upsert path must be idempotent for same key.
- reasoning generation should skip duplicate successful artifacts unless forced.

## 8. Error taxonomy and observability

Error codes:

- `enqueue_failed`
- `event_load_failed`
- `rules_call_failed`
- `case_persist_failed`
- `reasoning_call_failed`
- `reasoning_persist_failed`
- `dlq_write_failed`

Observability:

- propagate `trace_id` through enqueue, worker, downstream service calls.
- structured logs include task id, queue name, attempt, error_code.
- metrics: queue depth, retry count, processing latency, DLQ size.

## 9. Operational API (internal/admin)

- `GET /ops/queue/stats`: queue depth, retries, throughput.
- `GET /ops/queue/dlq`: list dead tasks with failure metadata.
- `POST /ops/queue/dlq/{task_id}/retry`: replay one DLQ task.

These endpoints are internal and require service-level authentication.

## 10. Enterprise Go/Python structure

Go (ingestion-service):

- `internal/queue` - Asynq enqueue/dequeue adapter.
- `internal/worker` - task handlers and retry policy.
- `internal/domain` - idempotent process_event use case.
- `internal/clients` - rules/case/reasoning clients.
- `internal/ops` - queue stats and DLQ management handlers.

Python (anomaly-service):

- keep reasoning API contract stable.
- no queue orchestration in Python.
- strict schema validation + typed errors.

## 11. Testing strategy

Unit tests:

- enqueue payload validation.
- idempotency key generation and duplicate handling.
- retry policy mapping by error class.

Integration tests:

- worker with Redis and mocked downstream services.
- DLQ movement after max attempts.
- DLQ retry endpoint replays successfully.

Contract tests:

- reasoning endpoint request/response schema.
- case-service persistence payload compatibility.

End-to-end smoke:

- ingest -> queue -> rules -> case -> reasoning -> persisted artifact.

## 12. Acceptance criteria

- synchronous ingest path no longer performs full downstream processing inline.
- worker processes queued events successfully under normal conditions.
- retries and DLQ behavior are visible and operable.
- duplicate events do not create duplicate cases.
- reasoning remains gated by severity/type policy and persists as expected.

## 13. Temporal-ready migration boundary

Keep interfaces abstracted:

- `QueuePublisher`
- `TaskWorker`
- `ProcessEventUseCase`

As long as these interfaces remain stable, replacing Asynq with Temporal later does not require rewriting domain logic.
