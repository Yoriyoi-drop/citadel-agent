package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/citadel-agent/backend/internal/auth"
	"github.com/citadel-agent/backend/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// RBACMiddleware provides role-based access control for routes
type RBACMiddleware struct {
	rbacManager *auth.RBACManager
}

// NewRBACMiddleware creates a new RBAC middleware
func NewRBACMiddleware(rbacManager *auth.RBACManager) *RBACMiddleware {
	return &RBACMiddleware{
		rbacManager: rbacManager,
	}
}

// RequirePermission creates a middleware that checks if the user has a specific permission
func (rbac *RBACMiddleware) RequirePermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := getUserIDFromContext(c)
		if userID == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Unauthorized: no user context",
			})
		}

		hasPermission, err := rbac.rbacManager.HasPermission(c.Context(), userID, permission)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": fmt.Sprintf("Error checking permission: %v", err),
			})
		}

		if !hasPermission {
			return c.Status(403).JSON(fiber.Map{
				"error": "Forbidden: insufficient permissions",
			})
		}

		return c.Next()
	}
}

// RequireAnyPermission creates a middleware that checks if the user has at least one of the required permissions
func (rbac *RBACMiddleware) RequireAnyPermission(permissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := getUserIDFromContext(c)
		if userID == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Unauthorized: no user context",
			})
		}

		hasAnyPermission := false
		var lastErr error

		for _, permission := range permissions {
			hasPermission, err := rbac.rbacManager.HasPermission(c.Context(), userID, permission)
			if err != nil {
				lastErr = err
				continue
			}
			if hasPermission {
				hasAnyPermission = true
				break
			}
		}

		if !hasAnyPermission {
			if lastErr != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": fmt.Sprintf("Error checking permissions: %v", lastErr),
				})
			}
			return c.Status(403).JSON(fiber.Map{
				"error": "Forbidden: insufficient permissions",
			})
		}

		return c.Next()
	}
}

// RequireAllPermissions creates a middleware that checks if the user has all of the required permissions
func (rbac *RBACMiddleware) RequireAllPermissions(permissions []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := getUserIDFromContext(c)
		if userID == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Unauthorized: no user context",
			})
		}

		authorized, err := rbac.rbacManager.Authorize(c.Context(), userID, permissions)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": fmt.Sprintf("Error checking permissions: %v", err),
			})
		}

		if !authorized {
			return c.Status(403).JSON(fiber.Map{
				"error": "Forbidden: insufficient permissions",
			})
		}

		return c.Next()
	}
}

// RequireRole creates a middleware that checks if the user has a specific role
func (rbac *RBACMiddleware) RequireRole(roleName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := getUserIDFromContext(c)
		if userID == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Unauthorized: no user context",
			})
		}

		userRoles, err := rbac.rbacManager.GetUserRoles(c.Context(), userID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": fmt.Sprintf("Error getting user roles: %v", err),
			})
		}

		hasRole := false
		for _, role := range userRoles {
			if role.Name == roleName || role.ID == roleName {
				hasRole = true
				break
			}
		}

		if !hasRole {
			return c.Status(403).JSON(fiber.Map{
				"error": "Forbidden: required role not assigned",
			})
		}

		return c.Next()
	}
}

// RequireAnyRole creates a middleware that checks if the user has at least one of the required roles
func (rbac *RBACMiddleware) RequireAnyRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := getUserIDFromContext(c)
		if userID == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Unauthorized: no user context",
			})
		}

		userRoles, err := rbac.rbacManager.GetUserRoles(c.Context(), userID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": fmt.Sprintf("Error getting user roles: %v", err),
			})
		}

		hasAnyRole := false
		for _, requiredRole := range roles {
			for _, userRole := range userRoles {
				if userRole.Name == requiredRole || userRole.ID == requiredRole {
					hasAnyRole = true
					break
				}
			}
			if hasAnyRole {
				break
			}
		}

		if !hasAnyRole {
			return c.Status(403).JSON(fiber.Map{
				"error": "Forbidden: required role not assigned",
			})
		}

		return c.Next()
	}
}

// ResourceBasedAccess creates a middleware that checks for specific resource-level access
func (rbac *RBACMiddleware) ResourceBasedAccess(resource, action string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := getUserIDFromContext(c)
		if userID == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Unauthorized: no user context",
			})
		}

		hasAccess, err := rbac.rbacManager.CheckUserHasResourceAccess(c.Context(), userID, resource, action)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": fmt.Sprintf("Error checking resource access: %v", err),
			})
		}

		if !hasAccess {
			return c.Status(403).JSON(fiber.Map{
				"error": "Forbidden: insufficient permissions for resource access",
			})
		}

		return c.Next()
	}
}

// getUserIDFromContext extracts the user ID from the request context
// This assumes that authentication middleware has already set the user ID in the context
func getUserIDFromContext(c *fiber.Ctx) string {
	// Try to get user ID from context (set by auth middleware)
	userID := c.Locals("user_id")
	if userID != nil {
		if id, ok := userID.(string); ok {
			return id
		}
	}

	// Try to get user ID from JWT token
	authHeader := c.Get("Authorization")
	if authHeader == "" || len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return ""
	}

	tokenString := authHeader[7:]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Here you would return your signing key
		// In a real implementation, you would get this from config
		return []byte("your-secret-key"), nil
	})

	if err != nil || !token.Valid {
		return ""
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return ""
	}

	if userIDClaim, exists := claims["user_id"]; exists {
		if id, ok := userIDClaim.(string); ok {
			return id
		}
	}

	return ""
}

// RBACResponse represents the structure of RBAC-related API responses
type RBACResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Error     string      `json:"error,omitempty"`
	ErrorCode string      `json:"error_code,omitempty"`
}

// RegisterRBACRoutes registers RBAC-related API routes
func RegisterRBACRoutes(app *fiber.App, rbacMiddleware *RBACMiddleware, rbacManager *auth.RBACManager) {
	rbacRoutes := app.Group("/api/v1/rbac")

	// Get all roles
	rbacRoutes.Get("/roles", rbacMiddleware.RequirePermission("rbac:read"), func(c *fiber.Ctx) error {
		roles, err := rbacManager.GetAllRoles(c.Context())
		if err != nil {
			return c.Status(500).JSON(RBACResponse{
				Success: false,
				Error:   err.Error(),
			})
		}

		return c.JSON(RBACResponse{
			Success: true,
			Data:    roles,
		})
	})

	// Check user permissions
	rbacRoutes.Get("/user/:userID/permissions", rbacMiddleware.RequirePermission("rbac:read"), func(c *fiber.Ctx) error {
		userID := c.Params("userID")
		if userID == "" {
			return c.Status(400).JSON(RBACResponse{
				Success: false,
				Error:   "User ID is required",
			})
		}

		userRoles, err := rbacManager.GetUserRoles(c.Context(), userID)
		if err != nil {
			return c.Status(500).JSON(RBACResponse{
				Success: false,
				Error:   err.Error(),
			})
		}

		// Combine all permissions from all roles
		var allPermissions []string
		for _, role := range userRoles {
			allPermissions = append(allPermissions, role.Permissions...)
		}

		return c.JSON(RBACResponse{
			Success: true,
			Data: map[string]interface{}{
				"roles":       userRoles,
				"permissions": allPermissions,
			},
		})
	})
}