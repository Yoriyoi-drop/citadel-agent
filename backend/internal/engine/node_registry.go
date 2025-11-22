// backend/internal/engine/node_registry.go
package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
	"github.com/citadel-agent/backend/internal/nodes"
	"github.com/citadel-agent/backend/internal/runtimes"
)

// BuiltInNodeType represents the type of built-in nodes
type BuiltInNodeType string

const (
	HTTPNodeType        BuiltInNodeType = "http_request"
	DelayNodeType       BuiltInNodeType = "delay"
	FunctionNodeType    BuiltInNodeType = "function"
	TriggerNodeType     BuiltInNodeType = "trigger"
	DataProcessNodeType BuiltInNodeType = "data_process"
	// Multi-language runtime nodes
	GoNodeType          BuiltInNodeType = "go_code"
	JSNodeType          BuiltInNodeType = "javascript_code"
	PythonNodeType      BuiltInNodeType = "python_code"
	JavaNodeType        BuiltInNodeType = "java_code"
	RubyNodeType        BuiltInNodeType = "ruby_code"
	PHPNodeType         BuiltInNodeType = "php_code"
	RustNodeType        BuiltInNodeType = "rust_code"
	CSharpNodeType      BuiltInNodeType = "csharp_code"
	ShellNodeType       BuiltInNodeType = "shell_script"
	// AI nodes
	AIAgentNodeType     BuiltInNodeType = "ai_agent"
	MultiRuntimeNodeType BuiltInNodeType = "multi_runtime"
)

// NodeRegistry holds the registration of all available nodes
type NodeRegistry struct {
	nodes map[string]NodeExecutor
	aiManager interfaces.AIManagerInterface
	runtimeMgr *runtimes.MultiRuntimeManager
}

// NewNodeRegistry creates a new instance of NodeRegistry
func NewNodeRegistry(aiManager interfaces.AIManagerInterface, runtimeMgr *runtimes.MultiRuntimeManager) *NodeRegistry {
	registry := &NodeRegistry{
		nodes: make(map[string]NodeExecutor),
		aiManager: aiManager,
		runtimeMgr: runtimeMgr,
	}

	// Register built-in nodes
	registry.RegisterNode(string(HTTPNodeType), &HTTPRequestNode{})
	registry.RegisterNode(string(DelayNodeType), &DelayNode{})
	registry.RegisterNode(string(FunctionNodeType), &FunctionNode{})
	registry.RegisterNode(string(TriggerNodeType), &TriggerNode{})
	registry.RegisterNode(string(DataProcessNodeType), &DataProcessNode{})

	// Register multi-language runtime nodes
	registry.RegisterNode(string(GoNodeType), &nodes.GoNode{})
	registry.RegisterNode(string(JSNodeType), &nodes.JavaScriptNode{})
	registry.RegisterNode(string(PythonNodeType), &nodes.PythonNode{})
	registry.RegisterNode(string(JavaNodeType), &nodes.JavaNode{})
	registry.RegisterNode(string(RubyNodeType), &nodes.RubyNode{})
	registry.RegisterNode(string(PHPNodeType), &nodes.PHPNode{})
	registry.RegisterNode(string(RustNodeType), &nodes.RustNode{})
	registry.RegisterNode(string(CSharpNodeType), &nodes.CSharpNode{})
	registry.RegisterNode(string(ShellNodeType), &nodes.ShellNode{})

	// Register AI nodes with their dependencies
	registry.RegisterNode(string(AIAgentNodeType), &nodes.AIAgentNode{AIManager: aiManager})
	registry.RegisterNode(string(MultiRuntimeNodeType), &nodes.MultiRuntimeNode{RuntimeMgr: runtimeMgr})

	return registry
}

// RegisterNode registers a node executor
func (nr *NodeRegistry) RegisterNode(nodeType string, executor NodeExecutor) {
	nr.nodes[nodeType] = executor
}

// GetNodeExecutor retrieves a node executor by type
func (nr *NodeRegistry) GetNodeExecutor(nodeType string) (NodeExecutor, bool) {
	executor, exists := nr.nodes[nodeType]
	return executor, exists
}

// HTTPRequestNode represents an HTTP request node
type HTTPRequestNode struct{}

// Execute executes the HTTP request node
func (h *HTTPRequestNode) Execute(ctx context.Context, input map[string]interface{}) (*ExecutionResult, error) {
	// Extract settings from input
	method, _ := input["method"].(string)
	if method == "" {
		method = "GET"
	}

	url, exists := input["url"].(string)
	if !exists {
		return &ExecutionResult{
			Status:    "error",
			Error:     "URL is required for HTTP request node",
			Timestamp: time.Now(),
		}, nil
	}

	// In a real implementation, we would make the HTTP request here
	// For now, we'll simulate the request
	log.Printf("Executing HTTP request: %s %s", method, url)

	// Simulate some processing time
	time.Sleep(100 * time.Millisecond)

	// Simulate response data
	responseData := map[string]interface{}{
		"status": 200,
		"data":   fmt.Sprintf("Response from %s", url),
		"headers": map[string]string{
			"content-type": "application/json",
		},
	}

	return &ExecutionResult{
		Status:    "success",
		Data:      responseData,
		Timestamp: time.Now(),
	}, nil
}

// DelayNode represents a delay node
type DelayNode struct{}

