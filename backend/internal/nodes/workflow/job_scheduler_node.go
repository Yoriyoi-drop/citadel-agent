package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/citadel-agent/backend/internal/engine"
)

// JobSchedulerNodeConfig represents the configuration for a Job Scheduler node
type JobSchedulerNodeConfig struct {
	ScheduleType   string                 `json:"schedule_type"`    // "interval", "delay", "cron", "once"
	Interval       int                    `json:"interval"`        // Interval in seconds for interval-based jobs
	Delay          int                    `json:"delay"`           // Delay in seconds for delayed jobs
	CronExpression string                 `json:"cron_expression"` // Cron expression for cron-based jobs
	JobType        string                 `json:"job_type"`        // Type of job
	JobPayload     map[string]interface{} `json:"job_payload"`     // Payload for the job
	MaxRetries     int                    `json:"max_retries"`     // Maximum number of retries
	Timeout        int                    `json:"timeout"`         // Job timeout in seconds
	Enabled        bool                   `json:"enabled"`         // Whether the job is enabled
	ConcurrentJobs int                    `json:"concurrent_jobs"` // Number of concurrent jobs allowed
	Description    string                 `json:"description"`     // Description of the job
}

// Job represents a scheduled job
type Job struct {
	ID          string
	Type        string
	Payload     map[string]interface{}
	ExecuteAt   time.Time
	MaxRetries  int
	Retries     int
	Completed   bool
	Callback    func(map[string]interface{}) error
}

// JobSchedulerNode represents a node that schedules and manages jobs
type JobSchedulerNode struct {
	config  *JobSchedulerNodeConfig
	jobs    []*Job
	mutex   sync.RWMutex
	stopCh  chan struct{}
	stopped bool
}

// NewJobSchedulerNode creates a new Job Scheduler node
func NewJobSchedulerNode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Convert interface{} map to JSON and back to struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}

	var jobConfig JobSchedulerNodeConfig
	err = json.Unmarshal(jsonData, &jobConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// Validate required fields
	if jobConfig.ScheduleType == "" {
		jobConfig.ScheduleType = "once" // default to one-time job
	}

	if jobConfig.ScheduleType == "cron" && jobConfig.CronExpression == "" {
		return nil, fmt.Errorf("cron_expression is required when schedule_type is cron")
	}

	if jobConfig.ScheduleType == "interval" && jobConfig.Interval <= 0 {
		return nil, fmt.Errorf("interval must be greater than 0 when schedule_type is interval")
	}

	return &JobSchedulerNode{
		config: &jobConfig,
		jobs:   make([]*Job, 0),
		stopCh: make(chan struct{}),
	}, nil
}

