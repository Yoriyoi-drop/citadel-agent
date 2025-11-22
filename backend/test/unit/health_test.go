package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestHealthEndpoint(t *testing.T) {
	// Create a new app instance
	app := fiber.New()

	// Define a simple health endpoint for testing
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"version": "1.0.0",
		})
	})

	// Create a request to the health endpoint
	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req, -1)

	// Assert that there is no error
	assert.NoError(t, err)
	
	// Assert that the status code is 200
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse the response body
	var healthResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&healthResp)
	assert.NoError(t, err)

	// Assert that the status is healthy
	assert.Equal(t, "healthy", healthResp["status"])
	assert.Equal(t, "1.0.0", healthResp["version"])
}