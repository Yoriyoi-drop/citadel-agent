package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"citadel-agent/backend/internal/nodes/base"
)

// HTTPRequestNodeV2 implements a node that makes HTTP requests (New System)
type HTTPRequestNodeV2 struct {
	*base.BaseNode
}

// HTTPRequestConfig holds HTTP request configuration
type HTTPRequestConfig struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    interface{}       `json:"body"`
	Timeout int               `json:"timeout"`
}

// NewHTTPRequestNodeWrapper creates a new HTTP request node for the registry
func NewHTTPRequestNodeWrapper() base.Node {
	metadata := base.NodeMetadata{
		ID:          "http_request",
		Name:        "HTTP Request",
		Category:    "http",
		Description: "Make HTTP requests (GET, POST, PUT, DELETE, PATCH)",
		Version:     "1.0.0",
		Author:      "Citadel Agent",
		Icon:        "globe",
		Color:       "#3b82f6",
		Inputs: []base.NodeInput{
			{
				ID:          "url",
				Name:        "URL",
				Type:        "string",
				Required:    false,
				Description: "Override URL",
			},
			{
				ID:          "body",
				Name:        "Body",
				Type:        "any",
				Required:    false,
				Description: "Request body",
			},
		},
		Outputs: []base.NodeOutput{
			{
				ID:          "response",
				Name:        "Response",
				Type:        "object",
				Description: "Response body",
			},
			{
				ID:          "status",
				Name:        "Status",
				Type:        "number",
				Description: "HTTP status code",
			},
			{
				ID:          "headers",
				Name:        "Headers",
				Type:        "object",
				Description: "Response headers",
			},
		},
		Config: []base.NodeConfig{
			{
				Name:        "url",
				Label:       "URL",
				Description: "Request URL",
				Type:        "string",
				Required:    true,
			},
			{
				Name:        "method",
				Label:       "Method",
				Description: "HTTP Method",
				Type:        "select",
				Required:    true,
				Default:     "GET",
				Options: []base.ConfigOption{
					{Label: "GET", Value: "GET"},
					{Label: "POST", Value: "POST"},
					{Label: "PUT", Value: "PUT"},
					{Label: "DELETE", Value: "DELETE"},
					{Label: "PATCH", Value: "PATCH"},
				},
			},
			{
				Name:        "headers",
				Label:       "Headers",
				Description: "Request headers",
				Type:        "json",
				Required:    false,
			},
			{
				Name:        "body",
				Label:       "Body",
				Description: "Request body",
				Type:        "json",
				Required:    false,
			},
			{
				Name:        "timeout",
				Label:       "Timeout (seconds)",
				Description: "Request timeout",
				Type:        "number",
				Required:    false,
				Default:     30,
			},
		},
		Tags: []string{"http", "api", "request"},
	}

	return &HTTPRequestNodeV2{
		BaseNode: base.NewBaseNode(metadata),
	}
}

// Execute makes HTTP request
func (n *HTTPRequestNodeV2) Execute(ctx *base.ExecutionContext, inputs map[string]interface{}) (*base.ExecutionResult, error) {
	startTime := time.Now()

	// Parse configuration
	var config HTTPRequestConfig
	if err := base.UnmarshalConfig(ctx.Variables, &config); err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Override from inputs
	if url, ok := inputs["url"].(string); ok && url != "" {
		config.URL = url
	}
	if body, ok := inputs["body"]; ok {
		config.Body = body
	}

	// Prepare request body
	var bodyReader io.Reader
	if config.Body != nil {
		if strBody, ok := config.Body.(string); ok {
			bodyReader = bytes.NewBufferString(strBody)
		} else {
			jsonBody, err := json.Marshal(config.Body)
			if err != nil {
				return base.CreateErrorResult(err, time.Since(startTime)), err
			}
			bodyReader = bytes.NewBuffer(jsonBody)
		}
	}

	// Create request
	req, err := http.NewRequest(config.Method, config.URL, bodyReader)
	if err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Set headers
	for k, v := range config.Headers {
		req.Header.Set(k, v)
	}

	// Set Content-Type if body is present and not set
	if config.Body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Client with timeout
	timeout := 30
	if config.Timeout > 0 {
		timeout = config.Timeout
	}
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Parse response body if JSON
	var parsedBody interface{}
	if err := json.Unmarshal(respBody, &parsedBody); err != nil {
		parsedBody = string(respBody)
	}

	// Prepare output
	respHeaders := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			respHeaders[k] = v[0]
		}
	}

	result := map[string]interface{}{
		"response": parsedBody,
		"status":   resp.StatusCode,
		"headers":  respHeaders,
	}

	return base.CreateSuccessResult(result, time.Since(startTime)), nil
}
