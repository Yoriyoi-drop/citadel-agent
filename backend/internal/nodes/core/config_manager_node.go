package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/citadel-agent/backend/internal/workflow/core/engine"
)

// ConfigManagerNodeConfig represents the configuration for a ConfigManager node
type ConfigManagerNodeConfig struct {
	Source       string                 `json:"source"`         // file, env, remote, defaults
	ConfigPath   string                 `json:"config_path,omitempty"` // path to config file
	RemoteURL    string                 `json:"remote_url,omitempty"` // URL for remote config
	Defaults     map[string]interface{} `json:"defaults,omitempty"` // default values
	WatchChanges bool                   `json:"watch_changes"`  // watch for config changes
	KeyMappings  map[string]string      `json:"key_mappings,omitempty"` // map input keys to config keys
}

// ConfigManagerNode handles configuration management using Viper
type ConfigManagerNode struct {
	config ConfigManagerNodeConfig
	viper  *viper.Viper
}

// NewConfigManagerNode creates a new ConfigManager node with the given configuration
func NewConfigManagerNode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Extract config values
	source := getStringValue(config["source"], "defaults")
	configPath := getStringValue(config["config_path"], "")
	remoteURL := getStringValue(config["remote_url"], "")

	watchChanges := false
	if wc, exists := config["watch_changes"]; exists {
		if wcBool, ok := wc.(bool); ok {
			watchChanges = wcBool
		}
	}

	// Extract defaults
	defaults := make(map[string]interface{})
	if defaultsVal, exists := config["defaults"]; exists {
		if defaultsMap, ok := defaultsVal.(map[string]interface{}); ok {
			defaults = defaultsMap
		}
	}

	// Extract key mappings
	keyMappings := make(map[string]string)
	if mappingsVal, exists := config["key_mappings"]; exists {
		if mappingsMap, ok := mappingsVal.(map[string]interface{}); ok {
			for k, v := range mappingsMap {
				if vStr, ok := v.(string); ok {
					keyMappings[k] = vStr
				}
			}
		}
	}

	configManagerConfig := ConfigManagerNodeConfig{
		Source:       source,
		ConfigPath:   configPath,
		RemoteURL:    remoteURL,
		Defaults:     defaults,
		WatchChanges: watchChanges,
		KeyMappings:  keyMappings,
	}

	// Initialize Viper
	v := viper.New()

	// Set up configuration based on source
	switch configManagerConfig.Source {
	case "file":
		if configManagerConfig.ConfigPath == "" {
			return nil, fmt.Errorf("config_path is required when source is file")
		}
		v.SetConfigFile(configManagerConfig.ConfigPath)
		err := v.ReadInConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %v", err)
		}
	case "env":
		v.AutomaticEnv()
		// Replace . with _ in env variable names
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	case "remote":
		// For remote configs, we'll just store the URL and fetch when Execute is called
		// The actual remote fetching would be done in Execute
	case "defaults":
		// Set default values
		for key, value := range configManagerConfig.Defaults {
			v.SetDefault(key, value)
		}
	default:
		// Default to using defaults
		for key, value := range configManagerConfig.Defaults {
			v.SetDefault(key, value)
		}
	}

	return &ConfigManagerNode{
		config: configManagerConfig,
		viper:  v,
	}, nil
}

// Execute implements the NodeInstance interface
func (c *ConfigManagerNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	var resultData map[string]interface{}

	switch c.config.Source {
	case "file":
		resultData = c.handleFileSource(input)
	case "env":
		resultData = c.handleEnvSource(input)
	case "remote":
		resultData = c.handleRemoteSource(input)
	case "defaults":
		resultData = c.handleDefaultsSource(input)
	default:
		resultData = c.handleDefaultsSource(input)
	}

	// Merge with input values if provided
	for k, v := range input {
		if _, exists := resultData[k]; !exists {
			resultData[k] = v
		}
	}

	resultData["success"] = true
	resultData["timestamp"] = time.Now().Unix()

	return resultData, nil
}

// handleFileSource processes configuration from file
func (c *ConfigManagerNode) handleFileSource(input map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Get all values from viper
	allSettings := c.viper.AllSettings()

	// Apply key mappings if provided
	if len(c.config.KeyMappings) > 0 {
		for configKey, configValue := range allSettings {
			if mappedKey, exists := c.config.KeyMappings[configKey]; exists {
				result[mappedKey] = configValue
			} else {
				result[configKey] = configValue
			}
		}
	} else {
		for configKey, configValue := range allSettings {
			result[configKey] = configValue
		}
	}

	// Override with input values if provided
	for k, v := range input {
		result[k] = v
	}

	return result
}

// handleEnvSource processes configuration from environment variables
func (c *ConfigManagerNode) handleEnvSource(input map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Get environment variables based on input keys or all if no specific keys requested
	if len(input) == 0 {
		// If no specific keys requested, get all config
		allSettings := c.viper.AllSettings()
		for k, v := range allSettings {
			result[k] = v
		}
	} else {
		// Get specific keys from environment
		for key := range input {
			if c.viper.IsSet(key) {
				result[key] = c.viper.Get(key)
			} else {
				// Use input value as fallback
				result[key] = input[key]
			}
		}
	}

	return result
}

// handleRemoteSource processes configuration from remote source
func (c *ConfigManagerNode) handleRemoteSource(input map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// In a real implementation, this would fetch config from a remote source
	// For now, we'll simulate the behavior
	if c.config.RemoteURL != "" {
		result["remote_url"] = c.config.RemoteURL
		result["fetch_status"] = "simulated"
		result["timestamp"] = time.Now().Unix()
	}

	// Add default values if any
	for k, v := range c.config.Defaults {
		if _, exists := result[k]; !exists {
			result[k] = v
		}
	}

	// Override with input values
	for k, v := range input {
		result[k] = v
	}

	return result
}

// handleDefaultsSource processes configuration from default values
func (c *ConfigManagerNode) handleDefaultsSource(input map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Use default values
	for k, v := range c.config.Defaults {
		result[k] = v
	}

	// Override with input values
	for k, v := range input {
		result[k] = v
	}

	return result
}

// getStringValue safely extracts a string value
func getStringValue(v interface{}, defaultValue string) string {
	if v == nil {
		return defaultValue
	}
	if s, ok := v.(string); ok {
		return s
	}
	return defaultValue
}