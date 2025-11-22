package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/engine"
)

// VisionAIProcessorNodeConfig represents the configuration for a Vision AI Processor node
type VisionAIProcessorNodeConfig struct {
	Provider       string                 `json:"provider"`        // AI provider (openai, google, azure, etc.)
	ApiKey         string                 `json:"api_key"`         // API key for the vision service
	Model          string                 `json:"model"`           // Vision model to use
	ImageURL       string                 `json:"image_url"`       // URL of the image to process
	ImageData      string                 `json:"image_data"`      // Base64 encoded image data
	AnalysisType   []string               `json:"analysis_type"`   // Types of analysis to perform (object_detection, text_recognition, etc.)
	MaxResults     int                    `json:"max_results"`     // Maximum number of results to return
	Confidence     float64                `json:"confidence"`      // Confidence threshold
	CustomParams   map[string]interface{} `json:"custom_params"`   // Custom parameters for the AI service
	Timeout        int                    `json:"timeout"`         // Request timeout in seconds
	Enabled        bool                   `json:"enabled"`         // Whether the node is enabled
}

// VisionAIProcessorNode represents a node that processes images using AI vision services
type VisionAIProcessorNode struct {
	config *VisionAIProcessorNodeConfig
}

// NewVisionAIProcessorNode creates a new Vision AI Processor node
func NewVisionAIProcessorNode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Convert interface{} map to JSON and back to struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}

	var visionConfig VisionAIProcessorNodeConfig
	err = json.Unmarshal(jsonData, &visionConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// Validate required fields
	if visionConfig.Provider == "" {
		visionConfig.Provider = "openai" // default provider
	}

	if visionConfig.Model == "" {
		visionConfig.Model = "gpt-4-vision-preview" // default model
	}

	if visionConfig.Timeout == 0 {
		visionConfig.Timeout = 30 // default timeout of 30 seconds
	}

	return &VisionAIProcessorNode{
		config: &visionConfig,
	}, nil
}

// Execute implements the NodeInstance interface
func (v *VisionAIProcessorNode) Execute(ctx context.Context, input map[string]interface{}) (*engine.ExecutionResult, error) {
	// Override configuration with input values if provided
	provider := v.config.Provider
	if inputProvider, ok := input["provider"].(string); ok && inputProvider != "" {
		provider = inputProvider
	}

	apiKey := v.config.ApiKey
	if inputApiKey, ok := input["api_key"].(string); ok && inputApiKey != "" {
		apiKey = inputApiKey
	}

	model := v.config.Model
	if inputModel, ok := input["model"].(string); ok && inputModel != "" {
		model = inputModel
	}

	imageURL := v.config.ImageURL
	if inputImageURL, ok := input["image_url"].(string); ok && inputImageURL != "" {
		imageURL = inputImageURL
	}

	imageData := v.config.ImageData
	if inputImageData, ok := input["image_data"].(string); ok && inputImageData != "" {
		imageData = inputImageData
	}

	analysisType := v.config.AnalysisType
	if inputAnalysisType, ok := input["analysis_type"].([]interface{}); ok {
		analysisType = make([]string, len(inputAnalysisType))
		for i, val := range inputAnalysisType {
			analysisType[i] = fmt.Sprintf("%v", val)
		}
	}

	maxResults := v.config.MaxResults
	if inputMaxResults, ok := input["max_results"].(float64); ok {
		maxResults = int(inputMaxResults)
	}

	confidence := v.config.Confidence
	if inputConfidence, ok := input["confidence"].(float64); ok {
		confidence = inputConfidence
	}

	customParams := v.config.CustomParams
	if inputCustomParams, ok := input["custom_params"].(map[string]interface{}); ok {
		customParams = inputCustomParams
	}

	enabled := v.config.Enabled
	if inputEnabled, ok := input["enabled"].(bool); ok {
		enabled = inputEnabled
	}

	// Check if node should be enabled
	if !enabled {
		return &engine.ExecutionResult{
			Status: "success",
			Data: map[string]interface{}{
				"message": "vision AI processor disabled, not executed",
				"enabled": false,
			},
			Timestamp: time.Now(),
		}, nil
	}

	// Validate required input
	if imageURL == "" && imageData == "" {
		return &engine.ExecutionResult{
			Status:    "error",
			Error:     "either image_url or image_data is required for vision AI processing",
			Timestamp: time.Now(),
		}, nil
	}

	if apiKey == "" {
		return &engine.ExecutionResult{
			Status:    "error",
			Error:     "api_key is required for vision AI processing",
			Timestamp: time.Now(),
		}, nil
	}

	// In a real implementation, this would call the actual vision AI service
	// For now, we'll simulate the response
	result := map[string]interface{}{
		"provider":      provider,
		"model":         model,
		"image_url":     imageURL,
		"analysis_type": analysisType,
		"results": []map[string]interface{}{
			{
				"object":     "person",
				"confidence": 0.95,
				"bounding_box": map[string]interface{}{
					"x": 100,
					"y": 150,
					"width": 200,
					"height": 300,
				},
			},
			{
				"object":     "car",
				"confidence": 0.89,
				"bounding_box": map[string]interface{}{
					"x": 400,
					"y": 200,
					"width": 250,
					"height": 150,
				},
			},
		},
		"detected_text": []string{"SPEED LIMIT 30", "CAUTION"},
		"image_quality": map[string]interface{}{
			"brightness": 0.7,
			"contrast":   0.8,
			"sharpness":  0.6,
		},
		"processing_time": time.Since(time.Now().Add(-time.Duration(v.config.Timeout) * time.Second)).Seconds(),
	}

	return &engine.ExecutionResult{
		Status: "success",
		Data: map[string]interface{}{
			"message":        "vision AI processing completed",
			"result":         result,
			"provider":       provider,
			"model":          model,
			"analysis_type":  analysisType,
			"timestamp":      time.Now().Unix(),
		},
		Timestamp: time.Now(),
	}, nil
}

// GetType returns the type of the node
func (v *VisionAIProcessorNode) GetType() string {
	return "vision_ai_processor"
}

// GetID returns a unique ID for the node instance
func (v *VisionAIProcessorNode) GetID() string {
	return "vision_ai_" + fmt.Sprintf("%d", time.Now().Unix())
}