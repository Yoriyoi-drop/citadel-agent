package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/models"
	"github.com/citadel-agent/backend/internal/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ExecutionService handles execution-related business logic
type ExecutionService struct {
	repo *repositories.ExecutionRepository
	repositoryFactory *repositories.RepositoryFactory
}

// NewExecutionService creates a new execution service
func NewExecutionService(db *gorm.DB) *ExecutionService {
	repositoryFactory := repositories.NewRepositoryFactory(db)

	return &ExecutionService{
		repo: repositoryFactory.GetExecutionRepository(),
		repositoryFactory: repositoryFactory,
	}
}

// CreateExecution creates a new execution with validation
func (s *ExecutionService) CreateExecution(execution *models.Execution) error {
	// Validate input
	if execution.WorkflowID == "" {
		return errors.New("workflow ID is required")
	}

	// Generate ID if not provided
	if execution.ID == "" {
		execution.ID = uuid.New().String()
	}

	// Set timestamps and defaults
	execution.CreatedAt = time.Now()
	execution.UpdatedAt = time.Now()
	if execution.Status == "" {
		execution.Status = "running"
	}
	if execution.StartedAt.IsZero() {
		execution.StartedAt = time.Now()
	}

	return s.repo.Create(execution)
}

// GetExecution retrieves an execution by ID
func (s *ExecutionService) GetExecution(id string) (*models.Execution, error) {
	if id == "" {
		return nil, errors.New("execution ID is required")
	}

	return s.repo.GetByID(id)
}

// GetExecutions retrieves all executions for a workflow
func (s *ExecutionService) GetExecutions(workflowID string) ([]*models.Execution, error) {
	if workflowID == "" {
		return nil, errors.New("workflow ID is required")
	}

	return s.repo.GetByWorkflowID(workflowID)
}

// GetExecutionsWithPagination retrieves executions for a workflow with pagination
func (s *ExecutionService) GetExecutionsWithPagination(workflowID string, offset, limit int) ([]*models.Execution, error) {
	if workflowID == "" {
		return nil, errors.New("workflow ID is required")
	}
	if offset < 0 || limit <= 0 || limit > 100 {
		return nil, errors.New("invalid pagination parameters")
	}

	return s.repo.GetByWorkflowIDWithPagination(workflowID, offset, limit)
}

// GetExecutionsByStatus retrieves executions by status
func (s *ExecutionService) GetExecutionsByStatus(status string) ([]*models.Execution, error) {
	if status == "" {
		return nil, errors.New("status is required")
	}

	return s.repo.GetByStatus(status)
}

// GetExecutionsByWorkflowAndStatus retrieves executions by workflow and status
func (s *ExecutionService) GetExecutionsByWorkflowAndStatus(workflowID, status string) ([]*models.Execution, error) {
	if workflowID == "" {
		return nil, errors.New("workflow ID is required")
	}
	if status == "" {
		return nil, errors.New("status is required")
	}

	return s.repo.GetByWorkflowIDAndStatus(workflowID, status)
}

// UpdateExecution updates an execution with validation
func (s *ExecutionService) UpdateExecution(execution *models.Execution) error {
	// Validate input
	if execution.ID == "" {
		return errors.New("execution ID is required")
	}

	// Check if execution exists
	existing, err := s.repo.GetByID(execution.ID)
	if err != nil {
		return fmt.Errorf("execution not found: %w", err)
	}

	// Update allowed fields
	existing.Status = execution.Status
	existing.Result = execution.Result
	existing.Error = execution.Error
	existing.EndedAt = execution.EndedAt
	existing.UpdatedAt = time.Now()

	return s.repo.Update(existing)
}

// CompleteExecution marks an execution as completed
func (s *ExecutionService) CompleteExecution(id string, result map[string]interface{}) error {
	if id == "" {
		return errors.New("execution ID is required")
	}

	execution, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("execution not found: %w", err)
	}

	execution.Status = "completed"
	execution.Result = result
	endTime := time.Now()
	execution.EndedAt = &endTime
	execution.UpdatedAt = time.Now()

	return s.repo.Update(execution)
}

// FailExecution marks an execution as failed
func (s *ExecutionService) FailExecution(id string, error string) error {
	if id == "" {
		return errors.New("execution ID is required")
	}

	execution, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("execution not found: %w", err)
	}

	execution.Status = "failed"
	execution.Error = error
	endTime := time.Now()
	execution.EndedAt = &endTime
	execution.UpdatedAt = time.Now()

	return s.repo.Update(execution)
}

// CancelExecution marks an execution as cancelled
func (s *ExecutionService) CancelExecution(id string) error {
	if id == "" {
		return errors.New("execution ID is required")
	}

	execution, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("execution not found: %w", err)
	}

	execution.Status = "cancelled"
	endTime := time.Now()
	execution.EndedAt = &endTime
	execution.UpdatedAt = time.Now()

	return s.repo.Update(execution)
}

