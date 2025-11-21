// backend/config/engine_config.go
package config

import (
	"time"
)

// EngineConfig holds the configuration for the workflow engine
type EngineConfig struct {
	// Core engine settings
	Parallelism             int           `json:"parallelism"`              // Number of concurrent executions
	MaxConcurrentExecutions int           `json:"max_concurrent_executions"` // Maximum concurrent executions
	ExecutionTimeout        time.Duration `json:"execution_timeout"`        // Timeout for workflow execution
	DefaultRetryAttempts    int           `json:"default_retry_attempts"`   // Default retry attempts for failed nodes

	// Security settings
	SecurityConfig *SecurityConfig `json:"security_config"`

	// Monitoring settings
	MonitoringConfig *MonitoringConfig `json:"monitoring_config"`

	// AI Agent settings
	AIConfig *AIConfig `json:"ai_config"`

	// Resource limits
	ResourceLimits *ResourceLimits `json:"resource_limits"`

	// Node settings
	NodeConfig *NodeConfig `json:"node_config"`

	// Database settings
	DatabaseConfig *DatabaseConfig `json:"database_config"`

	// Cache settings
	CacheConfig *CacheConfig `json:"cache_config"`

	// API settings
	APIConfig *APIConfig `json:"api_config"`

	// Notification settings
	NotificationConfig *NotificationConfig `json:"notification_config"`

	// Tenant settings
	TenantConfig *TenantConfig `json:"tenant_config"`

	// Sandbox settings
	SandboxConfig *SandboxConfig `json:"sandbox_config"`

	// Logging settings
	LoggingConfig *LoggingConfig `json:"logging_config"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	EnableRuntimeValidation bool     `json:"enable_runtime_validation"`
	AllowedHosts           []string `json:"allowed_hosts"`
	BlockedPaths           []string `json:"blocked_paths"`
	MaxExecutionTime       time.Duration `json:"max_execution_time"`
	MaxMemory              int64    `json:"max_memory"`
	EnablePermissionCheck  bool     `json:"enable_permission_check"`
	EnableResourceLimiting bool     `json:"enable_resource_limiting"`
	JWTSecret              string   `json:"jwt_secret"`
	EnableRateLimiting     bool     `json:"enable_rate_limiting"`
	RateLimitWindow        time.Duration `json:"rate_limit_window"`
	RateLimitRequests      int     `json:"rate_limit_requests"`
}

// MonitoringConfig holds monitoring-related configuration
type MonitoringConfig struct {
	EnableMetrics           bool     `json:"enable_metrics"`
	EnableAlerting          bool     `json:"enable_alerting"`
	EnableTracing           bool     `json:"enable_tracing"`
	PrometheusEndpoint      string   `json:"prometheus_endpoint"`
	AlertWebhookURL         string   `json:"alert_webhook_url"`
	MetricsCollectionInterval time.Duration `json:"metrics_collection_interval"`
	EnableHealthCheck       bool     `json:"enable_health_check"`
}

// AIConfig holds AI-related configuration
type AIConfig struct {
	EnableMemorySystem      bool              `json:"enable_memory_system"`
	EnableMultiAgentCoordination bool         `json:"enable_multi_agent_coordination"`
	EnableHumanInLoop       bool              `json:"enable_human_in_loop"`
	MaxMemorySize           int64             `json:"max_memory_size"`
	MemoryExpiry            time.Duration     `json:"memory_expiry"`
	OpenAIAPIKey            string            `json:"openai_api_key"`
	AnthropicAPIKey         string            `json:"anthropic_api_key"`
	EmbeddingModel          string            `json:"embedding_model"`
	EnableToolUse           bool              `json:"enable_tool_use"`
	MaxToolCalls            int               `json:"max_tool_calls"`
}

// ResourceLimits holds resource limitation configuration
type ResourceLimits struct {
	MaxWorkflowNodes       int           `json:"max_workflow_nodes"`
	MaxWorkflowSize        int64         `json:"max_workflow_size"`
	MaxExecutionTime       time.Duration `json:"max_execution_time"`
	MaxMemoryPerExecution  int64         `json:"max_memory_per_execution"`
	MaxNetworkRequests     int64         `json:"max_network_requests"`
	MaxFileOperations      int64         `json:"max_file_operations"`
}

// NodeConfig holds node-related configuration
type NodeConfig struct {
	EnableFileNodes         bool `json:"enable_file_nodes"`
	EnableHTTPNodes         bool `json:"enable_http_nodes"`
	EnableDatabaseNodes     bool `json:"enable_database_nodes"`
	EnableAINodes           bool `json:"enable_ai_nodes"`
	EnableLogicNodes        bool `json:"enable_logic_nodes"`
	EnableDataTransformNodes bool `json:"enable_data_transform_nodes"`
	EnableNotificationNodes bool `json:"enable_notification_nodes"`
	MaxNodesPerWorkflow     int  `json:"max_nodes_per_workflow"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
	SSLMode  string `json:"ssl_mode"`
	PoolSize int    `json:"pool_size"`
	Timeout  time.Duration `json:"timeout"`
}

