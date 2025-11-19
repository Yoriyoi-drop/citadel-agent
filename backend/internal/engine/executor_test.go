package engine

import (
	"context"
	"testing"
)

func TestExecutor_RegisterAndExecuteNode(t *testing.T) {
	executor := NewExecutor()
	
	// Create a mock node executor
	mockExecutor := &MockNodeExecutor{
		result: &ExecutionResult{
			Status: "success",
			Data:   map[string]interface{}{"result": "test"},
		},
	}
	
	// Register the executor
	executor.RegisterNodeExecutor("test_node", mockExecutor)
	
	// Execute a node
	input := map[string]interface{}{"input": "test"}
	result, err := executor.ExecuteNode(context.Background(), "test_node", input)
	
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	if result.Status != "success" {
		t.Errorf("Expected status 'success', got: %s", result.Status)
	}
	
	if result.Data == nil {
		t.Error("Expected result data, got nil")
	}
}

func TestExecutor_UnregisteredNode(t *testing.T) {
	executor := NewExecutor()
	
	// Try to execute an unregistered node
	input := map[string]interface{}{"input": "test"}
	result, err := executor.ExecuteNode(context.Background(), "unregistered_node", input)
	
	// Should not return error but should return error result
	if err != nil {
		t.Errorf("ExecuteNode should not return error for unregistered node, got: %v", err)
	}
	
	if result.Status != "error" {
		t.Errorf("Expected status 'error' for unregistered node, got: %s", result.Status)
	}
	
	if result.Error == "" {
		t.Error("Expected error message for unregistered node")
	}
}

func TestExecutor_MultipleNodeTypes(t *testing.T) {
	executor := NewExecutor()
	
	// Register multiple node executors
	executorA := &MockNodeExecutor{
		result: &ExecutionResult{
			Status: "success",
			Data:   map[string]interface{}{"node": "A"},
		},
	}
	
	executorB := &MockNodeExecutor{
		result: &ExecutionResult{
			Status: "success", 
			Data:   map[string]interface{}{"node": "B"},
		},
	}
	
	executor.RegisterNodeExecutor("node_A", executorA)
	executor.RegisterNodeExecutor("node_B", executorB)
	
	// Execute both nodes
	resultA, err := executor.ExecuteNode(context.Background(), "node_A", nil)
	if err != nil {
		t.Errorf("Expected no error for node A, got: %v", err)
	}
	
	resultB, err := executor.ExecuteNode(context.Background(), "node_B", nil)
	if err != nil {
		t.Errorf("Expected no error for node B, got: %v", err)
	}
	
	if resultA.Status != "success" || resultB.Status != "success" {
		t.Error("Expected both nodes to succeed")
	}
	
	// Check that data is correct for each node
	if dataA, ok := resultA.Data.(map[string]interface{}); !ok || dataA["node"] != "A" {
		t.Error("Expected node A result to contain 'A'")
	}
	
	if dataB, ok := resultB.Data.(map[string]interface{}); !ok || dataB["node"] != "B" {
		t.Error("Expected node B result to contain 'B'")
	}
}

func TestExecutionContext_Lifecycle(t *testing.T) {
	ctx := context.Background()
	workflow := &Workflow{
		ID:   "test-workflow",
		Name: "Test Workflow",
		Nodes: []Node{
			{ID: "A", Type: "test", Name: "Node A"},
			{ID: "B", Type: "test", Name: "Node B"},
		},
		Edges: []Edge{},
	}
	variables := map[string]interface{}{"initial": "value"}
	
	// Create execution context
	executionCtx := NewExecutionContext(ctx, workflow, variables)
	
	// Check initial state
	if executionCtx.ID == "" {
		t.Error("Execution context should have an ID")
	}
	
	if executionCtx.Workflow == nil {
		t.Error("Execution context should have a workflow")
	}
	
	if len(executionCtx.Variables) != 1 {
		t.Error("Execution context should have initial variables")
	}
	
	// Test variable operations
	value, exists := executionCtx.GetVariable("initial")
	if !exists || value != "value" {
		t.Error("Initial variable not found or incorrect")
	}
	
	// Add a new variable
	executionCtx.UpdateVariable("new_var", "new_value")
	value, exists = executionCtx.GetVariable("new_var")
	if !exists || value != "new_value" {
		t.Error("New variable not found or incorrect")
	}
	
	// Test node result operations
	resultA := &ExecutionResult{
		NodeID: "A",
		Status: "success",
		Data:   "result A",
	}
	
	// Add result
	executionCtx.AddNodeResult("A", resultA)
	
	// Verify result was added
	result, exists := executionCtx.GetNodeResult("A")
	if !exists {
		t.Error("Node result A was not stored")
	}
	
	if result == nil || result.Status != "success" {
		t.Error("Node result A is incorrect")
	}
	
	// Verify state updated
	if executionCtx.State.ProcessedNodes != 1 {
		t.Error("Processed node count not updated")
	}
	
	// Test cancellation
	executionCtx.Cancel()
	if !executionCtx.IsCancelled() {
		t.Error("Execution context should be cancelled")
	}
	
	// Test completion
	executionCtx.Complete()
	if executionCtx.EndedAt == nil {
		t.Error("Execution context should have ended at time")
	}
	
	// Get final execution result
	finalResult := executionCtx.GetExecutionResult()
	if finalResult == nil {
		t.Error("Final execution result should not be nil")
	}
	
	if finalResult.Status != "cancelled" {
		t.Error("Final status should be cancelled due to context cancellation")
	}
}

func TestExecutionContext_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	workflow := &Workflow{
		ID:   "test-workflow",
		Name: "Test Workflow",
		Nodes: []Node{{ID: "A", Type: "test", Name: "Node A"}},
		Edges: []Edge{},
	}
	
	executionCtx := NewExecutionContext(ctx, workflow, nil)
	
	// Test error handling
	executionCtx.SetError("test error")
	if !executionCtx.HasFailed() {
		t.Error("Execution context should have failed after setting error")
	}
	
	if executionCtx.State.LastError != "test error" {
		t.Error("Error message not set correctly")
	}
	
	// Get execution result after error
	finalResult := executionCtx.GetExecutionResult()
	if finalResult.Error != "test error" {
		t.Error("Error not propagated to final result")
	}
	
	if finalResult.Status != "failed" {
		t.Error("Status should be failed when error is set")
	}
}