package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// Simple config
var (
	jwtSecret = getEnv("JWT_SECRET", "default_secret_for_dev")
	
	// OAuth configs
	githubClientID     = getEnv("GITHUB_CLIENT_ID", "")
	githubClientSecret = getEnv("GITHUB_CLIENT_SECRET", "")
	githubRedirectURI  = getEnv("GITHUB_REDIRECT_URI", "http://localhost:5001/auth/github/callback")
	
	googleClientID     = getEnv("GOOGLE_CLIENT_ID", "")
	googleClientSecret = getEnv("GOOGLE_CLIENT_SECRET", "")
	googleRedirectURI  = getEnv("GOOGLE_REDIRECT_URI", "http://localhost:5001/auth/google/callback")
	
	databaseURL = getEnv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/citadel_lite")
)

// Simple user structure
type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Username     string `json:"username"`
	Provider     string `json:"provider"` // github, google, local
	ProviderID   string `json:"provider_id"`
	AvatarURL    string `json:"avatar_url"`
	CreatedAt    int64  `json:"created_at"`
	LastLoginAt  int64  `json:"last_login_at"`
}

// Simple token structure
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

func main() {
	// Create Fiber app with custom error handler
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Log the error
			log.Printf("Error: %v at path: %s", err, c.Path())

			// Return appropriate error response
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
				"path":  c.Path(),
			})
		},
	})

	// Middleware
	app.Use(recover.New()) // Recover from panics
	app.Use(logger.New())  // Log requests
	app.Use(cors.New())    // Enable CORS

	// Database connection
	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to database")

	// Routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Citadel Agent Lite - Simplified Workflow Engine")
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		// Test database connection
		if err := db.Ping(context.Background()); err != nil {
			log.Printf("Health check failed: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"status": "unhealthy",
				"time":   time.Now().Unix(),
				"error":  "Database connection failed",
			})
		}

		return c.JSON(fiber.Map{
			"status": "healthy",
			"time":   time.Now().Unix(),
			"uptime": time.Now().Unix(),
		})
	})

	// Auth routes
	setupAuthRoutes(app, db)

	// 404 handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{
			"error": "Route not found",
			"path":  c.Path(),
		})
	})

	// Start server
	port := getEnv("PORT", "5001")
	log.Printf("Starting Citadel Agent Lite on port %s", port)

	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupAuthRoutes(app *fiber.App, db *pgxpool.Pool) {
	// Local login
	app.Post("/auth/login", func(c *fiber.Ctx) error {
		log.Printf("Login attempt from IP: %s", c.IP())

		// For simplicity, this is a mock implementation
		// In a real app, you'd verify credentials against DB
		var req struct {
			Email    string `json:"email" validate:"required,email"`
			Password string `json:"password" validate:"required"`
		}

		if err := c.BodyParser(&req); err != nil {
			log.Printf("Invalid login request from %s: %v", c.IP(), err)
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request format",
				"code":  "INVALID_REQUEST",
			})
		}

		// Validate email format
		if req.Email == "" || req.Password == "" {
			log.Printf("Missing credentials from %s", c.IP())
			return c.Status(400).JSON(fiber.Map{
				"error": "Email and password are required",
				"code":  "MISSING_CREDENTIALS",
			})
		}

		// Mock user creation/verification
		user := User{
			ID:        "user_" + req.Email,
			Email:     req.Email,
			Username:  req.Email,
			Provider:  "local",
			CreatedAt: time.Now().Unix(),
			LastLoginAt: time.Now().Unix(),
		}

		// Generate simple token (in a real app, use JWT)
		token := fmt.Sprintf("token_%s_%d", user.ID, time.Now().Unix())

		log.Printf("Successful login for user: %s from IP: %s", req.Email, c.IP())

		return c.JSON(fiber.Map{
			"access_token": token,
			"user":         user,
			"message":      "Login successful",
		})
	})

	// GitHub OAuth
	app.Get("/auth/github", func(c *fiber.Ctx) error {
		if githubClientID == "" {
			log.Printf("GitHub OAuth not configured, request from: %s", c.IP())
			return c.Status(500).JSON(fiber.Map{
				"error": "GitHub OAuth not configured",
				"code":  "OAUTH_NOT_CONFIGURED",
			})
		}

		// Generate a random state to prevent CSRF
		state := fmt.Sprintf("state_%d", time.Now().Unix())

		config := &oauth2.Config{
			ClientID:     githubClientID,
			ClientSecret: githubClientSecret,
			RedirectURL:  githubRedirectURI,
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
		}

		url := config.AuthCodeURL(state, oauth2.AccessTypeOnline)

		log.Printf("Initiating GitHub OAuth for IP: %s, state: %s", c.IP(), state)
		return c.Redirect(url)
	})

	// GitHub callback
	app.Get("/auth/github/callback", func(c *fiber.Ctx) error {
		if githubClientID == "" {
			log.Printf("GitHub OAuth not configured, callback from: %s", c.IP())
			return c.Status(500).JSON(fiber.Map{
				"error": "GitHub OAuth not configured",
				"code":  "OAUTH_NOT_CONFIGURED",
			})
		}

		code := c.Query("code")
		if code == "" {
			log.Printf("Missing authorization code in GitHub callback from: %s", c.IP())
			return c.Status(400).JSON(fiber.Map{
				"error": "No authorization code provided",
				"code":  "MISSING_CODE",
			})
		}

		// Exchange code for token
		config := &oauth2.Config{
			ClientID:     githubClientID,
			ClientSecret: githubClientSecret,
			RedirectURL:  githubRedirectURI,
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
		}

		token, err := config.Exchange(context.Background(), code)
		if err != nil {
			log.Printf("Failed to exchange GitHub code for token from %s: %v", c.IP(), err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to exchange authorization code",
				"code":  "TOKEN_EXCHANGE_FAILED",
			})
		}

		// In a real app, you'd get user profile from GitHub API here
		username := os.Getenv("GITHUB_DEFAULT_USERNAME")
		if username == "" {
			username = generateRandomString(8, "github") // Generate random string as default
		}

		email := os.Getenv("GITHUB_DEFAULT_EMAIL")
		if email == "" {
			email = generateRandomString(8, "github") + "@example.com" // Generate random email as default
		}

		user := User{
			ID:        "github_user_" + token.AccessToken[:8],
			Email:     email, // In real app, get from GitHub API
			Username:  username,
			Provider:  "github",
			CreatedAt: time.Now().Unix(),
			LastLoginAt: time.Now().Unix(),
		}

		// Generate token
		accessToken := fmt.Sprintf("token_%s_%d", user.ID, time.Now().Unix())

		log.Printf("Successful GitHub OAuth for user: %s, IP: %s", user.Email, c.IP())

		// In real app, save user to database
		// saveUserToDB(db, user)

		return c.JSON(fiber.Map{
			"access_token": accessToken,
			"user":         user,
			"message":      "GitHub login successful",
		})
	})

	// Google OAuth
	app.Get("/auth/google", func(c *fiber.Ctx) error {
		if googleClientID == "" {
			log.Printf("Google OAuth not configured, request from: %s", c.IP())
			return c.Status(500).JSON(fiber.Map{
				"error": "Google OAuth not configured",
				"code":  "OAUTH_NOT_CONFIGURED",
			})
		}

		// Generate a random state to prevent CSRF
		state := fmt.Sprintf("state_%d", time.Now().Unix())

		config := &oauth2.Config{
			ClientID:     googleClientID,
			ClientSecret: googleClientSecret,
			RedirectURL:  googleRedirectURI,
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
			Endpoint:     google.Endpoint,
		}

		url := config.AuthCodeURL(state, oauth2.AccessTypeOnline)

		log.Printf("Initiating Google OAuth for IP: %s, state: %s", c.IP(), state)
		return c.Redirect(url)
	})

	// Google callback
	app.Get("/auth/google/callback", func(c *fiber.Ctx) error {
		if googleClientID == "" {
			log.Printf("Google OAuth not configured, callback from: %s", c.IP())
			return c.Status(500).JSON(fiber.Map{
				"error": "Google OAuth not configured",
				"code":  "OAUTH_NOT_CONFIGURED",
			})
		}

		code := c.Query("code")
		if code == "" {
			log.Printf("Missing authorization code in Google callback from: %s", c.IP())
			return c.Status(400).JSON(fiber.Map{
				"error": "No authorization code provided",
				"code":  "MISSING_CODE",
			})
		}

		// Exchange code for token
		config := &oauth2.Config{
			ClientID:     googleClientID,
			ClientSecret: googleClientSecret,
			RedirectURL:  googleRedirectURI,
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
			Endpoint:     google.Endpoint,
		}

		token, err := config.Exchange(context.Background(), code)
		if err != nil {
			log.Printf("Failed to exchange Google code for token from %s: %v", c.IP(), err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to exchange authorization code",
				"code":  "TOKEN_EXCHANGE_FAILED",
			})
		}

		// In a real app, you'd get user profile from Google API here
		username := os.Getenv("GOOGLE_DEFAULT_USERNAME")
		if username == "" {
			username = generateRandomString(8, "google") // Generate random string as default
		}

		email := os.Getenv("GOOGLE_DEFAULT_EMAIL")
		if email == "" {
			email = generateRandomString(8, "google") + "@example.com" // Generate random email as default
		}

		user := User{
			ID:        "google_user_" + token.AccessToken[:8],
			Email:     email, // In real app, get from Google API
			Username:  username,
			Provider:  "google",
			CreatedAt: time.Now().Unix(),
			LastLoginAt: time.Now().Unix(),
		}

		// Generate token
		accessToken := fmt.Sprintf("token_%s_%d", user.ID, time.Now().Unix())

		log.Printf("Successful Google OAuth for user: %s, IP: %s", user.Email, c.IP())

		// In real app, save user to database
		// saveUserToDB(db, user)

		return c.JSON(fiber.Map{
			"access_token": accessToken,
			"user":         user,
			"message":      "Google login successful",
		})
	})

	// Protected route example
	app.Get("/auth/me", func(c *fiber.Ctx) error {
		// In a real app, validate JWT token here
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			log.Printf("Unauthorized access attempt to /auth/me from: %s", c.IP())
			return c.Status(401).JSON(fiber.Map{
				"error": "Authorization header required",
				"code":  "UNAUTHORIZED",
			})
		}

		// Check if the token is in the right format
		if len(authHeader) < 7 || authHeader[:6] != "Bearer" {
			log.Printf("Invalid authorization header format from: %s", c.IP())
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid authorization header format",
				"code":  "INVALID_AUTH_FORMAT",
			})
		}

		token := authHeader[7:] // Remove "Bearer " prefix
		if token == "" {
			log.Printf("Empty token in authorization header from: %s", c.IP())
			return c.Status(401).JSON(fiber.Map{
				"error": "Empty token in authorization header",
				"code":  "EMPTY_TOKEN",
			})
		}

		// Mock user return - in a real app, verify the JWT token
		user := User{
			ID:        "current_user",
			Email:     "user@example.com",
			Username:  "CurrentUser",
			CreatedAt: time.Now().Unix() - 86400, // 1 day ago
			LastLoginAt: time.Now().Unix(),
		}

		log.Printf("Successful access to /auth/me for token: %s... from IP: %s", token[:min(10, len(token))], c.IP())

		return c.JSON(fiber.Map{
			"user":    user,
			"message": "User info retrieved successfully",
		})
	})
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper function to get environment variable with default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// generateRandomString generates a random string with a prefix
func generateRandomString(length int, prefix string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = "abcdefghijklmnopqrstuvwxyz0123456789"[time.Now().UnixNano()%int64(len("abcdefghijklmnopqrstuvwxyz0123456789"))]
	}
	if prefix != "" {
		return prefix + "_" + string(b)
	}
	return string(b)
}