// CacheConfig holds cache-related configuration
type CacheConfig struct {
	Host     string        `json:"host"`
	Port     int           `json:"port"`
	Password string        `json:"password"`
	DB       int           `json:"db"`
	PoolSize int           `json:"pool_size"`
	TTL      time.Duration `json:"ttl"`
}

// APIConfig holds API-related configuration
type APIConfig struct {
	Host             string        `json:"host"`
	Port             int           `json:"port"`
	ReadTimeout      time.Duration `json:"read_timeout"`
	WriteTimeout     time.Duration `json:"write_timeout"`
	IdleTimeout      time.Duration `json:"idle_timeout"`
	MaxRequestBodySize int64        `json:"max_request_body_size"`
	EnableCORS       bool          `json:"enable_cors"`
	CORSOrigins      []string      `json:"cors_origins"`
	EnableRateLimiting bool        `json:"enable_rate_limiting"`
	RateLimitWindow  time.Duration `json:"rate_limit_window"`
	RateLimitRequests int         `json:"rate_limit_requests"`
}

// NotificationConfig holds notification-related configuration
type NotificationConfig struct {
	EnableEmail    bool     `json:"enable_email"`
	EnableSlack    bool     `json:"enable_slack"`
	EnableWebhook  bool     `json:"enable_webhook"`
	EnablePush     bool     `json:"enable_push"`
	EnableSMS      bool     `json:"enable_sms"`
	EmailSettings  *EmailConfig `json:"email_settings"`
	SlackSettings  *SlackConfig `json:"slack_settings"`
	SMSSettings    *SMSConfig   `json:"sms_settings"`
	MaxRetries     int        `json:"max_retries"`
	RetryInterval  time.Duration `json:"retry_interval"`
}

// TenantConfig holds multi-tenant configuration
type TenantConfig struct {
	EnableMultiTenant     bool              `json:"enable_multi_tenant"`
	DefaultTenantLimit    int               `json:"default_tenant_limit"`
	IsolationLevel        string            `json:"isolation_level"` // "database", "schema", or "row"
	DefaultStorageLimit   int64             `json:"default_storage_limit"`
	EnableTenantQuotas    bool              `json:"enable_tenant_quotas"`
	TenantQuotas          *TenantQuotas     `json:"tenant_quotas"`
	EnableTenantBilling   bool              `json:"enable_tenant_billing"`
}

// TenantQuotas holds quota configuration per tenant
type TenantQuotas struct {
	MaxUsers           int `json:"max_users"`
	MaxWorkflows       int `json:"max_workflows"`
	MaxExecutions      int `json:"max_executions"`
	MaxStorage         int64 `json:"max_storage"`
	MaxAPIRequests     int64 `json:"max_api_requests"`
	MaxNotifications   int64 `json:"max_notifications"`
}

// SandboxConfig holds sandbox configuration
type SandboxConfig struct {
	EnableAdvancedSandboxing bool            `json:"enable_advanced_sandboxing"`
	EnableContainerSandbox   bool            `json:"enable_container_sandbox"`
	MaxExecutionTime         time.Duration   `json:"max_execution_time"`
	MaxMemory                int64           `json:"max_memory"`
	MaxCPU                   int             `json:"max_cpu"`
	AllowedCommands          []string        `json:"allowed_commands"`
	BlockedPaths             []string        `json:"blocked_paths"`
	EnableNetworkIsolation   bool            `json:"enable_network_isolation"`
	AllowedHosts             []string        `json:"allowed_hosts"`
	EnableFileAccessControl  bool            `json:"enable_file_access_control"`
	EnableProcessLimiting    bool            `json:"enable_process_limiting"`
	MaxProcesses             int             `json:"max_processes"`
	MaxOpenFiles             int             `json:"max_open_files"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `json:"level"`      // debug, info, warn, error
	Format     string `json:"format"`     // json, text
	Output     string `json:"output"`     // stdout, file, both
	Filepath   string `json:"filepath"`
	MaxSize    int    `json:"max_size"`   // in MB
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`    // in days
	Compress   bool   `json:"compress"`
}

// EmailConfig holds email notification settings
type EmailConfig struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	FromAddress  string `json:"from_address"`
	FromName     string `json:"from_name"`
	EnableTLS    bool   `json:"enable_tls"`
}

// SlackConfig holds Slack notification settings
type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
	Channel    string `json:"channel"`
	Username   string `json:"username"`
	IconURL    string `json:"icon_url"`
}

// SMSConfig holds SMS notification settings
type SMSConfig struct {
	Provider   string `json:"provider"`   // twilio, plivo, etc.
	AccountSID string `json:"account_sid"`
	AuthToken  string `json:"auth_token"`
	FromNumber string `json:"from_number"`
}

