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
	"github.com/jackc/pgx/v5/pgxpool"
)

type EvidenceItem struct {
	Kind string                 `json:"kind"`
	Data map[string]any         `json:"data"`
}

type CreateCaseRequest struct {
	TenantID       string         `json:"tenant_id"`
	CaseType       string         `json:"case_type"`
	Status         string         `json:"status,omitempty"`
	Severity       string         `json:"severity"`
	Title          string         `json:"title"`
	Summary        string         `json:"summary"`
	ExposureAmount *float64       `json:"exposure_amount,omitempty"`
	Currency       string         `json:"currency,omitempty"`
	Confidence     *float64       `json:"confidence,omitempty"`
	Evidence       []EvidenceItem  `json:"evidence,omitempty"`
}

type Case struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	CaseType       string    `json:"case_type"`
	Status         string    `json:"status"`
	Severity       string    `json:"severity"`
	Title          string    `json:"title"`
	Summary        string    `json:"summary"`
	ExposureAmount *float64  `json:"exposure_amount,omitempty"`
	Currency       string    `json:"currency"`
	Confidence     *float64  `json:"confidence,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Evidence       []EvidenceItem `json:"evidence,omitempty"`
}

func main() {
	port := env("PORT", "8083")
	dbURL := env("DATABASE_URL", "")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	r := chi.NewRouter()

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	r.Post("/cases", func(w http.ResponseWriter, r *http.Request) {
		var req CreateCaseRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if req.TenantID == "" || req.CaseType == "" || req.Severity == "" || req.Title == "" || req.Summary == "" {
			writeError(w, http.StatusBadRequest, errors.New("tenant_id, case_type, severity, title, summary are required"))
			return
		}
		if req.Status == "" {
			req.Status = "detected"
		}
		if req.Currency == "" {
			req.Currency = "USD"
		}

		caseID, createdAt, updatedAt, err := insertCase(ctx, pool, req)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		resp := Case{
			ID:             caseID,
			TenantID:       req.TenantID,
			CaseType:       req.CaseType,
			Status:         req.Status,
			Severity:       req.Severity,
			Title:          req.Title,
			Summary:        req.Summary,
			ExposureAmount: req.ExposureAmount,
			Currency:       req.Currency,
			Confidence:     req.Confidence,
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
			Evidence:       req.Evidence,
		}
		writeJSON(w, http.StatusCreated, resp)
	})

	r.Get("/cases", func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.URL.Query().Get("tenant_id")
		if tenantID == "" {
			tenantID = "demo-tenant"
		}
		rows, err := pool.Query(r.Context(), `
			SELECT id, tenant_id, case_type, status, severity, title, summary, exposure_amount, currency, confidence, created_at, updated_at
			FROM leak_case
			WHERE tenant_id=$1
			ORDER BY created_at DESC
			LIMIT 50
		`, tenantID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		defer rows.Close()

		cases := make([]Case, 0)
		for rows.Next() {
			var c Case
			if err := rows.Scan(&c.ID, &c.TenantID, &c.CaseType, &c.Status, &c.Severity, &c.Title, &c.Summary, &c.ExposureAmount, &c.Currency, &c.Confidence, &c.CreatedAt, &c.UpdatedAt); err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
			cases = append(cases, c)
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": cases})
	})

	r.Get("/cases/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var c Case
		err := pool.QueryRow(r.Context(), `
			SELECT id, tenant_id, case_type, status, severity, title, summary, exposure_amount, currency, confidence, created_at, updated_at
			FROM leak_case
			WHERE id=$1
		`, id).Scan(&c.ID, &c.TenantID, &c.CaseType, &c.Status, &c.Severity, &c.Title, &c.Summary, &c.ExposureAmount, &c.Currency, &c.Confidence, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			writeError(w, http.StatusNotFound, fmt.Errorf("case not found"))
			return
		}

		eRows, err := pool.Query(r.Context(), `SELECT kind, data FROM case_evidence WHERE case_id=$1 ORDER BY created_at ASC`, id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		defer eRows.Close()
		ev := make([]EvidenceItem, 0)
		for eRows.Next() {
			var kind string
			var dataBytes []byte
			if err := eRows.Scan(&kind, &dataBytes); err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
			var data map[string]any
			if len(dataBytes) > 0 {
				_ = json.Unmarshal(dataBytes, &data)
			}
			ev = append(ev, EvidenceItem{Kind: kind, Data: data})
		}
		c.Evidence = ev

		writeJSON(w, http.StatusOK, c)
	})

	addr := ":" + port
	log.Printf("case-service listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func insertCase(ctx context.Context, pool *pgxpool.Pool, req CreateCaseRequest) (id string, createdAt time.Time, updatedAt time.Time, err error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return "", time.Time{}, time.Time{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	err = tx.QueryRow(ctx, `
		INSERT INTO leak_case (tenant_id, case_type, status, severity, title, summary, exposure_amount, currency, confidence)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, created_at, updated_at
	`, req.TenantID, req.CaseType, req.Status, req.Severity, req.Title, req.Summary, req.ExposureAmount, req.Currency, req.Confidence).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return "", time.Time{}, time.Time{}, err
	}

	for _, e := range req.Evidence {
		if e.Kind == "" {
			continue
		}
		if e.Data == nil {
			e.Data = map[string]any{}
		}
		b, mErr := json.Marshal(e.Data)
		if mErr != nil {
			return "", time.Time{}, time.Time{}, mErr
		}
		_, err = tx.Exec(ctx, `INSERT INTO case_evidence (case_id, kind, data) VALUES ($1,$2,$3::jsonb)`, id, e.Kind, string(b))
		if err != nil {
			return "", time.Time{}, time.Time{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", time.Time{}, time.Time{}, err
	}
	return id, createdAt, updatedAt, nil
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
