package repositories

import (
	"citadel-agent/backend/internal/models"
)

// Repository interface that all repositories should implement
type Repository interface {
	// Common methods that repositories might implement
}

// NodeRepositoryInterface defines the interface for node operations
type NodeRepositoryInterface interface {
	Create(node *models.Node) error
	GetByID(id string) (*models.Node, error)
	GetByWorkflowID(workflowID string) ([]*models.Node, error)
	Update(node *models.Node) error
	Delete(id string) error
	DeleteByWorkflowID(workflowID string) error
	GetByType(nodeType string) ([]*models.Node, error)
	GetByTypes(nodeTypes []string) ([]*models.Node, error)
	GetWithPagination(offset, limit int) ([]*models.Node, error)
	Count() (int64, error)
	CountByWorkflowID(workflowID string) (int64, error)
}

// WorkflowRepositoryInterface defines the interface for workflow operations
type WorkflowRepositoryInterface interface {
	Create(workflow *models.Workflow) error
	GetByID(id string) (*models.Workflow, error)
	GetAll() ([]*models.Workflow, error)
	GetAllWithPagination(offset, limit int) ([]*models.Workflow, error)
	Update(workflow *models.Workflow) error
	Delete(id string) error
	GetByUserID(userID string) ([]*models.Workflow, error)
	GetByUserIDWithPagination(userID string, offset, limit int) ([]*models.Workflow, error)
	GetByStatus(status string) ([]*models.Workflow, error)
	GetByStatusAndUserID(userID, status string) ([]*models.Workflow, error)
	Count() (int64, error)
	CountByUserID(userID string) (int64, error)
	CountByStatus(status string) (int64, error)
	SearchByName(name string) ([]*models.Workflow, error)
	GetRecentlyCreated(limit int) ([]*models.Workflow, error)
}

// UserRepositoryInterface defines the interface for user operations
type UserRepositoryInterface interface {
	Create(user *models.User) error
	GetByID(id string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	Update(user *models.User) error
	Delete(id string) error
	GetAll() ([]*models.User, error)
	GetAllWithPagination(offset, limit int) ([]*models.User, error)
	GetByRole(role string) ([]*models.User, error)
	GetByStatus(status string) ([]*models.User, error)
	Count() (int64, error)
	CountByRole(role string) (int64, error)
	CountByStatus(status string) (int64, error)
	SearchByName(name string) ([]*models.User, error)
	SearchByEmail(email string) ([]*models.User, error)
	GetRecentlyCreated(limit int) ([]*models.User, error)
	GetByEmails(emails []string) ([]*models.User, error)
}

// ExecutionRepositoryInterface defines the interface for execution operations
type ExecutionRepositoryInterface interface {
	Create(execution *models.Execution) error
	GetByID(id string) (*models.Execution, error)
	GetByWorkflowID(workflowID string) ([]*models.Execution, error)
	GetByWorkflowIDWithPagination(workflowID string, offset, limit int) ([]*models.Execution, error)
	GetByStatus(status string) ([]*models.Execution, error)
	GetByWorkflowIDAndStatus(workflowID, status string) ([]*models.Execution, error)
	Update(execution *models.Execution) error
	Delete(id string) error
	DeleteByWorkflowID(workflowID string) error
	Count() (int64, error)
	CountByWorkflowID(workflowID string) (int64, error)
	CountByStatus(status string) (int64, error)
	GetRecentlyCompleted(limit int) ([]*models.Execution, error)
	GetRunning() ([]*models.Execution, error)
}