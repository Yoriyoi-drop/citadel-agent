package types

import (
	"context"
)

// NodeInput represents the input data for a node
type NodeInput struct {
	Data map[string]interface{}
}

// NodeOutput represents the output data from a node
type NodeOutput struct {
	Data map[string]interface{}
	Error error
}

// NodeMetadata contains metadata about a node type
type NodeMetadata struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Category    string                 `json:"category"`
	Description string                 `json:"description"`
	Inputs      map[string]interface{} `json:"inputs"`
	Outputs     map[string]interface{} `json:"outputs"`
	Icon        string                 `json:"icon"`
}

// NodeInstance is the interface that all nodes must implement
type NodeInstance interface {
	// Initialize sets up the node with configuration
	Initialize(config map[string]interface{}) error
	
	// Execute runs the node logic with the provided input
	Execute(ctx context.Context, input NodeInput) NodeOutput
	
	// Validate checks if the node configuration is valid
	Validate() error
	
	// Close performs cleanup operations
	Close() error
	
	// GetMetadata returns node metadata for UI
	GetMetadata() NodeMetadata
}

// NodeTypeRegistry holds all available node types
type NodeTypeRegistry interface {
	RegisterNodeType(id string, creator func() NodeInstance) error
	GetNodeType(id string) (func() NodeInstance, bool)
	ListNodeTypes() []NodeMetadata
}