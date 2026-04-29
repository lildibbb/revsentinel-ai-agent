# LeakGuard AI PRD

## Product overview

LeakGuard AI is an enterprise-grade revenue leakage detection agent that continuously watches transactions, invoices, discounts, credits, contracts, and billing events to detect silent profit loss before it compounds.[cite:226][cite:232][cite:234] The product is designed as an always-on agent, not a passive dashboard: it detects anomalies, validates them against business rules and policy context, quantifies impact, creates investigation cases, and recommends corrective action.[cite:221][cite:225][cite:227]

## Problem

Revenue leakage is the class of hidden losses caused by pricing errors, unauthorized discounts, missed billable events, invoice timing failures, contract mismatches, excessive credits, and manual overrides that quietly reduce realized revenue.[cite:226][cite:232][cite:234] Organizations often discover these issues too late because billing, sales ops, finance, and service delivery data live in separate systems, which means no single workflow continuously watches for leakage across the full revenue lifecycle.[cite:221][cite:223][cite:232] Research-backed commercial deployments show that revenue leakage detection can produce major financial impact, including one reported case that found $12.1 million in leakage in the first year with 302% ROI, which makes this category highly credible for a business-facing portfolio product.[cite:232]

## Product vision

LeakGuard AI acts like a revenue operations sentinel for organizations that want stronger profit control without scaling manual audit work.[cite:232][cite:233] Its purpose is to move teams from reactive audit and spreadsheet checks to continuous, autonomous revenue assurance where suspicious patterns are surfaced with evidence, business impact, and next-step recommendations.[cite:221][cite:225][cite:234]

## Goals

- Detect silent revenue leakage before month-end close or quarterly review.[cite:232][cite:234]
- Reduce manual audit effort by automating cross-system checks and anomaly triage.[cite:221][cite:225]
- Quantify leakage in business terms such as recovered revenue, prevented loss, and open financial exposure.[cite:232][cite:234]
- Provide a full audit trail for every agent decision, case state transition, and recommended action.[cite:221][cite:228]
- Demonstrate an enterprise-grade AI product that combines deterministic workflows with contextual LLM reasoning.[cite:221][cite:227][cite:230]

## Non-goals

- Full ERP replacement.
- General financial forecasting.
- Fully autonomous posting of accounting entries without human approval.
- Real-time payment orchestration or treasury management.

## Target customers

The product is best positioned for mid-market and enterprise organizations with complex pricing, approval, or billing flows, especially SaaS companies, services businesses, B2B distributors, telecom-like billing environments, and multi-team revenue operations functions.[cite:226][cite:232][cite:234] The internal users are typically finance analysts, revenue operations managers, billing teams, controllers, internal audit, and CFO-level stakeholders who care about hidden margin erosion and control quality.[cite:232][cite:234]

## Core users

| User | Primary need | What LeakGuard provides |
|---|---|---|
| Revenue operations manager | Detect leakage patterns across contracts, discounts, and billing | Central leak feed, case prioritization, root-cause reasoning.[cite:226][cite:232] |
| Finance analyst | Validate suspicious cases quickly | Evidence-backed cases with impacted records and policy references.[cite:225][cite:234] |
| Controller / CFO | Understand monetary exposure and controls effectiveness | Impact dashboard, prevented loss totals, trend summaries.[cite:232][cite:234] |
| Admin / platform owner | Tune rules and system behavior | Policy configuration, thresholds, role controls, audit logs.[cite:221][cite:228] |

## Core problem scenarios

### Pricing leakage

A customer contract allows a maximum 10% discount, but a rep or billing operator applies 22%, which creates a margin leak that may not be noticed until after invoicing.[cite:226][cite:234] LeakGuard should detect the variance, verify the authorized contract terms, estimate annualized impact, and open a case automatically.[cite:226][cite:232]

### Unbilled service delivery

Professional services or usage events are delivered, but the corresponding invoice is never generated because the billing event was missed or failed in the handoff.[cite:234] LeakGuard should correlate service logs, entitlement usage, and invoice records to detect completed but unbilled work.[cite:221][cite:234]

### Credit and adjustment abuse

Frequent manual credits, write-offs, or refunds are issued for the same customer segment, agent, or product line, creating a slow but meaningful revenue drain.[cite:229][cite:232][cite:234] LeakGuard should identify repeated adjustment patterns, cluster related activity, and explain why the pattern is suspicious relative to historical norms.[cite:229][cite:232]

### Billing timing failures

Invoices are generated too late, outside SLA windows, or not aligned to contract terms, causing cash-flow delays and possible revenue recognition issues.[cite:234] LeakGuard should monitor expected billing cadence and create a warning before the issue becomes a close-period problem.[cite:226][cite:234]

