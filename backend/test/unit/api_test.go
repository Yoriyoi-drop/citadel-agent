package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/citadel-agent/backend/internal/api"
	"github.com/citadel-agent/backend/internal/auth"
	"github.com/citadel-agent/backend/internal/ai"
	"github.com/citadel-agent/backend/internal/runtimes"
	"github.com/citadel-agent/backend/internal/engine"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// TestServerInitialization tests that the server initializes correctly
func TestServerInitialization(t *testing.T) {
	// Create services with mock dependencies
	authService := auth.NewAuthService(nil) // Using nil for db since we're testing initialization
	aiService := ai.NewAIService()
	runtimeMgr := runtimes.NewMultiRuntimeManager()
	nodeRegistry := engine.NewNodeRegistry()
	executor := engine.NewExecutor(nodeRegistry)
	runner := engine.NewRunner(executor)

	// Create server instance
	server := api.NewServer(
		nil, // Using nil for db in this test
		authService,
		aiService,
		nodeRegistry,
		executor,
		runner,
		runtimeMgr,
	)

	// Verify that server is created
	assert.NotNil(t, server)
	assert.NotNil(t, server.App)
}

// TestAuthEndpoints tests the authentication endpoints
func TestAuthEndpoints(t *testing.T) {
	// Create services with mock dependencies
	authService := auth.NewAuthService(nil)
	aiService := ai.NewAIService()
	runtimeMgr := runtimes.NewMultiRuntimeManager()
	nodeRegistry := engine.NewNodeRegistry()
	executor := engine.NewExecutor(nodeRegistry)
	runner := engine.NewRunner(executor)

	// Create server instance
	server := api.NewServer(
		nil, // Using nil for db in this test
		authService,
		aiService,
		nodeRegistry,
		executor,
		runner,
		runtimeMgr,
	)

	app := server.App

	// Test case: login endpoint exists
	loginPayload := `{
		"email": "test@example.com",
		"password": "password123"
	}`

	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(loginPayload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1) // -1 means no timeout

	assert.NoError(t, err)
	// The actual response may vary, but we expect it to not be a 404
	assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
}

// TestWorkflowEndpoints tests the workflow endpoints
func TestWorkflowEndpoints(t *testing.T) {
	// Create services with mock dependencies
	authService := auth.NewAuthService(nil)
	aiService := ai.NewAIService()
	runtimeMgr := runtimes.NewMultiRuntimeManager()
	nodeRegistry := engine.NewNodeRegistry()
	executor := engine.NewExecutor(nodeRegistry)
	runner := engine.NewRunner(executor)

	// Create server instance
	server := api.NewServer(
		nil, // Using nil for db in this test
		authService,
		aiService,
		nodeRegistry,
		executor,
		runner,
		runtimeMgr,
	)

	app := server.App

	// Test GET workflows endpoint
	req := httptest.NewRequest("GET", "/api/v1/workflows", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	// Expect 401 since this is a protected endpoint
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestNodeEndpoints tests the node endpoints
func TestNodeEndpoints(t *testing.T) {
	// Create services with mock dependencies
	authService := auth.NewAuthService(nil)
	aiService := ai.NewAIService()
	runtimeMgr := runtimes.NewMultiRuntimeManager()
	nodeRegistry := engine.NewNodeRegistry()
	executor := engine.NewExecutor(nodeRegistry)
	runner := engine.NewRunner(executor)

	// Create server instance
	server := api.NewServer(
		nil, // Using nil for db in this test
		authService,
		aiService,
		nodeRegistry,
		executor,
		runner,
		runtimeMgr,
	)

	app := server.App

	// Test GET node types endpoint
	req := httptest.NewRequest("GET", "/api/v1/nodes/types", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify response body
	var nodeTypesResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&nodeTypesResp)
	assert.NoError(t, err)

	// Check that node_types is in the response
	assert.Contains(t, nodeTypesResp, "node_types")

	// Check that node_types is an array
	nodeTypes, ok := nodeTypesResp["node_types"].([]interface{})
	if ok {
		assert.Greater(t, len(nodeTypes), 0)
	}
}