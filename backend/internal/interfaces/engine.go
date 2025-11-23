package interfaces

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ExecutionResult represents the result of a node execution
type ExecutionResult struct {
	Status    string      `json:"status"`
	Data      interface{} `json:"data"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// NodeRegistry interface for managing node definitions
type NodeRegistry interface {
	RegisterNodeType(nodeType string, constructor func(map[string]interface{}) (NodeInstance, error)) error
	CreateInstance(nodeType string, config map[string]interface{}) (NodeInstance, error)
	ListNodeTypes() []string
	GetNodeDefinition(nodeType string) (*NodeDefinition, bool)
}

// Concrete implementation of NodeRegistry
type ConcreteNodeRegistry struct {
	nodes       map[string]func(map[string]interface{}) (NodeInstance, error)
	definitions map[string]*NodeDefinition
	mutex       sync.RWMutex
}

// NewNodeRegistry creates a new node registry
func NewNodeRegistry() *ConcreteNodeRegistry {
	return &ConcreteNodeRegistry{
		nodes:       make(map[string]func(map[string]interface{}) (NodeInstance, error)),
		definitions: make(map[string]*NodeDefinition),
	}
}

// RegisterNodeType registers a new node type with its constructor
func (r *ConcreteNodeRegistry) RegisterNodeType(nodeType string, constructor func(map[string]interface{}) (NodeInstance, error)) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.nodes[nodeType] = constructor
	return nil
}

// CreateInstance creates a new instance of the specified node type
func (r *ConcreteNodeRegistry) CreateInstance(nodeType string, config map[string]interface{}) (NodeInstance, error) {
	r.mutex.RLock()
	constructor, exists := r.nodes[nodeType]
	r.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("node type %s not registered", nodeType)
	}
	return constructor(config)
}

// ListNodeTypes returns all registered node types
func (r *ConcreteNodeRegistry) ListNodeTypes() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	types := make([]string, 0, len(r.nodes))
	for nodeType := range r.nodes {
		types = append(types, nodeType)
	}
	return types
}

// GetNodeDefinition returns the definition for a node type
func (r *ConcreteNodeRegistry) GetNodeDefinition(nodeType string) (*NodeDefinition, bool) {
	r.mutex.RLock()
	def, exists := r.definitions[nodeType]
	r.mutex.RUnlock()
	return def, exists
}

// RegisterNodeDefinition registers the definition for a node type
func (r *ConcreteNodeRegistry) RegisterNodeDefinition(definition *NodeDefinition) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.definitions[definition.Type] = definition
}

// UnregisterNodeType removes a node type from the registry
func (r *ConcreteNodeRegistry) UnregisterNodeType(nodeType string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.nodes, nodeType)
	delete(r.definitions, nodeType)
}

// CountNodeTypes returns the total number of registered node types
func (r *ConcreteNodeRegistry) CountNodeTypes() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return len(r.nodes)
}

// GetAllNodeDefinitions returns all registered node definitions
func (r *ConcreteNodeRegistry) GetAllNodeDefinitions() map[string]*NodeDefinition {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	definitions := make(map[string]*NodeDefinition, len(r.definitions))
	for k, v := range r.definitions {
		definitions[k] = v
	}
	return definitions
}

// Engine interface for workflow execution engine
type Engine interface {
	ExecuteWorkflow(ctx context.Context, workflow *Workflow, triggerParams map[string]interface{}) (string, error)
	GetExecution(id string) (*Execution, error)
	GetNodeRegistry() NodeRegistry
	RegisterNode(nodeType string, constructor func(map[string]interface{}) (NodeInstance, error)) error
	ListNodeTypes() []string
	GetNodeDefinition(nodeType string) (*NodeDefinition, bool)
	GetMetrics() map[string]interface{}
	HealthCheck() map[string]interface{}
}

// Workflow represents a workflow definition
type Workflow struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Version     int                     `json:"version"`
	Nodes       []*Node                 `json:"nodes"`
	Connections []*Connection          `json:"connections"`
	Config      map[string]interface{} `json:"config"`
	Variables   map[string]interface{} `json:"variables"`
	Status      WorkflowStatus         `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	DeletedAt   *time.Time             `json:"deleted_at,omitempty"`
}

// Node represents a single node in a workflow
type Node struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Name          string                 `json:"name"`
	Label         string                 `json:"label"`
	Description   string                 `json:"description"`
	Config        map[string]interface{} `json:"config"`
	Dependencies  []string               `json:"dependencies"`
	Inputs        map[string]interface{} `json:"inputs"`
	Outputs       map[string]interface{} `json:"outputs"`
	Status        NodeStatus             `json:"status"`
	Position      Position               `json:"position"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// Connection represents a connection between nodes
type Connection struct {
	ID           string `json:"id"`
	SourceNodeID string `json:"source_node_id"`
	TargetNodeID string `json:"target_node_id"`
	SourceHandle string `json:"source_handle,omitempty"`
	TargetHandle string `json:"target_handle,omitempty"`
	Type         string `json:"type,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
}

// Execution represents an execution instance of a workflow
type Execution struct {
	ID            string                 `json:"id"`
	WorkflowID    string                 `json:"workflow_id"`
	Status        ExecutionStatus        `json:"status"`
	StartedAt     time.Time              `json:"started_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	Variables     map[string]interface{} `json:"variables"`
	NodeResults   map[string]*NodeResult `json:"node_results"`
	Error         *string                `json:"error,omitempty"`
	TriggeredBy   string                 `json:"triggered_by"`
	TriggerParams map[string]interface{} `json:"trigger_params"`
	ExecutionTime time.Duration          `json:"execution_time"`
	CancelledAt   *time.Time             `json:"cancelled_at,omitempty"`
}

// NodeResult represents the result of a single node execution during a workflow execution
type NodeResult struct {
	ID            string                 `json:"id"`
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

// Position represents the position of a node in the visual workflow
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// WorkflowStatus represents the status of a workflow
type WorkflowStatus string

const (
	WorkflowDraft     WorkflowStatus = "draft"
	WorkflowActive    WorkflowStatus = "active"
	WorkflowInactive  WorkflowStatus = "inactive"
	WorkflowArchived  WorkflowStatus = "archived"
	WorkflowDeleting  WorkflowStatus = "deleting"
)

// ExecutionStatus represents the status of an execution
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

// NodeStatus represents the status of a node
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


// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
	Value   interface{} `json:"value"`
}

// WorkflowValidationError represents validation errors for a workflow
type WorkflowValidationError struct {
	Errors []ValidationError `json:"errors"`
}

func (e *WorkflowValidationError) Error() string {
	if len(e.Errors) == 0 {
		return "workflow validation failed"
	}
	return fmt.Sprintf("workflow validation failed: %s", e.Errors[0].Message)
}