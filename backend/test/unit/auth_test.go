package test

import (
	"testing"

	"github.com/citadel-agent/backend/internal/auth"
	"github.com/stretchr/testify/assert"
)

// TestAuthServiceInitialization tests that the auth service initializes correctly
func TestAuthServiceInitialization(t *testing.T) {
	// Create auth service with nil db (for testing purposes)
	authService := auth.NewAuthService(nil)

	// Verify that service is created
	assert.NotNil(t, authService)
}

// TestAuthMiddleware tests the authentication middleware
func TestAuthMiddleware(t *testing.T) {
	// Create auth service
	authService := auth.NewAuthService(nil)

	// Verify that middleware function is returned
	middleware := authService.AuthMiddleware()
	assert.NotNil(t, middleware)
}

// TestAuthenticateUser tests the authenticate user function
func TestAuthenticateUser(t *testing.T) {
	// Create auth service
	authService := auth.NewAuthService(nil)

	// Test with dummy credentials
	err := authService.AuthenticateUser("test@example.com", "password")
	// This should not return an error in our mock implementation
	assert.NoError(t, err)
}