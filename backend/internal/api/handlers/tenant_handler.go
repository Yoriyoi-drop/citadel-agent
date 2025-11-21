// backend/internal/api/handlers/tenant_handler.go
package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"citadel-agent/backend/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// TenantHandler handles tenant-related HTTP requests
type TenantHandler struct {
	tenantService *services.TenantService
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(tenantService *services.TenantService) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
	}
}

// CreateTenant creates a new tenant
func (th *TenantHandler) CreateTenant(c *fiber.Ctx) error {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		OwnerID     string `json:"owner_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Tenant name is required",
		})
	}

	if req.OwnerID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Owner ID is required",
		})
	}

	// Validate UUID format for owner ID
	if _, err := uuid.Parse(req.OwnerID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid owner ID format",
		})
	}

	tenant, err := th.tenantService.CreateTenant(c.Context(), req.Name, req.Description, req.OwnerID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create tenant: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    tenant,
	})
}

// GetTenant retrieves a tenant by ID
func (th *TenantHandler) GetTenant(c *fiber.Ctx) error {
	tenantID := c.Params("tenantId")

	// Validate UUID format
	if _, err := uuid.Parse(tenantID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid tenant ID format",
		})
	}

	tenant, err := th.tenantService.GetTenant(c.Context(), tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get tenant: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    tenant,
	})
}

// UpdateTenant updates an existing tenant
func (th *TenantHandler) UpdateTenant(c *fiber.Ctx) error {
	tenantID := c.Params("tenantId")

	// Validate UUID format
	if _, err := uuid.Parse(tenantID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid tenant ID format",
		})
	}

	var req struct {
		Name        *string `json:"name,omitempty"`
		Description *string `json:"description,omitempty"`
		Status      *string `json:"status,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var status *services.TenantStatus
	if req.Status != nil {
		s := services.TenantStatus(*req.Status)
		status = &s
	}

	tenant, err := th.tenantService.UpdateTenant(c.Context(), tenantID, 
		getStringOrEmpty(req.Name), 
		getStringOrEmpty(req.Description), 
		status)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to update tenant: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    tenant,
	})
}

// DeleteTenant deletes a tenant
func (th *TenantHandler) DeleteTenant(c *fiber.Ctx) error {
	tenantID := c.Params("tenantId")

	// Validate UUID format
	if _, err := uuid.Parse(tenantID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid tenant ID format",
		})
	}

	err := th.tenantService.DeleteTenant(c.Context(), tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to delete tenant: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Tenant deleted successfully",
	})
}

// ListTenants retrieves a list of tenants
func (th *TenantHandler) ListTenants(c *fiber.Ctx) error {
	statusStr := c.Query("status")
	limitStr := c.Query("limit", "50")
	offsetStr := c.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	var status *services.TenantStatus
	if statusStr != "" {
		s := services.TenantStatus(statusStr)
		status = &s
	}

	tenants, err := th.tenantService.ListTenants(c.Context(), status, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to list tenants: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    tenants,
		"count":   len(tenants),
		"limit":   limit,
		"offset":  offset,
	})
}

// GetTenantByUser retrieves the tenant for a specific user
func (th *TenantHandler) GetTenantByUser(c *fiber.Ctx) error {
	userID := c.Params("userId")

	// Validate UUID format
	if _, err := uuid.Parse(userID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	tenant, err := th.tenantService.GetTenantByUser(c.Context(), userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get tenant for user: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    tenant,
	})
}

// InviteUser invites a user to a tenant
func (th *TenantHandler) InviteUser(c *fiber.Ctx) error {
	tenantID := c.Params("tenantId")

	// Validate UUID format
	if _, err := uuid.Parse(tenantID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid tenant ID format",
		})
	}

	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Email == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Email is required",
		})
	}

	if req.Role == "" {
		req.Role = "member" // Default role
	}

	err := th.tenantService.InviteUser(c.Context(), tenantID, req.Email, req.Role)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to invite user: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "User invited successfully",
	})
}

// RemoveUser removes a user from a tenant
func (th *TenantHandler) RemoveUser(c *fiber.Ctx) error {
	tenantID := c.Params("tenantId")
	userID := c.Params("userId")

	// Validate UUID format
	if _, err := uuid.Parse(tenantID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid tenant ID format",
		})
	}

	if _, err := uuid.Parse(userID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	err := th.tenantService.RemoveUser(c.Context(), tenantID, userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to remove user: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "User removed successfully",
	})
}

