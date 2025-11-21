// backend/internal/auth/audit.go
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// AuditEvent represents an auditable event
type AuditEvent struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Timestamp time.Time `json:"timestamp" gorm:"index"`
	UserID    *uint     `json:"user_id,omitempty" gorm:"index"` // Nullable for system events
	User      *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Action    string    `json:"action" gorm:"not null"`      // login, logout, create, update, delete, etc.
	Resource  string    `json:"resource" gorm:"not null"`    // user, workflow, api_key, etc.
	ResourceID *uint    `json:"resource_id,omitempty"`       // ID of the affected resource
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Success   bool      `json:"success"`
	Details   string    `json:"details"`                     // JSON string of additional details
	CreatedAt time.Time `json:"created_at"`
}

// AuditService provides audit logging functionality
type AuditService struct {
	DB *gorm.DB
}

// NewAuditService creates a new audit service
func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{DB: db}
}

// LogEvent logs an audit event
func (as *AuditService) LogEvent(ctx context.Context, userID *uint, action, resource string, resourceID *uint, success bool, details map[string]interface{}, c interface{}) error {
	var ipAddress, userAgent string
	
	// Extract request details if context contains them
	if c != nil {
		// Assuming c is a Gin context or similar web framework context
		// This will depend on your web framework
		// For now, we'll leave these empty
	}

	// Convert details to JSON string
	detailsJSON := ""
	if details != nil {
		if jsonBytes, err := json.Marshal(details); err == nil {
			detailsJSON = string(jsonBytes)
		}
	}

	auditEvent := &AuditEvent{
		Timestamp:  time.Now(),
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Success:    success,
		Details:    detailsJSON,
	}

	if err := as.DB.Create(auditEvent).Error; err != nil {
		return fmt.Errorf("failed to log audit event: %w", err)
	}

	return nil
}

// LogUserLogin logs a user login event
func (as *AuditService) LogUserLogin(ctx context.Context, userID uint, success bool, c interface{}) error {
	return as.LogEvent(ctx, &userID, "login", "user", &userID, success, nil, c)
}

// LogUserLogout logs a user logout event
func (as *AuditService) LogUserLogout(ctx context.Context, userID uint, c interface{}) error {
	return as.LogEvent(ctx, &userID, "logout", "user", &userID, true, nil, c)
}

// LogWorkflowAction logs a workflow-related action
func (as *AuditService) LogWorkflowAction(ctx context.Context, userID uint, action string, workflowID uint, success bool, details map[string]interface{}, c interface{}) error {
	return as.LogEvent(ctx, &userID, action, "workflow", &workflowID, success, details, c)
}

// LogAPIKeyAction logs an API key-related action
func (as *AuditService) LogAPIKeyAction(ctx context.Context, userID uint, action string, apiKeyID uint, success bool, details map[string]interface{}, c interface{}) error {
	return as.LogEvent(ctx, &userID, action, "api_key", &apiKeyID, success, details, c)
}

// LogSystemEvent logs a system-level event (not associated with a user)
func (as *AuditService) LogSystemEvent(ctx context.Context, action, resource string, resourceID *uint, details map[string]interface{}) error {
	return as.LogEvent(ctx, nil, action, resource, resourceID, true, details, nil)
}

// GetAuditEvents retrieves audit events with optional filters
func (as *AuditService) GetAuditEvents(ctx context.Context, filters AuditFilters) ([]AuditEvent, error) {
	var events []AuditEvent
	query := as.DB.Model(&AuditEvent{})

	if filters.StartTime != nil {
		query = query.Where("timestamp >= ?", *filters.StartTime)
	}
	
	if filters.EndTime != nil {
		query = query.Where("timestamp <= ?", *filters.EndTime)
	}
	
	if filters.UserID != nil {
		query = query.Where("user_id = ?", *filters.UserID)
	}
	
	if filters.Action != "" {
		query = query.Where("action = ?", filters.Action)
	}
	
	if filters.Resource != "" {
		query = query.Where("resource = ?", filters.Resource)
	}
	
	if filters.Success != nil {
		query = query.Where("success = ?", *filters.Success)
	}

	if filters.Limit > 0 {
		query = query.Limit(int(filters.Limit))
	}
	
	if filters.Offset > 0 {
		query = query.Offset(int(filters.Offset))
	}

	// Order by timestamp descending by default
	query = query.Order("timestamp DESC")

	if err := query.Find(&events).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve audit events: %w", err)
	}

	return events, nil
}

