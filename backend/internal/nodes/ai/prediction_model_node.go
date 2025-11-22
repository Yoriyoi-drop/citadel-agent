package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/engine"
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
func NewPredictionModelNode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Convert interface{} map to JSON and back to struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}

	var predictionConfig PredictionModelNodeConfig
	err = json.Unmarshal(jsonData, &predictionConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// Validate required fields
	if predictionConfig.Provider == "" {
		predictionConfig.Provider = "huggingface" // default provider
	}

	if predictionConfig.Model == "" {
		predictionConfig.Model = "microsoft/regnet-600mf" // default model
	}

	if predictionConfig.PredictionType == "" {
		predictionConfig.PredictionType = "classification" // default type
	}

	if predictionConfig.Confidence == 0 {
		predictionConfig.Confidence = 0.7 // default confidence of 70%
	}

	if predictionConfig.Timeout == 0 {
		predictionConfig.Timeout = 30 // default timeout of 30 seconds
	}

	return &PredictionModelNode{
		config: &predictionConfig,
	}, nil
}

// Execute implements the NodeInstance interface
func (p *PredictionModelNode) Execute(ctx context.Context, input map[string]interface{}) (*engine.ExecutionResult, error) {
	// Override configuration with input values if provided
	provider := p.config.Provider
	if inputProvider, ok := input["provider"].(string); ok && inputProvider != "" {
		provider = inputProvider
	}

	apiKey := p.config.ApiKey
	if inputApiKey, ok := input["api_key"].(string); ok && inputApiKey != "" {
		apiKey = inputApiKey
	}

	model := p.config.Model
	if inputModel, ok := input["model"].(string); ok && inputModel != "" {
		model = inputModel
	}

	features := p.config.Features
	if inputFeatures, ok := input["features"].([]interface{}); ok {
		features = inputFeatures
	}

	featureNames := p.config.FeatureNames
	if inputFeatureNames, ok := input["feature_names"].([]interface{}); ok {
		featureNames = make([]string, len(inputFeatureNames))
		for i, name := range inputFeatureNames {
			featureNames[i] = fmt.Sprintf("%v", name)
		}
	}

	predictionType := p.config.PredictionType
	if inputPredictionType, ok := input["prediction_type"].(string); ok && inputPredictionType != "" {
		predictionType = inputPredictionType
	}

	modelEndpoint := p.config.ModelEndpoint
	if inputEndpoint, ok := input["model_endpoint"].(string); ok && inputEndpoint != "" {
		modelEndpoint = inputEndpoint
	}

	confidence := p.config.Confidence
	if inputConfidence, ok := input["confidence"].(float64); ok {
		confidence = inputConfidence
	}

	returnFeatures := p.config.ReturnFeatures
	if inputReturnFeatures, ok := input["return_features"].(bool); ok {
		returnFeatures = inputReturnFeatures
	}

	customParams := p.config.CustomParams
	if inputCustomParams, ok := input["custom_params"].(map[string]interface{}); ok {
		customParams = inputCustomParams
	}

	timeout := p.config.Timeout
	if inputTimeout, ok := input["timeout"].(float64); ok {
		timeout = int(inputTimeout)
	}

	enabled := p.config.Enabled
	if inputEnabled, ok := input["enabled"].(bool); ok {
		enabled = inputEnabled
	}

	// Check if node should be enabled
	if !enabled {
		return &engine.ExecutionResult{
			Status: "success",
			Data: map[string]interface{}{
				"message": "prediction model processor disabled, not executed",
				"enabled": false,
			},
			Timestamp: time.Now(),
		}, nil
	}

	// Validate required input
	if len(features) == 0 {
		return &engine.ExecutionResult{
			Status:    "error",
			Error:     "features are required for prediction",
			Timestamp: time.Now(),
		}, nil
	}

	if apiKey == "" {
		return &engine.ExecutionResult{
			Status:    "error",
			Error:     "api_key is required for prediction",
			Timestamp: time.Now(),
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

	return &engine.ExecutionResult{
		Status: "success",
		Data: map[string]interface{}{
			"message":        "prediction completed",
			"result":         result,
			"provider":       provider,
			"model":          model,
			"prediction_type": predictionType,
			"return_features": returnFeatures,
			"timestamp":      time.Now().Unix(),
		},
		Timestamp: time.Now(),
	}, nil
}

// GetType returns the type of the node
func (p *PredictionModelNode) GetType() string {
	return "prediction_model"
}

// GetID returns a unique ID for the node instance
func (p *PredictionModelNode) GetID() string {
	return "prediction_model_" + fmt.Sprintf("%d", time.Now().Unix())
}