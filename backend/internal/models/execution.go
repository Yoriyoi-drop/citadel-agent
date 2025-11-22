package models

import (
	"time"

	"gorm.io/gorm"
)

// Execution represents a single execution of a workflow
type Execution struct {
	ID          string                 `gorm:"primaryKey" json:"id"`
	WorkflowID  string                 `gorm:"not null;index" json:"workflow_id"`
	Name        string                 `gorm:"not null" json:"name"` // Name of the execution
	Status      string                 `gorm:"not null;default:'running'" json:"status"` // 'running', 'completed', 'failed', 'cancelled'
	StartedAt   time.Time              `json:"started_at"`
	EndedAt     *time.Time             `json:"ended_at,omitempty"`
	Result      map[string]interface{} `gorm:"serializer:json" json:"result"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	DeletedAt   gorm.DeletedAt         `gorm:"index" json:"deleted_at,omitempty"`
	TriggeredBy string                 `gorm:"not null" json:"triggered_by"` // Who triggered the execution (user, schedule, api, etc.)
	RetryCount  int                    `gorm:"default:0" json:"retry_count"` // Number of retry attempts
	UserID      string                 `gorm:"index" json:"user_id,omitempty"` // User who triggered the execution
	TeamID      string                 `gorm:"index" json:"team_id,omitempty"` // Team ID if applicable
	ParentID    *string                `gorm:"index" json:"parent_id,omitempty"` // Parent execution ID for nested executions
	CompletedAt *time.Time             `json:"completed_at,omitempty"` // Added to match the reference used in code

	// Relationships
	Workflow Workflow `gorm:"foreignKey:WorkflowID" json:"-"`
}

// TableName specifies the table name for Execution
func (Execution) TableName() string {
	return "executions"
}