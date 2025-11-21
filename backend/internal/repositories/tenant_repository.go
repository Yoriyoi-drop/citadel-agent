// backend/internal/repositories/tenant_repository.go
package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TenantRepository handles tenant-related database operations
type TenantRepository struct {
	db *pgxpool.Pool
}

// Tenant represents a tenant in the database
type Tenant struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	OwnerID     string                 `json:"owner_id"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Settings    map[string]interface{} `json:"settings"`
}

// TenantFilters represents filters for listing tenants
type TenantFilters struct {
	Status *string
	Limit  int
	Offset int
}

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(db *pgxpool.Pool) *TenantRepository {
	return &TenantRepository{
		db: db,
	}
}

// Create creates a new tenant
func (tr *TenantRepository) Create(ctx context.Context, tenant *Tenant) (*Tenant, error) {
	// Serialize settings to JSON
	settingsJSON, err := json.Marshal(tenant.Settings)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `
		INSERT INTO tenants (
			id, name, description, owner_id, status, 
			created_at, updated_at, settings
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, name, description, owner_id, status, created_at, updated_at
	`

	var createdTenant Tenant
	err = tr.db.QueryRow(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Description,
		tenant.OwnerID,
		tenant.Status,
		tenant.CreatedAt,
		tenant.UpdatedAt,
		settingsJSON,
	).Scan(
		&createdTenant.ID,
		&createdTenant.Name,
		&createdTenant.Description,
		&createdTenant.OwnerID,
		&createdTenant.Status,
		&createdTenant.CreatedAt,
		&createdTenant.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Set settings from original input since they're not returned by the query
	createdTenant.Settings = tenant.Settings

	return &createdTenant, nil
}

// GetByID retrieves a tenant by ID
func (tr *TenantRepository) GetByID(ctx context.Context, id string) (*Tenant, error) {
	query := `
		SELECT id, name, description, owner_id, status, 
		       created_at, updated_at, settings
		FROM tenants
		WHERE id = $1
	`

	var tenant Tenant
	var settingsJSON []byte

	err := tr.db.QueryRow(ctx, query, id).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Description,
		&tenant.OwnerID,
		&tenant.Status,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
		&settingsJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// Deserialize settings
	if settingsJSON != nil {
		if err := json.Unmarshal(settingsJSON, &tenant.Settings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
		}
	}

	return &tenant, nil
}

// GetByName retrieves a tenant by name
func (tr *TenantRepository) GetByName(ctx context.Context, name string) (*Tenant, error) {
	query := `
		SELECT id, name, description, owner_id, status, 
		       created_at, updated_at, settings
		FROM tenants
		WHERE name = $1
	`

	var tenant Tenant
	var settingsJSON []byte

	err := tr.db.QueryRow(ctx, query, name).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Description,
		&tenant.OwnerID,
		&tenant.Status,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
		&settingsJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil, nil if not found
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// Deserialize settings
	if settingsJSON != nil {
		if err := json.Unmarshal(settingsJSON, &tenant.Settings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
		}
	}

	return &tenant, nil
}

// GetByOwner retrieves tenants by owner ID
func (tr *TenantRepository) GetByOwner(ctx context.Context, ownerID string) ([]*Tenant, error) {
	query := `
		SELECT id, name, description, owner_id, status, 
		       created_at, updated_at, settings
		FROM tenants
		WHERE owner_id = $1
	`

	rows, err := tr.db.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*Tenant
	for rows.Next() {
		var tenant Tenant
		var settingsJSON []byte

		err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.Description,
			&tenant.OwnerID,
			&tenant.Status,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
			&settingsJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}

		// Deserialize settings
		if settingsJSON != nil {
			if err := json.Unmarshal(settingsJSON, &tenant.Settings); err != nil {
				return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
			}
		}

		tenants = append(tenants, &tenant)
	}

	return tenants, nil
}

// Update updates an existing tenant
func (tr *TenantRepository) Update(ctx context.Context, tenant *Tenant) (*Tenant, error) {
	// Serialize settings to JSON
	settingsJSON, err := json.Marshal(tenant.Settings)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `
		UPDATE tenants
		SET name = $2, description = $3, status = $4, updated_at = $5, settings = $6
		WHERE id = $1
		RETURNING id, name, description, owner_id, status, created_at, updated_at
	`

	var updatedTenant Tenant
	err = tr.db.QueryRow(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Description,
		tenant.Status,
		tenant.UpdatedAt,
		settingsJSON,
	).Scan(
		&updatedTenant.ID,
		&updatedTenant.Name,
		&updatedTenant.Description,
		&updatedTenant.OwnerID,
		&updatedTenant.Status,
		&updatedTenant.CreatedAt,
		&updatedTenant.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	// Set settings from original input
	updatedTenant.Settings = tenant.Settings

	return &updatedTenant, nil
}

// UpdateStatus updates only the status of a tenant
func (tr *TenantRepository) UpdateStatus(ctx context.Context, id, status string) error {
	query := `
		UPDATE tenants
		SET status = $2, updated_at = $3
		WHERE id = $1
	`

	_, err := tr.db.Exec(ctx, query, id, status, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update tenant status: %w", err)
	}

	return nil
}

// List retrieves a list of tenants based on filters
func (tr *TenantRepository) List(ctx context.Context, filters TenantFilters) ([]*Tenant, error) {
	query := `
		SELECT id, name, description, owner_id, status, 
		       created_at, updated_at, settings
		FROM tenants
		WHERE 1=1
	`
	
	args := []interface{}{}
	argCount := 1

	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, *filters.Status)
		argCount++
	}
	
	query += " ORDER BY created_at DESC"
	
	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filters.Limit)
		argCount++
	}
	
	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filters.Offset)
	}
	
	rows, err := tr.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*Tenant
	for rows.Next() {
		var tenant Tenant
		var settingsJSON []byte

		err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.Description,
			&tenant.OwnerID,
			&tenant.Status,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
			&settingsJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}

		// Deserialize settings
		if settingsJSON != nil {
			if err := json.Unmarshal(settingsJSON, &tenant.Settings); err != nil {
				return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
			}
		}

		tenants = append(tenants, &tenant)
	}

	return tenants, nil
}

// Delete removes a tenant by ID
func (tr *TenantRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM tenants WHERE id = $1"

	result, err := tr.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("tenant not found: %s", id)
	}

	return nil
}

// Count counts the number of tenants
func (tr *TenantRepository) Count(ctx context.Context, status *string) (int64, error) {
	query := "SELECT COUNT(*) FROM tenants WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, *status)
		argCount++
	}

	var count int64
	err := tr.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count tenants: %w", err)
	}

	return count, nil
}