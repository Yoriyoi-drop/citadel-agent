package engine

import (
	"context"
	"testing"
	"time"
)

// MockNodeExecutor for testing purposes
type MockNodeExecutor struct {
	result *ExecutionResult
	delay  time.Duration
}

func (m *MockNodeExecutor) Execute(ctx context.Context, input map[string]interface{}) (*ExecutionResult, error) {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	
	if m.result != nil {
		return m.result, nil
	}
	
	// Default success result
	return &ExecutionResult{
		Status: "success",
		Data:   input,
	}, nil
}

func TestRunner_SimpleWorkflow(t *testing.T) {
	// Create a simple workflow: A -> B
	workflow := &Workflow{
		ID:   "test-workflow",
		Name: "Test Workflow",
		Nodes: []Node{
			{
				ID:   "A",
				Type: "mock",
				Name: "Node A",
			},
			{
				ID:   "B", 
				Type: "mock",
				Name: "Node B",
			},
		},
		Edges: []Edge{
			{
				ID:     "e1",
				Source: "A",
				Target: "B",
			},
		},
	}

	// Create executor and register mock node executor
	executor := NewExecutor()
	mockExecutor := &MockNodeExecutor{}
	executor.RegisterNodeExecutor("mock", mockExecutor)

	// Create runner
	runner := NewRunner(executor)

	// Run workflow
	ctx := context.Background()
	variables := map[string]interface{}{"test": "value"}
	
	execution, err := runner.RunWorkflow(ctx, workflow, variables)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if execution.Status != "completed" {
		t.Errorf("Expected execution status 'completed', got: %s", execution.Status)
	}

	// Check that both nodes were executed
	if len(execution.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(execution.Results))
	}
}

func TestRunner_ParallelWorkflow(t *testing.T) {
	// Create a parallel workflow: A -> B, A -> C
	workflow := &Workflow{
		ID:   "test-parallel-workflow",
		Name: "Test Parallel Workflow",
		Nodes: []Node{
			{
				ID:   "A",
				Type: "mock",
				Name: "Node A",
			},
			{
				ID:   "B",
				Type: "mock",
				Name: "Node B",
			},
			{
				ID:   "C",
				Type: "mock",
				Name: "Node C",
			},
		},
		Edges: []Edge{
			{
				ID:     "e1",
				Source: "A",
				Target: "B",
			},
			{
				ID:     "e2",
				Source: "A", 
				Target: "C",
			},
		},
	}

	// Create executor and register mock node executor
	executor := NewExecutor()
	mockExecutor := &MockNodeExecutor{}
	executor.RegisterNodeExecutor("mock", mockExecutor)

	// Create runner
	runner := NewRunner(executor)

	// Run workflow
	ctx := context.Background()
	variables := map[string]interface{}{}

	execution, err := runner.RunWorkflow(ctx, workflow, variables)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if execution.Status != "completed" {
		t.Errorf("Expected execution status 'completed', got: %s", execution.Status)
	}

	// Check that all 3 nodes were executed
	if len(execution.Results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(execution.Results))
	}
}

func TestRunner_FailingNode(t *testing.T) {
	// Create a simple workflow with a failing node
	workflow := &Workflow{
		ID:   "test-failing-workflow",
		Name: "Test Failing Workflow",
		Nodes: []Node{
			{
				ID:   "A",
				Type: "mock",
				Name: "Node A",
			},
			{
				ID:   "B",
				Type: "mock", 
				Name: "Node B",
			},
		},
		Edges: []Edge{
			{
				ID:     "e1",
				Source: "A",
				Target: "B",
			},
		},
	}

	// Create executor with a failing node
	executor := NewExecutor()
	failingExecutor := &MockNodeExecutor{
		result: &ExecutionResult{
			Status: "error",
			Error:  "Node failed intentionally",
		},
	}
	executor.RegisterNodeExecutor("mock", failingExecutor)

	// Create runner
	runner := NewRunner(executor)

	// Run workflow - this should fail
	ctx := context.Background()
	variables := map[string]interface{}{}

	execution, err := runner.RunWorkflow(ctx, workflow, variables)
	if err == nil {
		t.Error("Expected error when node fails, got nil")
	}

	if execution.Status != "failed" {
		t.Errorf("Expected execution status 'failed', got: %s", execution.Status)
	}
}

func TestRunner_CircularDependency(t *testing.T) {
	// Create a workflow with circular dependency: A -> B -> C -> A
	workflow := &Workflow{
		ID:   "test-circular-workflow",
		Name: "Test Circular Workflow",
		Nodes: []Node{
			{
				ID:   "A",
				Type: "mock",
				Name: "Node A",
			},
			{
				ID:   "B",
				Type: "mock",
				Name: "Node B",
			},
			{
				ID:   "C",
				Type: "mock",
				Name: "Node C",
			},
		},
		Edges: []Edge{
			{
				ID:     "e1",
				Source: "A",
				Target: "B",
			},
			{
				ID:     "e2",
				Source: "B",
				Target: "C",
			},
			{
				ID:     "e3",
				Source: "C",
				Target: "A", // Creates circular dependency
			},
		},
	}

	// Create executor
	executor := NewExecutor()
	mockExecutor := &MockNodeExecutor{}
	executor.RegisterNodeExecutor("mock", mockExecutor)

	// Create runner
	runner := NewRunner(executor)

	// Run workflow - this should fail due to circular dependency
	ctx := context.Background()
	variables := map[string]interface{}{}

	_, err := runner.RunWorkflow(ctx, workflow, variables)
	if err == nil {
		t.Error("Expected error for circular dependency, got nil")
	}
}

func TestRunner_ExecutionContext(t *testing.T) {
	// Test the ExecutionContext functionality
	ctx := context.Background()
	workflow := &Workflow{
		ID:   "test-workflow",
		Name: "Test Workflow",
		Nodes: []Node{
			{ID: "A", Type: "test", Name: "Node A"},
		},
		Edges: []Edge{},
	}
	variables := map[string]interface{}{"key": "value"}

	executionCtx := NewExecutionContext(ctx, workflow, variables)

	// Test variable management
	if val, exists := executionCtx.GetVariable("key"); !exists || val != "value" {
		t.Error("Failed to set/get variable in execution context")
	}

	executionCtx.UpdateVariable("newKey", "newValue")
	if val, exists := executionCtx.GetVariable("newKey"); !exists || val != "newValue" {
		t.Error("Failed to update/get new variable in execution context")
	}

	// Test node result management
	result := &ExecutionResult{
		NodeID: "A",
		Status: "success",
		Data:   "test data",
	}
	executionCtx.AddNodeResult("A", result)

	if res, exists := executionCtx.GetNodeResult("A"); !exists || res.Status != "success" {
		t.Error("Failed to add/get node result in execution context")
	}

	// Test completion
	executionCtx.Complete()
	if executionCtx.EndedAt == nil {
		t.Error("Execution context not properly completed")
	}

	if executionCtx.HasFailed() {
		t.Error("Execution context incorrectly marked as failed")
	}
}