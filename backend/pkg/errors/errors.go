package errors

import "fmt"

// WorkflowError represents a workflow-specific error
type WorkflowError struct {
	Code    string
	Message string
	Cause   error
}

// Error returns the error message
func (e *WorkflowError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *WorkflowError) Unwrap() error {
	return e.Cause
}

// NewWorkflowError creates a new workflow error
func NewWorkflowError(code, message string) *WorkflowError {
	return &WorkflowError{
		Code:    code,
		Message: message,
	}
}

// WrapWorkflowError wraps an existing error with workflow context
func WrapWorkflowError(code, message string, cause error) *WorkflowError {
	return &WorkflowError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// Common error codes
const (
	NodeInitializationError = "NODE_INIT_ERROR"
	NodeExecutionError      = "NODE_EXEC_ERROR"
	WorkflowValidationError = "WORKFLOW_VALIDATION_ERROR"
	ConnectionError         = "CONNECTION_ERROR"
	TimeoutError            = "TIMEOUT_ERROR"
)

// Helper functions for common error types
func NewNodeInitializationError(message string, cause error) *WorkflowError {
	return WrapWorkflowError(NodeInitializationError, message, cause)
}

func NewNodeExecutionError(message string, cause error) *WorkflowError {
	return WrapWorkflowError(NodeExecutionError, message, cause)
}

func NewWorkflowValidationError(message string) *WorkflowError {
	return NewWorkflowError(WorkflowValidationError, message)
}

func NewConnectionError(message string, cause error) *WorkflowError {
	return WrapWorkflowError(ConnectionError, message, cause)
}

func NewTimeoutError(message string) *WorkflowError {
	return NewWorkflowError(TimeoutError, message)
}