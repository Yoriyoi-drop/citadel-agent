// config/config.go
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/redis/go-redis/v9"
)

// Config holds the application configuration
type Config struct {
	// Server configuration
	Server ServerConfig `mapstructure:"server" json:"server"`

	// Database configuration
	Database DatabaseConfig `mapstructure:"database" json:"database"`

	// Redis configuration
	Redis RedisConfig `mapstructure:"redis" json:"redis"`

	// Security configuration
	Security SecurityConfig `mapstructure:"security" json:"security"`

	// Logging configuration
	Logging LoggingConfig `mapstructure:"logging" json:"logging"`

	// Workflow engine configuration
	Workflow WorkflowConfig `mapstructure:"workflow" json:"workflow"`

	// Plugin configuration
	Plugin PluginConfig `mapstructure:"plugin" json:"plugin"`

	// API configuration
	API APIConfig `mapstructure:"api" json:"api"`

	// Scheduler configuration
	Scheduler SchedulerConfig `mapstructure:"scheduler" json:"scheduler"`

	// Worker configuration
	Worker WorkerConfig `mapstructure:"worker" json:"worker"`

	// AI service configuration
	AI AIConfig `mapstructure:"ai" json:"ai"`

	// Sandbox configuration
	Sandbox SandboxConfig `mapstructure:"sandbox" json:"sandbox"`

	// Monitoring configuration
	Monitoring MonitoringConfig `mapstructure:"monitoring" json:"monitoring"`

	// Environment
	Environment string `mapstructure:"environment" json:"environment"`

	// Additional custom configuration
	Custom map[string]interface{} `mapstructure:",remain" json:"custom"`
}

// ServerConfig contains server-specific configuration
type ServerConfig struct {
	Host         string        `mapstructure:"host" json:"host"`
	Port         int           `mapstructure:"port" json:"port"`
	ShutdownTime time.Duration `mapstructure:"shutdown_time" json:"shutdown_time"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" json:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" json:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout" json:"idle_timeout"`
	CORS         CORSConfig    `mapstructure:"cors" json:"cors"`
	SSL          SSLConfig     `mapstructure:"ssl" json:"ssl"`
}

// CORSConfig defines CORS settings
type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins" json:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods" json:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers" json:"allow_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials" json:"allow_credentials"`
	ExposeHeaders    []string `mapstructure:"expose_headers" json:"expose_headers"`
	MaxAge           int      `mapstructure:"max_age" json:"max_age"`
}

// SSLConfig defines SSL/TLS settings
type SSLConfig struct {
	Enabled  bool   `mapstructure:"enabled" json:"enabled"`
	CertFile string `mapstructure:"cert_file" json:"cert_file"`
	KeyFile  string `mapstructure:"key_file" json:"key_file"`
}

// DatabaseConfig contains database-specific configuration
type DatabaseConfig struct {
	Host       string        `mapstructure:"host" json:"host"`
	Port       int           `mapstructure:"port" json:"port"`
	Name       string        `mapstructure:"name" json:"name"`
	User       string        `mapstructure:"user" json:"user"`
	Password   string        `mapstructure:"password" json:"-"` // Never expose password in JSON
	SSLMode    string        `mapstructure:"ssl_mode" json:"ssl_mode"`
	MaxOpenConns int         `mapstructure:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns int         `mapstructure:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" json:"conn_max_lifetime"`
	QueryTimeout    time.Duration `mapstructure:"query_timeout" json:"query_timeout"`
	MigrationPath   string        `mapstructure:"migration_path" json:"migration_path"`
}

// RedisConfig contains Redis configuration
type RedisConfig struct {
	Address  string        `mapstructure:"address" json:"address"`
	Password string        `mapstructure:"password" json:"-"` // Never expose password
	DB       int           `mapstructure:"db" json:"db"`
	PoolSize int           `mapstructure:"pool_size" json:"pool_size"`
	MinIdleConns int       `mapstructure:"min_idle_conns" json:"min_idle_conns"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout" json:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" json:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" json:"write_timeout"`
	PoolTimeout  time.Duration `mapstructure:"pool_timeout" json:"pool_timeout"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time" json:"conn_max_idle_time"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" json:"conn_max_lifetime"`
}

