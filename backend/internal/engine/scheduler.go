package engine

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// Schedule represents a workflow execution schedule
type Schedule struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	WorkflowID  string    `json:"workflow_id"`
	Cron        string    `json:"cron"` // Cron expression for scheduling
	Interval    string    `json:"interval"` // Interval-based scheduling
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Scheduler manages scheduled workflow executions
type Scheduler struct {
	runner *Runner
	tasks  map[string]*ScheduledTask
}

// ScheduledTask represents a single scheduled task
type ScheduledTask struct {
	ID          string    `json:"id"`
	Schedule    *Schedule `json:"schedule"`
	LastRun     time.Time `json:"last_run"`
	NextRun     time.Time `json:"next_run"`
	Status      string    `json:"status"` // "active", "paused", "error"
	Error       string    `json:"error,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewScheduler creates a new scheduler instance
func NewScheduler(runner *Runner) *Scheduler {
	return &Scheduler{
		runner: runner,
		tasks:  make(map[string]*ScheduledTask),
	}
}

// CreateSchedule creates a new schedule
func (s *Scheduler) CreateSchedule(schedule *Schedule) (*ScheduledTask, error) {
	if schedule.WorkflowID == "" {
		return nil, fmt.Errorf("workflow ID is required")
	}

	// Generate ID if not provided
	if schedule.ID == "" {
		schedule.ID = uuid.New().String()
	}

	// Create scheduled task
	task := &ScheduledTask{
		ID:       uuid.New().String(),
		Schedule: schedule,
		Status:   "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Calculate next run time based on cron or interval
	err := s.calculateNextRun(task)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate next run: %w", err)
	}

	// Store the task
	s.tasks[task.ID] = task

	return task, nil
}

// UpdateSchedule updates an existing schedule
func (s *Scheduler) UpdateSchedule(id string, updated *Schedule) (*ScheduledTask, error) {
	task, exists := s.tasks[id]
	if !exists {
		return nil, fmt.Errorf("schedule with ID %s not found", id)
	}

	// Update the schedule
	task.Schedule = updated
	task.UpdatedAt = time.Now()

	// Recalculate next run
	err := s.calculateNextRun(task)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate next run: %w", err)
	}

	return task, nil
}

// DeleteSchedule deletes a schedule
func (s *Scheduler) DeleteSchedule(id string) error {
	_, exists := s.tasks[id]
	if !exists {
		return fmt.Errorf("schedule with ID %s not found", id)
	}

	delete(s.tasks, id)
	return nil
}

// GetSchedule returns a specific schedule
func (s *Scheduler) GetSchedule(id string) (*ScheduledTask, error) {
	task, exists := s.tasks[id]
	if !exists {
		return nil, fmt.Errorf("schedule with ID %s not found", id)
	}

	return task, nil
}

// ListSchedules returns all schedules
func (s *Scheduler) ListSchedules() []*ScheduledTask {
	tasks := make([]*ScheduledTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// Start begins scheduling and executing tasks
func (s *Scheduler) Start(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			s.executeDueTasks()
		}
	}
}

// executeDueTasks executes all tasks that are due to run
func (s *Scheduler) executeDueTasks() {
	now := time.Now()
	
	for _, task := range s.tasks {
		if task.Status != "active" {
			continue
		}

		if now.After(task.NextRun) {
			s.executeTask(task)
			
			// Update task after execution
			task.LastRun = now
			task.UpdatedAt = now
			
			// Calculate next run
			s.calculateNextRun(task)
		}
	}
}

// executeTask executes a single scheduled task
func (s *Scheduler) executeTask(task *ScheduledTask) {
	ctx := context.Background()
	
	// In a real implementation, we would fetch the actual workflow
	// For now, we'll create a minimal workflow for testing
	workflow := &Workflow{
		ID:   task.Schedule.WorkflowID,
		Name: fmt.Sprintf("Scheduled: %s", task.Schedule.Name),
	}
	
	// Execute the workflow
	_, err := s.runner.RunWorkflow(ctx, workflow, map[string]interface{}{
		"schedule_id":     task.ID,
		"scheduled_time":  task.NextRun,
		"execution_type":  "scheduled",
	})
	
	if err != nil {
		task.Status = "error"
		task.Error = err.Error()
		log.Printf("Scheduled task %s failed: %v", task.ID, err)
	} else {
		task.Status = "active"
		task.Error = ""
	}
}

// calculateNextRun calculates the next run time for a task
// In a real implementation, this would parse cron expressions
// For this example, we'll implement a simple interval-based system
func (s *Scheduler) calculateNextRun(task *ScheduledTask) error {
	now := time.Now()
	
	if task.Schedule.Cron != "" {
		// In a real implementation, we'd parse the cron expression
		// For now, we'll default to an interval-based calculation
		log.Printf("Cron scheduling not fully implemented, using interval fallback")
	}
	
	if task.Schedule.Interval != "" {
		// Parse interval like "5m", "1h", "30s", etc.
		duration, err := time.ParseDuration(task.Schedule.Interval)
		if err != nil {
			return fmt.Errorf("invalid interval format: %w", err)
		}
		
		var baseTime time.Time
		if task.LastRun.IsZero() {
			baseTime = now
		} else {
			baseTime = task.LastRun
		}
		
		task.NextRun = baseTime.Add(duration)
	} else {
		// Default to 1 hour if no interval specified
		task.NextRun = now.Add(1 * time.Hour)
	}
	
	return nil
}

// PauseSchedule pauses a schedule
func (s *Scheduler) PauseSchedule(id string) error {
	task, exists := s.tasks[id]
	if !exists {
		return fmt.Errorf("schedule with ID %s not found", id)
	}
	
	task.Status = "paused"
	task.UpdatedAt = time.Now()
	
	return nil
}

// ResumeSchedule resumes a paused schedule
func (s *Scheduler) ResumeSchedule(id string) error {
	task, exists := s.tasks[id]
	if !exists {
		return fmt.Errorf("schedule with ID %s not found", id)
	}
	
	task.Status = "active"
	task.UpdatedAt = time.Now()
	
	// Recalculate next run
	return s.calculateNextRun(task)
}