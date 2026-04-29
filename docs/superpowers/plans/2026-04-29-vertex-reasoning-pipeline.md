# Vertex Reasoning Pipeline Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a backend-only, grounded Vertex Gemini reasoning pipeline for leak cases, with persisted reasoning artifacts and MYR-normalized money fields.

**Architecture:** Keep deterministic leak detection unchanged. Extend case-service with a `case_reasoning` artifact contract, then add reasoning orchestration in anomaly-service (FastAPI) that fetches case context, applies FX normalization, calls Vertex, validates structured output, and persists results back to case-service.

**Tech Stack:** Go (chi + pgx), Python (FastAPI + pydantic), PostgreSQL (jsonb), Vertex AI Gemini API, pytest, Go test.

---

## Scope check

This plan covers one subsystem only: **reasoning pipeline**. It does not include frontend implementation or full retrieval platform build-out.

## File structure map (lock this before coding)

- **Modify:** `infra/docker/postgres/init.sql`
  - Add `case_reasoning` table and indexes.
- **Modify:** `apps/case-service/main.go`
  - Add create/read reasoning endpoints and persistence helpers.
- **Create:** `apps/case-service/reasoning_models.go`
  - Dedicated request/response structs for reasoning artifact contract.
- **Create:** `apps/case-service/reasoning_handlers_test.go`
  - API contract tests for reasoning write/read paths.
- **Modify:** `apps/anomaly-service/requirements.txt`
  - Add test/runtime deps for reasoning path.
- **Create:** `apps/anomaly-service/app/schemas.py`
  - Pydantic schemas for request, context, output, persistence payload.
- **Create:** `apps/anomaly-service/app/fx.py`
  - FX adapter contract (MYR normalization).
- **Create:** `apps/anomaly-service/app/retrieval.py`
  - Policy/contract retrieval adapter contract.
- **Create:** `apps/anomaly-service/app/vertex_client.py`
  - Vertex Gemini adapter contract.
- **Modify:** `apps/anomaly-service/app/main.py`
  - Add `/reasoning/generate` and wiring.
- **Create:** `apps/anomaly-service/tests/test_reasoning_generate.py`
  - Endpoint behavior tests (success/failure/degraded cases).
- **Modify:** `docker-compose.yml`
  - Add Vertex/FX/reasoning env variables.
- **Modify:** `README.md`
  - Add reasoning API usage and env setup.

---

### Task 1: Add case-service reasoning persistence contract

**Files:**
- Modify: `infra/docker/postgres/init.sql`
- Create: `apps/case-service/reasoning_models.go`
- Modify: `apps/case-service/main.go`
- Test: `apps/case-service/reasoning_handlers_test.go`

- [ ] **Step 1: Write the failing test**

```go
// apps/case-service/reasoning_handlers_test.go
func TestCreateReasoning_ValidRequest_ReturnsCreated(t *testing.T) {
	reqBody := `{
	  "status":"success",
	  "model_provider":"vertex-ai",
	  "model_name":"gemini-2.5-pro",
	  "prompt_version":"v1",
	  "response_schema_version":"v1",
	  "summary":"Risk detected",
	  "recommended_action":"Review contract exception",
	  "citations":[{"source":"policy","reference":"POL-12","excerpt":"Max discount 10%"}],
	  "amount_original":12000,
	  "currency_original":"USD",
	  "amount_myr_normalized":56400,
	  "fx_rate_to_myr":4.7
	}`
	_ = reqBody
	t.Fatal("fail until handler + persistence contract are implemented")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd apps/case-service && go test ./... -run TestCreateReasoning_ValidRequest_ReturnsCreated -v`  
Expected: FAIL with explicit failure from test scaffold.

- [ ] **Step 3: Write minimal implementation**

