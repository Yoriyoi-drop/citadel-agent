package communication

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// SlackClient handles Slack API interactions
type SlackClient struct {
	webhookURL string
	token      string
	client     *http.Client
}

// NewSlackClient creates a new Slack client
func NewSlackClient(webhookURL, token string) *SlackClient {
	return &SlackClient{
		webhookURL: webhookURL,
		token:      token,
		client:     &http.Client{},
	}
}

// SendMessage sends a message to a Slack channel
func (s *SlackClient) SendMessage(ctx context.Context, channel, text string) error {
	payload := map[string]interface{}{
		"channel": channel,
		"text":    text,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := s.webhookURL
	if url == "" {
		url = "https://slack.com/api/chat.postMessage"
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if s.token != "" {
		req.Header.Set("Authorization", "Bearer "+s.token)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack API returned status %d", resp.StatusCode)
	}

	return nil
}