## Product positioning

LeakGuard AI is not a generic anomaly dashboard and not a pure rules engine.[cite:221][cite:225] The differentiator is a hybrid model: deterministic rules catch known failure patterns, statistical anomaly detection surfaces unfamiliar deviations, and Gemini via Vertex AI explains, contextualizes, and recommends action with policy grounding.[cite:227][cite:230][cite:234]

## Solution summary

The system ingests normalized commercial and financial events, runs a scheduled and event-driven watch loop, evaluates each event through rule checks and anomaly models, enriches suspicious findings with contract and policy context, and converts validated findings into trackable cases with business impact estimates.[cite:221][cite:225][cite:227] Each case contains the leak type, evidence, confidence level, affected records, exposure estimate, recommended next action, and audit trace so teams can resolve issues faster and more consistently.[cite:228][cite:232][cite:234]

## Key product capabilities

### 1. Revenue watch engine

The platform continuously watches invoices, quotes, contracts, service logs, discount approvals, billing events, credit notes, and manual overrides for evidence of leakage.[cite:226][cite:232][cite:234] It supports both scheduled scans and event-driven checks so the product feels like an always-on operational agent rather than a monthly reporting tool.[cite:221][cite:225]

### 2. Hybrid detection engine

The detection engine combines deterministic policy checks with statistical anomaly scoring because hybrid architectures are better suited to enterprise agent systems that need both precision and adaptability.[cite:221][cite:225][cite:229] Rules catch explicit violations such as discount thresholds or missing invoice mappings, while anomaly models identify unexpected adjustment behavior, pricing drift, or unusual revenue patterns.[cite:229][cite:234]

### 3. Policy-aware reasoning

Gemini is used to interpret business context, summarize why a case matters, compare the finding to pricing policy or contract language, and recommend the most likely next action.[cite:227][cite:230][cite:246] This reasoning layer should never be the source of truth for deterministic controls; instead, it enhances explainability and workflow quality on top of validated signals.[cite:221][cite:246]

### 4. Case management

Each confirmed or probable leak becomes a case with a lifecycle such as `detected`, `investigating`, `confirmed`, `actioned`, `resolved`, or `dismissed`.[cite:221][cite:228] This turns the product from an analytics surface into an operational system that owns follow-through and supports management reporting.[cite:225][cite:233]

### 5. Auditability and trust

Enterprise agents must be explainable and governable, so every detection path should be recorded: source event, rule hit, anomaly score, retrieval context, LLM summary, case action, reviewer decision, and timestamp.[cite:221][cite:225][cite:228] This is essential to make the product credible for finance and audit stakeholders.[cite:228][cite:234]

## Success metrics

| Metric | Definition |
|---|---|
| Leakage value detected | Total monetary value of all confirmed and probable leaks in a time period. |
| Prevented loss | Estimated value prevented after timely corrective action. |
| Mean time to detect | Time from leak-causing event to case creation. |
| Mean time to resolution | Time from case creation to case closure. |
| False positive rate | Percentage of detected cases dismissed by reviewers. |
| Policy coverage | Percentage of pricing and billing control rules represented in the system. |
| Analyst time saved | Estimated reduction in manual audit/review effort. |

## Functional requirements

### Event ingestion

- Ingest transaction, invoice, credit note, quote, usage, contract, service delivery, and approval events.
- Support batch import and streaming/event push.
- Normalize source records into a common ledger-style event model.
- Preserve source metadata for traceability.

### Rules engine

- Support configurable rules for discount thresholds, missing invoices, contract price mismatches, duplicate credits, late billing, expired entitlements, unauthorized overrides, and invoice cadence gaps.
- Allow per-tenant rule customization.
- Version every rule and support effective dates.

### Anomaly detection

- Score unusual behavior in credits, price overrides, invoice timing, adjustment frequency, and revenue variance.
- Store features and model version used for each case.
- Allow sensitivity tuning by organization or business unit.

### Retrieval and reasoning

- Retrieve contract terms, pricing policy, approval policy, and exception documentation for context.
- Use Gemini to create case summaries, recommended actions, and reviewer-friendly explanations.[cite:227][cite:246]
- Ground LLM outputs in retrieved text and source references.

### Case workflow

- Auto-create cases above severity threshold.
- Route cases to the correct queue based on type and organization unit.
- Support reviewer comments, assignment, state changes, and dismissal reasons.
- Generate action summaries and downloadable case evidence packs.

### Dashboards

