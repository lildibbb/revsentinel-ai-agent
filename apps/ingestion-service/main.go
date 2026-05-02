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
	EventID      string    `json:"event_id"`
	TenantID     string    `json:"tenant_id"`
	EventType    string    `json:"event_type"`
	OccurredAt   time.Time `json:"occurred_at"`
	Findings     []Finding `json:"findings"`
	CasesCreated []string  `json:"cases_created"`
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

func main() {
	port := env("PORT", "8081")
	dbURL := env("DATABASE_URL", "")
	redisAddr := env("REDIS_ADDR", "redis:6379")
	queueName := env("ASYNQ_QUEUE_DEFAULT", "default")
	tenantID := env("TENANT_ID", "demo-tenant")
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

		var sourcePtr *string
		if req.Source != nil {
			b, err := json.Marshal(req.Source)
			if err != nil {
				writeError(w, http.StatusBadRequest, fmt.Errorf("invalid source: %w", err))
				return
			}
			s := string(b)
			sourcePtr = &s
		}

		var eventID string
		err = pool.QueryRow(r.Context(), `
			INSERT INTO ingested_event (tenant_id, event_type, occurred_at, source, payload)
			VALUES ($1,$2,$3,$4::jsonb,$5::jsonb)
			RETURNING id
		`, tenantID, req.EventType, occurredAt, sourcePtr, payloadJSON).Scan(&eventID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			traceID = eventID
		}
		err = publisher.EnqueueProcessEvent(r.Context(), eventID, tenantID, req.EventType, occurredAt.Format(time.RFC3339Nano), traceID)
		if err != nil {
			writeEnqueueError(w, eventID, err)
			return
		}

		resp := IngestResponse{
			EventID:      eventID,
			TenantID:     tenantID,
			EventType:    req.EventType,
			OccurredAt:   occurredAt,
			Findings:     []Finding{},
			CasesCreated: []string{},
		}
		writeJSON(w, http.StatusOK, resp)
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

func writeEnqueueError(w http.ResponseWriter, eventID string, err error) {
	writeJSON(w, http.StatusBadGateway, map[string]any{
		"error":    fmt.Sprintf("enqueue process_event.v1 failed: %v", err),
		"event_id": eventID,
	})
}
