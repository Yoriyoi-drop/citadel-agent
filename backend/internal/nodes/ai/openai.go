package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"citadel-agent/backend/internal/nodes/base"
)

// OpenAINode implements OpenAI API integration
type OpenAINode struct {
	*base.BaseNode
}

// OpenAIConfig holds OpenAI configuration
type OpenAIConfig struct {
	APIKey       string  `json:"api_key"`
	Model        string  `json:"model"`
	Prompt       string  `json:"prompt"`
	Temperature  float64 `json:"temperature"`
	MaxTokens    int     `json:"max_tokens"`
	SystemPrompt string  `json:"system_prompt"`
}

// OpenAIRequest represents OpenAI API request
type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents OpenAI API response
type OpenAIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a response choice
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage represents token usage
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// NewOpenAIGPT4Node creates OpenAI GPT-4 node
func NewOpenAIGPT4Node() base.Node {
	metadata := base.NodeMetadata{
		ID:          "openai_gpt4",
		Name:        "OpenAI GPT-4",
		Category:    "ai_llm",
		Description: "GPT-4 text generation",
		Version:     "1.0.0",
		Author:      "Citadel Agent",
		Icon:        "brain",
		Color:       "#8b5cf6",
		Inputs: []base.NodeInput{
			{
				ID:          "prompt",
				Name:        "Prompt",
				Type:        "string",
				Required:    true,
				Description: "User prompt",
			},
		},
		Outputs: []base.NodeOutput{
			{
				ID:          "response",
				Name:        "Response",
				Type:        "string",
				Description: "AI response",
			},
			{
				ID:          "usage",
				Name:        "Usage",
				Type:        "object",
				Description: "Token usage",
			},
		},
		Config: []base.NodeConfig{
			{
				Name:        "api_key",
				Label:       "API Key",
				Description: "OpenAI API key",
				Type:        "password",
				Required:    true,
			},
			{
				Name:        "model",
				Label:       "Model",
				Description: "GPT model",
				Type:        "select",
				Required:    true,
				Default:     "gpt-4",
				Options: []base.ConfigOption{
					{Label: "GPT-4", Value: "gpt-4"},
					{Label: "GPT-4 Turbo", Value: "gpt-4-turbo-preview"},
					{Label: "GPT-4 32K", Value: "gpt-4-32k"},
				},
			},
			{
				Name:        "system_prompt",
				Label:       "System Prompt",
				Description: "System instructions",
				Type:        "textarea",
				Required:    false,
			},
			{
				Name:        "temperature",
				Label:       "Temperature",
				Description: "Randomness (0-2)",
				Type:        "number",
				Required:    false,
				Default:     0.7,
			},
			{
				Name:        "max_tokens",
				Label:       "Max Tokens",
				Description: "Maximum response tokens",
				Type:        "number",
				Required:    false,
				Default:     1000,
			},
		},
		Tags: []string{"openai", "gpt4", "llm", "ai"},
	}

	return &OpenAINode{
		BaseNode: base.NewBaseNode(metadata),
	}
}

// NewOpenAIGPT35Node creates OpenAI GPT-3.5 node
func NewOpenAIGPT35Node() base.Node {
	metadata := base.NodeMetadata{
		ID:          "openai_gpt35",
		Name:        "OpenAI GPT-3.5",
		Category:    "ai_llm",
		Description: "GPT-3.5 Turbo text generation",
		Version:     "1.0.0",
		Author:      "Citadel Agent",
		Icon:        "brain",
		Color:       "#8b5cf6",
		Inputs: []base.NodeInput{
			{
				ID:          "prompt",
				Name:        "Prompt",
				Type:        "string",
				Required:    true,
				Description: "User prompt",
			},
		},
		Outputs: []base.NodeOutput{
			{
				ID:          "response",
				Name:        "Response",
				Type:        "string",
				Description: "AI response",
			},
		},
		Config: []base.NodeConfig{
			{
				Name:        "api_key",
				Label:       "API Key",
				Description: "OpenAI API key",
				Type:        "password",
				Required:    true,
			},
			{
				Name:        "model",
				Label:       "Model",
				Description: "GPT model",
				Type:        "select",
				Required:    true,
				Default:     "gpt-3.5-turbo",
				Options: []base.ConfigOption{
					{Label: "GPT-3.5 Turbo", Value: "gpt-3.5-turbo"},
					{Label: "GPT-3.5 Turbo 16K", Value: "gpt-3.5-turbo-16k"},
				},
			},
			{
				Name:        "temperature",
				Label:       "Temperature",
				Description: "Randomness (0-2)",
				Type:        "number",
				Required:    false,
				Default:     0.7,
			},
			{
				Name:        "max_tokens",
				Label:       "Max Tokens",
				Description: "Maximum response tokens",
				Type:        "number",
				Required:    false,
				Default:     1000,
			},
		},
		Tags: []string{"openai", "gpt3.5", "llm", "ai"},
	}

	return &OpenAINode{
		BaseNode: base.NewBaseNode(metadata),
	}
}

// Execute calls OpenAI API
func (n *OpenAINode) Execute(ctx *base.ExecutionContext, inputs map[string]interface{}) (*base.ExecutionResult, error) {
	startTime := time.Now()

	// Parse configuration
	var config OpenAIConfig
	if err := base.UnmarshalConfig(ctx.Variables, &config); err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Get prompt from inputs
	prompt, ok := inputs["prompt"].(string)
	if !ok {
		prompt = config.Prompt
	}

	if prompt == "" {
		err := fmt.Errorf("prompt is required")
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Build messages
	messages := []Message{}
	if config.SystemPrompt != "" {
		messages = append(messages, Message{
			Role:    "system",
			Content: config.SystemPrompt,
		})
	}
	messages = append(messages, Message{
		Role:    "user",
		Content: prompt,
	})

	// Create request
	reqBody := OpenAIRequest{
		Model:       config.Model,
		Messages:    messages,
		Temperature: config.Temperature,
		MaxTokens:   config.MaxTokens,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Make API request
	req, err := http.NewRequestWithContext(ctx.Context, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("OpenAI API error: %s", string(body))
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Parse response
	var apiResp OpenAIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	if len(apiResp.Choices) == 0 {
		err := fmt.Errorf("no response from OpenAI")
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	result := map[string]interface{}{
		"response": apiResp.Choices[0].Message.Content,
		"usage": map[string]interface{}{
			"prompt_tokens":     apiResp.Usage.PromptTokens,
			"completion_tokens": apiResp.Usage.CompletionTokens,
			"total_tokens":      apiResp.Usage.TotalTokens,
		},
		"model":         apiResp.Model,
		"finish_reason": apiResp.Choices[0].FinishReason,
	}

	ctx.Logger.Info("OpenAI request completed", map[string]interface{}{
		"model":        config.Model,
		"total_tokens": apiResp.Usage.TotalTokens,
	})

	return base.CreateSuccessResult(result, time.Since(startTime)), nil
}
