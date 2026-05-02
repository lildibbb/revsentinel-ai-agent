CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS ingested_event (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id TEXT NOT NULL,
  event_type TEXT NOT NULL,
  occurred_at TIMESTAMPTZ NOT NULL,
  source JSONB NULL,
  payload JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS leak_case (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id TEXT NOT NULL,
  case_type TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'detected',
  severity TEXT NOT NULL,
  title TEXT NOT NULL,
  summary TEXT NOT NULL,
  exposure_amount DOUBLE PRECISION NULL,
  currency TEXT NOT NULL DEFAULT 'USD',
  confidence DOUBLE PRECISION NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS case_evidence (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  case_id UUID NOT NULL REFERENCES leak_case(id) ON DELETE CASCADE,
  kind TEXT NOT NULL,
  data JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

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

CREATE INDEX IF NOT EXISTS idx_ingested_event_tenant_time ON ingested_event(tenant_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_leak_case_tenant_time ON leak_case(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_case_evidence_case_id ON case_evidence(case_id);
CREATE INDEX IF NOT EXISTS idx_case_reasoning_case_created ON case_reasoning(case_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_case_reasoning_tenant_created ON case_reasoning(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_case_reasoning_trace_id ON case_reasoning(trace_id);
CREATE INDEX IF NOT EXISTS idx_event_processing_ledger_event ON event_processing_ledger(event_id);
