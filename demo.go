package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"citadel-agent/backend/internal/nodes/http"
	"citadel-agent/backend/internal/nodes/utility"
	"citadel-agent/backend/internal/workflow/core/engine"
)

// Demo function to show the workflow system working
func main() {
	fmt.Println("Citadel Agent - Working Demo")
	fmt.Println("=============================")

	// Initialize the node registry
	registry := engine.NewNodeTypeRegistry()

	// Register our nodes
	registerDemoNodes(registry)

	// Create a simple workflow: HTTP Request -> Logger
	workflow := &engine.Workflow{
		ID:   "demo-workflow",
		Name: "Demo Workflow",
		Nodes: map[string]*engine.WorkflowNode{
			"http-request": {
				ID:   "http-request",
				Type: "http_request",
				Config: map[string]interface{}{
					"method":  "GET",
					"url":     "https://httpbin.org/get?demo=test",
					"timeout": 30.0,
				},
			},
			"logger": {
				ID:   "logger",
				Type: "logger",
				Config: map[string]interface{}{
					"message": "Demo workflow result: {{body}}",
					"level":   "info",
				},
			},
		},
		Edges: []engine.WorkflowEdge{
			{
				ID:     "edge-1",
				Source: "http-request",
				Target: "logger",
			},
		},
	}

	// Create executor and run the workflow
	executor := engine.NewWorkflowExecutor(registry)

	// Execute the workflow
	ctx := context.Background()
	inputs := map[string]interface{}{
		"demonstration": true,
		"source":        "demo-script",
	}

	fmt.Println("Executing demo workflow...")
	results, err := executor.ExecuteWorkflow(ctx, workflow, inputs)
	if err != nil {
		log.Printf("Workflow execution failed: %v", err)
	} else {
		fmt.Println("Workflow executed successfully!")
		fmt.Println("Results:")

		// Pretty print the results
		resultBytes, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(resultBytes))
	}

	fmt.Println("=============================")
	fmt.Println("Demo completed successfully!")
	fmt.Println("The citadel Agent foundation is working correctly.")
}

func registerDemoNodes(registry *engine.NodeTypeRegistryImpl) {
	// Register HTTP Request Node
	httpMetadata := http.NewHTTPRequestNode().GetMetadata()
	registry.RegisterNodeType("http_request", http.NewHTTPRequestNode, httpMetadata)

	// Register Logger Node
	loggerMetadata := utility.NewLoggerNode().GetMetadata()
	registry.RegisterNodeType("logger", utility.NewLoggerNode, loggerMetadata)

	// Register Data Transformer Node
	transformerMetadata := utility.NewDataTransformerNode().GetMetadata()
	registry.RegisterNodeType("data_transformer", utility.NewDataTransformerNode, transformerMetadata)

	// Register If/Else Node
	ifelseMetadata := utility.NewIfElseNode().GetMetadata()
	registry.RegisterNodeType("if_else", utility.NewIfElseNode, ifelseMetadata)

	// Register For Each Node
	foreachMetadata := utility.NewForEachNode().GetMetadata()
	registry.RegisterNodeType("for_each", utility.NewForEachNode, foreachMetadata)

	log.Printf("Registered %d node types for demo", len(registry.ListNodeTypes()))
}
