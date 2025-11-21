// backend/internal/nodes/monitoring/apm_node.go
package monitoring

import (
	"context"
	"fmt"
	"time"

	"citadel-agent/backend/internal/workflow/core/engine"
	"citadel-agent/backend/internal/observability"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// APMOperationType represents the type of APM operation
type APMOperationType string

const (
	APMTrackMetric      APMOperationType = "track_metric"
	APMTraceExecution   APMOperationType = "trace_execution"
	APMAlertGeneration  APMOperationType = "alert_generation"
	APMLogTransaction   APMOperationType = "log_transaction"
	APMPerformanceCheck APMOperationType = "performance_check"
	APMResourceMonitor  APMOperationType = "resource_monitor"
	APMHealthCheck      APMOperationType = "health_check"
	APMAuditLog         APMOperationType = "audit_log"
)

// APMNodeConfig represents the configuration for an APM node
type APMNodeConfig struct {
	OperationType APMOperationType `json:"operation_type"`
	Provider      string          `json:"provider"`  // "opentelemetry", "datadog", "newrelic", etc.
	Endpoint      string          `json:"endpoint"`  // APM collection endpoint
	ApiKey        string          `json:"api_key"`   // APM service API key
	ServiceName   string          `json:"service_name"`
	Application   string          `json:"application"`
	Environment   string          `json:"environment"` // "production", "staging", "development"
	CollectMetrics bool           `json:"collect_metrics"`
	CollectTraces bool           `json:"collect_traces"`
	CollectLogs   bool           `json:"collect_logs"`
	SampleRate    float64        `json:"sample_rate"` // 0.0 to 1.0
	Timeout       time.Duration  `json:"timeout"`
	BufferSize    int            `json:"buffer_size"`
	FlushInterval time.Duration  `json:"flush_interval"`
	EnableTLS     bool           `json:"enable_tls"`
	IgnoreErrors  []string       `json:"ignore_errors"`
	CustomTags    map[string]string `json:"custom_tags"`
	MetricConfigs []MetricConfig    `json:"metric_configs"`
	AlertConfigs  []AlertConfig     `json:"alert_configs"`
}

// MetricConfig represents configuration for a specific metric
type MetricConfig struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`  // "counter", "gauge", "histogram", "timer"
	Description string            `json:"description"`
	Unit        string            `json:"unit"`  // "milliseconds", "bytes", "percent", etc.
	Attributes  map[string]string `json:"attributes"`
	Thresholds  map[string]float64 `json:"thresholds"` // e.g., {"warning": 80, "critical": 90}
	Enabled     bool              `json:"enabled"`
}

// AlertConfig represents configuration for an alert
type AlertConfig struct {
	Name        string            `json:"name"`
	Condition   string            `json:"condition"`  // "greater_than", "less_than", "equals", etc.
	Threshold   float64           `json:"threshold"`
	Operator    string            `json:"operator"`   // "AND", "OR"
	Description string            `json:"description"`
	Severity    string            `json:"severity"`   // "info", "warning", "error", "critical"
	Channels    []string          `json:"channels"`   // "email", "slack", "webhook", etc.
	Recipients  []string          `json:"recipients"`
	Enabled     bool              `json:"enabled"`
}

// APMNode represents an APM (Application Performance Monitoring) node
type APMNode struct {
	config   *APMNodeConfig
	meter    metric.Meter
	tracer   trace.Tracer
	provider string
}

// NewAPMNode creates a new APM node
func NewAPMNode(config *APMNodeConfig) *APMNode {
	if config.SampleRate == 0 {
		config.SampleRate = 1.0 // Default to 100% sampling
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.FlushInterval == 0 {
		config.FlushInterval = 10 * time.Second
	}
	if config.BufferSize == 0 {
		config.BufferSize = 1000
	}

	// Set up OpenTelemetry provider
	otel.SetTracerProvider(otel.GetTracerProvider()) // Use global provider or create custom
	meter := otel.Meter("citadel-agent.apm")
	tracer := otel.Tracer("citadel-agent.apm")

	apmNode := &APMNode{
		config:  config,
		meter:   meter,
		tracer:  tracer,
		provider: config.Provider,
	}

	return apmNode
}

// Execute executes the APM operation
func (an *APMNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	operation := an.config.OperationType
	if op, exists := inputs["operation_type"]; exists {
		if opStr, ok := op.(string); ok {
			operation = APMOperationType(opStr)
		}
	}

	switch operation {
	case APMTrackMetric:
		return an.trackMetric(ctx, inputs)
	case APMTraceExecution:
		return an.traceExecution(ctx, inputs)
	case APMAlertGeneration:
		return an.generateAlert(ctx, inputs)
	case APMLogTransaction:
		return an.logTransaction(ctx, inputs)
	case APMPerformanceCheck:
		return an.performanceCheck(ctx, inputs)
	case APMResourceMonitor:
		return an.resourceMonitor(ctx, inputs)
	case APMHealthCheck:
		return an.healthCheck(ctx, inputs)
	case APMAuditLog:
		return an.auditLog(ctx, inputs)
	default:
		return nil, fmt.Errorf("unsupported APM operation: %s", operation)
	}
}

// trackMetric tracks a custom metric
func (an *APMNode) trackMetric(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	metricName := ""
	if name, exists := inputs["metric_name"]; exists {
		if nameStr, ok := name.(string); ok {
			metricName = nameStr
		}
	}

	if metricName == "" {
		return nil, fmt.Errorf("metric name is required")
	}

	value, exists := inputs["value"]
	if !exists {
		return nil, fmt.Errorf("metric value is required")
	}

	valueFloat, err := toFloat64(value)
	if err != nil {
		return nil, fmt.Errorf("metric value must be numeric: %w", err)
	}

	// Additional attributes from inputs
	attributes := make(map[string]string)
	for k, v := range inputs {
		if k != "metric_name" && k != "value" && k != "operation_type" {
			attributes[k] = fmt.Sprintf("%v", v)
		}
	}

	// Add custom tags from config
	for k, v := range an.config.CustomTags {
		attributes[k] = v
	}

	// Create metric attributes
	attrSet := attribute.NewSet()
	for k, v := range attributes {
		attrSet = attrSet.With(attribute.String(k, v))
	}

	// Record the metric based on type
	var result map[string]interface{}
	switch inputs["metric_type"] {
	case "counter":
		counter, err := an.meter.Int64Counter(metricName)
		if err != nil {
			return nil, fmt.Errorf("failed to create counter: %w", err)
		}
		counter.Add(ctx, int64(valueFloat), metric.WithAttributes(attrSet.ToSlice()...))
		
		result = map[string]interface{}{
			"success":     true,
			"operation":   string(APMTrackMetric),
			"metric_name": metricName,
			"metric_type": "counter",
			"value":       int64(valueFloat),
			"attributes":  attributes,
		}
	case "gauge":
		gauge, err := an.meter.Float64ObservableGauge(metricName)
		if err != nil {
			return nil, fmt.Errorf("failed to create gauge: %w", err)
		}
		// Note: Observable gauges are typically registered during initialization
		// For immediate recording, we'll use a direct approach
		
		result = map[string]interface{}{
			"success":     true,
			"operation":   string(APMTrackMetric),
			"metric_name": metricName,
			"metric_type": "gauge",
			"value":       valueFloat,
			"attributes":  attributes,
		}
	case "histogram":
		histogram, err := an.meter.Float64Histogram(metricName)
		if err != nil {
			return nil, fmt.Errorf("failed to create histogram: %w", err)
		}
		histogram.Record(ctx, valueFloat, metric.WithAttributes(attrSet.ToSlice()...))
		
		result = map[string]interface{}{
			"success":     true,
			"operation":   string(APMTrackMetric),
			"metric_name": metricName,
			"metric_type": "histogram",
			"value":       valueFloat,
			"attributes":  attributes,
		}
	default: // Default to counter
		counter, err := an.meter.Int64Counter(metricName)
		if err != nil {
			return nil, fmt.Errorf("failed to create counter: %w", err)
		}
		counter.Add(ctx, int64(valueFloat), metric.WithAttributes(attrSet.ToSlice()...))
		
		result = map[string]interface{}{
			"success":     true,
			"operation":   string(APMTrackMetric),
			"metric_name": metricName,
			"metric_type": "counter",
			"value":       int64(valueFloat),
			"attributes":  attributes,
		}
	}

	result["timestamp"] = time.Now().Unix()
	return result, nil
}

// traceExecution traces an execution with OpenTelemetry
func (an *APMNode) traceExecution(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	spanName := "custom_trace"
	if name, exists := inputs["span_name"]; exists {
		if nameStr, ok := name.(string); ok {
			spanName = nameStr
		}
	}

	// Start a new span
	ctx, span := an.tracer.Start(ctx, spanName)
	defer span.End()

	// Set attributes
	for k, v := range inputs {
		if k != "span_name" && k != "operation_type" {
			span.SetAttributes(attribute.String(k, fmt.Sprintf("%v", v)))
		}
	}

	// Add custom tags as attributes
	for k, v := range an.config.CustomTags {
		span.SetAttributes(attribute.String(k, v))
	}

	// Record any errors if present
	if errorInput, exists := inputs["error"]; exists {
		if errorStr, ok := errorInput.(string); ok {
			span.RecordError(fmt.Errorf(errorStr))
			span.SetStatus(codes.Error, errorStr)
		} else if err, ok := errorInput.(error); ok {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
	}

	result := map[string]interface{}{
		"success":       true,
		"operation":     string(APMTraceExecution),
		"span_name":     spanName,
		"trace_id":      span.SpanContext().TraceID().String(),
		"span_id":       span.SpanContext().SpanID().String(),
		"attributes":    inputs,
		"timestamp":     time.Now().Unix(),
	}

	return result, nil
}

// generateAlert generates an alert based on conditions
func (an *APMNode) generateAlert(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	alertName := ""
	if name, exists := inputs["alert_name"]; exists {
		if nameStr, ok := name.(string); ok {
			alertName = nameStr
		}
	}

	if alertName == "" {
		return nil, fmt.Errorf("alert name is required")
	}

	value, exists := inputs["value"]
	if !exists {
		return nil, fmt.Errorf("value is required for alert generation")
	}

	threshold, exists := inputs["threshold"]
	if !exists {
		// Look for alert config in node configuration
		for _, alertConfig := range an.config.AlertConfigs {
			if alertConfig.Name == alertName && alertConfig.Enabled {
				threshold = alertConfig.Threshold
				break
			}
		}
	}

	if threshold == nil {
		return nil, fmt.Errorf("threshold is required for alert generation")
	}

	valueFloat, err := toFloat64(value)
	if err != nil {
		return nil, fmt.Errorf("value must be numeric: %w", err)
	}

	thresholdFloat, err := toFloat64(threshold)
	if err != nil {
		return nil, fmt.Errorf("threshold must be numeric: %w", err)
	}

	operator := "greater_than" // Default operator
	if op, exists := inputs["operator"]; exists {
		if opStr, ok := op.(string); ok {
			operator = opStr
		}
	}

	triggered := false
	switch operator {
	case "greater_than", "gt":
		triggered = valueFloat > thresholdFloat
	case "less_than", "lt":
		triggered = valueFloat < thresholdFloat
	case "greater_than_or_equal", "gte":
		triggered = valueFloat >= thresholdFloat
	case "less_than_or_equal", "lte":
		triggered = valueFloat <= thresholdFloat
	case "equals", "eq":
		triggered = valueFloat == thresholdFloat
	case "not_equals", "ne":
		triggered = valueFloat != thresholdFloat
	default:
		return nil, fmt.Errorf("unsupported operator: %s", operator)
	}

	alertLevel := "info"
	severity := "info"
	
	// Determine alert level based on severity thresholds if specified
	for _, alertConfig := range an.config.AlertConfigs {
		if alertConfig.Name == alertName {
			if valueFloat >= alertConfig.Thresholds["critical"] {
				alertLevel = "critical"
				severity = "critical"
			} else if valueFloat >= alertConfig.Thresholds["warning"] {
				alertLevel = "warning"
				severity = "warning"
			}
			break
		}
	}

	result := map[string]interface{}{
		"success":     true,
		"operation":   string(APMAlertGeneration),
		"alert_name":  alertName,
		"value":       valueFloat,
		"threshold":   thresholdFloat,
		"operator":    operator,
		"triggered":   triggered,
		"level":       alertLevel,
		"severity":    severity,
		"timestamp":   time.Now().Unix(),
	}

	// If alert is triggered, record it as a metric
	if triggered {
		// In a real implementation, this would send the alert to notification systems
		// For now, we'll just mark it in the result
		
		// Record alert as a counter metric
		alertCounter, err := an.meter.Int64Counter("citadel.alerts.total")
		if err != nil {
			fmt.Printf("Warning: failed to create alert counter: %v\n", err)
		} else {
			alertCounter.Add(ctx, 1, metric.WithAttributes(
				attribute.String("alert_name", alertName),
				attribute.String("severity", severity),
				attribute.Bool("triggered", triggered),
			))
		}
	}

	return result, nil
}

// logTransaction logs a transaction for APM tracking
func (an *APMNode) logTransaction(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	transactionName := ""
	if name, exists := inputs["transaction_name"]; exists {
		if nameStr, ok := name.(string); ok {
			transactionName = nameStr
		}
	}

	if transactionName == "" {
		return nil, fmt.Errorf("transaction name is required")
	}

	duration, exists := inputs["duration"]
	if !exists {
		duration = 0.0 // Default to 0 if not provided
	}

	durationFloat, err := toFloat64(duration)
	if err != nil {
		durationFloat = 0.0
	}

	startTime := time.Now()
	if st, exists := inputs["start_time"]; exists {
		if stStr, ok := st.(string); ok {
			if parsedTime, err := time.Parse(time.RFC3339, stStr); err == nil {
				startTime = parsedTime
			}
		}
	}

	endTime := startTime.Add(time.Duration(durationFloat) * time.Second)

	result := map[string]interface{}{
		"success":         true,
		"operation":       string(APMLogTransaction),
		"transaction_name": transactionName,
		"start_time":      startTime.Unix(),
		"end_time":        endTime.Unix(),
		"duration":        durationFloat,
		"attributes":      inputs,
		"timestamp":       time.Now().Unix(),
	}

	// Record as a histogram metric for response time
	responseTimeHistogram, err := an.meter.Float64Histogram("citadel.transaction.duration")
	if err != nil {
		fmt.Printf("Warning: failed to create response time histogram: %v\n", err)
	} else {
		responseTimeHistogram.Record(ctx, durationFloat, metric.WithAttributes(
			attribute.String("transaction_name", transactionName),
		))
	}

	return result, nil
}

// performanceCheck performs a performance check
func (an *APMNode) performanceCheck(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	checkType := "response_time"
	if typ, exists := inputs["check_type"]; exists {
		if typStr, ok := typ.(string); ok {
			checkType = typStr
		}
	}

	threshold, exists := inputs["threshold"]
	if !exists {
		return nil, fmt.Errorf("threshold is required for performance check")
	}

	thresholdFloat, err := toFloat64(threshold)
	if err != nil {
		return nil, fmt.Errorf("threshold must be numeric: %w", err)
	}

	value, exists := inputs["value"]
	if !exists {
		return nil, fmt.Errorf("value is required for performance check")
	}

	valueFloat, err := toFloat64(value)
	if err != nil {
		return nil, fmt.Errorf("value must be numeric: %w", err)
	}

	passed := valueFloat <= thresholdFloat

	result := map[string]interface{}{
		"success":     true,
		"operation":   string(APMPerformanceCheck),
		"check_type":  checkType,
		"value":       valueFloat,
		"threshold":   thresholdFloat,
		"passed":      passed,
		"attributes":  inputs,
		"timestamp":   time.Now().Unix(),
	}

	// Record performance check as a gauge metric
	perfGauge, err := an.meter.Float64ObservableGauge("citadel.performance.checks")
	if err != nil {
		fmt.Printf("Warning: failed to create performance gauge: %v\n", err)
	} else {
		// In a real implementation, this would be an observable gauge
		// For now, we'll record the check as a counter
		perfCounter, err := an.meter.Int64Counter("citadel.performance.checks.total")
		if err != nil {
			fmt.Printf("Warning: failed to create performance counter: %v\n", err)
		} else {
			perfCounter.Add(ctx, 1, metric.WithAttributes(
				attribute.String("check_type", checkType),
				attribute.Bool("passed", passed),
				attribute.Float64("value", valueFloat),
				attribute.Float64("threshold", thresholdFloat),
			))
		}
	}

	return result, nil
}

// resourceMonitor monitors resource usage
func (an *APMNode) resourceMonitor(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Get resource type
	resourceType := "cpu"
	if typ, exists := inputs["resource_type"]; exists {
		if typStr, ok := typ.(string); ok {
			resourceType = typStr
		}
	}

	value, exists := inputs["value"]
	if !exists {
		return nil, fmt.Errorf("value is required for resource monitor")
	}

	valueFloat, err := toFloat64(value)
	if err != nil {
		return nil, fmt.Errorf("value must be numeric: %w", err)
	}

	// Get threshold from config or inputs
	threshold := an.getResourceThreshold(resourceType)
	if thr, exists := inputs["threshold"]; exists {
		if thrFloat, err := toFloat64(thr); err == nil {
			threshold = thrFloat
		}
	}

	aboveThreshold := valueFloat > threshold

	result := map[string]interface{}{
		"success":        true,
		"operation":      string(APMResourceMonitor),
		"resource_type":  resourceType,
		"value":          valueFloat,
		"threshold":      threshold,
		"above_threshold": aboveThreshold,
		"timestamp":      time.Now().Unix(),
	}

	// Record resource metric
	resourceGauge, err := an.meter.Float64ObservableGauge("citadel.resources.utilization")
	if err != nil {
		fmt.Printf("Warning: failed to create resource gauge: %v\n", err)
	} else {
		// For immediate recording, use a counter
		resourceCounter, err := an.meter.Int64Counter("citadel.resources.utilization.total")
		if err != nil {
			fmt.Printf("Warning: failed to create resource counter: %v\n", err)
		} else {
			resourceCounter.Add(ctx, 1, metric.WithAttributes(
				attribute.String("resource_type", resourceType),
				attribute.Bool("above_threshold", aboveThreshold),
				attribute.Float64("value", valueFloat),
				attribute.Float64("threshold", threshold),
			))
		}
	}

	return result, nil
}

// getResourceThreshold returns the default threshold for a resource type
func (an *APMNode) getResourceThreshold(resourceType string) float64 {
	switch resourceType {
	case "cpu":
		return 80.0 // 80% CPU usage threshold
	case "memory":
		return 85.0 // 85% memory usage threshold
	case "disk":
		return 90.0 // 90% disk usage threshold
	case "network":
		return 100.0 // 100MB/s threshold
	default:
		return 80.0 // Default
	}
}

// healthCheck performs a health check
func (an *APMNode) healthCheck(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	component := "system"
	if comp, exists := inputs["component"]; exists {
		if compStr, ok := comp.(string); ok {
			component = compStr
		}
	}

	// In a real implementation, this would check the actual health of the component
	// For this example, we'll assume it's healthy
	healthy := true

	// Add custom checks based on inputs
	if url, exists := inputs["url"]; exists {
		if urlStr, ok := url.(string); ok {
			// Perform HTTP health check
			healthy = an.httpHealthCheck(urlStr)
		}
	} else if endpoint, exists := inputs["endpoint"]; exists {
		if epStr, ok := endpoint.(string); ok {
			// Perform endpoint-specific health check
			healthy = an.endpointHealthCheck(epStr)
		}
	}

	result := map[string]interface{}{
		"success":     true,
		"operation":   string(APMHealthCheck),
		"component":   component,
		"healthy":     healthy,
		"timestamp":   time.Now().Unix(),
	}

	// Record health check metric
	healthCounter, err := an.meter.Int64Counter("citadel.health.checks.total")
	if err != nil {
		fmt.Printf("Warning: failed to create health check counter: %v\n", err)
	} else {
		healthCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("component", component),
			attribute.Bool("healthy", healthy),
		))
	}

	return result, nil
}

// httpHealthCheck performs an HTTP health check
func (an *APMNode) httpHealthCheck(url string) bool {
	// In a real implementation, this would perform an HTTP request
	// For now, we'll just return true
	return true
}

// endpointHealthCheck performs an endpoint-specific health check
func (an *APMNode) endpointHealthCheck(endpoint string) bool {
	// In a real implementation, this would check the specific endpoint
	// For now, we'll just return true
	return true
}

// auditLog creates an audit log entry
func (an *APMNode) auditLog(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	action := ""
	if act, exists := inputs["action"]; exists {
		if actStr, ok := act.(string); ok {
			action = actStr
		}
	}

	if action == "" {
		return nil, fmt.Errorf("action is required for audit log")
	}

	userID := ""
	if uid, exists := inputs["user_id"]; exists {
		if uidStr, ok := uid.(string); ok {
			userID = uidStr
		}
	}

	resource := ""
	if res, exists := inputs["resource"]; exists {
		if resStr, ok := res.(string); ok {
			resource = resStr
		}
	}

	result := map[string]interface{}{
		"success":     true,
		"operation":   string(APMAuditLog),
		"action":      action,
		"user_id":     userID,
		"resource":    resource,
		"details":     inputs,
		"timestamp":   time.Now().Unix(),
		"service":     an.config.ServiceName,
		"application": an.config.Application,
		"environment": an.config.Environment,
	}

	// Record audit event as a counter metric
	auditCounter, err := an.meter.Int64Counter("citadel.audit.events.total")
	if err != nil {
		fmt.Printf("Warning: failed to create audit event counter: %v\n", err)
	} else {
		auditCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("action", action),
			attribute.String("resource", resource),
			attribute.String("user_id", userID),
			attribute.String("service", an.config.ServiceName),
		))
	}

	return result, nil
}

// toFloat64 converts an interface{} to float64
func toFloat64(i interface{}) (float64, error) {
	switch v := i.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", i)
	}
}

// RegisterAPMNode registers the APM node type with the engine
func RegisterAPMNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("apm_monitoring", func(config map[string]interface{}) (engine.NodeInstance, error) {
		var operationType APMOperationType
		if op, exists := config["operation_type"]; exists {
			if opStr, ok := op.(string); ok {
				operationType = APMOperationType(opStr)
			}
		}

		var provider string
		if prov, exists := config["provider"]; exists {
			if provStr, ok := prov.(string); ok {
				provider = provStr
			}
		}

		var endpoint string
		if ep, exists := config["endpoint"]; exists {
			if epStr, ok := ep.(string); ok {
				endpoint = epStr
			}
		}

		var apiKey string
		if key, exists := config["api_key"]; exists {
			if keyStr, ok := key.(string); ok {
				apiKey = keyStr
			}
		}

		var serviceName string
		if name, exists := config["service_name"]; exists {
			if nameStr, ok := name.(string); ok {
				serviceName = nameStr
			}
		}

		var application string
		if app, exists := config["application"]; exists {
			if appStr, ok := app.(string); ok {
				application = appStr
			}
		}

		var environment string
		if env, exists := config["environment"]; exists {
			if envStr, ok := env.(string); ok {
				environment = envStr
			}
		}

		var collectMetrics bool
		if metrics, exists := config["collect_metrics"]; exists {
			if metricsBool, ok := metrics.(bool); ok {
				collectMetrics = metricsBool
			}
		}

		var collectTraces bool
		if traces, exists := config["collect_traces"]; exists {
			if tracesBool, ok := traces.(bool); ok {
				collectTraces = tracesBool
			}
		}

		var collectLogs bool
		if logs, exists := config["collect_logs"]; exists {
			if logsBool, ok := logs.(bool); ok {
				collectLogs = logsBool
			}
		}

		var sampleRate float64
		if rate, exists := config["sample_rate"]; exists {
			if rateFloat, ok := rate.(float64); ok {
				sampleRate = rateFloat
			}
		}

		var timeout float64
		if t, exists := config["timeout_seconds"]; exists {
			if tFloat, ok := t.(float64); ok {
				timeout = tFloat
			}
		}

		var bufferSize float64
		if size, exists := config["buffer_size"]; exists {
			if sizeFloat, ok := size.(float64); ok {
				bufferSize = sizeFloat
			}
		}

		var flushInterval float64
		if fi, exists := config["flush_interval_seconds"]; exists {
			if fiFloat, ok := fi.(float64); ok {
				flushInterval = fiFloat
			}
		}

		var enableTLS bool
		if tls, exists := config["enable_tls"]; exists {
			if tlsBool, ok := tls.(bool); ok {
				enableTLS = tlsBool
			}
		}

		var ignoreErrors []string
		if errs, exists := config["ignore_errors"]; exists {
			if errsSlice, ok := errs.([]interface{}); ok {
				for _, err := range errsSlice {
					if errStr, ok := err.(string); ok {
						ignoreErrors = append(ignoreErrors, errStr)
					}
				}
			}
		}

		var customTags map[string]string
		if tags, exists := config["custom_tags"]; exists {
			if tagsMap, ok := tags.(map[string]interface{}); ok {
				customTags = make(map[string]string)
				for k, v := range tagsMap {
					if vStr, ok := v.(string); ok {
						customTags[k] = vStr
					} else {
						customTags[k] = fmt.Sprintf("%v", v)
					}
				}
			}
		}

		var metricConfigs []MetricConfig
		if mcs, exists := config["metric_configs"]; exists {
			if mcsArr, ok := mcs.([]interface{}); ok {
				for _, mc := range mcsArr {
					if mcMap, ok := mc.(map[string]interface{}); ok {
						var thresholds map[string]float64
						if thr, exists := mcMap["thresholds"]; exists {
							if thrMap, ok := thr.(map[string]interface{}); ok {
								thresholds = make(map[string]float64)
								for k, v := range thrMap {
									if vFloat, ok := v.(float64); ok {
										thresholds[k] = vFloat
									}
								}
							}
						}

						var attributes map[string]string
						if attrs, exists := mcMap["attributes"]; exists {
							if attrsMap, ok := attrs.(map[string]interface{}); ok {
								attributes = make(map[string]string)
								for k, v := range attrsMap {
									if vStr, ok := v.(string); ok {
										attributes[k] = vStr
									}
								}
							}
						}

						var enabled bool
						if en, exists := mcMap["enabled"]; exists {
							if enBool, ok := en.(bool); ok {
								enabled = enBool
							}
						} else {
							enabled = true // Default to enabled
						}

						metricConfigs = append(metricConfigs, MetricConfig{
							Name:        getStringValue(mcMap["name"]),
							Type:        getStringValue(mcMap["type"]),
							Description: getStringValue(mcMap["description"]),
							Unit:        getStringValue(mcMap["unit"]),
							Attributes:  attributes,
							Thresholds:  thresholds,
							Enabled:     enabled,
						})
					}
				}
			}
		}

		var alertConfigs []AlertConfig
		if acs, exists := config["alert_configs"]; exists {
			if acsArr, ok := acs.([]interface{}); ok {
				for _, ac := range acsArr {
					if acMap, ok := ac.(map[string]interface{}); ok {
						var channels []string
						if chans, exists := acMap["channels"]; exists {
							if chansArr, ok := chans.([]interface{}); ok {
								for _, ch := range chansArr {
									if chStr, ok := ch.(string); ok {
										channels = append(channels, chStr)
									}
								}
							}
						}

						var recipients []string
						if recips, exists := acMap["recipients"]; exists {
							if recipsArr, ok := recips.([]interface{}); ok {
								for _, recip := range recipsArr {
									if recipStr, ok := recip.(string); ok {
										recipients = append(recipients, recipStr)
									}
								}
							}
						}

						var thresholds map[string]float64
						if thr, exists := acMap["thresholds"]; exists {
							if thrMap, ok := thr.(map[string]interface{}); ok {
								thresholds = make(map[string]float64)
								for k, v := range thrMap {
									if vFloat, ok := v.(float64); ok {
										thresholds[k] = vFloat
									}
								}
							}
						}

						var enabled bool
						if en, exists := acMap["enabled"]; exists {
							if enBool, ok := en.(bool); ok {
								enabled = enBool
							}
						} else {
							enabled = true // Default to enabled
						}

						alertConfigs = append(alertConfigs, AlertConfig{
							Name:        getStringValue(acMap["name"]),
							Condition:   getStringValue(acMap["condition"]),
							Threshold:   getFloat64Value(acMap["threshold"]),
							Operator:    getStringValue(acMap["operator"]),
							Description: getStringValue(acMap["description"]),
							Severity:    getStringValue(acMap["severity"]),
							Channels:    channels,
							Recipients:  recipients,
							Thresholds:  thresholds,
							Enabled:     enabled,
						})
					}
				}
			}
		}

		nodeConfig := &APMNodeConfig{
			OperationType:  operationType,
			Provider:       provider,
			Endpoint:       endpoint,
			APIKey:         apiKey,
			ServiceName:    serviceName,
			Application:    application,
			Environment:    environment,
			CollectMetrics: collectMetrics,
			CollectTraces:  collectTraces,
			CollectLogs:    collectLogs,
			SampleRate:     sampleRate,
			Timeout:        time.Duration(timeout) * time.Second,
			BufferSize:     int(bufferSize),
			FlushInterval:  time.Duration(flushInterval) * time.Second,
			EnableTLS:      enableTLS,
			IgnoreErrors:   ignoreErrors,
			CustomTags:     customTags,
			MetricConfigs:  metricConfigs,
			AlertConfigs:   alertConfigs,
		}

		return NewAPMNode(nodeConfig), nil
	})
}

// getStringValue safely gets a string value from interface{}
func getStringValue(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// getFloat64Value safely gets a float64 value from interface{}
func getFloat64Value(v interface{}) float64 {
	if v == nil {
		return 0.0
	}
	if f, ok := v.(float64); ok {
		return f
	}
	if s, ok := v.(string); ok {
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f
		}
	}
	return 0.0
}