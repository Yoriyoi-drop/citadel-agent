package engine

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrPoolClosed    = errors.New("worker pool is closed")
	ErrJobTimeout    = errors.New("job execution timeout")
	ErrQueueFull     = errors.New("job queue is full")
	ErrInvalidWorker = errors.New("invalid number of workers")
)

// Job represents a unit of work
type Job struct {
	ID       string
	Task     func(context.Context) error
	Priority int
	Timeout  time.Duration
	Retry    int
}

// Result represents the result of a job execution
type Result struct {
	JobID     string
	Error     error
	Duration  time.Duration
	Timestamp time.Time
	Retried   int
}

// WorkerPool manages a pool of workers for concurrent job execution
type WorkerPool struct {
	workers      int
	jobQueue     chan Job
	resultChan   chan Result
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	metrics      *PoolMetrics
	maxQueueSize int
	started      atomic.Bool
}

// PoolMetrics tracks worker pool statistics
type PoolMetrics struct {
	JobsSubmitted atomic.Int64
	JobsCompleted atomic.Int64
	JobsFailed    atomic.Int64
	JobsRetried   atomic.Int64
	TotalDuration atomic.Int64 // in nanoseconds
	ActiveWorkers atomic.Int32
	QueuedJobs    atomic.Int32
	mu            sync.RWMutex
	jobDurations  []time.Duration
}

// WorkerPoolConfig holds configuration for worker pool
type WorkerPoolConfig struct {
	Workers      int
	QueueSize    int
	ResultBuffer int
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(ctx context.Context, config WorkerPoolConfig) (*WorkerPool, error) {
	if config.Workers <= 0 {
		return nil, ErrInvalidWorker
	}

	if config.QueueSize <= 0 {
		config.QueueSize = 100
	}

	if config.ResultBuffer <= 0 {
		config.ResultBuffer = 100
	}

	// Use parent context if provided, otherwise create new one
	if ctx == nil {
		ctx = context.Background()
	}
	poolCtx, cancel := context.WithCancel(ctx)

	pool := &WorkerPool{
		workers:      config.Workers,
		jobQueue:     make(chan Job, config.QueueSize),
		resultChan:   make(chan Result, config.ResultBuffer),
		ctx:          poolCtx,
		cancel:       cancel,
		metrics:      &PoolMetrics{},
		maxQueueSize: config.QueueSize,
	}

	return pool, nil
}

// Start starts the worker pool
func (p *WorkerPool) Start() {
	if !p.started.CompareAndSwap(false, true) {
		return // Already started
	}

	// Start workers
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// worker processes jobs from the queue
func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()

	p.metrics.ActiveWorkers.Add(1)
	defer p.metrics.ActiveWorkers.Add(-1)

	for {
		select {
		case <-p.ctx.Done():
			return

		case job, ok := <-p.jobQueue:
			if !ok {
				return
			}

			p.metrics.QueuedJobs.Add(-1)
			p.executeJob(job)
		}
	}
}

// executeJob executes a single job
func (p *WorkerPool) executeJob(job Job) {
	startTime := time.Now()
	var err error

	// Create context with timeout if specified
	ctx := p.ctx
	if job.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(p.ctx, job.Timeout)
		defer cancel()
	}

	// Execute job with retry logic
	retries := 0
	for attempt := 0; attempt <= job.Retry; attempt++ {
		err = job.Task(ctx)
		if err == nil {
			break
		}
		retries = attempt
		if attempt < job.Retry {
			p.metrics.JobsRetried.Add(1)
			time.Sleep(time.Second * time.Duration(attempt+1)) // Simple backoff
		}
	}

	duration := time.Since(startTime)

	// Update metrics
	if err != nil {
		p.metrics.JobsFailed.Add(1)
	} else {
		p.metrics.JobsCompleted.Add(1)
	}
	p.metrics.TotalDuration.Add(int64(duration))
	p.metrics.recordDuration(duration)

	// Send result
	result := Result{
		JobID:     job.ID,
		Error:     err,
		Duration:  duration,
		Timestamp: time.Now(),
		Retried:   retries,
	}

	select {
	case p.resultChan <- result:
	default:
		// Result channel full, drop result
	}
}

// Submit submits a job to the worker pool
func (p *WorkerPool) Submit(job Job) error {
	if !p.started.Load() {
		return errors.New("worker pool not started")
	}

	select {
	case <-p.ctx.Done():
		return ErrPoolClosed
	default:
	}

	// Try to submit job
	select {
	case p.jobQueue <- job:
		p.metrics.JobsSubmitted.Add(1)
		p.metrics.QueuedJobs.Add(1)
		return nil
	default:
		return ErrQueueFull
	}
}

// SubmitWithTimeout submits a job with a timeout
func (p *WorkerPool) SubmitWithTimeout(job Job, timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case p.jobQueue <- job:
		p.metrics.JobsSubmitted.Add(1)
		p.metrics.QueuedJobs.Add(1)
		return nil
	case <-timer.C:
		return ErrJobTimeout
	case <-p.ctx.Done():
		return ErrPoolClosed
	}
}

// Results returns the result channel
func (p *WorkerPool) Results() <-chan Result {
	return p.resultChan
}

// Shutdown gracefully shuts down the worker pool
func (p *WorkerPool) Shutdown(timeout time.Duration) error {
	// Stop accepting new jobs
	p.cancel()
	close(p.jobQueue)

	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		close(p.resultChan)
		return nil
	case <-time.After(timeout):
		return errors.New("shutdown timeout exceeded")
	}
}

// GetMetrics returns current pool metrics
func (p *WorkerPool) GetMetrics() PoolMetrics {
	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()

	return PoolMetrics{
		JobsSubmitted: atomic.Int64{},
		JobsCompleted: atomic.Int64{},
		JobsFailed:    atomic.Int64{},
		JobsRetried:   atomic.Int64{},
		TotalDuration: atomic.Int64{},
		ActiveWorkers: atomic.Int32{},
		QueuedJobs:    atomic.Int32{},
		jobDurations:  append([]time.Duration{}, p.metrics.jobDurations...),
	}
}

// GetQueueSize returns current queue size
func (p *WorkerPool) GetQueueSize() int {
	return int(p.metrics.QueuedJobs.Load())
}

// GetActiveWorkers returns number of active workers
func (p *WorkerPool) GetActiveWorkers() int {
	return int(p.metrics.ActiveWorkers.Load())
}

// recordDuration records job duration for metrics
func (m *PoolMetrics) recordDuration(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.jobDurations = append(m.jobDurations, d)

	// Keep only last 1000 durations
	if len(m.jobDurations) > 1000 {
		m.jobDurations = m.jobDurations[len(m.jobDurations)-1000:]
	}
}

// GetAverageDuration returns average job duration
func (m *PoolMetrics) GetAverageDuration() time.Duration {
	completed := m.JobsCompleted.Load()
	if completed == 0 {
		return 0
	}
	return time.Duration(m.TotalDuration.Load() / completed)
}

// GetSuccessRate returns job success rate (0-1)
func (m *PoolMetrics) GetSuccessRate() float64 {
	completed := m.JobsCompleted.Load()
	failed := m.JobsFailed.Load()
	total := completed + failed

	if total == 0 {
		return 0
	}

	return float64(completed) / float64(total)
}
