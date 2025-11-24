package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AuditLog represents an audit trail entry
type AuditLog struct {
	ID         string         `gorm:"type:uuid;primaryKey" json:"id"`
	UserID     string         `gorm:"type:uuid;index" json:"user_id"`
	Action     string         `gorm:"index;not null" json:"action"` // e.g., "workflow.create", "user.delete"
	Resource   string         `gorm:"not null" json:"resource"`     // e.g., "workflow", "user", "apikey"
	ResourceID string         `gorm:"index" json:"resource_id"`
	Changes    datatypes.JSON `json:"changes"` // JSON of what changed
	IPAddress  string         `json:"ip_address"`
	UserAgent  string         `json:"user_agent"`
	Status     string         `gorm:"index" json:"status"` // "success", "failure"
	ErrorMsg   string         `json:"error_msg,omitempty"`
	CreatedAt  time.Time      `gorm:"index" json:"created_at"`
}

// BeforeCreate hook to generate UUID
func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}

// AuditAction constants
const (
	// Workflow actions
	ActionWorkflowCreate  = "workflow.create"
	ActionWorkflowUpdate  = "workflow.update"
	ActionWorkflowDelete  = "workflow.delete"
	ActionWorkflowExecute = "workflow.execute"

	// User actions
	ActionUserCreate = "user.create"
	ActionUserUpdate = "user.update"
	ActionUserDelete = "user.delete"
	ActionUserLogin  = "user.login"
	ActionUserLogout = "user.logout"

	// Role actions
	ActionRoleCreate = "role.create"
	ActionRoleUpdate = "role.update"
	ActionRoleDelete = "role.delete"
	ActionRoleAssign = "role.assign"

	// API Key actions
	ActionAPIKeyCreate = "apikey.create"
	ActionAPIKeyRevoke = "apikey.revoke"
	ActionAPIKeyRotate = "apikey.rotate"

	// Settings actions
	ActionSettingsUpdate = "settings.update"
)

// AuditStatus constants
const (
	AuditStatusSuccess = "success"
	AuditStatusFailure = "failure"
)
