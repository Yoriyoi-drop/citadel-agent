package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the database
type User struct {
	ID        string         `gorm:"primaryKey" json:"id"`
	Email     string         `gorm:"not null;unique;index" json:"email"`
	Username  string         `gorm:"not null;unique;index" json:"username"`
	Password  string         `gorm:"not null" json:"password"` // Hashed password
	FirstName string         `json:"first_name"`
	LastName  string         `json:"last_name"`
	Role      string         `gorm:"default:'user'" json:"role"` // 'admin', 'user', 'viewer'
	Status    string         `gorm:"default:'active'" json:"status"` // 'active', 'inactive', 'suspended'
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relationships
	Workflows []Workflow `gorm:"foreignKey:OwnerID" json:"workflows"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}