// Execute executes the delay node
func (d *DelayNode) Execute(ctx context.Context, input map[string]interface{}) (*ExecutionResult, error) {
	// Extract delay duration from settings
	delaySecs, exists := input["seconds"].(float64)
	if !exists {
		// Try integer type as well
		if delaySecsInt, ok := input["seconds"].(int); ok {
			delaySecs = float64(delaySecsInt)
		} else {
			delaySecs = 1 // default to 1 second
		}
	}

	// Convert to time.Duration
	delayDuration := time.Duration(delaySecs * float64(time.Second))

	// Simulate delay
	select {
	case <-time.After(delayDuration):
		// Delay completed
	case <-ctx.Done():
		// Context cancelled
		return &ExecutionResult{
			Status:    "error",
			Error:     "Delay node cancelled",
			Timestamp: time.Now(),
		}, nil
	}

	return &ExecutionResult{
		Status: "success",
		Data: map[string]interface{}{
			"message":    "Delay completed",
			"duration":   delayDuration.String(),
			"input_data": input,
		},
		Timestamp: time.Now(),
	}, nil
}

// FunctionNode represents a function node that executes custom code
type FunctionNode struct{}

// Execute executes the function node
func (f *FunctionNode) Execute(ctx context.Context, input map[string]interface{}) (*ExecutionResult, error) {
	// In a real implementation, this would execute JavaScript/Python code from the input
	// For now, we'll simulate function execution

	code, exists := input["code"].(string)
	if !exists {
		// If no code is provided, just pass through input data
		return &ExecutionResult{
			Status:    "success",
			Data:      input,
			Timestamp: time.Now(),
		}, nil
	}

	// Simulate function execution
	log.Printf("Executing function: %s", code)

	// Just return the input with an additional field as an example
	resultData := make(map[string]interface{})
	for k, v := range input {
		resultData[k] = v
	}
	resultData["processed_by"] = "function_node"
	resultData["random_value"] = rand.Intn(100)

	return &ExecutionResult{
		Status:    "success",
		Data:      resultData,
		Timestamp: time.Now(),
	}, nil
}

// TriggerNode represents a trigger node that starts workflows
type TriggerNode struct{}

// Execute executes the trigger node
func (t *TriggerNode) Execute(ctx context.Context, input map[string]interface{}) (*ExecutionResult, error) {
	// Trigger nodes typically initialize workflows
	// They might check for conditions, receive webhooks, etc.

	// For this example, we'll just pass through the input
	triggerType, exists := input["trigger_type"].(string)
	if !exists {
		triggerType = "manual"
	}

	log.Printf("Trigger node executed with type: %s", triggerType)

	return &ExecutionResult{
		Status: "success",
		Data: map[string]interface{}{
			"trigger_type": triggerType,
			"timestamp":    time.Now().Unix(),
			"input_data":   input,
		},
		Timestamp: time.Now(),
	}, nil
}

// DataProcessNode represents a node for data processing operations
type DataProcessNode struct{}

// Execute executes the data processing node
func (d *DataProcessNode) Execute(ctx context.Context, input map[string]interface{}) (*ExecutionResult, error) {
	// Process data based on operation type
	operation, exists := input["operation"].(string)
	if !exists {
		operation = "transform" // default operation
	}

	log.Printf("Processing data with operation: %s", operation)

	// Perform different operations based on the operation type
	switch operation {
	case "transform":
		// Transform the input data
		resultData := transformData(input)
		return &ExecutionResult{
			Status:    "success",
			Data:      resultData,
			Timestamp: time.Now(),
		}, nil
	case "filter":
		// Filter the input data
		resultData := filterData(input)
		return &ExecutionResult{
			Status:    "success",
			Data:      resultData,
			Timestamp: time.Now(),
		}, nil
	case "aggregate":
		// Aggregate the input data
		resultData := aggregateData(input)
		return &ExecutionResult{
			Status:    "success",
			Data:      resultData,
			Timestamp: time.Now(),
		}, nil
	default:
		return &ExecutionResult{
			Status:    "error",
			Error:     fmt.Sprintf("Unknown operation: %s", operation),
			Timestamp: time.Now(),
		}, nil
	}
}

// Helper functions for data processing
func transformData(input map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Convert all values to strings as an example transformation
	for k, v := range input {
		switch val := v.(type) {
		case string:
			result[k] = val
		case int, int32, int64, float32, float64:
			result[k] = fmt.Sprintf("%v", val)
		case bool:
			result[k] = fmt.Sprintf("%v", val)
		default:
			// For complex types, marshal to JSON string
			jsonBytes, err := json.Marshal(val)
			if err != nil {
				result[k] = fmt.Sprintf("%v", val)
			} else {
				result[k] = string(jsonBytes)
			}
		}
	}

	return result
}

func filterData(input map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// For this example, filter out any keys that start with "temp_"
	for k, v := range input {
		if len(k) < 5 || k[:5] != "temp_" {
			result[k] = v
		}
	}

	return result
}

func aggregateData(input map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// For this example, count and sum all number values
	count := 0
	sum := 0.0

	for _, v := range input {
		switch val := v.(type) {
		case int:
			count++
			sum += float64(val)
		case int32:
			count++
			sum += float64(val)
		case int64:
			count++
			sum += float64(val)
		case float32:
			count++
			sum += float64(val)
		case float64:
			count++
			sum += val
		}
	}

	result["count"] = count
	result["sum"] = sum
	if count > 0 {
		result["average"] = sum / float64(count)
	}

	return result
}

