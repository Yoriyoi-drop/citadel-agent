// backend/internal/repositories/api_key_repository.go
package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// APIKeyRepository handles API key-related database operations
type APIKeyRepository struct {
	db *pgxpool.Pool
}

// APIKey represents an API key in the database
type APIKey struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	UserID      *string   `json:"user_id,omitempty"`
	TeamID      *string   `json:"team_id,omitempty"`
	KeyHash     string    `json:"key_hash"`
	KeyPrefix   string    `json:"key_prefix"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	CreatedBy   string    `json:"created_by"`
	Revoked     bool      `json:"revoked"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// APIKeyFilters represents filters for listing API keys
type APIKeyFilters struct {
	UserID  *string
	TeamID  *string
	Revoked *bool
	Page    int
	Limit   int
}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository(db *pgxpool.Pool) *APIKeyRepository {
	return &APIKeyRepository{
		db: db,
	}
}

// Create creates a new API key
func (akr *APIKeyRepository) Create(ctx context.Context, apiKey *APIKey) (*APIKey, error) {
	// Serialize permissions to JSON
	permissionsJSON, err := json.Marshal(apiKey.Permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal permissions: %w", err)
	}

	// Serialize metadata to JSON
	metadataJSON, err := json.Marshal(apiKey.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO api_keys (
			id, name, user_id, team_id, key_hash, key_prefix, permissions,
			created_at, expires_at, last_used_at, created_by, revoked, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, name, user_id, team_id, key_prefix, created_at, expires_at, last_used_at, created_by, revoked
	`

	var createdKey APIKey
	var permissionsBytes []byte
	var metadataBytes []byte

	err = akr.db.QueryRow(ctx, query,
		apiKey.ID,
		apiKey.Name,
		apiKey.UserID,
		apiKey.TeamID,
		apiKey.KeyHash,
		apiKey.KeyPrefix,
		permissionsJSON,
		apiKey.CreatedAt,
		apiKey.ExpiresAt,
		apiKey.LastUsedAt,
		apiKey.CreatedBy,
		apiKey.Revoked,
		metadataJSON,
	).Scan(
		&createdKey.ID,
		&createdKey.Name,
		&createdKey.UserID,
		&createdKey.TeamID,
		&createdKey.KeyPrefix,
		&createdKey.CreatedAt,
		&createdKey.ExpiresAt,
		&createdKey.LastUsedAt,
		&createdKey.CreatedBy,
		&createdKey.Revoked,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	// Unmarshal permissions
	if err := json.Unmarshal(permissionsJSON, &createdKey.Permissions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
	}

	// Unmarshal metadata
	if err := json.Unmarshal(metadataJSON, &createdKey.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &createdKey, nil
}

// GetByID retrieves an API key by ID
func (akr *APIKeyRepository) GetByID(ctx context.Context, id string) (*APIKey, error) {
	query := `
		SELECT id, name, user_id, team_id, key_prefix, permissions, 
		       created_at, expires_at, last_used_at, created_by, revoked, metadata
		FROM api_keys
		WHERE id = $1
	`

	var apiKey APIKey
	var permissionsBytes []byte
	var metadataBytes []byte

	err := akr.db.QueryRow(ctx, query, id).Scan(
		&apiKey.ID,
		&apiKey.Name,
		&apiKey.UserID,
		&apiKey.TeamID,
		&apiKey.KeyPrefix,
		&permissionsBytes,
		&apiKey.CreatedAt,
		&apiKey.ExpiresAt,
		&apiKey.LastUsedAt,
		&apiKey.CreatedBy,
		&apiKey.Revoked,
		&metadataBytes,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("API key not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	// Unmarshal permissions
	if err := json.Unmarshal(permissionsBytes, &apiKey.Permissions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
	}

	// Unmarshal metadata
	if err := json.Unmarshal(metadataBytes, &apiKey.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &apiKey, nil
}

// GetByPrefix retrieves an API key by its prefix
func (akr *APIKeyRepository) GetByPrefix(ctx context.Context, prefix string) (*APIKey, error) {
	query := `
		SELECT id, name, user_id, team_id, key_prefix, permissions, 
		       created_at, expires_at, last_used_at, created_by, revoked, metadata
		FROM api_keys
		WHERE key_prefix = $1
	`

	var apiKey APIKey
	var permissionsBytes []byte
	var metadataBytes []byte

	err := akr.db.QueryRow(ctx, query, prefix).Scan(
		&apiKey.ID,
		&apiKey.Name,
		&apiKey.UserID,
		&apiKey.TeamID,
		&apiKey.KeyPrefix,
		&permissionsBytes,
		&apiKey.CreatedAt,
		&apiKey.ExpiresAt,
		&apiKey.LastUsedAt,
		&apiKey.CreatedBy,
		&apiKey.Revoked,
		&metadataBytes,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("API key not found with prefix: %s", prefix)
		}
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	// Unmarshal permissions
	if err := json.Unmarshal(permissionsBytes, &apiKey.Permissions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
	}

	// Unmarshal metadata
	if err := json.Unmarshal(metadataBytes, &apiKey.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &apiKey, nil
}

// GetByUser retrieves API keys for a specific user
func (akr *APIKeyRepository) GetByUser(ctx context.Context, userID string) ([]*APIKey, error) {
	query := `
		SELECT id, name, user_id, team_id, key_prefix, permissions, 
		       created_at, expires_at, last_used_at, created_by, revoked, metadata
		FROM api_keys
		WHERE user_id = $1 AND revoked = false
		ORDER BY created_at DESC
	`

	rows, err := akr.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query API keys: %w", err)
	}
	defer rows.Close()

	var apiKeys []*APIKey
	for rows.Next() {
		var apiKey APIKey
		var permissionsBytes []byte
		var metadataBytes []byte

		err := rows.Scan(
			&apiKey.ID,
			&apiKey.Name,
			&apiKey.UserID,
			&apiKey.TeamID,
			&apiKey.KeyPrefix,
			&permissionsBytes,
			&apiKey.CreatedAt,
			&apiKey.ExpiresAt,
			&apiKey.LastUsedAt,
			&apiKey.CreatedBy,
			&apiKey.Revoked,
			&metadataBytes,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}

		// Unmarshal permissions
		if err := json.Unmarshal(permissionsBytes, &apiKey.Permissions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
		}

		// Unmarshal metadata
		if err := json.Unmarshal(metadataBytes, &apiKey.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		apiKeys = append(apiKeys, &apiKey)
	}

	return apiKeys, nil
}

// GetByTeam retrieves API keys for a specific team
func (akr *APIKeyRepository) GetByTeam(ctx context.Context, teamID string) ([]*APIKey, error) {
	query := `
		SELECT id, name, user_id, team_id, key_prefix, permissions, 
		       created_at, expires_at, last_used_at, created_by, revoked, metadata
		FROM api_keys
		WHERE team_id = $1 AND revoked = false
		ORDER BY created_at DESC
	`

	rows, err := akr.db.Query(ctx, query, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to query API keys: %w", err)
	}
	defer rows.Close()

	var apiKeys []*APIKey
	for rows.Next() {
		var apiKey APIKey
		var permissionsBytes []byte
		var metadataBytes []byte

		err := rows.Scan(
			&apiKey.ID,
			&apiKey.Name,
			&apiKey.UserID,
			&apiKey.TeamID,
			&apiKey.KeyPrefix,
			&permissionsBytes,
			&apiKey.CreatedAt,
			&apiKey.ExpiresAt,
			&apiKey.LastUsedAt,
			&apiKey.CreatedBy,
			&apiKey.Revoked,
			&metadataBytes,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}

		// Unmarshal permissions
		if err := json.Unmarshal(permissionsBytes, &apiKey.Permissions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
		}

		// Unmarshal metadata
		if err := json.Unmarshal(metadataBytes, &apiKey.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		apiKeys = append(apiKeys, &apiKey)
	}

	return apiKeys, nil
}

// Update updates an existing API key
func (akr *APIKeyRepository) Update(ctx context.Context, apiKey *APIKey) (*APIKey, error) {
	// Serialize permissions to JSON
	permissionsJSON, err := json.Marshal(apiKey.Permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal permissions: %w", err)
	}

	// Serialize metadata to JSON
	metadataJSON, err := json.Marshal(apiKey.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE api_keys
		SET name = $2, permissions = $3, expires_at = $4, last_used_at = $5, 
		    revoked = $6, metadata = $7, updated_at = $8
		WHERE id = $1
		RETURNING id, name, user_id, team_id, key_prefix, created_at, expires_at, last_used_at, created_by, revoked
	`

	var updatedKey APIKey
	var permissionsBytes []byte
	var metadataBytes []byte

	err = akr.db.QueryRow(ctx, query,
		apiKey.ID,
		apiKey.Name,
		permissionsJSON,
		apiKey.ExpiresAt,
		apiKey.LastUsedAt,
		apiKey.Revoked,
		metadataJSON,
		time.Now(),
	).Scan(
		&updatedKey.ID,
		&updatedKey.Name,
		&updatedKey.UserID,
		&updatedKey.TeamID,
		&updatedKey.KeyPrefix,
		&updatedKey.CreatedAt,
		&updatedKey.ExpiresAt,
		&updatedKey.LastUsedAt,
		&updatedKey.CreatedBy,
		&updatedKey.Revoked,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update API key: %w", err)
	}

	// Unmarshal permissions
	if err := json.Unmarshal(permissionsJSON, &updatedKey.Permissions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
	}

	// Unmarshal metadata
	if err := json.Unmarshal(metadataJSON, &updatedKey.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &updatedKey, nil
}

// Revoke sets an API key as revoked
func (akr *APIKeyRepository) Revoke(ctx context.Context, id string) error {
	query := `
		UPDATE api_keys
		SET revoked = true, updated_at = $2
		WHERE id = $1
	`

	_, err := akr.db.Exec(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to revoke API key: %w", err)
	}

	return nil
}

// UpdateLastUsedAt updates the last used timestamp of an API key
func (akr *APIKeyRepository) UpdateLastUsedAt(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE api_keys
		SET last_used_at = $2, updated_at = $3
		WHERE id = $1
	`

	_, err := akr.db.Exec(ctx, query, id, &now, now)
	if err != nil {
		return fmt.Errorf("failed to update last used timestamp: %w", err)
	}

	return nil
}

// List retrieves a list of API keys based on filters
func (akr *APIKeyRepository) List(ctx context.Context, filters APIKeyFilters) ([]*APIKey, error) {
	query := `
		SELECT id, name, user_id, team_id, key_prefix, permissions, 
		       created_at, expires_at, last_used_at, created_by, revoked, metadata
		FROM api_keys
		WHERE 1=1
	`
	
	args := []interface{}{}
	argCount := 1

	if filters.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, *filters.UserID)
		argCount++
	}

	if filters.TeamID != nil {
		query += fmt.Sprintf(" AND team_id = $%d", argCount)
		args = append(args, *filters.TeamID)
		argCount++
	}

	if filters.Revoked != nil {
		query += fmt.Sprintf(" AND revoked = $%d", argCount)
		args = append(args, *filters.Revoked)
		argCount++
	}
	
	query += " ORDER BY created_at DESC"
	
	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filters.Limit)
		argCount++
	}
	
	if filters.Page > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, (filters.Page-1)*filters.Limit)
	}
	
	rows, err := akr.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query API keys: %w", err)
	}
	defer rows.Close()

	var apiKeys []*APIKey
	for rows.Next() {
		var apiKey APIKey
		var permissionsBytes []byte
		var metadataBytes []byte

		err := rows.Scan(
			&apiKey.ID,
			&apiKey.Name,
			&apiKey.UserID,
			&apiKey.TeamID,
			&apiKey.KeyPrefix,
			&permissionsBytes,
			&apiKey.CreatedAt,
			&apiKey.ExpiresAt,
			&apiKey.LastUsedAt,
			&apiKey.CreatedBy,
			&apiKey.Revoked,
			&metadataBytes,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}

		// Unmarshal permissions
		if err := json.Unmarshal(permissionsBytes, &apiKey.Permissions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
		}

		// Unmarshal metadata
		if err := json.Unmarshal(metadataBytes, &apiKey.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		apiKeys = append(apiKeys, &apiKey)
	}

	return apiKeys, nil
}

// Count counts the number of API keys matching the filters
func (akr *APIKeyRepository) Count(ctx context.Context, filters APIKeyFilters) (int64, error) {
	query := "SELECT COUNT(*) FROM api_keys WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if filters.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, *filters.UserID)
		argCount++
	}

	if filters.TeamID != nil {
		query += fmt.Sprintf(" AND team_id = $%d", argCount)
		args = append(args, *filters.TeamID)
		argCount++
	}

	if filters.Revoked != nil {
		query += fmt.Sprintf(" AND revoked = $%d", argCount)
		args = append(args, *filters.Revoked)
		argCount++
	}

	var count int64
	err := akr.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count API keys: %w", err)
	}

	return count, nil
}