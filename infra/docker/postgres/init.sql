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

CREATE INDEX IF NOT EXISTS idx_ingested_event_tenant_time ON ingested_event(tenant_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_leak_case_tenant_time ON leak_case(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_case_evidence_case_id ON case_evidence(case_id);
