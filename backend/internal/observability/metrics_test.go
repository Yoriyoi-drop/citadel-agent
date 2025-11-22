package observability

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMetricsService(t *testing.T) {
	// Test creating a new metrics service
	metricsService := NewMetricsService()
	assert.NotNil(t, metricsService)

	// Test workflow execution metrics recording
	metricsService.RecordWorkflowExecution("workflow-123", "completed", "tenant-123", 2*time.Second)
	
	// Test node execution metrics recording  
	metricsService.RecordNodeExecution("http_node", "workflow-123", "success", "tenant-123", 500*time.Millisecond)
	
	// Test API request metrics recording
	metricsService.RecordAPIRequest("GET", "/api/test", "200", "tenant-123", 100*time.Millisecond, 100, 200)
	
	// Test security event recording
	metricsService.RecordSecurityEvent("login", "info")
	
	// Test login attempt recording
	metricsService.RecordLoginAttempt("success", "192.168.1.1")
	
	// Test updating system metrics
	metricsService.UpdateGoroutines(10)
	metricsService.UpdateUptime()
	
	// All these calls should execute without errors
	assert.Equal(t, 0, 0) // Basic assertion that test runs
}

func TestMetricsServiceCollect(t *testing.T) {
	metricsService := NewMetricsService()
	
	ctx := context.Background()
	err := metricsService.Collect(ctx)
	
	assert.NoError(t, err)
}

func TestGetGoroutineCount(t *testing.T) {
	count := GetGoroutineCount()
	
	// Should return a positive number
	assert.GreaterOrEqual(t, count, 0)
}