```sql
-- infra/docker/postgres/init.sql
CREATE TABLE IF NOT EXISTS case_reasoning (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  case_id UUID NOT NULL REFERENCES leak_case(id) ON DELETE CASCADE,
  tenant_id TEXT NOT NULL,
  status TEXT NOT NULL,
  model_provider TEXT NOT NULL,
  model_name TEXT NOT NULL,
  model_version TEXT NULL,
  prompt_version TEXT NOT NULL,
  response_schema_version TEXT NOT NULL,
  summary TEXT NOT NULL,
  recommended_action TEXT NOT NULL,
  rationale JSONB NULL,
  confidence DOUBLE PRECISION NULL,
  citations JSONB NOT NULL,
  token_input INT NULL,
  token_output INT NULL,
  latency_ms INT NULL,
  trace_id TEXT NOT NULL,
  error_code TEXT NULL,
  error_message TEXT NULL,
  amount_original DOUBLE PRECISION NULL,
  currency_original TEXT NULL,
  amount_myr_normalized DOUBLE PRECISION NULL,
  fx_rate_to_myr DOUBLE PRECISION NULL,
  fx_rate_timestamp TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

```go
// apps/case-service/main.go (add routes)
r.Post("/cases/{id}/reasoning", createReasoningHandler(pool))
r.Get("/cases/{id}/reasoning/latest", getLatestReasoningHandler(pool))
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd apps/case-service && go test ./... -run TestCreateReasoning_ValidRequest_ReturnsCreated -v`  
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add infra/docker/postgres/init.sql apps/case-service/main.go apps/case-service/reasoning_models.go apps/case-service/reasoning_handlers_test.go
git commit -m "feat(case-service): add case_reasoning persistence and API contract"
```

---

### Task 2: Add reasoning endpoint in anomaly-service with strict schema

**Files:**
- Modify: `apps/anomaly-service/requirements.txt`
- Create: `apps/anomaly-service/app/schemas.py`
- Modify: `apps/anomaly-service/app/main.py`
- Test: `apps/anomaly-service/tests/test_reasoning_generate.py`

- [ ] **Step 1: Write the failing test**

```python
# apps/anomaly-service/tests/test_reasoning_generate.py
from fastapi.testclient import TestClient
from app.main import app

def test_generate_reasoning_success_contract():
    client = TestClient(app)
    res = client.post("/reasoning/generate", json={"case_id": "00000000-0000-0000-0000-000000000001"})
    assert res.status_code == 200
    body = res.json()
    assert body["status"] in {"success", "failed"}
    assert "prompt_version" in body
    assert "response_schema_version" in body
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd apps/anomaly-service && python -m pytest tests/test_reasoning_generate.py::test_generate_reasoning_success_contract -q`  
Expected: FAIL with 404 or missing fields.

- [ ] **Step 3: Write minimal implementation**

```python
# apps/anomaly-service/app/schemas.py
from pydantic import BaseModel
from typing import Optional, List, Dict, Any

class GenerateReasoningRequest(BaseModel):
    case_id: str
    force_regenerate: bool = False

class Citation(BaseModel):
    source: str
    reference: str
    excerpt: str

class GenerateReasoningResponse(BaseModel):
    status: str
    prompt_version: str
    response_schema_version: str
    summary: str
    recommended_action: str
    citations: List[Citation]
    amount_original: Optional[float] = None
    currency_original: Optional[str] = None
    amount_myr_normalized: Optional[float] = None
    fx_rate_to_myr: Optional[float] = None
    metadata: Dict[str, Any] = {}
```

