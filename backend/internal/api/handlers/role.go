package handlers

import (
	"time"

	"github.com/citadel-agent/backend/internal/auth"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleHandler handles role operations
type RoleHandler struct {
	db          *gorm.DB
	rbacService *auth.RBACService
}

// NewRoleHandler creates a new role handler
func NewRoleHandler(db *gorm.DB, rbacService *auth.RBACService) *RoleHandler {
	return &RoleHandler{
		db:          db,
		rbacService: rbacService,
	}
}

// CreateRole creates a new role
// POST /api/v1/roles
func (h *RoleHandler) CreateRole(c *fiber.Ctx) error {
	// Parse request
	var req struct {
		Name        string   `json:"name" validate:"required,min=3,max=50"`
		Description string   `json:"description"`
		Permissions []string `json:"permissions" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate permissions
	if err := h.rbacService.ValidatePermissions(req.Permissions); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Create role
	roleID := uuid.New().String()
	err := h.db.Exec(`
		INSERT INTO roles (id, name, description, permissions, is_system, created_at, updated_at)
		VALUES ($1, $2, $3, $4, false, $5, $6)
	`, roleID, req.Name, req.Description, req.Permissions, time.Now(), time.Now())

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create role",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":          roleID,
		"name":        req.Name,
		"description": req.Description,
		"permissions": req.Permissions,
		"is_system":   false,
		"created_at":  time.Now(),
	})
}

// ListRoles lists all roles
// GET /api/v1/roles
func (h *RoleHandler) ListRoles(c *fiber.Ctx) error {
	var roles []map[string]interface{}

	err := h.db.Exec(`
		SELECT id, name, description, permissions, is_system, created_at
		FROM roles
		WHERE deleted_at IS NULL
		ORDER BY name ASC
	`)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch roles",
		})
	}

	return c.JSON(fiber.Map{
		"roles": roles,
	})
}

// GetRole gets a specific role
// GET /api/v1/roles/:id
func (h *RoleHandler) GetRole(c *fiber.Ctx) error {
	roleID := c.Params("id")

	var role map[string]interface{}
	err := h.db.Exec(`
		SELECT id, name, description, permissions, is_system, created_at
		FROM roles
		WHERE id = $1 AND deleted_at IS NULL
	`, roleID)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Role not found",
		})
	}

	return c.JSON(role)
}

// UpdateRole updates a role
// PUT /api/v1/roles/:id
func (h *RoleHandler) UpdateRole(c *fiber.Ctx) error {
	roleID := c.Params("id")

	// Parse request
	var req struct {
		Description string   `json:"description"`
		Permissions []string `json:"permissions"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check if role is system role
	var isSystem bool
	err := h.db.Exec(`
		SELECT is_system FROM roles WHERE id = $1 AND deleted_at IS NULL
	`, roleID)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Role not found",
		})
	}

	if isSystem {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Cannot modify system roles",
		})
	}

	// Validate permissions
	if len(req.Permissions) > 0 {
		if err := h.rbacService.ValidatePermissions(req.Permissions); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}

	// Update role
	err = h.db.Exec(`
		UPDATE roles
		SET description = $1, permissions = $2, updated_at = $3
		WHERE id = $4 AND deleted_at IS NULL
	`, req.Description, req.Permissions, time.Now(), roleID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update role",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Role updated successfully",
	})
}

// DeleteRole deletes a role
// DELETE /api/v1/roles/:id
func (h *RoleHandler) DeleteRole(c *fiber.Ctx) error {
	roleID := c.Params("id")

	// Check if role is system role
	var isSystem bool
	err := h.db.Exec(`
		SELECT is_system FROM roles WHERE id = $1 AND deleted_at IS NULL
	`, roleID)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Role not found",
		})
	}

	if isSystem {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Cannot delete system roles",
		})
	}

	// Soft delete role
	err = h.db.Exec(`
		UPDATE roles SET deleted_at = $1 WHERE id = $2
	`, time.Now(), roleID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete role",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Role deleted successfully",
	})
}

// AssignRole assigns a role to a user
// POST /api/v1/users/:userId/roles/:roleId
func (h *RoleHandler) AssignRole(c *fiber.Ctx) error {
	userID := c.Params("userId")
	roleID := c.Params("roleId")

	// Check if role exists
	var count int
	err := h.db.Exec(`
		SELECT COUNT(*) FROM roles WHERE id = $1 AND deleted_at IS NULL
	`, roleID)

	if err != nil || count == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Role not found",
		})
	}

	// Assign role
	err = h.db.Exec(`
		INSERT INTO user_roles (user_id, role_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`, userID, roleID, time.Now())

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to assign role",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Role assigned successfully",
	})
}

// RemoveRole removes a role from a user
// DELETE /api/v1/users/:userId/roles/:roleId
func (h *RoleHandler) RemoveRole(c *fiber.Ctx) error {
	userID := c.Params("userId")
	roleID := c.Params("roleId")

	err := h.db.Exec(`
		DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2
	`, userID, roleID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to remove role",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Role removed successfully",
	})
}

// GetUserRoles gets all roles for a user
// GET /api/v1/users/:userId/roles
func (h *RoleHandler) GetUserRoles(c *fiber.Ctx) error {
	userID := c.Params("userId")

	var roles []map[string]interface{}
	err := h.db.Exec(`
		SELECT r.id, r.name, r.description, r.permissions, r.is_system
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND r.deleted_at IS NULL
		ORDER BY r.name ASC
	`, userID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user roles",
		})
	}

	return c.JSON(fiber.Map{
		"roles": roles,
	})
}
