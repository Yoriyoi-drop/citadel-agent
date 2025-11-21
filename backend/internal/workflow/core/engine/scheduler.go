// workflow/core/engine/scheduler.go
package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// ScheduledWorkflow represents a scheduled workflow
type ScheduledWorkflow struct {
	ID          string                 `json:"id"`
	WorkflowID  string                 `json:"workflow_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Schedule    string                 `json:"schedule"` // Cron expression
	NextRunAt   time.Time              `json:"next_run_at"`
	LastRunAt   *time.Time             `json:"last_run_at,omitempty"`
	LastResult  *string                `json:"last_result,omitempty"`
	Status      ScheduledStatus        `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Params      map[string]interface{} `json:"params"` // Parameters to pass to the workflow
}

// ScheduledStatus represents the status of a scheduled workflow
type ScheduledStatus string

const (
	ScheduledActive   ScheduledStatus = "active"
	ScheduledPaused   ScheduledStatus = "paused"
	ScheduledDisabled ScheduledStatus = "disabled"
)

// EventTrigger represents an event-based trigger
type EventTrigger struct {
	ID            string                 `json:"id"`
	WorkflowID    string                 `json:"workflow_id"`
	Name          string                 `json:"name"`
	EventPattern  string                 `json:"event_pattern"` // Pattern to match events
	Conditions    map[string]interface{} `json:"conditions"`   // Conditions to evaluate
	Params        map[string]interface{} `json:"params"`       // Parameters to pass to the workflow
	Status        TriggerStatus          `json:"status"`
	CreatedAt     time.Time              `json:"created_at"`
	LastTriggered *time.Time             `json:"last_triggered,omitempty"`
}

// TriggerStatus represents the status of an event trigger
type TriggerStatus string

const (
	TriggerActive   TriggerStatus = "active"
	TriggerPaused   TriggerStatus = "paused"
	TriggerDisabled TriggerStatus = "disabled"
)

// Event represents an event that can trigger workflows
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// Scheduler manages scheduled and event-based workflow execution
type Scheduler struct {
	mutex            sync.RWMutex
	cronScheduler    *cron.Cron
	scheduledJobs    map[string]cron.EntryID
	eventTriggers    map[string]*EventTrigger
	eventQueue       chan *Event
	workflowEngine   *Engine
	ctx              context.Context
	cancelFunc       context.CancelFunc
	logger           Logger
}

// NewScheduler creates a new scheduler
func NewScheduler() *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	
	scheduler := &Scheduler{
		cronScheduler: cron.New(
			cron.WithSeconds(), // Use seconds precision
		),
		scheduledJobs:  make(map[string]cron.EntryID),
		eventTriggers:  make(map[string]*EventTrigger),
		eventQueue:     make(chan *Event, 100), // Buffered channel
		ctx:           ctx,
		cancelFunc:    cancel,
	}
	
	// Start the cron scheduler
	scheduler.cronScheduler.Start()
	
	return scheduler
}

// SetLogger sets the logger for the scheduler
func (s *Scheduler) SetLogger(logger Logger) {
	s.logger = logger
}

// SetWorkflowEngine sets the workflow engine
func (s *Scheduler) SetWorkflowEngine(engine *Engine) {
	s.workflowEngine = engine
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	// Start event processing goroutine
	go s.processEvents()
	
	if s.logger != nil {
		s.logger.Info("Scheduler started")
	}
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.cancelFunc()
	
	// Stop cron scheduler
	s.cronScheduler.Stop()
	
	// Close event queue
	close(s.eventQueue)
	
	if s.logger != nil {
		s.logger.Info("Scheduler stopped")
	}
}

