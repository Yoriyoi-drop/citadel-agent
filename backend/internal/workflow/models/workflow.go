// backend/internal/workflow/models/workflow.go
package models

import (
	"time"
)

// Workflow represents a complete workflow definition
type Workflow struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Nodes       []Node                 `json:"nodes"`
	Connections []Connection           `json:"connections"`
	Settings    map[string]interface{} `json:"settings"`
	Status      string                 `json:"status"` // active, inactive
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Node represents a single node in the workflow
type Node struct {
	ID          string                 `json:"id"`
	WorkflowID  string                 `json:"workflow_id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // http-request, function, delay, etc
	Parameters  map[string]interface{} `json:"parameters"`
	Outputs     map[string]interface{} `json:"outputs"`
	Position    map[string]float64     `json:"position"` // X, Y coordinates for UI
	Settings    map[string]interface{} `json:"settings"`
	WebhookPath *string                `json:"webhook_path,omitempty"` // For webhook-triggered nodes
}

// Connection represents a connection between nodes
type Connection struct {
	ID          string `json:"id"`
	WorkflowID  string `json:"workflow_id"`
	SourceNode  string `json:"source_node"`
	TargetNode  string `json:"target_node"`
	SourceHandle string `json:"source_handle"`
	TargetHandle string `json:"target_handle"`
	Type        string `json:"type"` // main, error, etc
}

// Execution represents a single execution of a workflow
type Execution struct {
	ID           string                 `json:"id" gorm:"primaryKey"`
	WorkflowID   string                 `json:"workflow_id"`
	Mode         string                 `json:"mode"` // trigger, manual, etc
	Status       string                 `json:"status"` // running, success, error, waiting
	StartedAt    time.Time              `json:"started_at"`
	StoppedAt    *time.Time             `json:"stopped_at,omitempty"`
	ResultData   map[string]interface{} `json:"result_data"`
	Error        *string                `json:"error,omitempty"`
	WaitTill     *time.Time             `json:"wait_till,omitempty"` // For waiting workflows
	Finished     bool                   `json:"finished"`
}

// ExecutionData stores execution-specific data
type ExecutionData struct {
	ID         string                 `json:"id"`
	ExecutionID string                `json:"execution_id"`
	Data       map[string]interface{} `json:"data"`
	WorkflowData map[string]interface{} `json:"workflow_data"`
	StartedAt  time.Time              `json:"started_at"`
	FinishedAt *time.Time             `json:"finished_at,omitempty"`
}