- Leak feed by severity, type, status, and business unit.
- Trend dashboards by period and category.
- Impact leaderboard by product, customer segment, or sales region.
- Resolution metrics and false-positive analytics.

### Notifications

- Slack, email, or webhook notifications for high-severity cases.
- Digest summaries for finance/revops managers.
- Escalation notifications for aging cases.

## Non-functional requirements

### Performance

- Event ingestion should support bursty data arrival and background processing.[cite:235][cite:244]
- High-severity rule matches should become visible in the UI within a short operational window.
- Rule evaluation must degrade gracefully even if LLM services are unavailable.[cite:221][cite:246]

### Reliability

- Case creation must be idempotent.
- Retries and dead-letter handling are required for ingestion and scoring pipelines.
- Core rule checks must continue operating even when the reasoning layer is degraded.

### Security

- Tenant data isolation is mandatory.
- Role-based access control for analyst, manager, admin, and executive views.
- Encryption in transit and at rest.
- Full audit log for all user and agent actions.

### Explainability

- Every case must show what triggered it, what evidence was used, and why the action was recommended.[cite:221][cite:228]
- LLM-generated content must be clearly labeled as recommendation, not authoritative accounting action.[cite:246]

## System architecture

The recommended architecture is a microservice system with Go as the primary systems/runtime language and Python as the specialized analytics service language, because Go offers stronger memory efficiency and concurrency for always-on services while Python remains stronger for anomaly detection and analytics experimentation.[cite:235][cite:239][cite:244] The AI reasoning layer is provided through Vertex AI / Gemini, which preserves the value of the user's GenAI credit while keeping deterministic control logic outside the LLM.[cite:227][cite:241][cite:246]

### Service map

| Service | Language | Responsibility |
|---|---|---|
| API Gateway / BFF | Go | Auth, tenant routing, API composition, WebSocket push. |
| Ingestion Service | Go | Event intake, normalization, validation, deduplication. |
| Rules Engine Service | Go | Deterministic leakage checks and policy enforcement. |
| Scheduler / Orchestrator | Go | Watch cycles, scan triggers, workflow coordination. |
| Case Management Service | Go | Case lifecycle, assignments, comments, states, audit trail. |
| Notification Service | Go | Slack/email/webhook notifications and digests. |
| Anomaly Detection Service | Python | Feature engineering, anomaly models, scoring. |
| AI Reasoning Service | Python or thin Go client | Calls Vertex AI / Gemini for grounded summaries and recommendations.[cite:227][cite:246] |
| Retrieval / Policy Service | Go | Policy document indexing, metadata lookup, retrieval orchestration. |
| Frontend Dashboard | React + Vite | Analyst and executive UI. |

## Why microservices

A microservice architecture is justified here because the system combines different workload types: low-latency APIs, streaming or queue-driven ingestion, deterministic control execution, long-running scoring jobs, and external AI calls with different scaling characteristics.[cite:221][cite:228] It also creates a stronger enterprise narrative for the portfolio project by showing clear bounded contexts, service contracts, queue-based resilience, and language specialization rather than a monolithic AI demo.[cite:221][cite:233]

## Why Go + Python

### Why Go for core services

Go is a better choice for the always-on operational core because it is memory-efficient, deploys cleanly as a single binary, handles concurrency well, and fits event-driven services, APIs, schedulers, and queue workers very naturally.[cite:235][cite:239][cite:244] It also gives the project stronger enterprise backend credibility than a pure Python app when the system has to behave like an operational platform instead of a notebook-backed prototype.[cite:239][cite:247]

### Why Python for analytics and AI integration

Python remains useful where its ecosystem is strongest: feature engineering, anomaly detection, model experimentation, and certain AI integration workflows.[cite:238][cite:240][cite:246] Rather than making Python the whole platform, the recommended design isolates it in a dedicated analytics service so the product keeps Go's operational advantages without giving up Python's modeling ecosystem.[cite:239][cite:245]

## Recommended tech stack

### Frontend

- React + Vite
- Tailwind CSS
- TanStack Query
- WebSocket client for real-time case updates
- Recharts or Apache ECharts for impact and trend charts

### Backend

- Go 1.23+
- HTTP framework: Fiber or Chi
- gRPC for service-to-service calls where useful
- PostgreSQL for transactional persistence
- Redis for cache, ephemeral agent state, distributed locks
- Asynq or Temporal for background jobs and workflow orchestration
- NATS or Kafka optional for event streaming if scope expands

### Analytics / AI

- Python 3.12
- FastAPI for analytics service wrapper
- pandas / polars for feature engineering
- scikit-learn / PyOD for anomaly models
- Vertex AI Gemini for contextual reasoning, summaries, and action recommendations.[cite:227][cite:246]
- Vertex AI Search or pgvector for retrieval over policy and contract text.[cite:230][cite:246]

