// backend/internal/nodes/integrations/github_node.go
package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"citadel-agent/backend/internal/workflow/core/engine"
)

// GitHubNodeConfig represents the configuration for a GitHub node
type GitHubNodeConfig struct {
	GithubToken    string                 `json:"github_token"`
	Repository     string                 `json:"repository"` // owner/repo format
	Endpoint       string                 `json:"endpoint"`   // API endpoint
	Method         string                 `json:"method"`     // GET, POST, PUT, PATCH, DELETE
	RequestBody    map[string]interface{} `json:"request_body,omitempty"`
	QueryParams    map[string]string      `json:"query_params,omitempty"`
}

// GitHubNode represents a GitHub integration node
type GitHubNode struct {
	config *GitHubNodeConfig
}

// NewGitHubNode creates a new GitHub node
func NewGitHubNode(config *GitHubNodeConfig) *GitHubNode {
	if config.Method == "" {
		config.Method = "GET"
	}
	return &GitHubNode{
		config: config,
	}
}

// Execute executes the GitHub operation
func (gn *GitHubNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Override config values with inputs if provided
	token := gn.config.GithubToken
	if tok, exists := inputs["github_token"]; exists {
		if tokStr, ok := tok.(string); ok {
			token = tokStr
		}
	}

	repo := gn.config.Repository
	if r, exists := inputs["repository"]; exists {
		if rStr, ok := r.(string); ok {
			repo = rStr
		}
	}

	endpoint := gn.config.Endpoint
	if ep, exists := inputs["endpoint"]; exists {
		if epStr, ok := ep.(string); ok {
			endpoint = epStr
		}
	}

	method := gn.config.Method
	if m, exists := inputs["method"]; exists {
		if mStr, ok := m.(string); ok {
			method = mStr
		}
	}

	// Construct the API URL
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s", repo, endpoint)

	// Prepare the request
	req, err := http.NewRequestWithContext(ctx, method, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "Citadel-Agent/1.0")

	// Add query parameters
	q := req.URL.Query()
	for key, value := range gn.config.QueryParams {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	// For POST/PUT/PATCH requests, encode the body
	if method == "POST" || method == "PUT" || method == "PATCH" {
		bodyData := gn.config.RequestBody
		if bodyMap, exists := inputs["request_body"]; exists {
			if body, ok := bodyMap.(map[string]interface{}); ok {
				bodyData = body
			}
		}

		if bodyData != nil {
			jsonData, err := json.Marshal(bodyData)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}

			req.Body = &nopCloser{jsonData}
			req.Header.Set("Content-Type", "application/json")
		}
	}

	// Execute the request
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody := make([]byte, 0)
	if resp.Body != nil {
		responseBody, _ = io.ReadAll(resp.Body)
	}

	// Parse response
	var responseData interface{}
	if len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, &responseData); err != nil {
			// If JSON parsing fails, return as string
			responseData = string(responseBody)
		}
	}

	result := map[string]interface{}{
		"success":    true,
		"status_code": resp.StatusCode,
		"headers":    resp.Header,
		"response":   responseData,
		"request": map[string]interface{}{
			"method":      method,
			"url":         apiURL,
			"query_params": gn.config.QueryParams,
			"request_body": gn.config.RequestBody,
		},
		"timestamp": time.Now().Unix(),
	}

	// Check for error status codes
	if resp.StatusCode >= 400 {
		result["success"] = false
		result["error"] = fmt.Sprintf("GitHub API returned status %d", resp.StatusCode)
	}

	return result, nil
}

// nopCloser wraps a byte slice and implements io.ReadCloser
type nopCloser struct {
	data []byte
}

func (n *nopCloser) Read(p []byte) (int, error) {
	if len(n.data) == 0 {
		return 0, io.EOF
	}
	n, m := copy(p, n.data)
	n.data = n.data[n:]
	return m, nil
}

func (n *nopCloser) Close() error {
	return nil
}

// RegisterGitHubNode registers the GitHub node type with the engine
func RegisterGitHubNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("github_integration", func(config map[string]interface{}) (engine.NodeInstance, error) {
		var token string
		if tok, exists := config["github_token"]; exists {
			if tokStr, ok := tok.(string); ok {
				token = tokStr
			}
		}

		var repo string
		if rep, exists := config["repository"]; exists {
			if repStr, ok := rep.(string); ok {
				repo = repStr
			}
		}

		var endpoint string
		if ep, exists := config["endpoint"]; exists {
			if epStr, ok := ep.(string); ok {
				endpoint = epStr
			}
		}

		var method string
		if m, exists := config["method"]; exists {
			if mStr, ok := m.(string); ok {
				method = mStr
			}
		}

		var requestBody map[string]interface{}
		if body, exists := config["request_body"]; exists {
			if bodyMap, ok := body.(map[string]interface{}); ok {
				requestBody = bodyMap
			}
		}

		var queryParams map[string]string
		if params, exists := config["query_params"]; exists {
			if paramMap, ok := params.(map[string]interface{}); ok {
				queryParams = make(map[string]string)
				for k, v := range paramMap {
					if vStr, ok := v.(string); ok {
						queryParams[k] = vStr
					}
				}
			}
		}

		nodeConfig := &GitHubNodeConfig{
			GithubToken:  token,
			Repository:   repo,
			Endpoint:     endpoint,
			Method:       method,
			RequestBody:  requestBody,
			QueryParams:  queryParams,
		}

		return NewGitHubNode(nodeConfig), nil
	})
}