// GetTenantUsers retrieves all users in a tenant
func (th *TenantHandler) GetTenantUsers(c *fiber.Ctx) error {
	tenantID := c.Params("tenantId")

	// Validate UUID format
	if _, err := uuid.Parse(tenantID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid tenant ID format",
		})
	}

	limitStr := c.Query("limit", "50")
	offsetStr := c.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	users, err := th.tenantService.GetTenantUsers(c.Context(), tenantID, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get tenant users: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    users,
		"count":   len(users),
		"limit":   limit,
		"offset":  offset,
	})
}

// UpdateTenantSettings updates tenant settings
func (th *TenantHandler) UpdateTenantSettings(c *fiber.Ctx) error {
	tenantID := c.Params("tenantId")

	// Validate UUID format
	if _, err := uuid.Parse(tenantID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid tenant ID format",
		})
	}

	var settings map[string]interface{}
	if err := c.BodyParser(&settings); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := th.tenantService.UpdateTenantSettings(c.Context(), tenantID, settings)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to update tenant settings: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Tenant settings updated successfully",
	})
}

// GetTenantSettings retrieves tenant settings
func (th *TenantHandler) GetTenantSettings(c *fiber.Ctx) error {
	tenantID := c.Params("tenantId")

	// Validate UUID format
	if _, err := uuid.Parse(tenantID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid tenant ID format",
		})
	}

	settings, err := th.tenantService.GetTenantSettings(c.Context(), tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get tenant settings: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    settings,
	})
}

// SwitchTenant allows a user to switch between tenants
func (th *TenantHandler) SwitchTenant(c *fiber.Ctx) error {
	userID := c.Locals("user_id") // Assuming this comes from auth middleware
	if userID == nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return c.Status(500).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	tenantID := c.Params("tenantId")

	// Validate UUID format
	if _, err := uuid.Parse(tenantID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid tenant ID format",
		})
	}

	err := th.tenantService.SwitchTenant(c.Context(), userIDStr, tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to switch tenant: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Tenant switched successfully",
		"data": fiber.Map{
			"tenant_id": tenantID,
			"user_id":   userIDStr,
		},
	})
}

// ValidateTenantAccess checks if a user has access to a tenant
func (th *TenantHandler) ValidateTenantAccess(c *fiber.Ctx) error {
	userID := c.Locals("user_id") // Assuming this comes from auth middleware
	if userID == nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return c.Status(500).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	tenantID := c.Params("tenantId")

	// Validate UUID format
	if _, err := uuid.Parse(tenantID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid tenant ID format",
		})
	}

	hasAccess, err := th.tenantService.ValidateTenantAccess(c.Context(), userIDStr, tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to validate tenant access: %v", err),
		})
	}

	if !hasAccess {
		return c.Status(403).JSON(fiber.Map{
			"error": "User does not have access to this tenant",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Access validated successfully",
		"data": fiber.Map{
			"tenant_id": tenantID,
			"user_id":   userIDStr,
			"has_access": hasAccess,
		},
	})
}

// getStringOrEmpty returns the string value or empty string if nil
func getStringOrEmpty(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// RegisterRoutes registers tenant handler routes
func (th *TenantHandler) RegisterRoutes(router fiber.Router) {
	// Authenticated endpoints
	authGroup := router.Use() // This would use the auth middleware in a real implementation
	{
		// Tenant management
		authGroup.Post("/tenants", th.CreateTenant)
		authGroup.Get("/tenants/:tenantId", th.GetTenant)
		authGroup.Put("/tenants/:tenantId", th.UpdateTenant)
		authGroup.Delete("/tenants/:tenantId", th.DeleteTenant)
		authGroup.Get("/tenants", th.ListTenants)

		// User-tenant relationship
		authGroup.Get("/users/:userId/tenant", th.GetTenantByUser)
		authGroup.Post("/tenants/:tenantId/invite", th.InviteUser)
		authGroup.Delete("/tenants/:tenantId/users/:userId", th.RemoveUser)
		authGroup.Get("/tenants/:tenantId/users", th.GetTenantUsers)

		// Settings management
		authGroup.Put("/tenants/:tenantId/settings", th.UpdateTenantSettings)
		authGroup.Get("/tenants/:tenantId/settings", th.GetTenantSettings)

		// Tenant switching and access control
		authGroup.Post("/tenants/:tenantId/switch", th.SwitchTenant)
		authGroup.Get("/tenants/:tenantId/access/validate", th.ValidateTenantAccess)
	}
}