// SecurityConfig holds security-related settings
type SecurityConfig struct {
	JWT JWTConfig `mapstructure:"jwt" json:"jwt"`
	SessionTimeout time.Duration `mapstructure:"session_timeout" json:"session_timeout"`
	AllowedOrigins []string      `mapstructure:"allowed_origins" json:"allowed_origins"`
	BlockedIPs     []string      `mapstructure:"blocked_ips" json:"blocked_ips"`
	TrustedProxies []string      `mapstructure:"trusted_proxies" json:"trusted_proxies"`
}

// JWTConfig defines JWT-related settings
type JWTConfig struct {
	Secret          string        `mapstructure:"secret" json:"-"`
	ExpirationTime  time.Duration `mapstructure:"expiration_time" json:"expiration_time"`
	RefreshTime     time.Duration `mapstructure:"refresh_time" json:"refresh_time"`
	Algorithm       string        `mapstructure:"algorithm" json:"algorithm"`
	CookieName      string        `mapstructure:"cookie_name" json:"cookie_name"`
	CookieSecure    bool          `mapstructure:"cookie_secure" json:"cookie_secure"`
	CookieHTTPOnly  bool          `mapstructure:"cookie_http_only" json:"cookie_http_only"`
	CookieSameSite  string        `mapstructure:"cookie_same_site" json:"cookie_same_site"`
	BlacklistEnabled bool         `mapstructure:"blacklist_enabled" json:"blacklist_enabled"`
}

// LoggingConfig defines logging settings
type LoggingConfig struct {
	Level      string `mapstructure:"level" json:"level"`
	Format     string `mapstructure:"format" json:"format"` // json, text
	Output     string `mapstructure:"output" json:"output"` // stdout, stderr, file
	File       string `mapstructure:"file" json:"file"`
	MaxSize    int    `mapstructure:"max_size" json:"max_size"`    // megabytes
	MaxBackups int    `mapstructure:"max_backups" json:"max_backups"` // number of files
	MaxAge     int    `mapstructure:"max_age" json:"max_age"`      // days
	Compress   bool   `mapstructure:"compress" json:"compress"`    // compress rotated files
	Verbose    bool   `mapstructure:"verbose" json:"verbose"`      // enable verbose logging
}

// WorkflowConfig defines workflow engine settings
type WorkflowConfig struct {
	MaxExecutionTime      time.Duration      `mapstructure:"max_execution_time" json:"max_execution_time"`
	MaxConcurrentExecutions int              `mapstructure:"max_concurrent_executions" json:"max_concurrent_executions"`
	MaxNodesPerWorkflow   int               `mapstructure:"max_nodes_per_workflow" json:"max_nodes_per_workflow"`
	MaxDepth              int               `mapstructure:"max_depth" json:"max_depth"`
	RetryPolicy           RetryPolicy       `mapstructure:"retry_policy" json:"retry_policy"`
	TimeoutPolicy         TimeoutPolicy     `mapstructure:"timeout_policy" json:"timeout_policy"`
	Persistence           PersistenceConfig `mapstructure:"persistence" json:"persistence"`
	StateRetentionDays    int               `mapstructure:"state_retention_days" json:"state_retention_days"`
	ResultRetentionDays   int               `mapstructure:"result_retention_days" json:"result_retention_days"`
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxAttempts   int           `mapstructure:"max_attempts" json:"max_attempts"`
	BackoffFactor float64       `mapstructure:"backoff_factor" json:"backoff_factor"`
	MaxWaitTime   time.Duration `mapstructure:"max_wait_time" json:"max_wait_time"`
	Enable        bool          `mapstructure:"enable" json:"enable"`
	Conditions    []string      `mapstructure:"conditions" json:"conditions"` // conditions under which to retry
}

// TimeoutPolicy defines timeout behavior
type TimeoutPolicy struct {
	NodeTimeout     time.Duration `mapstructure:"node_timeout" json:"node_timeout"`
	WorkflowTimeout time.Duration `mapstructure:"workflow_timeout" json:"workflow_timeout"`
	HTTPTimeout     time.Duration `mapstructure:"http_timeout" json:"http_timeout"`
	DatabaseTimeout time.Duration `mapstructure:"database_timeout" json:"database_timeout"`
	ScriptTimeout   time.Duration `mapstructure:"script_timeout" json:"script_timeout"`
}

