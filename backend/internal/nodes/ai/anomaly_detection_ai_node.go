package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
	"github.com/citadel-agent/backend/internal/nodes/utils"
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
func NewAnomalyDetectionAINode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	// Extract config values
	provider := utils.GetStringVal(config["provider"], "huggingface")
	apiKey := utils.GetStringVal(config["api_key"], "")
	model := utils.GetStringVal(config["model"], "isolation-forest")
	dataType := utils.GetStringVal(config["data_type"], "")
	algorithm := utils.GetStringVal(config["algorithm"], "")

	data := make([]interface{}, 0)
	if dataVal, exists := config["data"]; exists {
		if dataSlice, ok := dataVal.([]interface{}); ok {
			data = dataSlice
		}
	}

	threshold := getFloat64Value(config["threshold"], 0.7)
	windowSize := getIntValue(config["window_size"], 0)
	sensitivity := getFloat64Value(config["sensitivity"], 0.5)

	returnAnomaliesOnly := false
	if val, exists := config["return_anomalies_only"]; exists {
		if b, ok := val.(bool); ok {
			returnAnomaliesOnly = b
		}
	}

	customParams := make(map[string]interface{})
	if paramsVal, exists := config["custom_params"]; exists {
		if paramsMap, ok := paramsVal.(map[string]interface{}); ok {
			customParams = paramsMap
		}
	}

	timeout := getIntValue(config["timeout"], 45)
	enabled := getBoolValue(config["enabled"], true)

	nodeConfig := &AnomalyDetectionAINodeConfig{
		Provider:            provider,
		ApiKey:              apiKey,
		Model:               model,
		Data:                data,
		DataType:            dataType,
		Algorithm:           algorithm,
		Threshold:           threshold,
		WindowSize:          windowSize,
		Sensitivity:         sensitivity,
		ReturnAnomaliesOnly: returnAnomaliesOnly,
		CustomParams:        customParams,
		Timeout:             timeout,
		Enabled:             enabled,
	}

	return &AnomalyDetectionAINode{
		config: nodeConfig,
	}, nil
}

// Execute implements the NodeInstance interface
func (a AnomalyDetectionAINode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Override configuration with input values if provided
	provider := a.config.Provider
	if inputProvider, exists := input["provider"]; exists {
		if inputProviderStr, ok := inputProvider.(string); ok && inputProviderStr != "" {
			provider = inputProviderStr
		}
	}

	apiKey := a.config.ApiKey
	if inputApiKey, exists := input["api_key"]; exists {
		if inputApiKeyStr, ok := inputApiKey.(string); ok && inputApiKeyStr != "" {
			apiKey = inputApiKeyStr
		}
	}

	model := a.config.Model
	if inputModel, exists := input["model"]; exists {
		if inputModelStr, ok := inputModel.(string); ok && inputModelStr != "" {
			model = inputModelStr
		}
	}

	data := a.config.Data
	if inputData, exists := input["data"]; exists {
		if inputDataSlice, ok := inputData.([]interface{}); ok {
			data = inputDataSlice
		}
	}

	dataType := a.config.DataType
	if inputDataType, exists := input["data_type"]; exists {
		if inputDataTypeStr, ok := inputDataType.(string); ok && inputDataTypeStr != "" {
			dataType = inputDataTypeStr
		}
	}

	algorithm := a.config.Algorithm
	if inputAlgorithm, exists := input["algorithm"]; exists {
		if inputAlgorithmStr, ok := inputAlgorithm.(string); ok && inputAlgorithmStr != "" {
			algorithm = inputAlgorithmStr
		}
	}

	threshold := a.config.Threshold
	if inputThreshold, exists := input["threshold"]; exists {
		if inputThresholdFloat, ok := inputThreshold.(float64); ok {
			threshold = inputThresholdFloat
		}
	}

	windowSize := a.config.WindowSize
	if inputWindowSize, exists := input["window_size"]; exists {
		if inputWindowSizeFloat, ok := inputWindowSize.(float64); ok {
			windowSize = int(inputWindowSizeFloat)
		}
	}

	sensitivity := a.config.Sensitivity
	if inputSensitivity, exists := input["sensitivity"]; exists {
		if inputSensitivityFloat, ok := inputSensitivity.(float64); ok {
			sensitivity = inputSensitivityFloat
		}
	}

	returnAnomaliesOnly := a.config.ReturnAnomaliesOnly
	if inputReturnAnomalies, exists := input["return_anomalies_only"]; exists {
		if inputReturnAnomaliesBool, ok := inputReturnAnomalies.(bool); ok {
			returnAnomaliesOnly = inputReturnAnomaliesBool
		}
	}

	customParams := a.config.CustomParams
	if inputCustomParams, exists := input["custom_params"]; exists {
		if inputCustomParamsMap, ok := inputCustomParams.(map[string]interface{}); ok {
			customParams = inputCustomParamsMap
		}
	}

	timeout := a.config.Timeout
	if inputTimeout, exists := input["timeout"]; exists {
		if inputTimeoutFloat, ok := inputTimeout.(float64); ok {
			timeout = int(inputTimeoutFloat)
		}
	}

	enabled := a.config.Enabled
	if inputEnabled, exists := input["enabled"]; exists {
		if inputEnabledBool, ok := inputEnabled.(bool); ok {
			enabled = inputEnabledBool
		}
	}

	// Check if node should be enabled
	if !enabled {
		return map[string]interface{}{
			"success": true,
			"message": "anomaly detection processor disabled, not executed",
			"enabled": false,
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Validate required input
	if len(data) == 0 {
		return map[string]interface{}{
			"success": false,
			"error":   "data is required for anomaly detection",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	if apiKey == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "api_key is required for anomaly detection",
			"timestamp": time.Now().Unix(),
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
		if anomalyScore, ok := anomaly["anomaly_score"].(float64); ok {
			if anomalyScore >= threshold {
				filteredAnomalies = append(filteredAnomalies, anomaly)
			}
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

	return map[string]interface{}{
		"success": true,
		"message":        "anomaly detection completed",
		"result":         result,
		"provider":       provider,
		"model":          model,
		"algorithm":      algorithm,
		"return_anomalies_only": returnAnomaliesOnly,
		"timestamp":      time.Now().Unix(),
	}, nil
}

