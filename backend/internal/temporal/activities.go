package temporal

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
	"github.com/citadel-agent/backend/internal/workflow/core/engine"
	"go.temporal.io/sdk/activity"
)

// ExecuteNodeActivity executes a single node in the workflow
func ExecuteNodeActivity(ctx context.Context, input NodeInput) (NodeOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing node activity", "NodeID", input.NodeID, "NodeType", input.NodeType)

	startTime := time.Now()

	// Create a node instance based on the node type
	nodeInstance, err := createNodeInstance(input.NodeType, input.Config)
	if err != nil {
		return NodeOutput{
			NodeID:   input.NodeID,
			Status:   "failed",
			Error:    fmt.Sprintf("failed to create node instance: %v", err),
			Duration: time.Since(startTime),
		}, nil
	}

	// Execute the node
	output, err := nodeInstance.Execute(ctx, input.Variables)
	if err != nil {
		return NodeOutput{
			NodeID:   input.NodeID,
			Status:   "failed",
			Error:    fmt.Sprintf("node execution failed: %v", err),
			Duration: time.Since(startTime),
		}, nil
	}

	// Success case
	return NodeOutput{
		NodeID:   input.NodeID,
		Status:   "success",
		Output:   output,
		Duration: time.Since(startTime),
	}, nil
}

// createNodeInstance creates a node instance based on type and config
func createNodeInstance(nodeType string, config map[string]interface{}) (interfaces.NodeInstance, error) {
	// This function would typically use a registry to map node types to implementations
	// For now, we'll implement a simple mapping for common node types
	switch nodeType {
	case "http_request":
		return createHTTPRequestNode(config)
	case "condition":
		return createConditionNode(config)
	case "delay":
		return createDelayNode(config)
	case "database_query":
		return createDatabaseNode(config)
	case "script_execution":
		return createScriptNode(config)
	case "ai_agent":
		return createAIAgentNode(config)
	case "data_transformer":
		return createDataTransformerNode(config)
	case "notification":
		return createNotificationNode(config)
	case "loop":
		return createLoopNode(config)
	case "error_handler":
		return createErrorHandlerNode(config)
	default:
		// Try to use plugin system for custom node types
		return createPluginNode(nodeType, config)
	}
}

// createHTTPRequestNode creates an HTTP request node
func createHTTPRequestNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	return &HTTPRequestNode{Config: config}, nil
}

// createConditionNode creates a condition node
func createConditionNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	return &ConditionNode{Config: config}, nil
}

// createDelayNode creates a delay node
func createDelayNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	return &DelayNode{Config: config}, nil
}

// createDatabaseNode creates a database node
func createDatabaseNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	return &DatabaseNode{Config: config}, nil
}

// createScriptNode creates a script execution node
func createScriptNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	return &ScriptNode{Config: config}, nil
}

// createAIAgentNode creates an AI agent node
func createAIAgentNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	return &AIAgentNode{Config: config}, nil
}

// createDataTransformerNode creates a data transformer node
func createDataTransformerNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	return &DataTransformerNode{Config: config}, nil
}

// createNotificationNode creates a notification node
func createNotificationNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	return &NotificationNode{Config: config}, nil
}

// createLoopNode creates a loop node
func createLoopNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	return &LoopNode{Config: config}, nil
}

// createErrorHandlerNode creates an error handler node
func createErrorHandlerNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	return &ErrorHandlerNode{Config: config}, nil
}

// createPluginNode creates a plugin-based node
func createPluginNode(nodeType string, config map[string]interface{}) (interfaces.NodeInstance, error) {
	// This would integrate with the plugin system we created earlier
	// For now, return an error indicating the node type is not supported
	return nil, fmt.Errorf("node type '%s' not supported", nodeType)
}

// HTTPRequestNode represents an HTTP request node
type HTTPRequestNode struct {
	Config map[string]interface{}
}

func (n *HTTPRequestNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Simulate HTTP request execution
	result := map[string]interface{}{
		"result":    "HTTP request executed",
		"config":    n.Config,
		"inputs":    inputs,
		"timestamp": time.Now().Unix(),
	}
	return result, nil
}

// ConditionNode represents a condition node
type ConditionNode struct {
	Config map[string]interface{}
}

func (n *ConditionNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Simulate condition evaluation
	result := map[string]interface{}{
		"result":    "Condition evaluated",
		"config":    n.Config,
		"inputs":    inputs,
		"timestamp": time.Now().Unix(),
	}
	return result, nil
}

// DelayNode represents a delay/sleep node
type DelayNode struct {
	Config map[string]interface{}
}

func (n *DelayNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Simulate delay
	delaySecs := 1.0
	if secs, ok := n.Config["seconds"].(float64); ok {
		delaySecs = secs
	}

	time.Sleep(time.Duration(delaySecs) * time.Second)

	result := map[string]interface{}{
		"result":       "Delay completed",
		"delayed_by":   delaySecs,
		"config":       n.Config,
		"inputs":       inputs,
		"timestamp":    time.Now().Unix(),
	}
	return result, nil
}

// DatabaseNode represents a database query node
type DatabaseNode struct {
	Config map[string]interface{}
}

func (n *DatabaseNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Simulate database query
	result := map[string]interface{}{
		"result":    "Database query executed",
		"config":    n.Config,
		"inputs":    inputs,
		"timestamp": time.Now().Unix(),
	}
	return result, nil
}

// ScriptNode represents a script execution node
type ScriptNode struct {
	Config map[string]interface{}
}

func (n *ScriptNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Simulate script execution
	result := map[string]interface{}{
		"result":    "Script executed",
		"config":    n.Config,
		"inputs":    inputs,
		"timestamp": time.Now().Unix(),
	}
	return result, nil
}

// AIAgentNode represents an AI agent node
type AIAgentNode struct {
	Config map[string]interface{}
}

func (n *AIAgentNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Simulate AI agent execution
	result := map[string]interface{}{
		"result":    "AI agent executed",
		"config":    n.Config,
		"inputs":    inputs,
		"timestamp": time.Now().Unix(),
	}
	return result, nil
}

// DataTransformerNode represents a data transformer node
type DataTransformerNode struct {
	Config map[string]interface{}
}

func (n *DataTransformerNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Simulate data transformation
	result := map[string]interface{}{
		"result":    "Data transformed",
		"config":    n.Config,
		"inputs":    inputs,
		"timestamp": time.Now().Unix(),
	}
	return result, nil
}

// NotificationNode represents a notification node
type NotificationNode struct {
	Config map[string]interface{}
}

func (n *NotificationNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Simulate notification sending
	result := map[string]interface{}{
		"result":    "Notification sent",
		"config":    n.Config,
		"inputs":    inputs,
		"timestamp": time.Now().Unix(),
	}
	return result, nil
}

// LoopNode represents a loop node
type LoopNode struct {
	Config map[string]interface{}
}

func (n *LoopNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Simulate loop execution
	result := map[string]interface{}{
		"result":    "Loop executed",
		"config":    n.Config,
		"inputs":    inputs,
		"timestamp": time.Now().Unix(),
	}
	return result, nil
}

// ErrorHandlerNode represents an error handler node
type ErrorHandlerNode struct {
	Config map[string]interface{}
}

func (n *ErrorHandlerNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Simulate error handling
	result := map[string]interface{}{
		"result":    "Error handled",
		"config":    n.Config,
		"inputs":    inputs,
		"timestamp": time.Now().Unix(),
	}
	return result, nil
}