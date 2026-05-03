package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"leakguard.local/ingestion-service/internal/domain"
	"leakguard.local/ingestion-service/internal/queue"
)

// Client interfaces for dependency injection
type RulesClient interface {
	Evaluate(ctx context.Context, eventType, eventID string) ([]map[string]interface{}, error)
}

type CaseClient interface {
	CreateFromFindings(ctx context.Context, tenantID string, findings []map[string]interface{}) (string, error)
}

type ReasoningClient interface {
	Generate(ctx context.Context, caseID string) error
}

type Handler struct {
	rules     RulesClient
	cases     CaseClient
	reasoning ReasoningClient
}

func NewHandler(rules RulesClient, cases CaseClient, reasoning ReasoningClient) *Handler {
	return &Handler{
		rules:     rules,
		cases:     cases,
		reasoning: reasoning,
	}
}

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

	if domain.ShouldTriggerReasoning(findings) {
		if err := h.reasoning.Generate(ctx, caseID); err != nil {
			return fmt.Errorf("reasoning_call_failed: %w", err)
		}
	}

	return nil
}
