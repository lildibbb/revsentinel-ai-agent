package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
)

type EvaluateRequest struct {
	TenantID  string                 `json:"tenant_id"`
	EventID   string                 `json:"event_id,omitempty"`
	EventType string                 `json:"event_type"`
	Payload   map[string]any         `json:"payload"`
	OccurredAt *time.Time            `json:"occurred_at,omitempty"`
}

type Finding struct {
	CaseType       string            `json:"case_type"`
	Severity       string            `json:"severity"`
	Title          string            `json:"title"`
	Summary        string            `json:"summary"`
	ExposureAmount *float64          `json:"exposure_amount,omitempty"`
	Currency       string            `json:"currency,omitempty"`
	Confidence     *float64          `json:"confidence,omitempty"`
	Evidence       map[string]any    `json:"evidence,omitempty"`
}

type EvaluateResponse struct {
	Findings []Finding `json:"findings"`
}

func main() {
	port := env("PORT", "8082")

	r := chi.NewRouter()
	
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	r.Post("/evaluate", func(w http.ResponseWriter, r *http.Request) {
		var req EvaluateRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if req.EventType == "" || req.Payload == nil {
			writeError(w, http.StatusBadRequest, errors.New("event_type and payload are required"))
			return
		}

		resp := EvaluateResponse{Findings: evaluate(req)}
		writeJSON(w, http.StatusOK, resp)
	})

	addr := ":" + port
	log.Printf("rules-service listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

func evaluate(req EvaluateRequest) []Finding {
	switch req.EventType {
	case "discount_event":
		return evalDiscount(req.Payload)
	case "service_delivery_event":
		return evalServiceDelivery(req.Payload)
	case "credit_note", "adjustment_event":
		return evalCredit(req.Payload)
	default:
		return nil
	}
}

func evalDiscount(p map[string]any) []Finding {
	dPct, ok1 := getFloat(p, "discount_pct")
	aPct, ok2 := getFloat(p, "allowed_discount_pct")
	amount, _ := getFloat(p, "amount")
	currency := getString(p, "currency")
	if currency == "" {
		currency = "USD"
	}
	if !ok1 || !ok2 {
		return nil
	}
	if dPct <= aPct {
		return nil
	}
	exposure := amount * (dPct - aPct)
	confidence := 0.92
	return []Finding{{
		CaseType:       "discount_threshold_violation",
		Severity:       "high",
		Title:          "Discount exceeds allowed threshold",
		Summary:        "A discount was applied above the contract/policy maximum.",
		ExposureAmount: &exposure,
		Currency:       currency,
		Confidence:     &confidence,
		Evidence: map[string]any{
			"discount_pct":         dPct,
			"allowed_discount_pct": aPct,
			"customer_id":          getString(p, "customer_id"),
			"contract_id":          getString(p, "contract_id"),
			"invoice_id":           getString(p, "invoice_id"),
			"amount":               amount,
		},
	}}}
}

func evalServiceDelivery(p map[string]any) []Finding {
	invoiceID := getString(p, "invoice_id")
	value, ok := getFloat(p, "value")
	currency := getString(p, "currency")
	if currency == "" {
		currency = "USD"
	}
	if invoiceID != "" {
		return nil
	}
	confidence := 0.85
	var exposure *float64
	if ok {
		exposure = &value
	}
	return []Finding{{
		CaseType:       "unbilled_service_delivery",
		Severity:       "high",
		Title:          "Service delivered but no invoice detected",
		Summary:        "A service delivery event has no linked invoice; review billing handoff and entitlements.",
		ExposureAmount: exposure,
		Currency:       currency,
		Confidence:     &confidence,
		Evidence: map[string]any{
			"customer_id": getString(p, "customer_id"),
			"service_id":  getString(p, "service_id"),
			"value":       value,
		},
	}}}
}

func evalCredit(p map[string]any) []Finding {
	amount, ok := getFloat(p, "amount")
	if !ok {
		return nil
	}
	currency := getString(p, "currency")
	if currency == "" {
		currency = "USD"
	}
	threshold := 1000.0
	if amount < threshold {
		return nil
	}
	confidence := 0.70
	return []Finding{{
		CaseType:       "suspicious_credit_or_adjustment",
		Severity:       "medium",
		Title:          "High-value credit/adjustment",
		Summary:        "A high-value credit/adjustment may indicate leakage or abuse; review frequency and authorization.",
		ExposureAmount: &amount,
		Currency:       currency,
		Confidence:     &confidence,
		Evidence: map[string]any{
			"customer_id": getString(p, "customer_id"),
			"amount":      amount,
			"reason":      getString(p, "reason"),
		},
	}}}
}

func getFloat(m map[string]any, key string) (float64, bool) {
	v, ok := m[key]
	if !ok || v == nil {
		return 0, false
	}
	switch t := v.(type) {
	case float64:
		return t, true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case json.Number:
		f, err := t.Float64()
		if err == nil {
			return f, true
		}
	}
	return 0, false
}

func getString(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
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
