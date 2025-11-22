package temporal

import (
	"time"

	"go.temporal.io/sdk/client"
)

// Default configuration values
const (
	DefaultTemporalAddress  = "localhost:7233"
	DefaultNamespace        = "default"
	DefaultTaskQueue        = "citadel-agent-workflows"
	DefaultWorkflowTimeout  = 60 * time.Minute
	DefaultActivityTimeout  = 10 * time.Minute
	DefaultHeartbeatTimeout = 5 * time.Minute
)

// AdvancedConfig holds advanced configuration for Temporal integration
type AdvancedConfig struct {
	// Basic Temporal settings
	Address      string
	Namespace    string
	TaskQueue    string
	
	// Workflow settings
	WorkflowTimeout time.Duration
	RetryAttempts   int
	
	// Activity settings
	ActivityTimeout    time.Duration
	HeartbeatTimeout   time.Duration
	MaxConcurrentTasks int
	
	// Retry policy for workflows
	RetryPolicy *client.RetryPolicy
	
	// Connection settings
	ConnectionTimeout time.Duration
	RefreshInterval   time.Duration
	
	// Metrics and logging
	EnableMetrics bool
	EnableLogging bool
	
	// Security settings
	UseTLS     bool
	CertFile   string
	KeyFile    string
	CaFile     string
	ServerName string
}

// GetDefaultConfig returns a configuration with default values
func GetDefaultConfig() *AdvancedConfig {
	return &AdvancedConfig{
		Address:           DefaultTemporalAddress,
		Namespace:         DefaultNamespace,
		TaskQueue:         DefaultTaskQueue,
		WorkflowTimeout:   DefaultWorkflowTimeout,
		ActivityTimeout:   DefaultActivityTimeout,
		HeartbeatTimeout:  DefaultHeartbeatTimeout,
		RetryAttempts:     3,
		MaxConcurrentTasks: 100,
		ConnectionTimeout: 30 * time.Second,
		RefreshInterval:   10 * time.Second,
		EnableMetrics:     true,
		EnableLogging:     true,
		UseTLS:            false,
	}
}

// Validate validates the configuration
func (c *AdvancedConfig) Validate() error {
	if c.Address == "" {
		return &ConfigError{"Address cannot be empty"}
	}
	
	if c.Namespace == "" {
		return &ConfigError{"Namespace cannot be empty"}
	}
	
	if c.TaskQueue == "" {
		return &ConfigError{"TaskQueue cannot be empty"}
	}
	
	if c.WorkflowTimeout <= 0 {
		return &ConfigError{"WorkflowTimeout must be positive"}
	}
	
	if c.ActivityTimeout <= 0 {
		return &ConfigError{"ActivityTimeout must be positive"}
	}
	
	if c.HeartbeatTimeout <= 0 {
		return &ConfigError{"HeartbeatTimeout must be positive"}
	}
	
	return nil
}

// ConfigError represents an error in configuration
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}