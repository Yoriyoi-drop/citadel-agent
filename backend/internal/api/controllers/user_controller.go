package controllers

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/citadel-agent/backend/internal/services"
)

// UserController handles user-related HTTP requests
type UserController struct {
	service *services.UserService
}

// NewUserController creates a new user controller
func NewUserController(userService *services.UserService) *UserController {
	return &UserController{
		service: userService,
	}
}

// GetUser retrieves a user by ID
func (c *UserController) GetUser(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	user, err := c.service.GetUser(id)
	if err != nil {
		log.Printf("Error getting user %s: %v", id, err)
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return ctx.JSON(user)
}

// GetUsers retrieves all users with pagination
func (c *UserController) GetUsers(ctx *fiber.Ctx) error {
	// Get query parameters for pagination
	page, err := strconv.Atoi(ctx.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.Query("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	// Convert pagination to offset and limit for the service
	offset := (page - 1) * limit

	users, err := c.service.GetAllUsersWithPagination(offset, limit)
	if err != nil {
		log.Printf("Error getting users: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot retrieve users",
		})
	}

	return ctx.JSON(users)
}

// UpdateUser updates a user
func (c *UserController) UpdateUser(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	user, err := c.service.GetUser(id)
	if err != nil {
		log.Printf("Error getting user %s: %v", id, err)
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	if err := ctx.BodyParser(user); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}
	// Keep the original ID
	user.ID = id

	if err := c.service.UpdateUser(user); err != nil {
		log.Printf("Error updating user %s: %v", id, err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot update user",
		})
	}

	return ctx.JSON(user)
}

// DeleteUser deletes a user by ID
func (c *UserController) DeleteUser(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	if err := c.service.DeleteUser(id); err != nil {
		log.Printf("Error deleting user %s: %v", id, err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot delete user",
		})
	}

	return ctx.Status(fiber.StatusNoContent).JSON(fiber.Map{
		"message": "User deleted successfully",
	})
}