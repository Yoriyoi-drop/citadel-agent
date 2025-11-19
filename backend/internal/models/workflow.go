package models

import (
	"time"

	"gorm.io/gorm"
)

// Workflow represents a workflow in the database
type Workflow struct {
	ID          string         `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `json:"description"`
	Status      string         `gorm:"default:'active'" json:"status"` // 'active', 'inactive', 'archived'
	OwnerID     string         `json:"owner_id"` // Foreign key to User
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relationships
	Nodes []Node `gorm:"foreignKey:WorkflowID" json:"nodes"`
}

// TableName specifies the table name for Workflow
func (Workflow) TableName() string {
	return "workflows"
}