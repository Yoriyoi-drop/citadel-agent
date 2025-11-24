package engine

import (
	"sync/atomic"
	"time"
)

// Metrics tracks workflow engine metrics
type Metrics struct {
	// Workflow metrics
	WorkflowsExecuted  atomic.Int64
	WorkflowsSucceeded atomic.Int64
	WorkflowsFailed    atomic.Int64
	WorkflowsDuration  atomic.Int64 // Total duration in nanoseconds

	// Node metrics
	NodesExecuted  atomic.Int64
	NodesSucceeded atomic.Int64
	NodesFailed    atomic.Int64
	NodesDuration  atomic.Int64

	// Circuit breaker metrics
	CircuitBreakerTrips  atomic.Int64
	CircuitBreakerResets atomic.Int64

	// Retry metrics
	TotalRetries      atomic.Int64
	SuccessfulRetries atomic.Int64

	// Worker pool metrics (embedded)
	PoolMetrics *PoolMetrics
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		PoolMetrics: &PoolMetrics{},
	}
}

// RecordWorkflowExecution records a workflow execution
func (m *Metrics) RecordWorkflowExecution(success bool, duration time.Duration) {
	m.WorkflowsExecuted.Add(1)
	m.WorkflowsDuration.Add(int64(duration))

	if success {
		m.WorkflowsSucceeded.Add(1)
	} else {
		m.WorkflowsFailed.Add(1)
	}
}

// RecordNodeExecution records a node execution
func (m *Metrics) RecordNodeExecution(success bool, duration time.Duration) {
	m.NodesExecuted.Add(1)
	m.NodesDuration.Add(int64(duration))

	if success {
		m.NodesSucceeded.Add(1)
	} else {
		m.NodesFailed.Add(1)
	}
}

// RecordCircuitBreakerTrip records a circuit breaker trip
func (m *Metrics) RecordCircuitBreakerTrip() {
	m.CircuitBreakerTrips.Add(1)
}

// RecordCircuitBreakerReset records a circuit breaker reset
func (m *Metrics) RecordCircuitBreakerReset() {
	m.CircuitBreakerResets.Add(1)
}

// RecordRetry records a retry attempt
func (m *Metrics) RecordRetry(success bool) {
	m.TotalRetries.Add(1)
	if success {
		m.SuccessfulRetries.Add(1)
	}
}

// GetWorkflowSuccessRate returns workflow success rate (0-1)
func (m *Metrics) GetWorkflowSuccessRate() float64 {
	total := m.WorkflowsExecuted.Load()
	if total == 0 {
		return 0
	}
	return float64(m.WorkflowsSucceeded.Load()) / float64(total)
}

// GetNodeSuccessRate returns node success rate (0-1)
func (m *Metrics) GetNodeSuccessRate() float64 {
	total := m.NodesExecuted.Load()
	if total == 0 {
		return 0
	}
	return float64(m.NodesSucceeded.Load()) / float64(total)
}

// GetAverageWorkflowDuration returns average workflow duration
func (m *Metrics) GetAverageWorkflowDuration() time.Duration {
	executed := m.WorkflowsExecuted.Load()
	if executed == 0 {
		return 0
	}
	return time.Duration(m.WorkflowsDuration.Load() / executed)
}

// GetAverageNodeDuration returns average node duration
func (m *Metrics) GetAverageNodeDuration() time.Duration {
	executed := m.NodesExecuted.Load()
	if executed == 0 {
		return 0
	}
	return time.Duration(m.NodesDuration.Load() / executed)
}

// GetRetrySuccessRate returns retry success rate (0-1)
func (m *Metrics) GetRetrySuccessRate() float64 {
	total := m.TotalRetries.Load()
	if total == 0 {
		return 0
	}
	return float64(m.SuccessfulRetries.Load()) / float64(total)
}

// Snapshot returns a snapshot of current metrics
func (m *Metrics) Snapshot() MetricsSnapshot {
	return MetricsSnapshot{
		WorkflowsExecuted:    m.WorkflowsExecuted.Load(),
		WorkflowsSucceeded:   m.WorkflowsSucceeded.Load(),
		WorkflowsFailed:      m.WorkflowsFailed.Load(),
		NodesExecuted:        m.NodesExecuted.Load(),
		NodesSucceeded:       m.NodesSucceeded.Load(),
		NodesFailed:          m.NodesFailed.Load(),
		CircuitBreakerTrips:  m.CircuitBreakerTrips.Load(),
		CircuitBreakerResets: m.CircuitBreakerResets.Load(),
		TotalRetries:         m.TotalRetries.Load(),
		SuccessfulRetries:    m.SuccessfulRetries.Load(),
		AvgWorkflowDuration:  m.GetAverageWorkflowDuration(),
		AvgNodeDuration:      m.GetAverageNodeDuration(),
		WorkflowSuccessRate:  m.GetWorkflowSuccessRate(),
		NodeSuccessRate:      m.GetNodeSuccessRate(),
		RetrySuccessRate:     m.GetRetrySuccessRate(),
	}
}

// MetricsSnapshot represents a point-in-time snapshot of metrics
type MetricsSnapshot struct {
	WorkflowsExecuted    int64
	WorkflowsSucceeded   int64
	WorkflowsFailed      int64
	NodesExecuted        int64
	NodesSucceeded       int64
	NodesFailed          int64
	CircuitBreakerTrips  int64
	CircuitBreakerResets int64
	TotalRetries         int64
	SuccessfulRetries    int64
	AvgWorkflowDuration  time.Duration
	AvgNodeDuration      time.Duration
	WorkflowSuccessRate  float64
	NodeSuccessRate      float64
	RetrySuccessRate     float64
}
