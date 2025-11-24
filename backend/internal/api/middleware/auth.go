package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware handles both JWT and API key authentication
type AuthMiddleware struct {
	jwtSecret     string
	apiKeyService interface {
		ValidateAPIKey(key string) (string, error)
	}
	rbacService interface {
		HasPermission(userID string, permission string) (bool, error)
		GetUserPermissions(userID string) ([]string, error)
	}
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jwtSecret string, apiKeyService, rbacService interface{}) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret:     jwtSecret,
		apiKeyService: apiKeyService,
		rbacService:   rbacService,
	}
}

// Authenticate handles both JWT and API key authentication
func (m *AuthMiddleware) Authenticate() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Try API key first
		apiKey := extractAPIKey(c)
		if apiKey != "" {
			return m.authenticateWithAPIKey(c, apiKey)
		}

		// Fall back to JWT
		token := extractJWT(c)
		if token != "" {
			return m.authenticateWithJWT(c, token)
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing authentication credentials",
		})
	}
}

// authenticateWithAPIKey validates API key and sets user context
func (m *AuthMiddleware) authenticateWithAPIKey(c *fiber.Ctx, apiKey string) error {
	// Validate API key
	userID, err := m.apiKeyService.ValidateAPIKey(apiKey)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid API key",
		})
	}

	// Set user context
	c.Locals("userID", userID)
	c.Locals("authType", "apikey")

	return c.Next()
}

// authenticateWithJWT validates JWT token and sets user context
func (m *AuthMiddleware) authenticateWithJWT(c *fiber.Ctx, token string) error {
	// TODO: Implement JWT validation
	// For now, just pass through
	// This should validate the JWT and extract user ID

	// Placeholder - extract user ID from JWT
	userID := "user-from-jwt" // TODO: Extract from JWT

	c.Locals("userID", userID)
	c.Locals("authType", "jwt")

	return c.Next()
}

// RequirePermission creates a middleware that checks for specific permission
func (m *AuthMiddleware) RequirePermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("userID")
		if userID == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		// Check permission
		hasPermission, err := m.rbacService.HasPermission(userID.(string), permission)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to check permissions",
			})
		}

		if !hasPermission {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":      "Permission denied",
				"permission": permission,
			})
		}

		return c.Next()
	}
}

// RequireAnyPermission checks if user has any of the specified permissions
func (m *AuthMiddleware) RequireAnyPermission(permissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("userID")
		if userID == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		// Check if user has any of the permissions
		for _, perm := range permissions {
			hasPermission, err := m.rbacService.HasPermission(userID.(string), perm)
			if err != nil {
				continue
			}
			if hasPermission {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":       "Permission denied",
			"permissions": permissions,
		})
	}
}

// RequireAllPermissions checks if user has all of the specified permissions
func (m *AuthMiddleware) RequireAllPermissions(permissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("userID")
		if userID == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		// Check if user has all permissions
		for _, perm := range permissions {
			hasPermission, err := m.rbacService.HasPermission(userID.(string), perm)
			if err != nil || !hasPermission {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error":      "Permission denied",
					"permission": perm,
				})
			}
		}

		return c.Next()
	}
}

// extractAPIKey extracts API key from Authorization header or query parameter
func extractAPIKey(c *fiber.Ctx) string {
	// Try Authorization header first (Bearer token)
	auth := c.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		token := strings.TrimPrefix(auth, "Bearer ")
		// Check if it's an API key (starts with cta_)
		if strings.HasPrefix(token, "cta_") {
			return token
		}
	}

	// Try X-API-Key header
	apiKey := c.Get("X-API-Key")
	if apiKey != "" {
		return apiKey
	}

	// Try query parameter (not recommended for production)
	apiKey = c.Query("api_key")
	if apiKey != "" {
		return apiKey
	}

	return ""
}

// extractJWT extracts JWT token from Authorization header
func extractJWT(c *fiber.Ctx) string {
	auth := c.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		token := strings.TrimPrefix(auth, "Bearer ")
		// Make sure it's not an API key
		if !strings.HasPrefix(token, "cta_") {
			return token
		}
	}
	return ""
}