// AddScheduledWorkflow adds a new scheduled workflow
func (s *Scheduler) AddScheduledWorkflow(workflow *ScheduledWorkflow) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if workflow.Status != ScheduledActive {
		return fmt.Errorf("cannot add scheduled workflow with status %s", workflow.Status)
	}

	// Validate cron expression
	_, err := cron.ParseStandard(workflow.Schedule)
	if err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	// Add cron job
	job := &scheduledJob{
		scheduler:    s,
		workflow:     workflow,
		triggerParam: map[string]interface{}{"scheduled": true, "timestamp": time.Now()},
	}

	entryID, err := s.cronScheduler.AddJob(workflow.Schedule, job)
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	// Store the entry ID
	s.scheduledJobs[workflow.ID] = entryID
	
	if s.logger != nil {
		s.logger.Info("Added scheduled workflow %s with schedule %s", workflow.ID, workflow.Schedule)
	}

	return nil
}

// RemoveScheduledWorkflow removes a scheduled workflow
func (s *Scheduler) RemoveScheduledWorkflow(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	entryID, exists := s.scheduledJobs[id]
	if !exists {
		return fmt.Errorf("scheduled workflow %s not found", id)
	}

	s.cronScheduler.Remove(entryID)
	delete(s.scheduledJobs, id)
	
	if s.logger != nil {
		s.logger.Info("Removed scheduled workflow %s", id)
	}

	return nil
}

// PauseScheduledWorkflow pauses a scheduled workflow
func (s *Scheduler) PauseScheduledWorkflow(id string) error {
	return s.RemoveScheduledWorkflow(id)
}

// ResumeScheduledWorkflow resumes a paused scheduled workflow
func (s *Scheduler) ResumeScheduledWorkflow(workflow *ScheduledWorkflow) error {
	if workflow.Status != ScheduledActive {
		return fmt.Errorf("workflow status must be active to resume")
	}
	
	return s.AddScheduledWorkflow(workflow)
}

// AddEventTrigger adds a new event-based trigger
func (s *Scheduler) AddEventTrigger(trigger *EventTrigger) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if trigger.Status != TriggerActive {
		return fmt.Errorf("cannot add event trigger with status %s", trigger.Status)
	}

	s.eventTriggers[trigger.ID] = trigger
	
	if s.logger != nil {
		s.logger.Info("Added event trigger %s for workflow %s", trigger.ID, trigger.WorkflowID)
	}

	return nil
}

// RemoveEventTrigger removes an event-based trigger
func (s *Scheduler) RemoveEventTrigger(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exists := s.eventTriggers[id]
	if !exists {
		return fmt.Errorf("event trigger %s not found", id)
	}

	delete(s.eventTriggers, id)
	
	if s.logger != nil {
		s.logger.Info("Removed event trigger %s", id)
	}

	return nil
}

// PauseEventTrigger pauses an event-based trigger
func (s *Scheduler) PauseEventTrigger(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	trigger, exists := s.eventTriggers[id]
	if !exists {
		return fmt.Errorf("event trigger %s not found", id)
	}

	trigger.Status = TriggerPaused
	
	if s.logger != nil {
		s.logger.Info("Paused event trigger %s", id)
	}

	return nil
}

// ResumeEventTrigger resumes a paused event-based trigger
func (s *Scheduler) ResumeEventTrigger(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	trigger, exists := s.eventTriggers[id]
	if !exists {
		return fmt.Errorf("event trigger %s not found", id)
	}

	if trigger.Status != TriggerPaused {
		return fmt.Errorf("event trigger %s is not paused", id)
	}

	trigger.Status = TriggerActive
	
	if s.logger != nil {
		s.logger.Info("Resumed event trigger %s", id)
	}

	return nil
}

// processEvents processes incoming events and triggers workflows
func (s *Scheduler) processEvents() {
	for {
		select {
		case event := <-s.eventQueue:
			if event != nil {
				s.processEvent(event)
			}
		case <-s.ctx.Done():
			return
		}
	}
}

