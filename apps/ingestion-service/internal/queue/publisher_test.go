package queue

import (
	"encoding/json"
	"testing"
)

func TestNewProcessEventTask_ContainsRequiredFields(t *testing.T) {
	t.Parallel()

	task, err := NewProcessEventTask("evt-1", "demo-tenant", "discount_event", "2026-04-29T00:00:00Z", "trace-1")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if task.Type() != TaskTypeProcessEvent {
		t.Fatalf("unexpected task type: %s", task.Type())
	}

	var payload ProcessEventPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if payload.EventID != "evt-1" {
		t.Fatalf("unexpected event_id: %s", payload.EventID)
	}
	if payload.TenantID != "demo-tenant" {
		t.Fatalf("unexpected tenant_id: %s", payload.TenantID)
	}
	if payload.EventType != "discount_event" {
		t.Fatalf("unexpected event_type: %s", payload.EventType)
	}
	if payload.OccurredAt != "2026-04-29T00:00:00Z" {
		t.Fatalf("unexpected occurred_at: %s", payload.OccurredAt)
	}
	if payload.TraceID != "trace-1" {
		t.Fatalf("unexpected trace_id: %s", payload.TraceID)
	}
	if payload.SchemaVersion != "v1" {
		t.Fatalf("unexpected schema_version: %s", payload.SchemaVersion)
	}
}

func TestNewProcessEventTask_Validation(t *testing.T) {
	t.Parallel()

	_, err := NewProcessEventTask("", "demo-tenant", "discount_event", "2026-04-29T00:00:00Z", "trace-1")
	if err == nil {
		t.Fatal("expected validation error")
	}
}
