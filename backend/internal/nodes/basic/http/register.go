package http

import (
	"citadel-agent/backend/internal/engine"
	"citadel-agent/backend/internal/nodes"
)

// RegisterHTTPNodes registers all HTTP-related nodes
func RegisterHTTPNodes(registry *nodes.NodeRegistry) {
	// Register the enhanced HTTP request node
	httpNode := NewHTTPRequestNode()
	
	registry.RegisterNode("http_request_enhanced", httpNode, &nodes.NodeDefinition{
		Name:        "Enhanced HTTP Request",
		Description: "Advanced HTTP client with security, timeout, and response handling",
		Type:        nodes.Basic,
		Category:    nodes.HTTP,
		Icon:        "http",
		SettingsSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"url": map[string]interface{}{
					"type":        "string",
					"description": "Target URL for the request",
					"format":      "uri",
				},
				"method": map[string]interface{}{
					"type":        "string",
					"description": "HTTP method (GET, POST, PUT, DELETE, etc.)",
					"enum":        []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
					"default":     "GET",
				},
				"headers": map[string]interface{}{
					"type":        "object",
					"description": "Request headers",
					"additionalProperties": map[string]interface{}{
						"type": "string",
					},
				},
				"body": map[string]interface{}{
					"type":        "object",
					"description": "Request body (for POST, PUT, PATCH)",
				},
			},
			"required": []string{"url"},
		},
	})
}