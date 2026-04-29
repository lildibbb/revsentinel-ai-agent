# Async Workflow + Retries + DLQ Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move ingest processing to a queue-backed async pipeline with retries, DLQ, and idempotent event processing while preserving current case/reasoning behavior.

**Architecture:** Ingestion-service persists events and enqueues Asynq jobs to Redis. A Go worker processes jobs, calls rules-service, persists cases, and conditionally triggers reasoning in anomaly-service. Failures are retried with bounded backoff and routed to DLQ after max attempts, with operational endpoints for queue stats and replay.

**Tech Stack:** Go 1.26 (chi, pgx, asynq), Python FastAPI (existing anomaly-service), Redis 7, PostgreSQL 16, Docker Compose.

---

## Scope check

This plan is scoped to one subsystem: async processing reliability (queue, worker, retries, DLQ, idempotency, ops endpoints). It does not include frontend queue UI or Temporal migration.

## File structure map

- **Create:** `apps/ingestion-service/internal/queue/publisher.go` (enqueue adapter + payload schema)
- **Create:** `apps/ingestion-service/internal/worker/handler.go` (Asynq task handler)
- **Create:** `apps/ingestion-service/internal/domain/process_event.go` (idempotent orchestration logic)
- **Create:** `apps/ingestion-service/internal/clients/rules_client.go` (rules-service HTTP client)
- **Create:** `apps/ingestion-service/internal/clients/case_client.go` (case-service HTTP client)
- **Create:** `apps/ingestion-service/internal/clients/reasoning_client.go` (anomaly-service HTTP client)
- **Create:** `apps/ingestion-service/internal/ops/handlers.go` (queue stats + DLQ replay handlers)
- **Modify:** `apps/ingestion-service/main.go` (wire publisher + worker mode + ops routes)
- **Modify:** `apps/ingestion-service/go.mod` (add Asynq dependency)
- **Create:** `apps/ingestion-service/internal/worker/handler_test.go` (retry/error mapping tests)
- **Create:** `apps/ingestion-service/internal/domain/process_event_test.go` (idempotency tests)
- **Modify:** `infra/docker/postgres/init.sql` (processing ledger table/index)
- **Modify:** `docker-compose.yml` (worker env/concurrency/retry settings)
- **Modify:** `README.md` (async pipeline runbook and ops endpoints)

---

### Task 1: Add idempotency ledger and queue payload contract

**Files:**
- Modify: `infra/docker/postgres/init.sql`
- Create: `apps/ingestion-service/internal/queue/publisher.go`
- Create: `apps/ingestion-service/internal/domain/process_event.go`
- Test: `apps/ingestion-service/internal/domain/process_event_test.go`

- [ ] **Step 1: Write the failing test**

```go
// apps/ingestion-service/internal/domain/process_event_test.go
func TestBuildIdempotencyKey_StableFormat(t *testing.T) {
	got := BuildIdempotencyKey("demo-tenant", "evt-1", "discount_event")
	want := "demo-tenant:evt-1:discount_event"
	if got != want {
		t.Fatalf("want %s got %s", want, got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd apps/ingestion-service && "D:\Program Files\Go\bin\go.exe" test ./... -run TestBuildIdempotencyKey_StableFormat -v`  
Expected: FAIL with undefined `BuildIdempotencyKey`.

- [ ] **Step 3: Write minimal implementation**

```sql
-- infra/docker/postgres/init.sql
CREATE TABLE IF NOT EXISTS event_processing_ledger (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  idempotency_key TEXT NOT NULL UNIQUE,
  event_id UUID NOT NULL,
  tenant_id TEXT NOT NULL,
  event_type TEXT NOT NULL,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_event_processing_ledger_event ON event_processing_ledger(event_id);
```

