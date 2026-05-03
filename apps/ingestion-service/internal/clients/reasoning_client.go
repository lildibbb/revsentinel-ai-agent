package clients

import (
	"context"
	"fmt"
	"net/http"
)

type ReasoningClient struct {
	endpoint string
	client   *http.Client
}

func NewReasoningClient(endpoint string, client *http.Client) *ReasoningClient {
	if client == nil {
		client = http.DefaultClient
	}
	return &ReasoningClient{
		endpoint: endpoint,
		client:   client,
	}
}

func (c *ReasoningClient) Generate(ctx context.Context, caseID string) error {
	url := fmt.Sprintf("%s/generate?caseID=%s", c.endpoint, caseID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call reasoning service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("reasoning service returned status %d", resp.StatusCode)
	}

	return nil
}
