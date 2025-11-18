package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/websocket/v1"
)

// ExecutionResult represents the result of a node execution
type ExecutionResult struct {
	NodeID    string      `json:"node_id"`
	Status    string      `json:"status"` // "success", "error", "running"
	Data      interface{} `json:"data"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// NodeExecutor defines the interface for executing a workflow node
type NodeExecutor interface {
	Execute(ctx context.Context, input map[string]interface{}) (*ExecutionResult, error)
}

// Executor manages the execution of workflow nodes
type Executor struct {
	nodeExecutors map[string]NodeExecutor
	webSocketConn *websocket.Conn
}

// NewExecutor creates a new instance of Executor
func NewExecutor() *Executor {
	return &Executor{
		nodeExecutors: make(map[string]NodeExecutor),
	}
}

// RegisterNodeExecutor registers a node executor for a specific node type
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

	return result, nil
}

// SetWebSocketConn sets the WebSocket connection for real-time updates
func (e *Executor) SetWebSocketConn(conn *websocket.Conn) {
	e.webSocketConn = conn
}

// SendUpdate sends execution updates via WebSocket if available
func (e *Executor) SendUpdate(result *ExecutionResult) {
	if e.webSocketConn != nil {
		data, err := json.Marshal(result)
		if err != nil {
			log.Printf("Error marshaling execution result: %v", err)
			return
		}
		
		if err := e.webSocketConn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("Error sending WebSocket update: %v", err)
		}
	}
}