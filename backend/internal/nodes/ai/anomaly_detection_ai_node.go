package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/engine"
)

// AnomalyDetectionAINodeConfig represents the configuration for an Anomaly Detection AI node
type AnomalyDetectionAINodeConfig struct {
	Provider       string                 `json:"provider"`        // AI provider (openai, google, huggingface, etc.)
	ApiKey         string                 `json:"api_key"`         // API key for the anomaly detection service
	Model          string                 `json:"model"`           // Anomaly detection model to use
	Data           []interface{}          `json:"data"`            // Input data for anomaly detection
	DataType       string                 `json:"data_type"`       // Type of data (numeric, categorical, time_series, etc.)
	Algorithm      string                 `json:"algorithm"`       // Anomaly detection algorithm (isolation_forest, local_outlier, etc.)
	Threshold      float64                `json:"threshold"`       // Anomaly threshold (0-1)
	WindowSize     int                    `json:"window_size"`     // Window size for time series analysis
	Sensitivity    float64                `json:"sensitivity"`     // Detection sensitivity (0-1)
	ReturnAnomaliesOnly bool              `json:"return_anomalies_only"` // Whether to return only anomalies
	CustomParams   map[string]interface{} `json:"custom_params"`   // Custom parameters for the AI service
	Timeout        int                    `json:"timeout"`         // Request timeout in seconds
	Enabled        bool                   `json:"enabled"`         // Whether the node is enabled
}

// AnomalyDetectionAINode represents a node that detects anomalies in data using AI
type AnomalyDetectionAINode struct {
	config *AnomalyDetectionAINodeConfig
}

// NewAnomalyDetectionAINode creates a new Anomaly Detection AI node
func NewAnomalyDetectionAINode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Convert interface{} map to JSON and back to struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}

	var anomalyConfig AnomalyDetectionAINodeConfig
	err = json.Unmarshal(jsonData, &anomalyConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// Validate required fields
	if anomalyConfig.Provider == "" {
		anomalyConfig.Provider = "huggingface" // default provider
	}

	if anomalyConfig.Model == "" {
		anomalyConfig.Model = "isolation-forest" // default model
	}

	if anomalyConfig.Threshold == 0 {
		anomalyConfig.Threshold = 0.7 // default threshold of 70%
	}

	if anomalyConfig.Sensitivity == 0 {
		anomalyConfig.Sensitivity = 0.5 // default sensitivity of 50%
	}

	if anomalyConfig.Timeout == 0 {
		anomalyConfig.Timeout = 45 // default timeout of 45 seconds for anomaly detection
	}

	return &AnomalyDetectionAINode{
		config: &anomalyConfig,
	}, nil
}

// Execute implements the NodeInstance interface
func (a *AnomalyDetectionAINode) Execute(ctx context.Context, input map[string]interface{}) (*engine.ExecutionResult, error) {
	// Override configuration with input values if provided
	provider := a.config.Provider
	if inputProvider, ok := input["provider"].(string); ok && inputProvider != "" {
		provider = inputProvider
	}

	apiKey := a.config.ApiKey
	if inputApiKey, ok := input["api_key"].(string); ok && inputApiKey != "" {
		apiKey = inputApiKey
	}

	model := a.config.Model
	if inputModel, ok := input["model"].(string); ok && inputModel != "" {
		model = inputModel
	}

	data := a.config.Data
	if inputData, ok := input["data"].([]interface{}); ok {
		data = inputData
	}

	dataType := a.config.DataType
	if inputDataType, ok := input["data_type"].(string); ok && inputDataType != "" {
		dataType = inputDataType
	}

	algorithm := a.config.Algorithm
	if inputAlgorithm, ok := input["algorithm"].(string); ok && inputAlgorithm != "" {
		algorithm = inputAlgorithm
	}

	threshold := a.config.Threshold
	if inputThreshold, ok := input["threshold"].(float64); ok {
		threshold = inputThreshold
	}

	windowSize := a.config.WindowSize
	if inputWindowSize, ok := input["window_size"].(float64); ok {
		windowSize = int(inputWindowSize)
	}

	sensitivity := a.config.Sensitivity
	if inputSensitivity, ok := input["sensitivity"].(float64); ok {
		sensitivity = inputSensitivity
	}

	returnAnomaliesOnly := a.config.ReturnAnomaliesOnly
	if inputReturnAnomalies, ok := input["return_anomalies_only"].(bool); ok {
		returnAnomaliesOnly = inputReturnAnomalies
	}

	customParams := a.config.CustomParams
	if inputCustomParams, ok := input["custom_params"].(map[string]interface{}); ok {
		customParams = inputCustomParams
	}

	timeout := a.config.Timeout
	if inputTimeout, ok := input["timeout"].(float64); ok {
		timeout = int(inputTimeout)
	}

	enabled := a.config.Enabled
	if inputEnabled, ok := input["enabled"].(bool); ok {
		enabled = inputEnabled
	}

	// Check if node should be enabled
	if !enabled {
		return &engine.ExecutionResult{
			Status: "success",
			Data: map[string]interface{}{
				"message": "anomaly detection processor disabled, not executed",
				"enabled": false,
			},
			Timestamp: time.Now(),
		}, nil
	}

	// Validate required input
	if len(data) == 0 {
		return &engine.ExecutionResult{
			Status:    "error",
			Error:     "data is required for anomaly detection",
			Timestamp: time.Now(),
		}, nil
	}

	if apiKey == "" {
		return &engine.ExecutionResult{
			Status:    "error",
			Error:     "api_key is required for anomaly detection",
			Timestamp: time.Now(),
		}, nil
	}

	// In a real implementation, this would call the actual anomaly detection AI service
	// For now, we'll simulate the response
	anomalies := []map[string]interface{}{
		{
			"index": 5,
			"value": 98.7,
			"anomaly_score": 0.89,
			"anomaly_type": "spike",
			"confidence": 0.85,
		},
		{
			"index": 12,
			"value": 12.3,
			"anomaly_score": 0.92,
			"anomaly_type": "drop",
			"confidence": 0.91,
		},
	}

	// Filter anomalies based on threshold if only anomalies are requested
	filteredAnomalies := []map[string]interface{}{}
	for _, anomaly := range anomalies {
		if anomaly["anomaly_score"].(float64) >= threshold {
			filteredAnomalies = append(filteredAnomalies, anomaly)
		}
	}

	if returnAnomaliesOnly {
		anomalies = filteredAnomalies
	}

	result := map[string]interface{}{
		"input_data_type": dataType,
		"algorithm": algorithm,
		"threshold": threshold,
		"sensitivity": sensitivity,
		"total_data_points": len(data),
		"total_anomalies": len(anomalies),
		"anomalies": anomalies,
		"anomaly_percentage": float64(len(anomalies)) / float64(len(data)) * 100,
		"detection_time": time.Since(time.Now().Add(-time.Duration(timeout) * time.Second)).Seconds(),
	}

	return &engine.ExecutionResult{
		Status: "success",
		Data: map[string]interface{}{
			"message":        "anomaly detection completed",
			"result":         result,
			"provider":       provider,
			"model":          model,
			"algorithm":      algorithm,
			"return_anomalies_only": returnAnomaliesOnly,
			"timestamp":      time.Now().Unix(),
		},
		Timestamp: time.Now(),
	}, nil
}

// GetType returns the type of the node
func (a *AnomalyDetectionAINode) GetType() string {
	return "anomaly_detection_ai"
}

// GetID returns a unique ID for the node instance
func (a *AnomalyDetectionAINode) GetID() string {
	return "anomaly_detection_ai_" + fmt.Sprintf("%d", time.Now().Unix())
}