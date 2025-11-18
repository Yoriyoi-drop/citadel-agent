package engine

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ExecutionContext holds the execution context for a workflow run
type ExecutionContext struct {
	// Unique ID for this execution context
	ID string

	// Reference to the workflow being executed
	Workflow *Workflow

	// The execution being processed
	Execution *Execution

	// Shared variables across nodes in this execution
	Variables map[string]interface{}

	// Timestamps
	StartedAt time.Time
	EndedAt   *time.Time

	// Context for cancellation and timeouts
	Ctx context.Context
	Cancel context.CancelFunc

	// Current execution state
	State ExecutionState
}

// ExecutionState represents the current state of an execution
type ExecutionState struct {
	// Current node being executed
	CurrentNodeID string

	// Results from previously executed nodes
	NodeResults map[string]*ExecutionResult

	// Execution statistics
	ProcessedNodes int
	TotalNodes     int

	// Error tracking
	LastError string
}

// NewExecutionContext creates a new execution context
func NewExecutionContext(ctx context.Context, workflow *Workflow, variables map[string]interface{}) *ExecutionContext {
	// Create a cancellable context
	ctx, cancel := context.WithCancel(ctx)

	return &ExecutionContext{
		ID:          uuid.New().String(),
		Workflow:    workflow,
		Variables:   variables,
		StartedAt:   time.Now(),
		Ctx:         ctx,
		Cancel:      cancel,
		NodeResults: make(map[string]*ExecutionResult),
		State: ExecutionState{
			NodeResults: make(map[string]*ExecutionResult),
			TotalNodes:  len(workflow.Nodes),
		},
	}
}

// UpdateVariable updates a variable in the execution context
func (ec *ExecutionContext) UpdateVariable(key string, value interface{}) {
	if ec.Variables == nil {
		ec.Variables = make(map[string]interface{})
	}
	ec.Variables[key] = value
}

// GetVariable retrieves a variable from the execution context
func (ec *ExecutionContext) GetVariable(key string) (interface{}, bool) {
	if ec.Variables == nil {
		return nil, false
	}
	value, exists := ec.Variables[key]
	return value, exists
}

// AddNodeResult adds a result from a node execution
func (ec *ExecutionContext) AddNodeResult(nodeID string, result *ExecutionResult) {
	ec.State.NodeResults[nodeID] = result
	ec.State.ProcessedNodes++
}

// GetNodeResult retrieves a result from a previously executed node
func (ec *ExecutionContext) GetNodeResult(nodeID string) (*ExecutionResult, bool) {
	result, exists := ec.State.NodeResults[nodeID]
	return result, exists
}

// IsCancelled checks if the execution has been cancelled
func (ec *ExecutionContext) IsCancelled() bool {
	select {
	case <-ec.Ctx.Done():
		return true
	default:
		return false
	}
}

// Cancel cancels the execution context
func (ec *ExecutionContext) Cancel() {
	ec.Cancel()
}

// Complete marks the execution as completed
func (ec *ExecutionContext) Complete() {
	endTime := time.Now()
	ec.EndedAt = &endTime
}

// HasFailed checks if the execution has failed
func (ec *ExecutionContext) HasFailed() bool {
	return ec.State.LastError != ""
}

// SetError sets an error for the execution
func (ec *ExecutionContext) SetError(err string) {
	ec.State.LastError = err
}

// GetExecutionResult returns the final execution result
func (ec *ExecutionContext) GetExecutionResult() *Execution {
	return &Execution{
		ID:         ec.ID,
		WorkflowID: ec.Workflow.ID,
		Status:     getExecutionStatus(ec),
		StartedAt:  ec.StartedAt,
		EndedAt:    ec.EndedAt,
		Results:    convertResults(ec.State.NodeResults),
		Error:      ec.State.LastError,
		Variables:  ec.Variables,
	}
}

// Helper function to determine execution status
func getExecutionStatus(ec *ExecutionContext) string {
	if ec.IsCancelled() {
		return "cancelled"
	}
	if ec.HasFailed() {
		return "failed"
	}
	if ec.EndedAt != nil {
		return "completed"
	}
	return "running"
}

// Helper function to convert ExecutionResult map for final output
func convertResults(results map[string]*ExecutionResult) map[string]interface{} {
	output := make(map[string]interface{})
	for nodeID, result := range results {
		output[nodeID] = result
	}
	return output
}