// PersistenceConfig defines persistence settings
type PersistenceConfig struct {
	Enabled          bool          `mapstructure:"enabled" json:"enabled"`
	SnapshotInterval time.Duration `mapstructure:"snapshot_interval" json:"snapshot_interval"`
	Compression      bool          `mapstructure:"compression" json:"compression"`
	Encryption       bool          `mapstructure:"encryption" json:"encryption"`
	StorageType      string        `mapstructure:"storage_type" json:"storage_type"` // database, file, cloud
	StoragePath      string        `mapstructure:"storage_path" json:"storage_path"`
}

// PluginConfig defines plugin system settings
type PluginConfig struct {
	Directory           string        `mapstructure:"directory" json:"directory"`
	AllowedDirectories  []string      `mapstructure:"allowed_directories" json:"allowed_directories"`
	ScanInterval        time.Duration `mapstructure:"scan_interval" json:"scan_interval"`
	RequireSignature    bool          `mapstructure:"require_signature" json:"require_signature"`
	MaxFileSize         int64         `mapstructure:"max_file_size" json:"max_file_size"` // bytes
	MaxConcurrentLoads  int           `mapstructure:"max_concurrent_loads" json:"max_concurrent_loads"`
	Whitelist           []string      `mapstructure:"whitelist" json:"whitelist"`
	Blacklist           []string      `mapstructure:"blacklist" json:"blacklist"`
	AutoUpdate          bool          `mapstructure:"auto_update" json:"auto_update"`
	RepositoryURLs      []string      `mapstructure:"repository_urls" json:"repository_urls"`
	DownloadTimeout     time.Duration `mapstructure:"download_timeout" json:"download_timeout"`
	ValidationTimeout   time.Duration `mapstructure:"validation_timeout" json:"validation_timeout"`
	SecurityScanEnabled bool          `mapstructure:"security_scan_enabled" json:"security_scan_enabled"`
}

// APIConfig defines API-related settings
type APIConfig struct {
	Version               string        `mapstructure:"version" json:"version"`
	EnableCORS            bool          `mapstructure:"enable_cors" json:"enable_cors"`
	MaxRequestBodySize    int64         `mapstructure:"max_request_size" json:"max_request_size"` // bytes
	EnableRateLimiting    bool          `mapstructure:"enable_rate_limiting" json:"enable_rate_limiting"`
	RateLimitWindow       time.Duration `mapstructure:"rate_limit_window" json:"rate_limit_window"`
	RateLimitMaxRequests  int           `mapstructure:"rate_limit_max_requests" json:"rate_limit_max_requests"`
	EnableAPIDocs         bool          `mapstructure:"enable_api_docs" json:"enable_api_docs"`
	APIDocsPath           string        `mapstructure:"api_docs_path" json:"api_docs_path"`
	RequestTimeout        time.Duration `mapstructure:"request_timeout" json:"request_timeout"`
	MaxConcurrentRequests int           `mapstructure:"max_concurrent_requests" json:"max_concurrent_requests"`
	EnableDebugHeaders    bool          `mapstructure:"enable_debug_headers" json:"enable_debug_headers"`
	TrustedProxies        []string      `mapstructure:"trusted_proxies" json:"trusted_proxies"`
}

// SchedulerConfig defines scheduler settings
type SchedulerConfig struct {
	Enabled              bool          `mapstructure:"enabled" json:"enabled"`
	PollInterval         time.Duration `mapstructure:"poll_interval" json:"poll_interval"`
	MaxConcurrentTasks   int           `mapstructure:"max_concurrent_tasks" json:"max_concurrent_tasks"`
	QueueType            string        `mapstructure:"queue_type" json:"queue_type"` // redis, database, memory
	WorkerCount          int           `mapstructure:"worker_count" json:"worker_count"`
	MaxRetries           int           `mapstructure:"max_retries" json:"max_retries"`
	CronEnabled          bool          `mapstructure:"cron_enabled" json:"cron_enabled"`
	EventBasedEnabled    bool          `mapstructure:"event_based_enabled" json:"event_based_enabled"`
	BacklogCheckInterval time.Duration `mapstructure:"backlog_check_interval" json:"backlog_check_interval"`
	DeadLetterEnabled    bool          `mapstructure:"dead_letter_enabled" json:"dead_letter_enabled"`
	MaxBacklog           int           `mapstructure:"max_backlog" json:"max_backlog"`
}

