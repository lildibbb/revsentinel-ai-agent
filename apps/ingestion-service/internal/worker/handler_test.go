package worker

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hibiken/asynq"
	"leakguard.local/ingestion-service/internal/queue"
)

// Test mocks for clients
type mockRulesClient struct {
	called bool
	findings []map[string]interface{}
	err    error
}

func (m *mockRulesClient) Evaluate(ctx context.Context, eventType, eventID string) ([]map[string]interface{}, error) {
	m.called = true
	return m.findings, m.err
}

type mockCaseClient struct {
	called bool
	caseID string
	err    error
}

func (m *mockCaseClient) CreateFromFindings(ctx context.Context, tenantID string, findings []map[string]interface{}) (string, error) {
	m.called = true
	return m.caseID, m.err
}

type mockReasoningClient struct {
	called bool
	err    error
}

func (m *mockReasoningClient) Generate(ctx context.Context, caseID string) error {
	m.called = true
	return m.err
}

// Test handler setup
type testHandler struct {
	*Handler
	rulesClient     *mockRulesClient
	caseClient      *mockCaseClient
	reasoningClient *mockReasoningClient
}

func newTestHandlerWithMocks() *testHandler {
	rules := &mockRulesClient{}
	cases := &mockCaseClient{caseID: "case-123"}
	reasoning := &mockReasoningClient{}

	return &testHandler{
		Handler: &Handler{
			rules:     rules,
			cases:     cases,
			reasoning: reasoning,
		},
		rulesClient:     rules,
		caseClient:      cases,
		reasoningClient: reasoning,
	}
}

func (t *testHandler) rulesClientCalled() bool {
	return t.rulesClient.called
}

func (t *testHandler) caseClientCalled() bool {
	return t.caseClient.called
}

func (t *testHandler) reasoningClientCalled() bool {
	return t.reasoningClient.called
}

func mustTask(eventID, tenantID, eventType, occurredAt, traceID string) *asynq.Task {
	payload := queue.ProcessEventPayload{
		EventID:       eventID,
		TenantID:      tenantID,
		EventType:     eventType,
		OccurredAt:    occurredAt,
		TraceID:       traceID,
		SchemaVersion: "v1",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	return asynq.NewTask(queue.TaskTypeProcessEvent, body)
}

// Tests
func TestHandleProcessEvent_RulesFindings_CreateCaseAndReasoning(t *testing.T) {
	h := newTestHandlerWithMocks()
	h.rulesClient.findings = []map[string]interface{}{
		{"severity": "high", "type": "anomaly"},
	}
	
	task := mustTask("evt-1", "demo-tenant", "discount_event", "2026-04-29T00:00:00Z", "trace-1")
	err := h.HandleProcessEvent(context.Background(), task)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !h.caseClientCalled() || !h.reasoningClientCalled() {
		t.Fatalf("expected case and reasoning calls")
	}
}

func TestHandleProcessEvent_NoFindings_NoCase(t *testing.T) {
	h := newTestHandlerWithMocks()
	h.rulesClient.findings = []map[string]interface{}{}
	
	task := mustTask("evt-2", "demo-tenant", "discount_event", "2026-04-29T00:00:00Z", "trace-2")
	err := h.HandleProcessEvent(context.Background(), task)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if h.caseClientCalled() {
		t.Fatalf("expected no case call when no findings")
	}
}

func TestHandleProcessEvent_RulesError(t *testing.T) {
	h := newTestHandlerWithMocks()
	h.rulesClient.err = context.DeadlineExceeded
	
	task := mustTask("evt-3", "demo-tenant", "discount_event", "2026-04-29T00:00:00Z", "trace-3")
	err := h.HandleProcessEvent(context.Background(), task)
	if err == nil {
		t.Fatalf("expected error from rules call")
	}
	if err.Error() != "rules_call_failed: context deadline exceeded" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestHandleProcessEvent_CaseError(t *testing.T) {
	h := newTestHandlerWithMocks()
	h.rulesClient.findings = []map[string]interface{}{
		{"severity": "high", "type": "anomaly"},
	}
	h.caseClient.err = context.DeadlineExceeded
	
	task := mustTask("evt-4", "demo-tenant", "discount_event", "2026-04-29T00:00:00Z", "trace-4")
	err := h.HandleProcessEvent(context.Background(), task)
	if err == nil {
		t.Fatalf("expected error from case call")
	}
	if err.Error() != "case_persist_failed: context deadline exceeded" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestHandleProcessEvent_ReasoningError(t *testing.T) {
	h := newTestHandlerWithMocks()
	h.rulesClient.findings = []map[string]interface{}{
		{"severity": "high", "type": "anomaly"},
	}
	h.reasoningClient.err = context.DeadlineExceeded
	
	task := mustTask("evt-5", "demo-tenant", "discount_event", "2026-04-29T00:00:00Z", "trace-5")
	err := h.HandleProcessEvent(context.Background(), task)
	if err == nil {
		t.Fatalf("expected error from reasoning call")
	}
	if err.Error() != "reasoning_call_failed: context deadline exceeded" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestHandleProcessEvent_InvalidPayload(t *testing.T) {
	h := newTestHandlerWithMocks()
	
	task := asynq.NewTask(queue.TaskTypeProcessEvent, []byte("invalid json"))
	err := h.HandleProcessEvent(context.Background(), task)
	if err == nil {
		t.Fatalf("expected error from invalid payload")
	}
	errMsg := err.Error()
	if len(errMsg) < 18 || errMsg[:18] != "event_load_failed:" {
		t.Fatalf("unexpected error message: %v", err)
	}
}
