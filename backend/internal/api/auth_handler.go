// backend/internal/api/auth_handler.go
package api

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/citadel-agent/backend/internal/services"
	"github.com/citadel-agent/backend/internal/models"
)

// AuthHandler handles HTTP requests for authentication operations
type AuthHandler struct {
	service *services.AuthService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(service *services.AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

// Register handles POST /api/v1/auth/register
func (ah *AuthHandler) Register(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if req.Email == "" || req.Password == "" || req.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Email, password, and name are required",
		})
	}

	result, err := ah.service.RegisterUser(c.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": fmt.Sprintf("Registration failed: %v", err),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"user": result.User,
		"token": fiber.Map{
			"access_token":  result.Token,
			"refresh_token": result.Refresh,
			"token_type":    result.TokenType,
			"expires_in":    result.ExpiresIn,
		},
		"team": fiber.Map{
			"id":   result.TeamID,
			"role": result.TeamRole,
		},
		"message": "Registration successful",
	})
}

// Login handles POST /api/v1/auth/login
func (ah *AuthHandler) Login(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if req.Email == "" || req.Password == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Email and password are required",
		})
	}

	result, err := ah.service.LoginUser(c.Context(), req.Email, req.Password)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": fmt.Sprintf("Login failed: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"user": result.User,
		"token": fiber.Map{
			"access_token":  result.Token,
			"refresh_token": result.Refresh,
			"token_type":    result.TokenType,
			"expires_in":    result.ExpiresIn,
		},
		"team": fiber.Map{
			"id":   result.TeamID,
			"role": result.TeamRole,
		},
		"message": "Login successful",
	})
}

// Refresh handles POST /api/v1/auth/refresh
func (ah *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	result, err := ah.service.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": fmt.Sprintf("Token refresh failed: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"token": fiber.Map{
			"access_token":  result.Token,
			"token_type":    result.TokenType,
			"expires_in":    result.ExpiresIn,
		},
		"message": "Token refreshed successfully",
	})
}

// Me handles GET /api/v1/auth/me
func (ah *AuthHandler) Me(c *fiber.Ctx) error {
	// Extract user ID from context (assumes middleware adds user info)
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	user, err := ah.service.GetCurrentUser(c.Context(), userID.(string))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get user: %v", err),
		})
	}

	return c.JSON(user)
}

// UpdateProfile handles PUT /api/v1/auth/profile
func (ah *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	// Extract user ID from context
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	var req models.UserProfile
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	updatedUser, err := ah.service.UpdateUserProfile(c.Context(), userID.(string), req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to update profile: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"user": updatedUser,
		"message": "Profile updated successfully",
	})
}

// UpdatePreferences handles PUT /api/v1/auth/preferences
func (ah *AuthHandler) UpdatePreferences(c *fiber.Ctx) error {
	// Extract user ID from context
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	var req map[string]interface{}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	updatedUser, err := ah.service.UpdateUserPreferences(c.Context(), userID.(string), req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to update preferences: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"user": updatedUser,
		"message": "Preferences updated successfully",
	})
}

// ChangePassword handles PUT /api/v1/auth/password
func (ah *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	// Extract user ID from context
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Old and new passwords are required",
		})
	}

	err := ah.service.ChangePassword(c.Context(), userID.(string), req.OldPassword, req.NewPassword)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": fmt.Sprintf("Password change failed: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Password changed successfully",
	})
}

// CreateAPIKey handles POST /api/v1/auth/api-keys
func (ah *AuthHandler) CreateAPIKey(c *fiber.Ctx) error {
	// Extract user ID from context
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	var req struct {
		Name        string    `json:"name"`
		Permissions []string  `json:"permissions"`
		ExpiresIn   *int      `json:"expires_in_days"` // Optional: days until expiration
		TeamID      *string   `json:"team_id"`         // Optional: team ID
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var expiresAt *time.Time
	if req.ExpiresIn != nil {
		expiryTime := time.Now().AddDate(0, 0, *req.ExpiresIn)
		expiresAt = &expiryTime
	}

	userIDStr := userID.(string)
	teamID := req.TeamID
	if teamID == nil {
		// If no team ID provided, use the user's default team
		// In a real system, you would look this up
	}

	apiKey, err := ah.service.CreateAPIKey(c.Context(), &userIDStr, teamID, req.Name, req.Permissions, expiresAt)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create API key: %v", err),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"api_key": fiber.Map{
			"id":          apiKey.ID,
			"name":        apiKey.Name,
			"prefix":      apiKey.Prefix, // Only return prefix for security
			"permissions": apiKey.Permissions,
			"created_at":  apiKey.CreatedAt,
			"expires_at":  apiKey.ExpiresAt,
		},
		"message": "API key created successfully",
	})
}

// GetAPIKeys handles GET /api/v1/auth/api-keys
func (ah *AuthHandler) GetAPIKeys(c *fiber.Ctx) error {
	// Extract user ID from context
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// In a real system, you would query the repository to get API keys
	// For now, return an empty array
	return c.JSON(fiber.Map{
		"data":      []interface{}{}, // Return empty list for now
		"message":   "API keys retrieved",
	})
}

// RevokeAPIKey handles DELETE /api/v1/auth/api-keys/:id
func (ah *AuthHandler) RevokeAPIKey(c *fiber.Ctx) error {
	// Extract user ID from context
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	apiKeyID := c.Params("id")
	
	err := ah.service.RevokeAPIKey(c.Context(), apiKeyID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to revoke API key: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message": "API key revoked successfully",
	})
}

// Logout handles POST /api/v1/auth/logout
func (ah *AuthHandler) Logout(c *fiber.Ctx) error {
	// In a real implementation, you would add the token to a blacklist
	// For now, just return a success message
	return c.JSON(fiber.Map{
		"message": "Successfully logged out",
	})
}

// RegisterAuthRoutes registers authentication routes
func (ah *AuthHandler) RegisterAuthRoutes(app *fiber.App) {
	auth := app.Group("/api/v1/auth")
	
	auth.Post("/register", ah.Register)
	auth.Post("/login", ah.Login)
	auth.Post("/refresh", ah.Refresh)
	auth.Post("/logout", ah.Logout)
	
	// Protected routes (require authentication middleware)
	protected := auth.Use(ah.AuthMiddleware)
	protected.Get("/me", ah.Me)
	protected.Put("/profile", ah.UpdateProfile)
	protected.Put("/preferences", ah.UpdatePreferences)
	protected.Put("/password", ah.ChangePassword)
	protected.Post("/api-keys", ah.CreateAPIKey)
	protected.Get("/api-keys", ah.GetAPIKeys)
	protected.Delete("/api-keys/:id", ah.RevokeAPIKey)
}

// AuthMiddleware is a placeholder for JWT authentication middleware
func (ah *AuthHandler) AuthMiddleware(c *fiber.Ctx) error {
	// This is a placeholder implementation
	// In a real system, you would validate JWT tokens and extract user info
	
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(401).JSON(fiber.Map{
			"error": "Authorization header missing",
		})
	}
	
	// Check if it's a Bearer token
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return c.Status(401).JSON(fiber.Map{
			"error": "Invalid authorization header format",
		})
	}
	
	// Extract token
	token := authHeader[7:]
	
	// In a real implementation, you would validate the JWT token here
	// For now, we'll just set a placeholder user ID
	// This is where you would decode the JWT and extract user information
	c.Locals("user_id", "placeholder_user_id")
	c.Locals("user_role", "admin")
	
	return c.Next()
}