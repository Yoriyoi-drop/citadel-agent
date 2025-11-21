// workflow/core/plugin_system.go
package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"plugin"
	"sync"

	"github.com/hashicorp/go-plugin"
	"github.com/tidwall/gjson"
)

// Plugin defines the interface for Citadel Agent plugins
type Plugin interface {
	// Initialize the plugin with configuration
	Initialize(config map[string]interface{}) error
	
	// GetName returns the plugin name
	GetName() string
	
	// GetVersion returns the plugin version
	GetVersion() string
	
	// GetDescription returns the plugin description
	GetDescription() string
	
	// GetType returns the plugin type (node, connector, processor, etc.)
	GetType() PluginType
	
	// Validate checks if the plugin configuration is valid
	Validate(config map[string]interface{}) error
	
	// Execute performs the plugin's main function
	Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)
}

// PluginType represents different types of plugins
type PluginType string

const (
	NodePlugin      PluginType = "node"
	ConnectorPlugin PluginType = "connector"
	ProcessorPlugin PluginType = "processor"
	TriggerPlugin   PluginType = "trigger"
	AIServicePlugin PluginType = "ai_service"
)

// PluginInfo contains metadata about a plugin
type PluginInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Type        PluginType        `json:"type"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Homepage    string            `json:"homepage"`
	License     string            `json:"license"`
	Tags        []string          `json:"tags"`
	Inputs      []PluginParameter `json:"inputs"`
	Outputs     []PluginParameter `json:"outputs"`
	Configuration []PluginParameter `json:"configuration"`
	Requirements PluginRequirements `json:"requirements"`
	Installed   bool              `json:"installed"`
	Enabled     bool              `json:"enabled"`
	CreatedAt   int64             `json:"created_at"`
	UpdatedAt   int64             `json:"updated_at"`
}

// PluginParameter describes a parameter for a plugin
type PluginParameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"` // string, number, boolean, object, array, file
	Required    bool        `json:"required"`
	Default     interface{} `json:"default"`
	Description string      `json:"description"`
	Validation  ParameterValidation `json:"validation"`
	UIHint      UIHint      `json:"ui_hint"` // UI presentation hint
}

// ParameterValidation describes validation rules for a parameter
type ParameterValidation struct {
	Min         *float64    `json:"min,omitempty"`
	Max         *float64    `json:"max,omitempty"`
	MinLength   *int        `json:"min_length,omitempty"`
	MaxLength   *int        `json:"max_length,omitempty"`
	Pattern     *string     `json:"pattern,omitempty"`
	Options     []string    `json:"options,omitempty"`
	AllowEmpty  bool        `json:"allow_empty"`
	UniqueItems bool        `json:"unique_items"`
}

// UIHint describes how to present a parameter in UI
type UIHint struct {
	Widget      string            `json:"widget"` // textbox, textarea, dropdown, checkbox, etc.
	Placeholder string            `json:"placeholder"`
	Group       string            `json:"group"` // UI grouping
	Order       int               `json:"order"` // UI ordering
	Visible     bool              `json:"visible"`
	Conditional *ConditionalUI    `json:"conditional,omitempty"`
}

// ConditionalUI defines when a parameter should be visible
type ConditionalUI struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // equals, not_equals, contains, etc.
	Value    interface{} `json:"value"`
}

// PluginRequirements describes system requirements for a plugin
type PluginRequirements struct {
	OS      []string `json:"os"`       // Supported operating systems
	Arch    []string `json:"arch"`     // Supported architectures
	GoVersion string `json:"go_version"` // Required Go version
	Deps    []string `json:"dependencies"` // Required system dependencies
	Memory  string   `json:"memory"`    // Required memory
	CPU     string   `json:"cpu"`       // Required CPU
}

// PluginManager manages the loading and execution of plugins
type PluginManager struct {
	plugins     map[string]*LoadedPlugin
	pluginDirs  []string
	mutex       sync.RWMutex
	logger      Logger
	config      *Config
	hub         *PluginHub
}

// LoadedPlugin represents a loaded plugin instance
type LoadedPlugin struct {
	PluginInfo
	Instance Plugin
	File     string
	Plugin   *plugin.Plugin
	Healthy  bool
	LastUsed int64
}

