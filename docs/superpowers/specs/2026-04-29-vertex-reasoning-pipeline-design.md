# RevSentinel AI - Vertex Reasoning Pipeline Design

## 1. Purpose and scope

This spec defines the first implementation milestone for leveraging Google Cloud GenAI Builder credit in RevSentinel AI.

In scope:

- Backend-only reasoning pipeline for existing leak cases
- Vertex AI Gemini call for grounded case explanation and recommended action
- Multi-currency handling with original values plus MYR normalized values
- Persistence and retrieval of reasoning results through service APIs

Out of scope:

- Frontend UX changes for reasoning display
- Full retrieval platform build-out across all policy/document sources
- Autonomous financial action execution

## 2. Architectural decision

Selected approach: **Thin Python reasoning service** in `apps/anomaly-service` (FastAPI), responsible for retrieval, currency normalization, and Gemini invocation in a single service boundary for this milestone.

Rationale:

- Fastest route to demonstrable GenAI value
- Clear integration seam with existing Go services
- Keeps deterministic leak detection unchanged as source-of-truth

## 3. High-level architecture

### 3.1 Services involved

- `case-service` (Go): source of case/evidence context and sink for reasoning artifacts
- `anomaly-service` (Python/FastAPI): reasoning orchestrator for this milestone
- `api-gateway` (Go): optional trigger path for operators/internal API clients

### 3.2 Logical flow

1. Trigger reasoning for a case ID.
2. Reasoning service fetches case details and evidence from case-service.
3. Reasoning service retrieves policy/contract snippets relevant to the case.
4. Reasoning service normalizes monetary values to MYR using FX adapter.
5. Reasoning service calls Vertex AI Gemini with grounded context.
6. Reasoning service validates structured output schema.
7. Reasoning result is persisted to case-service as a case artifact.
8. Case read API exposes latest reasoning result and metadata.

## 4. API design

### 4.1 Reasoning service endpoints

- `POST /reasoning/generate`
  - Request: `case_id`, optional `reasoning_mode`, optional `force_regenerate`
  - Response: reasoning artifact metadata + structured result payload
- `GET /reasoning/healthz`
  - Service health and dependency probe summary

### 4.2 Case-service integration

- Fetch case context:
  - `GET /cases/{id}` (existing)
- Persist reasoning artifact (new internal contract):
  - `POST /cases/{id}/reasoning` (or equivalent artifact endpoint)
- Retrieve latest reasoning:
  - Included in `GET /cases/{id}` response under a dedicated reasoning block

## 5. Data model and schema updates

### 5.1 New artifact table

Proposed table: `case_reasoning`

Core fields:

- `id` (UUID PK)
- `case_id` (UUID FK)
- `tenant_id` (text)
- `status` (`pending | success | failed`)
- `model_provider` (text, e.g., `vertex-ai`)
- `model_name` (text)
- `model_version` (text nullable)
- `prompt_version` (text)
- `response_schema_version` (text)
- `summary` (text)
- `recommended_action` (text)
- `rationale` (jsonb/text)
- `confidence` (double precision nullable)
- `citations` (jsonb)
- `token_input` (int nullable)
- `token_output` (int nullable)
- `latency_ms` (int nullable)
- `trace_id` (text)
- `error_code` (text nullable)
- `error_message` (text nullable)
- `created_at`, `updated_at` (timestamptz)

### 5.2 Multi-currency fields

For reasoning outputs and monetary evidence:

- `amount_original` (double precision nullable)
- `currency_original` (text nullable, ISO code expected)
- `amount_myr_normalized` (double precision nullable)
- `fx_rate_to_myr` (double precision nullable)
- `fx_rate_timestamp` (timestamptz nullable)

Rule: store both original and normalized values whenever amount data is present.

## 6. Prompting and grounding contract

### 6.1 Prompt inputs

- Case core fields: type, severity, title, summary, timestamps
- Deterministic finding evidence from rules path
- Retrieved policy/contract snippets with source references
- Monetary context including original and MYR-normalized values

### 6.2 Prompt output schema (strict JSON)

- `explanation_summary` (string)
- `why_this_is_risky` (string)
- `recommended_next_action` (string)
- `confidence` (number 0-1)
- `citations` (array of objects: source, reference, excerpt)

Validation failure returns explicit `schema_validation_failed`.

## 7. Reliability and failure behavior

### 7.1 Error classes

- `context_fetch_failed`
- `retrieval_failed`
- `fx_failed`
- `vertex_call_failed`
- `schema_validation_failed`

### 7.2 Failure rules

- No silent fallback to ungrounded output.
- Failures persist structured status and error details.
- Retry policy is explicit and bounded; retries are traceable.

## 8. Security, governance, and credit controls

- Restrict reasoning invocation by severity and case type allowlists.
- Tenant-scoped quota and daily call cap.
- Max token and timeout limits per request.
- Persist model ID, prompt version, and token usage for auditability.
- Keep deterministic controls authoritative; reasoning is recommendation only.

## 9. Testing strategy

### 9.1 Automated tests

- Contract tests for reasoning API response shape.
- Integration tests with mocked Vertex and FX adapters.
- Golden tests for prompt template + parser behavior.
- Negative tests for missing retrieval context and invalid currency payloads.

### 9.2 Acceptance criteria

- Backend trigger generates reasoning for an existing case.
- Result is grounded and includes citations.
- Monetary output includes original + MYR-normalized fields.
- Case read API returns latest reasoning artifact.
- Failures are persisted with typed error code and trace metadata.

## 10. Implementation boundaries

This milestone intentionally keeps scope focused to one vertical backend slice:

- Reasoning generation API
- Case-service persistence contract for reasoning artifacts
- Vertex integration with grounded context and MYR normalization

UI rendering and broader retrieval platform expansion are separate follow-up work.
