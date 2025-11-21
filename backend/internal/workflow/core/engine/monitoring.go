package engine

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MetricsCollector collects and stores execution metrics
type MetricsCollector struct {
	metrics    map[string]*Metric
	mutex      sync.RWMutex
	aggregator *MetricAggregator
	exporter   *MetricExporter
}

// Metric represents a single metric
type Metric struct {
	Name       string
	Value      float64
	Timestamp  time.Time
	Labels     map[string]string
	Aggregator string // "sum", "avg", "max", "min", "count"
}

// MetricAggregator aggregates metrics over time
type MetricAggregator struct {
	aggregatedMetrics map[string]*AggregatedMetric
	windowSize        time.Duration
}

// AggregatedMetric represents aggregated metrics
type AggregatedMetric struct {
	Sum   float64
	Count int
	Avg   float64
	Max   float64
	Min   float64
	Last  float64
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	mc := &MetricsCollector{
		metrics:    make(map[string]*Metric),
		aggregator: NewMetricAggregator(),
		exporter:   NewMetricExporter(),
	}

	// Start background aggregation
	go mc.aggregateMetrics()
	
	return mc
}

// NewMetricAggregator creates a new metric aggregator
func NewMetricAggregator() *MetricAggregator {
	return &MetricAggregator{
		aggregatedMetrics: make(map[string]*AggregatedMetric),
		windowSize:        5 * time.Minute,
	}
}

// RecordMetric records a new metric
func (mc *MetricsCollector) RecordMetric(name string, value float64, labels map[string]string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	metric := &Metric{
		Name:      name,
		Value:     value,
		Timestamp: time.Now(),
		Labels:    labels,
	}

	// Generate unique key for this metric
	key := fmt.Sprintf("%s_%v", name, labels)
	mc.metrics[key] = metric
	
	// Add to aggregator for time-based aggregation
	mc.aggregator.AddMetric(name, value)
}

// GetMetric retrieves a specific metric
func (mc *MetricsCollector) GetMetric(name string, labels map[string]string) (*Metric, error) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	key := fmt.Sprintf("%s_%v", name, labels)
	metric, exists := mc.metrics[key]
	if !exists {
		return nil, fmt.Errorf("metric %s not found", name)
	}

	return metric, nil
}

// GetAggregatedMetrics retrieves aggregated metrics
func (mc *MetricsCollector) GetAggregatedMetrics() map[string]*AggregatedMetric {
	return mc.aggregator.GetAggregatedMetrics()
}

// aggregateMetrics runs in background to periodically aggregate metrics
func (mc *MetricsCollector) aggregateMetrics() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		mc.mutex.Lock()
		// Clean up old metrics (older than 1 hour)
		cutoff := time.Now().Add(-1 * time.Hour)
		for key, metric := range mc.metrics {
			if metric.Timestamp.Before(cutoff) {
				delete(mc.metrics, key)
			}
		}
		mc.mutex.Unlock()
	}
}

// AddMetric adds a metric to the aggregator
func (ma *MetricAggregator) AddMetric(name string, value float64) {
	ma.aggregatedMetrics[name] = &AggregatedMetric{
		Sum:   ma.aggregatedMetrics[name].Sum + value,
		Count: ma.aggregatedMetrics[name].Count + 1,
		Avg:   (ma.aggregatedMetrics[name].Sum + value) / float64(ma.aggregatedMetrics[name].Count+1),
		Max:   max(ma.aggregatedMetrics[name].Max, value),
		Min:   min(ma.aggregatedMetrics[name].Min, value),
		Last:  value,
	}
}

// GetAggregatedMetrics returns all aggregated metrics
func (ma *MetricAggregator) GetAggregatedMetrics() map[string]*AggregatedMetric {
	return ma.aggregatedMetrics
}

// MetricExporter exports metrics to external systems
type MetricExporter struct {
	exporters []MetricExporterInterface
}

// MetricExporterInterface defines the interface for metric exporters
type MetricExporterInterface interface {
	Export(metrics []*Metric) error
}

// NewMetricExporter creates a new metric exporter
func NewMetricExporter() *MetricExporter {
	return &MetricExporter{
		exporters: []MetricExporterInterface{},
	}
}

// AddExporter adds a new exporter
func (me *MetricExporter) AddExporter(exporter MetricExporterInterface) {
	me.exporters = append(me.exporters, exporter)
}

// Export exports metrics to all registered exporters
func (me *MetricExporter) Export(metrics []*Metric) error {
	for _, exporter := range me.exporters {
		if err := exporter.Export(metrics); err != nil {
			return fmt.Errorf("exporter failed: %w", err)
		}
	}
	return nil
}

// TraceCollector collects execution traces for debugging
type TraceCollector struct {
	traces   map[string]*Trace
	mutex    sync.RWMutex
	exporter *TraceExporter
}