// PluginHub manages plugin discovery and marketplace
type PluginHub struct {
	Registry   *PluginRegistry
	Cache      map[string]*CachedPluginInfo
	LocalStore *LocalPluginStore
	Logger     Logger
	Mutex      sync.RWMutex
}

// PluginRegistry handles plugin discovery and metadata
type PluginRegistry struct {
	Plugins map[string]*PluginInfo
	Mutex   sync.RWMutex
	Logger  Logger
}

// LocalPluginStore manages local plugin storage
type LocalPluginStore struct {
	PluginDir string
	Mutex     sync.RWMutex
	Logger    Logger
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(config *Config, logger Logger) *PluginManager {
	pm := &PluginManager{
		plugins:    make(map[string]*LoadedPlugin),
		pluginDirs: config.PluginDirs,
		logger:     logger,
		config:     config,
		hub:        NewPluginHub(config, logger),
	}

	return pm
}

// NewPluginHub creates a new plugin hub
func NewPluginHub(config *Config, logger Logger) *PluginHub {
	return &PluginHub{
		Registry:   NewPluginRegistry(logger),
		Cache:      make(map[string]*CachedPluginInfo),
		LocalStore: NewLocalPluginStore(config.PluginDir, logger),
		Logger:     logger,
	}
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry(logger Logger) *PluginRegistry {
	return &PluginRegistry{
		Plugins: make(map[string]*PluginInfo),
		Logger:  logger,
	}
}

// NewLocalPluginStore creates a new local plugin store
func NewLocalPluginStore(pluginDir string, logger Logger) *LocalPluginStore {
	// Ensure plugin directory exists
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		if logger != nil {
			logger.Error("Failed to create plugin directory: %v", err)
		}
	}

	return &LocalPluginStore{
		PluginDir: pluginDir,
		Logger:    logger,
	}
}

// LoadPlugin loads a plugin from a file
func (pm *PluginManager) LoadPlugin(pluginPath string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Validate plugin path
	if !filepath.IsAbs(pluginPath) {
		return fmt.Errorf("plugin path must be absolute: %s", pluginPath)
	}

	// Load the plugin
	pluginFile, err := plugin.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to open plugin: %w", err)
	}

	// Extract plugin info from filename or metadata
	pluginID := filepath.Base(pluginPath)
	pluginID = pluginID[:len(pluginID)-len(filepath.Ext(pluginID))] // Remove extension

	// Attempt to get plugin interface
	pluginSymbol, err := pluginFile.Lookup("Plugin")
	if err != nil {
		return fmt.Errorf("plugin does not export symbol 'Plugin': %w", err)
	}

	pluginInstance, ok := pluginSymbol.(Plugin)
	if !ok {
		return fmt.Errorf("plugin does not implement Plugin interface")
	}

	// Initialize plugin with empty config to get metadata
	if err := pluginInstance.Initialize(make(map[string]interface{})); err != nil {
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	// Create loaded plugin
	loadedPlugin := &LoadedPlugin{
		PluginInfo: PluginInfo{
			ID:          pluginID,
			Name:        pluginInstance.GetName(),
			Version:     pluginInstance.GetVersion(),
			Type:        pluginInstance.GetType(),
			Description: pluginInstance.GetDescription(),
			Installed:   true,
			Enabled:     true,
			CreatedAt:   time.Now().Unix(),
			UpdatedAt:   time.Now().Unix(),
		},
		Instance: pluginInstance,
		File:     pluginPath,
		Plugin:   pluginFile,
		Healthy:  true,
		LastUsed: time.Now().Unix(),
	}

	pm.plugins[pluginID] = loadedPlugin

	if pm.logger != nil {
		pm.logger.Info("Loaded plugin: %s (v%s) - %s", loadedPlugin.Name, loadedPlugin.Version, loadedPlugin.Description)
	}

	return nil
}

// LoadAllPlugins loads all plugins from configured directories
func (pm *PluginManager) LoadAllPlugins() error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	for _, pluginDir := range pm.pluginDirs {
		files, err := ioutil.ReadDir(pluginDir)
		if err != nil {
			if pm.logger != nil {
				pm.logger.Warn("Failed to read plugin directory %s: %v", pluginDir, err)
			}
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			// Only load files with certain extensions
			ext := filepath.Ext(file.Name())
			if ext == ".so" || ext == ".plugin" || ext == ".dll" {
				pluginPath := filepath.Join(pluginDir, file.Name())
				
				if err := pm.loadPluginAtPath(pluginPath); err != nil {
					if pm.logger != nil {
						pm.logger.Error("Failed to load plugin %s: %v", pluginPath, err)
					}
				}
			}
		}
	}

	return nil
}

