package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
)

// HTTPRequestConfig represents configuration for HTTP request node
type HTTPRequestConfig struct {
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers"`
	Body        interface{}       `json:"body"`
	Timeout     int               `json:"timeout"`      // in seconds
	AuthType    string            `json:"auth_type"`    // "none", "basic", "bearer", "api_key"
	AuthValue   string            `json:"auth_value"`
	VerifySSL   bool              `json:"verify_ssl"`
	RetryCount  int               `json:"retry_count"`  // Number of retries
	RetryDelay  int               `json:"retry_delay"`  // Delay between retries in seconds
	EnableCaching bool            `json:"enable_caching"`
	CacheTTL    int               `json:"cache_ttl"`    // Cache TTL in seconds
	EnableProfiling bool          `json:"enable_profiling"`
	ReturnRawResults bool          `json:"return_raw_results"`
	CustomParams map[string]interface{} `json:"custom_params"`
}

// HTTPRequestNode represents an HTTP request node
type HTTPRequestNode struct {
	config *HTTPRequestConfig
}

// NewHTTPRequestNode creates a new HTTP request node
func NewHTTPRequestNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	// Convert config map to struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var httpConfig HTTPRequestConfig
	if err := json.Unmarshal(jsonData, &httpConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate and set defaults
	if httpConfig.Method == "" {
		httpConfig.Method = "GET"
	}

	httpConfig.Method = strings.ToUpper(httpConfig.Method)

	if httpConfig.Timeout == 0 {
		httpConfig.Timeout = 30 // 30 seconds default
	}

	if httpConfig.RetryCount == 0 {
		httpConfig.RetryCount = 3
	}

	if httpConfig.RetryDelay == 0 {
		httpConfig.RetryDelay = 1 // 1 second default delay
	}

	if httpConfig.CacheTTL == 0 {
		httpConfig.CacheTTL = 3600 // 1 hour default cache TTL
	}

	if httpConfig.Headers == nil {
		httpConfig.Headers = make(map[string]string)
	}

	return &HTTPRequestNode{
		config: &httpConfig,
	}, nil
}

