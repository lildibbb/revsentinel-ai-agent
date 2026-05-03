package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type CaseClient struct {
	endpoint string
	client   *http.Client
}

type CreateCaseRequest struct {
	TenantID string                   `json:"tenant_id"`
	Findings []map[string]interface{} `json:"findings"`
}

type CreateCaseResponse struct {
	CaseID string `json:"case_id"`
}

func NewCaseClient(endpoint string, client *http.Client) *CaseClient {
	if client == nil {
		client = http.DefaultClient
	}
	return &CaseClient{
		endpoint: endpoint,
		client:   client,
	}
}

func (c *CaseClient) CreateFromFindings(ctx context.Context, tenantID string, findings []map[string]interface{}) (string, error) {
	req := CreateCaseRequest{
		TenantID: tenantID,
		Findings: findings,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/cases", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to call case service: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("case service returned status %d", resp.StatusCode)
	}

	var respObj CreateCaseResponse
	if err := json.Unmarshal(respBody, &respObj); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return respObj.CaseID, nil
}