### Infra / DevOps

- Docker / Docker Compose for local orchestration
- Google Cloud Run for services with moderate traffic and simple ops.[cite:227][cite:241]
- GKE optional if demonstrating full container orchestration is part of the portfolio target
- Google Cloud SQL for PostgreSQL
- Memorystore for Redis if fully on GCP
- GitHub Actions for CI/CD
- OpenTelemetry + Prometheus / Grafana or GCP-native logging/monitoring

## Deployment topology

A practical portfolio deployment can run the Go API gateway, Go rules service, Go case service, Python anomaly service, PostgreSQL, and Redis in containers with internal networking, while the reasoning layer calls Vertex AI externally.[cite:227][cite:241] This shows realistic service separation without forcing unnecessary cloud complexity too early.[cite:221][cite:225]

## Proposed data model

### Core entities

- `tenant`
- `customer`
- `contract`
- `pricing_rule`
- `approval_policy`
- `invoice`
- `invoice_line`
- `usage_event`
- `service_delivery_event`
- `discount_event`
- `credit_note`
- `adjustment_event`
- `leak_case`
- `case_evidence`
- `case_action`
- `audit_log`
- `model_run`
- `rule_version`

### Example leakage features

- Contract price vs invoice price delta
- Discount percentage vs allowed threshold
- Days from service delivery to invoice creation
- Credit note frequency by rep / customer / product
- Refund-to-revenue ratio by segment
- Manual override count per billing cycle
- Invoice cadence drift vs expected schedule
- Repeated revenue reversal patterns

## Agent loop design

The agent should run in two modes: event-driven and scheduled.[cite:221][cite:225] Event-driven mode handles newly ingested records and high-confidence deterministic violations immediately, while scheduled mode performs broader scans for timing gaps, cumulative patterns, and slow-moving anomalies.[cite:221][cite:225][cite:234]

### Loop steps

1. Ingest new or changed records.
2. Normalize and enrich with tenant, contract, and policy metadata.
3. Run deterministic rules.
4. Send relevant aggregates to anomaly scoring.
5. Retrieve supporting policy or contract context.
6. Ask Gemini to summarize the finding and recommend next action with grounded context.[cite:227][cite:246]
7. Create or update case.
8. Notify stakeholders if severity threshold is met.
9. Track human feedback and final resolution.
10. Feed reviewer outcomes back into tuning and thresholding.

## Detection strategy

### Deterministic rules first

The platform should treat deterministic rules as the first line of defense because finance control environments require explicit, explainable checks for known policy boundaries.[cite:221][cite:228] This includes contract max-discount breaches, billing without valid entitlement, unbilled completed work, and approval-path violations.[cite:226][cite:234]

### Anomaly scoring second

Anomaly models should be used to detect unusual but not explicitly codified behavior, such as a sudden spike in credits by a specific region or a subtle shift in invoice timing.[cite:229][cite:234] The point is not full ML autonomy; the point is to prioritize patterns humans would likely miss.[cite:229][cite:232]

### LLM reasoning third

The LLM should be used after evidence exists, not before.[cite:246] Its role is to generate readable explanations, compare evidence against policy text, suggest likely root causes, and draft recommended actions, which keeps the system governable and enterprise-safe.[cite:221][cite:227][cite:246]

## User workflows

### Workflow 1: Discount abuse detection

1. Discount event arrives from CRM or billing export.
2. Ingestion service normalizes it.
3. Rules engine compares discount to contract and approval policy.
4. If breached, case is created immediately.
5. Gemini generates explanation and recommended remediation.
6. Analyst reviews and confirms.
7. Notification is sent to revops manager.
8. Case is resolved after corrective invoice or approval exception.

### Workflow 2: Unbilled work detection

1. Service delivery events arrive from operational source.
2. Scheduled scan looks for missing invoice mapping after grace window.
3. Anomaly service checks if the lag is unusual relative to historical pattern.
4. Case is created with estimated value at risk.
5. Analyst reviews missing invoice evidence.
6. Finance triggers billing recovery action.

### Workflow 3: Adjustment anomaly cluster

1. Credit notes and write-offs accumulate through the week.
2. Scheduled scan aggregates by rep, customer segment, and product line.
3. Python service scores anomaly cluster.
4. Gemini explains why the cluster deviates from normal and what to inspect first.
5. Manager receives a digest of open high-impact cases.

## Human-in-the-loop design

