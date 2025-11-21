// backend/internal/interfaces/engine.go
package interfaces

import "time"

// ExecutionResult represents the result of a node execution
type ExecutionResult struct {
	Status    string      `json:"status"`
	Data      interface{} `json:"data"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}