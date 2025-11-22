// backend/internal/api/handlers/api_key_handler.go
package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/citadel-agent/backend/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// APIKeyHandler handles API key-related HTTP requests
type APIKeyHandler struct {
	apiKeyService *services.APIKeyService
}

// NewAPIKeyHandler creates a new API key handler
func NewAPIKeyHandler(apiKeyService *services.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyService: apiKeyService,
	}
}

// CreateAPIKey creates a new API key
func (akh *APIKeyHandler) CreateAPIKey(c *fiber.Ctx) error {
	var req struct {
		Name        string    `json:"name"`
		UserID      *string   `json:"user_id,omitempty"`
		TeamID      *string   `json:"team_id,omitempty"`
		Permissions []string  `json:"permissions"`
		ExpiresAt   *string   `json:"expires_at,omitempty"` // RFC3339 format
		CreatedBy   string    `json:"created_by"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "API key name is required",
		})
	}

	if len(req.Permissions) == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "At least one permission is required",
		})
	}

	if req.CreatedBy == "" {
		// Get user ID from context if available (from auth middleware)
		if userID := c.Locals("user_id"); userID != nil {
			if id, ok := userID.(string); ok {
				req.CreatedBy = id
			}
		}
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		parsed, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid expires_at format, must be RFC3339",
			})
		}
		expiresAt = &parsed
	}

	// Validate UUID formats if provided
	if req.UserID != nil {
		if _, err := uuid.Parse(*req.UserID); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid user ID format",
			})
		}
	}

	if req.TeamID != nil {
		if _, err := uuid.Parse(*req.TeamID); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid team ID format",
			})
		}
	}

	apiKey, err := akh.apiKeyService.CreateAPIKey(c.Context(), req.UserID, req.TeamID, req.Name, req.Permissions, expiresAt, req.CreatedBy)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create API key: %v", err),
		})
	}

	// Remove the actual key from response if not needed (security)
	returnKey := *apiKey
	delete(returnKey.Metadata, "key") // Don't return the actual key in subsequent calls

	return c.JSON(fiber.Map{
		"success": true,
		"data":    apiKey, // Include the key in the initial creation response
		"message": "API key created successfully",
	})
}

// GetAPIKey retrieves an API key by ID
func (akh *APIKeyHandler) GetAPIKey(c *fiber.Ctx) error {
	apiKeyID := c.Params("apiKeyId")

	// Validate UUID format
	if _, err := uuid.Parse(apiKeyID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid API key ID format",
		})
	}

	// Note: In a real implementation, we would retrieve the key from the database
	// but not return the actual secret key value for security
	// For now, this is a placeholder implementation
	// The actual retrieval should happen via prefix as in validation

	// We'll return an error since we don't have a method to retrieve by ID directly
	// for security reasons (don't return the actual key)
	return c.JSON(fiber.Map{
		"success": false,
		"error":   "API key details retrieval by ID requires special permissions for security",
		"data": fiber.Map{
			"api_key_id": apiKeyID,
		},
	})
}

// ListAPIKeys retrieves a list of API keys
func (akh *APIKeyHandler) ListAPIKeys(c *fiber.Ctx) error {
	userIDStr := c.Query("user_id")
	teamIDStr := c.Query("team_id")
	revokedStr := c.Query("revoked")
	pageStr := c.Query("page", "1")
	limitStr := c.Query("limit", "50")

	var userID *string
	if userIDStr != "" {
		if _, err := uuid.Parse(userIDStr); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid user ID format",
			})
		}
		userID = &userIDStr
	}

	var teamID *string
	if teamIDStr != "" {
		if _, err := uuid.Parse(teamIDStr); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid team ID format",
			})
		}
		teamID = &teamIDStr
	}

	var revoked *bool
	if revokedStr != "" {
		r, err := strconv.ParseBool(revokedStr)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid revoked parameter, must be true or false",
			})
		}
		revoked = &r
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 50
	}

	apiKeys, err := akh.apiKeyService.ListAPIKeys(c.Context(), userID, teamID, revoked, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to list API keys: %v", err),
		})
	}

	// Don't return the actual key values for security
	for _, key := range apiKeys {
		delete(key.Metadata, "key")
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    apiKeys,
		"count":   len(apiKeys),
		"page":    page,
		"limit":   limit,
	})
}

// RevokeAPIKey revokes an API key
func (akh *APIKeyHandler) RevokeAPIKey(c *fiber.Ctx) error {
	apiKeyID := c.Params("apiKeyId")

	// Validate UUID format
	if _, err := uuid.Parse(apiKeyID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid API key ID format",
		})
	}

	// Get user ID from context (from auth middleware)
	revokedBy := ""
	if userID := c.Locals("user_id"); userID != nil {
		if id, ok := userID.(string); ok {
			revokedBy = id
		}
	}

	if revokedBy == "" {
		return c.Status(401).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	err := akh.apiKeyService.RevokeAPIKey(c.Context(), apiKeyID, revokedBy)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to revoke API key: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "API key revoked successfully",
	})
}

// RotateAPIKey rotates an existing API key (creates new one, revokes old one)
func (akh *APIKeyHandler) RotateAPIKey(c *fiber.Ctx) error {
	apiKeyID := c.Params("apiKeyId")

	// Validate UUID format
	if _, err := uuid.Parse(apiKeyID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid API key ID format",
		})
	}

	// Get user ID from context (from auth middleware)
	rotatedBy := ""
	if userID := c.Locals("user_id"); userID != nil {
		if id, ok := userID.(string); ok {
			rotatedBy = id
		}
	}

	if rotatedBy == "" {
		return c.Status(401).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	newKey, err := akh.apiKeyService.RotateAPIKey(c.Context(), apiKeyID, rotatedBy)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to rotate API key: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    newKey,
		"message": "API key rotated successfully",
	})
}

// ValidateAPIKey validates an API key (this would typically be middleware)
func (akh *APIKeyHandler) ValidateAPIKey(c *fiber.Ctx) error {
	var req struct {
		Key string `json:"key"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Key == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "API key is required",
		})
	}

	apiKey, err := akh.apiKeyService.ValidateAPIKey(c.Context(), req.Key)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid API key: %v", err),
		})
	}

	// Don't return the actual key for security
	returnKey := *apiKey
	delete(returnKey.Metadata, "key")

	return c.JSON(fiber.Map{
		"success": true,
		"data":    returnKey,
		"message": "API key is valid",
	})
}

