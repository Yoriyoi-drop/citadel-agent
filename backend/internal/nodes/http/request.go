package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"citadel-agent/backend/internal/interfaces"
)

// HTTPRequestNode implements a node that makes HTTP requests
type HTTPRequestNode struct {
	id          string
	nodeType    string
	method      string
	url         string
	headers     map[string]string
	body        string
	timeout     time.Duration
	authType    string
	authValue   string
	config      map[string]interface{}
}

// Initialize sets up the HTTP request node with configuration
func (h *HTTPRequestNode) Initialize(config map[string]interface{}) error {
	h.config = config

	if method, ok := config["method"]; ok {
		if m, ok := method.(string); ok {
			h.method = m
		} else {
			return fmt.Errorf("method must be a string")
		}
	} else {
		h.method = "GET" // default method
	}

	if url, ok := config["url"]; ok {
		if u, ok := url.(string); ok {
			h.url = u
		} else {
			return fmt.Errorf("url must be a string")
		}
	} else {
		return fmt.Errorf("url is required")
	}

	if headers, ok := config["headers"]; ok {
		if hMap, ok := headers.(map[string]interface{}); ok {
			h.headers = make(map[string]string)
			for k, v := range hMap {
				if vStr, ok := v.(string); ok {
					h.headers[k] = vStr
				} else {
					h.headers[k] = fmt.Sprintf("%v", v)
				}
			}
		} else {
			return fmt.Errorf("headers must be an object")
		}
	} else {
		h.headers = make(map[string]string)
	}

	if body, ok := config["body"]; ok {
		if b, ok := body.(string); ok {
			h.body = b
		} else {
			// Try to serialize the body if it's not a string
			bodyBytes, err := json.Marshal(body)
			if err != nil {
				return fmt.Errorf("failed to serialize body: %v", err)
			}
			h.body = string(bodyBytes)
		}
	}

	if timeout, ok := config["timeout"]; ok {
		if t, ok := timeout.(float64); ok {
			h.timeout = time.Duration(t) * time.Second
		} else if t, ok := timeout.(int); ok {
			h.timeout = time.Duration(t) * time.Second
		} else {
			return fmt.Errorf("timeout must be a number")
		}
	} else {
		h.timeout = 30 * time.Second // default timeout
	}

	if authType, ok := config["auth_type"]; ok {
		if at, ok := authType.(string); ok {
			h.authType = at
		}
	}

	if authValue, ok := config["auth_value"]; ok {
		if av, ok := authValue.(string); ok {
			h.authValue = av
		}
	}

	return nil
}

// Execute runs the HTTP request
func (h *HTTPRequestNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	client := &http.Client{
		Timeout: h.timeout,
	}

	// Prepare request body
	var bodyReader io.Reader
	if h.body != "" {
		bodyReader = bytes.NewBufferString(h.body)
	} else if len(inputs) > 0 {
		// If no explicit body, try to use inputs
		inputBytes, err := json.Marshal(inputs)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal input data: %v", err)
		}
		bodyReader = bytes.NewBuffer(inputBytes)
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, h.method, h.url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	for key, value := range h.headers {
		req.Header.Set(key, value)
	}

	// Set content type if not already set and we have a body
	if h.body != "" || len(inputs) > 0 {
		if req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}
	}

	// Set up authentication if configured
	if h.authType != "" && h.authValue != "" {
		switch h.authType {
		case "bearer":
			req.Header.Set("Authorization", "Bearer "+h.authValue)
		case "api_key":
			req.Header.Set("Authorization", h.authValue)
		case "basic":
			// For basic auth, the auth_value should be in format "username:password"
			req.Header.Set("Authorization", "Basic "+h.authValue)
		}
	}

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Prepare response data
	result := map[string]interface{}{
		"status_code": resp.StatusCode,
		"status":      resp.Status,
		"headers":     resp.Header,
		"body":        string(respBody),
		"method":      h.method,
		"url":         h.url,
	}

	return result, nil
}

// GetType returns the type of the node
func (h *HTTPRequestNode) GetType() string {
	return h.nodeType
}

// GetID returns the unique identifier for this node instance
func (h *HTTPRequestNode) GetID() string {
	return h.id
}

// NewHTTPRequestNode creates a new HTTP request node constructor for the registry
func NewHTTPRequestNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	node := &HTTPRequestNode{
		id:       fmt.Sprintf("http_%d", time.Now().UnixNano()),
		nodeType: "http_request",
	}

	if err := node.Initialize(config); err != nil {
		return nil, err
	}

	return node, nil
}