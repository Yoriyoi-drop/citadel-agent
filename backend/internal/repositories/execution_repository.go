// citadel-agent/backend/internal/repositories/execution_repository.go
package repositories

import (
	"citadel-agent/backend/internal/models"
	"gorm.io/gorm"
)

// ExecutionRepository handles execution database operations
type ExecutionRepository struct {
	BaseRepository
}

// NewExecutionRepository creates a new execution repository instance
func NewExecutionRepository(db *gorm.DB) *ExecutionRepository {
	return &ExecutionRepository{
		BaseRepository: *NewBaseRepository(db),
	}
}

// Create creates a new execution
func (r *ExecutionRepository) Create(execution *models.Execution) error {
	return r.BaseRepository.db.Create(execution).Error
}

// GetByID retrieves an execution by ID
func (r *ExecutionRepository) GetByID(id string) (*models.Execution, error) {
	var execution models.Execution
	err := r.BaseRepository.db.Where("id = ?", id).First(&execution).Error
	if err != nil {
		return nil, err
	}
	return &execution, nil
}

// GetByWorkflowID retrieves all executions for a workflow
func (r *ExecutionRepository) GetByWorkflowID(workflowID string) ([]*models.Execution, error) {
	var executions []*models.Execution
	err := r.BaseRepository.db.Where("workflow_id = ?", workflowID).Find(&executions).Error
	if err != nil {
		return nil, err
	}
	return executions, nil
}

// GetByWorkflowIDWithPagination retrieves executions for a workflow with pagination
func (r *ExecutionRepository) GetByWorkflowIDWithPagination(workflowID string, offset, limit int) ([]*models.Execution, error) {
	var executions []*models.Execution
	err := r.BaseRepository.db.Where("workflow_id = ?", workflowID).Offset(offset).Limit(limit).Find(&executions).Error
	if err != nil {
		return nil, err
	}
	return executions, nil
}

// GetByStatus retrieves executions by status
func (r *ExecutionRepository) GetByStatus(status string) ([]*models.Execution, error) {
	var executions []*models.Execution
	err := r.BaseRepository.db.Where("status = ?", status).Find(&executions).Error
	if err != nil {
		return nil, err
	}
	return executions, nil
}

// GetByWorkflowIDAndStatus retrieves executions by workflow and status
func (r *ExecutionRepository) GetByWorkflowIDAndStatus(workflowID, status string) ([]*models.Execution, error) {
	var executions []*models.Execution
	err := r.BaseRepository.db.Where("workflow_id = ? AND status = ?", workflowID, status).Find(&executions).Error
	if err != nil {
		return nil, err
	}
	return executions, nil
}

// Update updates an execution
func (r *ExecutionRepository) Update(execution *models.Execution) error {
	return r.BaseRepository.db.Save(execution).Error
}

// Delete deletes an execution by ID
func (r *ExecutionRepository) Delete(id string) error {
	return r.BaseRepository.db.Delete(&models.Execution{}, "id = ?", id).Error
}

// DeleteByWorkflowID deletes all executions for a workflow
func (r *ExecutionRepository) DeleteByWorkflowID(workflowID string) error {
	return r.BaseRepository.db.Where("workflow_id = ?", workflowID).Delete(&models.Execution{}).Error
}

// GetRunning retrieves all running executions
func (r *ExecutionRepository) GetRunning() ([]*models.Execution, error) {
	var executions []*models.Execution
	err := r.BaseRepository.db.Where("status = ?", "running").Find(&executions).Error
	if err != nil {
		return nil, err
	}
	return executions, nil
}

// GetRecentlyCompleted retrieves recently completed executions
func (r *ExecutionRepository) GetRecentlyCompleted(limit int) ([]*models.Execution, error) {
	var executions []*models.Execution
	err := r.BaseRepository.db.Where("status = ? OR status = ?", "completed", "failed").Order("ended_at DESC").Limit(limit).Find(&executions).Error
	if err != nil {
		return nil, err
	}
	return executions, nil
}

// Count counts all executions
func (r *ExecutionRepository) Count() (int64, error) {
	var count int64
	err := r.BaseRepository.db.Model(&models.Execution{}).Count(&count).Error
	return count, err
}

// CountByWorkflowID counts executions for a specific workflow
func (r *ExecutionRepository) CountByWorkflowID(workflowID string) (int64, error) {
	var count int64
	err := r.BaseRepository.db.Model(&models.Execution{}).Where("workflow_id = ?", workflowID).Count(&count).Error
	return count, err
}

// CountByStatus counts executions by status
func (r *ExecutionRepository) CountByStatus(status string) (int64, error) {
	var count int64
	err := r.BaseRepository.db.Model(&models.Execution{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// GetExecutionResults gets execution results for a specific execution
func (r *ExecutionRepository) GetExecutionResults(executionID string) (map[string]interface{}, error) {
	execution, err := r.GetByID(executionID)
	if err != nil {
		return nil, err
	}
	
	return execution.Result, nil
}

// GetExecutionsByWorkflowAndStatus retrieves executions by workflow and status
func (r *ExecutionRepository) GetExecutionsByWorkflowAndStatus(workflowID, status string) ([]*models.Execution, error) {
	var executions []*models.Execution
	err := r.BaseRepository.db.Where("workflow_id = ? AND status = ?", workflowID, status).Find(&executions).Error
	if err != nil {
		return nil, err
	}
	return executions, nil
}