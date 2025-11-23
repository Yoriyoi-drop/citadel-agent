package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/citadel-agent/backend/internal/nodes/ai"
)

// OpenAIProvider implements the AI Provider interface for OpenAI
type OpenAIProvider struct {
	apiKey string
	client *http.Client
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider() *OpenAIProvider {
	return &OpenAIProvider{
		apiKey: os.Getenv("OPENAI_API_KEY"),
		client: &http.Client{},
	}
}

// Generate generates text using OpenAI
func (p *OpenAIProvider) Generate(ctx context.Context, req ai.Request) (*ai.Response, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key not set")
	}

	url := "https://api.openai.com/v1/chat/completions"

	payload := map[string]interface{}{
		"model": req.ModelName,
		"messages": []map[string]string{
			{"role": "user", "content": req.Prompt},
		},
	}

	jsonPayload, _ := json.Marshal(payload)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("OpenAI API error: %s", string(body))
	}

	// Simplified response parsing
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	choices := result["choices"].([]interface{})
	firstChoice := choices[0].(map[string]interface{})
	message := firstChoice["message"].(map[string]interface{})
	content := message["content"].(string)

	return &ai.Response{
		Text: content,
	}, nil
}
