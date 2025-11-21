package runtimes

import (
	"context"
	"time"
)

// Runtime interface defines the contract for language runtimes
type Runtime interface {
	ExecuteCode(ctx context.Context, code string, inputs map[string]interface{}, timeout time.Duration) (map[string]interface{}, error)
	ValidateCode(code string) error
	Initialize() error
	Dispose() error
	GetInfo() RuntimeInfo
}

// RuntimeInfo contains information about a runtime
type RuntimeInfo struct {
	Name    string
	Version string
	Status  string
	Stats   RuntimeStats
}

// RuntimeStats contains runtime statistics
type RuntimeStats struct {
	Executions       int64
	Errors           int64
	TotalTime        time.Duration
	AvgExecutionTime time.Duration
}