// GetUserAPIKeys retrieves all API keys for a user
func (akh *APIKeyHandler) GetUserAPIKeys(c *fiber.Ctx) error {
	userID := c.Params("userId")

	// Validate UUID format
	if _, err := uuid.Parse(userID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	// Check if the requesting user is the same as the target user
	requestingUserID := c.Locals("user_id")
	if requestingUserID == nil || requestingUserID != userID {
		return c.Status(403).JSON(fiber.Map{
			"error": "Access denied: cannot access other users' API keys",
		})
	}

	apiKeys, err := akh.apiKeyService.GetUserAPIKeys(c.Context(), userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get user API keys: %v", err),
		})
	}

	// Don't return the actual key values for security
	for _, key := range apiKeys {
		delete(key.Metadata, "key")
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    apiKeys,
		"count":   len(apiKeys),
	})
}

// GetTeamAPIKeys retrieves all API keys for a team
func (akh *APIKeyHandler) GetTeamAPIKeys(c *fiber.Ctx) error {
	teamID := c.Params("teamId")

	// Validate UUID format
	if _, err := uuid.Parse(teamID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid team ID format",
		})
	}

	apiKeys, err := akh.apiKeyService.GetTeamAPIKeys(c.Context(), teamID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get team API keys: %v", err),
		})
	}

	// Don't return the actual key values for security
	for _, key := range apiKeys {
		delete(key.Metadata, "key")
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    apiKeys,
		"count":   len(apiKeys),
	})
}

// CheckPermission checks if an API key has a specific permission
func (akh *APIKeyHandler) CheckPermission(c *fiber.Ctx) error {
	var req struct {
		Key        string `json:"key"`
		Permission string `json:"permission"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Key == "" || req.Permission == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "API key and permission are required",
		})
	}

	// Validate the API key first
	apiKey, err := akh.apiKeyService.ValidateAPIKey(c.Context(), req.Key)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid API key: %v", err),
		})
	}

	// Check if the key has the required permission
	hasPermission := akh.apiKeyService.CheckPermission(c.Context(), apiKey, req.Permission)

	return c.JSON(fiber.Map{
		"success":    true,
		"has_permission": hasPermission,
		"permission": req.Permission,
		"api_key_id": apiKey.ID,
	})
}

// RegisterRoutes registers API key handler routes
func (akh *APIKeyHandler) RegisterRoutes(router fiber.Router) {
	// Public validation endpoint
	router.Post("/api-keys/validate", akh.ValidateAPIKey)
	router.Post("/api-keys/check-permission", akh.CheckPermission)
	
	// Authenticated endpoints
	authGroup := router.Use() // This would use the auth middleware in a real implementation
	{
		// Key management
		authGroup.Post("/api-keys", akh.CreateAPIKey)
		authGroup.Get("/api-keys", akh.ListAPIKeys)
		authGroup.Post("/api-keys/:apiKeyId/revoke", akh.RevokeAPIKey)
		authGroup.Post("/api-keys/:apiKeyId/rotate", akh.RotateAPIKey)
		
		// User/Team specific endpoints
		authGroup.Get("/users/:userId/api-keys", akh.GetUserAPIKeys)
		authGroup.Get("/teams/:teamId/api-keys", akh.GetTeamAPIKeys)
	}
}