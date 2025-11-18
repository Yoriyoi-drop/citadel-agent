package plugins

import (
	"context"
	"time"
)

// PluginType defines the type of plugin
type PluginType string

const (
	JavascriptPlugin PluginType = "javascript"
	PythonPlugin     PluginType = "python"
	BuiltinPlugin    PluginType = "builtin"
)

// Plugin represents a single plugin
type Plugin struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Type        PluginType  `json:"type"`
	Code        string      `json:"code"`
	Schema      interface{} `json:"schema"` // JSON schema for plugin configuration
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// PluginResult represents the result of plugin execution
type PluginResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	ExecTime time.Duration `json:"exec_time,omitempty"`
}

// PluginExecutor defines the interface for executing plugins
type PluginExecutor interface {
	Execute(ctx context.Context, code string, input map[string]interface{}) (*PluginResult, error)
}

// ExecutionContext holds the context for plugin execution
type ExecutionContext struct {
	ID      string
	Timeout time.Duration
	Inputs  map[string]interface{}
	Outputs map[string]interface{}
}