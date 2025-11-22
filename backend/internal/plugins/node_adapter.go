package plugins

import (
	"context"
	"encoding/json"

	"github.com/citadel-agent/backend/internal/interfaces"
)

// NodeInstanceAdapter adapts a NodeInstance to work with the plugin system
type NodeInstanceAdapter struct {
	instance interfaces.NodeInstance
	metadata NodeMetadata
}

// NewNodeInstanceAdapter creates a new adapter from a NodeInstance
func NewNodeInstanceAdapter(instance interfaces.NodeInstance, metadata NodeMetadata) *NodeInstanceAdapter {
	return &NodeInstanceAdapter{
		instance: instance,
		metadata: metadata,
	}
}

// Execute implements NodePlugin interface
func (a *NodeInstanceAdapter) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	return a.instance.Execute(ctx, inputs)
}

// GetConfigSchema implements NodePlugin interface
func (a *NodeInstanceAdapter) GetConfigSchema() map[string]interface{} {
	// Default schema - in practice this would come from the node or be defined separately
	schema := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
		"required":   []string{},
	}
	
	if a.metadata.Schema != nil {
		// If metadata contains schema, use it
		schemaBytes, err := json.Marshal(a.metadata.Schema)
		if err == nil {
			json.Unmarshal(schemaBytes, &schema)
		}
	}
	
	return schema
}

// GetMetadata implements NodePlugin interface
func (a *NodeInstanceAdapter) GetMetadata() NodeMetadata {
	return a.metadata
}

// NodeInstancePluginWrapper wraps a NodeInstance to make it compatible with the plugin interface
type NodeInstancePluginWrapper struct {
	NodeName string
	Constructor func(config map[string]interface{}) (interfaces.NodeInstance, error)
	Metadata NodeMetadata
}

// CreatePluginAdapter creates a NodeInstanceAdapter for a specific node type
func (n *NodeInstancePluginWrapper) CreateAdapter(config map[string]interface{}) (*NodeInstanceAdapter, error) {
	instance, err := n.Constructor(config)
	if err != nil {
		return nil, err
	}

	return NewNodeInstanceAdapter(instance, n.Metadata), nil
}