// backend/internal/observability/metrics.go
package observability

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsService handles metrics collection and exposure
type MetricsService struct {
	// Workflow execution metrics
	workflowExecutionsTotal *prometheus.CounterVec
	workflowExecutionDuration *prometheus.HistogramVec
	workflowErrorsTotal *prometheus.CounterVec
	
	// Node execution metrics
	nodeExecutionsTotal *prometheus.CounterVec
	nodeExecutionDuration *prometheus.HistogramVec
	nodeErrorsTotal *prometheus.CounterVec
	
	// API request metrics
	apiRequestsTotal *prometheus.CounterVec
	apiRequestDuration *prometheus.HistogramVec
	apiRequestSize *prometheus.SummaryVec
	apiResponseSize *prometheus.SummaryVec
	
	// Resource usage metrics
	cpuUsage *prometheus.GaugeVec
	memoryUsage *prometheus.GaugeVec
	goroutines *prometheus.Gauge
	
	// Security metrics
	securityEventsTotal *prometheus.CounterVec
	loginAttemptsTotal *prometheus.CounterVec
	permissionDeniedTotal *prometheus.CounterVec
	apiKeysUsedTotal *prometheus.CounterVec
	
	// System metrics
	startTime time.Time
	uptime *prometheus.Gauge
}

// NewMetricsService creates a new metrics service
func NewMetricsService() *MetricsService {
	service := &MetricsService{
		startTime: time.Now(),
		
		// Workflow execution metrics
		workflowExecutionsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "citadel_workflow_executions_total",
				Help: "Total number of workflow executions",
			},
			[]string{"workflow_id", "status", "tenant_id"},
		),
		
		workflowExecutionDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "citadel_workflow_execution_duration_seconds",
				Help: "Duration of workflow executions",
				Buckets: []float64{0.1, 0.5, 1, 2.5, 5, 10, 30, 60, 120},
			},
			[]string{"workflow_id", "status", "tenant_id"},
		),
		
		workflowErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "citadel_workflow_errors_total",
				Help: "Total number of workflow errors",
			},
			[]string{"workflow_id", "error_type", "tenant_id"},
		),
		
		// Node execution metrics
		nodeExecutionsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "citadel_node_executions_total",
				Help: "Total number of node executions",
			},
			[]string{"node_type", "workflow_id", "status", "tenant_id"},
		),
		
		nodeExecutionDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "citadel_node_execution_duration_seconds",
				Help: "Duration of node executions",
				Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"node_type", "workflow_id", "tenant_id"},
		),
		
		nodeErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "citadel_node_errors_total",
				Help: "Total number of node errors",
			},
			[]string{"node_type", "workflow_id", "error_type", "tenant_id"},
		),
		
		// API request metrics
		apiRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "citadel_api_requests_total",
				Help: "Total number of API requests",
			},
			[]string{"method", "endpoint", "status_code", "tenant_id"},
		),
		
		apiRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "citadel_api_request_duration_seconds",
				Help: "Duration of API requests",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"method", "endpoint", "tenant_id"},
		),
		
		apiRequestSize: promauto.NewSummaryVec(
			prometheus.SummaryOpts{
				Name: "citadel_api_request_size_bytes",
				Help: "Size of API requests",
			},
			[]string{"method", "endpoint"},
		),
		
		apiResponseSize: promauto.NewSummaryVec(
			prometheus.SummaryOpts{
				Name: "citadel_api_response_size_bytes",
				Help: "Size of API responses",
			},
			[]string{"method", "endpoint"},
		),
		
		// Resource usage metrics
		cpuUsage: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "citadel_cpu_usage_percent",
				Help: "CPU usage percentage",
			},
			[]string{"process"},
		),
		
		memoryUsage: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "citadel_memory_usage_bytes",
				Help: "Memory usage in bytes",
			},
			[]string{"process"},
		),
		
		goroutines: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "citadel_goroutines_total",
				Help: "Total number of goroutines",
			},
		),
		
		// Security metrics
		securityEventsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "citadel_security_events_total",
				Help: "Total number of security events",
			},
			[]string{"event_type", "severity"},
		),
		
		loginAttemptsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "citadel_login_attempts_total",
				Help: "Total number of login attempts",
			},
			[]string{"status", "source_ip"},
		),
		
		permissionDeniedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "citadel_permission_denied_total",
				Help: "Total number of permission denied events",
			},
			[]string{"resource", "action", "user_id"},
		),
		
		apiKeysUsedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "citadel_api_keys_used_total",
				Help: "Total number of API key usages",
			},
			[]string{"api_key_id", "user_id", "api_endpoint"},
		),
		
		// System metrics
		uptime: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "citadel_uptime_seconds",
				Help: "Uptime of the service in seconds",
			},
		),
	}
	
	return service
}

