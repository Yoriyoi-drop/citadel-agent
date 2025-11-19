package middlewares

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/jwt/v3"

	"citadel-agent/backend/internal/config"
)

// AuthMiddleware creates an authentication middleware
func AuthMiddleware(cfg *config.Config) fiber.Handler {
	return jwt.New(jwt.Config{
		SigningKey: []byte(cfg.JWT.Secret),
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Printf("Authentication error: %v", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized access",
			})
		},
	})
}

// OptionalAuthMiddleware creates an optional authentication middleware
func OptionalAuthMiddleware(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			// No token provided, continue without user context
			return c.Next()
		}

		// Use the standard JWT middleware for validation
		return jwt.New(jwt.Config{
			SigningKey: []byte(cfg.JWT.Secret),
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				// If token is invalid, continue without user context instead of error
				return c.Next()
			},
		})(c)
	}
}