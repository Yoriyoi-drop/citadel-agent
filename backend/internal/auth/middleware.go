// backend/internal/auth/middleware.go
package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Context keys for user info
const (
	UserContextKey = "user"
	UserIDContextKey = "user_id"
)

// AuthMiddleware creates a middleware that validates JWT tokens
func (s *AuthService) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		tokenString := ""
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization format, use 'Bearer <token>'",
			})
			c.Abort()
			return
		}

		user, err := s.ValidateToken(c.Request.Context(), tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Add user to context
		c.Set(UserContextKey, user)
		c.Set(UserIDContextKey, user.ID)
		c.Next()
	}
}

// RBACMiddleware creates a middleware that checks permissions for specific resource and action
func (s *AuthService) RBACMiddleware(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, ensure user is authenticated
		userData, exists := c.Get(UserContextKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		user, ok := userData.(*User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid user data in context",
			})
			c.Abort()
			return
		}

		// Check if user has required permission
		hasPermission, err := s.HasPermission(c.Request.Context(), user.ID, resource, action)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to check permissions",
			})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserFromContext retrieves the authenticated user from the request context
func GetUserFromContext(c *gin.Context) (*User, error) {
	userData, exists := c.Get(UserContextKey)
	if !exists {
		return nil, errors.New("no user in context")
	}

	user, ok := userData.(*User)
	if !ok {
		return nil, errors.New("invalid user type in context")
	}

	return user, nil
}

// GetUserIDFromContext retrieves the authenticated user ID from the request context
func GetUserIDFromContext(c *gin.Context) (uint, error) {
	userIDData, exists := c.Get(UserIDContextKey)
	if !exists {
		return 0, errors.New("no user ID in context")
	}

	userID, ok := userIDData.(uint)
	if !ok {
		return 0, errors.New("invalid user ID type in context")
	}

	return userID, nil
}

// RequireRoleMiddleware creates a middleware that checks if the user has a specific role
func (s *AuthService) RequireRoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := GetUserFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		// Check if user has the required role
		hasRole := false
		for _, role := range user.Roles {
			if role.Name == requiredRole {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "User does not have required role: " + requiredRole,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminOnlyMiddleware restricts access to admin users only
func (s *AuthService) AdminOnlyMiddleware() gin.HandlerFunc {
	return s.RequireRoleMiddleware("admin")
}

// PermissionChecker checks permissions for specific resource and action
func (s *AuthService) PermissionChecker(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := GetUserFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		hasPermission, err := s.HasPermission(c.Request.Context(), user.ID, resource, action)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to check permissions",
			})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions for " + resource + ":" + action,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ContextUserLoader loads the user data into context (useful for non-authenticated routes that may need user info)
func (s *AuthService) ContextUserLoader() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth header, continue without user in context
			c.Next()
			return
		}

		// Extract token from "Bearer <token>" format
		tokenString := ""
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			c.Next() // Not a proper bearer token, continue
			return
		}

		user, err := s.ValidateToken(c.Request.Context(), tokenString)
		if err != nil {
			// Invalid token, continue without user in context
			c.Next()
			return
		}

		// Add user to context
		c.Set(UserContextKey, user)
		c.Set(UserIDContextKey, user.ID)
		c.Next()
	}
}