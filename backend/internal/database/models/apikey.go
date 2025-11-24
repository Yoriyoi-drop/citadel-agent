package models

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// APIKey represents an API key for programmatic access
type APIKey struct {
	ID          string         `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      string         `gorm:"type:uuid;index;not null" json:"user_id"`
	Name        string         `gorm:"not null" json:"name"`
	Key         string         `gorm:"uniqueIndex;not null" json:"-"` // Never expose in JSON
	KeyPrefix   string         `gorm:"index" json:"key_prefix"`       // First 8 chars for identification
	Permissions []string       `gorm:"type:text[]" json:"permissions"`
	ExpiresAt   *time.Time     `json:"expires_at"`
	LastUsedAt  *time.Time     `json:"last_used_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate hook to generate UUID and API key
func (a *APIKey) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}

	// Generate API key if not provided
	if a.Key == "" {
		key, err := GenerateAPIKey()
		if err != nil {
			return err
		}
		a.Key = key
		a.KeyPrefix = key[:8] // Store first 8 chars for display
	}

	return nil
}

// IsExpired checks if the API key has expired
func (a *APIKey) IsExpired() bool {
	if a.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*a.ExpiresAt)
}

// UpdateLastUsed updates the last used timestamp
func (a *APIKey) UpdateLastUsed(tx *gorm.DB) error {
	now := time.Now()
	a.LastUsedAt = &now
	return tx.Model(a).Update("last_used_at", now).Error
}

// GenerateAPIKey generates a secure random API key
func GenerateAPIKey() (string, error) {
	// Generate 32 random bytes
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// Encode to base64 and add prefix
	key := "cta_" + base64.URLEncoding.EncodeToString(b)
	return key, nil
}

// APIKeyCreateRequest represents the request to create an API key
type APIKeyCreateRequest struct {
	Name        string     `json:"name" validate:"required,min=3,max=100"`
	Permissions []string   `json:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

// APIKeyResponse represents the API key response (with full key only on creation)
type APIKeyResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Key         string     `json:"key,omitempty"` // Only included on creation
	KeyPrefix   string     `json:"key_prefix"`
	Permissions []string   `json:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// ToResponse converts APIKey to APIKeyResponse
func (a *APIKey) ToResponse(includeKey bool) APIKeyResponse {
	resp := APIKeyResponse{
		ID:          a.ID,
		Name:        a.Name,
		KeyPrefix:   a.KeyPrefix,
		Permissions: a.Permissions,
		ExpiresAt:   a.ExpiresAt,
		LastUsedAt:  a.LastUsedAt,
		CreatedAt:   a.CreatedAt,
	}

	if includeKey {
		resp.Key = a.Key
	}

	return resp
}