// loadPluginAtPath loads a single plugin at the specified path
func (pm *PluginManager) loadPluginAtPath(pluginPath string) error {
	// Validate plugin path
	if !filepath.IsAbs(pluginPath) {
		return fmt.Errorf("plugin path must be absolute: %s", pluginPath)
	}

	// Load the plugin
	pluginFile, err := plugin.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to open plugin: %w", err)
	}

	// Extract plugin info from filename or metadata
	pluginID := filepath.Base(pluginPath)
	pluginID = pluginID[:len(pluginID)-len(filepath.Ext(pluginID))] // Remove extension

	// Attempt to get plugin interface
	pluginSymbol, err := pluginFile.Lookup("Plugin")
	if err != nil {
		return fmt.Errorf("plugin does not export symbol 'Plugin': %w", err)
	}

	pluginInstance, ok := pluginSymbol.(Plugin)
	if !ok {
		return fmt.Errorf("plugin does not implement Plugin interface")
	}

	// Initialize plugin with empty config to get metadata
	if err := pluginInstance.Initialize(make(map[string]interface{})); err != nil {
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	// Create loaded plugin
	loadedPlugin := &LoadedPlugin{
		PluginInfo: PluginInfo{
			ID:          pluginID,
			Name:        pluginInstance.GetName(),
			Version:     pluginInstance.GetVersion(),
			Type:        pluginInstance.GetType(),
			Description: pluginInstance.GetDescription(),
			Installed:   true,
			Enabled:     true,
			CreatedAt:   time.Now().Unix(),
			UpdatedAt:   time.Now().Unix(),
		},
		Instance: pluginInstance,
		File:     pluginPath,
		Plugin:   pluginFile,
		Healthy:  true,
		LastUsed: time.Now().Unix(),
	}

	pm.plugins[pluginID] = loadedPlugin

	if pm.logger != nil {
		pm.logger.Info("Loaded plugin: %s (v%s) - %s", loadedPlugin.Name, loadedPlugin.Version, loadedPlugin.Description)
	}

	return nil
}

// ExecutePlugin executes a plugin with the given input
func (pm *PluginManager) ExecutePlugin(ctx context.Context, pluginID string, input map[string]interface{}) (map[string]interface{}, error) {
	pm.mutex.RLock()
	loadedPlugin, exists := pm.plugins[pluginID]
	pm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginID)
	}

	if !loadedPlugin.Enabled {
		return nil, fmt.Errorf("plugin %s is disabled", pluginID)
	}

	// Validate input against plugin's input schema
	if err := pm.validatePluginInput(loadedPlugin.Instance, input); err != nil {
		return nil, fmt.Errorf("plugin input validation failed: %w", err)
	}

	// Execute the plugin
	result, err := loadedPlugin.Instance.Execute(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("plugin execution failed: %w", err)
	}

	// Update last used timestamp
	pm.mutex.Lock()
	loadedPlugin.LastUsed = time.Now().Unix()
	pm.mutex.Unlock()

	if pm.logger != nil {
		pm.logger.Info("Executed plugin %s, input size: %d, output size: %d", pluginID, len(input), len(result))
	}

	return result, nil
}

// validatePluginInput validates input against plugin's requirements
func (pm *PluginManager) validatePluginInput(pluginInstance Plugin, input map[string]interface{}) error {
	// In a real implementation, this would validate the input against the plugin's schema
	// For now, we'll just do basic validation
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	return nil
}

// GetPluginInfo returns information about a specific plugin
func (pm *PluginManager) GetPluginInfo(pluginID string) (*PluginInfo, error) {
	pm.mutex.RLock()
	loadedPlugin, exists := pm.plugins[pluginID]
	pm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginID)
	}

	info := loadedPlugin.PluginInfo
	return &info, nil
}

