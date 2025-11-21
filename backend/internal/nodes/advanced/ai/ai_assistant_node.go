package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/engine"
)

// AIAssistantNode represents an AI-powered node for natural language processing
type AIAssistantNode struct {
	Model       string                 `json:"model"`
	Prompt      string                 `json:"prompt"`
	MaxTokens   int                    `json:"max_tokens"`
	Temperature float64                `json:"temperature"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// Execute executes the AI assistant node
func (ai *AIAssistantNode) Execute(ctx context.Context, input map[string]interface{}) (*engine.ExecutionResult, error) {
	// Extract prompt from settings or input
	prompt := ai.Prompt
	if promptInput, exists := input["prompt"].(string); exists && promptInput != "" {
		prompt = promptInput
	} else if promptInput, exists := input["input"].(string); exists {
		prompt = promptInput
	}

	if prompt == "" {
		return &engine.ExecutionResult{
			Status:    "error",
			Error:     "Prompt is required for AI assistant node",
			Timestamp: time.Now(),
		}, nil
	}

	// Simulate AI processing time
	time.Sleep(500 * time.Millisecond)

	// For now, simulate AI response - in real implementation this would call an AI API
	simulatedResponse := fmt.Sprintf("AI Response to: %s", prompt)
	
	// Add context from input
	extendedContext, _ := json.Marshal(input)
	
	result := map[string]interface{}{
		"response":       simulatedResponse,
		"input_context":  string(extendedContext),
		"model":          ai.Model,
		"processing_time": 500 * time.Millisecond,
		"tokens_used":    len(prompt) * 4, // Very rough estimation
	}

	return &engine.ExecutionResult{
		Status:    "success",
		Data:      result,
		Timestamp: time.Now(),
	}, nil
}

// Validate ensures the AI assistant node is configured correctly
func (ai *AIAssistantNode) Validate() error {
	if ai.Model == "" {
		return fmt.Errorf("AI model is required")
	}
	
	if ai.MaxTokens < 0 {
		return fmt.Errorf("MaxTokens must be non-negative")
	}
	
	if ai.Temperature < 0 || ai.Temperature > 2 {
		return fmt.Errorf("Temperature must be between 0 and 2")
	}

	return nil
}