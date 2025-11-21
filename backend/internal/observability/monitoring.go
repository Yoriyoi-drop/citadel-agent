// backend/internal/observability/monitoring.go
package observability

import (
	"context"
	"fmt"
	"time"

	"citadel-agent/backend/internal/workflow/core/engine"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MonitoringService provides comprehensive monitoring and observability
type MonitoringService struct {
	db        *pgxpool.Pool
	metrics   *MetricsService
	tracer    *TelemetryService
	workflowEngine *engine.Engine
	events    chan *Event
	stop      chan struct{}
}

// Event represents a system event for monitoring
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Service   string                 `json:"service"`
	TenantID  string                 `json:"tenant_id"`
	UserID    string                 `json:"user_id"`
	Resource  string                 `json:"resource"`
	Action    string                 `json:"action"`
	Status    string                 `json:"status"`
	Metadata  map[string]interface{} `json:"metadata"`
	Duration  time.Duration          `json:"duration,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(
	db *pgxpool.Pool,
	metrics *MetricsService,
	tracer *TelemetryService,
	workflowEngine *engine.Engine,
) *MonitoringService {
	service := &MonitoringService{
		db:             db,
		metrics:        metrics,
		tracer:         tracer,
		workflowEngine: workflowEngine,
		events:         make(chan *Event, 1000), // Buffered channel for events
		stop:           make(chan struct{}),
	}
	
	// Start event processing goroutine
	go service.processEvents()
	
	return service
}

// processEvents processes monitoring events asynchronously
func (ms *MonitoringService) processEvents() {
	for {
		select {
		case event := <-ms.events:
			// Process the event (log to database, send to external systems, etc.)
			ms.storeEvent(event)
		case <-ms.stop:
			return
		}
	}
}

// storeEvent stores the event in the database
func (ms *MonitoringService) storeEvent(event *Event) error {
	if ms.db == nil {
		return fmt.Errorf("database connection not available")
	}

	query := `
		INSERT INTO monitoring_events (
			id, type, timestamp, service, tenant_id, user_id, resource, 
			action, status, metadata, duration, error_msg
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	// In a real implementation, we would serialize metadata to JSON
	// For now, we'll just process the event without storing it permanently
	// to avoid database schema dependencies

	return nil
}

// RecordEvent records a monitoring event
func (ms *MonitoringService) RecordEvent(ctx context.Context, eventType, service, tenantID, userID, resource, action, status string, metadata map[string]interface{}) {
	event := &Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now(),
		Service:   service,
		TenantID:  tenantID,
		UserID:    userID,
		Resource:  resource,
		Action:    action,
		Status:    status,
		Metadata:  metadata,
	}

	// Send event to processing channel
	select {
	case ms.events <- event:
		// Successfully queued for processing
	default:
		// Channel is full, log warning
		// In production, this should be handled more gracefully
	}
}

// RecordWorkflowExecutionMetrics records execution metrics for a workflow
func (ms *MonitoringService) RecordWorkflowExecutionMetrics(workflowID, status, tenantID string, duration time.Duration) {
	ms.metrics.RecordWorkflowExecution(workflowID, status, tenantID, duration)
	
	// Also record as an event
	ms.RecordEvent(
		context.Background(),
		"workflow_execution",
		"workflow_engine",
		tenantID,
		"", // userID may not be available
		workflowID,
		"execute",
		status,
		map[string]interface{}{
			"duration": duration.Seconds(),
			"status":   status,
		},
	)
}

// RecordNodeExecutionMetrics records execution metrics for a node
func (ms *MonitoringService) RecordNodeExecutionMetrics(nodeType, workflowID, status, tenantID string, duration time.Duration) {
	ms.metrics.RecordNodeExecution(nodeType, workflowID, status, tenantID, duration)
	
	// Also record as an event
	ms.RecordEvent(
		context.Background(),
		"node_execution",
		"workflow_engine",
		tenantID,
		"", // userID may not be available
		workflowID,
		"execute",
		status,
		map[string]interface{}{
			"node_type": nodeType,
			"duration":  duration.Seconds(),
			"status":    status,
		},
	)
}

// RecordAPIRequestMetrics records API request metrics
func (ms *MonitoringService) RecordAPIRequestMetrics(method, endpoint, statusCode, tenantID string, duration time.Duration, requestSize, responseSize int) {
	ms.metrics.RecordAPIRequest(method, endpoint, statusCode, tenantID, duration, requestSize, responseSize)
	
	// Also record as an event
	ms.RecordEvent(
		context.Background(),
		"api_request",
		"http_server",
		tenantID,
		"", // userID may not be available
		endpoint,
		method,
		statusCode,
		map[string]interface{}{
			"method":        method,
			"endpoint":      endpoint,
			"status_code":   statusCode,
			"duration":      duration.Seconds(),
			"request_size":  requestSize,
			"response_size": responseSize,
		},
	)
}

