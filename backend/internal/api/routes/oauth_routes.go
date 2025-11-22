package routes

import (
	"github.com/citadel-agent/backend/internal/api/controllers"
	"github.com/gofiber/fiber/v2"
)

// RegisterOAuthRoutes registers OAuth-related routes
func RegisterOAuthRoutes(app *fiber.App, authController *controllers.AuthController) {
	// Create OAuth controller
	oauthController := controllers.NewOAuthController()

	// OAuth routes
	oauth := app.Group("/api/v1/auth")

	// GitHub OAuth routes
	oauth.Get("/github", oauthController.GithubAuth)
	oauth.Get("/github/callback", oauthController.GithubCallback)

	// Google OAuth routes
	oauth.Get("/google", oauthController.GoogleAuth)
	oauth.Get("/google/callback", oauthController.GoogleCallback)
}