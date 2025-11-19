package services

import (
	"errors"
	"fmt"
	"log"
	"time"

	"citadel-agent/backend/internal/models"
	"citadel-agent/backend/internal/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WorkflowService handles workflow-related business logic
type WorkflowService struct {
	repo *repositories.WorkflowRepository
	nodeRepo *repositories.NodeRepository
	repositoryFactory *repositories.RepositoryFactory
	executionService *ExecutionService
}

// NewWorkflowService creates a new workflow service
func NewWorkflowService(db *gorm.DB) *WorkflowService {
	repositoryFactory := repositories.NewRepositoryFactory(db)

	return &WorkflowService{
		repo: repositoryFactory.GetWorkflowRepository(),
		nodeRepo: repositoryFactory.GetNodeRepository(),
		repositoryFactory: repositoryFactory,
		executionService: NewExecutionService(db),
	}
}

// CreateWorkflow creates a new workflow with validation
func (s *WorkflowService) CreateWorkflow(workflow *models.Workflow) error {
	// Validate input
	if workflow.Name == "" {
		return errors.New("workflow name is required")
	}

	// Generate ID if not provided
	if workflow.ID == "" {
		workflow.ID = uuid.New().String()
	}

	// Set timestamps
	workflow.CreatedAt = time.Now()
	workflow.UpdatedAt = time.Now()

	// Begin transaction
	tx := s.repo.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create workflow
	if err := tx.Create(workflow).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create workflow: %w", err)
	}

	return tx.Commit().Error
}

// GetWorkflow retrieves a workflow by ID with nodes
func (s *WorkflowService) GetWorkflow(id string) (*models.Workflow, error) {
	// Validate ID
	if id == "" {
		return nil, errors.New("workflow ID is required")
	}

	workflow, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}

	// Load associated nodes
	nodes, err := s.nodeRepo.GetByWorkflowID(id)
	if err != nil {
		log.Printf("Error loading nodes for workflow %s: %v", id, err)
		// Don't fail the entire operation just because of node loading
	}
	workflow.Nodes = make([]models.Node, len(nodes))
	for i, node := range nodes {
		workflow.Nodes[i] = *node
	}

	return workflow, nil
}

// GetWorkflows retrieves all workflows with optional pagination
func (s *WorkflowService) GetWorkflows() ([]*models.Workflow, error) {
	return s.repo.GetAll()
}

// GetExecutionsByWorkflowID retrieves all executions for a specific workflow
// This method is kept for backward compatibility - it should delegate to execution service
func (s *WorkflowService) GetExecutionsByWorkflowID(workflowID string) ([]*models.Execution, error) {
	if workflowID == "" {
		return nil, errors.New("workflow ID is required")
	}

	if s.executionService == nil {
		return nil, errors.New("execution service is not initialized")
	}

	return s.executionService.GetExecutions(workflowID)
}

// GetWorkflowsWithPagination retrieves workflows with pagination
func (s *WorkflowService) GetWorkflowsWithPagination(offset, limit int) ([]*models.Workflow, error) {
	if offset < 0 || limit <= 0 || limit > 100 {
		return nil, errors.New("invalid pagination parameters")
	}

	return s.repo.GetAllWithPagination(offset, limit)
}

// GetWorkflowsByUser retrieves workflows for a specific user
func (s *WorkflowService) GetWorkflowsByUser(userID string) ([]*models.Workflow, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	return s.repo.GetByUserID(userID)
}

// GetWorkflowsByUserWithPagination retrieves workflows for a specific user with pagination
func (s *WorkflowService) GetWorkflowsByUserWithPagination(userID string, offset, limit int) ([]*models.Workflow, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}
	if offset < 0 || limit <= 0 || limit > 100 {
		return nil, errors.New("invalid pagination parameters")
	}

	return s.repo.GetByUserIDWithPagination(userID, offset, limit)
}

// GetWorkflowsByStatus retrieves workflows by status
func (s *WorkflowService) GetWorkflowsByStatus(status string) ([]*models.Workflow, error) {
	if status == "" {
		return nil, errors.New("status is required")
	}

	return s.repo.GetByStatus(status)
}

