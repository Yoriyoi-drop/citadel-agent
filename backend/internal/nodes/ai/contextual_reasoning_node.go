package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
	"github.com/citadel-agent/backend/internal/nodes/utils"
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
func NewContextualReasoningNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	// Extract config values
	provider := utils.GetStringVal(config["provider"], "openai")
	apiKey := utils.GetStringVal(config["api_key"], "")
	model := utils.GetStringVal(config["model"], "gpt-4")
	inputText := utils.GetStringVal(config["input_text"], "")
	contextText := utils.GetStringVal(config["context"], "")
	reasoningType := utils.GetStringVal(config["reasoning_type"], "")

	maxSteps := utils.GetIntVal(config["max_steps"], 10)
	returnSteps := utils.GetBoolVal(config["return_steps"], false)
	confidence := utils.GetFloat64Val(config["confidence"], 0.7)
	timeout := utils.GetIntVal(config["timeout"], 60)
	enabled := utils.GetBoolVal(config["enabled"], true)

	customParams := make(map[string]interface{})
	if paramsVal, exists := config["custom_params"]; exists {
		if paramsMap, ok := paramsVal.(map[string]interface{}); ok {
			customParams = paramsMap
		}
	}

	nodeConfig := &ContextualReasoningNodeConfig{
		Provider:       provider,
		ApiKey:         apiKey,
		Model:          model,
		InputText:      inputText,
		Context:        contextText,
		MaxSteps:       maxSteps,
		ReasoningType:  reasoningType,
		ReturnSteps:    returnSteps,
		Confidence:     confidence,
		CustomParams:   customParams,
		Timeout:        timeout,
		Enabled:        enabled,
	}

	return &ContextualReasoningNode{
		config: nodeConfig,
	}, nil
}

// Execute implements the NodeInstance interface
func (c *ContextualReasoningNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Override configuration with input values if provided
	provider := c.config.Provider
	if inputProvider, exists := input["provider"]; exists {
		if inputProviderStr, ok := inputProvider.(string); ok && inputProviderStr != "" {
			provider = inputProviderStr
		}
	}

	apiKey := c.config.ApiKey
	if inputApiKey, exists := input["api_key"]; exists {
		if inputApiKeyStr, ok := inputApiKey.(string); ok && inputApiKeyStr != "" {
			apiKey = inputApiKeyStr
		}
	}

	model := c.config.Model
	if inputModel, exists := input["model"]; exists {
		if inputModelStr, ok := inputModel.(string); ok && inputModelStr != "" {
			model = inputModelStr
		}
	}

	inputText := c.config.InputText
	if inputInputText, exists := input["input_text"]; exists {
		if inputInputTextStr, ok := inputInputText.(string); ok && inputInputTextStr != "" {
			inputText = inputInputTextStr
		}
	}

	contextText := c.config.Context
	if inputContext, exists := input["context"]; exists {
		if inputContextStr, ok := inputContext.(string); ok && inputContextStr != "" {
			contextText = inputContextStr
		}
	}

	maxSteps := c.config.MaxSteps
	if inputMaxSteps, exists := input["max_steps"]; exists {
		if inputMaxStepsFloat, ok := inputMaxSteps.(float64); ok {
			maxSteps = int(inputMaxStepsFloat)
		}
	}

	reasoningType := c.config.ReasoningType
	if inputReasoningType, exists := input["reasoning_type"]; exists {
		if inputReasoningTypeStr, ok := inputReasoningType.(string); ok && inputReasoningTypeStr != "" {
			reasoningType = inputReasoningTypeStr
		}
	}

	returnSteps := c.config.ReturnSteps
	if inputReturnSteps, exists := input["return_steps"]; exists {
		if inputReturnStepsBool, ok := inputReturnSteps.(bool); ok {
			returnSteps = inputReturnStepsBool
		}
	}

	confidence := c.config.Confidence
	if inputConfidence, exists := input["confidence"]; exists {
		if inputConfidenceFloat, ok := inputConfidence.(float64); ok {
			confidence = inputConfidenceFloat
		}
	}

	customParams := c.config.CustomParams
	if inputCustomParams, exists := input["custom_params"]; exists {
		if inputCustomParamsMap, ok := inputCustomParams.(map[string]interface{}); ok {
			customParams = inputCustomParamsMap
		}
	}

	timeout := c.config.Timeout
	if inputTimeout, exists := input["timeout"]; exists {
		if inputTimeoutFloat, ok := inputTimeout.(float64); ok {
			timeout = int(inputTimeoutFloat)
		}
	}

	enabled := c.config.Enabled
	if inputEnabled, exists := input["enabled"]; exists {
		if inputEnabledBool, ok := inputEnabled.(bool); ok {
			enabled = inputEnabledBool
		}
	}

	// Check if node should be enabled
	if !enabled {
		return map[string]interface{}{
			"success": true,
			"message": "contextual reasoning processor disabled, not executed",
			"enabled": false,
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Validate required input
	if inputText == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "input_text is required for contextual reasoning",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	if apiKey == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "api_key is required for contextual reasoning",
			"timestamp": time.Now().Unix(),
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
	if resultConfidence, ok := result["confidence"].(float64); ok && resultConfidence < confidence {
		result["warning"] = fmt.Sprintf("Confidence (%.2f) is below threshold (%.2f)",
			resultConfidence, confidence)
	}

	return map[string]interface{}{
		"success": true,
		"message":        "contextual reasoning completed",
		"result":         result,
		"provider":       provider,
		"model":          model,
		"reasoning_type": reasoningType,
		"steps_returned": returnSteps,
		"timestamp":      time.Now().Unix(),
	}, nil
}


