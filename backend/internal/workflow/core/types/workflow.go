package types

import (
	"time"
)

// Workflow represents a complete workflow definition
type Workflow struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     int                    `json:"version"`
	Nodes       []*Node                `json:"nodes"`
	Connections []*Connection          `json:"connections"`
	Config      map[string]interface{} `json:"config"`
	Variables   map[string]interface{} `json:"variables"`
	Status      WorkflowStatus         `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	DeletedAt   *time.Time             `json:"deleted_at,omitempty"`
}

// Node represents a single node in the workflow
type Node struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Label       string                 `json:"label"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	Inputs      map[string]interface{} `json:"inputs"`
	Outputs     map[string]interface{} `json:"outputs"`
	Position    Position               `json:"position"`
	Dependencies []string              `json:"dependencies"`
	Status      NodeStatus             `json:"status"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Error       *string                `json:"error,omitempty"`
}

// Position represents the position of a node in the visual workflow
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Connection represents a connection between nodes
type Connection struct {
	ID           string `json:"id"`
	SourceNodeID string `json:"source_node_id"`
	TargetNodeID string `json:"target_node_id"`
	SourceHandle string `json:"source_handle,omitempty"` // Port name
	TargetHandle string `json:"target_handle,omitempty"` // Port name
	Type         string `json:"type,omitempty"`          // Connection type
	Data         map[string]interface{} `json:"data,omitempty"` // Additional connection data
}

// Execution represents a single execution of a workflow
type Execution struct {
	ID            string                 `json:"id" gorm:"primaryKey"`
	WorkflowID    string                 `json:"workflow_id"`
	Status        ExecutionStatus        `json:"status"`
	StartedAt     time.Time              `json:"started_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	Variables     map[string]interface{} `json:"variables"`
	NodeResults   map[string]*NodeResult `json:"node_results"`
	Error         *string                `json:"error,omitempty"`
	TriggeredBy   string                 `json:"triggered_by"`
	TriggerParams map[string]interface{} `json:"trigger_params,omitempty"`
	ExecutionTime time.Duration          `json:"execution_time,omitempty"`
	Retries       int                    `json:"retries"`
	ParentID      *string                `json:"parent_id,omitempty"` // For sub-workflows
	CancelledAt   *time.Time             `json:"cancelled_at,omitempty"`
}

// NodeResult represents the result of a single node execution
type NodeResult struct {
	ID            string                 `json:"id" gorm:"primaryKey"`
	ExecutionID   string                 `json:"execution_id"`
	NodeID        string                 `json:"node_id"`
	Status        NodeStatus             `json:"status"`
	Output        map[string]interface{} `json:"output"`
	Error         *string                `json:"error,omitempty"`
	StartedAt     time.Time              `json:"started_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	ExecutionTime time.Duration          `json:"execution_time"`
	RetryCount    int                    `json:"retry_count"`
	InputsUsed    map[string]interface{} `json:"inputs_used"`
	OutputsCached bool                   `json:"outputs_cached"`
}

// WorkflowStatus represents the status of a workflow definition
type WorkflowStatus string

const (
	WorkflowDraft     WorkflowStatus = "draft"
	WorkflowActive    WorkflowStatus = "active"
	WorkflowInactive  WorkflowStatus = "inactive"
	WorkflowArchived  WorkflowStatus = "archived"
	WorkflowDeleting  WorkflowStatus = "deleting"
)

// ExecutionStatus represents the status of a workflow execution
type ExecutionStatus string

const (
	ExecutionCreated    ExecutionStatus = "created"
	ExecutionQueued     ExecutionStatus = "queued"
	ExecutionRunning    ExecutionStatus = "running"
	ExecutionPaused     ExecutionStatus = "paused"
	ExecutionResuming   ExecutionStatus = "resuming"
	ExecutionCancelled  ExecutionStatus = "cancelled"
	ExecutionFailed     ExecutionStatus = "failed"
	ExecutionSucceeded  ExecutionStatus = "succeeded"
	ExecutionTimeout    ExecutionStatus = "timeout"
	ExecutionRetrying   ExecutionStatus = "retrying"
)

// NodeStatus represents the status of a node execution
type NodeStatus string

const (
	NodeScheduled    NodeStatus = "scheduled"
	NodePending      NodeStatus = "pending"
	NodeRunning      NodeStatus = "running"
	NodeCompleted    NodeStatus = "completed"
	NodeFailed       NodeStatus = "failed"
	NodeSkipped      NodeStatus = "skipped"
	NodeCancelled    NodeStatus = "cancelled"
	NodeTimeout      NodeStatus = "timeout"
	NodeRetrying     NodeStatus = "retrying"
	NodeInterrupted  NodeStatus = "interrupted"
)

// TriggerType represents how a workflow execution was triggered
type TriggerType string

const (
	TriggerManual     TriggerType = "manual"
	TriggerSchedule   TriggerType = "schedule"
	TriggerWebhook    TriggerType = "webhook"
	TriggerEvent      TriggerType = "event"
	TriggerAPI        TriggerType = "api"
	TriggerOther      TriggerType = "other"
)

// WorkflowStatistics holds statistics for a workflow
type WorkflowStatistics struct {
	TotalExecutions     int64     `json:"total_executions"`
	SuccessfulExecutions int64     `json:"successful_executions"`
	FailedExecutions    int64     `json:"failed_executions"`
	AverageExecutionTime time.Duration `json:"average_execution_time"`
	LastExecutionAt     *time.Time `json:"last_execution_at,omitempty"`
	CurrentExecutions   int       `json:"current_executions"`
	LastExecutionStatus ExecutionStatus `json:"last_execution_status"`
}

// NodeMetadata holds metadata about a node
type NodeMetadata struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Category    string                 `json:"category"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Author      string                 `json:"author"`
	Icon        string                 `json:"icon"`
	Tags        []string               `json:"tags"`
	Inputs      map[string]interface{} `json:"inputs"`
	Outputs     map[string]interface{} `json:"outputs"`
	Settings    map[string]interface{} `json:"settings"`
}

// ConnectionMetadata holds metadata about a connection
type ConnectionMetadata struct {
	PortType    string                 `json:"port_type"`      // "input" or "output"
	DataType    string                 `json:"data_type"`      // "any", "string", "number", etc.
	IsRequired  bool                   `json:"is_required"`    // Whether this connection is required
	Label       string                 `json:"label"`          // Label for the connection
	Schema      map[string]interface{} `json:"schema"`         // JSON schema for validation
	Validation  map[string]interface{} `json:"validation"`     // Validation rules
	MaxConnections int                `json:"max_connections"` // Maximum number of connections allowed
}

// ValidationError represents an error during validation
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
	Value   interface{} `json:"value"`
}

// WorkflowValidationError represents a collection of validation errors
type WorkflowValidationError struct {
	Errors []ValidationError `json:"errors"`
}

func (e *WorkflowValidationError) Error() string {
	return "workflow validation failed"
}