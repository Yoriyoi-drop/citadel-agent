package auth

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Role represents a user role with permissions
type Role struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null;unique"`
	Description string    `json:"description"`
	Permissions []string  `json:"permissions" gorm:"type:text[]"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Permission represents a specific permission
type Permission struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null;unique"`
	Description string    `json:"description"`
	Resource    string    `json:"resource"` // The resource this permission applies to
	Action      string    `json:"action"`   // The action allowed on the resource
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserRole represents the assignment of a role to a user
type UserRole struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id"`
	RoleID    string    `json:"role_id"`
	GrantedAt time.Time `json:"granted_at"`
	GrantedBy string    `json:"granted_by"`
	CreatedAt time.Time `json:"created_at"`
}

// RBACManager handles role-based access control
type RBACManager struct {
	db      *pgxpool.Pool
	roles   map[string]*Role
	mutex   sync.RWMutex
	cacheExpiry time.Duration
}

// NewRBACManager creates a new RBAC manager
func NewRBACManager(db *pgxpool.Pool) *RBACManager {
	manager := &RBACManager{
		db:      db,
		roles:   make(map[string]*Role),
		mutex:   sync.RWMutex{},
		cacheExpiry: 5 * time.Minute, // Cache roles for 5 minutes
	}

	// Initialize default roles
	manager.initDefaultRoles()

	return manager
}

// initDefaultRoles initializes default roles
func (rbac *RBACManager) initDefaultRoles() {
	defaultRoles := []Role{
		{
			ID:          "admin",
			Name:        "Administrator",
			Description: "Full access to all resources",
			Permissions: []string{
				"users:*",
				"workflows:*",
				"executions:*",
				"teams:*",
				"settings:*",
				"api_keys:*",
			},
		},
		{
			ID:          "member",
			Name:        "Member",
			Description: "Standard access to workflows and executions",
			Permissions: []string{
				"workflows:create",
				"workflows:read",
				"workflows:update",
				"workflows:delete",
				"executions:create",
				"executions:read",
				"executions:update",
				"executions:delete",
			},
		},
		{
			ID:          "viewer",
			Name:        "Viewer",
			Description: "Read-only access to workflows and executions",
			Permissions: []string{
				"workflows:read",
				"executions:read",
			},
		},
		{
			ID:          "api_user",
			Name:        "API User",
			Description: "Limited API access",
			Permissions: []string{
				"workflows:read",
				"executions:create",
				"executions:read",
			},
		},
	}

	// This would normally be saved to the database
	// For now, we'll store them in memory
	for _, role := range defaultRoles {
		rbac.roles[role.ID] = &role
	}
}

// CreateRole creates a new role
func (rbac *RBACManager) CreateRole(ctx context.Context, role *Role) error {
	if role.ID == "" {
		role.ID = uuid.New().String()
	}

	if role.Name == "" {
		return fmt.Errorf("role name is required")
	}

	// Check if role already exists
	if _, exists := rbac.roles[role.ID]; exists {
		return fmt.Errorf("role with ID %s already exists", role.ID)
	}

	// Validate permissions format
	if err := rbac.validatePermissions(role.Permissions); err != nil {
		return fmt.Errorf("invalid permissions: %w", err)
	}

	// Add role to memory
	rbac.mutex.Lock()
	rbac.roles[role.ID] = role
	rbac.mutex.Unlock()

	// In a real implementation, you would also save to the database
	// Here's what the database operation would look like:
	/*
		query := `
			INSERT INTO roles (id, name, description, permissions)
			VALUES ($1, $2, $3, $4)
		`
		_, err := rbac.db.Exec(ctx, query, role.ID, role.Name, role.Description, role.Permissions)
		if err != nil {
			return fmt.Errorf("failed to create role in database: %w", err)
		}
	*/

	return nil
}

// GetRole retrieves a role by ID
func (rbac *RBACManager) GetRole(ctx context.Context, roleID string) (*Role, error) {
	rbac.mutex.RLock()
	role, exists := rbac.roles[roleID]
	rbac.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("role with ID %s not found", roleID)
	}

	// In a real implementation, you would fetch from the database
	// with a query like:
	/*
		query := `SELECT id, name, description, permissions, created_at, updated_at FROM roles WHERE id = $1`
		var role Role
		err := rbac.db.QueryRow(ctx, query, roleID).Scan(&role.ID, &role.Name, &role.Description, &role.Permissions, &role.CreatedAt, &role.UpdatedAt)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil, fmt.Errorf("role with ID %s not found", roleID)
			}
			return nil, fmt.Errorf("failed to get role from database: %w", err)
		}
	*/

	return role, nil
}

