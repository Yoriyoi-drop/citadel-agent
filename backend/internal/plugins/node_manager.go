package plugins

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/hashicorp/go-plugin"
)

// NodeManager manages node plugins
type NodeManager struct {
	plugins     map[string]NodePlugin
	pluginClients map[string]*plugin.Client
	pluginPaths map[string]string
	mutex       sync.RWMutex
}

// NewNodeManager creates a new node plugin manager
func NewNodeManager() *NodeManager {
	return &NodeManager{
		plugins:       make(map[string]NodePlugin),
		pluginClients: make(map[string]*plugin.Client),
		pluginPaths:   make(map[string]string),
	}
}

// RegisterPluginAtPath registers and loads a plugin from file path
func (nm *NodeManager) RegisterPluginAtPath(pluginID, pluginPath string) error {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	// Kill any existing client for this plugin
	if client, exists := nm.pluginClients[pluginID]; exists {
		client.Kill()
		delete(nm.pluginClients, pluginID)
	}

	// Configure the plugin client
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: Handshake,
		Plugins: map[string]plugin.Plugin{
			"node": &NodePluginImpl{},
		},
		Cmd:              plugin.ComputeCmd(filepath.Join(pluginPath)),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return fmt.Errorf("failed to connect to plugin: %w", err)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("node")
	if err != nil {
		client.Kill()
		return fmt.Errorf("failed to dispense plugin: %w", err)
	}

	pluginImpl := raw.(NodePlugin)

	// Store the client and plugin
	nm.pluginClients[pluginID] = client
	nm.plugins[pluginID] = pluginImpl
	nm.pluginPaths[pluginID] = pluginPath

	return nil
}

// GetNodePlugin retrieves a node plugin by ID
func (nm *NodeManager) GetNodePlugin(pluginID string) (NodePlugin, error) {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()

	plugin, exists := nm.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("node plugin with ID %s not found", pluginID)
	}

	return plugin, nil
}

// ExecuteNode executes a node plugin with given inputs
func (nm *NodeManager) ExecuteNode(ctx context.Context, pluginID string, inputs map[string]interface{}) (map[string]interface{}, error) {
	plugin, err := nm.GetNodePlugin(pluginID)
	if err != nil {
		return nil, err
	}

	return plugin.Execute(ctx, inputs)
}

// GetNodeMetadata retrieves metadata for a node plugin
func (nm *NodeManager) GetNodeMetadata(pluginID string) (NodeMetadata, error) {
	plugin, err := nm.GetNodePlugin(pluginID)
	if err != nil {
		return NodeMetadata{}, err
	}

	return plugin.GetMetadata(), nil
}

// GetNodeConfigSchema retrieves configuration schema for a node plugin
func (nm *NodeManager) GetNodeConfigSchema(pluginID string) (map[string]interface{}, error) {
	plugin, err := nm.GetNodePlugin(pluginID)
	if err != nil {
		return nil, err
	}

	return plugin.GetConfigSchema()
}

// ListAvailablePlugins lists all registered plugins
func (nm *NodeManager) ListAvailablePlugins() []string {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()

	ids := make([]string, 0, len(nm.plugins))
	for id := range nm.plugins {
		ids = append(ids, id)
	}
	return ids
}

// UnregisterPlugin removes and cleans up a plugin
func (nm *NodeManager) UnregisterPlugin(pluginID string) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	if client, exists := nm.pluginClients[pluginID]; exists {
		client.Kill()
		delete(nm.pluginClients, pluginID)
	}

	delete(nm.plugins, pluginID)
	delete(nm.pluginPaths, pluginID)
}

// CloseAll closes all plugin connections
func (nm *NodeManager) CloseAll() {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	for id, client := range nm.pluginClients {
		client.Kill()
		delete(nm.pluginClients, id)
	}

	nm.plugins = make(map[string]NodePlugin)
	nm.pluginPaths = make(map[string]string)
}