// WorkerConfig defines worker settings
type WorkerConfig struct {
	Enabled              bool             `mapstructure:"enabled" json:"enabled"`
	PoolSize             int              `mapstructure:"pool_size" json:"pool_size"`
	MaxConcurrentTasks   int              `mapstructure:"max_concurrent_tasks" json:"max_concurrent_tasks"`
	TaskQueueType        string           `mapstructure:"task_queue_type" json:"task_queue_type"`
	HeartbeatInterval    time.Duration    `mapstructure:"heartbeat_interval" json:"heartbeat_interval"`
	GracefulShutdownTime time.Duration    `mapstructure:"graceful_shutdown_time" json:"graceful_shutdown_time"`
	ResourceLimits       ResourceLimits   `mapstructure:"resource_limits" json:"resource_limits"`
	ErrorHandling        ErrorHandlingConfig `mapstructure:"error_handling" json:"error_handling"`
	MaxTaskRetries       int              `mapstructure:"max_task_retries" json:"max_task_retries"`
	TaskTimeout          time.Duration    `mapstructure:"task_timeout" json:"task_timeout"`
}

// ResourceLimits defines resource constraints for workers
type ResourceLimits struct {
	MaxMemoryPerTask string `mapstructure:"max_memory_per_task" json:"max_memory_per_task"` // e.g., "128MB"
	MaxCPUPerTask    string `mapstructure:"max_cpu_per_task" json:"max_cpu_per_task"`       // e.g., "50%"
	TaskMemoryLimit  string `mapstructure:"task_memory_limit" json:"task_memory_limit"`
	TaskCPULimit     string `mapstructure:"task_cpu_limit" json:"task_cpu_limit"`
}

// ErrorHandlingConfig defines how errors should be handled
type ErrorHandlingConfig struct {
	MaxRetries         int      `mapstructure:"max_retries" json:"max_retries"`
	BackoffStrategy    string   `mapstructure:"backoff_strategy" json:"backoff_strategy"`
	NotifyOnError      bool     `mapstructure:"notify_on_error" json:"notify_on_error"`
	NotificationEmails []string `mapstructure:"notification_emails" json:"notification_emails"`
	LogFailedTasks     bool     `mapstructure:"log_failed_tasks" json:"log_failed_tasks"`
	IngestionStrategy  string   `mapstructure:"ingestion_strategy" json:"ingestion_strategy"` // continue, stop, retry
	AlertThreshold     int      `mapstructure:"alert_threshold" json:"alert_threshold"`       // number of failures before alerting
}

// AIConfig defines AI service settings
type AIConfig struct {
	Enabled              bool              `mapstructure:"enabled" json:"enabled"`
	MaxConcurrentAgents  int               `mapstructure:"max_concurrent_agents" json:"max_concurrent_agents"`
	DefaultModel         string            `mapstructure:"default_model" json:"default_model"`
	APIKeys              map[string]string `mapstructure:"api_keys" json:"-"` // Never expose API keys in JSON
	Timeout              time.Duration     `mapstructure:"timeout" json:"timeout"`
	MaxTokens            int               `mapstructure:"max_tokens" json:"max_tokens"`
	Temperature          float64           `mapstructure:"temperature" json:"temperature"`
	SystemPrompt         string            `mapstructure:"system_prompt" json:"system_prompt"`
	MemoryEnabled        bool              `mapstructure:"memory_enabled" json:"memory_enabled"`
	MemoryRetentionHours int               `mapstructure:"memory_retention_hours" json:"memory_retention_hours"`
	SafeMode             bool              `mapstructure:"safe_mode" json:"safe_mode"`
	MaxExecutionTime     time.Duration     `mapstructure:"max_execution_time" json:"max_execution_time"`
}

// SandboxConfig defines sandboxing settings
type SandboxConfig struct {
	Enabled             bool               `mapstructure:"enabled" json:"enabled"`
	Type                string             `mapstructure:"type" json:"type"` // container, process, namespace
	ContainerRuntime    string             `mapstructure:"container_runtime" json:"container_runtime"` // docker, podman, containerd
	MaxContainers       int                `mapstructure:"max_containers" json:"max_containers"`
	MaxProcessesPerTask int                `mapstructure:"max_processes_per_task" json:"max_processes_per_task"`
	NetworkIsolation    bool               `mapstructure:"network_isolation" json:"network_isolation"`
	FSEscapePrevention  bool               `mapstructure:"fs_escape_prevention" json:"fs_escape_prevention"`
	ReadOnlyRoot        bool               `mapstructure:"read_only_root" json:"read_only_root"`
	SeccompEnabled      bool               `mapstructure:"seccomp_enabled" json:"seccomp_enabled"`
	AppArmorEnabled     bool               `mapstructure:"app_armor_enabled" json:"app_armor_enabled"`
	WhitelistedSyscalls []string           `mapstructure:"whitelisted_syscalls" json:"whitelisted_syscalls"`
	BlacklistedSyscalls []string           `mapstructure:"blacklisted_syscalls" json:"blacklisted_syscalls"`
	RuntimePermissions  RuntimePermissions `mapstructure:"runtime_permissions" json:"runtime_permissions"`
}

