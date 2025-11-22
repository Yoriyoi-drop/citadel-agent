package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/citadel-agent/backend/internal/engine"
	"github.com/citadel-agent/backend/internal/models"
	"github.com/google/uuid"
)

// ExecutionManagerService manages workflow execution orchestration
type ExecutionManagerService struct {
	workflowService *WorkflowService
	nodeService     *NodeService
	executionService *ExecutionService
	runner          *engine.Runner
	nodeRegistry    *engine.NodeRegistry
	executionCache  map[string]*engine.ExecutionContext
	cacheMutex      sync.RWMutex
}

// NewExecutionManagerService creates a new execution manager service
func NewExecutionManagerService(
	workflowService *WorkflowService,
	nodeService *NodeService,
	executionService *ExecutionService,
	runner *engine.Runner,
	nodeRegistry *engine.NodeRegistry,
) *ExecutionManagerService {
	return &ExecutionManagerService{
		workflowService: workflowService,
		nodeService:     nodeService,
		executionService: executionService,
		runner:          runner,
		nodeRegistry:    nodeRegistry,
		executionCache:  make(map[string]*engine.ExecutionContext),
	}
}

// ExecuteWorkflow executes a workflow asynchronously
func (s *ExecutionManagerService) ExecuteWorkflow(workflowID string, variables map[string]interface{}) (string, error) {
	if workflowID == "" {
		return "", errors.New("workflow ID is required")
	}
	
	// Get the workflow with nodes
	workflow, err := s.workflowService.GetWorkflow(workflowID)
	if err != nil {
		return "", fmt.Errorf("failed to get workflow: %w", err)
	}
	
	// Create execution record
	execution := &models.Execution{
		ID:         uuid.New().String(),
		WorkflowID: workflowID,
		Status:     "running",
		StartedAt:  time.Now(),
	}
	
	if err := s.executionService.CreateExecution(execution); err != nil {
		return "", fmt.Errorf("failed to create execution record: %w", err)
	}
	
	// Convert the models.Workflow to engine.Workflow
	engineWorkflow := &engine.Workflow{
		ID:          workflow.ID,
		Name:        workflow.Name,
		Description: workflow.Description,
		CreatedAt:   workflow.CreatedAt,
		UpdatedAt:   workflow.UpdatedAt,
	}
	
	// Convert nodes
	for _, node := range workflow.Nodes {
		engineNode := engine.Node{
			ID:       node.ID,
			Type:     node.Type,
			Name:     node.Name,
			Settings: node.Settings,
			Inputs:   []string{}, // Will be determined by dependency resolver
			Outputs:  []string{}, // Will be determined by dependency resolver
		}
		engineWorkflow.Nodes = append(engineWorkflow.Nodes, engineNode)
	}
	
	// Execute workflow asynchronously
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute) // Set a reasonable timeout
		defer cancel()
		
		// Execute the workflow
		result, err := s.runner.RunWorkflow(ctx, engineWorkflow, variables)
		if err != nil {
			log.Printf("Error executing workflow %s: %v", workflowID, err)
			// Mark execution as failed
			if updateErr := s.executionService.FailExecution(execution.ID, err.Error()); updateErr != nil {
				log.Printf("Error updating execution status to failed: %v", updateErr)
			}
			return
		}
		
		// Mark execution as completed
		if result.Status == "completed" {
			if updateErr := s.executionService.CompleteExecution(execution.ID, result.Results); updateErr != nil {
				log.Printf("Error updating execution status to completed: %v", updateErr)
			}
		} else {
			// If it's not completed, it might be cancelled or failed
			if updateErr := s.executionService.UpdateExecution(&models.Execution{
				ID:     execution.ID,
				Status: result.Status,
				Error:  result.Error,
			}); updateErr != nil {
				log.Printf("Error updating execution status: %v", updateErr)
			}
		}
	}()
	
	return execution.ID, nil
}