// DefaultEngineConfig returns the default configuration for the engine
func DefaultEngineConfig() *EngineConfig {
	return &EngineConfig{
		Parallelism:             10,
		MaxConcurrentExecutions: 100,
		ExecutionTimeout:        30 * time.Minute,
		DefaultRetryAttempts:    3,
		SecurityConfig: &SecurityConfig{
			EnableRuntimeValidation: true,
			AllowedHosts:           []string{"api.github.com", "api.openai.com", "httpbin.org"},
			BlockedPaths:           []string{"/etc/", "/proc/", "/sys/"},
			MaxExecutionTime:       10 * time.Minute,
			MaxMemory:              200 * 1024 * 1024, // 200MB
			EnablePermissionCheck:  true,
			EnableResourceLimiting: true,
			JWTSecret:              "your-super-secret-jwt-key-here-at-least-32-characters-for-production",
			EnableRateLimiting:     true,
			RateLimitWindow:        1 * time.Minute,
			RateLimitRequests:      1000,
		},
		MonitoringConfig: &MonitoringConfig{
			EnableMetrics:           true,
			EnableAlerting:          true,
			EnableTracing:           false,
			PrometheusEndpoint:      "/metrics",
			MetricsCollectionInterval: 30 * time.Second,
			EnableHealthCheck:       true,
		},
		AIConfig: &AIConfig{
			EnableMemorySystem:      true,
			EnableMultiAgentCoordination: true,
			EnableHumanInLoop:       true,
			MaxMemorySize:           100 * 1024 * 1024, // 100MB
			MemoryExpiry:            24 * time.Hour,
			EmbeddingModel:          "text-embedding-ada-002",
			EnableToolUse:           true,
			MaxToolCalls:            10,
		},
		ResourceLimits: &ResourceLimits{
			MaxWorkflowNodes:       100,
			MaxWorkflowSize:        10 * 1024 * 1024, // 10MB
			MaxExecutionTime:       30 * time.Minute,
			MaxMemoryPerExecution:  500 * 1024 * 1024, // 500MB
			MaxNetworkRequests:     100,
			MaxFileOperations:      50,
		},
		NodeConfig: &NodeConfig{
			EnableFileNodes:         true,
			EnableHTTPNodes:         true,
			EnableDatabaseNodes:     true,
			EnableAINodes:           true,
			EnableLogicNodes:        true,
			EnableDataTransformNodes: true,
			EnableNotificationNodes: true,
			MaxNodesPerWorkflow:     200,
		},
		APIConfig: &APIConfig{
			Host:                 "0.0.0.0",
			Port:                 5001,
			ReadTimeout:          30 * time.Second,
			WriteTimeout:         60 * time.Second,
			IdleTimeout:          60 * time.Second,
			MaxRequestBodySize:   10 * 1024 * 1024, // 10MB
			EnableCORS:           true,
			CORSOrigins:          []string{"http://localhost:3000", "http://localhost:3001", "https://yourdomain.com"},
			EnableRateLimiting:   true,
			RateLimitWindow:      1 * time.Minute,
			RateLimitRequests:    1000,
		},
		NotificationConfig: &NotificationConfig{
			EnableEmail:   true,
			EnableSlack:   true,
			EnableWebhook: true,
			EnablePush:    false,
			EnableSMS:     false,
			MaxRetries:    3,
			RetryInterval: 5 * time.Second,
		},
		TenantConfig: &TenantConfig{
			EnableMultiTenant:  true,
			DefaultTenantLimit: 50,
			IsolationLevel:     "row",
			DefaultStorageLimit: 10 * 1024 * 1024 * 1024, // 10GB
			EnableTenantQuotas: true,
			TenantQuotas: &TenantQuotas{
				MaxUsers:         100,
				MaxWorkflows:     500,
				MaxExecutions:    10000,
				MaxStorage:       10 * 1024 * 1024 * 1024, // 10GB
				MaxAPIRequests:   100000,
				MaxNotifications: 50000,
			},
			EnableTenantBilling: false,
		},
		SandboxConfig: &SandboxConfig{
			EnableAdvancedSandboxing: true,
			EnableContainerSandbox:   false, // Enable for high security needs
			MaxExecutionTime:         5 * time.Minute,
			MaxMemory:                100 * 1024 * 1024, // 100MB
			MaxCPU:                   80, // Percentage
			AllowedCommands:          []string{"ls", "cat", "echo", "grep", "awk", "sed"},
			BlockedPaths:             []string{"/etc/", "/proc/", "/sys/", "/root/", "/home/"},
			EnableNetworkIsolation:   true,
			AllowedHosts:             []string{"api.github.com", "api.openai.com", "httpbin.org"},
			EnableFileAccessControl:  true,
			EnableProcessLimiting:    true,
			MaxProcesses:             10,
			MaxOpenFiles:             20,
		},
		LoggingConfig: &LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100, // 100MB
			MaxBackups: 3,
			MaxAge:     28, // 28 days
			Compress:   true,
		},
	}
}