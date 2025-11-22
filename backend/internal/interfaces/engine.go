// backend/internal/interfaces/engine.go
package interfaces

import (
	"context"
	"time"
)

// ExecutionResult represents the result of a node execution
type ExecutionResult struct {
	Status    string      `json:"status"`
	Data      interface{} `json:"data"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// NodeInstance interface for all node types
// This breaks the circular dependency between engine and nodes packages
type NodeInstance interface {
	Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error)
}