// Execute implements the NodeInstance interface
func (j *JobSchedulerNode) Execute(ctx context.Context, input map[string]interface{}) (*engine.ExecutionResult, error) {
	// Override configuration with input values if provided
	scheduleType := j.config.ScheduleType
	if inputType, ok := input["schedule_type"].(string); ok && inputType != "" {
		scheduleType = inputType
	}

	interval := j.config.Interval
	if inputInterval, ok := input["interval"].(float64); ok {
		interval = int(inputInterval)
	}

	delay := j.config.Delay
	if inputDelay, ok := input["delay"].(float64); ok {
		delay = int(inputDelay)
	}

	jobPayload := j.config.JobPayload
	if inputPayload, ok := input["job_payload"].(map[string]interface{}); ok {
		jobPayload = inputPayload
	}

	jobType := j.config.JobType
	if inputType, ok := input["job_type"].(string); ok && inputType != "" {
		jobType = inputType
	}

	maxRetries := j.config.MaxRetries
	if inputRetries, ok := input["max_retries"].(float64); ok {
		maxRetries = int(inputRetries)
	}

	timeout := j.config.Timeout
	if inputTimeout, ok := input["timeout"].(float64); ok {
		timeout = int(inputTimeout)
	}

	enabled := j.config.Enabled
	if inputEnabled, ok := input["enabled"].(bool); ok {
		enabled = inputEnabled
	}

	// Check if scheduler should be enabled
	if !enabled {
		return &engine.ExecutionResult{
			Status: "success",
			Data: map[string]interface{}{
				"message": "job scheduler disabled, not started",
				"enabled": false,
			},
			Timestamp: time.Now(),
		}, nil
	}

	// Create the job
	job := &Job{
		ID:         fmt.Sprintf("job_%d", time.Now().UnixNano()),
		Type:       jobType,
		Payload:    jobPayload,
		MaxRetries: maxRetries,
	}

	// Schedule the job based on the schedule type
	switch scheduleType {
	case "interval":
		if interval <= 0 {
			return &engine.ExecutionResult{
				Status:    "error",
				Error:     "interval must be greater than 0 for interval jobs",
				Timestamp: time.Now(),
			}, nil
		}
		go j.runIntervalJob(ctx, job, time.Duration(interval)*time.Second, timeout)

	case "delay":
		if delay <= 0 {
			// Default to 1 second delay
			delay = 1
		}
		time.Sleep(time.Duration(delay) * time.Second)
		result, err := j.executeJob(job, timeout)
		if err != nil {
			return &engine.ExecutionResult{
				Status:    "error",
				Error:     err.Error(),
				Timestamp: time.Now(),
			}, nil
		}
		return result, nil

	case "cron":
		// For cron, we would normally parse the expression and schedule accordingly
		// For simplicity, we'll just return a message about the scheduled job
		return &engine.ExecutionResult{
			Status: "success",
			Data: map[string]interface{}{
				"message":        "cron job scheduled",
				"cron_expression": j.config.CronExpression,
				"job_id":         job.ID,
				"job_type":       job.Type,
				"payload":        job.Payload,
				"timestamp":      time.Now().Unix(),
			},
			Timestamp: time.Now(),
		}, nil

	case "once":
		fallthrough
	default:
		// Execute job immediately
		result, err := j.executeJob(job, timeout)
		if err != nil {
			return &engine.ExecutionResult{
				Status:    "error",
				Error:     err.Error(),
				Timestamp: time.Now(),
			}, nil
		}
		return result, nil
	}

	return &engine.ExecutionResult{
		Status: "success",
		Data: map[string]interface{}{
			"message":    "job scheduler started",
			"job_type":   jobType,
			"schedule":   scheduleType,
			"payload":    jobPayload,
			"timestamp":  time.Now().Unix(),
		},
		Timestamp: time.Now(),
	}, nil
}

// runIntervalJob runs a job at specified intervals
func (j *JobSchedulerNode) runIntervalJob(ctx context.Context, job *Job, interval time.Duration, timeout int) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Execute the job
			_, err := j.executeJob(job, timeout)
			if err != nil {
				// Log error but continue with the ticker
				fmt.Printf("Error executing job: %v\n", err)
			}
		case <-ctx.Done():
			// Context cancelled, stop the ticker
			return
		case <-j.stopCh:
			// Node stopped
			return
		}
	}
}

// executeJob executes a single job
func (j *JobSchedulerNode) executeJob(job *Job, timeout int) (*engine.ExecutionResult, error) {
	// Create a context with timeout
	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
	} else {
		// Default to 30 seconds if no timeout specified
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	// Simulate job execution
	// In a real implementation, this would call actual services or workflows
	result := map[string]interface{}{
		"job_id":    job.ID,
		"executed":  true,
		"timestamp": time.Now().Unix(),
		"payload":   job.Payload,
	}

	return &engine.ExecutionResult{
		Status: "success",
		Data:   result,
		Timestamp: time.Now(),
	}, nil
}

// GetType returns the type of the node
func (j *JobSchedulerNode) GetType() string {
	return "job_scheduler"
}

// GetID returns a unique ID for the node instance
func (j *JobSchedulerNode) GetID() string {
	return "job_scheduler_" + fmt.Sprintf("%d", time.Now().Unix())
}