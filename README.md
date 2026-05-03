# LeakGuard AI

LeakGuard AI is a revenue leakage detection platform.  
This repository is a **single-repo monorepo** containing backend services, frontend dashboard, shared schemas, and local infrastructure.

## What this repository contains

```text
ai-agent-leakguard/
  apps/
    api-gateway/          # Go HTTP gateway for external API access
    ingestion-service/    # Go event ingestion + persistence + orchestration
    rules-service/        # Go deterministic leakage checks
    case-service/         # Go case + evidence persistence/query APIs
    anomaly-service/      # Python FastAPI service for anomaly scoring
    web-dashboard/        # React + Vite dashboard
    scheduler-service/    # service placeholder
    notification-service/ # service placeholder
    retrieval-service/    # service placeholder
  packages/
    proto/                # shared contracts/proto placeholder
    shared-schemas/       # shared JSON schemas
  infra/
    docker/
      postgres/init.sql   # database bootstrap script
  docs/
```

## System context

The platform ingests revenue-related events (discounts, service delivery, credits/adjustments), evaluates leakage risk, and creates investigation cases.

Current end-to-end flow:

1. Client sends event to `api-gateway` (`POST /api/ingest`)
2. `ingestion-service` persists the event (`ingested_event`)
3. `rules-service` evaluates deterministic checks
4. `case-service` creates `leak_case` and `case_evidence`
5. `web-dashboard` reads and displays open cases (`GET /api/cases`)

## Services

| Service | Language | Responsibility |
|---|---|---|
| `api-gateway` | Go | API entrypoint, request routing, CORS |
| `ingestion-service` | Go | Event intake, persistence, rule/case orchestration |
| `rules-service` | Go | Deterministic checks for leakage patterns |
| `case-service` | Go | Case lifecycle data and evidence storage/query |
| `anomaly-service` | Python (FastAPI) | Anomaly scoring and backend reasoning APIs |
| `web-dashboard` | React + Vite | Analyst-facing case feed UI |

## Detection patterns currently implemented

- Discount threshold violation
- Unbilled service delivery (missing invoice reference)
- High-value credit/adjustment flag

## Data model (current)

PostgreSQL bootstrap script creates:

- `ingested_event`
- `leak_case`
- `case_evidence`

Schema file: `infra/docker/postgres/init.sql`

## Local development

Prerequisites:

- Docker Desktop (daemon running)
- pnpm 8.x (for local frontend development)

Start the stack:

```bash
docker compose up --build
```

Main endpoints:

- Web dashboard: `http://localhost:5173`
- API gateway health: `http://localhost:8080/healthz`
- Case list: `http://localhost:8080/api/cases`

### Backend reasoning flow (anomaly-service)

`anomaly-service` exposes `POST /reasoning/generate` for backend-triggered case reasoning.
When reasoning is enabled, the service generates grounded output and posts the artifact to `case-service`.

Required environment variables:

- `REASONING_ENABLED=true`
- `CASE_SERVICE_URL` (for local compose: `http://case-service:8083`)
- `VERTEX_PROJECT_ID`
- `VERTEX_LOCATION`
- `VERTEX_MODEL`
- `BASE_CURRENCY=MYR`

### Async event processing

Events are processed asynchronously via Asynq + Redis:

1. `ingestion-service` persists the event and enqueues a `process_event.v1` task
2. HTTP response returns `202 Accepted` with task ID (does not wait for processing)
3. Worker processes task: calls `rules-service` → creates case → triggers reasoning
4. Failures are retried with exponential backoff (5s → 30s max)
5. After 10 retries, failed tasks move to DLQ

#### Operations endpoints (internal-only)

- `GET /ops/queue/stats` - Queue metrics (size, processed, failed count)
- `GET /ops/queue/dlq` - List tasks in dead letter queue
- `POST /ops/queue/dlq/retry/{task_id}` - Manually retry a DLQ task

Example:

```bash
# View queue stats
curl http://localhost:8081/ops/queue/stats

# List DLQ tasks
curl http://localhost:8081/ops/queue/dlq

# Retry a failed task
curl -X POST http://localhost:8081/ops/queue/dlq/retry/task-id-here
```

Worker concurrency and retry behavior configured via environment:

- `WORKER_CONCURRENCY=10` - number of concurrent task workers
- `DLQ_MAX_RETRY=10` - max retry attempts before DLQ
- `MIN_BACKOFF_SECS=5` - initial retry delay
- `MAX_BACKOFF_SECS=30` - maximum retry delay

Frontend local commands:

```bash
cd apps/web-dashboard
pnpm install
pnpm dev
```

## Example ingest request

```bash
curl -X POST http://localhost:8080/api/ingest ^
  -H "Content-Type: application/json" ^
  -d "{\"event_type\":\"discount_event\",\"occurred_at\":\"2026-04-29T00:00:00Z\",\"payload\":{\"customer_id\":\"cust_123\",\"contract_id\":\"ctr_001\",\"invoice_id\":\"inv_1001\",\"amount\":12000,\"currency\":\"USD\",\"discount_pct\":0.22,\"allowed_discount_pct\":0.10}}"
```

## Frontend stack

`web-dashboard` uses **React + Vite** and communicates with backend APIs through `/api` via the gateway.

## Notes for contributors

- Keep service contracts explicit and versioned in `packages/`
- Keep domain logic in service boundaries (`ingestion`, `rules`, `case`)
- Keep DB changes in migration/bootstrap scripts under `infra/`
