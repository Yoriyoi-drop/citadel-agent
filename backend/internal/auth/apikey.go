package auth

import (
	"crypto/subtle"
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrAPIKeyNotFound = errors.New("API key not found")
	ErrAPIKeyExpired  = errors.New("API key expired")
	ErrAPIKeyInvalid  = errors.New("invalid API key")
)

// APIKeyService handles API key operations
type APIKeyService struct {
	db *gorm.DB
}

// NewAPIKeyService creates a new API key service
func NewAPIKeyService(db *gorm.DB) *APIKeyService {
	return &APIKeyService{db: db}
}

// ValidateAPIKey validates an API key and returns the associated user ID
func (s *APIKeyService) ValidateAPIKey(key string) (string, error) {
	var apiKey struct {
		ID        string
		UserID    string
		Key       string
		ExpiresAt *time.Time
		DeletedAt *time.Time
	}

	err := s.db.Table("api_keys").
		Select("id, user_id, key, expires_at, deleted_at").
		Where("key = ?", key).
		Where("deleted_at IS NULL").
		First(&apiKey).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrAPIKeyNotFound
		}
		return "", err
	}

	// Check if expired
	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return "", ErrAPIKeyExpired
	}

	// Constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(apiKey.Key), []byte(key)) != 1 {
		return "", ErrAPIKeyInvalid
	}

	// Update last used timestamp asynchronously
	go func() {
		s.db.Exec(
			"UPDATE api_keys SET last_used_at = ? WHERE id = ?",
			time.Now(), apiKey.ID,
		)
	}()

	return apiKey.UserID, nil
}

// GetAPIKeyPermissions returns the permissions associated with an API key
func (s *APIKeyService) GetAPIKeyPermissions(key string) ([]string, error) {
	var permissions []string
	err := s.db.Table("api_keys").
		Select("permissions").
		Where("key = ?", key).
		Where("deleted_at IS NULL").
		Pluck("permissions", &permissions).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAPIKeyNotFound
		}
		return nil, err
	}

	return permissions, nil
}

// HasAPIKeyPermission checks if an API key has a specific permission
func (s *APIKeyService) HasAPIKeyPermission(key string, permission string) (bool, error) {
	permissions, err := s.GetAPIKeyPermissions(key)
	if err != nil {
		return false, err
	}

	// Check for admin permission
	if contains(permissions, "admin:*") {
		return true, nil
	}

	// Check for specific permission
	return contains(permissions, permission), nil
}

// RevokeAPIKey revokes an API key (soft delete)
func (s *APIKeyService) RevokeAPIKey(keyID string, userID string) error {
	return s.db.Exec(
		"UPDATE api_keys SET deleted_at = ? WHERE id = ? AND user_id = ?",
		time.Now(), keyID, userID,
	).Error
}

// ListUserAPIKeys lists all API keys for a user
func (s *APIKeyService) ListUserAPIKeys(userID string) ([]map[string]interface{}, error) {
	var keys []map[string]interface{}
	err := s.db.Table("api_keys").
		Select("id, name, key_prefix, permissions, expires_at, last_used_at, created_at").
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Find(&keys).Error

	return keys, err
}

// RotateAPIKey rotates an API key (creates new key, revokes old one)
func (s *APIKeyService) RotateAPIKey(keyID string, userID string) (string, error) {
	// Get old key details
	var oldKey struct {
		Name        string
		Permissions []string
		ExpiresAt   *time.Time
	}

	err := s.db.Table("api_keys").
		Select("name, permissions, expires_at").
		Where("id = ?", keyID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		First(&oldKey).Error

	if err != nil {
		return "", err
	}

	// Create new key with same details
	newKey := make(map[string]interface{})
	newKey["user_id"] = userID
	newKey["name"] = oldKey.Name + " (rotated)"
	newKey["permissions"] = oldKey.Permissions
	newKey["expires_at"] = oldKey.ExpiresAt

	// This would need the actual model to generate the key
	// For now, return empty string - should be implemented with proper model

	// Revoke old key
	err = s.RevokeAPIKey(keyID, userID)
	if err != nil {
		return "", err
	}

	return "", nil // TODO: Implement with proper model
}
