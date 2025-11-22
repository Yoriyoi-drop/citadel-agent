package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
	"github.com/citadel-agent/backend/internal/utils"
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
	provider := utils.GetString(config["provider"], "openai")
	apiKey := utils.GetString(config["api_key"], "")
	model := utils.GetString(config["model"], "gpt-4-vision-preview")
	imageURL := utils.GetString(config["image_url"], "")
	imageData := utils.GetString(config["image_data"], "")

	maxResults := utils.GetInt(config["max_results"], 10)
	confidence := utils.GetFloat64(config["confidence"], 0.7)
	timeout := utils.GetInt(config["timeout"], 60)
	enabled := utils.GetBool(config["enabled"], true)
	
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

	timeout := v.config.Timeout
	if inputTimeout, exists := input["timeout"]; exists {
		if inputTimeoutFloat, ok := inputTimeout.(float64); ok {
			timeout = int(inputTimeoutFloat)
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
			"error":   "either image_url or image_data is required for vision processing",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	if apiKey == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "api_key is required for vision processing",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// In a real implementation, this would call the actual vision AI service
	// For now, we'll simulate the response
	result := map[string]interface{}{
		"image_processed":   imageURL,
		"image_type":        imageData, // For simulation purposes
		"objects_detected": []map[string]interface{}{
			{"name": "person", "confidence": 0.95, "bounding_box": map[string]interface{}{"x": 10, "y": 20, "width": 100, "height": 150}},
			{"name": "car", "confidence": 0.88, "bounding_box": map[string]interface{}{"x": 200, "y": 150, "width": 80, "height": 40}},
		},
		"text_recognized": []string{"SAMPLE TEXT", "ON IMAGE"},
		"image_description": "A sample image with person and car",
		"confidence_threshold": confidence,
		"analysis_types": analysisTypes,
		"max_results": maxResults,
		"processing_time": time.Since(time.Now().Add(-time.Duration(timeout) * time.Second)).Seconds(),
	}

	return map[string]interface{}{
		"success": true,
		"message":        "vision AI processing completed",
		"result":         result,
		"provider":       provider,
		"model":          model,
		"analysis_types": analysisTypes,
		"timestamp":      time.Now().Unix(),
	}, nil
}