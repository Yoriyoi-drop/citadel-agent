package handlers

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// AuditLogHandler handles audit log operations
type AuditLogHandler struct {
	db *gorm.DB
}

// NewAuditLogHandler creates a new audit log handler
func NewAuditLogHandler(db *gorm.DB) *AuditLogHandler {
	return &AuditLogHandler{db: db}
}

// ListAuditLogs lists audit logs with filtering
// GET /api/v1/audit-logs
func (h *AuditLogHandler) ListAuditLogs(c *fiber.Ctx) error {
	// Parse query parameters
	userID := c.Query("user_id")
	action := c.Query("action")
	resource := c.Query("resource")
	status := c.Query("status")
	limit := c.QueryInt("limit", 100)
	offset := c.QueryInt("offset", 0)

	// Build query
	query := `
		SELECT id, user_id, action, resource, resource_id,
		       changes, ip_address, user_agent, status, error_msg, created_at
		FROM audit_logs
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	if userID != "" {
		query += " AND user_id = $" + string(rune(argCount))
		args = append(args, userID)
		argCount++
	}

	if action != "" {
		query += " AND action = $" + string(rune(argCount))
		args = append(args, action)
		argCount++
	}

	if resource != "" {
		query += " AND resource = $" + string(rune(argCount))
		args = append(args, resource)
		argCount++
	}

	if status != "" {
		query += " AND status = $" + string(rune(argCount))
		args = append(args, status)
		argCount++
	}

	query += " ORDER BY created_at DESC LIMIT $" + string(rune(argCount)) + " OFFSET $" + string(rune(argCount+1))
	args = append(args, limit, offset)

	var logs []map[string]interface{}
	err := h.db.Exec(query, args...)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch audit logs",
		})
	}

	return c.JSON(fiber.Map{
		"logs":   logs,
		"limit":  limit,
		"offset": offset,
	})
}

// GetAuditLog gets a specific audit log
// GET /api/v1/audit-logs/:id
func (h *AuditLogHandler) GetAuditLog(c *fiber.Ctx) error {
	logID := c.Params("id")

	var log map[string]interface{}
	err := h.db.Exec(`
		SELECT id, user_id, action, resource, resource_id,
		       changes, ip_address, user_agent, status, error_msg, created_at
		FROM audit_logs
		WHERE id = $1
	`, logID)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Audit log not found",
		})
	}

	return c.JSON(log)
}

// ExportAuditLogs exports audit logs as CSV
// GET /api/v1/audit-logs/export
func (h *AuditLogHandler) ExportAuditLogs(c *fiber.Ctx) error {
	// Parse query parameters (same as ListAuditLogs)
	userID := c.Query("user_id")
	action := c.Query("action")
	resource := c.Query("resource")
	status := c.Query("status")

	// Build query
	query := `
		SELECT id, user_id, action, resource, resource_id,
		       ip_address, user_agent, status, created_at
		FROM audit_logs
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	if userID != "" {
		query += " AND user_id = $" + string(rune(argCount))
		args = append(args, userID)
		argCount++
	}

	if action != "" {
		query += " AND action = $" + string(rune(argCount))
		args = append(args, action)
		argCount++
	}

	if resource != "" {
		query += " AND resource = $" + string(rune(argCount))
		args = append(args, resource)
		argCount++
	}

	if status != "" {
		query += " AND status = $" + string(rune(argCount))
		args = append(args, status)
		argCount++
	}

	query += " ORDER BY created_at DESC"

	var logs []map[string]interface{}
	err := h.db.Exec(query, args...)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to export audit logs",
		})
	}

	// Convert to CSV
	csv := "ID,User ID,Action,Resource,Resource ID,IP Address,User Agent,Status,Created At\n"
	for _, log := range logs {
		csv += log["id"].(string) + "," +
			log["user_id"].(string) + "," +
			log["action"].(string) + "," +
			log["resource"].(string) + "," +
			log["resource_id"].(string) + "," +
			log["ip_address"].(string) + "," +
			log["user_agent"].(string) + "," +
			log["status"].(string) + "," +
			log["created_at"].(string) + "\n"
	}

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=audit_logs.csv")
	return c.SendString(csv)
}