// ExecuteWorkflowSync executes a workflow synchronously
func (s *ExecutionManagerService) ExecuteWorkflowSync(workflowID string, variables map[string]interface{}) (*engine.Execution, error) {
	if workflowID == "" {
		return nil, errors.New("workflow ID is required")
	}
	
	// Get the workflow with nodes
	workflow, err := s.workflowService.GetWorkflow(workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}
	
	// Convert the models.Workflow to engine.Workflow
	engineWorkflow := &engine.Workflow{
		ID:          workflow.ID,
		Name:        workflow.Name,
		Description: workflow.Description,
		CreatedAt:   workflow.CreatedAt,
		UpdatedAt:   workflow.UpdatedAt,
	}
	
	// Convert nodes
	for _, node := range workflow.Nodes {
		engineNode := engine.Node{
			ID:       node.ID,
			Type:     node.Type,
			Name:     node.Name,
			Settings: node.Settings,
			Inputs:   []string{}, // Will be determined by dependency resolver
			Outputs:  []string{}, // Will be determined by dependency resolver
		}
		engineWorkflow.Nodes = append(engineWorkflow.Nodes, engineNode)
	}
	
	// Execute the workflow
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	
	result, err := s.runner.RunWorkflow(ctx, engineWorkflow, variables)
	if err != nil {
		return nil, fmt.Errorf("workflow execution failed: %w", err)
	}
	
	return result, nil
}

// GetExecutionStatus retrieves the status of a workflow execution
func (s *ExecutionManagerService) GetExecutionStatus(executionID string) (*models.Execution, error) {
	if executionID == "" {
		return nil, errors.New("execution ID is required")
	}
	
	return s.executionService.GetExecution(executionID)
}

// GetExecutionResults retrieves the results of a workflow execution
func (s *ExecutionManagerService) GetExecutionResults(executionID string) (map[string]interface{}, error) {
	if executionID == "" {
		return nil, errors.New("execution ID is required")
	}
	
	execution, err := s.executionService.GetExecution(executionID)
	if err != nil {
		return nil, fmt.Errorf("execution not found: %w", err)
	}
	
	return execution.Result, nil
}

// CancelExecution attempts to cancel a running workflow execution
func (s *ExecutionManagerService) CancelExecution(executionID string) error {
	if executionID == "" {
		return errors.New("execution ID is required")
	}
	
	// Check if execution exists and is running
	execution, err := s.executionService.GetExecution(executionID)
	if err != nil {
		return fmt.Errorf("execution not found: %w", err)
	}
	
	if execution.Status != "running" {
		return errors.New("execution is not running and cannot be cancelled")
	}
	
	// Update execution status to cancelled
	return s.executionService.CancelExecution(executionID)
}

// GetWorkflowExecutions retrieves all executions for a workflow
func (s *ExecutionManagerService) GetWorkflowExecutions(workflowID string) ([]*models.Execution, error) {
	if workflowID == "" {
		return nil, errors.New("workflow ID is required")
	}
	
	return s.executionService.GetExecutions(workflowID)
}

// GetRecentExecutions retrieves recent workflow executions
func (s *ExecutionManagerService) GetRecentExecutions(limit int) ([]*models.Execution, error) {
	if limit <= 0 || limit > 100 {
		return nil, errors.New("limit must be between 1 and 100")
	}
	
	return s.executionService.GetRecentExecutions(limit)
}

// GetRunningExecutions retrieves all currently running executions
func (s *ExecutionManagerService) GetRunningExecutions() ([]*models.Execution, error) {
	return s.executionService.GetRunningExecutions()
}

// GetExecutionStats retrieves execution statistics for a workflow
func (s *ExecutionManagerService) GetExecutionStats(workflowID string) (map[string]interface{}, error) {
	if workflowID == "" {
		return nil, errors.New("workflow ID is required")
	}
	
	return s.executionService.GetExecutionStats(workflowID)
}

