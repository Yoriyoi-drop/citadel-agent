package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
	"github.com/citadel-agent/backend/internal/nodes/utils"
)

// PredictionModelNodeConfig represents the configuration for a Prediction Model node
type PredictionModelNodeConfig struct {
	Provider       string                 `json:"provider"`        // AI provider (openai, google, huggingface, etc.)
	ApiKey         string                 `json:"api_key"`         // API key for the prediction service
	Model          string                 `json:"model"`           // Prediction model to use
	Features       []interface{}          `json:"features"`        // Input features for prediction
	FeatureNames   []string               `json:"feature_names"`   // Names of the input features
	PredictionType string                 `json:"prediction_type"` // Type of prediction (classification, regression, forecasting, etc.)
	ModelEndpoint  string                 `json:"model_endpoint"`  // Endpoint of the prediction model
	Confidence     float64                `json:"confidence"`      // Required confidence threshold
	ReturnFeatures bool                   `json:"return_features"` // Whether to return input features in the result
	CustomParams   map[string]interface{} `json:"custom_params"`   // Custom parameters for the prediction model
	Timeout        int                    `json:"timeout"`         // Request timeout in seconds
	Enabled        bool                   `json:"enabled"`         // Whether the node is enabled
}

// PredictionModelNode represents a node that makes predictions using AI models
type PredictionModelNode struct {
	config *PredictionModelNodeConfig
}

// NewPredictionModelNode creates a new Prediction Model node
func NewPredictionModelNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	// Extract config values
	provider := utils.GetStringVal(config["provider"], "huggingface")
	apiKey := utils.GetStringVal(config["api_key"], "")
	model := utils.GetStringVal(config["model"], "microsoft/regnet-600mf")
	predictionType := utils.GetStringVal(config["prediction_type"], "classification")
	modelEndpoint := utils.GetStringVal(config["model_endpoint"], "")

	features := make([]interface{}, 0)
	if featuresVal, exists := config["features"]; exists {
		if featuresSlice, ok := featuresVal.([]interface{}); ok {
			features = featuresSlice
		}
	}

	featureNames := make([]string, 0)
	if namesVal, exists := config["feature_names"]; exists {
		if namesSlice, ok := namesVal.([]interface{}); ok {
			featureNames = make([]string, len(namesSlice))
			for i, name := range namesSlice {
				featureNames[i] = fmt.Sprintf("%v", name)
			}
		}
	}

	confidence := getFloat64Value(config["confidence"], 0.7)
	customParams := make(map[string]interface{})
	if paramsVal, exists := config["custom_params"]; exists {
		if paramsMap, ok := paramsVal.(map[string]interface{}); ok {
			customParams = paramsMap
		}
	}
	timeout := getIntValue(config["timeout"], 30)
	enabled := getBoolValue(config["enabled"], true)

	returnFeatures := false
	if val, exists := config["return_features"]; exists {
		if b, ok := val.(bool); ok {
			returnFeatures = b
		}
	}

	nodeConfig := &PredictionModelNodeConfig{
		Provider:       provider,
		ApiKey:         apiKey,
		Model:          model,
		Features:       features,
		FeatureNames:   featureNames,
		PredictionType: predictionType,
		ModelEndpoint:  modelEndpoint,
		Confidence:     confidence,
		ReturnFeatures: returnFeatures,
		CustomParams:   customParams,
		Timeout:        timeout,
		Enabled:        enabled,
	}

	return &PredictionModelNode{
		config: nodeConfig,
	}, nil
}

