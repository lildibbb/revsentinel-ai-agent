package main

import "time"

type ReasoningCitation struct {
	Source    string `json:"source"`
	Reference string `json:"reference"`
	Excerpt   string `json:"excerpt"`
}

type CreateReasoningRequest struct {
	Status                string              `json:"status"`
	ModelProvider         string              `json:"model_provider"`
	ModelName             string              `json:"model_name"`
	ModelVersion          *string             `json:"model_version,omitempty"`
	PromptVersion         string              `json:"prompt_version"`
	ResponseSchemaVersion string              `json:"response_schema_version"`
	Summary               string              `json:"summary"`
	RecommendedAction     string              `json:"recommended_action"`
	Rationale             map[string]any      `json:"rationale,omitempty"`
	Confidence            *float64            `json:"confidence,omitempty"`
	Citations             []ReasoningCitation `json:"citations"`
	TokenInput            *int                `json:"token_input,omitempty"`
	TokenOutput           *int                `json:"token_output,omitempty"`
	LatencyMs             *int                `json:"latency_ms,omitempty"`
	TraceID               string              `json:"trace_id"`
	ErrorCode             *string             `json:"error_code,omitempty"`
	ErrorMessage          *string             `json:"error_message,omitempty"`
	AmountOriginal        *float64            `json:"amount_original,omitempty"`
	CurrencyOriginal      *string             `json:"currency_original,omitempty"`
	AmountMYRNormalized   *float64            `json:"amount_myr_normalized,omitempty"`
	FXRateToMYR           *float64            `json:"fx_rate_to_myr,omitempty"`
	FXRateTimestamp       *time.Time          `json:"fx_rate_timestamp,omitempty"`
}

type CaseReasoningResponse struct {
	ID                    string              `json:"id"`
	CaseID                string              `json:"case_id"`
	TenantID              string              `json:"tenant_id"`
	Status                string              `json:"status"`
	ModelProvider         string              `json:"model_provider"`
	ModelName             string              `json:"model_name"`
	ModelVersion          *string             `json:"model_version,omitempty"`
	PromptVersion         string              `json:"prompt_version"`
	ResponseSchemaVersion string              `json:"response_schema_version"`
	Summary               string              `json:"summary"`
	RecommendedAction     string              `json:"recommended_action"`
	Rationale             map[string]any      `json:"rationale,omitempty"`
	Confidence            *float64            `json:"confidence,omitempty"`
	Citations             []ReasoningCitation `json:"citations"`
	TokenInput            *int                `json:"token_input,omitempty"`
	TokenOutput           *int                `json:"token_output,omitempty"`
	LatencyMs             *int                `json:"latency_ms,omitempty"`
	TraceID               string              `json:"trace_id"`
	ErrorCode             *string             `json:"error_code,omitempty"`
	ErrorMessage          *string             `json:"error_message,omitempty"`
	AmountOriginal        *float64            `json:"amount_original,omitempty"`
	CurrencyOriginal      *string             `json:"currency_original,omitempty"`
	AmountMYRNormalized   *float64            `json:"amount_myr_normalized,omitempty"`
	FXRateToMYR           *float64            `json:"fx_rate_to_myr,omitempty"`
	FXRateTimestamp       *time.Time          `json:"fx_rate_timestamp,omitempty"`
	CreatedAt             time.Time           `json:"created_at"`
	UpdatedAt             time.Time           `json:"updated_at"`
}
