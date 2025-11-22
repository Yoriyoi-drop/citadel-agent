package runtimes

import (
	"context"
)


// CodeExecutionResult represents the result of code execution
type CodeExecutionResult struct {
	Output    string                 `json:"output"`
	Error     string                 `json:"error"`
	Success   bool                   `json:"success"`
	ExecutionTime int64              `json:"execution_time"`
	Resources map[string]interface{} `json:"resources,omitempty"`
}

