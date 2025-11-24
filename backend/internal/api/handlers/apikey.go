package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// APIKeyHandler handles API key operations
type APIKeyHandler struct {
	db interface {
		Create(value interface{}) error
		Find(dest interface{}, conds ...interface{}) error
		First(dest interface{}, conds ...interface{}) error
		Save(value interface{}) error
		Delete(value interface{}, conds ...interface{}) error
		Exec(sql string, values ...interface{}) error
	}
}

// NewAPIKeyHandler creates a new API key handler
func NewAPIKeyHandler(db interface{}) *APIKeyHandler {
	return &APIKeyHandler{db: db}
}

// CreateAPIKey creates a new API key
// POST /api/v1/apikeys
func (h *APIKeyHandler) CreateAPIKey(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	// Parse request
	var req struct {
		Name        string     `json:"name" validate:"required,min=3,max=100"`
		Permissions []string   `json:"permissions"`
		ExpiresAt   *time.Time `json:"expires_at"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Generate API key
	keyBytes := make([]byte, 32)
	// In production, use crypto/rand
	key := "cta_" + uuid.New().String() // Simplified for now

	// Create API key record
	apiKey := map[string]interface{}{
		"id":          uuid.New().String(),
		"user_id":     userID,
		"name":        req.Name,
		"key":         key,
		"key_prefix":  key[:8],
		"permissions": req.Permissions,
		"expires_at":  req.ExpiresAt,
		"created_at":  time.Now(),
		"updated_at":  time.Now(),
	}

	// Save to database
	err := h.db.Exec(`
		INSERT INTO api_keys (id, user_id, name, key, key_prefix, permissions, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, apiKey["id"], apiKey["user_id"], apiKey["name"], apiKey["key"],
		apiKey["key_prefix"], apiKey["permissions"], apiKey["expires_at"],
		apiKey["created_at"], apiKey["updated_at"])

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create API key",
		})
	}

	// Return response with full key (only time it's shown)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":          apiKey["id"],
		"name":        apiKey["name"],
		"key":         key, // Full key only shown on creation
		"key_prefix":  apiKey["key_prefix"],
		"permissions": apiKey["permissions"],
		"expires_at":  apiKey["expires_at"],
		"created_at":  apiKey["created_at"],
	})
}

// ListAPIKeys lists all API keys for the current user
// GET /api/v1/apikeys
func (h *APIKeyHandler) ListAPIKeys(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var keys []map[string]interface{}
	err := h.db.Exec(`
		SELECT id, name, key_prefix, permissions, expires_at, last_used_at, created_at
		FROM api_keys
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, userID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch API keys",
		})
	}

	return c.JSON(fiber.Map{
		"keys": keys,
	})
}

// GetAPIKey gets a specific API key
// GET /api/v1/apikeys/:id
func (h *APIKeyHandler) GetAPIKey(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	keyID := c.Params("id")

	var key map[string]interface{}
	err := h.db.Exec(`
		SELECT id, name, key_prefix, permissions, expires_at, last_used_at, created_at
		FROM api_keys
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, keyID, userID)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "API key not found",
		})
	}

	return c.JSON(key)
}

// RevokeAPIKey revokes (soft deletes) an API key
// DELETE /api/v1/apikeys/:id
func (h *APIKeyHandler) RevokeAPIKey(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	keyID := c.Params("id")

	err := h.db.Exec(`
		UPDATE api_keys
		SET deleted_at = $1
		WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
	`, time.Now(), keyID, userID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to revoke API key",
		})
	}

	return c.JSON(fiber.Map{
		"message": "API key revoked successfully",
	})
}

// RotateAPIKey rotates an API key (creates new, revokes old)
// PUT /api/v1/apikeys/:id/rotate
func (h *APIKeyHandler) RotateAPIKey(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	keyID := c.Params("id")

	// Get old key details
	var oldKey struct {
		Name        string
		Permissions []string
		ExpiresAt   *time.Time
	}

	err := h.db.Exec(`
		SELECT name, permissions, expires_at
		FROM api_keys
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, keyID, userID)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "API key not found",
		})
	}

	// Generate new key
	newKey := "cta_" + uuid.New().String()
	newKeyID := uuid.New().String()

	// Create new key
	err = h.db.Exec(`
		INSERT INTO api_keys (id, user_id, name, key, key_prefix, permissions, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, newKeyID, userID, oldKey.Name+" (rotated)", newKey, newKey[:8],
		oldKey.Permissions, oldKey.ExpiresAt, time.Now(), time.Now())

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create new API key",
		})
	}

	// Revoke old key
	h.db.Exec(`
		UPDATE api_keys SET deleted_at = $1 WHERE id = $2
	`, time.Now(), keyID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":         newKeyID,
		"name":       oldKey.Name + " (rotated)",
		"key":        newKey, // Full key only shown on creation
		"key_prefix": newKey[:8],
		"created_at": time.Now(),
	})
}
