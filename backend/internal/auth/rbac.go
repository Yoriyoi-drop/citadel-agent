package auth

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

var (
	ErrPermissionDenied = errors.New("permission denied")
	ErrRoleNotFound     = errors.New("role not found")
	ErrInvalidRole      = errors.New("invalid role")
)

// RBACService handles role-based access control
type RBACService struct {
	db *gorm.DB
}

// NewRBACService creates a new RBAC service
func NewRBACService(db *gorm.DB) *RBACService {
	return &RBACService{db: db}
}

// HasPermission checks if a user has a specific permission
func (s *RBACService) HasPermission(userID string, permission string) (bool, error) {
	// Get user's roles
	var roleIDs []string
	err := s.db.Table("user_roles").
		Where("user_id = ?", userID).
		Pluck("role_id", &roleIDs).Error
	if err != nil {
		return false, err
	}

	if len(roleIDs) == 0 {
		return false, nil
	}

	// Get permissions from roles
	var roles []struct {
		Permissions []string `gorm:"serializer:json"`
	}
	err = s.db.Table("roles").
		Select("permissions").
		Where("id IN ?", roleIDs).
		Where("deleted_at IS NULL").
		Find(&roles).Error
	if err != nil {
		return false, err
	}

	// Check if user has admin permission (grants all)
	for _, role := range roles {
		if contains(role.Permissions, "admin:*") {
			return true, nil
		}
	}

	// Check for specific permission
	for _, role := range roles {
		if contains(role.Permissions, permission) {
			return true, nil
		}
	}

	return false, nil
}

// HasAnyPermission checks if a user has any of the specified permissions
func (s *RBACService) HasAnyPermission(userID string, permissions []string) (bool, error) {
	for _, perm := range permissions {
		has, err := s.HasPermission(userID, perm)
		if err != nil {
			return false, err
		}
		if has {
			return true, nil
		}
	}
	return false, nil
}

// HasAllPermissions checks if a user has all of the specified permissions
func (s *RBACService) HasAllPermissions(userID string, permissions []string) (bool, error) {
	for _, perm := range permissions {
		has, err := s.HasPermission(userID, perm)
		if err != nil {
			return false, err
		}
		if !has {
			return false, nil
		}
	}
	return true, nil
}

// GetUserPermissions returns all permissions for a user
func (s *RBACService) GetUserPermissions(userID string) ([]string, error) {
	// Get user's roles
	var roleIDs []string
	err := s.db.Table("user_roles").
		Where("user_id = ?", userID).
		Pluck("role_id", &roleIDs).Error
	if err != nil {
		return nil, err
	}

	if len(roleIDs) == 0 {
		return []string{}, nil
	}

	// Get permissions from roles
	var roles []struct {
		Permissions []string
	}
	err = s.db.Table("roles").
		Select("permissions").
		Where("id IN ?", roleIDs).
		Where("deleted_at IS NULL").
		Find(&roles).Error
	if err != nil {
		return nil, err
	}

	// Flatten and deduplicate permissions
	permMap := make(map[string]bool)
	for _, role := range roles {
		for _, perm := range role.Permissions {
			permMap[perm] = true
		}
	}

	permissions := make([]string, 0, len(permMap))
	for perm := range permMap {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// AssignRole assigns a role to a user
func (s *RBACService) AssignRole(userID, roleID string) error {
	// Check if role exists
	var count int64
	err := s.db.Table("roles").
		Where("id = ?", roleID).
		Where("deleted_at IS NULL").
		Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrRoleNotFound
	}

	// Check if already assigned
	err = s.db.Table("user_roles").
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // Already assigned
	}

	// Assign role
	return s.db.Exec(
		"INSERT INTO user_roles (user_id, role_id, created_at) VALUES (?, ?, NOW())",
		userID, roleID,
	).Error
}

// RemoveRole removes a role from a user
func (s *RBACService) RemoveRole(userID, roleID string) error {
	return s.db.Exec(
		"DELETE FROM user_roles WHERE user_id = ? AND role_id = ?",
		userID, roleID,
	).Error
}

// GetUserRoles returns all roles for a user
func (s *RBACService) GetUserRoles(userID string) ([]string, error) {
	var roleIDs []string
	err := s.db.Table("user_roles").
		Where("user_id = ?", userID).
		Pluck("role_id", &roleIDs).Error
	return roleIDs, err
}

// ValidatePermissions checks if all permissions are valid
func (s *RBACService) ValidatePermissions(permissions []string) error {
	validPrefixes := []string{
		"workflow:", "node:", "execution:", "user:",
		"role:", "apikey:", "auditlog:", "admin:",
	}

	for _, perm := range permissions {
		valid := false
		for _, prefix := range validPrefixes {
			if strings.HasPrefix(perm, prefix) {
				valid = true
				break
			}
		}
		if !valid {
			return errors.New("invalid permission: " + perm)
		}
	}
	return nil
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
