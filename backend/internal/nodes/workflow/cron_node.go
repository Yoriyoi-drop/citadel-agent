package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/citadel-agent/backend/internal/engine"
)

// CronSchedulerNodeConfig represents the configuration for a Cron Scheduler node
type CronSchedulerNodeConfig struct {
	CronExpression string                 `json:"cron_expression"` // Cron expression for scheduling
	NextAction     map[string]interface{} `json:"next_action"`     // Action to execute when cron fires
	Timezone       string                 `json:"timezone"`        // Timezone for the schedule
	Enabled        bool                   `json:"enabled"`         // Whether the scheduler is enabled
	Description    string                 `json:"description"`     // Description of the scheduled task
	MaxJobs        int                    `json:"max_jobs"`        // Maximum number of jobs to run simultaneously
}

// CronSchedulerNode represents a node that schedules tasks using cron
type CronSchedulerNode struct {
	config *CronSchedulerNodeConfig
	cron   *cron.Cron
	jobID  cron.EntryID
}

// NewCronSchedulerNode creates a new Cron Scheduler node
func NewCronSchedulerNode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Convert interface{} map to JSON and back to struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}

	var cronConfig CronSchedulerNodeConfig
	err = json.Unmarshal(jsonData, &cronConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// Validate required fields
	if cronConfig.CronExpression == "" {
		return nil, fmt.Errorf("cron_expression is required for Cron Scheduler node")
	}

	// Create cron instance with timezone if specified
	var cronOpts []cron.Option
	if cronConfig.Timezone != "" {
		tz, err := time.LoadLocation(cronConfig.Timezone)
		if err != nil {
			return nil, fmt.Errorf("invalid timezone: %v", err)
		}
		cronOpts = append(cronOpts, cron.WithLocation(tz))
	} else {
		cronOpts = append(cronOpts, cron.WithSeconds()) // Use seconds precision
	}

	c := cron.New(cronOpts...)

	return &CronSchedulerNode{
		config: &cronConfig,
		cron:   c,
	}, nil
}

// Execute implements the NodeInstance interface
// For cron scheduler, executing means starting the scheduler
func (c *CronSchedulerNode) Execute(ctx context.Context, input map[string]interface{}) (*engine.ExecutionResult, error) {
	// Update config with input if provided
	nextAction := c.config.NextAction
	if inputNextAction, ok := input["next_action"].(map[string]interface{}); ok {
		nextAction = inputNextAction
	}

	enabled := c.config.Enabled
	if inputEnabled, ok := input["enabled"].(bool); ok {
		enabled = inputEnabled
	}

	// Check if scheduler should be enabled
	if !enabled {
		return &engine.ExecutionResult{
			Status: "success",
			Data: map[string]interface{}{
				"message": "scheduler disabled, not started",
				"enabled": false,
			},
			Timestamp: time.Now(),
		}, nil
	}

	// Stop any existing job
	if c.jobID != 0 {
		c.cron.Remove(c.jobID)
	}

	// Define the function to execute when cron fires
	jobFunc := func() {
		// In a real implementation, this would trigger the workflow
		// For now, we'll just log that it was executed
		fmt.Printf("Cron job executed at %v with action: %v\n", time.Now(), nextAction)
	}

	// Add the job to cron
	jobID, err := c.cron.AddFunc(c.config.CronExpression, jobFunc)
	if err != nil {
		return &engine.ExecutionResult{
			Status:    "error",
			Error:     fmt.Sprintf("failed to schedule cron job: %v", err),
			Timestamp: time.Now(),
		}, nil
	}

	c.jobID = jobID

	// Start the cron scheduler in a goroutine
	// Note: In a real implementation, there would need to be proper lifecycle management
	// The cron would run continuously, and this node would need to manage that lifecycle
	go func() {
		// Wait for context to be cancelled to stop the cron
		<-ctx.Done()
		c.cron.Stop()
	}()

	// Start the scheduler
	c.cron.Start()

	return &engine.ExecutionResult{
		Status: "success",
		Data: map[string]interface{}{
			"message":        "cron scheduler started",
			"cron_expression": c.config.CronExpression,
			"job_id":         int(jobID),
			"next_action":    nextAction,
			"enabled":        true,
			"timestamp":      time.Now().Unix(),
		},
		Timestamp: time.Now(),
	}, nil
}

// GetType returns the type of the node
func (c *CronSchedulerNode) GetType() string {
	return "cron_scheduler"
}

// GetID returns a unique ID for the node instance
func (c *CronSchedulerNode) GetID() string {
	return "cron_" + fmt.Sprintf("%d", time.Now().Unix())
}