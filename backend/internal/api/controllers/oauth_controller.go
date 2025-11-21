package controllers

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// OAuthController handles OAuth authentication flows
type OAuthController struct {
	githubConfig  *oauth2.Config
	googleConfig  *oauth2.Config
	jwtSecret     string
	callbackURL   string
}

// NewOAuthController creates a new OAuth controller
func NewOAuthController() *OAuthController {
	// Initialize GitHub OAuth config
	githubConfig := &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GITHUB_CALLBACK_URL"), // e.g., http://localhost:5001/api/v1/auth/github/callback
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	// Initialize Google OAuth config
	googleConfig := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_CALLBACK_URL"), // e.g., http://localhost:5001/api/v1/auth/google/callback
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}

	return &OAuthController{
		githubConfig: githubConfig,
		googleConfig: googleConfig,
		jwtSecret:    os.Getenv("JWT_SECRET"),
	}
}

// GithubAuth redirects to GitHub OAuth
func (c *OAuthController) GithubAuth(ctx *fiber.Ctx) error {
	url := c.githubConfig.AuthCodeURL("state-token", oauth2.AccessTypeOnline)
	return ctx.Redirect(url)
}

// GithubCallback handles GitHub OAuth callback
func (c *OAuthController) GithubCallback(ctx *fiber.Ctx) error {
	// Get the authorization code from the callback
	code := ctx.Query("code")
	if code == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No authorization code provided",
		})
	}

	// Exchange code for token
	token, err := c.githubConfig.Exchange(ctx.Context(), code)
	if err != nil {
		log.Printf("Failed to exchange code for token: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to exchange code for token",
		})
	}

	// Get user info from GitHub API
	user, err := c.getGithubUser(token.AccessToken)
	if err != nil {
		log.Printf("Failed to get GitHub user: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user info from GitHub",
		})
	}

	// Create or update user in database
	// This would involve creating a UserService and calling appropriate methods
	// For now, this is a simplified implementation

	// Generate JWT token for the user
	// tokenString, err := c.generateJWT(user.ID, user.Email)
	// if err != nil {
	//     return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
	//         "error": "Failed to generate JWT",
	//     })
	// }

	// For now, return the user info as a simple response
	return ctx.JSON(fiber.Map{
		"message": "GitHub authentication successful",
		"user":    user,
		"token":   token.AccessToken, // In a real implementation, this would be your JWT
	})
}

// GoogleAuth redirects to Google OAuth
func (c *OAuthController) GoogleAuth(ctx *fiber.Ctx) error {
	url := c.googleConfig.AuthCodeURL("state-token", oauth2.AccessTypeOnline)
	return ctx.Redirect(url)
}

// GoogleCallback handles Google OAuth callback
func (c *OAuthController) GoogleCallback(ctx *fiber.Ctx) error {
	// Get the authorization code from the callback
	code := ctx.Query("code")
	if code == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No authorization code provided",
		})
	}

	// Exchange code for token
	token, err := c.googleConfig.Exchange(ctx.Context(), code)
	if err != nil {
		log.Printf("Failed to exchange code for token: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to exchange code for token",
		})
	}

	// Get user info from Google API
	user, err := c.getGoogleUser(token.AccessToken)
	if err != nil {
		log.Printf("Failed to get Google user: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user info from Google",
		})
	}

	// Create or update user in database
	// This would involve creating a UserService and calling appropriate methods

	// For now, return the user info as a simple response
	return ctx.JSON(fiber.Map{
		"message": "Google authentication successful",
		"user":    user,
		"token":   token.AccessToken, // In a real implementation, this would be your JWT
	})
}

// getGithubUser retrieves user information from GitHub API
func (c *OAuthController) getGithubUser(accessToken string) (map[string]interface{}, error) {
	// In a real implementation, this would make an HTTP request to GitHub API
	// For example: GET https://api.github.com/user with Authorization: Bearer {token}
	
	// Mock implementation
	user := map[string]interface{}{
		"id":    "github_user_id",
		"name":  "GitHub User",
		"email": "githubuser@example.com",
		"avatar_url": "https://avatars.githubusercontent.com/u/123456789",
	}
	
	fmt.Println("Retrieving user info from GitHub with token:", accessToken[:10]+"...")
	return user, nil
}

// getGoogleUser retrieves user information from Google API
func (c *OAuthController) getGoogleUser(accessToken string) (map[string]interface{}, error) {
	// In a real implementation, this would make an HTTP request to Google API
	// For example: GET https://www.googleapis.com/oauth2/v2/userinfo with Authorization: Bearer {token}
	
	// Mock implementation
	user := map[string]interface{}{
		"id":    "google_user_id",
		"name":  "Google User",
		"email": "googleuser@example.com",
		"avatar_url": "https://lh3.googleusercontent.com/a-/123456789",
	}
	
	fmt.Println("Retrieving user info from Google with token:", accessToken[:10]+"...")
	return user, nil
}

// generateJWT creates a JWT token for the user
func (c *OAuthController) generateJWT(userID, email string) (string, error) {
	// In a real implementation, this would create a JWT token
	// For now, we return a mock token
	return "mock-jwt-token", nil
}