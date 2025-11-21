package auth

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// RegisterFiberRoutes registers authentication routes with Fiber
func (s *AuthService) RegisterFiberRoutes(app *fiber.App) {
	// Public routes
	app.Post("/auth/login", s.LoginHandler)
	app.Get("/auth/logout", s.LogoutHandler)
	app.Get("/auth/oauth/github", s.GithubLoginHandler)
	app.Get("/auth/oauth/github/callback", s.GithubCallbackHandler)
	app.Get("/auth/oauth/google", s.GoogleLoginHandler)
	app.Get("/auth/oauth/google/callback", s.GoogleCallbackHandler)
	app.Post("/auth/device", s.DeviceCodeInitHandler)
	app.Post("/auth/device/verify", s.DeviceCodeVerifyHandler)
	app.Post("/auth/token/refresh", s.RefreshTokenHandler)

	// Protected routes
	app.Get("/auth/me", s.AuthMiddleware(), s.MeHandler)
}

// AuthMiddleware provides authentication middleware for protected routes
func (s *AuthService) AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from header or cookie
		authHeader := c.Get("Authorization")
		var tokenString string

		if authHeader != "" {
			// Check if it's a Bearer token
			if len(authHeader) >= 7 && strings.ToUpper(authHeader[:6]) == "BEARER" {
				tokenString = authHeader[7:]
			} else {
				return c.Status(401).JSON(fiber.Map{
					"error": "Invalid authorization header format",
				})
			}
		} else {
			// Try to get token from cookie
			cookie := c.Cookies("access_token")
			if cookie == "" {
				return c.Status(401).JSON(fiber.Map{
					"error": "Authorization token required",
				})
			}
			tokenString = cookie
		}

		// Validate JWT token
		claims, err := s.validateToken(tokenString)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// Store user info in context for use by next handlers
		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("username", claims.Username)

		return c.Next()
	}
}

// LoginHandler handles local email/password login
func (s *AuthService) LoginHandler(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// In a real application, you would validate the credentials here
	// For this example, we'll simulate a successful login with mock user data
	user := &User{
		ID:        s.generateRandomString(16),
		Email:     req.Email,
		Username:  req.Email, // In a real app, this would be fetched from DB
		Provider:  "local",
		ProviderID: "",
		AvatarURL: "",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := s.generateJWT(user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to generate tokens",
		})
	}

	// Update last login
	s.updateLastLogin(user.ID)

	// Set secure cookie with access token (only if needed)
	cookie := fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   3600, // 1 hour
		HTTPOnly: true,
		Secure:   s.isSecureCookie(), // Set to true in production with HTTPS
		SameSite: "Strict",
	}
	c.Cookie(&cookie)

	return c.JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    3600,
		"token_type":    "Bearer",
		"user": fiber.Map{
			"id":       user.ID,
			"email":    user.Email,
			"username": user.Username,
		},
	})
}

// LogoutHandler handles user logout
func (s *AuthService) LogoutHandler(c *fiber.Ctx) error {
	// Clear the access token cookie
	c.ClearCookie("access_token")

	// In production, invalidate refresh token in database

	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

// GithubLoginHandler redirects to GitHub OAuth
func (s *AuthService) GithubLoginHandler(c *fiber.Ctx) error {
	state := s.GenerateState()
	url := s.oauth[GitHub].AuthCodeURL(state, oauth2.AccessTypeOnline)

	// Store state in session or database (simplified for this example)
	// In production, use secure session management

	return c.Redirect(url)
}

// GithubCallbackHandler handles GitHub OAuth callback
func (s *AuthService) GithubCallbackHandler(c *fiber.Ctx) error {
	// Get the authorization code from the callback
	code := c.Query("code")
	if code == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "No authorization code provided",
		})
	}

	// Exchange code for token
	token, err := s.oauth[GitHub].Exchange(c.Context(), code)
	if err != nil {
		log.Printf("Failed to exchange code for token: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to exchange code for token",
		})
	}

	// Get user info from GitHub API
	user, err := s.getGithubUser(token.AccessToken)
	if err != nil {
		log.Printf("Failed to get GitHub user: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get user info from GitHub",
		})
	}

	// Create or update user in database
	dbUser, err := s.createOrUpdateUser(user, string(GitHub))
	if err != nil {
		log.Printf("Failed to create/update user: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := s.generateJWT(dbUser)
	if err != nil {
		log.Printf("Failed to generate JWT: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to generate tokens",
		})
	}

	// Update last login
	s.updateLastLogin(dbUser.ID)

	// Set secure cookie with access token
	cookie := fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   3600, // 1 hour
		HTTPOnly: true,
		Secure:   s.isSecureCookie(), // Set to true in production with HTTPS
		SameSite: "Strict",
	}
	c.Cookie(&cookie)

	// Redirect to frontend with success (or return token in JSON for API)
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	
	return c.Redirect(frontendURL + "/auth/success?token=" + accessToken)
}

// GoogleLoginHandler redirects to Google OAuth
func (s *AuthService) GoogleLoginHandler(c *fiber.Ctx) error {
	state := s.GenerateState()
	url := s.oauth[Google].AuthCodeURL(state, oauth2.AccessTypeOnline)

	return c.Redirect(url)
}