// RuntimePermissions defines permissions for different runtimes
type RuntimePermissions struct {
	Network          bool `mapstructure:"network" json:"network"`
	FileSystem       bool `mapstructure:"file_system" json:"file_system"`
	ProcessControl   bool `mapstructure:"process_control" json:"process_control"`
	SystemCalls      bool `mapstructure:"system_calls" json:"system_calls"`
	EnvironmentVars  bool `mapstructure:"environment_vars" json:"environment_vars"`
	InterProcessComm bool `mapstructure:"inter_process_comm" json:"inter_process_comm"`
}

// MonitoringConfig defines monitoring settings
type MonitoringConfig struct {
	Enabled           bool             `mapstructure:"enabled" json:"enabled"`
	CollectMetrics    bool             `mapstructure:"collect_metrics" json:"collect_metrics"`
	ExportMetrics     bool             `mapstructure:"export_metrics" json:"export_metrics"`
	MetricsEndpoint   string           `mapstructure:"metrics_endpoint" json:"metrics_endpoint"`
	ProfilingEnabled  bool             `mapstructure:"profiling_enabled" json:"profiling_enabled"`
	ProfilingEndpoint string           `mapstructure:"profiling_endpoint" json:"profiling_endpoint"`
	LogLevel          string           `mapstructure:"log_level" json:"log_level"`
	AlertingEnabled   bool             `mapstructure:"alerting_enabled" json:"alerting_enabled"`
	AlertingEndpoint  string           `mapstructure:"alerting_endpoint" json:"alerting_endpoint"`
	TracingEnabled    bool             `mapstructure:"tracing_enabled" json:"tracing_enabled"`
	TracingEndpoint   string           `mapstructure:"tracing_endpoint" json:"tracing_endpoint"`
	RetentionPeriod   time.Duration    `mapstructure:"retention_period" json:"retention_period"`
	AlertRules        []AlertRule      `mapstructure:"alert_rules" json:"alert_rules"`
	AlertChannels     []AlertChannel   `mapstructure:"alert_channels" json:"alert_channels"`
}

// AlertRule defines a monitoring alert rule
type AlertRule struct {
	ID          string            `mapstructure:"id" json:"id"`
	Name        string            `mapstructure:"name" json:"name"`
	Expression  string            `mapstructure:"expression" json:"expression"`
	Description string            `mapstructure:"description" json:"description"`
	For         string            `mapstructure:"for" json:"for"` // duration string
	Labels      map[string]string `mapstructure:"labels" json:"labels"`
	Annotations map[string]string `mapstructure:"annotations" json:"annotations"`
}

