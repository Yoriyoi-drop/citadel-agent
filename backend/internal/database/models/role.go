package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role represents a user role with associated permissions
type Role struct {
	ID          string         `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string         `gorm:"uniqueIndex;not null" json:"name"`
	Description string         `json:"description"`
	Permissions []string       `gorm:"type:text[];not null" json:"permissions"`
	IsSystem    bool           `gorm:"default:false" json:"is_system"` // System roles cannot be deleted
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate hook to generate UUID
func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

// UserRole represents the many-to-many relationship between users and roles
type UserRole struct {
	UserID    string    `gorm:"type:uuid;primaryKey" json:"user_id"`
	RoleID    string    `gorm:"type:uuid;primaryKey" json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Permission constants
const (
	// Workflow permissions
	PermissionWorkflowCreate  = "workflow:create"
	PermissionWorkflowRead    = "workflow:read"
	PermissionWorkflowUpdate  = "workflow:update"
	PermissionWorkflowDelete  = "workflow:delete"
	PermissionWorkflowExecute = "workflow:execute"

	// Node permissions
	PermissionNodeCreate = "node:create"
	PermissionNodeRead   = "node:read"
	PermissionNodeUpdate = "node:update"
	PermissionNodeDelete = "node:delete"

	// Execution permissions
	PermissionExecutionRead   = "execution:read"
	PermissionExecutionCancel = "execution:cancel"

	// User permissions
	PermissionUserCreate = "user:create"
	PermissionUserRead   = "user:read"
	PermissionUserUpdate = "user:update"
	PermissionUserDelete = "user:delete"

	// Role permissions
	PermissionRoleCreate = "role:create"
	PermissionRoleRead   = "role:read"
	PermissionRoleUpdate = "role:update"
	PermissionRoleDelete = "role:delete"

	// API Key permissions
	PermissionAPIKeyCreate = "apikey:create"
	PermissionAPIKeyRead   = "apikey:read"
	PermissionAPIKeyRevoke = "apikey:revoke"

	// Audit log permissions
	PermissionAuditLogRead = "auditlog:read"

	// Admin permission (grants all permissions)
	PermissionAdmin = "admin:*"
)

// Default roles
var (
	RoleAdmin = Role{
		Name:        "admin",
		Description: "Administrator with full access",
		Permissions: []string{PermissionAdmin},
		IsSystem:    true,
	}

	RoleEditor = Role{
		Name:        "editor",
		Description: "Can create and edit workflows",
		Permissions: []string{
			PermissionWorkflowCreate,
			PermissionWorkflowRead,
			PermissionWorkflowUpdate,
			PermissionWorkflowExecute,
			PermissionNodeRead,
			PermissionExecutionRead,
		},
		IsSystem: true,
	}

	RoleViewer = Role{
		Name:        "viewer",
		Description: "Read-only access",
		Permissions: []string{
			PermissionWorkflowRead,
			PermissionNodeRead,
			PermissionExecutionRead,
		},
		IsSystem: true,
	}

	RoleExecutor = Role{
		Name:        "executor",
		Description: "Can execute workflows",
		Permissions: []string{
			PermissionWorkflowRead,
			PermissionWorkflowExecute,
			PermissionExecutionRead,
		},
		IsSystem: true,
	}
)
