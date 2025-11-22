package temporal

import (
	"context"
	"fmt"

	"github.com/citadel-agent/backend/internal/interfaces"
	"github.com/citadel-agent/backend/internal/plugins"
)

// PluginNode represents a node that runs as a Temporal activity
type PluginNode struct {
	pluginManager *plugins.NodeManager
	nodeType      string
	config        map[string]interface{}
}

// NewPluginNode creates a new plugin node
func NewPluginNode(pluginManager *plugins.NodeManager, nodeType string, config map[string]interface{}) *PluginNode {
	return &PluginNode{
		pluginManager: pluginManager,
		nodeType:      nodeType,
		config:        config,
	}
}

func (pn *PluginNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Get the plugin node from the manager
	pluginNode, err := pn.pluginManager.GetNodePlugin(pn.nodeType)
	if err != nil {
		return nil, fmt.Errorf("plugin node %s not found: %w", pn.nodeType, err)
	}

	// Combine config and inputs for the plugin
	combinedInputs := make(map[string]interface{})
	
	// Add configuration
	for k, v := range pn.config {
		combinedInputs[k] = v
	}
	
	// Override with runtime inputs
	for k, v := range inputs {
		combinedInputs[k] = v
	}

	// Execute the plugin node
	return pluginNode.Execute(ctx, combinedInputs)
}

// CreatePluginNodeActivity creates a plugin node activity
func CreatePluginNodeActivity(pluginManager *plugins.NodeManager, nodeType string, config map[string]interface{}) interfaces.NodeInstance {
	return NewPluginNode(pluginManager, nodeType, config)
}

// PluginNodeAdapterFactory creates node instances from plugin types
type PluginNodeAdapterFactory struct {
	pluginManager *plugins.NodeManager
}

// NewPluginNodeAdapterFactory creates a new plugin node adapter factory
func NewPluginNodeAdapterFactory(pluginManager *plugins.NodeManager) *PluginNodeAdapterFactory {
	return &PluginNodeAdapterFactory{
		pluginManager: pluginManager,
	}
}

// CreateNode creates a node instance, trying plugin first then falling back to built-in
func (f *PluginNodeAdapterFactory) CreateNode(nodeType string, config map[string]interface{}) (interfaces.NodeInstance, error) {
	// First, try to create as a plugin
	pluginNode, err := f.pluginManager.GetNodePlugin(nodeType)
	if err == nil {
		// Plugin exists, create a wrapper
		return NewPluginNode(f.pluginManager, nodeType, config), nil
	}

	// If plugin doesn't exist, fall back to built-in node creation
	// This matches the behavior in the activities.go file
	return createNodeInstance(nodeType, config)
}