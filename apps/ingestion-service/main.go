package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"leakguard.local/ingestion-service/internal/queue"
)

type IngestRequest struct {
	EventType  string         `json:"event_type"`
	OccurredAt *time.Time     `json:"occurred_at,omitempty"`
	Source     map[string]any `json:"source,omitempty"`
	Payload    map[string]any `json:"payload"`
}

type IngestResponse struct {
	EventID          string    `json:"event_id"`
	TenantID         string    `json:"tenant_id"`
	EventType        string    `json:"event_type"`
	OccurredAt       time.Time `json:"occurred_at"`
	ProcessingStatus string    `json:"processing_status"`
	QueueTaskID      string    `json:"queue_task_id,omitempty"`
}

type Finding struct {
	CaseType       string   `json:"case_type"`
	Severity       string   `json:"severity"`
	Title          string   `json:"title"`
	Summary        string   `json:"summary"`
	ExposureAmount *float64 `json:"exposure_amount,omitempty"`
	Currency       string   `json:"currency,omitempty"`
	Confidence     *float64 `json:"confidence,omitempty"`
}

const (
	enqueueMetadataKey     = "_ingestion"
	enqueueStatusPending   = "enqueue_pending"
	enqueueStatusQueued    = "queued"
	enqueueStatusFailed    = "enqueue_failed"
	defaultEnqueueTimeout  = 5 * time.Second
	defaultDBUpdateTimeout = 3 * time.Second
)

func main() {
	port := env("PORT", "8081")
	dbURL := env("DATABASE_URL", "")
	redisAddr := env("REDIS_ADDR", "redis:6379")
	queueName := env("ASYNQ_QUEUE_DEFAULT", "default")
	tenantID := env("TENANT_ID", "demo-tenant")
	enqueueTimeout := durationEnv("ENQUEUE_TIMEOUT", defaultEnqueueTimeout)
	dbUpdateTimeout := durationEnv("DB_UPDATE_TIMEOUT", defaultDBUpdateTimeout)
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	defer asynqClient.Close()
	publisher := queue.NewPublisher(asynqClient, queueName)

	r := chi.NewRouter()

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	r.Post("/ingest", func(w http.ResponseWriter, r *http.Request) {
		var req IngestRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if req.EventType == "" || req.Payload == nil {
			writeError(w, http.StatusBadRequest, errors.New("event_type and payload are required"))
			return
		}

		occurredAt := time.Now().UTC()
		if req.OccurredAt != nil {
			occurredAt = req.OccurredAt.UTC()
		}

		payloadBytes, err := json.Marshal(req.Payload)
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %w", err))
			return
		}
		payloadJSON := string(payloadBytes)

		source := map[string]any{}
		for k, v := range req.Source {
			source[k] = v
		}
		source[enqueueMetadataKey] = map[string]any{
			"enqueue_status": enqueueStatusPending,
			"task_type":      queue.TaskTypeProcessEvent,
		}
		sourceBytes, err := json.Marshal(source)
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf("invalid source: %w", err))
			return
		}
		sourceJSON := string(sourceBytes)

		var eventID string
		err = pool.QueryRow(r.Context(), `
			INSERT INTO ingested_event (tenant_id, event_type, occurred_at, source, payload)
			VALUES ($1,$2,$3,$4::jsonb,$5::jsonb)
			RETURNING id
		`, tenantID, req.EventType, occurredAt, sourceJSON, payloadJSON).Scan(&eventID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			traceID = eventID
		}
		enqueueCtx, cancelEnqueue := context.WithTimeout(context.Background(), enqueueTimeout)
		taskID, err := publisher.EnqueueProcessEvent(enqueueCtx, eventID, tenantID, req.EventType, occurredAt.Format(time.RFC3339Nano), traceID)
		cancelEnqueue()
		if err != nil {
			failureMeta := map[string]any{
				"enqueue_status": enqueueStatusFailed,
				"task_type":      queue.TaskTypeProcessEvent,
				"last_error":     err.Error(),
			}
			persistEnqueueMetadata(pool, eventID, failureMeta, dbUpdateTimeout)
			writeEnqueueError(w, eventID, enqueueStatusFailed, err)
			return
		}
		successMeta := map[string]any{
			"enqueue_status": enqueueStatusQueued,
			"task_type":      queue.TaskTypeProcessEvent,
		}
		if taskID != "" {
			successMeta["task_id"] = taskID
		}
		persistEnqueueMetadata(pool, eventID, successMeta, dbUpdateTimeout)

		resp := IngestResponse{
			EventID:          eventID,
			TenantID:         tenantID,
			EventType:        req.EventType,
			OccurredAt:       occurredAt,
			ProcessingStatus: enqueueStatusQueued,
			QueueTaskID:      taskID,
		}
		writeJSON(w, http.StatusAccepted, resp)
	})

	addr := ":" + port
	log.Printf("ingestion-service listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

func env(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func durationEnv(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		return def
	}
	return d
}

func decodeJSON(r *http.Request, out any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(out)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{"error": err.Error()})
}

func writeEnqueueError(w http.ResponseWriter, eventID, status string, err error) {
	writeJSON(w, http.StatusBadGateway, map[string]any{
		"error":             fmt.Sprintf("enqueue process_event.v1 failed: %v", err),
		"event_id":          eventID,
		"processing_status": status,
	})
}

func persistEnqueueMetadata(pool *pgxpool.Pool, eventID string, metadata map[string]any, timeout time.Duration) {
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		log.Printf("failed to marshal enqueue metadata for event %s: %v", eventID, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if _, err := pool.Exec(ctx, `
		UPDATE ingested_event
		SET source = jsonb_set(COALESCE(source, '{}'::jsonb), '{_ingestion}', $2::jsonb, true)
		WHERE id = $1
	`, eventID, string(metadataBytes)); err != nil {
		log.Printf("failed to persist enqueue metadata for event %s: %v", eventID, err)
	}
}
