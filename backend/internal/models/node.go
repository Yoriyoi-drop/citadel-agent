package models

import (
	"time"

	"gorm.io/gorm"
)

// Node represents a workflow node in the database
type Node struct {
	ID          string                 `gorm:"primaryKey" json:"id"`
	WorkflowID  string                 `gorm:"not null;index" json:"workflow_id"`
	Type        string                 `gorm:"not null" json:"type"` // e.g., "http_request", "function", "trigger"
	Name        string                 `gorm:"not null" json:"name"`
	Description string                 `json:"description"`
	PositionX   float64                `json:"position_x"`
	PositionY   float64                `json:"position_y"`
	Settings    map[string]interface{} `gorm:"serializer:json" json:"settings"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	DeletedAt   gorm.DeletedAt         `gorm:"index" json:"deleted_at,omitempty"`

	// Relationships
	Workflow Workflow `gorm:"foreignKey:WorkflowID" json:"-"`
}

// TableName specifies the table name for Node
func (Node) TableName() string {
	return "nodes"
}