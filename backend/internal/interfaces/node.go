package interfaces

import (
	"context"
)

// NodeInstance interface defines the contract for all workflow nodes
// This interface is crucial for breaking circular dependencies between packages
type NodeInstance interface {
	// Execute executes the node with given inputs and returns the result
	Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error)

	// GetType returns the type of the node
	GetType() string

	// GetID returns the unique identifier for this node instance
	GetID() string
}

// NodeDefinition represents the static definition of a node type
type NodeDefinition struct {
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Icon        string                 `json:"icon"`
	Category    string                 `json:"category"`
	Config      map[string]interface{} `json:"config"`
	InputSchema map[string]interface{} `json:"input_schema"`
	OutputSchema map[string]interface{} `json:"output_schema"`
}

// NodeFactory creates instances of NodeInstance based on type
type NodeFactory interface {
	CreateInstance(nodeType string, config map[string]interface{}) (NodeInstance, error)
	RegisterNodeType(nodeType string, constructor func(map[string]interface{}) (NodeInstance, error)) error
	GetNodeDefinition(nodeType string) (*NodeDefinition, bool)
	ListNodeTypes() []string
}

// NodeNotFoundError indicates that a node type is not registered
type NodeNotFoundError struct {
	NodeType string
}

func (e *NodeNotFoundError) Error() string {
	return "node type not found: " + e.NodeType
}