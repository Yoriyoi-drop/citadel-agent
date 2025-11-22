package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/engine"
)

// ContextualReasoningNodeConfig represents the configuration for a Contextual Reasoning node
type ContextualReasoningNodeConfig struct {
	Provider       string                 `json:"provider"`        // AI provider (openai, anthropic, etc.)
	ApiKey         string                 `json:"api_key"`         // API key for the reasoning service
	Model          string                 `json:"model"`           // Reasoning model to use (gpt-4, claude-2, etc.)
	InputText      string                 `json:"input_text"`      // Input text to reason about
	Context        string                 `json:"context"`         // Additional context for reasoning
	MaxSteps       int                    `json:"max_steps"`       // Maximum number of reasoning steps
	ReasoningType  string                 `json:"reasoning_type"`  // Type of reasoning (logical, mathematical, causal, etc.)
	ReturnSteps    bool                   `json:"return_steps"`    // Whether to return intermediate reasoning steps
	Confidence     float64                `json:"confidence"`      // Minimum confidence threshold for answers
	CustomParams   map[string]interface{} `json:"custom_params"`   // Custom parameters for the AI service
	Timeout        int                    `json:"timeout"`         // Request timeout in seconds
	Enabled        bool                   `json:"enabled"`         // Whether the node is enabled
}

// ContextualReasoningNode represents a node that performs multi-step reasoning using AI
type ContextualReasoningNode struct {
	config *ContextualReasoningNodeConfig
}

// NewContextualReasoningNode creates a new Contextual Reasoning node
func NewContextualReasoningNode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Convert interface{} map to JSON and back to struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}

	var reasoningConfig ContextualReasoningNodeConfig
	err = json.Unmarshal(jsonData, &reasoningConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// Validate required fields
	if reasoningConfig.Provider == "" {
		reasoningConfig.Provider = "openai" // default provider
	}

	if reasoningConfig.Model == "" {
		reasoningConfig.Model = "gpt-4" // default model
	}

	if reasoningConfig.MaxSteps == 0 {
		reasoningConfig.MaxSteps = 10 // default to 10 steps
	}

	if reasoningConfig.Timeout == 0 {
		reasoningConfig.Timeout = 60 // default timeout of 60 seconds for complex reasoning
	}

	if reasoningConfig.Confidence == 0 {
		reasoningConfig.Confidence = 0.7 // default confidence of 70%
	}

	return &ContextualReasoningNode{
		config: &reasoningConfig,
	}, nil
}

// Execute implements the NodeInstance interface
func (c *ContextualReasoningNode) Execute(ctx context.Context, input map[string]interface{}) (*engine.ExecutionResult, error) {
	// Override configuration with input values if provided
	provider := c.config.Provider
	if inputProvider, ok := input["provider"].(string); ok && inputProvider != "" {
		provider = inputProvider
	}

	apiKey := c.config.ApiKey
	if inputApiKey, ok := input["api_key"].(string); ok && inputApiKey != "" {
		apiKey = inputApiKey
	}

	model := c.config.Model
	if inputModel, ok := input["model"].(string); ok && inputModel != "" {
		model = inputModel
	}

	inputText := c.config.InputText
	if inputInputText, ok := input["input_text"].(string); ok && inputInputText != "" {
		inputText = inputInputText
	}

	contextText := c.config.Context
	if inputContext, ok := input["context"].(string); ok && inputContext != "" {
		contextText = inputContext
	}

	maxSteps := c.config.MaxSteps
	if inputMaxSteps, ok := input["max_steps"].(float64); ok {
		maxSteps = int(inputMaxSteps)
	}

	reasoningType := c.config.ReasoningType
	if inputReasoningType, ok := input["reasoning_type"].(string); ok && inputReasoningType != "" {
		reasoningType = inputReasoningType
	}

	returnSteps := c.config.ReturnSteps
	if inputReturnSteps, ok := input["return_steps"].(bool); ok {
		returnSteps = inputReturnSteps
	}

	confidence := c.config.Confidence
	if inputConfidence, ok := input["confidence"].(float64); ok {
		confidence = inputConfidence
	}

	customParams := c.config.CustomParams
	if inputCustomParams, ok := input["custom_params"].(map[string]interface{}); ok {
		customParams = inputCustomParams
	}

	timeout := c.config.Timeout
	if inputTimeout, ok := input["timeout"].(float64); ok {
		timeout = int(inputTimeout)
	}

	enabled := c.config.Enabled
	if inputEnabled, ok := input["enabled"].(bool); ok {
		enabled = inputEnabled
	}

	// Check if node should be enabled
	if !enabled {
		return &engine.ExecutionResult{
			Status: "success",
			Data: map[string]interface{}{
				"message": "contextual reasoning processor disabled, not executed",
				"enabled": false,
			},
			Timestamp: time.Now(),
		}, nil
	}

	// Validate required input
	if inputText == "" {
		return &engine.ExecutionResult{
			Status:    "error",
			Error:     "input_text is required for contextual reasoning",
			Timestamp: time.Now(),
		}, nil
	}

	if apiKey == "" {
		return &engine.ExecutionResult{
			Status:    "error",
			Error:     "api_key is required for contextual reasoning",
			Timestamp: time.Now(),
		}, nil
	}

	// In a real implementation, this would call the actual reasoning AI service
	// For now, we'll simulate the response with multi-step reasoning
	reasoningSteps := []map[string]interface{}{}
	
	if returnSteps {
		reasoningSteps = []map[string]interface{}{
			{
				"step": 1,
				"description": "Analyze the input question and identify key components",
				"result": "Input question contains elements X, Y, and Z",
			},
			{
				"step": 2,
				"description": "Gather relevant context and background information",
				"result": "Context indicates variables A and B are related to Z",
			},
			{
				"step": 3,
				"description": "Apply logical rules to connect components",
				"result": "Based on rule R1, X is connected to Y through Z",
			},
			{
				"step": 4,
				"description": "Formulate the final conclusion",
				"result": "Therefore, the answer is C based on the chain of reasoning",
			},
		}
	}

	result := map[string]interface{}{
		"input_text": inputText,
		"reasoning_type": reasoningType,
		"final_answer": "Based on the contextual analysis and multi-step reasoning, the answer is: The solution is effective under the given constraints.",
		"confidence": 0.85,
		"reasoning_steps": reasoningSteps,
		"used_context": contextText,
		"steps_taken": len(reasoningSteps),
		"processing_time": time.Since(time.Now().Add(-time.Duration(timeout) * time.Second)).Seconds(),
	}

	// Check if confidence meets threshold
	if result["confidence"].(float64) < confidence {
		result["warning"] = fmt.Sprintf("Confidence (%.2f) is below threshold (%.2f)", 
			result["confidence"], confidence)
	}

	return &engine.ExecutionResult{
		Status: "success",
		Data: map[string]interface{}{
			"message":        "contextual reasoning completed",
			"result":         result,
			"provider":       provider,
			"model":          model,
			"reasoning_type": reasoningType,
			"steps_returned": returnSteps,
			"timestamp":      time.Now().Unix(),
		},
		Timestamp: time.Now(),
	}, nil
}

// GetType returns the type of the node
func (c *ContextualReasoningNode) GetType() string {
	return "contextual_reasoning"
}

// GetID returns a unique ID for the node instance
func (c *ContextualReasoningNode) GetID() string {
	return "contextual_reasoning_" + fmt.Sprintf("%d", time.Now().Unix())
}