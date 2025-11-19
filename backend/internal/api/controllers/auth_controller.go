package controllers

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"citadel-agent/backend/internal/models"
	"citadel-agent/backend/internal/services"
)

// AuthController handles authentication-related HTTP requests
type AuthController struct {
	service *services.UserService
}

// NewAuthController creates a new auth controller
func NewAuthController(userService *services.UserService) *AuthController {
	return &AuthController{
		service: userService,
	}
}

// Login handles user login
func (c *AuthController) Login(ctx *fiber.Ctx) error {
	var loginReq struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	if err := ctx.BodyParser(&loginReq); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// Authenticate the user
	user, err := c.service.AuthenticateUser(loginReq.Email, loginReq.Password)
	if err != nil {
		log.Printf("Authentication failed for user %s: %v", loginReq.Email, err)
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	// Generate JWT token
	token, err := c.service.GenerateUserToken(user, "default_secret_for_dev", 86400)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot generate token",
		})
	}

	return ctx.JSON(fiber.Map{
		"token": token,
		"user":  user,
	})
}

// Register handles user registration
func (c *AuthController) Register(ctx *fiber.Ctx) error {
	var registerReq struct {
		Email     string `json:"email" validate:"required,email"`
		Password  string `json:"password" validate:"required"`
		Username  string `json:"username" validate:"required"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	if err := ctx.BodyParser(&registerReq); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// Create user model
	user := &models.User{
		Email:     registerReq.Email,
		Username:  registerReq.Username,
		Password:  registerReq.Password, // Service will hash the password
		FirstName: registerReq.FirstName,
		LastName:  registerReq.LastName,
	}

	// Create the user
	if err := c.service.CreateUser(user); err != nil {
		log.Printf("Error creating user %s: %v", registerReq.Email, err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot create user",
		})
	}

	// Authenticate the user after registration
	authUser, err := c.service.AuthenticateUser(registerReq.Email, registerReq.Password)
	if err != nil {
		log.Printf("Error authenticating new user %s: %v", registerReq.Email, err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "User created but authentication failed",
		})
	}

	// Generate JWT token
	token, err := c.service.GenerateUserToken(authUser, "default_secret_for_dev", 86400)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot generate token",
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"token": token,
		"user":  authUser,
	})
}

// Profile returns the authenticated user's profile
func (c *AuthController) Profile(ctx *fiber.Ctx) error {
	// In a real implementation, we would extract user info from the token context
	// For now, we'll return a dummy user
	// You would typically use a middleware to extract the user ID from the token
	userID, ok := ctx.Locals("user_id").(string)
	if !ok {
		// If user_id is not in context, we can't determine the user
		// This means the authentication middleware wasn't applied
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	user, err := c.service.GetUser(userID)
	if err != nil {
		log.Printf("Error getting user profile %s: %v", userID, err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot retrieve user profile",
		})
	}

	return ctx.JSON(user)
}