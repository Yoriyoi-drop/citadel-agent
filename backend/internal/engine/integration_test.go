package engine

import (
	"context"
	"testing"
)

func TestIntegration_FullWorkflowExecution(t *testing.T) {
	// Create a more complex workflow: A -> B -> D, A -> C -> D
	workflow := &Workflow{
		ID:   "integration-test-workflow",
		Name: "Integration Test Workflow",
		Nodes: []Node{
			{ID: "start", Type: "mock_success", Name: "Start Node"},
			{ID: "process1", Type: "mock_success", Name: "Process Node 1"},
			{ID: "process2", Type: "mock_success", Name: "Process Node 2"},
			{ID: "merge", Type: "mock_success", Name: "Merge Node"},
		},
		Edges: []Edge{
			{ID: "e1", Source: "start", Target: "process1"},
			{ID: "e2", Source: "start", Target: "process2"},
			{ID: "e3", Source: "process1", Target: "merge"},
			{ID: "e4", Source: "process2", Target: "merge"},
		},
	}

	// Create executor with mock success node
	executor := NewExecutor()
	mockSuccess := &MockNodeExecutor{
		result: &ExecutionResult{
			Status: "success",
			Data:   map[string]interface{}{"processed": true},
		},
	}
	executor.RegisterNodeExecutor("mock_success", mockSuccess)

	// Create runner
	runner := NewRunner(executor)

	// Run workflow
	ctx := context.Background()
	variables := map[string]interface{}{
		"test_id": "integration_test",
		"data":    "initial_data",
	}

	execution, err := runner.RunWorkflow(ctx, workflow, variables)
	if err != nil {
		t.Fatalf("Expected no error during workflow execution, got: %v", err)
	}

	if execution.Status != "completed" {
		t.Errorf("Expected execution status 'completed', got: %s", execution.Status)
	}

	// Verify all nodes executed
	expectedNodeCount := 4
	if len(execution.Results) != expectedNodeCount {
		t.Errorf("Expected %d node results, got %d", expectedNodeCount, len(execution.Results))
	}

	// Verify variables were preserved
	if execution.Variables["test_id"] != "integration_test" {
		t.Error("Workflow variables not properly preserved")
	}

	// Check that execution time is reasonable (not negative)
	if execution.EndedAt == nil || execution.StartedAt.After(*execution.EndedAt) {
		t.Error("Execution timing seems incorrect")
	}
}

func TestIntegration_FailedWorkflowExecution(t *testing.T) {
	// Create a workflow where one node fails: A -> B (fails) -> C
	workflow := &Workflow{
		ID:   "failed-workflow-test",
		Name: "Failed Workflow Test",
		Nodes: []Node{
			{ID: "start", Type: "mock_success", Name: "Start Node"},
			{ID: "failing", Type: "mock_fail", Name: "Failing Node"},
			{ID: "end", Type: "mock_success", Name: "End Node"},
		},
		Edges: []Edge{
			{ID: "e1", Source: "start", Target: "failing"},
			{ID: "e2", Source: "failing", Target: "end"},
		},
	}

	// Create executor with mixed success/failure nodes
	executor := NewExecutor()
	
	mockSuccess := &MockNodeExecutor{
		result: &ExecutionResult{
			Status: "success",
			Data:   map[string]interface{}{"result": "success"},
		},
	}
	
	mockFail := &MockNodeExecutor{
		result: &ExecutionResult{
			Status: "error",
			Error:  "Intentional test failure",
		},
	}
	
	executor.RegisterNodeExecutor("mock_success", mockSuccess)
	executor.RegisterNodeExecutor("mock_fail", mockFail)

	// Create runner
	runner := NewRunner(executor)

	// Run workflow - should fail
	ctx := context.Background()
	variables := map[string]interface{}{}

	_, err := runner.RunWorkflow(ctx, workflow, variables)
	if err == nil {
		t.Error("Expected error for workflow with failing node, got nil")
	}

	// The workflow should stop execution after the failing node,
	// so only the first node should have executed
	// This is verified through the execution result in a full implementation
}

func TestIntegration_EmptyWorkflow(t *testing.T) {
	// Create a workflow with no nodes
	workflow := &Workflow{
		ID:    "empty-workflow-test",
		Name:  "Empty Workflow Test",
		Nodes: []Node{}, // No nodes
		Edges: []Edge{}, // No edges
	}

	executor := NewExecutor()
	runner := NewRunner(executor)

	ctx := context.Background()
	variables := map[string]interface{}{"test": "empty"}

	execution, err := runner.RunWorkflow(ctx, workflow, variables)
	if err != nil {
		t.Errorf("Expected no error for empty workflow, got: %v", err)
	}

	if execution.Status != "completed" {
		t.Errorf("Expected completed status for empty workflow, got: %s", execution.Status)
	}

	if len(execution.Results) != 0 {
		t.Errorf("Expected 0 results for empty workflow, got %d", len(execution.Results))
	}
}

func TestIntegration_SingleNodeWorkflow(t *testing.T) {
	// Create a workflow with a single node
	workflow := &Workflow{
		ID:   "single-node-test",
		Name: "Single Node Test",
		Nodes: []Node{
			{ID: "single", Type: "mock_success", Name: "Single Node"},
		},
		Edges: []Edge{}, // No edges
	}

	executor := NewExecutor()
	mockSuccess := &MockNodeExecutor{
		result: &ExecutionResult{
			Status: "success",
			Data:   map[string]interface{}{"result": "single_node_result"},
		},
	}
	executor.RegisterNodeExecutor("mock_success", mockSuccess)

	runner := NewRunner(executor)

	ctx := context.Background()
	variables := map[string]interface{}{"input": "single_node_input"}

	execution, err := runner.RunWorkflow(ctx, workflow, variables)
	if err != nil {
		t.Errorf("Expected no error for single node workflow, got: %v", err)
	}

	if execution.Status != "completed" {
		t.Errorf("Expected completed status for single node workflow, got: %s", execution.Status)
	}

	if len(execution.Results) != 1 {
		t.Errorf("Expected 1 result for single node workflow, got %d", len(execution.Results))
	}

	// The single node should have been executed
	if _, exists := execution.Results["single"]; !exists {
		t.Error("Expected single node result to exist")
	}
}