```python
# apps/anomaly-service/app/main.py (add endpoint)
@app.post("/reasoning/generate", response_model=GenerateReasoningResponse)
def generate_reasoning(req: GenerateReasoningRequest):
    return GenerateReasoningResponse(
        status="success",
        prompt_version="v1",
        response_schema_version="v1",
        summary="Case requires analyst review due to policy mismatch.",
        recommended_action="Validate discount exception approval path.",
        citations=[Citation(source="policy", reference="POL-12", excerpt="Maximum discount 10%")],
        amount_original=12000.0,
        currency_original="USD",
        amount_myr_normalized=56400.0,
        fx_rate_to_myr=4.7,
        metadata={"trace_id": "local-test"},
    )
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd apps/anomaly-service && python -m pytest tests/test_reasoning_generate.py::test_generate_reasoning_success_contract -q`  
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add apps/anomaly-service/requirements.txt apps/anomaly-service/app/schemas.py apps/anomaly-service/app/main.py apps/anomaly-service/tests/test_reasoning_generate.py
git commit -m "feat(anomaly-service): add reasoning generate endpoint with strict schema"
```

---

### Task 3: Add adapters for retrieval, FX (MYR), and Vertex call with typed failures

**Files:**
- Create: `apps/anomaly-service/app/fx.py`
- Create: `apps/anomaly-service/app/retrieval.py`
- Create: `apps/anomaly-service/app/vertex_client.py`
- Modify: `apps/anomaly-service/app/main.py`
- Test: `apps/anomaly-service/tests/test_reasoning_generate.py`

- [ ] **Step 1: Write the failing test**

```python
def test_generate_reasoning_returns_failed_status_on_vertex_error(monkeypatch):
    from fastapi.testclient import TestClient
    from app.main import app
    import app.vertex_client as vc

    def _raise(*args, **kwargs):
        raise RuntimeError("vertex_call_failed")

    monkeypatch.setattr(vc, "generate_grounded_reasoning", _raise)
    client = TestClient(app)
    res = client.post("/reasoning/generate", json={"case_id": "00000000-0000-0000-0000-000000000001"})
    assert res.status_code == 502
    assert res.json()["error_code"] == "vertex_call_failed"
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd apps/anomaly-service && python -m pytest tests/test_reasoning_generate.py::test_generate_reasoning_returns_failed_status_on_vertex_error -q`  
Expected: FAIL until adapter + error mapping exists.

- [ ] **Step 3: Write minimal implementation**

```python
# apps/anomaly-service/app/fx.py
from datetime import datetime, timezone

def normalize_to_myr(amount: float, currency: str) -> dict:
    if currency == "MYR":
        return {"amount_myr_normalized": amount, "fx_rate_to_myr": 1.0, "fx_rate_timestamp": datetime.now(timezone.utc).isoformat()}
    rates = {"USD": 4.70, "SGD": 3.45}
    if currency not in rates:
        raise ValueError("fx_failed")
    return {"amount_myr_normalized": amount * rates[currency], "fx_rate_to_myr": rates[currency], "fx_rate_timestamp": datetime.now(timezone.utc).isoformat()}
```

```python
# apps/anomaly-service/app/vertex_client.py
def generate_grounded_reasoning(case_context: dict, retrieved_context: list) -> dict:
    return {
        "summary": "Grounded summary",
        "recommended_action": "Review and confirm policy exception.",
        "citations": [{"source": "policy", "reference": "POL-12", "excerpt": "Discount above threshold requires approval"}],
        "confidence": 0.86,
    }
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd apps/anomaly-service && python -m pytest tests/test_reasoning_generate.py -q`  
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add apps/anomaly-service/app/fx.py apps/anomaly-service/app/retrieval.py apps/anomaly-service/app/vertex_client.py apps/anomaly-service/app/main.py apps/anomaly-service/tests/test_reasoning_generate.py
git commit -m "feat(anomaly-service): add retrieval/fx/vertex adapters and typed error mapping"
```

---

### Task 4: Persist reasoning output to case-service and expose latest artifact

**Files:**
- Modify: `apps/anomaly-service/app/main.py`
- Modify: `apps/case-service/main.go`
- Test: `apps/anomaly-service/tests/test_reasoning_generate.py`
- Test: `apps/case-service/reasoning_handlers_test.go`

- [ ] **Step 1: Write the failing test**

```python
def test_generate_reasoning_persists_to_case_service(monkeypatch):
    from fastapi.testclient import TestClient
    from app.main import app
    sent = {}

    class DummyResp:
        status_code = 201

    def fake_post(url, json, timeout):
        sent["url"] = url
        sent["json"] = json
        return DummyResp()

    monkeypatch.setattr("app.main.requests.post", fake_post)
    client = TestClient(app)
    res = client.post("/reasoning/generate", json={"case_id": "00000000-0000-0000-0000-000000000001"})
    assert res.status_code == 200
    assert sent["url"].endswith("/cases/00000000-0000-0000-0000-000000000001/reasoning")
    assert "amount_original" in sent["json"]
    assert "currency_original" in sent["json"]
    assert "amount_myr_normalized" in sent["json"]
    assert "fx_rate_to_myr" in sent["json"]
    assert isinstance(sent["json"]["citations"], list)
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd apps/anomaly-service && python -m pytest tests/test_reasoning_generate.py::test_generate_reasoning_persists_to_case_service -q`  
Expected: FAIL.

