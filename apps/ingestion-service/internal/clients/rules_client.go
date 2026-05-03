package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type RulesClient struct {
	endpoint string
	client   *http.Client
}

func NewRulesClient(endpoint string, client *http.Client) *RulesClient {
	if client == nil {
		client = http.DefaultClient
	}
	return &RulesClient{
		endpoint: endpoint,
		client:   client,
	}
}

func (c *RulesClient) Evaluate(ctx context.Context, eventType, eventID string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/evaluate?eventType=%s&eventID=%s", c.endpoint, eventType, eventID)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call rules service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rules service returned status %d", resp.StatusCode)
	}

	var findings []map[string]interface{}
	if err := json.Unmarshal(body, &findings); err != nil {
		return nil, fmt.Errorf("failed to parse findings: %w", err)
	}

	return findings, nil
}