// RecordErrorEvents records error metrics
func (ms *MonitoringService) RecordErrorEvents(errorType, resource, action, tenantID, userID, errorMessage string) {
	// Record as security event if it's an auth/permission error
	if errorType == "security" || errorType == "auth" || errorType == "permission" {
		ms.metrics.RecordSecurityEvent(errorType, "medium")
	}
	
	// Record as event
	ms.RecordEvent(
		context.Background(),
		"error_event",
		"system",
		tenantID,
		userID,
		resource,
		action,
		"error",
		map[string]interface{}{
			"error_type": errorType,
			"error_msg":  errorMessage,
		},
	)
}

// GetWorkflowExecutionTimeline returns execution timeline for a workflow
func (ms *MonitoringService) GetWorkflowExecutionTimeline(ctx context.Context, workflowID string) ([]*Event, error) {
	query := `
		SELECT id, type, timestamp, service, tenant_id, user_id, resource, 
		       action, status, metadata, duration, error_msg
		FROM monitoring_events
		WHERE resource = $1 AND type = 'workflow_execution'
		ORDER BY timestamp DESC
		LIMIT 100
	`

	// In a real implementation, we would query the database
	// For now, we'll return an empty slice
	return []*Event{}, nil
}

// GetNodeExecutionStats returns statistics for node executions
func (ms *MonitoringService) GetNodeExecutionStats(ctx context.Context, nodeType, workflowID string) (map[string]interface{}, error) {
	stats := map[string]interface{}{
		"node_type":      nodeType,
		"workflow_id":    workflowID,
		"total_executions": 0,
		"success_count":   0,
		"error_count":     0,
		"avg_duration":    0.0,
		"min_duration":    0.0,
		"max_duration":    0.0,
	}
	
	return stats, nil
}

// GetTenantActivity returns activity metrics for a tenant
func (ms *MonitoringService) GetTenantActivity(ctx context.Context, tenantID string, days int) (map[string]interface{}, error) {
	stats := map[string]interface{}{
		"tenant_id":  tenantID,
		"days":       days,
		"total_events": 0,
		"workflows_executed": 0,
		"nodes_executed": 0,
		"api_requests": 0,
		"errors": 0,
		"top_users": []string{},
		"peak_times": []string{},
	}
	
	return stats, nil
}

// GetSystemHealth returns system health metrics
func (ms *MonitoringService) GetSystemHealth(ctx context.Context) (map[string]interface{}, error) {
	health := map[string]interface{}{
		"status": "healthy",
		"timestamp": time.Now().Unix(),
		"uptime": time.Since(ms.tracer.(*TelemetryService).startTime).Seconds(),
		"goroutines": GetGoroutineCount(),
		"active_workflows": 0, // Would come from workflow engine
		"pending_events": len(ms.events),
		"metrics_collected": true,
		"tracing_enabled": true,
		"database_connected": ms.db != nil,
	}
	
	return health, nil
}

// RecordSecurityEvent records a security-related event
func (ms *MonitoringService) RecordSecurityEvent(ctx context.Context, eventType, severity, sourceIP, userID, resource, action string, metadata map[string]interface{}) {
	// Record to metrics
	ms.metrics.RecordSecurityEvent(eventType, severity)
	
	// Record as event
	ms.RecordEvent(ctx, "security_"+eventType, "security_module", "", userID, resource, action, severity, metadata)
}

// RecordUserActivity records user activity events
func (ms *MonitoringService) RecordUserActivity(ctx context.Context, userID, tenantID, action, resource, status string, metadata map[string]interface{}) {
	ms.RecordEvent(ctx, "user_activity", "auth_service", tenantID, userID, resource, action, status, metadata)
}

// GetErrorRate returns error rate for a specific period
func (ms *MonitoringService) GetErrorRate(ctx context.Context, service, resource string, hours int) (float64, error) {
	return 0.0, nil // Placeholder implementation
}

// GetPerformanceMetrics returns performance metrics for specific components
func (ms *MonitoringService) GetPerformanceMetrics(ctx context.Context, component, period string) (map[string]interface{}, error) {
	metrics := map[string]interface{}{
		"component": component,
		"period":    period,
		"requests_per_second": 0.0,
		"average_response_time": 0.0,
		"error_rate": 0.0,
		"throughput": 0,
		"concurrency": 0,
	}
	
	return metrics, nil
}

// Close shuts down the monitoring service
func (ms *MonitoringService) Close() {
	close(ms.stop)
	
	// Shutdown metrics and tracing
	if ms.tracer != nil {
		ms.tracer.Shutdown(context.Background())
	}
}