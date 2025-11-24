package middleware

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
)

// AuditMiddleware logs all requests for audit purposes
type AuditMiddleware struct {
	db interface {
		Exec(query string, args ...interface{}) error
	}
}

// NewAuditMiddleware creates a new audit middleware
func NewAuditMiddleware(db interface{}) *AuditMiddleware {
	return &AuditMiddleware{db: db}
}

// Log creates a middleware that logs actions for audit
func (m *AuditMiddleware) Log(action string, resource string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context
		userID := c.Locals("userID")
		if userID == nil {
			userID = "anonymous"
		}

		// Store start time
		startTime := time.Now()

		// Continue with request
		err := c.Next()

		// Determine status
		status := "success"
		errorMsg := ""
		if err != nil {
			status = "failure"
			errorMsg = err.Error()
		} else if c.Response().StatusCode() >= 400 {
			status = "failure"
		}

		// Extract resource ID from params or body
		resourceID := c.Params("id")
		if resourceID == "" {
			resourceID = c.Query("id")
		}

		// Get request/response for changes
		changes := make(map[string]interface{})

		// Try to parse request body
		if len(c.Body()) > 0 {
			var body map[string]interface{}
			if err := json.Unmarshal(c.Body(), &body); err == nil {
				changes["request"] = body
			}
		}

		// Add response status
		changes["status_code"] = c.Response().StatusCode()
		changes["duration_ms"] = time.Since(startTime).Milliseconds()

		changesJSON, _ := json.Marshal(changes)

		// Log asynchronously to avoid blocking
		go func() {
			m.db.Exec(`
				INSERT INTO audit_logs (
					id, user_id, action, resource, resource_id,
					changes, ip_address, user_agent, status, error_msg, created_at
				) VALUES (
					gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, NOW()
				)
			`,
				userID, action, resource, resourceID,
				changesJSON, c.IP(), c.Get("User-Agent"), status, errorMsg,
			)
		}()

		return err
	}
}

// LogWorkflowAction logs workflow-specific actions
func (m *AuditMiddleware) LogWorkflowAction(action string) fiber.Handler {
	return m.Log(action, "workflow")
}

// LogUserAction logs user-specific actions
func (m *AuditMiddleware) LogUserAction(action string) fiber.Handler {
	return m.Log(action, "user")
}

// LogAPIKeyAction logs API key-specific actions
func (m *AuditMiddleware) LogAPIKeyAction(action string) fiber.Handler {
	return m.Log(action, "apikey")
}

// LogRoleAction logs role-specific actions
func (m *AuditMiddleware) LogRoleAction(action string) fiber.Handler {
	return m.Log(action, "role")
}