// RetryExecution retries a failed execution
func (s *ExecutionManagerService) RetryExecution(executionID string, variables map[string]interface{}) (string, error) {
	if executionID == "" {
		return "", errors.New("execution ID is required")
	}
	
	// Get the original execution
	originalExecution, err := s.executionService.GetExecution(executionID)
	if err != nil {
		return "", fmt.Errorf("execution not found: %w", err)
	}
	
	// Get the associated workflow
	workflow, err := s.workflowService.GetWorkflow(originalExecution.WorkflowID)
	if err != nil {
		return "", fmt.Errorf("failed to get workflow: %w", err)
	}
	
	// Create a new execution record for the retry
	newExecution := &models.Execution{
		ID:         uuid.New().String(),
		WorkflowID: originalExecution.WorkflowID,
		Status:     "running",
		StartedAt:  time.Now(),
	}
	
	if err := s.executionService.CreateExecution(newExecution); err != nil {
		return "", fmt.Errorf("failed to create retry execution record: %w", err)
	}
	
	// Execute the workflow asynchronously
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		
		// Convert the models.Workflow to engine.Workflow
		engineWorkflow := &engine.Workflow{
			ID:          workflow.ID,
			Name:        workflow.Name,
			Description: workflow.Description,
			CreatedAt:   workflow.CreatedAt,
			UpdatedAt:   workflow.UpdatedAt,
		}
		
		// Convert nodes
		for _, node := range workflow.Nodes {
			engineNode := engine.Node{
				ID:       node.ID,
				Type:     node.Type,
				Name:     node.Name,
				Settings: node.Settings,
				Inputs:   []string{}, // Will be determined by dependency resolver
				Outputs:  []string{}, // Will be determined by dependency resolver
			}
			engineWorkflow.Nodes = append(engineWorkflow.Nodes, engineNode)
		}
		
		// Execute the workflow
		result, err := s.runner.RunWorkflow(ctx, engineWorkflow, variables)
		if err != nil {
			log.Printf("Error retrying workflow execution %s: %v", newExecution.ID, err)
			// Mark execution as failed
			if updateErr := s.executionService.FailExecution(newExecution.ID, err.Error()); updateErr != nil {
				log.Printf("Error updating retry execution status to failed: %v", updateErr)
			}
			return
		}
		
		// Mark execution as completed
		if result.Status == "completed" {
			if updateErr := s.executionService.CompleteExecution(newExecution.ID, result.Results); updateErr != nil {
				log.Printf("Error updating retry execution status to completed: %v", updateErr)
			}
		} else {
			// If it's not completed, it might be cancelled or failed
			if updateErr := s.executionService.UpdateExecution(&models.Execution{
				ID:     newExecution.ID,
				Status: result.Status,
				Error:  result.Error,
			}); updateErr != nil {
				log.Printf("Error updating retry execution status: %v", updateErr)
			}
		}
	}()
	
	return newExecution.ID, nil
}

// StopAllExecutions stops all running executions (useful for maintenance)
func (s *ExecutionManagerService) StopAllExecutions() error {
	runningExecutions, err := s.executionService.GetRunningExecutions()
	if err != nil {
		return fmt.Errorf("failed to get running executions: %w", err)
	}
	
	for _, execution := range runningExecutions {
		if err := s.executionService.CancelExecution(execution.ID); err != nil {
			log.Printf("Error cancelling execution %s: %v", execution.ID, err)
			// Continue with other executions even if one fails
		}
	}
	
	return nil
}

// CleanupOldExecutions removes executions older than the specified number of days
func (s *ExecutionManagerService) CleanupOldExecutions(days int) error {
	if days <= 0 {
		return errors.New("days must be greater than 0")
	}
	
	// This would typically involve a background job to cleanup old executions
	// For now, we'll just return a placeholder implementation
	// In a real system, this might call a function to delete executions older than X days
	
	log.Printf("Cleanup requested for executions older than %d days", days)
	return nil
}