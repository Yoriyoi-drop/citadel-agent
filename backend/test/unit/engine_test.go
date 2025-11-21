package test

import (
	"context"
	"testing"

	"github.com/citadel-agent/backend/internal/engine"
	"github.com/stretchr/testify/assert"
)

// TestNodeRegistryInitialization tests that the node registry initializes correctly
func TestNodeRegistryInitialization(t *testing.T) {
	// Create node registry
	registry := engine.NewNodeRegistry()

	// Verify that registry is created
	assert.NotNil(t, registry)
	
	// Verify that default nodes are registered
	definitions := registry.GetAllNodeDefinitions()
	assert.Greater(t, len(definitions), 0)
}

// TestGetNodeDefinition tests retrieving a specific node definition
func TestGetNodeDefinition(t *testing.T) {
	// Create node registry
	registry := engine.NewNodeRegistry()

	// Try to get a known node type
	def, err := registry.GetNodeDefinition("http_request")
	assert.NoError(t, err)
	assert.NotNil(t, def)
	assert.Equal(t, "http_request", def.Type)
	assert.Equal(t, "HTTP Request", def.Name)
}

// TestGetUnknownNodeDefinition tests retrieving an unknown node definition
func TestGetUnknownNodeDefinition(t *testing.T) {
	// Create node registry
	registry := engine.NewNodeRegistry()

	// Try to get an unknown node type
	def, err := registry.GetNodeDefinition("unknown_node")
	assert.Error(t, err)
	assert.Nil(t, def)
}

// TestExecutorInitialization tests that the executor initializes correctly
func TestExecutorInitialization(t *testing.T) {
	// Create node registry
	registry := engine.NewNodeRegistry()
	
	// Create executor
	executor := engine.NewExecutor(registry)

	// Verify that executor is created
	assert.NotNil(t, executor)
	assert.Equal(t, registry, executor.Registry)
}

// TestRunnerInitialization tests that the runner initializes correctly
func TestRunnerInitialization(t *testing.T) {
	// Create node registry
	registry := engine.NewNodeRegistry()
	
	// Create executor
	executor := engine.NewExecutor(registry)
	
	// Create runner
	runner := engine.NewRunner(executor)

	// Verify that runner is created
	assert.NotNil(t, runner)
	assert.Equal(t, executor, runner.Executor)
}

// TestNodeExecution tests executing a node through a workflow
func TestNodeExecution(t *testing.T) {
	// Create node registry
	registry := engine.NewNodeRegistry()

	// Create executor
	executor := engine.NewExecutor(registry)

	// Create a simple workflow with one node
	workflow := &engine.Workflow{
		ID:   "test-workflow-1",
		Name: "Test Workflow",
		Nodes: []*engine.Node{
			{
				ID:   "test-node-1",
				Type: "http_request",
				Input: map[string]interface{}{
					"url":    "https://httpbin.org/get",
					"method": "GET",
				},
			},
		},
	}

	// Execute the workflow which will execute the node
	err := executor.ExecuteWorkflow(context.Background(), workflow)
	// This should complete without error
	assert.NoError(t, err)

	// Check that the node status is updated
	node := workflow.Nodes[0]
	assert.Equal(t, "completed", node.Status)
	assert.NotNil(t, node.Output)
}

// TestWorkflowExecution tests executing a basic workflow
func TestWorkflowExecution(t *testing.T) {
	// Create node registry
	registry := engine.NewNodeRegistry()
	
	// Create executor
	executor := engine.NewExecutor(registry)
	
	// Create runner
	runner := engine.NewRunner(executor)

	// Create a simple workflow
	workflow := &engine.Workflow{
		ID:   "test-workflow-1",
		Name: "Test Workflow",
		Nodes: []*engine.Node{
			{
				ID:   "test-node-1",
				Type: "http_request",
				Input: map[string]interface{}{
					"url":    "https://httpbin.org/get",
					"method": "GET",
				},
			},
		},
	}

	// Execute the workflow
	err := runner.RunWorkflow(context.Background(), workflow)
	// This should complete without error
	assert.NoError(t, err)
	
	// Check that workflow nodes are executed
	for _, node := range workflow.Nodes {
		assert.Equal(t, "completed", node.Status)
		assert.NotNil(t, node.Output)
	}
}