// DeleteExecution deletes an execution by ID
func (s *ExecutionService) DeleteExecution(id string) error {
	if id == "" {
		return errors.New("execution ID is required")
	}

	return s.repo.Delete(id)
}

// DeleteExecutionsByWorkflow deletes all executions for a specific workflow
func (s *ExecutionService) DeleteExecutionsByWorkflow(workflowID string) error {
	if workflowID == "" {
		return errors.New("workflow ID is required")
	}

	return s.repo.DeleteByWorkflowID(workflowID)
}

// CountExecutions counts all executions
func (s *ExecutionService) CountExecutions() (int64, error) {
	return s.repo.Count()
}

// CountExecutionsByWorkflow counts executions for a specific workflow
func (s *ExecutionService) CountExecutionsByWorkflow(workflowID string) (int64, error) {
	if workflowID == "" {
		return 0, errors.New("workflow ID is required")
	}

	return s.repo.CountByWorkflowID(workflowID)
}

// CountExecutionsByStatus counts executions by status
func (s *ExecutionService) CountExecutionsByStatus(status string) (int64, error) {
	if status == "" {
		return 0, errors.New("status is required")
	}

	return s.repo.CountByStatus(status)
}

// GetRecentExecutions retrieves recent executions (completed or failed)
func (s *ExecutionService) GetRecentExecutions(limit int) ([]*models.Execution, error) {
	if limit <= 0 || limit > 100 {
		return nil, errors.New("invalid limit parameter")
	}

	return s.repo.GetRecentlyCompleted(limit)
}

// GetRunningExecutions retrieves all running executions
func (s *ExecutionService) GetRunningExecutions() ([]*models.Execution, error) {
	return s.repo.GetRunning()
}

// GetExecutionResults retrieves the results of a specific execution
func (s *ExecutionService) GetExecutionResults(executionID string) (map[string]interface{}, error) {
	if executionID == "" {
		return nil, errors.New("execution ID is required")
	}

	execution, err := s.repo.GetByID(executionID)
	if err != nil {
		return nil, fmt.Errorf("execution not found: %w", err)
	}

	return execution.Result, nil
}

// GetExecutionStats retrieves execution statistics for a workflow
func (s *ExecutionService) GetExecutionStats(workflowID string) (map[string]interface{}, error) {
	if workflowID == "" {
		return nil, errors.New("workflow ID is required")
	}

	// Get total count
	totalCount, err := s.repo.CountByWorkflowID(workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Get counts by status by retrieving and filtering
	completedExecutions, err := s.repo.GetByWorkflowIDAndStatus(workflowID, "completed")
	if err != nil {
		return nil, fmt.Errorf("failed to get completed executions: %w", err)
	}
	completedCount := int64(len(completedExecutions))

	failedExecutions, err := s.repo.GetByWorkflowIDAndStatus(workflowID, "failed")
	if err != nil {
		return nil, fmt.Errorf("failed to get failed executions: %w", err)
	}
	failedCount := int64(len(failedExecutions))

	runningExecutions, err := s.repo.GetByWorkflowIDAndStatus(workflowID, "running")
	if err != nil {
		return nil, fmt.Errorf("failed to get running executions: %w", err)
	}
	runningCount := int64(len(runningExecutions))

	stats := map[string]interface{}{
		"total":     totalCount,
		"completed": completedCount,
		"failed":    failedCount,
		"running":   runningCount,
		"success_rate": 0.0,
		"execution_times": map[string]interface{}{},
	}

	if totalCount > 0 {
		successRate := float64(completedCount) / float64(totalCount) * 100
		stats["success_rate"] = successRate
	}

	// Calculate average execution time for completed workflows if possible
	averageTime := time.Duration(0)
	if completedCount > 0 {
		totalDuration := time.Duration(0)
		for _, exec := range completedExecutions {
			if exec.EndedAt != nil && exec.StartedAt.After((time.Time{})) && exec.EndedAt.After(exec.StartedAt) {
				totalDuration += exec.EndedAt.Sub(exec.StartedAt)
			}
		}
		if totalDuration > 0 {
			averageTime = totalDuration / time.Duration(completedCount)
		}
	}

	executionTimes := map[string]interface{}{
		"average_duration_ms": averageTime.Milliseconds(),
		"average_duration_s":  averageTime.Seconds(),
	}
	stats["execution_times"] = executionTimes

	return stats, nil
}

// CleanupOldExecutions removes executions older than the specified number of days
func (s *ExecutionService) CleanupOldExecutions(days int) (int64, error) {
	if days <= 0 {
		return 0, errors.New("days must be greater than 0")
	}

	// This would require a more complex query to find and delete old executions
	// For now, we'll just return a placeholder implementation
	return 0, nil
}