// ListPlugins returns a list of all loaded plugins
func (pm *PluginManager) ListPlugins(pluginType *PluginType) []*PluginInfo {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	var plugins []*PluginInfo
	for _, loadedPlugin := range pm.plugins {
		if pluginType == nil || loadedPlugin.Type == *pluginType {
			info := loadedPlugin.PluginInfo
			plugins = append(plugins, &info)
		}
	}

	return plugins
}

// EnablePlugin enables a plugin
func (pm *PluginManager) EnablePlugin(pluginID string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	loadedPlugin, exists := pm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	loadedPlugin.Enabled = true
	loadedPlugin.UpdatedAt = time.Now().Unix()

	if pm.logger != nil {
		pm.logger.Info("Enabled plugin: %s", pluginID)
	}

	return nil
}

// DisablePlugin disables a plugin
func (pm *PluginManager) DisablePlugin(pluginID string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	loadedPlugin, exists := pm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	loadedPlugin.Enabled = false
	loadedPlugin.UpdatedAt = time.Now().Unix()

	if pm.logger != nil {
		pm.logger.Info("Disabled plugin: %s", pluginID)
	}

	return nil
}

// UpdatePlugin updates a plugin to a new version
func (pm *PluginManager) UpdatePlugin(pluginID, newVersionPath string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	loadedPlugin, exists := pm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	// Load the new version
	newPluginFile, err := plugin.Open(newVersionPath)
	if err != nil {
		return fmt.Errorf("failed to open new plugin version: %w", err)
	}

	// Verify the new plugin has the same interface
	newPluginSymbol, err := newPluginFile.Lookup("Plugin")
	if err != nil {
		return fmt.Errorf("new plugin does not export symbol 'Plugin': %w", err)
	}

	newPluginInstance, ok := newPluginSymbol.(Plugin)
	if !ok {
		return fmt.Errorf("new plugin does not implement Plugin interface")
	}

	// Validate that the new plugin has the same ID
	if newPluginInstance.GetName() != loadedPlugin.Name {
		return fmt.Errorf("new plugin has different name, expected %s, got %s", loadedPlugin.Name, newPluginInstance.GetName())
	}

	// Perform the update
	oldFile := loadedPlugin.File
	loadedPlugin.Plugin = newPluginFile
	loadedPlugin.Instance = newPluginInstance
	loadedPlugin.File = newVersionPath
	loadedPlugin.Version = newPluginInstance.GetVersion()
	loadedPlugin.UpdatedAt = time.Now().Unix()

	// Clean up old plugin (in a real implementation, we'd need to handle cleanup carefully)
	// For now, we'll just log the update

	if pm.logger != nil {
		pm.logger.Info("Updated plugin %s from %s to %s", pluginID, oldFile, newVersionPath)
	}

	return nil
}

// UninstallPlugin removes a plugin
func (pm *PluginManager) UninstallPlugin(pluginID string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	loadedPlugin, exists := pm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	// Close the plugin (if supported by the plugin framework)
	// In the case of Go plugins, they typically can't be unloaded cleanly
	// So we'll just remove from our internal registry

	delete(pm.plugins, pluginID)

	if pm.logger != nil {
		pm.logger.Info("Uninstalled plugin: %s", pluginID)
	}

	return nil
}

// ValidatePluginConfig validates plugin configuration
func (pm *PluginManager) ValidatePluginConfig(pluginID string, config map[string]interface{}) error {
	pm.mutex.RLock()
	loadedPlugin, exists := pm.plugins[pluginID]
	pm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	return loadedPlugin.Instance.Validate(config)
}

// RegisterPluginType registers a new plugin type with validation rules
func (ph *PluginHub) RegisterPluginType(pluginType PluginType, validator func(info *PluginInfo) error) {
	// In a real implementation, we would store validators for different plugin types
	// For now, we'll just acknowledge the registration
	if ph.Logger != nil {
		ph.Logger.Info("Registered plugin type: %s", pluginType)
	}
}

// SearchPlugins searches for plugins in the registry
func (ph *PluginHub) SearchPlugins(query string, filters map[string]string) ([]*PluginInfo, error) {
	ph.Mutex.RLock()
	defer ph.Mutex.RUnlock()

	var results []*PluginInfo
	
	// In a real implementation, this would query a remote registry
	// For now, we'll just return an empty list
	// In a real system, this would connect to a plugin marketplace
	
	return results, nil
}