// UpdateWorkflow updates a workflow with validation
func (s *WorkflowService) UpdateWorkflow(workflow *models.Workflow) error {
	// Validate input
	if workflow.ID == "" {
		return errors.New("workflow ID is required")
	}
	if workflow.Name == "" {
		return errors.New("workflow name is required")
	}

	// Check if workflow exists
	existing, err := s.repo.GetByID(workflow.ID)
	if err != nil {
		return fmt.Errorf("workflow not found: %w", err)
	}

	// Update only allowed fields
	existing.Name = workflow.Name
	existing.Description = workflow.Description
	existing.Status = workflow.Status
	existing.OwnerID = workflow.OwnerID
	existing.UpdatedAt = time.Now()

	return s.repo.Update(existing)
}

// DeleteWorkflow deletes a workflow by ID and its associated nodes
func (s *WorkflowService) DeleteWorkflow(id string) error {
	if id == "" {
		return errors.New("workflow ID is required")
	}

	// Begin transaction to ensure consistency
	tx := s.repo.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete associated nodes first
	if err := s.nodeRepo.DeleteByWorkflowID(id); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete associated nodes: %w", err)
	}

	// Delete the workflow
	if err := tx.Delete(&models.Workflow{}, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete workflow: %w", err)
	}

	return tx.Commit().Error
}

// CountWorkflows counts all workflows
func (s *WorkflowService) CountWorkflows() (int64, error) {
	return s.repo.Count()
}

// CountWorkflowsByUser counts workflows for a specific user
func (s *WorkflowService) CountWorkflowsByUser(userID string) (int64, error) {
	if userID == "" {
		return 0, errors.New("user ID is required")
	}

	return s.repo.CountByUserID(userID)
}

// SearchWorkflows searches workflows by name
func (s *WorkflowService) SearchWorkflows(name string) ([]*models.Workflow, error) {
	if name == "" {
		return nil, errors.New("search term is required")
	}

	return s.repo.SearchByName(name)
}

// ActivateWorkflow sets workflow status to active
func (s *WorkflowService) ActivateWorkflow(id string) error {
	if id == "" {
		return errors.New("workflow ID is required")
	}

	workflow, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("workflow not found: %w", err)
	}

	workflow.Status = "active"
	workflow.UpdatedAt = time.Now()

	return s.repo.Update(workflow)
}

// DeactivateWorkflow sets workflow status to inactive
func (s *WorkflowService) DeactivateWorkflow(id string) error {
	if id == "" {
		return errors.New("workflow ID is required")
	}

	workflow, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("workflow not found: %w", err)
	}

	workflow.Status = "inactive"
	workflow.UpdatedAt = time.Now()

	return s.repo.Update(workflow)
}

// DuplicateWorkflow creates a copy of an existing workflow
func (s *WorkflowService) DuplicateWorkflow(id string, newName string) (*models.Workflow, error) {
	if id == "" {
		return nil, errors.New("workflow ID is required")
	}

	if newName == "" {
		return nil, errors.New("new workflow name is required")
	}

	// Get the original workflow with nodes
	original, err := s.GetWorkflow(id)
	if err != nil {
		return nil, fmt.Errorf("original workflow not found: %w", err)
	}

	// Create a new workflow based on the original
	newWorkflow := &models.Workflow{
		ID:          uuid.New().String(),
		Name:        newName,
		Description: original.Description,
		Status:      "inactive", // New workflow starts as inactive
		OwnerID:     original.OwnerID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create the new workflow
	if err := s.CreateWorkflow(newWorkflow); err != nil {
		return nil, fmt.Errorf("failed to create new workflow: %w", err)
	}

	// Copy nodes
	for _, node := range original.Nodes {
		newNode := &models.Node{
			ID:          uuid.New().String(),
			WorkflowID:  newWorkflow.ID,
			Type:        node.Type,
			Name:        node.Name,
			Description: node.Description,
			PositionX:   node.PositionX,
			PositionY:   node.PositionY,
			Settings:    node.Settings,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := s.nodeRepo.Create(newNode); err != nil {
			// If node creation fails, try to delete the workflow we just created
			s.repo.Delete(newWorkflow.ID)
			return nil, fmt.Errorf("failed to create node: %w", err)
		}
	}

	return newWorkflow, nil
}