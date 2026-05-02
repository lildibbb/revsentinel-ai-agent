package queue

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/hibiken/asynq"
)

const TaskTypeProcessEvent = "process_event.v1"

type ProcessEventPayload struct {
	EventID       string `json:"event_id"`
	TenantID      string `json:"tenant_id"`
	EventType     string `json:"event_type"`
	OccurredAt    string `json:"occurred_at"`
	TraceID       string `json:"trace_id"`
	SchemaVersion string `json:"schema_version"`
}

type Publisher struct {
	client    *asynq.Client
	queueName string
}

func NewPublisher(client *asynq.Client, queueName string) *Publisher {
	if queueName == "" {
		queueName = "default"
	}
	return &Publisher{
		client:    client,
		queueName: queueName,
	}
}

func NewProcessEventTask(eventID, tenantID, eventType, occurredAt, traceID string) (*asynq.Task, error) {
	if eventID == "" {
		return nil, errors.New("event_id is required")
	}
	if tenantID == "" {
		return nil, errors.New("tenant_id is required")
	}
	if eventType == "" {
		return nil, errors.New("event_type is required")
	}
	if occurredAt == "" {
		return nil, errors.New("occurred_at is required")
	}

	payload := ProcessEventPayload{
		EventID:       eventID,
		TenantID:      tenantID,
		EventType:     eventType,
		OccurredAt:    occurredAt,
		TraceID:       traceID,
		SchemaVersion: "v1",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TaskTypeProcessEvent, body), nil
}

func (p *Publisher) EnqueueProcessEvent(ctx context.Context, eventID, tenantID, eventType, occurredAt, traceID string) error {
	if p == nil || p.client == nil {
		return errors.New("publisher client is not configured")
	}

	task, err := NewProcessEventTask(eventID, tenantID, eventType, occurredAt, traceID)
	if err != nil {
		return err
	}
	_, err = p.client.EnqueueContext(ctx, task, asynq.Queue(p.queueName))
	return err
}