// AuditFilters represents filters for retrieving audit events
type AuditFilters struct {
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	UserID    *uint      `json:"user_id,omitempty"`
	Action    string     `json:"action,omitempty"`
	Resource  string     `json:"resource,omitempty"`
	Success   *bool      `json:"success,omitempty"`
	Limit     uint       `json:"limit,omitempty"`
	Offset    uint       `json:"offset,omitempty"`
}

// GetAuditSummary returns summary statistics of audit events
func (as *AuditService) GetAuditSummary(ctx context.Context, startTime, endTime *time.Time) (*AuditSummary, error) {
	query := as.DB.Model(&AuditEvent{})
	
	if startTime != nil {
		query = query.Where("timestamp >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("timestamp <= ?", *endTime)
	}

	var totalEvents int64
	if err := query.Count(&totalEvents).Error; err != nil {
		return nil, fmt.Errorf("failed to count total events: %w", err)
	}

	// Get success/failure counts
	var successCount, failureCount int64
	successQuery := query.Where("success = ?", true)
	failureQuery := query.Where("success = ?", false)
	
	if err := successQuery.Count(&successCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count successful events: %w", err)
	}
	
	if err := failureQuery.Count(&failureCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count failed events: %w", err)
	}

	// Get top actions
	var topActions []struct {
		Action string
		Count  int64
	}
	actionQuery := as.DB.Model(&AuditEvent{})
	if startTime != nil {
		actionQuery = actionQuery.Where("timestamp >= ?", *startTime)
	}
	if endTime != nil {
		actionQuery = actionQuery.Where("timestamp <= ?", *endTime)
	}
	actionQuery = actionQuery.Select("action, COUNT(*) as count").Group("action").Order("count DESC").Limit(10)
	
	if err := actionQuery.Find(&topActions).Error; err != nil {
		return nil, fmt.Errorf("failed to get top actions: %w", err)
	}

	// Get top resources
	var topResources []struct {
		Resource string
		Count    int64
	}
	resourceQuery := as.DB.Model(&AuditEvent{})
	if startTime != nil {
		resourceQuery = resourceQuery.Where("timestamp >= ?", *startTime)
	}
	if endTime != nil {
		resourceQuery = resourceQuery.Where("timestamp <= ?", *endTime)
	}
	resourceQuery = resourceQuery.Select("resource, COUNT(*) as count").Group("resource").Order("count DESC").Limit(10)
	
	if err := resourceQuery.Find(&topResources).Error; err != nil {
		return nil, fmt.Errorf("failed to get top resources: %w", err)
	}

	summary := &AuditSummary{
		TotalEvents:    totalEvents,
		SuccessfulEvents: successCount,
		FailedEvents:   failureCount,
		TopActions:     topActions,
		TopResources:   topResources,
	}

	return summary, nil
}

// AuditSummary represents summary statistics of audit events
type AuditSummary struct {
	TotalEvents      int64                   `json:"total_events"`
	SuccessfulEvents int64                   `json:"successful_events"`
	FailedEvents     int64                   `json:"failed_events"`
	TopActions       []struct {
		Action string `json:"action"`
		Count  int64  `json:"count"`
	} `json:"top_actions"`
	TopResources []struct {
		Resource string `json:"resource"`
		Count    int64  `json:"count"`
	} `json:"top_resources"`
}

// CleanupOldEvents removes audit events older than the specified duration
func (as *AuditService) CleanupOldEvents(ctx context.Context, olderThan time.Duration) (int64, error) {
	deleteBefore := time.Now().Add(-olderThan)
	
	result := as.DB.Where("timestamp < ?", deleteBefore).Delete(&AuditEvent{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup old audit events: %w", result.Error)
	}
	
	return result.RowsAffected, nil
}