// backend/internal/database/config.go
package database

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds database configuration
type Config struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	User            string        `json:"user"`
	Password        string        `json:"password"`
	Database        string        `json:"database"`
	SSLMode         string        `json:"ssl_mode"`
	MaxConns        int           `json:"max_conns"`
	MinConns        int           `json:"min_conns"`
	MaxConnLifetime time.Duration `json:"max_conn_lifetime"`
	MaxConnIdleTime time.Duration `json:"max_conn_idle_time"`
	HealthCheck     time.Duration `json:"health_check"`
	PoolTimeout     time.Duration `json:"pool_timeout"`
	DriverName      string        `json:"driver_name"`
	URL             string        `json:"url"`
}

// DefaultConfig returns the default database configuration
func DefaultConfig() *Config {
	return &Config{
		Host:            getEnvOrDefault("DB_HOST", "localhost"),
		Port:            getEnvOrDefaultInt("DB_PORT", 5432),
		User:            getEnvOrDefault("DB_USER", "postgres"),
		Password:        getEnvOrDefault("DB_PASSWORD", "postgres"),
		Database:        getEnvOrDefault("DB_NAME", "citadel_agent"),
		SSLMode:         getEnvOrDefault("DB_SSL_MODE", "disable"),
		MaxConns:        20,
		MinConns:        5,
		MaxConnLifetime: 30 * time.Minute,
		MaxConnIdleTime: 15 * time.Minute,
		HealthCheck:     30 * time.Second,
		PoolTimeout:     5 * time.Second,
		DriverName:      "pgx",
		URL:             "",
	}
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

// Validate validates the database configuration
func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("database port must be between 1 and 65535")
	}

	if c.User == "" {
		return fmt.Errorf("database user is required")
	}

	if c.Database == "" {
		return fmt.Errorf("database name is required")
	}

	if c.MaxConns <= 0 {
		return fmt.Errorf("max connections must be greater than 0")
	}

	if c.MinConns < 0 {
		return fmt.Errorf("min connections cannot be negative")
	}

	if c.MinConns > c.MaxConns {
		return fmt.Errorf("min connections cannot be greater than max connections")
	}

	if c.MaxConnLifetime <= 0 {
		return fmt.Errorf("max connection lifetime must be greater than 0")
	}

	if c.MaxConnIdleTime <= 0 {
		return fmt.Errorf("max connection idle time must be greater than 0")
	}

	if c.PoolTimeout <= 0 {
		return fmt.Errorf("pool timeout must be greater than 0")
	}

	return nil
}

// ConnectionString generates a connection string from the config
func (c *Config) ConnectionString() string {
	if c.URL != "" {
		return c.URL
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.SSLMode,
	)
}

// ConnectionStringWithParams generates a connection string with additional parameters
func (c *Config) ConnectionStringWithParams(params map[string]string) string {
	connStr := c.ConnectionString()
	
	if len(params) == 0 {
		return connStr
	}
	
	for key, value := range params {
		connStr += fmt.Sprintf("&%s=%s", key, value)
	}
	
	return connStr
}