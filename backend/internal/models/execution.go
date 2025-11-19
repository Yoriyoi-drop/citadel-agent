package models

import (
	"time"

	"gorm.io/gorm"
)

// Execution represents a single execution of a workflow
type Execution struct {
	ID          string                 `gorm:"primaryKey" json:"id"`
	WorkflowID  string                 `gorm:"not null;index" json:"workflow_id"`
	Status      string                 `gorm:"not null;default:'running'" json:"status"` // 'running', 'completed', 'failed', 'cancelled'
	StartedAt   time.Time              `json:"started_at"`
	EndedAt     *time.Time             `json:"ended_at,omitempty"`
	Result      map[string]interface{} `gorm:"serializer:json" json:"result"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	DeletedAt   gorm.DeletedAt         `gorm:"index" json:"deleted_at,omitempty"`

	// Relationships
	Workflow Workflow `gorm:"foreignKey:WorkflowID" json:"-"`
}

// TableName specifies the table name for Execution
func (Execution) TableName() string {
	return "executions"
}