// UpdateRole updates an existing role
func (rbac *RBACManager) UpdateRole(ctx context.Context, roleID string, updatedRole *Role) error {
	rbac.mutex.Lock()
	defer rbac.mutex.Unlock()

	existingRole, exists := rbac.roles[roleID]
	if !exists {
		return fmt.Errorf("role with ID %s not found", roleID)
	}

	// Validate permissions format
	if err := rbac.validatePermissions(updatedRole.Permissions); err != nil {
		return fmt.Errorf("invalid permissions: %w", err)
	}

	// Update the role
	existingRole.Name = updatedRole.Name
	existingRole.Description = updatedRole.Description
	existingRole.Permissions = updatedRole.Permissions
	existingRole.UpdatedAt = time.Now()

	// In a real implementation, you would update the database
	/*
		query := `
			UPDATE roles
			SET name = $1, description = $2, permissions = $3, updated_at = $4
			WHERE id = $5
		`
		_, err := rbac.db.Exec(ctx, query, updatedRole.Name, updatedRole.Description, updatedRole.Permissions, time.Now(), roleID)
		if err != nil {
			return fmt.Errorf("failed to update role in database: %w", err)
		}
	*/

	return nil
}

// DeleteRole deletes a role
func (rbac *RBACManager) DeleteRole(ctx context.Context, roleID string) error {
	if roleID == "admin" || roleID == "member" || roleID == "viewer" || roleID == "api_user" {
		return fmt.Errorf("cannot delete system default role: %s", roleID)
	}

	rbac.mutex.Lock()
	defer rbac.mutex.Unlock()

	_, exists := rbac.roles[roleID]
	if !exists {
		return fmt.Errorf("role with ID %s not found", roleID)
	}

	delete(rbac.roles, roleID)

	// In a real implementation, you would delete from the database
	/*
		query := `DELETE FROM roles WHERE id = $1`
		_, err := rbac.db.Exec(ctx, query, roleID)
		if err != nil {
			return fmt.Errorf("failed to delete role from database: %w", err)
		}
	*/

	return nil
}

// AssignRole assigns a role to a user
func (rbac *RBACManager) AssignRole(ctx context.Context, userID, roleID, grantedBy string) error {
	// Check if role exists
	rbac.mutex.RLock()
	role, roleExists := rbac.roles[roleID]
	rbac.mutex.RUnlock()

	if !roleExists {
		return fmt.Errorf("role with ID %s not found", roleID)
	}

	// In a real implementation, you would save the user-role assignment to the database
	/*
		query := `
			INSERT INTO user_roles (id, user_id, role_id, granted_at, granted_by)
			VALUES ($1, $2, $3, $4, $5)
		`
		userRoleID := uuid.New().String()
		_, err := rbac.db.Exec(ctx, query, userRoleID, userID, roleID, time.Now(), grantedBy)
		if err != nil {
			return fmt.Errorf("failed to assign role to user: %w", err)
		}
	*/

	return nil
}

// RevokeRole revokes a role from a user
func (rbac *RBACManager) RevokeRole(ctx context.Context, userID, roleID string) error {
	// In a real implementation, you would delete the user-role assignment from the database
	/*
		query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`
		_, err := rbac.db.Exec(ctx, query, userID, roleID)
		if err != nil {
			return fmt.Errorf("failed to revoke role from user: %w", err)
		}
	*/

	return nil
}