// Execute implements the NodeInstance interface
func (p PredictionModelNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Override configuration with input values if provided
	provider := p.config.Provider
	if inputProvider, exists := input["provider"]; exists {
		if inputProviderStr, ok := inputProvider.(string); ok && inputProviderStr != "" {
			provider = inputProviderStr
		}
	}

	apiKey := p.config.ApiKey
	if inputApiKey, exists := input["api_key"]; exists {
		if inputApiKeyStr, ok := inputApiKey.(string); ok && inputApiKeyStr != "" {
			apiKey = inputApiKeyStr
		}
	}

	model := p.config.Model
	if inputModel, exists := input["model"]; exists {
		if inputModelStr, ok := inputModel.(string); ok && inputModelStr != "" {
			model = inputModelStr
		}
	}

	features := p.config.Features
	if inputFeatures, exists := input["features"]; exists {
		if inputFeaturesSlice, ok := inputFeatures.([]interface{}); ok {
			features = inputFeaturesSlice
		}
	}

	featureNames := p.config.FeatureNames
	if inputFeatureNames, exists := input["feature_names"]; exists {
		if inputFeatureNamesSlice, ok := inputFeatureNames.([]interface{}); ok {
			featureNames = make([]string, len(inputFeatureNamesSlice))
			for i, name := range inputFeatureNamesSlice {
				featureNames[i] = fmt.Sprintf("%v", name)
			}
		}
	}

	predictionType := p.config.PredictionType
	if inputPredictionType, exists := input["prediction_type"]; exists {
		if inputPredictionTypeStr, ok := inputPredictionType.(string); ok && inputPredictionTypeStr != "" {
			predictionType = inputPredictionTypeStr
		}
	}

	modelEndpoint := p.config.ModelEndpoint
	if inputEndpoint, exists := input["model_endpoint"]; exists {
		if inputEndpointStr, ok := inputEndpoint.(string); ok && inputEndpointStr != "" {
			modelEndpoint = inputEndpointStr
		}
	}

	confidence := p.config.Confidence
	if inputConfidence, exists := input["confidence"]; exists {
		if inputConfidenceFloat, ok := inputConfidence.(float64); ok {
			confidence = inputConfidenceFloat
		}
	}

	returnFeatures := p.config.ReturnFeatures
	if inputReturnFeatures, exists := input["return_features"]; exists {
		if inputReturnFeaturesBool, ok := inputReturnFeatures.(bool); ok {
			returnFeatures = inputReturnFeaturesBool
		}
	}

	customParams := p.config.CustomParams
	if inputCustomParams, exists := input["custom_params"]; exists {
		if inputCustomParamsMap, ok := inputCustomParams.(map[string]interface{}); ok {
			customParams = inputCustomParamsMap
		}
	}

	timeout := p.config.Timeout
	if inputTimeout, exists := input["timeout"]; exists {
		if inputTimeoutFloat, ok := inputTimeout.(float64); ok {
			timeout = int(inputTimeoutFloat)
		}
	}

	enabled := p.config.Enabled
	if inputEnabled, exists := input["enabled"]; exists {
		if inputEnabledBool, ok := inputEnabled.(bool); ok {
			enabled = inputEnabledBool
		}
	}

	// Check if node should be enabled
	if !enabled {
		return map[string]interface{}{
			"success": true,
			"message": "prediction model processor disabled, not executed",
			"enabled": false,
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Validate required input
	if len(features) == 0 {
		return map[string]interface{}{
			"success": false,
			"error":   "features are required for prediction",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	if apiKey == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "api_key is required for prediction",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// In a real implementation, this would call the actual prediction AI service
	// For now, we'll simulate the response
	var prediction interface{}
	var predictionConfidence float64
	var predictionDetails interface{}

	switch predictionType {
	case "classification":
		prediction = "category_A"
		predictionConfidence = 0.85
		predictionDetails = map[string]interface{}{
			"class_probabilities": map[string]interface{}{
				"category_A": 0.85,
				"category_B": 0.12,
				"category_C": 0.03,
			},
		}
	case "regression":
		prediction = 24.5
		predictionConfidence = 0.92
		predictionDetails = map[string]interface{}{
			"confidence_interval": map[string]interface{}{
				"lower_bound": 22.1,
				"upper_bound": 26.9,
			},
		}
	case "forecasting":
		prediction = []interface{}{25.1, 26.3, 27.8}
		predictionConfidence = 0.78
		predictionDetails = map[string]interface{}{
			"forecast_horizon": 3,
			"trend": "increasing",
		}
	default:
		prediction = "unknown"
		predictionConfidence = 0.5
		predictionDetails = map[string]interface{}{
			"error": "unsupported prediction type",
		}
	}

	result := map[string]interface{}{
		"prediction": prediction,
		"confidence": predictionConfidence,
		"prediction_type": predictionType,
		"prediction_details": predictionDetails,
		"model": model,
		"features_count": len(features),
		"prediction_time": time.Since(time.Now().Add(-time.Duration(timeout) * time.Second)).Seconds(),
	}

	// Add features to result if requested
	if returnFeatures {
		result["input_features"] = features
		result["feature_names"] = featureNames
	}

	// Check if confidence meets threshold
	if predictionConfidence < confidence {
		result["warning"] = fmt.Sprintf("Prediction confidence (%.2f) is below threshold (%.2f)",
			predictionConfidence, confidence)
	}

	return map[string]interface{}{
		"success": true,
		"message":        "prediction completed",
		"result":         result,
		"provider":       provider,
		"model":          model,
		"prediction_type": predictionType,
		"return_features": returnFeatures,
		"timestamp":      time.Now().Unix(),
	}, nil
}

