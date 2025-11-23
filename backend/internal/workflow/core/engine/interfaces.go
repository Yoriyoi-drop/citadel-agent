package engine

import (
	"context"

	"github.com/citadel-agent/backend/internal/workflow/core/types"
)

// Storage interface for storing and retrieving workflow data
type Storage interface {
	// Execution operations
	CreateExecution(execution *types.Execution) error
	UpdateExecution(execution *types.Execution) error
	GetExecution(id string) (*types.Execution, error)
	DeleteExecution(id string) error
	ListExecutions(workflowID string, limit, offset int) ([]*types.Execution, error)
	GetExecutionHistory(workflowID string, limit, offset int) ([]*types.Execution, error)
	GetLastExecution(workflowID string) (*types.Execution, error)
	GetRecentExecutions(limit int) ([]*types.Execution, error)
	GetExecutionCount(workflowID string) (int64, error)
	GetExecutionCountByStatus(workflowID string, status types.ExecutionStatus) (int64, error)

	// Node result operations
	CreateNodeResult(result *types.NodeResult) error
	UpdateNodeResult(result *types.NodeResult) error
	GetNodeResult(executionID, nodeID string) (*types.NodeResult, error)
	GetNodeResults(executionID string) (map[string]*types.NodeResult, error) // Get all node results for an execution
	DeleteNodeResult(id string) error
	ListNodeResults(executionID string, limit, offset int) ([]*types.NodeResult, error)

	// Workflow operations
	CreateWorkflow(workflow *types.Workflow) error
	UpdateWorkflow(workflow *types.Workflow) error
	GetWorkflow(id string) (*types.Workflow, error)
	DeleteWorkflow(id string) error
	ListWorkflows(limit, offset int) ([]*types.Workflow, error)
	GetWorkflowByName(name string) (*types.Workflow, error)

	// Variable operations
	GetVariable(executionID, key string) (interface{}, error)
	SetVariable(executionID, key string, value interface{}) error
	DeleteVariable(executionID, key string) error

	// Statistics operations
	GetWorkflowStatistics(workflowID string) (*types.WorkflowStatistics, error)
	GetExecutionStatistics(from, to string) (*types.WorkflowStatistics, error)
	GetNodeExecutionStats(nodeType string) (*types.WorkflowStatistics, error)

	// Cleanup operations
	CleanupExecutions(olderThanDays int) error
	CleanupNodeResults(olderThanDays int) error
	CleanupVariables(olderThanDays int) error

	// Batch operations
	BatchCreateExecutions(executions []*types.Execution) error
	BatchUpdateExecutions(executions []*types.Execution) error
	BatchDeleteExecutions(executionIDs []string) error

	// Index operations (for performance)
	IndexExecutionByStatus(status types.ExecutionStatus, workflowID string) error
	IndexExecutionByDate(dateRange string) error
	IndexExecutionByTrigger(triggerType string) error

	// Transaction support
	BeginTransaction() (Tx, error)
	InTransaction(fn func(Tx) error) error

	// Health check
	HealthCheck() error
}

// Tx represents a database transaction
type Tx interface {
	Storage
	Commit() error
	Rollback() error
}

// RetryManager manages retry logic for failed operations
type RetryManager struct{}

// CircuitBreakerManager manages circuit breakers for external calls
type CircuitBreakerManager struct{}

// Logger interface for logging
type Logger interface {
	Debug(msg string, fields ...map[string]interface{})
	Info(msg string, fields ...map[string]interface{})
	Warn(msg string, fields ...map[string]interface{})
	Error(msg string, fields ...map[string]interface{})
}

// Scheduler interface for scheduling jobs
type Scheduler interface {
	ScheduleJob(workflowID string, scheduleExpression string, triggerParams map[string]interface{}) error
	CancelJob(jobID string) error
	ListScheduledJobs(workflowID string) ([]*ScheduledJob, error)
}

// ScheduledJob represents a scheduled job
type ScheduledJob struct {
	ID            string                 `json:"id"`
	WorkflowID    string                 `json:"workflow_id"`
	ScheduleExpr  string                 `json:"schedule_expression"`
	TriggerParams map[string]interface{} `json:"trigger_params"`
	CreatedAt     int64                  `json:"created_at"`
	LastRunAt     *int64                 `json:"last_run_at,omitempty"`
	NextRunAt     int64                  `json:"next_run_at"`
	Status        string                 `json:"status"` // "active", "paused", "cancelled"
	Error         *string                `json:"error,omitempty"`
}