// GetUserRoles retrieves all roles assigned to a user
func (rbac *RBACManager) GetUserRoles(ctx context.Context, userID string) ([]*Role, error) {
	var userRoles []*Role

	// In a real implementation, you would query the database:
	/*
		query := `
			SELECT r.id, r.name, r.description, r.permissions, r.created_at, r.updated_at
			FROM roles r
			JOIN user_roles ur ON r.id = ur.role_id
			WHERE ur.user_id = $1
		`
		rows, err := rbac.db.Query(ctx, query, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user roles from database: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var role Role
			err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.Permissions, &role.CreatedAt, &role.UpdatedAt)
			if err != nil {
				return nil, fmt.Errorf("failed to scan role: %w", err)
			}
			userRoles = append(userRoles, &role)
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating over rows: %w", err)
		}
	*/

	// For now, return a default role based on a simple lookup
	// This is a simplified implementation
	defaultRole := rbac.getDefaultRoleForUser(userID)
	if defaultRole != nil {
		userRoles = append(userRoles, defaultRole)
	}

	return userRoles, nil
}

// getDefaultRoleForUser returns a default role for a user (simplified implementation)
func (rbac *RBACManager) getDefaultRoleForUser(userID string) *Role {
	// This is a simplified implementation
	// In a real system, you would look up the actual assigned roles from the database

	// For demonstration purposes, return a viewer role
	rbac.mutex.RLock()
	defer rbac.mutex.RUnlock()

	if role, exists := rbac.roles["viewer"]; exists {
		return role
	}

	return nil
}

// HasPermission checks if a user has a specific permission
func (rbac *RBACManager) HasPermission(ctx context.Context, userID, permission string) (bool, error) {
	userRoles, err := rbac.GetUserRoles(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user roles: %w", err)
	}

	for _, role := range userRoles {
		for _, rolePermission := range role.Permissions {
			if rbac.permissionMatches(rolePermission, permission) {
				return true, nil
			}
		}
	}

	return false, nil
}

// permissionMatches checks if a role permission matches the requested permission
// Supports wildcard matching (e.g. "workflows:*" matches "workflows:create")
func (rbac *RBACManager) permissionMatches(rolePermission, requestedPermission string) bool {
	// Exact match
	if rolePermission == requestedPermission {
		return true
	}

	// Check for wildcard in role permission
	if strings.Contains(rolePermission, ":*") {
		roleResource := strings.Split(rolePermission, ":*")[0]
		requestedResource := strings.Split(requestedPermission, ":")[0]
		
		if roleResource == requestedResource {
			return true
		}
	}

	return false
}

// validatePermissions validates that permissions are in the correct format
func (rbac *RBACManager) validatePermissions(permissions []string) error {
	for _, perm := range permissions {
		if !strings.Contains(perm, ":") {
			return fmt.Errorf("permission %s is not in the correct format (resource:action)", perm)
		}
	}
	return nil
}

// CheckUserHasResourceAccess checks if a user has access to a specific resource with a specific action
func (rbac *RBACManager) CheckUserHasResourceAccess(ctx context.Context, userID, resource, action string) (bool, error) {
	return rbac.HasPermission(ctx, userID, fmt.Sprintf("%s:%s", resource, action))
}

// GetAllRoles returns all available roles
func (rbac *RBACManager) GetAllRoles(ctx context.Context) ([]*Role, error) {
	rbac.mutex.RLock()
	defer rbac.mutex.RUnlock()

	roles := make([]*Role, 0, len(rbac.roles))
	for _, role := range rbac.roles {
		roles = append(roles, role)
	}

	return roles, nil
}

// CreatePermission creates a new permission
func (rbac *RBACManager) CreatePermission(ctx context.Context, permission *Permission) error {
	if permission.ID == "" {
		permission.ID = uuid.New().String()
	}

	// In a real implementation, you would save to the database
	// For now, we'll just validate the format
	if permission.Resource == "" || permission.Action == "" {
		return fmt.Errorf("permission resource and action are required")
	}

	return nil
}

// GetPermission returns a specific permission
func (rbac *RBACManager) GetPermission(ctx context.Context, permissionID string) (*Permission, error) {
	// In a real implementation, you would fetch from the database
	return nil, fmt.Errorf("not implemented")
}

// Authorize checks if a user has the required permissions for an action
// This provides a high-level authorization check
func (rbac *RBACManager) Authorize(ctx context.Context, userID string, requiredPermissions []string) (bool, error) {
	for _, requiredPerm := range requiredPermissions {
		hasPerm, err := rbac.HasPermission(ctx, userID, requiredPerm)
		if err != nil {
			return false, fmt.Errorf("failed to check permission %s: %w", requiredPerm, err)
		}
		if !hasPerm {
			return false, nil
		}
	}
	return true, nil
}