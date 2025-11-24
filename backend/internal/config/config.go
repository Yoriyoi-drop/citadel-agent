package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	AppName     string `mapstructure:"app_name"`
	AppEnv      string `mapstructure:"app_env"`
	AppPort     string `mapstructure:"app_port"`
	AppDebug    bool   `mapstructure:"app_debug"`
	AppTimezone string `mapstructure:"app_timezone"`

	// Database
	DBHost     string `mapstructure:"db_host"`
	DBPort     int    `mapstructure:"db_port"`
	DBUser     string `mapstructure:"db_user"`
	DBPassword string `mapstructure:"db_password"`
	DBName     string `mapstructure:"db_name"`
	DBSSLMode  string `mapstructure:"db_ssl_mode"`

	// Redis
	RedisHost     string `mapstructure:"redis_host"`
	RedisPort     int    `mapstructure:"redis_port"`
	RedisPassword string `mapstructure:"redis_password"`
	RedisDB       int    `mapstructure:"redis_db"`

	// JWT
	JWTSecret           string        `mapstructure:"jwt_secret"`
	JWTExpiresIn        time.Duration `mapstructure:"jwt_expires_in"`
	JWTRefreshSecret    string        `mapstructure:"jwt_refresh_secret"`
	JWTRefreshExpiresIn time.Duration `mapstructure:"jwt_refresh_expires_in"`

	// Temporal
	TemporalAddress   string `mapstructure:"temporal_address"`
	TemporalNamespace string `mapstructure:"temporal_namespace"`

	// Security
	SecureCookies      bool   `mapstructure:"secure_cookies"`
	CORSAllowedOrigins string `mapstructure:"cors_allowed_origins"`
	RateLimitRequests  int    `mapstructure:"rate_limit_requests"`
	RateLimitWindow    int    `mapstructure:"rate_limit_window"`

	// AI Models
	AILlamaModelPath   string `mapstructure:"ai_llama_model_path"`
	AIMistralModelPath string `mapstructure:"ai_mistral_model_path"`
	AIClipModelPath    string `mapstructure:"ai_clip_model_path"`
	AIWhisperModelPath string `mapstructure:"ai_whisper_model_path"`

	// File Uploads
	MaxUploadSize    string `mapstructure:"max_upload_size"`
	AllowedFileTypes string `mapstructure:"allowed_file_types"`
	TempFileDir      string `mapstructure:"temp_file_dir"`

	// Monitoring
	PrometheusEnabled bool   `mapstructure:"prometheus_enabled"`
	GrafanaEnabled    bool   `mapstructure:"grafana_enabled"`
	LogLevel          string `mapstructure:"log_level"`
	LokiURL           string `mapstructure:"loki_url"`

	// Workflow Engine
	MaxConcurrentExecutions int           `mapstructure:"max_concurrent_executions"`
	MaxConcurrentNodes      int           `mapstructure:"max_concurrent_nodes"`
	DefaultWorkflowTimeout  time.Duration `mapstructure:"default_workflow_timeout"`
	MaxRetries              int           `mapstructure:"max_retries"`
	RetryDelay              time.Duration `mapstructure:"retry_delay"`
	EnableProfiling         bool          `mapstructure:"enable_profiling"`
	EnableCaching           bool          `mapstructure:"enable_caching"`
	CacheTTL                time.Duration `mapstructure:"cache_ttl"`
}

// LoadConfig loads the application configuration
func LoadConfig() (*Config, error) {
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("./config")

	// Set default values
	viper.SetDefault("app_name", "Citadel Agent")
	viper.SetDefault("app_env", "development")
	viper.SetDefault("app_port", "8080")
	viper.SetDefault("app_debug", true)
	viper.SetDefault("app_timezone", "UTC")

	viper.SetDefault("db_host", "localhost")
	viper.SetDefault("db_port", 5432)
	viper.SetDefault("db_user", "postgres")
	viper.SetDefault("db_password", "")
	viper.SetDefault("db_name", "citadel_agent")
	viper.SetDefault("db_ssl_mode", "disable")

	viper.SetDefault("redis_host", "localhost")
	viper.SetDefault("redis_port", 6379)
	viper.SetDefault("redis_password", "")
	viper.SetDefault("redis_db", 0)

	// Remove hardcoded secrets - require via environment
	viper.SetDefault("jwt_secret", "")
	viper.SetDefault("jwt_expires_in", "24h")
	viper.SetDefault("jwt_refresh_secret", "")
	viper.SetDefault("jwt_refresh_expires_in", "720h")

	viper.SetDefault("temporal_address", "localhost:7233")
	viper.SetDefault("temporal_namespace", "default")

	viper.SetDefault("secure_cookies", false)
	viper.SetDefault("cors_allowed_origins", "http://localhost:3000,http://localhost:8080")
	viper.SetDefault("rate_limit_requests", 100)
	viper.SetDefault("rate_limit_window", 60)

	viper.SetDefault("max_upload_size", "10MB") // Reduced from 100MB
	viper.SetDefault("allowed_file_types", "json,csv,txt,pdf,doc,docx,xlsx")
	viper.SetDefault("temp_file_dir", "/tmp/citadel_uploads")

	viper.SetDefault("prometheus_enabled", true)
	viper.SetDefault("grafana_enabled", true)
	viper.SetDefault("log_level", "info")
	viper.SetDefault("loki_url", "")

	viper.SetDefault("max_concurrent_executions", 100)
	viper.SetDefault("max_concurrent_nodes", 50)
	viper.SetDefault("default_workflow_timeout", "30m")
	viper.SetDefault("max_retries", 3)
	viper.SetDefault("retry_delay", "1s")
	viper.SetDefault("enable_profiling", false)
	viper.SetDefault("enable_caching", true)
	viper.SetDefault("cache_ttl", "1h")

	// Set environment variable prefix
	viper.SetEnvPrefix("CITADEL")
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		// Config file not found, that's ok
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate critical configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// validateConfig validates critical configuration values
func validateConfig(cfg *Config) error {
	// Validate JWT secrets in production
	if cfg.AppEnv == "production" {
		if cfg.JWTSecret == "" || len(cfg.JWTSecret) < 32 {
			return fmt.Errorf("jwt_secret must be set and at least 32 characters in production")
		}
		if cfg.JWTRefreshSecret == "" || len(cfg.JWTRefreshSecret) < 32 {
			return fmt.Errorf("jwt_refresh_secret must be set and at least 32 characters in production")
		}
	}

	return nil
}