// processEvent processes a single event and triggers matching workflows
func (s *Scheduler) processEvent(event *Event) {
	s.mutex.RLock()
	triggers := make([]*EventTrigger, 0, len(s.eventTriggers))
	for _, trigger := range s.eventTriggers {
		triggers = append(triggers, trigger)
	}
	s.mutex.RUnlock()

	for _, trigger := range triggers {
		if s.matchesEvent(trigger, event) && s.evaluateConditions(trigger, event) {
			// Trigger the workflow
			go s.triggerWorkflow(trigger.WorkflowID, map[string]interface{}{
				"event": event,
				"triggered_by": "event",
			})
		}
	}
}

// matchesEvent checks if an event matches the trigger pattern
func (s *Scheduler) matchesEvent(trigger *EventTrigger, event *Event) bool {
	// Simple pattern matching - in real implementation, this could be more complex
	pattern := trigger.EventPattern
	
	// If pattern is empty, it matches all events
	if pattern == "" {
		return true
	}
	
	// Pattern format: "type.subtype" or "*" for all
	if pattern == "*" {
		return true
	}
	
	// For now, just match the event type
	return event.Type == pattern
}

// evaluateConditions evaluates conditions for the trigger
func (s *Scheduler) evaluateConditions(trigger *EventTrigger, event *Event) bool {
	// In a real implementation, this would evaluate complex conditions
	// For now, we'll just return true as a placeholder
	// This would normally involve evaluating expressions in trigger.Conditions
	
	return true
}

// triggerWorkflow triggers a workflow execution
func (s *Scheduler) triggerWorkflow(workflowID string, params map[string]interface{}) {
	if s.workflowEngine == nil {
		if s.logger != nil {
			s.logger.Error("Workflow engine not set, cannot trigger workflow %s", workflowID)
		}
		return
	}

	// In a real implementation, we would fetch the workflow definition
	// For now, we'll just log
	if s.logger != nil {
		s.logger.Info("Triggering workflow %s with params %v", workflowID, params)
	}
	
	// TODO: Implement actual workflow triggering
	// s.workflowEngine.ExecuteWorkflow(context.Background(), workflow, params)
}

// PublishEvent publishes an event to the scheduler
func (s *Scheduler) PublishEvent(event *Event) {
	select {
	case s.eventQueue <- event:
		if s.logger != nil {
			s.logger.Info("Published event %s of type %s", event.ID, event.Type)
		}
	default:
		if s.logger != nil {
			s.logger.Warn("Event queue full, dropping event %s", event.ID)
		}
	}
}

// GetScheduledWorkflows returns all scheduled workflows
func (s *Scheduler) GetScheduledWorkflows() []*ScheduledWorkflow {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	workflows := make([]*ScheduledWorkflow, 0, len(s.scheduledJobs))
	for id := range s.scheduledJobs {
		// In real implementation, we would retrieve from storage
		// For now, we'll return empty structs
		workflow := &ScheduledWorkflow{ID: id}
		workflows = append(workflows, workflow)
	}

	return workflows
}

// GetEventTriggers returns all event triggers
func (s *Scheduler) GetEventTriggers() []*EventTrigger {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	triggers := make([]*EventTrigger, 0, len(s.eventTriggers))
	for _, trigger := range s.eventTriggers {
		triggers = append(triggers, trigger)
	}

	return triggers
}

// scheduledJob implements the cron.Job interface
type scheduledJob struct {
	scheduler    *Scheduler
	workflow     *ScheduledWorkflow
	triggerParam map[string]interface{}
}

// Run implements the cron.Job interface
func (j *scheduledJob) Run() {
	if j.scheduler.workflowEngine == nil {
		if j.scheduler.logger != nil {
			j.scheduler.logger.Error("Workflow engine not set, cannot run scheduled job %s", j.workflow.ID)
		}
		return
	}

	// Update the workflow's last run time
	now := time.Now()
	j.workflow.LastRunAt = &now
	
	if j.scheduler.logger != nil {
		j.scheduler.logger.Info("Running scheduled workflow %s at %s", j.workflow.ID, now.Format(time.RFC3339))
	}

	// Trigger the workflow execution
	go j.scheduler.triggerWorkflow(j.workflow.WorkflowID, j.triggerParam)
}