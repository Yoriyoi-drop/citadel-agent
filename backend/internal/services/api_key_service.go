// backend/internal/services/api_key_service.go
package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/citadel-agent/backend/internal/models"
	"github.com/citadel-agent/backend/internal/repositories"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// APIKeyService handles API key management operations
type APIKeyService struct {
	db           *pgxpool.Pool
	apiKeyRepo   *repositories.APIKeyRepository
	userRepo     *repositories.UserRepository
	teamRepo     *repositories.TeamRepository
	permissions  []string
}

// APIKey represents an API key with its metadata
type APIKey struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	UserID      *string   `json:"user_id,omitempty"`
	TeamID      *string   `json:"team_id,omitempty"`
	KeyPrefix   string    `json:"key_prefix"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	CreatedBy   string    `json:"created_by"`
	Revoked     bool      `json:"revoked"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewAPIKeyService creates a new API key service
func NewAPIKeyService(
	db *pgxpool.Pool,
	apiKeyRepo *repositories.APIKeyRepository,
	userRepo *repositories.UserRepository,
	teamRepo *repositories.TeamRepository,
) *APIKeyService {
	return &APIKeyService{
		db:         db,
		apiKeyRepo: apiKeyRepo,
		userRepo:   userRepo,
		teamRepo:   teamRepo,
		permissions: []string{
			"workflows:read",
			"workflows:write",
			"workflows:execute",
			"executions:read",
			"executions:write",
			"executions:control",
			"api_keys:read",
			"api_keys:write",
			"users:read",
			"teams:read",
		},
	}
}

// CreateAPIKey creates a new API key
func (s *APIKeyService) CreateAPIKey(ctx context.Context, userID, teamID *string, name string, permissions []string, expiresAt *time.Time, createdBy string) (*APIKey, error) {
	// Validate inputs
	if name == "" {
		return nil, fmt.Errorf("API key name is required")
	}

	if len(permissions) == 0 {
		return nil, fmt.Errorf("at least one permission is required")
	}

	// Validate permissions
	for _, perm := range permissions {
		if !s.isValidPermission(perm) {
			return nil, fmt.Errorf("invalid permission: %s", perm)
		}
	}

	// Generate API key
	apiKeyID := uuid.New().String()
	secretKey, err := s.generateSecureKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// Create key prefix (first 8 characters of the key)
	keyPrefix := secretKey[:8]

	// Hash the secret key for storage
	keyHash := s.hashKey(secretKey)

	// Create the API key in the database
	dbAPIKey := &models.APIKey{
		ID:          apiKeyID,
		Name:        name,
		UserID:      userID,
		TeamID:      teamID,
		KeyHash:     keyHash,
		KeyPrefix:   keyPrefix,
		Permissions: permissions,
		CreatedAt:   time.Now(),
		ExpiresAt:   expiresAt,
		CreatedBy:   createdBy,
		Revoked:     false,
	}

	createdAPIKey, err := s.apiKeyRepo.Create(ctx, dbAPIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create API key in database: %w", err)
	}

	// Create return APIKey object with the actual key for the caller
	returnKey := &APIKey{
		ID:          createdAPIKey.ID,
		Name:        createdAPIKey.Name,
		UserID:      createdAPIKey.UserID,
		TeamID:      createdAPIKey.TeamID,
		KeyPrefix:   createdAPIKey.KeyPrefix,
		Permissions: createdAPIKey.Permissions,
		CreatedAt:   createdAPIKey.CreatedAt,
		ExpiresAt:   createdAPIKey.ExpiresAt,
		LastUsedAt:  createdAPIKey.LastUsedAt,
		CreatedBy:   createdAPIKey.CreatedBy,
		Revoked:     createdAPIKey.Revoked,
	}

	// Add the actual key to the result (this is the only place where the full key is returned)
	returnKeyWithSecret := &APIKey{
		ID:          returnKey.ID,
		Name:        returnKey.Name,
		UserID:      returnKey.UserID,
		TeamID:      returnKey.TeamID,
		KeyPrefix:   returnKey.KeyPrefix,
		Permissions: returnKey.Permissions,
		CreatedAt:   returnKey.CreatedAt,
		ExpiresAt:   returnKey.ExpiresAt,
		LastUsedAt:  returnKey.LastUsedAt,
		CreatedBy:   returnKey.CreatedBy,
		Revoked:     returnKey.Revoked,
		Metadata:    map[string]interface{}{"key": secretKey}, // Include the key in metadata for initial return
	}

	return returnKeyWithSecret, nil
}