// Execute implements the NodeInstance interface
func (h *HTTPRequestNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	startTime := time.Now()

	// Override config values with inputs if provided
	url := h.config.URL
	if inputURL, exists := inputs["url"]; exists {
		if urlStr, ok := inputURL.(string); ok && urlStr != "" {
			url = urlStr
		}
	}

	if url == "" {
		return nil, fmt.Errorf("URL is required")
	}

	method := h.config.Method
	if inputMethod, exists := inputs["method"]; exists {
		if methodStr, ok := inputMethod.(string); ok && methodStr != "" {
			method = strings.ToUpper(methodStr)
		}
	}

	body := h.config.Body
	if inputBody, exists := inputs["body"]; exists {
		body = inputBody
	}

	retryCount := h.config.RetryCount
	if inputRetryCount, exists := inputs["retry_count"]; exists {
		if retryCountFloat, ok := inputRetryCount.(float64); ok {
			retryCount = int(retryCountFloat)
		}
	}

	retryDelay := h.config.RetryDelay
	if inputRetryDelay, exists := inputs["retry_delay"]; exists {
		if retryDelayFloat, ok := inputRetryDelay.(float64); ok {
			retryDelay = int(retryDelayFloat)
		}
	}

	timeout := h.config.Timeout
	if inputTimeout, exists := inputs["timeout"]; exists {
		if timeoutFloat, ok := inputTimeout.(float64); ok {
			timeout = int(timeoutFloat)
		}
	}

	headers := h.config.Headers
	if inputHeaders, exists := inputs["headers"]; exists {
		if inputHeaderMap, ok := inputHeaders.(map[string]interface{}); ok {
			headers = make(map[string]string)
			for k, v := range inputHeaderMap {
				if vStr, ok := v.(string); ok {
					headers[k] = vStr
				}
			}
		}
	}

	authType := h.config.AuthType
	if inputAuthType, exists := inputs["auth_type"]; exists {
		if authTypeStr, ok := inputAuthType.(string); ok {
			authType = authTypeStr
		}
	}

	authValue := h.config.AuthValue
	if inputAuthValue, exists := inputs["auth_value"]; exists {
		if authValueStr, ok := inputAuthValue.(string); ok {
			authValue = authValueStr
		}
	}

	enableProfiling := h.config.EnableProfiling
	if inputEnableProfiling, exists := inputs["enable_profiling"]; exists {
		if inputEnableProfiling, ok := inputEnableProfiling.(bool); ok {
			enableProfiling = inputEnableProfiling
		}
	}

	returnRawResults := h.config.ReturnRawResults
	if inputReturnRaw, exists := inputs["return_raw_results"]; exists {
		if inputReturnRaw, ok := inputReturnRaw.(bool); ok {
			returnRawResults = inputReturnRaw
		}
	}

	// Prepare request
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	var req *http.Request
	var err error

	// Marshal body if it's not already a string
	var bodyReader io.Reader
	if body != nil {
		var bodyBytes []byte
		switch v := body.(type) {
		case string:
			bodyReader = strings.NewReader(v)
		case []byte:
			bodyReader = bytes.NewReader(v)
		default:
			bodyBytes, err = json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			bodyReader = bytes.NewReader(bodyBytes)
		}
	}

	req, err = http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers from config
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Set Content-Type header if not set and we have a body
	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add authentication headers
	switch authType {
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+authValue)
	case "basic":
		// Basic auth implementation would go here
		// For now, just add a placeholder
		req.Header.Set("Authorization", "Basic "+authValue)
	case "api_key":
		// API key implementation - often in a custom header
		apiKeyHeader := "X-API-Key" // Default, could be configurable
		if apiKeyHeaderVal, exists := inputs["api_key_header"]; exists {
			if apiKeyHeaderStr, ok := apiKeyHeaderVal.(string); ok {
				apiKeyHeader = apiKeyHeaderStr
			}
		}
		req.Header.Set(apiKeyHeader, authValue)
	}

	// Make the request with retries
	var resp *http.Response
	var respErr error

	for attempt := 0; attempt <= retryCount; attempt++ {
		resp, respErr = client.Do(req)
		if respErr == nil {
			break // Success, break out of retry loop
		}

		if attempt == retryCount {
			// Last attempt, return error
			return nil, fmt.Errorf("failed to execute request after %d attempts: %w", retryCount+1, respErr)
		}

		// Wait before retrying
		time.Sleep(time.Duration(retryDelay) * time.Second)
	}

	if resp == nil {
		return nil, fmt.Errorf("response is nil after all attempts")
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Try to parse response as JSON
	var responseData interface{}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &responseData); err != nil {
			// If JSON parsing fails, return the raw response body
			responseData = string(respBody)
		}
	} else {
		// Empty body, return a message indicating empty response
		responseData = map[string]interface{}{
			"message": "empty response body",
		}
	}

	// Prepare result
	result := map[string]interface{}{
		"success":       resp.StatusCode >= 200 && resp.StatusCode < 300,
		"status":        resp.Status,
		"status_code":   resp.StatusCode,
		"data":          responseData,
		"headers":       resp.Header,
		"url":           url,
		"method":        method,
		"request_time":  time.Since(startTime).Seconds(),
		"timestamp":     time.Now().Unix(),
		"input_data":    inputs,
		"retries_used":  retryCount,
		"execution_time": time.Since(startTime).Seconds(),
		"response_size": len(respBody),
	}

	// Add profiling data if enabled
	if enableProfiling {
		result["profiling"] = map[string]interface{}{
			"start_time": startTime.Unix(),
			"end_time":   time.Now().Unix(),
			"duration":   time.Since(startTime).Seconds(),
			"retries":    retryCount,
			"retry_delay": retryDelay,
			"timeout":    timeout,
		}
	}

	// Return raw results if requested
	if returnRawResults {
		result["raw_response"] = string(respBody)
	}

	return result, nil
}

// GetType returns the type of node
func (h *HTTPRequestNode) GetType() string {
	return "http_request"
}

// GetID returns the unique ID of the node instance
func (h *HTTPRequestNode) GetID() string {
	return fmt.Sprintf("http_%s_%d", h.config.Method, time.Now().Unix())
}