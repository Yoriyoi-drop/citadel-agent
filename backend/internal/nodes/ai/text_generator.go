package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
)

// TextGeneratorConfig represents the configuration for text generator AI node
type TextGeneratorConfig struct {
	ModelName     string            `json:"model_name"`      // e.g., "gpt-3.5-turbo", "llama-3.2-3b"
	Provider      string            `json:"provider"`        // "openai", "anthropic", "local", etc.
	ApiKey        string            `json:"api_key"`         // API key for the provider
	Temperature   float64           `json:"temperature"`     // Creativity level (0.0-2.0)
	MaxTokens     int               `json:"max_tokens"`      // Max tokens in response
	Prompt        string            `json:"prompt"`          // The main prompt
	SystemPrompt  string            `json:"system_prompt"`   // System message for the AI
	Timeout       int               `json:"timeout"`         // Timeout in seconds
	Parameters    map[string]interface{} `json:"parameters"` // Additional parameters
	EnableCaching bool              `json:"enable_caching"`  // Enable result caching
	CacheTTL      int               `json:"cache_ttl"`       // Cache TTL in seconds
	EnableProfiling bool            `json:"enable_profiling"` // Enable profiling
}

// TextGeneratorNode represents an AI-powered text generation node
type TextGeneratorNode struct {
	config *TextGeneratorConfig
	aiManager *AIManager
}

// AIManager handles the actual AI operations
type AIManager struct {
	// This would interact with various AI providers
	// For now, using a mock implementation
}

// NewAIManager creates a new AI manager
func NewAIManager() *AIManager {
	return &AIManager{}
}

// GenerateText generates text based on the provided prompt
func (am *AIManager) GenerateText(config *TextGeneratorConfig) (string, error) {
	// In a real implementation, this would:
	// 1. Route to the appropriate AI provider (local, API, etc.)
	// 2. Handle rate limiting
	// 3. Manage costs
	// 4. Handle caching
	// 5. Apply safety filters
	// 6. Process the response
	
	// For this example, we'll simulate the call
	time.Sleep(100 * time.Millisecond) // Simulate API call
	
	// Mock response
	response := fmt.Sprintf("Generated text based on prompt: '%s' using model %s from provider %s", 
		config.Prompt, config.ModelName, config.Provider)
	
	return response, nil
}

// NewTextGeneratorNode creates a new text generator node
func NewTextGeneratorNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	// Convert config map to struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var tgConfig TextGeneratorConfig
	if err := json.Unmarshal(jsonData, &tgConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set defaults
	if tgConfig.Temperature == 0 {
		tgConfig.Temperature = 0.7 // Default creativity
	}
	
	if tgConfig.MaxTokens == 0 {
		tgConfig.MaxTokens = 512 // Default token count
	}
	
	if tgConfig.Timeout == 0 {
		tgConfig.Timeout = 30 // Default timeout (30 seconds)
	}
	
	if tgConfig.CacheTTL == 0 {
		tgConfig.CacheTTL = 3600 // Default cache TTL (1 hour)
	}

	// Create AI manager
	aiManager := NewAIManager()

	return &TextGeneratorNode{
		config:    &tgConfig,
		aiManager: aiManager,
	}, nil
}

// Execute executes the text generation node
func (tg *TextGeneratorNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	startTime := time.Now()

	// Override config values with inputs if provided
	prompt := tg.config.Prompt
	if inputPrompt, exists := inputs["prompt"]; exists {
		if promptStr, ok := inputPrompt.(string); ok && promptStr != "" {
			prompt = promptStr
		}
	}

	modelName := tg.config.ModelName
	if inputModel, exists := inputs["model_name"]; exists {
		if modelStr, ok := inputModel.(string); ok && modelStr != "" {
			modelName = modelStr
		}
	}

	systemPrompt := tg.config.SystemPrompt
	if inputSystemPrompt, exists := inputs["system_prompt"]; exists {
		if sysPromptStr, ok := inputSystemPrompt.(string); ok {
			systemPrompt = sysPromptStr
		}
	}

	temperature := tg.config.Temperature
	if inputTemp, exists := inputs["temperature"]; exists {
		if tempFloat, ok := inputTemp.(float64); ok {
			temperature = tempFloat
		}
	}

	maxTokens := tg.config.MaxTokens
	if inputMaxTokens, exists := inputs["max_tokens"]; exists {
		if maxTokFloat, ok := inputMaxTokens.(float64); ok {
			maxTokens = int(maxTokFloat)
		}
	}

	// Prepare config for execution
	execConfig := &TextGeneratorConfig{
		ModelName:    modelName,
		Provider:     tg.config.Provider,
		ApiKey:       tg.config.ApiKey,
		Temperature:  temperature,
		MaxTokens:    maxTokens,
		Prompt:       prompt,
		SystemPrompt: systemPrompt,
		Timeout:      tg.config.Timeout,
		Parameters:   tg.config.Parameters,
	}

	// Run the AI operation
	result, err := tg.aiManager.GenerateText(execConfig)
	if err != nil {
		return map[string]interface{}{
			"success":   false,
			"error":     err.Error(),
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Prepare response
	response := map[string]interface{}{
		"success":         true,
		"generated_text":  result,
		"model":           execConfig.ModelName,
		"provider":        execConfig.Provider,
		"temperature":     execConfig.Temperature,
		"max_tokens":      execConfig.MaxTokens,
		"prompt_used":     execConfig.Prompt,
		"system_prompt":   execConfig.SystemPrompt,
		"execution_time":  time.Since(startTime).Seconds(),
		"timestamp":       time.Now().Unix(),
		"input_data":      inputs,
		"config_used":     execConfig,
	}

	// Add profiling data if enabled
	if tg.config.EnableProfiling {
		response["profiling"] = map[string]interface{}{
			"start_time": startTime.Unix(),
			"end_time":   time.Now().Unix(),
			"duration":   time.Since(startTime).Seconds(),
			"model":      execConfig.ModelName,
			"provider":   execConfig.Provider,
		}
	}

	return response, nil
}

// GetType returns the type of the node
func (tg *TextGeneratorNode) GetType() string {
	return "ai_text_generator"
}

// GetID returns the unique ID of the node instance
func (tg *TextGeneratorNode) GetID() string {
	return fmt.Sprintf("ai_tg_%s_%d", tg.config.ModelName, time.Now().Unix())
}