// ValidateAPIKey validates an API key and returns its details
func (s *APIKeyService) ValidateAPIKey(ctx context.Context, key string) (*APIKey, error) {
	if len(key) < 8 {
		return nil, fmt.Errorf("invalid key format")
	}

	// Extract the prefix from the key (first 8 characters)
	keyPrefix := key[:8]

	// Get the API key from the database by prefix
	dbAPIKey, err := s.apiKeyRepo.GetByPrefix(ctx, keyPrefix)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid API key")
		}
		return nil, fmt.Errorf("failed to retrieve API key: %w", err)
	}

	// Check if the key is revoked
	if dbAPIKey.Revoked {
		return nil, fmt.Errorf("API key has been revoked")
	}

	// Check if the key has expired
	if dbAPIKey.ExpiresAt != nil && time.Now().After(*dbAPIKey.ExpiresAt) {
		return nil, fmt.Errorf("API key has expired")
	}

	// Hash the provided key and compare with stored hash
	providedKeyHash := s.hashKey(key)
	if providedKeyHash != dbAPIKey.KeyHash {
		return nil, fmt.Errorf("invalid API key")
	}

	// Update the last used time
	now := time.Now()
	dbAPIKey.LastUsedAt = &now
	_, err = s.apiKeyRepo.Update(ctx, dbAPIKey)
	if err != nil {
		// Log the error but don't fail the validation
		// The API call can proceed even if we can't update the last used time
		fmt.Printf("Warning: failed to update last used time for API key %s: %v\n", dbAPIKey.ID, err)
	}

	// Return the API key details without the secret
	return &APIKey{
		ID:          dbAPIKey.ID,
		Name:        dbAPIKey.Name,
		UserID:      dbAPIKey.UserID,
		TeamID:      dbAPIKey.TeamID,
		KeyPrefix:   dbAPIKey.KeyPrefix,
		Permissions: dbAPIKey.Permissions,
		CreatedAt:   dbAPIKey.CreatedAt,
		ExpiresAt:   dbAPIKey.ExpiresAt,
		LastUsedAt:  dbAPIKey.LastUsedAt,
		CreatedBy:   dbAPIKey.CreatedBy,
		Revoked:     dbAPIKey.Revoked,
	}, nil
}

// GetUserAPIKeys retrieves all API keys for a user
func (s *APIKeyService) GetUserAPIKeys(ctx context.Context, userID string) ([]*APIKey, error) {
	dbAPIKeys, err := s.apiKeyRepo.GetByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user API keys: %w", err)
	}

	apiKeys := make([]*APIKey, len(dbAPIKeys))
	for i, dbKey := range dbAPIKeys {
		apiKeys[i] = &APIKey{
			ID:          dbKey.ID,
			Name:        dbKey.Name,
			UserID:      dbKey.UserID,
			TeamID:      dbKey.TeamID,
			KeyPrefix:   dbKey.KeyPrefix,
			Permissions: dbKey.Permissions,
			CreatedAt:   dbKey.CreatedAt,
			ExpiresAt:   dbKey.ExpiresAt,
			LastUsedAt:  dbKey.LastUsedAt,
			CreatedBy:   dbKey.CreatedBy,
			Revoked:     dbKey.Revoked,
		}
	}

	return apiKeys, nil
}

// GetTeamAPIKeys retrieves all API keys for a team
func (s *APIKeyService) GetTeamAPIKeys(ctx context.Context, teamID string) ([]*APIKey, error) {
	dbAPIKeys, err := s.apiKeyRepo.GetByTeam(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve team API keys: %w", err)
	}

	apiKeys := make([]*APIKey, len(dbAPIKeys))
	for i, dbKey := range dbAPIKeys {
		apiKeys[i] = &APIKey{
			ID:          dbKey.ID,
			Name:        dbKey.Name,
			UserID:      dbKey.UserID,
			TeamID:      dbKey.TeamID,
			KeyPrefix:   dbKey.KeyPrefix,
			Permissions: dbKey.Permissions,
			CreatedAt:   dbKey.CreatedAt,
			ExpiresAt:   dbKey.ExpiresAt,
			LastUsedAt:  dbKey.LastUsedAt,
			CreatedBy:   dbKey.CreatedBy,
			Revoked:     dbKey.Revoked,
		}
	}

	return apiKeys, nil
}

// RevokeAPIKey revokes an API key
func (s *APIKeyService) RevokeAPIKey(ctx context.Context, apiKeyID, revokedBy string) error {
	dbAPIKey, err := s.apiKeyRepo.GetByID(ctx, apiKeyID)
	if err != nil {
		return fmt.Errorf("failed to retrieve API key: %w", err)
	}

	dbAPIKey.Revoked = true
	_, err = s.apiKeyRepo.Update(ctx, dbAPIKey)
	if err != nil {
		return fmt.Errorf("failed to revoke API key: %w", err)
	}

	return nil
}

