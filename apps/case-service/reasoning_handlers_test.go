package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

type stubReasoningStore struct {
	insertResp CaseReasoningResponse
	insertErr  error
	latestResp CaseReasoningResponse
	latestErr  error
}

func (s stubReasoningStore) insertCaseReasoning(_ context.Context, _ string, _ CreateReasoningRequest) (CaseReasoningResponse, error) {
	return s.insertResp, s.insertErr
}

func (s stubReasoningStore) getLatestCaseReasoning(_ context.Context, _ string) (CaseReasoningResponse, error) {
	return s.latestResp, s.latestErr
}

func TestCreateReasoning_ValidRequest_ReturnsCreated(t *testing.T) {
	now := time.Date(2026, 4, 29, 8, 0, 0, 0, time.UTC)
	store := stubReasoningStore{
		insertResp: CaseReasoningResponse{
			ID:                    "11111111-1111-1111-1111-111111111111",
			CaseID:                "00000000-0000-0000-0000-000000000001",
			TenantID:              "demo-tenant",
			Status:                "success",
			ModelProvider:         "vertex-ai",
			ModelName:             "gemini-2.5-pro",
			PromptVersion:         "v1",
			ResponseSchemaVersion: "v1",
			Summary:               "Risk detected",
			RecommendedAction:     "Review contract exception",
			Citations: []ReasoningCitation{
				{Source: "policy", Reference: "POL-12", Excerpt: "Max discount 10%"},
			},
			TraceID:   "trace-1",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	r := chi.NewRouter()
	r.Post("/cases/{id}/reasoning", createReasoningHandler(store))

	reqBody := `{
		"status":"success",
		"model_provider":"vertex-ai",
		"model_name":"gemini-2.5-pro",
		"prompt_version":"v1",
		"response_schema_version":"v1",
		"summary":"Risk detected",
		"recommended_action":"Review contract exception",
		"citations":[{"source":"policy","reference":"POL-12","excerpt":"Max discount 10%"}],
		"amount_original":12000,
		"currency_original":"USD",
		"amount_myr_normalized":56400,
		"fx_rate_to_myr":4.7,
		"trace_id":"trace-1"
	}`

	req := httptest.NewRequest(http.MethodPost, "/cases/00000000-0000-0000-0000-000000000001/reasoning", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected %d, got %d body=%s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var body CaseReasoningResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.ID == "" || body.CaseID == "" {
		t.Fatalf("expected reasoning identifiers in response")
	}
	if body.Summary != "Risk detected" {
		t.Fatalf("expected summary in response")
	}
}

func TestGetLatestReasoning_ExistingReasoning_ReturnsOK(t *testing.T) {
	now := time.Date(2026, 4, 29, 9, 0, 0, 0, time.UTC)
	store := stubReasoningStore{
		latestResp: CaseReasoningResponse{
			ID:                    "22222222-2222-2222-2222-222222222222",
			CaseID:                "00000000-0000-0000-0000-000000000001",
			TenantID:              "demo-tenant",
			Status:                "success",
			ModelProvider:         "vertex-ai",
			ModelName:             "gemini-2.5-pro",
			PromptVersion:         "v1",
			ResponseSchemaVersion: "v1",
			Summary:               "Latest reasoning",
			RecommendedAction:     "Approve with monitoring",
			Citations: []ReasoningCitation{
				{Source: "policy", Reference: "POL-14", Excerpt: "Exception allowed with approval"},
			},
			TraceID:   "trace-2",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	r := chi.NewRouter()
	r.Get("/cases/{id}/reasoning/latest", getLatestReasoningHandler(store))

	req := httptest.NewRequest(http.MethodGet, "/cases/00000000-0000-0000-0000-000000000001/reasoning/latest", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var body CaseReasoningResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Summary != "Latest reasoning" {
		t.Fatalf("unexpected summary: %s", body.Summary)
	}
}

func TestGetLatestReasoning_NotFound_ReturnsNotFound(t *testing.T) {
	store := stubReasoningStore{latestErr: errReasoningNotFound}

	r := chi.NewRouter()
	r.Get("/cases/{id}/reasoning/latest", getLatestReasoningHandler(store))

	req := httptest.NewRequest(http.MethodGet, "/cases/00000000-0000-0000-0000-000000000001/reasoning/latest", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d body=%s", http.StatusNotFound, rec.Code, rec.Body.String())
	}
}

func TestCreateReasoning_DatabaseFailure_ReturnsInternalServerError(t *testing.T) {
	store := stubReasoningStore{insertErr: errors.New("db down")}

	r := chi.NewRouter()
	r.Post("/cases/{id}/reasoning", createReasoningHandler(store))

	reqBody := `{
		"status":"success",
		"model_provider":"vertex-ai",
		"model_name":"gemini-2.5-pro",
		"prompt_version":"v1",
		"response_schema_version":"v1",
		"summary":"Risk detected",
		"recommended_action":"Review contract exception",
		"citations":[{"source":"policy","reference":"POL-12","excerpt":"Max discount 10%"}],
		"trace_id":"trace-1"
	}`

	req := httptest.NewRequest(http.MethodPost, "/cases/00000000-0000-0000-0000-000000000001/reasoning", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d, got %d body=%s", http.StatusInternalServerError, rec.Code, rec.Body.String())
	}
}
