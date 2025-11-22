package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
)

// HTTPNodeConfig represents the configuration for an HTTP node
type HTTPNodeConfig struct {
	Method      string                 `json:"method"`
	URL         string                 `json:"url"`
	Headers     map[string]string      `json:"headers,omitempty"`
	Body        interface{}            `json:"body,omitempty"`
	QueryParams map[string]string      `json:"query_params,omitempty"`
	Timeout     int                    `json:"timeout,omitempty"` // in seconds
	Auth        *HTTPAuthConfig        `json:"auth,omitempty"`
}

// HTTPAuthConfig represents authentication configuration for HTTP requests
type HTTPAuthConfig struct {
	Type     string `json:"type"` // "basic", "bearer", "api_key"
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token,omitempty"`
	ApiKey   string `json:"api_key,omitempty"`
	KeyName  string `json:"key_name,omitempty"` // for API key header name
}

// HTTPNode executes HTTP requests with full configuration support
type HTTPNode struct {
	config HTTPNodeConfig
}

// NewHTTPNode creates a new HTTP node with the given configuration
func NewHTTPNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	// Convert interface{} map to JSON and back to struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}

	var httpConfig HTTPNodeConfig
	err = json.Unmarshal(jsonData, &httpConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// Validate required fields
	if httpConfig.URL == "" {
		return nil, fmt.Errorf("URL is required for HTTP node")
	}

	return &HTTPNode{
		config: httpConfig,
	}, nil
}

// Execute implements the NodeInstance interface
func (h *HTTPNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Override configuration with input values if provided
	method := h.config.Method
	if inputMethod, ok := input["method"].(string); ok && inputMethod != "" {
		method = inputMethod
	}
	if method == "" {
		method = "GET"
	}

	url := h.config.URL
	if inputURL, ok := input["url"].(string); ok && inputURL != "" {
		url = inputURL
	}

	// Create HTTP client with timeout
	timeout := time.Duration(h.config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second // default timeout
	}
	
	client := &http.Client{
		Timeout: timeout,
	}

	// Prepare request body
	var bodyReader io.Reader
	if h.config.Body != nil {
		bodyBytes, err := json.Marshal(h.config.Body)
		if err != nil {
			return map[string]interface{}{
				"status":    "error",
				"error":     fmt.Sprintf("failed to marshal request body: %v", err),
				"timestamp": time.Now().Unix(),
			}, nil
		}
		bodyReader = bytes.NewReader(bodyBytes)
	} else if inputBody, exists := input["body"]; exists {
		bodyBytes, err := json.Marshal(inputBody)
		if err != nil {
			return map[string]interface{}{
				"status":    "error",
				"error":     fmt.Sprintf("failed to marshal input body: %v", err),
				"timestamp": time.Now().Unix(),
			}, nil
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return map[string]interface{}{
			"status":    "error",
			"error":     fmt.Sprintf("failed to create HTTP request: %v", err),
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Set headers from config
	for key, value := range h.config.Headers {
		req.Header.Set(key, value)
	}

	// Add query parameters
	q := req.URL.Query()
	for key, value := range h.config.QueryParams {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	// Add authentication headers if configured
	if h.config.Auth != nil {
		switch h.config.Auth.Type {
		case "basic":
			req.SetBasicAuth(h.config.Auth.Username, h.config.Auth.Password)
		case "bearer":
			req.Header.Set("Authorization", "Bearer "+h.config.Auth.Token)
		case "api_key":
			keyName := h.config.Auth.KeyName
			if keyName == "" {
				keyName = "X-API-Key" // default API key header
			}
			req.Header.Set(keyName, h.config.Auth.ApiKey)
		}
	}

	// Set default content type if not set and we have a body
	if req.Body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{
			"status":    "error",
			"error":     fmt.Sprintf("HTTP request failed: %v", err),
			"timestamp": time.Now().Unix(),
		}, nil
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]interface{}{
			"status":    "error",
			"error":     fmt.Sprintf("failed to read response body: %v", err),
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Parse response body as JSON if possible
	var parsedRespBody interface{}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &parsedRespBody); err != nil {
			// If it's not JSON, return as string
			parsedRespBody = string(respBody)
		}
	} else {
		parsedRespBody = nil
	}

	// Prepare response headers
	respHeaders := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			respHeaders[key] = values[0] // take first value for simplicity
		}
	}

	resultData := map[string]interface{}{
		"status_code":  resp.StatusCode,
		"status":       resp.Status,
		"headers":      respHeaders,
		"body":         parsedRespBody,
		"response_time": time.Since(time.Now().Add(-timeout)).String(),
	}

	return map[string]interface{}{
		"status":    "success",
		"data":      resultData,
		"timestamp": time.Now().Unix(),
	}, nil
}

