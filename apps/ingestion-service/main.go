package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IngestRequest struct {
	EventType  string         `json:"event_type"`
	OccurredAt *time.Time     `json:"occurred_at,omitempty"`
	Source     map[string]any `json:"source,omitempty"`
	Payload    map[string]any `json:"payload"`
}

type IngestResponse struct {
	EventID       string    `json:"event_id"`
	TenantID      string    `json:"tenant_id"`
	EventType     string    `json:"event_type"`
	OccurredAt    time.Time `json:"occurred_at"`
	Findings      []Finding `json:"findings"`
	CasesCreated  []string  `json:"cases_created"`
}

type EvaluateResponse struct {
	Findings []Finding `json:"findings"`
}

type Finding struct {
	CaseType       string         `json:"case_type"`
	Severity       string         `json:"severity"`
	Title          string         `json:"title"`
	Summary        string         `json:"summary"`
	ExposureAmount *float64       `json:"exposure_amount,omitempty"`
	Currency       string         `json:"currency,omitempty"`
	Confidence     *float64       `json:"confidence,omitempty"`
	Evidence       map[string]any `json:"evidence,omitempty"`
}

type EvidenceItem struct {
	Kind string         `json:"kind"`
	Data map[string]any `json:"data"`
}

type CreateCaseRequest struct {
	TenantID       string        `json:"tenant_id"`
	CaseType       string        `json:"case_type"`
	Status         string        `json:"status"`
	Severity       string        `json:"severity"`
	Title          string        `json:"title"`
	Summary        string        `json:"summary"`
	ExposureAmount *float64      `json:"exposure_amount,omitempty"`
	Currency       string        `json:"currency,omitempty"`
	Confidence     *float64      `json:"confidence,omitempty"`
	Evidence       []EvidenceItem `json:"evidence,omitempty"`
}

type CreateCaseResponse struct {
	ID string `json:"id"`
}

func main() {
	port := env("PORT", "8081")
	dbURL := env("DATABASE_URL", "")
	rulesURL := env("RULES_URL", "")
	casesURL := env("CASES_URL", "")
	tenantID := env("TENANT_ID", "demo-tenant")
	if dbURL == "" || rulesURL == "" || casesURL == "" {
		log.Fatal("DATABASE_URL, RULES_URL, and CASES_URL are required")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	client := &http.Client{Timeout: 10 * time.Second}

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

		findings, err := callRules(r.Context(), client, rulesURL, tenantID, eventID, req.EventType, occurredAt, req.Payload)
		if err != nil {
			writeError(w, http.StatusBadGateway, err)
			return
		}

		caseIDs := make([]string, 0)
		for _, f := range findings {
			caseID, err := createCase(r.Context(), client, casesURL, tenantID, eventID, req.EventType, occurredAt, req.Payload, f)
			if err != nil {
				writeError(w, http.StatusBadGateway, err)
				return
			}
			caseIDs = append(caseIDs, caseID)
		}

		resp := IngestResponse{
			EventID:      eventID,
			TenantID:     tenantID,
			EventType:    req.EventType,
			OccurredAt:   occurredAt,
			Findings:     findings,
			CasesCreated: caseIDs,
		}
		writeJSON(w, http.StatusOK, resp)
	})

	addr := ":" + port
	log.Printf("ingestion-service listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

func callRules(ctx context.Context, client *http.Client, rulesURL, tenantID, eventID, eventType string, occurredAt time.Time, payload map[string]any) ([]Finding, error) {
	body, _ := json.Marshal(map[string]any{
		"tenant_id":   tenantID,
		"event_id":    eventID,
		"event_type":  eventType,
		"occurred_at": occurredAt,
		"payload":     payload,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rulesURL+"/evaluate", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("rules-service %d: %s", resp.StatusCode, string(b))
	}

	var out EvaluateResponse
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&out); err != nil {
		return nil, err
	}
	return out.Findings, nil
}

func createCase(ctx context.Context, client *http.Client, casesURL, tenantID, eventID, eventType string, occurredAt time.Time, payload map[string]any, f Finding) (string, error) {
	reqBody := CreateCaseRequest{
		TenantID:       tenantID,
		CaseType:       f.CaseType,
		Status:         "detected",
		Severity:       f.Severity,
		Title:          f.Title,
		Summary:        f.Summary,
		ExposureAmount: f.ExposureAmount,
		Currency:       f.Currency,
		Confidence:     f.Confidence,
		Evidence: []EvidenceItem{
			{Kind: "trigger_event", Data: map[string]any{"event_id": eventID, "event_type": eventType, "occurred_at": occurredAt}},
			{Kind: "event_payload", Data: payload},
			{Kind: "rule_finding", Data: map[string]any{"case_type": f.CaseType, "severity": f.Severity, "evidence": f.Evidence}},
		},
	}
	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, casesURL+"/cases", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("case-service %d: %s", resp.StatusCode, string(body))
	}

	var out CreateCaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.ID == "" {
		return "", errors.New("case-service returned empty id")
	}
	return out.ID, nil
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
