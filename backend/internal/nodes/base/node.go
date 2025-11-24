package base

import (
	"context"
	"encoding/json"
	"time"
)

// NodeType represents the type of node
type NodeType string

const (
	NodeTypeHTTP        NodeType = "http"
	NodeTypeDatabase    NodeType = "database"
	NodeTypeAI          NodeType = "ai"
	NodeTypeTransform   NodeType = "transform"
	NodeTypeFlow        NodeType = "flow"
	NodeTypeValidation  NodeType = "validation"
	NodeTypeIntegration NodeType = "integration"
	NodeTypeStorage     NodeType = "storage"
	NodeTypeSecurity    NodeType = "security"
	NodeTypeUtility     NodeType = "utility"
)

// ExecutionContext provides runtime context for node execution
type ExecutionContext struct {
	WorkflowID  string
	ExecutionID string
	NodeID      string
	UserID      string
	Timeout     time.Duration
	Variables   map[string]interface{}
	Secrets     map[string]string
	Context     context.Context
	Logger      Logger
	StartTime   time.Time
}

// Logger interface for node logging
type Logger interface {
	Debug(msg string, fields map[string]interface{})
	Info(msg string, fields map[string]interface{})
	Warn(msg string, fields map[string]interface{})
	Error(msg string, err error, fields map[string]interface{})
}

// NodeInput represents an input port
type NodeInput struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Required    bool                   `json:"required"`
	Description string                 `json:"description"`
	Default     interface{}            `json:"default,omitempty"`
	Validation  map[string]interface{} `json:"validation,omitempty"`
}

// NodeOutput represents an output port
type NodeOutput struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Schema      interface{} `json:"schema,omitempty"`
}

// NodeConfig represents node configuration
type NodeConfig struct {
	Name        string                 `json:"name"`
	Label       string                 `json:"label"`
	Description string                 `json:"description"`
	Required    bool                   `json:"required"`
	Type        string                 `json:"type"` // string, number, boolean, select, password, etc.
	Default     interface{}            `json:"default,omitempty"`
	Options     []ConfigOption         `json:"options,omitempty"`
	Validation  map[string]interface{} `json:"validation,omitempty"`
}

// ConfigOption for select-type configs
type ConfigOption struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}

// NodeMetadata contains node metadata
type NodeMetadata struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Category    string       `json:"category"`
	Description string       `json:"description"`
	Version     string       `json:"version"`
	Author      string       `json:"author"`
	Icon        string       `json:"icon"`
	Color       string       `json:"color"`
	Inputs      []NodeInput  `json:"inputs"`
	Outputs     []NodeOutput `json:"outputs"`
	Config      []NodeConfig `json:"config"`
	Tags        []string     `json:"tags"`
	Deprecated  bool         `json:"deprecated"`
}

// ExecutionResult represents the result of node execution
type ExecutionResult struct {
	Success   bool                   `json:"success"`
	Data      map[string]interface{} `json:"data"`
	Error     string                 `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
}

// Node is the interface that all nodes must implement
type Node interface {
	// GetMetadata returns node metadata
	GetMetadata() NodeMetadata

	// Validate validates the node configuration
	Validate(config map[string]interface{}) error

	// Execute executes the node with given inputs
	Execute(ctx *ExecutionContext, inputs map[string]interface{}) (*ExecutionResult, error)

	// OnStart is called when the node starts (optional lifecycle hook)
	OnStart(ctx *ExecutionContext) error

	// OnStop is called when the node stops (optional lifecycle hook)
	OnStop(ctx *ExecutionContext) error
}

// BaseNode provides common functionality for all nodes
type BaseNode struct {
	metadata NodeMetadata
}

// NewBaseNode creates a new base node
func NewBaseNode(metadata NodeMetadata) *BaseNode {
	return &BaseNode{
		metadata: metadata,
	}
}

// GetMetadata returns node metadata
func (n *BaseNode) GetMetadata() NodeMetadata {
	return n.metadata
}

// Validate validates configuration (default implementation)
func (n *BaseNode) Validate(config map[string]interface{}) error {
	// Validate required fields
	for _, cfg := range n.metadata.Config {
		if cfg.Required {
			if _, exists := config[cfg.Name]; !exists {
				return &ValidationError{
					Field:   cfg.Name,
					Message: "required field missing",
				}
			}
		}
	}
	return nil
}

// OnStart default implementation (no-op)
func (n *BaseNode) OnStart(ctx *ExecutionContext) error {
	return nil
}

// OnStop default implementation (no-op)
func (n *BaseNode) OnStop(ctx *ExecutionContext) error {
	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// ExecutionError represents an execution error
type ExecutionError struct {
	NodeID  string
	Message string
	Cause   error
}

func (e *ExecutionError) Error() string {
	if e.Cause != nil {
		return e.NodeID + ": " + e.Message + " - " + e.Cause.Error()
	}
	return e.NodeID + ": " + e.Message
}

// Helper functions

// CreateSuccessResult creates a successful execution result
func CreateSuccessResult(data map[string]interface{}, duration time.Duration) *ExecutionResult {
	return &ExecutionResult{
		Success:   true,
		Data:      data,
		Duration:  duration,
		Timestamp: time.Now(),
	}
}

// CreateErrorResult creates an error execution result
func CreateErrorResult(err error, duration time.Duration) *ExecutionResult {
	return &ExecutionResult{
		Success:   false,
		Error:     err.Error(),
		Duration:  duration,
		Timestamp: time.Now(),
	}
}

// MarshalConfig marshals config to JSON
func MarshalConfig(config interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// UnmarshalConfig unmarshals config from map
func UnmarshalConfig(config map[string]interface{}, target interface{}) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, target)
}
