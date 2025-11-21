// backend/internal/auth/models.go
package auth

import (
	"time"

	"gorm.io/gorm"
)

// User represents a system user
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	
	Email       string `json:"email" gorm:"uniqueIndex;not null"`
	Username    string `json:"username" gorm:"uniqueIndex"`
	Password    string `json:"password" gorm:"not null"` // Hashed
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	LastLogin   *time.Time `json:"last_login"`
	
	// Relationships
	Roles      []Role      `json:"roles" gorm:"many2many:user_roles;"`
	Workflows  []Workflow  `json:"workflows" gorm:"foreignKey:OwnerID"`
	APIKeys    []APIKey    `json:"api_keys" gorm:"foreignKey:UserID"`
}

// Role represents a set of permissions
type Role struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"uniqueIndex;not null"` // admin, user, viewer, etc.
	Description string `json:"description"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	
	// Relationships
	Users     []User     `json:"users" gorm:"many2many:user_roles;"`
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// Permission represents a specific action that can be performed
type Permission struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"uniqueIndex;not null"` // workflow:create, workflow:edit, etc.
	Description string `json:"description"`
	Resource    string `json:"resource" gorm:"not null"` // workflow, user, system, etc.
	Action      string `json:"action" gorm:"not null"`   // create, read, update, delete
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// APIKey represents an API key for programmatic access
type APIKey struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	Name      string    `json:"name" gorm:"not null"`
	KeyHash   string    `json:"key_hash" gorm:"not null"` // Store only hash
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// Relationships
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// Workflow represents a workflow entity for permissions
type Workflow struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	OwnerID   uint      `json:"owner_id" gorm:"not null"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// Relationships
	Owner User `json:"owner" gorm:"foreignKey:OwnerID"`
}