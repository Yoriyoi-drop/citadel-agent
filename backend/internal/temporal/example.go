package temporal

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/citadel-agent/backend/internal/workflow/core/engine"
)

// Example demonstrates how to use the Temporal integration
func Example() {
	// Create base engine
	baseEngine := engine.NewEngine(&engine.Config{
		Parallelism: 5,
		// Logger and Storage need to be configured
	})

	// Create Temporal client
	config := GetDefaultConfig()
	config.Address = "localhost:7233" // Use your Temporal server address
	config.Namespace = "default"
	
	temporalClient, err := NewTemporalClient(&Config{
		Address:   config.Address,
		Namespace: config.Namespace,
	})
	if err != nil {
		log.Fatal("Failed to create Temporal client:", err)
	}

	// Create Temporal workflow service
	workflowService := NewTemporalWorkflowService(temporalClient, baseEngine)
	workflowService.RegisterNodeTypes()

	// Define a sample workflow
	workflowDef := &WorkflowDefinition{
		ID:          "sample-workflow",
		Name:        "Sample Workflow",
		Description: "A sample workflow demonstrating Temporal integration",
		Options: WorkflowOptions{
			Parallelism:   3,
			Timeout:       time.Minute * 10,
			RetryAttempts: 3,
			ErrorHandling: "continue",
			RetryPolicy: RetryPolicy{
				InitialInterval:    time.Second,
				BackoffCoefficient: 2.0,
				MaximumInterval:    time.Minute,
				MaximumAttempts:    3,
			},
		},
		Nodes: []NodeDefinition{
			{
				ID:   "http-node-1",
				Type: "http_request",
				Name: "HTTP Request Node",
				Config: map[string]interface{}{
					"url":    "https://httpbin.org/get",
					"method": "GET",
				},
				Inputs: map[string]interface{}{
					"headers": map[string]string{
						"Content-Type": "application/json",
					},
				},
				Options: NodeExecutionOptions{
					RetryAttempts:   3,
					Timeout:         time.Second * 30,
					RetryOnFailure:  true,
					MaxConcurrent:   1,
				},
			},
			{
				ID:   "condition-node-1",
				Type: "condition",
				Name: "Condition Node",
				Config: map[string]interface{}{
					"expression": "response.status_code == 200",
				},
				Options: NodeExecutionOptions{
					RetryAttempts: 1,
					Timeout:       time.Second * 10,
				},
			},
			{
				ID:   "delay-node-1",
				Type: "delay",
				Name: "Delay Node",
				Config: map[string]interface{}{
					"seconds": 2.0,
				},
				Options: NodeExecutionOptions{
					RetryAttempts: 1,
					Timeout:       time.Second * 5,
				},
			},
		},
		Connections: []ConnectionDefinition{
			{
				SourceNodeID: "http-node-1",
				TargetNodeID: "condition-node-1",
			},
			{
				SourceNodeID: "condition-node-1",
				TargetNodeID: "delay-node-1",
			},
		},
	}

	// Register the workflow definition
	workflowService.RegisterWorkflowDefinition(workflowDef)

	// Execute the workflow
	params := map[string]interface{}{
		"user_id": 123,
		"action":  "test",
	}

	workflowRunID, err := workflowService.ExecuteWorkflowWithDefinition(
		context.Background(),
		workflowDef,
		params,
	)
	if err != nil {
		log.Fatal("Failed to execute workflow:", err)
	}

	fmt.Printf("Started workflow with RunID: %s\n", workflowRunID)

	// In a real scenario, you would monitor the workflow execution
	// For this example, we'll just wait a bit
	time.Sleep(5 * time.Second)

	// The workflow execution will continue in the Temporal cluster
	// You can query its status, get results, or signal it as needed
}

// AdvancedExample demonstrates more advanced usage with custom error handling
func AdvancedExample() {
	// This would include examples of:
	// - Custom retry policies
	// - Circuit breakers
	// - Workflow signaling
	// - Querying workflow state
	// - Error handling strategies
	fmt.Println("Advanced Temporal integration example")
	
	// Example of how to handle workflow results
	// result, err := workflowService.GetWorkflowResult(ctx, workflowID, runID)
	// if err != nil {
	//     fmt.Printf("Error getting workflow result: %v\n", err)
	// } else {
	//     fmt.Printf("Workflow result: %+v\n", result)
	// }
}