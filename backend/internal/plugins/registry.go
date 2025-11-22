package plugins

import (
	"fmt"

	"github.com/citadel-agent/backend/internal/interfaces"
)

// PluginAwareNodeRegistry combines local nodes and plugin nodes
type PluginAwareNodeRegistry struct {
	localNodes   map[string]func(map[string]interface{}) (interfaces.NodeInstance, error)
	pluginManager *NodeManager
}

// NewPluginAwareNodeRegistry creates a new registry that can handle both local and plugin nodes
func NewPluginAwareNodeRegistry(pluginManager *NodeManager) *PluginAwareNodeRegistry {
	return &PluginAwareNodeRegistry{
		localNodes:    make(map[string]func(map[string]interface{}) (interfaces.NodeInstance, error)),
		pluginManager: pluginManager,
	}
}

// RegisterNodeType registers a local node type with its constructor
func (r *PluginAwareNodeRegistry) RegisterNodeType(nodeType string, constructor func(map[string]interface{}) (interfaces.NodeInstance, error)) {
	r.localNodes[nodeType] = constructor
}

// RegisterPluginNode registers a plugin node
func (r *PluginAwareNodeRegistry) RegisterPluginNode(pluginID string) error {
	// Verify that the plugin exists and is valid
	_, err := r.pluginManager.GetNodeMetadata(pluginID)
	if err != nil {
		return fmt.Errorf("plugin %s is not available: %w", pluginID, err)
	}
	
	// We don't need to store it separately since it's already in the plugin manager
	return nil
}

// CreateInstance creates a new instance of the specified node type
// It first tries to find a local node, then falls back to plugin nodes
func (r *PluginAwareNodeRegistry) CreateInstance(nodeType string, config map[string]interface{}) (interfaces.NodeInstance, error) {
	// First, try to find a local node
	if constructor, exists := r.localNodes[nodeType]; exists {
		return constructor(config)
	}
	
	// If not found locally, try to find as a plugin node
	plugin, err := r.pluginManager.GetNodePlugin(nodeType)
	if err != nil {
		return nil, fmt.Errorf("node type %s not registered (local or plugin)", nodeType)
	}
	
	// Create an adapter to make the plugin compatible with NodeInstance interface
	metadata := plugin.GetMetadata()
	adapter := NewNodeInstanceAdapter(nil, metadata) // This is a simplified approach
	
	// For plugin nodes, we'll need a different approach
	// Create a wrapper that implements NodeInstance using the plugin
	wrapper := &PluginNodeWrapper{
		plugin: plugin,
		config: config,
	}
	
	return wrapper, nil
}

// PluginNodeWrapper is a wrapper that makes a plugin compatible with NodeInstance interface
type PluginNodeWrapper struct {
	plugin NodePlugin
	config map[string]interface{}
}

// Execute implements the NodeInstance interface for plugin nodes
func (p *PluginNodeWrapper) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Combine config and inputs for the plugin
	combinedInputs := make(map[string]interface{})
	
	// Add configuration
	for k, v := range p.config {
		combinedInputs[k] = v
	}
	
	// Override with runtime inputs
	for k, v := range inputs {
		combinedInputs[k] = v
	}
	
	return p.plugin.Execute(ctx, combinedInputs)
}