// AlertChannel defines an alert delivery channel
type AlertChannel struct {
	Type    string             `mapstructure:"type" json:"type"` // webhook, email, slack, etc.
	Name    string             `mapstructure:"name" json:"name"`
	Config  map[string]interface{} `mapstructure:"config" json:"config"`
	Enabled bool               `mapstructure:"enabled" json:"enabled"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         "0.0.0.0",
			Port:         5001,
			ShutdownTime: 30 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 60 * time.Second,
			IdleTimeout:  120 * time.Second,
			CORS: CORSConfig{
				AllowOrigins:     []string{"*"},
				AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
				AllowCredentials: true,
				MaxAge:           3600,
			},
			SSL: SSLConfig{
				Enabled: false,
			},
		},
		Database: DatabaseConfig{
			Host:            getEnvOrDefault("DB_HOST", "localhost"),
			Port:            getEnvOrDefaultInt("DB_PORT", 5432),
			Name:            getEnvOrDefault("DB_NAME", "citadel_agent"),
			User:            getEnvOrDefault("DB_USER", "postgres"),
			Password:        getEnvOrDefault("DB_PASSWORD", "postgres"),
			SSLMode:         getEnvOrDefault("DB_SSL_MODE", "disable"),
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
			QueryTimeout:    30 * time.Second,
			MigrationPath:   "./migrations",
		},
		Redis: RedisConfig{
			Address:         "localhost:6379",
			Password:        "",
			DB:              0,
			PoolSize:        10,
			MinIdleConns:    5,
			DialTimeout:     5 * time.Second,
			ReadTimeout:     3 * time.Second,
			WriteTimeout:    3 * time.Second,
			PoolTimeout:     4 * time.Second,
			ConnMaxIdleTime: 30 * time.Second,
			ConnMaxLifetime: 0, // 0 means no limit
		},
		Security: SecurityConfig{
			SessionTimeout: 24 * time.Hour,
			AllowedOrigins: []string{"http://localhost:3000", "http://localhost:5001"},
			BlockedIPs:     []string{},
			TrustedProxies: []string{"127.0.0.1", "::1"},
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    50,      // 50MB
			MaxBackups: 3,       // 3 files
			MaxAge:     28,      // 28 days
			Compress:   true,    // compress rotated files
			Verbose:    false,   // verbose logging off
		},
		Workflow: WorkflowConfig{
			MaxExecutionTime:      1 * time.Hour,
			MaxConcurrentExecutions: 100,
			MaxNodesPerWorkflow:   100,
			MaxDepth:             20,
			RetryPolicy: RetryPolicy{
				MaxAttempts:   3,
				BackoffFactor: 2.0,
				MaxWaitTime:   30 * time.Second,
				Enable:        true,
				Conditions:    []string{"network_error", "timeout", "resource_unavailable"},
			},
			TimeoutPolicy: TimeoutPolicy{
				NodeTimeout:     10 * time.Minute,
				WorkflowTimeout: 1 * time.Hour,
				HTTPTimeout:     30 * time.Second,
				DatabaseTimeout: 10 * time.Second,
				ScriptTimeout:   5 * time.Minute,
			},
			Persistence: PersistenceConfig{
				Enabled:          true,
				SnapshotInterval: 5 * time.Minute,
				Compression:      true,
				Encryption:       false,
				StorageType:      "database",
				StoragePath:      "",
			},
			StateRetentionDays:   30,
			ResultRetentionDays:  7,
		},
		Plugin: PluginConfig{
			Directory:            "./plugins",
			AllowedDirectories:   []string{"./plugins", "/opt/citadel-agent/plugins"},
			ScanInterval:         5 * time.Minute,
			RequireSignature:     false,
			MaxFileSize:          50 * 1024 * 1024, // 50MB
			MaxConcurrentLoads:   5,
			Whitelist:            []string{},
			Blacklist:            []string{},
			AutoUpdate:           false,
			RepositoryURLs:       []string{"https://plugins.citadel-agent.com"},
			DownloadTimeout:      5 * time.Minute,
			ValidationTimeout:    30 * time.Second,
			SecurityScanEnabled:  true,
		},
		API: APIConfig{
			Version:              "v1",
			EnableCORS:           true,
			MaxRequestBodySize:   10 * 1024 * 1024, // 10MB
			EnableRateLimiting:   true,
			RateLimitWindow:      1 * time.Minute,
			RateLimitMaxRequests: 1000,
			EnableAPIDocs:        true,
			APIDocsPath:          "/docs",
			RequestTimeout:       60 * time.Second,
			MaxConcurrentRequests: 1000,
			EnableDebugHeaders:   false,
			TrustedProxies:       []string{"127.0.0.1", "::1"},
		},
		Scheduler: SchedulerConfig{
			Enabled:              true,
			PollInterval:         5 * time.Second,
			MaxConcurrentTasks:   20,
			QueueType:            "redis",
			WorkerCount:          5,
			MaxRetries:           3,
			CronEnabled:          true,
			EventBasedEnabled:    true,
			BacklogCheckInterval: 1 * time.Minute,
			DeadLetterEnabled:    true,
			MaxBacklog:           1000,
		},
		Worker: WorkerConfig{
			Enabled:              true,
			PoolSize:             10,
			MaxConcurrentTasks:   10,
			TaskQueueType:        "redis",
			HeartbeatInterval:    10 * time.Second,
			GracefulShutdownTime: 30 * time.Second,
			ResourceLimits: ResourceLimits{
				MaxMemoryPerTask: "256MB",
				MaxCPUPerTask:    "50%",
				TaskMemoryLimit:  "512MB",
				TaskCPULimit:     "100%",
			},
			ErrorHandling: ErrorHandlingConfig{
				MaxRetries:         3,
				BackoffStrategy:    "exponential",
				NotifyOnError:      true,
				NotificationEmails: []string{"admin@citadel-agent.com"},
				LogFailedTasks:     true,
				IngestionStrategy:  "continue",
				AlertThreshold:     5,
			},
			MaxTaskRetries: 3,
			TaskTimeout:    10 * time.Minute,
		},
		AI: AIConfig{
			Enabled:              true,
			MaxConcurrentAgents:  50,
			DefaultModel:         "gpt-3.5-turbo",
			Timeout:              2 * time.Minute,
			MaxTokens:            2048,
			Temperature:          0.7,
			SystemPrompt:         "You are a helpful AI assistant for the Citadel Agent workflow platform.",
			MemoryEnabled:        true,
			MemoryRetentionHours: 24,
			SafeMode:             true,
			MaxExecutionTime:     5 * time.Minute,
		},
		Sandbox: SandboxConfig{
			Enabled:              true,
			Type:                 "container",
			ContainerRuntime:     "docker",
			MaxContainers:        100,
			MaxProcessesPerTask:  20,
			NetworkIsolation:     true,
			FSEscapePrevention:   true,
			ReadOnlyRoot:         true,
			SeccompEnabled:       true,
			AppArmorEnabled:      false,
			WhitelistedSyscalls:  []string{"read", "write", "open", "close", "brk"},
			BlacklistedSyscalls:  []string{"execve", "fork", "clone"},
			RuntimePermissions: RuntimePermissions{
				Network:        false,
				FileSystem:     true,
				ProcessControl: false,
				SystemCalls:    false,
				EnvironmentVars: true,
				InterProcessComm: false,
			},
		},
		Monitoring: MonitoringConfig{
			Enabled:              true,
			CollectMetrics:       true,
			ExportMetrics:        true,
			MetricsEndpoint:      "/metrics",
			ProfilingEnabled:     false,
			ProfilingEndpoint:    "/debug/pprof",
			LogLevel:             "info",
			AlertingEnabled:      true,
			AlertingEndpoint:     "/alerts",
			TracingEnabled:       false,
			TracingEndpoint:      "/trace",
			RetentionPeriod:      30 * 24 * time.Hour,
		},
		Environment: "development",
		Custom:      make(map[string]interface{}),
	}
}

// LoadConfig loads configuration from various sources
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults from our default config
	defaultCfg := DefaultConfig()
	setViperDefaults(v, defaultCfg)

	// Set config file if specified
	if configPath != "" {
		dir := filepath.Dir(configPath)
		file := filepath.Base(configPath)
		
		v.SetConfigName(strings.TrimSuffix(file, filepath.Ext(file)))
		v.AddConfigPath(dir)
		v.AddConfigPath(".") // Also look in current directory
		v.AddConfigPath("/etc/citadel-agent/") // System config directory
		v.AddConfigPath("$HOME/.citadel-agent") // User config directory
	} else {
		// Set defaults for config file lookup
		v.SetConfigName("config")
		v.AddConfigPath(".")
		v.AddConfigPath("/etc/citadel-agent/")
		v.AddConfigPath("$HOME/.citadel-agent")
	}

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		// If config file not found, proceed with defaults and environment variables
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Read environment variables
	v.AutomaticEnv()
	v.SetEnvPrefix("CITADEL") // CITADEL_SERVER_PORT, CITADEL_DATABASE_HOST, etc.
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Unmarshal config
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	// Apply environment variable overrides
	applyEnvOverrides(&cfg, v)

	return &cfg, nil
}

// setViperDefaults sets default values in viper from a config struct
func setViperDefaults(v *viper.Viper, cfg interface{}) {
	val := reflect.ValueOf(cfg)
	typ := reflect.TypeOf(cfg)

	// For pointers, dereference them
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	setDefaultsRecursive(v, val, typ, "")
}

// setDefaultsRecursive recursively sets defaults for nested structs
func setDefaultsRecursive(v *viper.Viper, val reflect.Value, typ reflect.Type, prefix string) {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// Skip unexported fields
		if !fieldVal.CanInterface() {
			continue
		}

		tag := field.Tag.Get("mapstructure")
		if tag == "" {
			tag = strings.ToLower(field.Name)
		}

		fullKey := tag
		if prefix != "" {
			fullKey = prefix + "." + tag
		}

		// If it's a struct, recurse
		if fieldVal.Kind() == reflect.Struct {
			setDefaultsRecursive(v, fieldVal, fieldVal.Type(), fullKey)
		} else {
			// Set the default value
			v.SetDefault(fullKey, fieldVal.Interface())
		}
	}
}

// applyEnvOverrides applies environment variable overrides to config
func applyEnvOverrides(cfg *Config, v *viper.Viper) {
	// Server config overrides
	if host := v.GetString("server.host"); host != "" {
		cfg.Server.Host = host
	}
	if port := v.GetInt("server.port"); port != 0 {
		cfg.Server.Port = port
	}

	// Database config overrides
	if dbHost := v.GetString("database.host"); dbHost != "" {
		cfg.Database.Host = dbHost
	}
	if dbPort := v.GetInt("database.port"); dbPort != 0 {
		cfg.Database.Port = dbPort
	}
	if dbName := v.GetString("database.name"); dbName != "" {
		cfg.Database.Name = dbName
	}
	if dbUser := v.GetString("database.user"); dbUser != "" {
		cfg.Database.User = dbUser
	}
	if dbPassword := v.GetString("database.password"); dbPassword != "" {
		cfg.Database.Password = dbPassword
	}
	if sslMode := v.GetString("database.ssl_mode"); sslMode != "" {
		cfg.Database.SSLMode = sslMode
	}

	// Redis config overrides
	if redisAddr := v.GetString("redis.address"); redisAddr != "" {
		cfg.Redis.Address = redisAddr
	}
	if redisPassword := v.GetString("redis.password"); redisPassword != "" {
		cfg.Redis.Password = redisPassword
	}
	if redisDB := v.GetInt("redis.db"); redisDB != 0 {
		cfg.Redis.DB = redisDB
	}

	// Security config overrides
	if env := v.GetString("environment"); env != "" {
		cfg.Environment = env
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate server configuration
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535, got %d", c.Server.Port)
	}

	// Validate database configuration
	if c.Database.Host == "" {
		return fmt.Errorf("database host cannot be empty")
	}
	if c.Database.Port <= 0 || c.Database.Port > 65535 {
		return fmt.Errorf("database port must be between 1 and 65535, got %d", c.Database.Port)
	}

	// Validate workflow configuration
	if c.Workflow.MaxConcurrentExecutions <= 0 {
		return fmt.Errorf("max concurrent executions must be greater than 0")
	}
	if c.Workflow.MaxNodesPerWorkflow <= 0 {
		return fmt.Errorf("max nodes per workflow must be greater than 0")
	}

	// Validate plugin configuration
	if c.Plugin.Directory == "" {
		return fmt.Errorf("plugin directory cannot be empty")
	}

	return nil
}

// GetConnectionString returns the database connection string
func (c *DatabaseConfig) GetConnectionString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
		c.SSLMode,
	)
}

// GetRedisOptions returns Redis options
func (c *RedisConfig) GetRedisOptions() *redis.Options {
	return &redis.Options{
		Addr:         c.Address,
		Password:     c.Password,
		DB:           c.DB,
		PoolSize:     c.PoolSize,
		MinIdleConns: c.MinIdleConns,
		DialTimeout:  c.DialTimeout,
		ReadTimeout:  c.ReadTimeout,
		WriteTimeout: c.WriteTimeout,
		PoolTimeout:  c.PoolTimeout,
	}
}

// GetJWTExpirationTime returns the JWT expiration time
func (c *JWTConfig) GetJWTExpirationTime() time.Duration {
	if c.ExpirationTime == 0 {
		return 24 * time.Hour // default to 24 hours
	}
	return c.ExpirationTime
}

// IsProduction returns if the environment is production
func (c *Config) IsProduction() bool {
	return strings.ToLower(c.Environment) == "production"
}

// IsDevelopment returns if the environment is development
func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.Environment) == "development"
}

// IsTesting returns if the environment is testing
func (c *Config) IsTesting() bool {
	return strings.ToLower(c.Environment) == "testing"
}

// Helper functions to get environment variables with defaults
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvOrDefaultInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvOrDefaultBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}