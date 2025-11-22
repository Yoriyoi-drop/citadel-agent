// backend/internal/nodes/ai/ai_agent.go
package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
)

// AIProvider represents the AI service provider
type AIProvider string

const (
	ProviderOpenAI   AIProvider = "openai"
	ProviderAnthropic AIProvider = "anthropic"
	ProviderLocal    AIProvider = "local"
)

// AINodeConfig represents the configuration for an AI node
type AINodeConfig struct {
	Provider     AIProvider         `json:"provider"`
	Model        string             `json:"model"`
	APIKey       string             `json:"api_key"`
	MaxTokens    int                `json:"max_tokens"`
	Temperature  float64            `json:"temperature"`
	TopP         float64            `json:"top_p"`
	StopSequences []string          `json:"stop_sequences"`
	SystemPrompt string             `json:"system_prompt"`
	Tools        []AITool           `json:"tools"`
	Memory       *AIMemoryConfig    `json:"memory"`
	Timeout      time.Duration      `json:"timeout"`
}

// AITool represents an available tool for the AI agent
type AITool struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// AIMemoryConfig represents memory configuration for the AI agent
type AIMemoryConfig struct {
	EnableShortTerm bool   `json:"enable_short_term"`
	EnableLongTerm  bool   `json:"enable_long_term"`
	CollectionName  string `json:"collection_name"`
	ContextSize     int    `json:"context_size"`
}

// AINode represents an AI agent node
type AINode struct {
	config *AINodeConfig
}

// NewAINode creates a new AI node
func NewAINode(config *AINodeConfig) *AINode {
	// Set defaults if not provided
	if config.MaxTokens == 0 {
		config.MaxTokens = 1024
	}
	if config.Temperature == 0 {
		config.Temperature = 0.7
	}
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}

	return &AINode{
		config: config,
	}
}

// Execute executes the AI node
func (an *AINode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Extract prompt from inputs
	prompt := ""
	if promptVal, exists := inputs["prompt"]; exists {
		if promptStr, ok := promptVal.(string); ok {
			prompt = promptStr
		}
	}

	// Extract other parameters from inputs
	model := an.config.Model
	if modelVal, exists := inputs["model"]; exists {
		if modelStr, ok := modelVal.(string); ok {
			model = modelStr
		}
	}

	maxTokens := an.config.MaxTokens
	if tokensVal, exists := inputs["max_tokens"]; exists {
		if tokensFloat, ok := tokensVal.(float64); ok {
			maxTokens = int(tokensFloat)
		}
	}

	temperature := an.config.Temperature
	if tempVal, exists := inputs["temperature"]; exists {
		if tempFloat, ok := tempVal.(float64); ok {
			temperature = tempFloat
		}
	}

	// Prepare the AI request based on provider
	switch an.config.Provider {
	case ProviderOpenAI:
		return an.executeOpenAI(ctx, prompt, model, maxTokens, temperature)
	case ProviderAnthropic:
		return an.executeAnthropic(ctx, prompt, model, maxTokens, temperature)
	case ProviderLocal:
		return an.executeLocal(ctx, prompt, model, maxTokens, temperature)
	default:
		return an.executeOpenAI(ctx, prompt, model, maxTokens, temperature) // Default to OpenAI
	}
}