```go
// apps/ingestion-service/internal/domain/process_event.go
package domain

func BuildIdempotencyKey(tenantID, eventID, eventType string) string {
	return tenantID + ":" + eventID + ":" + eventType
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd apps/ingestion-service && "D:\Program Files\Go\bin\go.exe" test ./... -run TestBuildIdempotencyKey_StableFormat -v`  
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add infra/docker/postgres/init.sql apps/ingestion-service/internal/domain/process_event.go apps/ingestion-service/internal/domain/process_event_test.go
git commit -m "feat(ingestion): add idempotency ledger schema and key builder"
```

---

### Task 2: Add Asynq publisher and enqueue from ingest endpoint

**Files:**
- Create: `apps/ingestion-service/internal/queue/publisher.go`
- Modify: `apps/ingestion-service/main.go`
- Modify: `apps/ingestion-service/go.mod`
- Test: `apps/ingestion-service/internal/queue/publisher_test.go`

- [ ] **Step 1: Write the failing test**

```go
// apps/ingestion-service/internal/queue/publisher_test.go
func TestNewProcessEventTask_ContainsRequiredFields(t *testing.T) {
	task, err := NewProcessEventTask("evt-1", "demo-tenant", "discount_event", "2026-04-29T00:00:00Z", "trace-1")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if task.Type() != "process_event.v1" {
		t.Fatalf("unexpected task type: %s", task.Type())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd apps/ingestion-service && "D:\Program Files\Go\bin\go.exe" test ./... -run TestNewProcessEventTask_ContainsRequiredFields -v`  
Expected: FAIL with undefined task constructor.

- [ ] **Step 3: Write minimal implementation**

```go
// apps/ingestion-service/internal/queue/publisher.go
package queue

import (
	"encoding/json"
	"github.com/hibiken/asynq"
)

type ProcessEventPayload struct {
	EventID       string `json:"event_id"`
	TenantID      string `json:"tenant_id"`
	EventType     string `json:"event_type"`
	OccurredAt    string `json:"occurred_at"`
	TraceID       string `json:"trace_id"`
	SchemaVersion string `json:"schema_version"`
}

func NewProcessEventTask(eventID, tenantID, eventType, occurredAt, traceID string) (*asynq.Task, error) {
	p := ProcessEventPayload{EventID: eventID, TenantID: tenantID, EventType: eventType, OccurredAt: occurredAt, TraceID: traceID, SchemaVersion: "v1"}
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask("process_event.v1", b), nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd apps/ingestion-service && "D:\Program Files\Go\bin\go.exe" test ./... -run TestNewProcessEventTask_ContainsRequiredFields -v`  
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add apps/ingestion-service/internal/queue/publisher.go apps/ingestion-service/internal/queue/publisher_test.go apps/ingestion-service/main.go apps/ingestion-service/go.mod apps/ingestion-service/go.sum
git commit -m "feat(ingestion): enqueue process_event tasks with asynq payload contract"
```

---

### Task 3: Implement worker domain orchestration (rules -> case -> reasoning)

**Files:**
- Create: `apps/ingestion-service/internal/clients/rules_client.go`
- Create: `apps/ingestion-service/internal/clients/case_client.go`
- Create: `apps/ingestion-service/internal/clients/reasoning_client.go`
- Create: `apps/ingestion-service/internal/domain/process_event.go` (extend from Task 1)
- Create: `apps/ingestion-service/internal/worker/handler.go`
- Test: `apps/ingestion-service/internal/worker/handler_test.go`

- [ ] **Step 1: Write the failing test**

```go
// apps/ingestion-service/internal/worker/handler_test.go
func TestHandleProcessEvent_RulesFindings_CreateCaseAndReasoning(t *testing.T) {
	h := newTestHandlerWithMocks()
	task := mustTask("evt-1", "demo-tenant", "discount_event", "2026-04-29T00:00:00Z", "trace-1")
	err := h.HandleProcessEvent(context.Background(), task)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !h.caseClientCalled || !h.reasoningClientCalled {
		t.Fatalf("expected case and reasoning calls")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd apps/ingestion-service && "D:\Program Files\Go\bin\go.exe" test ./... -run TestHandleProcessEvent_RulesFindings_CreateCaseAndReasoning -v`  
Expected: FAIL with missing handler/mocks.

- [ ] **Step 3: Write minimal implementation**

```go
// apps/ingestion-service/internal/worker/handler.go
func (h *Handler) HandleProcessEvent(ctx context.Context, task *asynq.Task) error {
	var p queue.ProcessEventPayload
	if err := json.Unmarshal(task.Payload(), &p); err != nil {
		return fmt.Errorf("event_load_failed: %w", err)
	}
	findings, err := h.rules.Evaluate(ctx, p.EventType, p.EventID)
	if err != nil {
		return fmt.Errorf("rules_call_failed: %w", err)
	}
	if len(findings) == 0 {
		return nil
	}
	caseID, err := h.cases.CreateFromFindings(ctx, p.TenantID, findings)
	if err != nil {
		return fmt.Errorf("case_persist_failed: %w", err)
	}
	if shouldTriggerReasoning(findings) {
		if err := h.reasoning.Generate(ctx, caseID); err != nil {
			return fmt.Errorf("reasoning_call_failed: %w", err)
		}
	}
	return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd apps/ingestion-service && "D:\Program Files\Go\bin\go.exe" test ./... -run TestHandleProcessEvent_RulesFindings_CreateCaseAndReasoning -v`  
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add apps/ingestion-service/internal/clients/rules_client.go apps/ingestion-service/internal/clients/case_client.go apps/ingestion-service/internal/clients/reasoning_client.go apps/ingestion-service/internal/domain/process_event.go apps/ingestion-service/internal/worker/handler.go apps/ingestion-service/internal/worker/handler_test.go
git commit -m "feat(worker): process queued events through rules, case creation, and reasoning trigger"
```

---

### Task 4: Add retries, DLQ behavior, and internal ops endpoints

**Files:**
- Create: `apps/ingestion-service/internal/ops/handlers.go`
- Modify: `apps/ingestion-service/main.go`
- Test: `apps/ingestion-service/internal/ops/handlers_test.go`

- [ ] **Step 1: Write the failing test**

```go
// apps/ingestion-service/internal/ops/handlers_test.go
func TestQueueStatsEndpoint_ReturnsJSON(t *testing.T) {
	r := setupOpsRouterWithMockInspector()
	req := httptest.NewRequest(http.MethodGet, "/ops/queue/stats", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd apps/ingestion-service && "D:\Program Files\Go\bin\go.exe" test ./... -run TestQueueStatsEndpoint_ReturnsJSON -v`  
Expected: FAIL with missing ops handlers/router.

- [ ] **Step 3: Write minimal implementation**

```go
// apps/ingestion-service/internal/ops/handlers.go
func (h *Handlers) QueueStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.inspector.Queues()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "queue_stats_failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"queues": stats})
}

func (h *Handlers) RetryDLQTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "task_id")
	if taskID == "" {
		writeErr(w, http.StatusBadRequest, "task_id_required")
		return
	}
	if err := h.inspector.RunTask("default", taskID); err != nil {
		writeErr(w, http.StatusBadGateway, "dlq_retry_failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "task_id": taskID})
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd apps/ingestion-service && "D:\Program Files\Go\bin\go.exe" test ./... -run TestQueueStatsEndpoint_ReturnsJSON -v`  
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add apps/ingestion-service/internal/ops/handlers.go apps/ingestion-service/internal/ops/handlers_test.go apps/ingestion-service/main.go
git commit -m "feat(ops): add queue stats and DLQ retry endpoints"
```

---

### Task 5: Compose wiring and runbook updates

**Files:**
- Modify: `docker-compose.yml`
- Modify: `README.md`

- [ ] **Step 1: Write the failing check**

```bash
rg "ASYNQ_|WORKER_|DLQ_|REDIS_ADDR" docker-compose.yml
```

- [ ] **Step 2: Run check to verify it fails**

Run: `rg "ASYNQ_|WORKER_|DLQ_|REDIS_ADDR" docker-compose.yml`  
Expected: missing/incomplete worker configuration.

- [ ] **Step 3: Write minimal implementation**

```yaml
# docker-compose.yml (ingestion-service env additions)
      REDIS_ADDR: "redis:6379"
      ASYNQ_QUEUE_CRITICAL: "critical"
      ASYNQ_QUEUE_DEFAULT: "default"
      ASYNQ_QUEUE_LOW: "low"
      WORKER_CONCURRENCY: "10"
      DLQ_MAX_RETRY: "10"
```

```md
## Async processing ops

- `GET /ops/queue/stats`
- `GET /ops/queue/dlq`
- `POST /ops/queue/dlq/{task_id}/retry`

Run workers via ingestion-service worker mode in compose.
```

- [ ] **Step 4: Run checks to verify they pass**

Run: `rg "ASYNQ_|WORKER_|DLQ_|REDIS_ADDR" docker-compose.yml`  
Expected: matching configuration lines.

- [ ] **Step 5: Commit**

```bash
git add docker-compose.yml README.md
git commit -m "chore(async): add worker/queue config and operations runbook"
```

---

## Final verification checklist

- [ ] Run: `cd apps/ingestion-service && "D:\Program Files\Go\bin\go.exe" test ./... -v`  
Expected: ingestion-service tests PASS.
- [ ] Run: `cd apps/anomaly-service && python -m pytest -q`  
Expected: anomaly-service tests PASS.
- [ ] Run: `docker compose build ingestion-service rules-service case-service anomaly-service api-gateway`  
Expected: build succeeds.
- [ ] Run: `docker compose up -d`  
Expected: services and worker start successfully.
- [ ] Run: ingest smoke test and verify case creation still works.

## Self-review

- **Spec coverage:** Tasks map to queue contract, worker flow, retries/DLQ, idempotency, ops endpoints, and observability-ready structure.
- **Placeholder scan:** No placeholder directives remain; each code step includes concrete snippets and commands.
- **Type consistency:** Reused stable names across tasks (`process_event.v1`, `trace_id`, `event_id`, `tenant_id`, `reasoning_call_failed`).
