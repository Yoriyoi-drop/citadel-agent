// backend/internal/auth/secrets.go
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SecretType represents the type of secret
type SecretType string

const (
	APIKeySecret       SecretType = "api_key"
	DatabaseCredential SecretType = "db_credential"
	ExternalServiceKey SecretType = "ext_service_key"
	Certificate        SecretType = "certificate"
)

// Secret represents a stored secret
type Secret struct {
	ID        uint         `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	
	// Metadata
	Name        string     `json:"name" gorm:"not null"`
	Description string     `json:"description"`
	Type        SecretType `json:"type" gorm:"not null"`
	
	// References
	OwnerID   *uint `json:"owner_id,omitempty" gorm:"index"` // User who owns the secret
	OwnerType string `json:"owner_type"` // "user", "workflow", "system", etc.
	
	// Encrypted Value
	Value           string    `json:"-" gorm:"column:value"` // Encrypted field
	ValueHash       string    `json:"-" gorm:"column:value_hash"` // For comparison without decrypting
	ValueType       string    `json:"value_type"` // "text", "json", "file", etc.
	
	// Metadata
	Tags            []string  `json:"tags" gorm:"-"` // Not stored in DB, managed separately
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	RotatedAt       *time.Time `json:"rotated_at,omitempty"`
	CreatedBy       *uint     `json:"created_by,omitempty"` // User who created the secret
	LastAccessedAt  *time.Time `json:"last_accessed_at,omitempty"`
	
	// Rotation settings
	RotationInterval *time.Duration `json:"rotation_interval,omitempty" gorm:"-"`
	AutoRotate       bool          `json:"auto_rotate" gorm:"default:false"`
	
	// Status
	IsActive   bool   `json:"is_active" gorm:"default:true"`
	IsRevoked  bool   `json:"is_revoked" gorm:"default:false"`
	Version    int    `json:"version" gorm:"default:1"`
	
	// References to related entities
	WorkflowID *uint `json:"workflow_id,omitempty" gorm:"index"`
	Service    string `json:"service"` // The service this secret is for
}

// SecretManager provides secret management functionality
type SecretManager struct {
	DB        *gorm.DB
	Encryptor *EncryptionService
}

// NewSecretManager creates a new secret manager
func NewSecretManager(db *gorm.DB, encryptor *EncryptionService) *SecretManager {
	return &SecretManager{
		DB:        db,
		Encryptor: encryptor,
	}
}

// CreateSecret creates a new secret with encryption
func (sm *SecretManager) CreateSecret(ctx context.Context, secret *Secret) error {
	if secret.Name == "" {
		return fmt.Errorf("secret name is required")
	}
	
	if secret.Type == "" {
		return fmt.Errorf("secret type is required")
	}

	// Generate a unique ID for the secret
	secretID := uuid.New().String()
	
	// Encrypt the secret value
	encryptedValue, err := sm.Encryptor.EncryptString(secret.Value)
	if err != nil {
		return fmt.Errorf("failed to encrypt secret value: %w", err)
	}
	
	// Store encrypted value and hash
	secret.Value = encryptedValue
	
	// Create a hash of the original value for comparison
	hash, err := sm.hashSecretValue(secret.Value)
	if err != nil {
		return fmt.Errorf("failed to hash secret value: %w", err)
	}
	secret.ValueHash = hash
	
	// Set creation timestamp
	secret.CreatedAt = time.Now()
	secret.UpdatedAt = time.Now()
	
	// Create the secret in the database
	if err := sm.DB.Create(secret).Error; err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}
	
	return nil
}

// GetSecret retrieves a secret by ID (with decryption)
func (sm *SecretManager) GetSecret(ctx context.Context, id uint, userID *uint) (*Secret, error) {
	var secret Secret
	if err := sm.DB.First(&secret, id).Error; err != nil {
		return nil, fmt.Errorf("secret not found: %w", err)
	}

	// Check if user has permission to access this secret
	if userID != nil && secret.OwnerID != nil && *secret.OwnerID != 0 && *secret.OwnerID != *userID {
		return nil, fmt.Errorf("user does not have permission to access this secret")
	}

	if !secret.IsActive || secret.IsRevoked {
		return nil, fmt.Errorf("secret is inactive or revoked")
	}

	// Decrypt the value for return
	decryptedValue, err := sm.Encryptor.DecryptString(secret.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secret value: %w", err)
	}
	
	// Create a copy to return with decrypted value
	result := secret
	result.Value = decryptedValue
	
	// Update last accessed time
	go sm.updateLastAccessed(secret.ID) // Update asynchronously
	
	return &result, nil
}

// ListSecrets retrieves secrets for a user or owner
func (sm *SecretManager) ListSecrets(ctx context.Context, userID *uint, ownerType string, filters map[string]interface{}) ([]Secret, error) {
	var secrets []Secret
	query := sm.DB.Model(&Secret{})

	// Filter by user
	if userID != nil {
		query = query.Where("owner_id = ? OR owner_id IS NULL", *userID)
	}
	
	// Apply other filters
	for key, value := range filters {
		switch key {
		case "type":
			query = query.Where("type = ?", value)
		case "service":
			query = query.Where("service = ?", value)
		case "active":
			query = query.Where("is_active = ?", value)
		case "owner_type":
			query = query.Where("owner_type = ?", value)
		}
	}

	if err := query.Find(&secrets).Error; err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	return secrets, nil
}

// UpdateSecret updates an existing secret
func (sm *SecretManager) UpdateSecret(ctx context.Context, secretID uint, userID *uint, updates map[string]interface{}) error {
	// Get the existing secret to check permissions
	existingSecret, err := sm.GetSecret(ctx, secretID, userID)
	if err != nil {
		return err
	}

	// Check permissions
	if userID != nil && existingSecret.OwnerID != nil && *existingSecret.OwnerID != *userID {
		return fmt.Errorf("user does not have permission to update this secret")
	}

	// Prepare updates
	dbUpdates := make(map[string]interface{})
	for key, value := range updates {
		switch key {
		case "name", "description", "service", "expires_at", "auto_rotate":
			dbUpdates[key] = value
		case "value":
			// If value is being updated, encrypt it
			encryptedValue, err := sm.Encryptor.EncryptString(value.(string))
			if err != nil {
				return fmt.Errorf("failed to encrypt secret value: %w", err)
			}
			
			hash, err := sm.hashSecretValue(value.(string))
			if err != nil {
				return fmt.Errorf("failed to hash secret value: %w", err)
			}
			
			dbUpdates["value"] = encryptedValue
			dbUpdates["value_hash"] = hash
			dbUpdates["rotated_at"] = time.Now()
			dbUpdates["version"] = gorm.Expr("version + 1")
		}
	}

	// Add last updated timestamp
	dbUpdates["updated_at"] = time.Now()

	if err := sm.DB.Model(&Secret{}).Where("id = ?", secretID).Updates(dbUpdates).Error; err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}

	return nil
}

// DeleteSecret soft deletes a secret (mark as revoked)
func (sm *SecretManager) DeleteSecret(ctx context.Context, secretID uint, userID *uint) error {
	// Get the existing secret to check permissions
	var existingSecret Secret
	if err := sm.DB.First(&existingSecret, secretID).Error; err != nil {
		return fmt.Errorf("secret not found: %w", err)
	}

	// Check permissions
	if userID != nil && existingSecret.OwnerID != nil && *existingSecret.OwnerID != *userID {
		return fmt.Errorf("user does not have permission to delete this secret")
	}

	// Soft delete by marking as revoked
	updates := map[string]interface{}{
		"is_revoked": true,
		"is_active":  false,
		"updated_at": time.Now(),
	}

	if err := sm.DB.Model(&Secret{}).Where("id = ?", secretID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	return nil
}

// RotateSecret rotates a secret's value
func (sm *SecretManager) RotateSecret(ctx context.Context, secretID uint, userID *uint, newValue string) error {
	// Get the existing secret to check permissions
	var existingSecret Secret
	if err := sm.DB.First(&existingSecret, secretID).Error; err != nil {
		return fmt.Errorf("secret not found: %w", err)
	}

	// Check permissions
	if userID != nil && existingSecret.OwnerID != nil && *existingSecret.OwnerID != *userID {
		return fmt.Errorf("user does not have permission to rotate this secret")
	}

	// Encrypt the new value
	encryptedValue, err := sm.Encryptor.EncryptString(newValue)
	if err != nil {
		return fmt.Errorf("failed to encrypt new secret value: %w", err)
	}

	// Create hash of new value
	hash, err := sm.hashSecretValue(newValue)
	if err != nil {
		return fmt.Errorf("failed to hash new secret value: %w", err)
	}

	// Update the secret with new encrypted value
	updates := map[string]interface{}{
		"value":       encryptedValue,
		"value_hash":  hash,
		"rotated_at":  time.Now(),
		"version":     gorm.Expr("version + 1"),
		"updated_at":  time.Now(),
	}

	if err := sm.DB.Model(&Secret{}).Where("id = ?", secretID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to rotate secret: %w", err)
	}

	return nil
}

// ValidateSecret checks if a secret is valid and not expired
func (sm *SecretManager) ValidateSecret(ctx context.Context, secretID uint, expectedValue string) (bool, error) {
	var secret Secret
	if err := sm.DB.Select("id, value_hash, is_active, is_revoked, expires_at").First(&secret, secretID).Error; err != nil {
		return false, fmt.Errorf("secret not found: %w", err)
	}

	// Check if secret is active and not revoked
	if !secret.IsActive || secret.IsRevoked {
		return false, nil
	}

	// Check expiration
	if secret.ExpiresAt != nil && time.Now().After(*secret.ExpiresAt) {
		return false, nil
	}

	// Hash the expected value and compare with stored hash
	expectedHash, err := sm.hashSecretValue(expectedValue)
	if err != nil {
		return false, fmt.Errorf("failed to hash expected value: %w", err)
	}

	return secret.ValueHash == expectedHash, nil
}

// hashSecretValue creates a hash of the secret value
func (sm *SecretManager) hashSecretValue(value string) (string, error) {
	// In a real implementation, you'd use a proper hashing function
	// For now, we'll use a placeholder (use bcrypt or scrypt in production)
	return value, nil // Placeholder - implement proper hashing
}

// updateLastAccessed updates the last accessed timestamp
func (sm *SecretManager) updateLastAccessed(secretID uint) {
	// Update asynchronously to not block the main operation
	now := time.Now()
	sm.DB.Model(&Secret{}).Where("id = ?", secretID).Update("last_accessed_at", now)
}

// CleanupExpiredSecrets removes or disables expired secrets
func (sm *SecretManager) CleanupExpiredSecrets(ctx context.Context) (int64, error) {
	now := time.Now()
	
	// Find expired secrets that are still active
	result := sm.DB.Model(&Secret{}).
		Where("expires_at IS NOT NULL AND expires_at < ? AND is_active = ?", now, true).
		Updates(map[string]interface{}{
			"is_active": false,
			"updated_at": now,
		})
	
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup expired secrets: %w", result.Error)
	}
	
	return result.RowsAffected, nil
}

// ScheduleAutomaticRotation schedules automatic rotation of secrets
func (sm *SecretManager) ScheduleAutomaticRotation(ctx context.Context) error {
	// Find secrets that need rotation
	var secretsToRotate []Secret
	now := time.Now()
	
	err := sm.DB.Where("auto_rotate = ? AND rotation_interval IS NOT NULL", true).
		Where("rotated_at IS NULL OR rotated_at + rotation_interval < ?", now).
		Find(&secretsToRotate).Error
	
	if err != nil {
		return fmt.Errorf("failed to get secrets for rotation: %w", err)
	}
	
	// Rotate each secret
	for _, secret := range secretsToRotate {
		// Generate new value (in a real implementation, this would be more sophisticated)
		newValue := generateNewSecretValue()
		
		if err := sm.RotateSecret(ctx, secret.ID, nil, newValue); err != nil {
			// Log error but continue with other secrets
			continue
		}
	}
	
	return nil
}

// generateNewSecretValue generates a new secret value (placeholder implementation)
func generateNewSecretValue() string {
	// In a real implementation, you'd generate a cryptographically secure random value
	// For now, return a placeholder
	return uuid.New().String()
}