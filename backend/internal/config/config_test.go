package config

import (
	"os"
	"testing"
)

func TestGetEnvOrDefault(t *testing.T) {
	// Test case 1: Environment variable exists
	os.Setenv("TEST_VAR", "test_value")
	result := getEnvOrDefault("TEST_VAR", "default_value")
	if result != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", result)
	}

	// Test case 2: Environment variable does not exist
	os.Unsetenv("TEST_VAR")
	result = getEnvOrDefault("TEST_VAR", "default_value")
	if result != "default_value" {
		t.Errorf("Expected 'default_value', got '%s'", result)
	}

	// Test case 3: Empty environment variable value
	os.Setenv("TEST_EMPTY", "")
	result = getEnvOrDefault("TEST_EMPTY", "default_value")
	if result != "default_value" {
		t.Errorf("Expected 'default_value', got '%s'", result)
	}

	// Cleanup
	os.Unsetenv("TEST_EMPTY")
}

func TestLoadConfig(t *testing.T) {
	// Set up test environment variables
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("DATABASE_URL", "postgresql://test:test@test:5432/testdb")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("ENVIRONMENT", "test")

	// Load config
	config := LoadConfig()

	// Verify values
	if config.ServerPort != "8080" {
		t.Errorf("Expected ServerPort '8080', got '%s'", config.ServerPort)
	}

	if config.DatabaseURL != "postgresql://test:test@test:5432/testdb" {
		t.Errorf("Expected DatabaseURL 'postgresql://test:test@test:5432/testdb', got '%s'", config.DatabaseURL)
	}

	if config.JWTSecret != "test-secret" {
		t.Errorf("Expected JWTSecret 'test-secret', got '%s'", config.JWTSecret)
	}

	if config.Environment != "test" {
		t.Errorf("Expected Environment 'test', got '%s'", config.Environment)
	}

	// Test with defaults
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("ENVIRONMENT")

	config = LoadConfig()

	if config.ServerPort != "5001" {
		t.Errorf("Expected default ServerPort '5001', got '%s'", config.ServerPort)
	}

	if config.Environment != "development" {
		t.Errorf("Expected default Environment 'development', got '%s'", config.Environment)
	}
}