// Trace represents a single execution trace
type Trace struct {
	ID         string
	WorkflowID string
	ExecutionID string
	NodeID     string
	StartTime  time.Time
	EndTime    time.Time
	Status     string
	Error      string
	Data       map[string]interface{}
	ParentID   string
}

// TraceExporter exports traces to external systems
type TraceExporter struct {
	exporters []TraceExporterInterface
}

// TraceExporterInterface defines the interface for trace exporters
type TraceExporterInterface interface {
	ExportTrace(trace *Trace) error
}

// NewTraceCollector creates a new trace collector
func NewTraceCollector() *TraceCollector {
	return &TraceCollector{
		traces:   make(map[string]*Trace),
		exporter: NewTraceExporter(),
	}
}

// NewTraceExporter creates a new trace exporter
func NewTraceExporter() *TraceExporter {
	return &TraceExporter{
		exporters: []TraceExporterInterface{},
	}
}

// StartTrace starts a new trace
func (tc *TraceCollector) StartTrace(workflowID, executionID, nodeID string, parentID string) *Trace {
	trace := &Trace{
		ID:         generateTraceID(),
		WorkflowID: workflowID,
		ExecutionID: executionID,
		NodeID:     nodeID,
		StartTime:  time.Now(),
		Status:     "running",
		ParentID:   parentID,
		Data:       make(map[string]interface{}),
	}

	tc.mutex.Lock()
	tc.traces[trace.ID] = trace
	tc.mutex.Unlock()

	return trace
}

// EndTrace ends an existing trace
func (tc *TraceCollector) EndTrace(traceID string, status string, errorStr string, data map[string]interface{}) error {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	trace, exists := tc.traces[traceID]
	if !exists {
		return fmt.Errorf("trace %s not found", traceID)
	}

	trace.EndTime = time.Now()
	trace.Status = status
	trace.Error = errorStr
	trace.Data = data

	// Export the completed trace
	if tc.exporter != nil {
		tc.exporter.ExportTrace(trace)
	}

	return nil
}

// GetTrace retrieves a specific trace
func (tc *TraceCollector) GetTrace(traceID string) (*Trace, error) {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()

	trace, exists := tc.traces[traceID]
	if !exists {
		return nil, fmt.Errorf("trace %s not found", traceID)
	}

	return trace, nil
}

// Alerter sends alerts based on metrics and conditions
type Alerter struct {
	alerts     []*Alert
	conditions []*AlertCondition
	mutex      sync.RWMutex
	notifiers  []Notifier
}

// Alert represents a single alert
type Alert struct {
	ID          string
	Name        string
	Description string
	Severity    string // "low", "medium", "high", "critical"
	Status      string // "active", "resolved", "suppressed"
	Timestamp   time.Time
	TriggerData map[string]interface{}
}

// AlertCondition defines conditions that trigger alerts
type AlertCondition struct {
	ID        string
	Name      string
	Metric    string
	Operator  string // "gt", "lt", "eq", "gte", "lte"
	Threshold float64
	Window    time.Duration
}

// Notifier interface for sending alerts
type Notifier interface {
	Notify(alert *Alert) error
}

// NewAlerter creates a new alerter
func NewAlerter() *Alerter {
	a := &Alerter{
		alerts:     make([]*Alert, 0),
		conditions: make([]*AlertCondition, 0),
		notifiers:  make([]Notifier, 0),
	}
	
	// Start background alert checking
	go a.checkAlerts()
	
	return a
}

// AddCondition adds a new alert condition
func (a *Alerter) AddCondition(condition *AlertCondition) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	
	a.conditions = append(a.conditions, condition)
}

// AddNotifier adds a new notifier
func (a *Alerter) AddNotifier(notifier Notifier) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	
	a.notifiers = append(a.notifiers, notifier)
}

// checkAlerts runs in background to periodically check alert conditions
func (a *Alerter) checkAlerts() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		a.checkConditions()
	}
}

// checkConditions checks all alert conditions
func (a *Alerter) checkConditions() {
	// In a real implementation, this would check against collected metrics
	// For now, we'll just simulate checking
	
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	
	for _, condition := range a.conditions {
		// This is where we would check the actual metric values
		// For now, we'll just log the condition
		fmt.Printf("Checking condition: %s for metric: %s\n", condition.Name, condition.Metric)
	}
}

// TriggerAlert manually triggers an alert
func (a *Alerter) TriggerAlert(name, description, severity string, data map[string]interface{}) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	alert := &Alert{
		ID:          generateAlertID(),
		Name:        name,
		Description: description,
		Severity:    severity,
		Status:      "active",
		Timestamp:   time.Now(),
		TriggerData: data,
	}

	a.alerts = append(a.alerts, alert)

	// Notify all configured notifiers
	for _, notifier := range a.notifiers {
		go notifier.Notify(alert)
	}
}

// Helper functions
func generateTraceID() string {
	return fmt.Sprintf("trace_%d", time.Now().UnixNano())
}

func generateAlertID() string {
	return fmt.Sprintf("alert_%d", time.Now().UnixNano())
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}