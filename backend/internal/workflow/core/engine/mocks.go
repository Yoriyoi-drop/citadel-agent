package engine

import (
	"context"

	"github.com/citadel-agent/backend/internal/workflow/core/types"
)

// Mock implementations for interface constructors
// These will be replaced with actual implementations later

// MockRuntimeValidator is a placeholder implementation
type MockRuntimeValidator struct{}

func NewRuntimeValidator() RuntimeValidator {
	return &MockRuntimeValidator{}
}

func (m *MockRuntimeValidator) ValidateWorkflow(workflow *types.Workflow) error {
	return nil
}

func (m *MockRuntimeValidator) ValidateNode(node *types.Node, inputs map[string]interface{}) error {
	return nil
}

func (m *MockRuntimeValidator) ValidateExecutionConstraints(execution *types.Execution) error {
	return nil
}

// MockPermissionChecker is a placeholder implementation
type MockPermissionChecker struct{}

func NewPermissionChecker() PermissionChecker {
	return &MockPermissionChecker{}
}

func (m *MockPermissionChecker) CheckPermission(userID, resource, action string) (bool, error) {
	return true, nil
}

func (m *MockPermissionChecker) HasResourceAccess(userID, resourceID string) (bool, error) {
	return true, nil
}

func (m *MockPermissionChecker) ValidateAPIKey(apiKey string) (bool, string, error) {
	return true, "mock-user", nil
}

// MockResourceLimiter is a placeholder implementation
type MockResourceLimiter struct{}

func NewResourceLimiter() ResourceLimiter {
	return &MockResourceLimiter{}
}

func (m *MockResourceLimiter) CheckResourceUsage(userID string, resourceType string) (bool, error) {
	return true, nil
}

func (m *MockResourceLimiter) IncrementResourceUsage(userID, resourceType string, amount int64) error {
	return nil
}

func (m *MockResourceLimiter) GetResourceQuota(userID, resourceType string) (int64, error) {
	return 1000000, nil
}

// MockMetricsCollector is a placeholder implementation
type MockMetricsCollector struct{}

func NewMetricsCollector() MetricsCollector {
	return &MockMetricsCollector{}
}

func (m *MockMetricsCollector) RecordExecutionStart(workflowID, executionID string) {}

func (m *MockMetricsCollector) RecordExecutionEnd(workflowID, executionID string, success bool, duration float64) {
}

func (m *MockMetricsCollector) RecordNodeExecution(nodeType, executionID string, success bool, duration float64) {
}

func (m *MockMetricsCollector) RecordError(workflowID, executionID, nodeID, errorType string) {}

func (m *MockMetricsCollector) GetWorkflowMetrics(workflowID string) *WorkflowMetrics {
	return &WorkflowMetrics{}
}

func (m *MockMetricsCollector) GetSystemMetrics() *SystemMetrics {
	return &SystemMetrics{}
}

// MockTraceCollector is a placeholder implementation
type MockTraceCollector struct{}

func NewTraceCollector() TraceCollector {
	return &MockTraceCollector{}
}

func (m *MockTraceCollector) StartSpan(operationName, executionID string) string {
	return "mock-trace-id"
}

func (m *MockTraceCollector) EndSpan(traceID string, success bool, duration float64) {}

func (m *MockTraceCollector) AddEvent(traceID, eventName string, attributes map[string]interface{}) {}

// MockAlerter is a placeholder implementation
type MockAlerter struct{}

func NewAlerter() Alerter {
	return &MockAlerter{}
}

func (m *MockAlerter) SendAlert(title, message string, severity string, metadata map[string]interface{}) error {
	return nil
}

func (m *MockAlerter) RegisterAlertHandler(handler AlertHandler) error {
	return nil
}

// MockAIManager is a placeholder implementation
type MockAIManager struct{}

func NewAIManager() AIManager {
	return &MockAIManager{}
}

func (m *MockAIManager) GenerateText(ctx context.Context, prompt string, config map[string]interface{}) (string, error) {
	return "mock AI response", nil
}

func (m *MockAIManager) ProcessImage(ctx context.Context, imageData []byte, config map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{"result": "mock image processing"}, nil
}

func (m *MockAIManager) TranscribeAudio(ctx context.Context, audioData []byte, config map[string]interface{}) (string, error) {
	return "mock transcription", nil
}
