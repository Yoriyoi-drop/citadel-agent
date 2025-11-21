package config

import (
	"os"
)

// Config holds the application configuration
type Config struct {
	ServerPort    string
	DatabaseURL   string
	JWTSecret     string
	GithubClientID     string
	GithubClientSecret string
	GithubRedirectURI  string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURI  string
	Environment   string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		ServerPort: getEnvOrDefault("SERVER_PORT", "5001"),
		DatabaseURL: getEnvOrDefault("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/citadel_agent"),
		JWTSecret: getEnvOrDefault("JWT_SECRET", "your-default-jwt-secret-for-dev-change-in-production"),
		GithubClientID: getEnvOrDefault("GITHUB_CLIENT_ID", ""),
		GithubClientSecret: getEnvOrDefault("GITHUB_CLIENT_SECRET", ""),
		GithubRedirectURI: getEnvOrDefault("GITHUB_REDIRECT_URI", "http://localhost:5001/auth/oauth/github/callback"),
		GoogleClientID: getEnvOrDefault("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnvOrDefault("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURI: getEnvOrDefault("GOOGLE_REDIRECT_URI", "http://localhost:5001/auth/oauth/google/callback"),
		Environment: getEnvOrDefault("ENVIRONMENT", "development"),
	}
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}