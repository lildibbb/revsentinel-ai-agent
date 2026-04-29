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
| `anomaly-service` | Python (FastAPI) | Anomaly scoring API (current stub implementation) |
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