// RotateAPIKey creates a new API key with the same properties and revokes the old one
func (s *APIKeyService) RotateAPIKey(ctx context.Context, apiKeyID, rotatedBy string) (*APIKey, error) {
	dbAPIKey, err := s.apiKeyRepo.GetByID(ctx, apiKeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve API key: %w", err)
	}

	// Create a new API key with the same properties
	newKey, err := s.CreateAPIKey(
		ctx,
		dbAPIKey.UserID,
		dbAPIKey.TeamID,
		dbAPIKey.Name+" (rotated)",
		dbAPIKey.Permissions,
		dbAPIKey.ExpiresAt,
		rotatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create new API key: %w", err)
	}

	// Revoke the old key
	err = s.RevokeAPIKey(ctx, apiKeyID, rotatedBy)
	if err != nil {
		// If rotation fails after creating the new key, we should handle this
		// In a real implementation, we'd want to rollback
		return nil, fmt.Errorf("failed to revoke old API key after rotation: %w", err)
	}

	return newKey, nil
}

// CheckPermission checks if an API key has a specific permission
func (s *APIKeyService) CheckPermission(ctx context.Context, apiKey *APIKey, permission string) bool {
	// Admin keys (with wildcard permission) have all permissions
	for _, perm := range apiKey.Permissions {
		if perm == "*:*" {
			return true
		}
	}

	// Check for exact match or wildcard match
	for _, perm := range apiKey.Permissions {
		if s.permissionMatches(perm, permission) {
			return true
		}
	}

	return false
}

// permissionMatches checks if a key permission matches the requested permission
func (s *APIKeyService) permissionMatches(keyPermission, requestedPermission string) bool {
	// Exact match
	if keyPermission == requestedPermission {
		return true
	}

	// Check for wildcard in key permission
	if strings.Contains(keyPermission, ":*") {
		keyResource := strings.Split(keyPermission, ":*")[0]
		requestedResource := strings.Split(requestedPermission, ":")[0]

		if keyResource == requestedResource {
			return true
		}
	}

	return false
}

// generateSecureKey generates a cryptographically secure API key
func (s *APIKeyService) generateSecureKey() (string, error) {
	// Generate 32 bytes of random data
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Encode as hex string (64 characters)
	return hex.EncodeToString(bytes), nil
}

// hashKey creates a SHA-256 hash of the API key for secure storage
func (s *APIKeyService) hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// isValidPermission checks if a permission is valid
func (s *APIKeyService) isValidPermission(permission string) bool {
	// Check if it's a wildcard permission
	if permission == "*:*" {
		return true
	}

	// Check if it matches the resource:action format
	parts := strings.Split(permission, ":")
	if len(parts) != 2 {
		return false
	}

	// Check if the permission exists in our allowed list
	for _, allowedPerm := range s.permissions {
		if s.permissionMatches(allowedPerm, permission) {
			return true
		}
	}

	// Allow any permission that follows the correct format and resource exists
	resource := parts[0]
	validResources := []string{
		"workflows", "executions", "api_keys", "users", "teams", 
		"settings", "notifications", "monitoring", "logs",
	}

	for _, validResource := range validResources {
		if resource == validResource {
			return true
		}
	}

	return false
}

// ListAPIKeys returns a paginated list of API keys with optional filters
func (s *APIKeyService) ListAPIKeys(ctx context.Context, userID, teamID *string, revoked *bool, page, limit int) ([]*APIKey, error) {
	filters := repositories.APIKeyFilters{
		UserID:  userID,
		TeamID:  teamID,
		Revoked: revoked,
		Page:    page,
		Limit:   limit,
	}

	dbAPIKeys, err := s.apiKeyRepo.List(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	apiKeys := make([]*APIKey, len(dbAPIKeys))
	for i, dbKey := range dbAPIKeys {
		apiKeys[i] = &APIKey{
			ID:          dbKey.ID,
			Name:        dbKey.Name,
			UserID:      dbKey.UserID,
			TeamID:      dbKey.TeamID,
			KeyPrefix:   dbKey.KeyPrefix,
			Permissions: dbKey.Permissions,
			CreatedAt:   dbKey.CreatedAt,
			ExpiresAt:   dbKey.ExpiresAt,
			LastUsedAt:  dbKey.LastUsedAt,
			CreatedBy:   dbKey.CreatedBy,
			Revoked:     dbKey.Revoked,
		}
	}

	return apiKeys, nil
}