// executeOpenAI executes the AI request with OpenAI
func (an *AINode) executeOpenAI(ctx context.Context, prompt, model string, maxTokens int, temperature float64) (map[string]interface{}, error) {
	if an.config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	// Prepare the request payload
	requestBody := map[string]interface{}{
		"model":       model,
		"messages":    []map[string]string{{"role": "user", "content": prompt}},
		"max_tokens":  maxTokens,
		"temperature": temperature,
	}

	if an.config.SystemPrompt != "" {
		// Insert system prompt at the beginning
		messages := []map[string]string{
			{"role": "system", "content": an.config.SystemPrompt},
			{"role": "user", "content": prompt},
		}
		requestBody["messages"] = messages
	}

	// Add tools if configured
	if len(an.config.Tools) > 0 {
		requestBody["tools"] = an.convertToolsToOpenAIFormat()
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request with timeout
	httpCtx, cancel := context.WithTimeout(ctx, an.config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(httpCtx, "POST", "https://api.openai.com/v1/chat/completions", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+an.config.APIKey)

	// Execute request
	client := &http.Client{Timeout: an.config.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse response
	var aiResponse map[string]interface{}
	if err := json.Unmarshal(responseBody, &aiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract the content from the response
	choices, ok := aiResponse["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid choice format")
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid message format")
	}

	content, ok := message["content"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid content format")
	}

	// Prepare the result
	result := map[string]interface{}{
		"success": true,
		"content": content,
		"model":   model,
		"provider": string(ProviderOpenAI),
		"tokens_used": map[string]interface{}{
			"input":  len(prompt),
			"output": len(content),
		},
		"timestamp": time.Now().Unix(),
	}

	// Add tool calls if present
	if toolCalls, exists := message["tool_calls"]; exists {
		result["tool_calls"] = toolCalls
	}

	return result, nil
}

// executeAnthropic executes the AI request with Anthropic
func (an *AINode) executeAnthropic(ctx context.Context, prompt, model string, maxTokens int, temperature float64) (map[string]interface{}, error) {
	if an.config.APIKey == "" {
		return nil, fmt.Errorf("Anthropic API key is required")
	}

	// Prepare the request payload
	requestBody := map[string]interface{}{
		"model":       model,
		"messages":    []map[string]string{{"role": "user", "content": prompt}},
		"max_tokens":  maxTokens,
		"temperature": temperature,
	}

	if an.config.SystemPrompt != "" {
		requestBody["system"] = an.config.SystemPrompt
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request with timeout
	httpCtx, cancel := context.WithTimeout(ctx, an.config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(httpCtx, "POST", "https://api.anthropic.com/v1/messages", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", an.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Execute request
	client := &http.Client{Timeout: an.config.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse response
	var aiResponse map[string]interface{}
	if err := json.Unmarshal(responseBody, &aiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract the content from the response
	contentArray, ok := aiResponse["content"].([]interface{})
	if !ok || len(contentArray) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	contentBlock, ok := contentArray[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid content block format")
	}

	content, ok := contentBlock["text"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid content text format")
	}

	// Prepare the result
	result := map[string]interface{}{
		"success": true,
		"content": content,
		"model":   model,
		"provider": string(ProviderAnthropic),
		"tokens_used": map[string]interface{}{
			"input":  len(prompt),
			"output": len(content),
		},
		"timestamp": time.Now().Unix(),
	}

	return result, nil
}

// executeLocal executes the AI request with a local model (simulated)
func (an *AINode) executeLocal(ctx context.Context, prompt, model string, maxTokens int, temperature float64) (map[string]interface{}, error) {
	// Simulate local AI execution
	// In a real implementation, this would call a local AI model

	// For simulation purposes, return a mock response
	content := fmt.Sprintf("This is a simulated response from local model '%s' to the prompt: '%s'", model, prompt)
	
	if len(content) > maxTokens {
		content = content[:maxTokens]
	}

	result := map[string]interface{}{
		"success": true,
		"content": content,
		"model":   model,
		"provider": string(ProviderLocal),
		"tokens_used": map[string]interface{}{
			"input":  len(prompt),
			"output": len(content),
		},
		"timestamp": time.Now().Unix(),
		"local_execution": true,
	}

	return result, nil
}

// convertToolsToOpenAIFormat converts tools to OpenAI format
func (an *AINode) convertToolsToOpenAIFormat() []map[string]interface{} {
	var tools []map[string]interface{}
	
	for _, tool := range an.config.Tools {
		toolDef := map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.Parameters,
			},
		}
		tools = append(tools, toolDef)
	}
	
	return tools
}

// RegisterAINode registers the AI node type with the engine
func RegisterAINode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("ai_agent", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		var provider AIProvider
		if providerVal, exists := config["provider"]; exists {
			if providerStr, ok := providerVal.(string); ok {
				provider = AIProvider(providerStr)
			}
		}

		var model string
		if modelVal, exists := config["model"]; exists {
			if modelStr, ok := modelVal.(string); ok {
				model = modelStr
			}
		}

		var apiKey string
		if keyVal, exists := config["api_key"]; exists {
			if keyStr, ok := keyVal.(string); ok {
				apiKey = keyStr
			}
		}

		var maxTokens float64
		if tokensVal, exists := config["max_tokens"]; exists {
			if tokensFloat, ok := tokensVal.(float64); ok {
				maxTokens = tokensFloat
			}
		}

		var temperature float64
		if tempVal, exists := config["temperature"]; exists {
			if tempFloat, ok := tempVal.(float64); ok {
				temperature = tempFloat
			}
		}

		var topP float64
		if topPVal, exists := config["top_p"]; exists {
			if topPFloat, ok := topPVal.(float64); ok {
				topP = topPFloat
			}
		}

		var systemPrompt string
		if sysPromptVal, exists := config["system_prompt"]; exists {
			if sysPromptStr, ok := sysPromptVal.(string); ok {
				systemPrompt = sysPromptStr
			}
		}

		var timeout float64
		if timeoutVal, exists := config["timeout_seconds"]; exists {
			if timeoutFloat, ok := timeoutVal.(float64); ok {
				timeout = timeoutFloat
			}
		}

		var stopSequences []string
		if stopVal, exists := config["stop_sequences"]; exists {
			if stopSlice, ok := stopVal.([]interface{}); ok {
				for _, seq := range stopSlice {
					if seqStr, ok := seq.(string); ok {
						stopSequences = append(stopSequences, seqStr)
					}
				}
			}
		}

		var tools []AITool
		if toolsVal, exists := config["tools"]; exists {
			if toolsSlice, ok := toolsVal.([]interface{}); ok {
				for _, tool := range toolsSlice {
					if toolMap, ok := tool.(map[string]interface{}); ok {
						var toolObj AITool
						if name, exists := toolMap["name"]; exists {
							if nameStr, ok := name.(string); ok {
								toolObj.Name = nameStr
							}
						}
						if desc, exists := toolMap["description"]; exists {
							if descStr, ok := desc.(string); ok {
								toolObj.Description = descStr
							}
						}
						if params, exists := toolMap["parameters"]; exists {
							if paramsMap, ok := params.(map[string]interface{}); ok {
								toolObj.Parameters = paramsMap
							}
						}
						tools = append(tools, toolObj)
					}
				}
			}
		}

		var memory *AIMemoryConfig
		if memVal, exists := config["memory"]; exists {
			if memMap, ok := memVal.(map[string]interface{}); ok {
				var memConfig AIMemoryConfig
				if shortTerm, exists := memMap["enable_short_term"]; exists {
					if shortBool, ok := shortTerm.(bool); ok {
						memConfig.EnableShortTerm = shortBool
					}
				}
				if longTerm, exists := memMap["enable_long_term"]; exists {
					if longBool, ok := longTerm.(bool); ok {
						memConfig.EnableLongTerm = longBool
					}
				}
				if collection, exists := memMap["collection_name"]; exists {
					if collectionStr, ok := collection.(string); ok {
						memConfig.CollectionName = collectionStr
					}
				}
				if contextSize, exists := memMap["context_size"]; exists {
					if contextFloat, ok := contextSize.(float64); ok {
						memConfig.ContextSize = int(contextFloat)
					}
				}
				memory = &memConfig
			}
		}

		nodeConfig := &AINodeConfig{
			Provider:      provider,
			Model:         model,
			APIKey:        apiKey,
			MaxTokens:     int(maxTokens),
			Temperature:   temperature,
			TopP:          topP,
			StopSequences: stopSequences,
			SystemPrompt:  systemPrompt,
			Tools:         tools,
			Memory:        memory,
			Timeout:       time.Duration(timeout) * time.Second,
		}

		return NewAINode(nodeConfig), nil
	})
}