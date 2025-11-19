package http

import (
	"context"
	"testing"

	"citadel-agent/backend/internal/engine"
)

func TestEnhancedHTTPNode_Execute(t *testing.T) {
	node := &EnhancedHTTPNode{}
	
	// Test basic execution
	input := map[string]interface{}{
		"message": "test input",
	}
	
	result, err := node.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	if result.Status != "success" {
		t.Errorf("Expected status 'success', got: %s", result.Status)
	}
	
	if result.Data == nil {
		t.Error("Expected result data, got nil")
	}
	
	dataMap, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Error("Expected result data to be a map")
	}
	
	if msg, exists := dataMap["message"]; !exists || msg != "Enhanced HTTP functionality" {
		t.Error("Expected 'Enhanced HTTP functionality' message in result")
	}
	
	// Verify input is passed through
	if inputResult, exists := dataMap["input"]; !exists || inputResult != input {
		t.Error("Expected input to be passed through in result")
	}
}

// For testing the full HTTP functionality, we would need a test server, but for now
// we'll focus on the basic functionality since the main HTTPRequestNode is 
// implemented in the registry.go file
func TestHTTPRequestNodeInRegistry(t *testing.T) {
	// This test would ensure that the HTTPRequestNode in registry.go works properly
	// It's tested as part of the entire system, but we can still have a basic validation
	
	// Create a mock HTTP request node to test
	node := NewHTTPRequestNode()
	
	if node == nil {
		t.Error("NewHTTPRequestNode should not return nil")
	}
	
	if node.client == nil {
		t.Error("HTTPRequestNode should have an HTTP client initialized")
	}
	
	// Test with invalid input (no URL)
	invalidInput := map[string]interface{}{
		"method": "GET",
		// Missing URL
	}
	
	result, err := node.Execute(context.Background(), invalidInput)
	if err != nil {
		t.Errorf("Execute should not return error, got: %v", err)
	}
	
	if result.Status != "error" {
		t.Error("Expected error status when URL is missing")
	}
	
	if result.Error == "" {
		t.Error("Expected error message when URL is missing")
	}
}

// Additional tests for other HTTP node functionality would require a test server
// which can be implemented in further test iterations