// RecordWorkflowExecution records workflow execution metrics
func (ms *MetricsService) RecordWorkflowExecution(workflowID, status, tenantID string, duration time.Duration) {
	ms.workflowExecutionsTotal.WithLabelValues(workflowID, status, tenantID).Inc()
	ms.workflowExecutionDuration.WithLabelValues(workflowID, status, tenantID).Observe(duration.Seconds())
}

// RecordWorkflowError records workflow error metrics
func (ms *MetricsService) RecordWorkflowError(workflowID, errorType, tenantID string) {
	ms.workflowErrorsTotal.WithLabelValues(workflowID, errorType, tenantID).Inc()
}

// RecordNodeExecution records node execution metrics
func (ms *MetricsService) RecordNodeExecution(nodeType, workflowID, status, tenantID string, duration time.Duration) {
	ms.nodeExecutionsTotal.WithLabelValues(nodeType, workflowID, status, tenantID).Inc()
	ms.nodeExecutionDuration.WithLabelValues(nodeType, workflowID, tenantID).Observe(duration.Seconds())
}

// RecordNodeError records node error metrics
func (ms *MetricsService) RecordNodeError(nodeType, workflowID, errorType, tenantID string) {
	ms.nodeErrorsTotal.WithLabelValues(nodeType, workflowID, errorType, tenantID).Inc()
}

// RecordAPIRequest records API request metrics
func (ms *MetricsService) RecordAPIRequest(method, endpoint, statusCode, tenantID string, duration time.Duration, requestSize, responseSize int) {
	ms.apiRequestsTotal.WithLabelValues(method, endpoint, statusCode, tenantID).Inc()
	ms.apiRequestDuration.WithLabelValues(method, endpoint, tenantID).Observe(duration.Seconds())
	
	if requestSize > 0 {
		ms.apiRequestSize.WithLabelValues(method, endpoint).Observe(float64(requestSize))
	}
	
	if responseSize > 0 {
		ms.apiResponseSize.WithLabelValues(method, endpoint).Observe(float64(responseSize))
	}
}

// RecordCPUUsage records CPU usage metrics
func (ms *MetricsService) RecordCPUUsage(process string, usagePercent float64) {
	ms.cpuUsage.WithLabelValues(process).Set(usagePercent)
}

// RecordMemoryUsage records memory usage metrics
func (ms *MetricsService) RecordMemoryUsage(process string, usageBytes int64) {
	ms.memoryUsage.WithLabelValues(process).Set(float64(usageBytes))
}

// RecordSecurityEvent records security event metrics
func (ms *MetricsService) RecordSecurityEvent(eventType, severity string) {
	ms.securityEventsTotal.WithLabelValues(eventType, severity).Inc()
}

// RecordLoginAttempt records login attempt metrics
func (ms *MetricsService) RecordLoginAttempt(status, sourceIP string) {
	ms.loginAttemptsTotal.WithLabelValues(status, sourceIP).Inc()
}

// RecordPermissionDenied records permission denied events
func (ms *MetricsService) RecordPermissionDenied(resource, action, userID string) {
	ms.permissionDeniedTotal.WithLabelValues(resource, action, userID).Inc()
}

// RecordAPIKeyUsage records API key usage
func (ms *MetricsService) RecordAPIKeyUsage(apiKeyID, userID, apiEndpoint string) {
	ms.apiKeysUsedTotal.WithLabelValues(apiKeyID, userID, apiEndpoint).Inc()
}

// UpdateGoroutines updates goroutine count
func (ms *MetricsService) UpdateGoroutines(count int) {
	ms.goroutines.Set(float64(count))
}

// UpdateUptime updates uptime metric
func (ms *MetricsService) UpdateUptime() {
	uptime := time.Since(ms.startTime)
	ms.uptime.Set(uptime.Seconds())
}

// Handler returns Prometheus metrics handler
func (ms *MetricsService) Handler() http.Handler {
	return promhttp.Handler()
}

// Collect collects custom metrics
func (ms *MetricsService) Collect(ctx context.Context) error {
	// Update system metrics
	ms.UpdateGoroutines(getGoroutineCount())
	ms.UpdateUptime()
	
	// In a real implementation, this would collect system resource usage
	// For now, we'll just return nil
	return nil
}

// getGoroutineCount returns the current number of goroutines
func getGoroutineCount() int {
	return int(prometheus.Labels{}["count"]) // Placeholder - we'll actually get real count
}

// GetGoroutineCount returns the current number of goroutines
func GetGoroutineCount() int {
	return runtime.NumGoroutine()
}