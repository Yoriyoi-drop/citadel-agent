package main

import (
	"context"
	"testing"

	"citadel-agent/backend/internal/nodes/http"
	"citadel-agent/backend/internal/workflow/core/types"
)

func TestHTTPRequestNode(t *testing.T) {
	// Create a new HTTP request node
	node := http.NewHTTPRequestNode()
	
	// Initialize the node with configuration
	config := map[string]interface{}{
		"method":  "GET",
		"url":     "https://httpbin.org/get",
		"timeout": 30.0,
	}
	
	err := node.Initialize(config)
	if err != nil {
		t.Fatalf("Failed to initialize HTTP request node: %v", err)
	}
	
	// Validate the node
	err = node.Validate()
	if err != nil {
		t.Fatalf("HTTP request node validation failed: %v", err)
	}
	
	// Execute the node with empty input
	ctx := context.Background()
	input := types.NodeInput{Data: make(map[string]interface{})}
	output := node.Execute(ctx, input)
	
	// Check if execution was successful (it might fail due to network, but shouldn't panic)
	if output.Error != nil {
		// For this test, we'll allow network errors since we're testing external service
		t.Logf("HTTP request execution returned error (expected for external service): %v", output.Error)
	}
	
	// Test metadata
	metadata := node.GetMetadata()
	if metadata.ID != "http_request" {
		t.Errorf("Expected metadata ID 'http_request', got '%s'", metadata.ID)
	}
	if metadata.Name != "HTTP Request" {
		t.Errorf("Expected metadata name 'HTTP Request', got '%s'", metadata.Name)
	}
	
	// Test with invalid configuration
	invalidConfig := map[string]interface{}{
		"method": "INVALID_METHOD",
		"url":    "",
		"timeout": -1.0,
	}
	
	invalidNode := http.NewHTTPRequestNode()
	err = invalidNode.Initialize(invalidConfig)
	if err == nil {
		t.Error("Expected error when initializing with invalid config, but got none")
	}
	
	invalidErr := invalidNode.Validate()
	if invalidErr == nil {
		t.Error("Expected validation error with invalid config, but got none")
	}
}