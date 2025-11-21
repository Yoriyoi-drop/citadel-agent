// backend/internal/engine/executor.go
package engine

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
	"github.com/google/uuid"
)

// NodeExecutor defines the interface for executing nodes
type NodeExecutor interface {
	Execute(ctx context.Context, input map[string]interface{}) (*ExecutionResult, error)
}

// ExecutionResult represents the result of a node execution
type ExecutionResult struct {
	Status    string      `json:"status"`
	Data      interface{} `json:"data"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// Executor manages the execution of workflow nodes
type Executor struct {
	nodeExecutors map[string]NodeExecutor
	aiManager    interfaces.AIManagerInterface // Use interface instead of concrete type
}

// NewExecutor creates a new instance of Executor
func NewExecutor(aiManager interfaces.AIManagerInterface) *Executor {
	executor := &Executor{
		nodeExecutors: make(map[string]NodeExecutor),
		aiManager:    aiManager,
	}

	// Register built-in nodes
	executor.RegisterNodeExecutor("http_request", &HTTPRequestNode{})
	executor.RegisterNodeExecutor("delay", &DelayNode{})
	executor.RegisterNodeExecutor("function", &FunctionNode{})
	executor.RegisterNodeExecutor("trigger", &TriggerNode{})
	executor.RegisterNodeExecutor("data_process", &DataProcessNode{})
	executor.RegisterNodeExecutor("ai_agent", &AIAgentNode{AIManager: aiManager})

	return executor
}

// RegisterNodeExecutor registers a node executor
func (e *Executor) RegisterNodeExecutor(nodeType string, executor NodeExecutor) {
	e.nodeExecutors[nodeType] = executor
}

// ExecuteNode executes a single node
func (e *Executor) ExecuteNode(ctx context.Context, nodeType string, input map[string]interface{}) (*ExecutionResult, error) {
	executor, exists := e.nodeExecutors[nodeType]
	if !exists {
		return &ExecutionResult{
			Status:    "error",
			Error:     fmt.Sprintf("Node type %s not registered", nodeType),
			Timestamp: time.Now(),
		}, nil
	}

	result, err := executor.Execute(ctx, input)
	if err != nil {
		return &ExecutionResult{
			Status:    "error",
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, nil
	}

	// Send result via WebSocket or other notification mechanism
	// In a real implementation, this would send the result to interested parties
	// For now, we'll just log it
	log.Printf("Execution result for node type %s: %+v", nodeType, result)

	return result, nil
}

// HTTPRequestNode executes HTTP requests
type HTTPRequestNode struct{}

func (h *HTTPRequestNode) Execute(ctx context.Context, input map[string]interface{}) (*ExecutionResult, error) {
	// Extract HTTP parameters from input
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

	// In a real implementation, this would make the actual HTTP request
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

// DelayNode implements a delay/wait node
type DelayNode struct{}

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

	// Create a channel to receive the completion signal
	done := make(chan bool, 1)

	// Execute delay in a goroutine to allow potential cancellation
	go func() {
		select {
		case <-time.After(delayDuration):
			done <- true
		case <-ctx.Done():
			// Context cancelled
			done <- false
		}
	}()

	// Wait for delay to complete or context cancellation
	completed := <-done

	if !completed {
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

// FunctionNode executes custom functions
type FunctionNode struct{}

func (f *FunctionNode) Execute(ctx context.Context, input map[string]interface{}) (*ExecutionResult, error) {
	// In a real implementation, this would execute JavaScript/Python code
	// For now, we'll just return the input with an additional field
	resultData := make(map[string]interface{})
	for k, v := range input {
		resultData[k] = v
	}
	resultData["processed_by"] = "function_node"
	resultData["timestamp"] = time.Now().Unix()

	return &ExecutionResult{
		Status:    "success",
		Data:      resultData,
		Timestamp: time.Now(),
	}, nil
}

// TriggerNode represents a trigger node that initiates workflows
type TriggerNode struct{}

func (t *TriggerNode) Execute(ctx context.Context, input map[string]interface{}) (*ExecutionResult, error) {
	// Extract trigger type from input
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

// DataProcessNode processes data based on operation type
type DataProcessNode struct{}

func (d *DataProcessNode) Execute(ctx context.Context, input map[string]interface{}) (*ExecutionResult, error) {
	// Extract operation type from input
	operation, exists := input["operation"].(string)
	if !exists {
		operation = "transform"
	}

	// Process data based on operation type
	var resultData interface{}
	switch operation {
	case "transform":
		// Transform the input data
		resultData = transformData(input)
	case "filter":
		// Filter the input data
		resultData = filterData(input)
	case "aggregate":
		// Aggregate the input data
		resultData = aggregateData(input)
	default:
		// For unknown operations, just return the input
		resultData = input
	}

	return &ExecutionResult{
		Status: "success",
		Data:   resultData,
		Timestamp: time.Now(),
	}, nil
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

// AIAgentNode executes an AI agent (now uses the interface)
type AIAgentNode struct {
	AIManager interfaces.AIManagerInterface
}

func (a *AIAgentNode) Execute(ctx context.Context, input map[string]interface{}) (*ExecutionResult, error) {
	agentID, exists := input["agent_id"].(string)
	if !exists {
		return &ExecutionResult{
			Status: "error",
			Error:  "agent_id is required for AI agent node",
			Timestamp: time.Now(),
		}, nil
	}

	agentInput, exists := input["input"].(string)
	if !exists {
		// If input is not a string, convert it to string
		agentInput = fmt.Sprintf("%v", input)
	}

	// Use the AI manager interface (which could be any AI implementation)
	if a.AIManager == nil {
		return &ExecutionResult{
			Status: "error",
			Error:  "AI manager not initialized for AIAgentNode",
			Timestamp: time.Now(),
		}, nil
	}

	result, err := a.AIManager.ExecuteAgent(ctx, agentID, map[string]interface{}{"input": agentInput})
	if err != nil {
		return &ExecutionResult{
			Status: "error",
			Error:  fmt.Sprintf("AI agent execution failed: %v", err),
			Timestamp: time.Now(),
		}, nil
	}

	executionResult := &ExecutionResult{
		Status: "success",
		Data:   result,
		Timestamp: time.Now(),
	}
	
	return executionResult, nil
}