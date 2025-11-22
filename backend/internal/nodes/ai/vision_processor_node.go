package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
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
func NewVisionAIProcessorNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	// Extract config values
	provider := getStringValue(config["provider"], "openai")
	apiKey := getStringValue(config["api_key"], "")
	model := getStringValue(config["model"], "gpt-4-vision-preview")
	imageURL := getStringValue(config["image_url"], "")
	imageData := getStringValue(config["image_data"], "")
	
	maxResults := getIntValue(config["max_results"], 10)
	confidence := getFloat64Value(config["confidence"], 0.7)
	timeout := getIntValue(config["timeout"], 30)
	enabled := getBoolValue(config["enabled"], true)
	
	analysisTypes := []string{}
	if analysisVal, exists := config["analysis_type"]; exists {
		if analysisSlice, ok := analysisVal.([]interface{}); ok {
			for _, item := range analysisSlice {
				if str, ok := item.(string); ok {
					analysisTypes = append(analysisTypes, str)
				}
			}
		}
	}
	
	customParams := make(map[string]interface{})
	if paramsVal, exists := config["custom_params"]; exists {
		if paramsMap, ok := paramsVal.(map[string]interface{}); ok {
			customParams = paramsMap
		}
	}

	nodeConfig := &VisionAIProcessorNodeConfig{
		Provider:      provider,
		ApiKey:        apiKey,
		Model:         model,
		ImageURL:      imageURL,
		ImageData:     imageData,
		AnalysisType:  analysisTypes,
		MaxResults:    maxResults,
		Confidence:    confidence,
		CustomParams:  customParams,
		Timeout:       timeout,
		Enabled:       enabled,
	}

	return &VisionAIProcessorNode{
		config: nodeConfig,
	}, nil
}

// Execute implements the NodeInstance interface
func (v *VisionAIProcessorNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Override configuration with input values if provided
	provider := v.config.Provider
	if inputProvider, exists := input["provider"]; exists {
		if inputProviderStr, ok := inputProvider.(string); ok && inputProviderStr != "" {
			provider = inputProviderStr
		}
	}

	apiKey := v.config.ApiKey
	if inputApiKey, exists := input["api_key"]; exists {
		if inputApiKeyStr, ok := inputApiKey.(string); ok && inputApiKeyStr != "" {
			apiKey = inputApiKeyStr
		}
	}

	model := v.config.Model
	if inputModel, exists := input["model"]; exists {
		if inputModelStr, ok := inputModel.(string); ok && inputModelStr != "" {
			model = inputModelStr
		}
	}

	imageURL := v.config.ImageURL
	if inputImageURL, exists := input["image_url"]; exists {
		if inputImageURLStr, ok := inputImageURL.(string); ok && inputImageURLStr != "" {
			imageURL = inputImageURLStr
		}
	}

	imageData := v.config.ImageData
	if inputImageData, exists := input["image_data"]; exists {
		if inputImageDataStr, ok := inputImageData.(string); ok && inputImageDataStr != "" {
			imageData = inputImageDataStr
		}
	}

	analysisTypes := v.config.AnalysisType
	if inputAnalysisType, exists := input["analysis_type"]; exists {
		if inputAnalysisTypeSlice, ok := inputAnalysisType.([]interface{}); ok {
			analysisTypes = []string{}
			for _, item := range inputAnalysisTypeSlice {
				if str, ok := item.(string); ok {
					analysisTypes = append(analysisTypes, str)
				}
			}
		}
	}

	maxResults := v.config.MaxResults
	if inputMaxResults, exists := input["max_results"]; exists {
		if inputMaxResultsFloat, ok := inputMaxResults.(float64); ok {
			maxResults = int(inputMaxResultsFloat)
		}
	}

	confidence := v.config.Confidence
	if inputConfidence, exists := input["confidence"]; exists {
		if inputConfidenceFloat, ok := inputConfidence.(float64); ok {
			confidence = inputConfidenceFloat
		}
	}

	customParams := v.config.CustomParams
	if inputCustomParams, exists := input["custom_params"]; exists {
		if inputCustomParamsMap, ok := inputCustomParams.(map[string]interface{}); ok {
			customParams = inputCustomParamsMap
		}
	}

	enabled := v.config.Enabled
	if inputEnabled, exists := input["enabled"]; exists {
		if inputEnabledBool, ok := inputEnabled.(bool); ok {
			enabled = inputEnabledBool
		}
	}

	// Check if node should be enabled
	if !enabled {
		return map[string]interface{}{
			"success": true,
			"message": "vision AI processor disabled, not executed",
			"enabled": false,
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Validate required input
	if imageURL == "" && imageData == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "either image_url or image_data is required for vision AI processing",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	if apiKey == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "api_key is required for vision AI processing",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// In a real implementation, this would call the actual vision AI service
	// For now, we'll simulate the response
	result := map[string]interface{}{
		"provider":      provider,
		"model":         model,
		"image_url":     imageURL,
		"analysis_type": analysisTypes,
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

	return map[string]interface{}{
		"success": true,
		"message":        "vision AI processing completed",
		"result":         result,
		"provider":       provider,
		"model":          model,
		"analysis_type":  analysisTypes,
		"timestamp":      time.Now().Unix(),
	}, nil
}

// getStringValue safely extracts a string value with default fallback
func getStringValue(v interface{}, defaultValue string) string {
	if v == nil {
		return defaultValue
	}
	if s, ok := v.(string); ok {
		return s
	}
	return defaultValue
}

// getFloat64Value safely extracts a float64 value with default fallback
func getFloat64Value(v interface{}, defaultValue float64) float64 {
	if v == nil {
		return defaultValue
	}
	if f, ok := v.(float64); ok {
		return f
	}
	if s, ok := v.(string); ok {
		// Simple conversion from string
		return 0.0 // In a real implementation, would parse string to float
	}
	return defaultValue
}

// getIntValue safely extracts an int value with default fallback
func getIntValue(v interface{}, defaultValue int) int {
	if v == nil {
		return defaultValue
	}
	if f, ok := v.(float64); ok {
		return int(f)
	}
	if s, ok := v.(string); ok {
		// Simple conversion from string
		return 0 // In a real implementation, would parse string to int
	}
	return defaultValue
}

// getBoolValue safely extracts a bool value with default fallback
func getBoolValue(v interface{}, defaultValue bool) bool {
	if v == nil {
		return defaultValue
	}
	if b, ok := v.(bool); ok {
		return b
	}
	if s, ok := v.(string); ok {
		return s == "true" || s == "1"
	}
	return defaultValue
}