// GoogleCallbackHandler handles Google OAuth callback
func (s *AuthService) GoogleCallbackHandler(c *fiber.Ctx) error {
	// Get the authorization code from the callback
	code := c.Query("code")
	if code == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "No authorization code provided",
		})
	}

	// Exchange code for token
	token, err := s.oauth[Google].Exchange(c.Context(), code)
	if err != nil {
		log.Printf("Failed to exchange code for token: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to exchange code for token",
		})
	}

	// Get user info from Google API
	user, err := s.getGoogleUser(token.AccessToken)
	if err != nil {
		log.Printf("Failed to get Google user: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get user info from Google",
		})
	}

	// Create or update user in database
	dbUser, err := s.createOrUpdateUser(user, string(Google))
	if err != nil {
		log.Printf("Failed to create/update user: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := s.generateJWT(dbUser)
	if err != nil {
		log.Printf("Failed to generate JWT: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to generate tokens",
		})
	}

	// Update last login
	s.updateLastLogin(dbUser.ID)

	// Set secure cookie with access token
	cookie := fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   3600, // 1 hour
		HTTPOnly: true,
		Secure:   s.isSecureCookie(), // Set to true in production with HTTPS
		SameSite: "Strict",
	}
	c.Cookie(&cookie)

	// Redirect to frontend with success
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	
	return c.Redirect(frontendURL + "/auth/success?token=" + accessToken)
}

// DeviceCodeInitHandler handles device code initiation
func (s *AuthService) DeviceCodeInitHandler(c *fiber.Ctx) error {
	var req DeviceCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate provider
	if req.Provider != GitHub && req.Provider != Google {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid provider",
		})
	}

	// Generate codes
	deviceCode := s.generateRandomString(32)
	userCode := s.generateUserCode()

	// Create device session
	session := &DeviceSession{
		DeviceCode: deviceCode,
		UserCode:   userCode,
		Provider:   string(req.Provider),
		ExpiresAt:  time.Now().Add(10 * time.Minute), // 10 minutes expiry
		Status:     "pending",
	}

	// Store session (in production, use Redis or database)
	s.deviceCode[deviceCode] = session

	// Return device code and user instructions
	response := DeviceCodeResponse{
		UserCode:        userCode,
		DeviceCode:      deviceCode,
		VerificationURI: "https://github.com/login/device", // Use provider-specific URL
		ExpiresIn:       600, // 10 minutes
		Interval:        5,   // Poll every 5 seconds
	}

	return c.JSON(response)
}

// DeviceCodeVerifyHandler handles device code verification
func (s *AuthService) DeviceCodeVerifyHandler(c *fiber.Ctx) error {
	var req DeviceVerifyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get stored session
	session, exists := s.deviceCode[req.DeviceCode]
	if !exists {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid device code",
		})
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		session.Status = "expired"
		delete(s.deviceCode, req.DeviceCode)
		return c.Status(400).JSON(fiber.Map{
			"error": "Device code expired",
		})
	}

	// Check if already approved
	if session.Status == "approved" {
		// Return tokens
		response := TokenResponse{
			AccessToken:  session.AccessToken,
			RefreshToken: session.RefreshToken,
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}
		
		return c.JSON(response)
	}

	// Return pending status
	return c.Status(202).JSON(fiber.Map{
		"status":  "pending",
		"message": "Waiting for user to approve",
	})
}

// RefreshTokenHandler handles JWT refresh token
func (s *AuthService) RefreshTokenHandler(c *fiber.Ctx) error {
	// Extract refresh token from authorization header or request body
	authHeader := c.Get("Authorization")
	var refreshToken string

	if authHeader != "" {
		if len(authHeader) >= 7 && strings.ToUpper(authHeader[:6]) == "BEARER" {
			refreshToken = authHeader[7:]
		} else {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}
	} else {
		// Try to get refresh token from request body
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Refresh token required in request body or authorization header",
			})
		}
		refreshToken = req.RefreshToken
	}

	if refreshToken == "" {
		return c.Status(401).JSON(fiber.Map{
			"error": "Refresh token required",
		})
	}

	// In a real implementation, validate refresh token and generate new access token
	// Check if refresh token exists in database and is not revoked

	// For this example, we return an error as the implementation requires database storage
	return c.Status(501).JSON(fiber.Map{
		"error": "Refresh token functionality not implemented in this example",
	})
}

// MeHandler returns authenticated user info
func (s *AuthService) MeHandler(c *fiber.Ctx) error {
	// Get user info from context (set by AuthMiddleware)
	userID := c.Locals("user_id")
	email := c.Locals("email")
	username := c.Locals("username")

	if userID == nil || email == nil || username == nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Return user info
	user := fiber.Map{
		"id":       userID,
		"email":    email,
		"username": username,
		"provider": "local", // This would be determined from the user record in a real app
	}

	return c.JSON(user)
}

// isSecureCookie returns true if cookies should be secure (in production)
func (s *AuthService) isSecureCookie() bool {
	environment := os.Getenv("ENVIRONMENT")
	return environment == "production"
}

// validateToken validates JWT token and returns claims
func (s *AuthService) validateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}