// InstallPlugin installs a plugin from the hub
func (ph *PluginHub) InstallPlugin(pluginID string) error {
	ph.Mutex.Lock()
	defer ph.Mutex.Unlock()

	// In a real implementation, this would download the plugin from the hub
	// For now, we'll just return an error
	return fmt.Errorf("plugin hub functionality not implemented")
}

// GetPluginManifest gets the manifest for a plugin
func (ph *PluginHub) GetPluginManifest(pluginID string) (*PluginInfo, error) {
	ph.Mutex.RLock()
	defer ph.Mutex.RUnlock()

	// In a real implementation, this would fetch the manifest from the hub
	// For now, we'll return an error
	return nil, fmt.Errorf("plugin hub functionality not implemented")
}

// GetPluginUpdates checks for updates for installed plugins
func (ph *PluginHub) GetPluginUpdates() (map[string]string, error) {
	ph.Mutex.RLock()
	defer ph.Mutex.RUnlock()

	// In a real implementation, this would query the hub for updates
	// For now, we'll return an empty map
	return make(map[string]string), nil
}

// Integration with the workflow engine
func (e *Engine) initializePluginSystem() {
	// Initialize plugin manager
	e.pluginManager = NewPluginManager(e.config, e.logger)
	
	// Set up plugin directories
	pluginDir := filepath.Join(e.config.DataDir, "plugins")
	e.pluginManager.config.PluginDir = pluginDir
	
	// Create plugin directory if it doesn't exist
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		e.logger.Error("Failed to create plugin directory: %v", err)
	} else {
		e.logger.Info("Plugin directory created: %s", pluginDir)
	}
	
	// Add the plugin directory to the search path
	e.pluginManager.pluginDirs = append(e.pluginManager.pluginDirs, pluginDir)
	
	// Load all plugins
	if err := e.pluginManager.LoadAllPlugins(); err != nil {
		e.logger.Error("Failed to load plugins: %v", err)
	} else {
		plugins := e.pluginManager.ListPlugins(nil)
		e.logger.Info("Loaded %d plugins successfully", len(plugins))
	}
	
	// Register default node types that come from plugins
	e.registerPluginNodeTypes()
	
	if e.logger != nil {
		e.logger.Info("Plugin system initialized")
	}
}

// registerPluginNodeTypes registers node types from plugins
func (e *Engine) registerPluginNodeTypes() {
	plugins := e.pluginManager.ListPlugins(nil)
	
	for _, pluginInfo := range plugins {
		if pluginInfo.Type == NodePlugin && pluginInfo.Enabled {
			// In a real implementation, we would register the plugin as a node type
			// For now, we'll just log the registration
			if e.logger != nil {
				e.logger.Info("Registered plugin node: %s (type: %s)", pluginInfo.Name, pluginInfo.Type)
			}
		}
	}
}

// ExecutePluginNode executes a plugin-based node
func (e *Engine) ExecutePluginNode(ctx context.Context, execution *Execution, node *Node) error {
	// Extract plugin ID from node configuration
	pluginID, exists := node.Config["plugin_id"].(string)
	if !exists {
		return fmt.Errorf("plugin node configuration missing 'plugin_id'")
	}

	// Prepare input from node
	input := make(map[string]interface{})
	if node.Input != nil {
		input = node.Input
	}

	// Execute the plugin
	result, err := e.pluginManager.ExecutePlugin(ctx, pluginID, input)
	if err != nil {
		return fmt.Errorf("plugin execution failed: %w", err)
	}

	// Update node result
	node.Status = NodeSuccess
	node.Output = result
	node.CompletedAt = time.Now()

	return nil
}

// ValidatePluginNodeConfig validates a plugin node's configuration
func (e *Engine) ValidatePluginNodeConfig(node *Node) error {
	pluginID, exists := node.Config["plugin_id"].(string)
	if !exists {
		return fmt.Errorf("plugin node configuration missing 'plugin_id'")
	}

	config, exists := node.Config["plugin_config"].(map[string]interface{})
	if !exists {
		config = make(map[string]interface{}) // Use empty config if not provided
	}

	return e.pluginManager.ValidatePluginConfig(pluginID, config)
}