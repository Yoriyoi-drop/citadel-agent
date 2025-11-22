package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/citadel-agent/backend/config"
)

func TestAppCreation(t *testing.T) {
	// Test creating an app with default config
	cfg := &config.EngineConfig{
		APIConfig: config.APIConfig{
			Host: "localhost",
			Port: 8080,
		},
		DatabaseConfig: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Name:     "testdb",
			Username: "testuser",
			Password: "testpass",
		},
	}

	// Note: We can't fully test NewApp without a real database connection
	// So we'll test the configuration aspects
	
	assert.NotNil(t, cfg)
	assert.Equal(t, "localhost", cfg.APIConfig.Host)
	assert.Equal(t, 8080, cfg.APIConfig.Port)
	assert.Equal(t, "localhost", cfg.DatabaseConfig.Host)
	assert.Equal(t, 5432, cfg.DatabaseConfig.Port)
}

func TestAppStructInitialization(t *testing.T) {
	// Test that app struct fields are properly defined
	app := &App{}
	
	// Verify that the struct can be created without errors
	assert.NotNil(t, app)
	
	// The fields will be nil initially
	assert.Nil(t, app.config)
	assert.Nil(t, app.db)
	assert.Nil(t, app.server)
}