- [ ] **Step 3: Write minimal implementation**

```python
# apps/anomaly-service/app/main.py (inside /reasoning/generate handler)
payload = {
    "status": "success",
    "model_provider": "vertex-ai",
    "model_name": "gemini-2.5-pro",
    "prompt_version": "v1",
    "response_schema_version": "v1",
    "summary": reasoning["summary"],
    "recommended_action": reasoning["recommended_action"],
    "citations": reasoning["citations"],
    "amount_original": case_amount,
    "currency_original": case_currency,
    "amount_myr_normalized": fx["amount_myr_normalized"],
    "fx_rate_to_myr": fx["fx_rate_to_myr"],
}
requests.post(f"{CASE_SERVICE_URL}/cases/{req.case_id}/reasoning", json=payload, timeout=10)
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd apps/anomaly-service && python -m pytest tests/test_reasoning_generate.py -q`  
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add apps/anomaly-service/app/main.py apps/anomaly-service/tests/test_reasoning_generate.py apps/case-service/main.go apps/case-service/reasoning_handlers_test.go
git commit -m "feat(reasoning): persist generated reasoning artifacts in case-service"
```

---

### Task 5: Configuration, compose wiring, and operator runbook

**Files:**
- Modify: `docker-compose.yml`
- Modify: `README.md`

- [ ] **Step 1: Write the failing test**

```bash
# Failing check: required env vars absent in compose/anomaly-service section
rg "VERTEX_PROJECT_ID|VERTEX_LOCATION|VERTEX_MODEL|REASONING_ENABLED" docker-compose.yml
```

- [ ] **Step 2: Run check to verify it fails**

Run: `rg "VERTEX_PROJECT_ID|VERTEX_LOCATION|VERTEX_MODEL|REASONING_ENABLED" docker-compose.yml`  
Expected: missing keys (non-zero or incomplete output).

- [ ] **Step 3: Write minimal implementation**

```yaml
# docker-compose.yml (anomaly-service env additions)
      REASONING_ENABLED: "true"
      CASE_SERVICE_URL: "http://case-service:8083"
      VERTEX_PROJECT_ID: "${VERTEX_PROJECT_ID}"
      VERTEX_LOCATION: "${VERTEX_LOCATION:-asia-southeast1}"
      VERTEX_MODEL: "${VERTEX_MODEL:-gemini-2.5-pro}"
      BASE_CURRENCY: "MYR"
```

```md
<!-- README.md -->
### Generate case reasoning (backend)

`POST /reasoning/generate` in anomaly-service creates grounded reasoning and stores it in case-service.

Required env vars:
- `VERTEX_PROJECT_ID`
- `VERTEX_LOCATION`
- `VERTEX_MODEL`
- `REASONING_ENABLED=true`
```

- [ ] **Step 4: Run checks to verify they pass**

Run: `rg "VERTEX_PROJECT_ID|VERTEX_LOCATION|VERTEX_MODEL|REASONING_ENABLED" docker-compose.yml`  
Expected: matching lines printed.

- [ ] **Step 5: Commit**

```bash
git add docker-compose.yml README.md
git commit -m "chore(config): wire reasoning env vars and document backend reasoning flow"
```

---

## Final verification checklist

- [ ] Run: `cd apps/case-service && go test ./... -v`  
Expected: all case-service tests PASS.
- [ ] Run: `cd apps/anomaly-service && python -m pytest -q`  
Expected: all anomaly-service tests PASS.
- [ ] Run: `docker compose config`  
Expected: valid compose output.
- [ ] Run: `docker compose up --build`  
Expected: services start; reasoning endpoint reachable.

## Self-review results

- **Spec coverage:** All approved spec areas are mapped: architecture, API, schema, currency model (original + MYR), failure classes, governance config, and backend acceptance checks.
- **Placeholder scan:** Removed ambiguous placeholder-style statements from implementation steps; each code step includes executable snippets/commands.
- **Type consistency:** Field names are consistent across plan tasks (`amount_original`, `currency_original`, `amount_myr_normalized`, `fx_rate_to_myr`, `prompt_version`, `response_schema_version`).