// AIManager interface for managing AI operations
type AIManager interface {
	GenerateText(ctx context.Context, prompt string, config map[string]interface{}) (string, error)
	ProcessImage(ctx context.Context, imageData []byte, config map[string]interface{}) (map[string]interface{}, error)
	TranscribeAudio(ctx context.Context, audioData []byte, config map[string]interface{}) (string, error)
}

// RuntimeValidator interface for validating runtime conditions
type RuntimeValidator interface {
	ValidateWorkflow(workflow *types.Workflow) error
	ValidateNode(node *types.Node, inputs map[string]interface{}) error
	ValidateExecutionConstraints(execution *types.Execution) error
}

// PermissionChecker interface for checking permissions
type PermissionChecker interface {
	CheckPermission(userID, resource, action string) (bool, error)
	HasResourceAccess(userID, resourceID string) (bool, error)
	ValidateAPIKey(apiKey string) (bool, string, error) // returns isValid, userID, error
}

// ResourceLimiter interface for managing resource limits
type ResourceLimiter interface {
	CheckResourceUsage(userID string, resourceType string) (bool, error)
	IncrementResourceUsage(userID, resourceType string, amount int64) error
	GetResourceQuota(userID, resourceType string) (int64, error)
}

// MetricsCollector interface for collecting workflow metrics
type MetricsCollector interface {
	RecordExecutionStart(workflowID, executionID string)
	RecordExecutionEnd(workflowID, executionID string, success bool, duration float64)
	RecordNodeExecution(nodeType, executionID string, success bool, duration float64)
	RecordError(workflowID, executionID, nodeID, errorType string)
	GetWorkflowMetrics(workflowID string) *WorkflowMetrics
	GetSystemMetrics() *SystemMetrics
}

// TraceCollector interface for collecting execution traces
type TraceCollector interface {
	StartSpan(operationName, executionID string) string // returns traceID
	EndSpan(traceID string, success bool, duration float64)
	AddEvent(traceID, eventName string, attributes map[string]interface{})
}

// Alerter interface for handling alerts
type Alerter interface {
	SendAlert(title, message string, severity string, metadata map[string]interface{}) error
	RegisterAlertHandler(handler AlertHandler) error
}

// AlertHandler interface for handling specific alert types
type AlertHandler interface {
	HandleAlert(alert *Alert) error
	CanHandle(alertType string) bool
}

// Alert represents an alert
type Alert struct {
	ID         string                 `json:"id"`
	Title      string                 `json:"title"`
	Message    string                 `json:"message"`
	Severity   string                 `json:"severity"`    // "info", "warning", "error", "critical"
	Type       string                 `json:"alert_type"`  // "execution_failure", "performance", "security", etc.
	ResourceID string                 `json:"resource_id"` // Workflow ID, execution ID, etc.
	Timestamp  int64                  `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata"`
	Handled    bool                   `json:"handled"`
	HandledAt  *int64                 `json:"handled_at,omitempty"`
}

// WorkflowMetrics represents workflow execution metrics
type WorkflowMetrics struct {
	TotalExecutions      int64   `json:"total_executions"`
	SuccessfulExecutions int64   `json:"successful_executions"`
	FailedExecutions     int64   `json:"failed_executions"`
	AverageExecutionTime float64 `json:"average_execution_time"` // in seconds
	AverageNodeTime      float64 `json:"average_node_time"`      // in seconds
	LastErrorTime        *int64  `json:"last_error_time,omitempty"`
}

// SystemMetrics represents system-level metrics
type SystemMetrics struct {
	ActiveExecutions int64   `json:"active_executions"`
	MemoryUsage      float64 `json:"memory_usage_mb"`
	CPUUsage         float64 `json:"cpu_usage_percent"`
	Uptime           float64 `json:"uptime_seconds"`
	RequestRate      float64 `json:"request_rate_per_second"`
	ErrorRate        float64 `json:"error_rate_per_second"`
}