Enterprise grade means the platform should support autonomy with control, not full unsupervised execution.[cite:221][cite:228] High-confidence low-risk actions can create notifications or draft artifacts automatically, but accounting-impacting actions should remain subject to review or approval.[cite:221][cite:246]

### Action classes

- **Auto**: notify, summarize, categorize, create case, attach evidence.
- **Suggested**: draft email, draft correction workflow, propose severity or owner.
- **Approval required**: post adjustment, dismiss high-severity case, change policy rule, trigger financial reversal.

## Security and governance

- RBAC by tenant, business unit, and role.
- Immutable audit log for all agent and human actions.[cite:228]
- Model version and prompt version tracking for every AI-generated recommendation.
- Configurable data retention and masking for sensitive customer records.
- Rate limits and outbound policy for LLM calls.
- Environment separation across dev, staging, and prod.

## Reliability and ops

- Queue-backed processing with retries and dead-letter queues.
- Idempotent case creation and event handling.
- Health checks for all services.
- Circuit breaker or fallback mode when Vertex AI is degraded.
- Reconciliation jobs to ensure no ingestion gaps.
- Structured logging with trace IDs across services.

## MVP scope

The MVP should focus on 3 leakage patterns only so the product feels sharp instead of bloated:

1. Discount threshold violations.
2. Unbilled service delivery.
3. Suspicious credit / adjustment clusters.

These three are enough to prove the product thesis and demonstrate hybrid detection, case management, and AI explanation end to end.[cite:226][cite:232][cite:234]

## Phase roadmap

### Phase 1 — Foundations

- Define event schema.
- Build synthetic tenant data generator.
- Implement PostgreSQL schema.
- Create ingestion pipeline.
- Build base RBAC and auth.

### Phase 2 — Deterministic controls

- Implement discount, unbilled work, and adjustment rules.
- Build case creation workflow.
- Add audit trail.
- Build leak feed UI.

### Phase 3 — Analytics and AI

- Add Python anomaly scoring service.
- Add Vertex AI grounded reasoning.
- Add policy/contract retrieval.
- Add executive impact dashboard.

### Phase 4 — Enterprise polish

- Add notification workflows.
- Add case assignment and approvals.
- Add digest summaries.
- Add observability and resiliency hardening.
- Add multi-tenant isolation and configuration UI.

## Build steps

### Step 1: Domain design

Model the revenue lifecycle, identify explicit leakage patterns, define entities, and write the case taxonomy.

### Step 2: Synthetic data generation

Generate realistic tenants, customers, contracts, pricing rules, invoices, discounts, usage events, and adjustment behavior so the demo world is coherent and repeatable.[cite:201][cite:194][cite:195]

### Step 3: Core backend services

Implement Go services for ingestion, rules, orchestration, case management, and notifications.

### Step 4: Analytics microservice

Implement the Python anomaly service with clear feature contracts and service APIs.

### Step 5: AI reasoning and retrieval

Integrate Gemini and retrieval so every high-value case comes with grounded explanation and recommended action.[cite:227][cite:246]

### Step 6: Frontend

Build analyst and executive dashboards that show a live leak feed, case drill-down, evidence panel, and impact summaries.

### Step 7: Ops and hardening

Add tracing, metrics, retries, idempotency, and failure fallbacks.

### Step 8: Portfolio packaging

Create a polished case study that explains the business problem, product design, architecture, sample cases, and estimated ROI.

## Suggested repository structure

```text
leakguard-ai/
  apps/
    web-dashboard/
    api-gateway/
    ingestion-service/
    rules-service/
    case-service/
    notification-service/
    scheduler-service/
    anomaly-service/
    retrieval-service/
  packages/
    proto/
    shared-schemas/
    ui-kit/
  infra/
    docker/
    k8s/
    terraform/
  docs/
    prd.md
    architecture.md
    api-contracts/
    prompts/
    data-model/
```

## Why this project is portfolio-worthy

This project stands out because it is clearly tied to profit protection, not generic AI novelty.[cite:232][cite:234] It also demonstrates enterprise thinking across architecture, service boundaries, governance, event-driven processing, AI grounding, and human-in-the-loop controls, which makes it much more credible than a simple prompt wrapper or chatbot app.[cite:221][cite:225][cite:228]

## Final recommendation

LeakGuard AI should be built as a Go-first, Python-assisted, Vertex-powered revenue assurance platform that treats AI as a contextual reasoning layer on top of a deterministic control core.[cite:227][cite:235][cite:246] That architecture gives the strongest balance of enterprise credibility, technical depth, operational realism, and portfolio impact for a candidate who wants to look valuable to recruiters beyond pure software implementation.[cite:85][cite:239][cite:244]
