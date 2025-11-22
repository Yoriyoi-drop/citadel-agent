package plugins

import (
	"context"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/workflow/core/engine"
)

// Example demonstrates how to use the plugin-aware engine
func Example() {
	// Create base engine
	baseEngine := engine.NewEngine(&engine.Config{
		Parallelism: 5,
		// Logger and Storage need to be configured
	})

	// Create plugin manager
	pluginManager := NewNodeManager()

	// Create plugin-aware engine
	pluginAwareEngine := NewPluginAwareEngine(baseEngine, pluginManager)

	// Register local node types (existing nodes)
	pluginAwareEngine.RegisterLocalNodeType("http_request", engine.NewHTTPRequestNode)
	pluginAwareEngine.RegisterLocalNodeType("condition", engine.NewConditionNode)
	pluginAwareEngine.RegisterLocalNodeType("delay", engine.NewDelayNode)

	// Example: If we have a plugin file at ./plugins/security_plugin, we would register it like this:
	// err := pluginManager.RegisterPluginAtPath("security_operation", "./plugins/security_plugin")
	// if err != nil {
	//     fmt.Printf("Failed to register plugin: %v\n", err)
	//     return
	// }
	//
	// err = pluginAwareEngine.RegisterPluginNodeType("security_operation")
	// if err != nil {
	//     fmt.Printf("Failed to register plugin node type: %v\n", err)
	//     return
	// }

	// Create a sample workflow with both local and plugin nodes
	workflow := &engine.Workflow{
		ID:          "test-workflow-1",
		Name:        "Test Workflow",
		Description: "Workflow demonstrating local and plugin nodes",
		Nodes: []*engine.Node{
			{
				ID:   "node-1",
				Type: "http_request", // Local node
				Name: "HTTP Request",
				Config: map[string]interface{}{
					"url": "https://api.example.com/data",
					"method": "GET",
				},
			},
			{
				ID:   "node-2", 
				Type: "security_operation", // Plugin node (if registered)
				Name: "Security Operation",
				Config: map[string]interface{}{
					"operation": "hash",
					"algorithm": "sha256",
					"data": "sensitive data",
				},
				Dependencies: []string{"node-1"},
			},
		},
		Connections: []*engine.Connection{
			{
				SourceNodeID: "node-1",
				TargetNodeID: "node-2",
			},
		},
	}

	// Execute the workflow with plugin support
	ctx := context.Background()
	executionID, err := pluginAwareEngine.ExecuteWithPlugins(ctx, workflow, nil)
	if err != nil {
		fmt.Printf("Failed to execute workflow: %v\n", err)
		return
	}

	fmt.Printf("Started execution with ID: %s\n", executionID)

	// Wait for execution to complete (in real usage, you'd have a different mechanism to track completion)
	time.Sleep(5 * time.Second)

	// Get the execution result
	result, err := baseEngine.GetExecution(executionID)
	if err != nil {
		fmt.Printf("Failed to get execution result: %v\n", err)
		return
	}

	fmt.Printf("Execution %s completed with status: %s\n", executionID, result.Status)
}

// MigrationPath shows the steps to migrate from local nodes to plugin nodes
func MigrationPath() {
	fmt.Println("Migration Path from Local Nodes to Plugin Nodes:")
	fmt.Println("1. Identify existing node types that should be converted to plugins")
	fmt.Println("2. Create plugin implementations for each node type")
	fmt.Println("3. Register plugin paths with the PluginManager")
	fmt.Println("4. Update workflow definitions to use plugin node IDs")
	fmt.Println("5. Maintain backward compatibility by keeping local nodes during transition")
	fmt.Println("6. Gradually phase out local nodes once all workflows use plugins")
}