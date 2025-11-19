// citadel-agent/backend/internal/repositories/workflow_repository.go
package repositories

import (
	"citadel-agent/backend/internal/models"
	"gorm.io/gorm"
)

// WorkflowRepository handles workflow database operations
type WorkflowRepository struct {
	BaseRepository
}

// NewWorkflowRepository creates a new workflow repository instance
func NewWorkflowRepository(db *gorm.DB) *WorkflowRepository {
	return &WorkflowRepository{
		BaseRepository: *NewBaseRepository(db),
	}
}

// Create creates a new workflow
func (r *WorkflowRepository) Create(workflow *models.Workflow) error {
	return r.BaseRepository.db.Create(workflow).Error
}

// GetByID retrieves a workflow by ID
func (r *WorkflowRepository) GetByID(id string) (*models.Workflow, error) {
	var workflow models.Workflow
	err := r.BaseRepository.db.Preload("Nodes").Where("id = ?", id).First(&workflow).Error
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}

// GetAll retrieves all workflows with nodes
func (r *WorkflowRepository) GetAll() ([]*models.Workflow, error) {
	var workflows []*models.Workflow
	err := r.BaseRepository.db.Preload("Nodes").Find(&workflows).Error
	if err != nil {
		return nil, err
	}
	return workflows, nil
}

// GetAllWithPagination retrieves workflows with pagination
func (r *WorkflowRepository) GetAllWithPagination(offset, limit int) ([]*models.Workflow, error) {
	var workflows []*models.Workflow
	err := r.BaseRepository.db.Preload("Nodes").Offset(offset).Limit(limit).Find(&workflows).Error
	if err != nil {
		return nil, err
	}
	return workflows, nil
}

// Update updates a workflow
func (r *WorkflowRepository) Update(workflow *models.Workflow) error {
	return r.BaseRepository.db.Save(workflow).Error
}

// Delete deletes a workflow by ID
func (r *WorkflowRepository) Delete(id string) error {
	// First delete related nodes due to foreign key constraint
	if err := r.BaseRepository.db.Where("workflow_id = ?", id).Delete(&models.Node{}).Error; err != nil {
		return err
	}
	
	return r.BaseRepository.db.Delete(&models.Workflow{}, "id = ?", id).Error
}

// GetByUserID retrieves workflows for a specific user
func (r *WorkflowRepository) GetByUserID(userID string) ([]*models.Workflow, error) {
	var workflows []*models.Workflow
	err := r.BaseRepository.db.Preload("Nodes").Where("owner_id = ?", userID).Find(&workflows).Error
	if err != nil {
		return nil, err
	}
	return workflows, nil
}

// GetByUserIDWithPagination retrieves workflows for a specific user with pagination
func (r *WorkflowRepository) GetByUserIDWithPagination(userID string, offset, limit int) ([]*models.Workflow, error) {
	var workflows []*models.Workflow
	err := r.BaseRepository.db.Preload("Nodes").Where("owner_id = ?", userID).Offset(offset).Limit(limit).Find(&workflows).Error
	if err != nil {
		return nil, err
	}
	return workflows, nil
}

// GetByStatus retrieves workflows by status
func (r *WorkflowRepository) GetByStatus(status string) ([]*models.Workflow, error) {
	var workflows []*models.Workflow
	err := r.BaseRepository.db.Preload("Nodes").Where("status = ?", status).Find(&workflows).Error
	if err != nil {
		return nil, err
	}
	return workflows, nil
}

// GetByStatusAndUserID retrieves workflows by status and user ID
func (r *WorkflowRepository) GetByStatusAndUserID(userID, status string) ([]*models.Workflow, error) {
	var workflows []*models.Workflow
	err := r.BaseRepository.db.Preload("Nodes").Where("owner_id = ? AND status = ?", userID, status).Find(&workflows).Error
	if err != nil {
		return nil, err
	}
	return workflows, nil
}

// Count counts all workflows
func (r *WorkflowRepository) Count() (int64, error) {
	var count int64
	err := r.BaseRepository.db.Model(&models.Workflow{}).Count(&count).Error
	return count, err
}

// CountByUserID counts workflows for a specific user
func (r *WorkflowRepository) CountByUserID(userID string) (int64, error) {
	var count int64
	err := r.BaseRepository.db.Model(&models.Workflow{}).Where("owner_id = ?", userID).Count(&count).Error
	return count, err
}

// CountByStatus counts workflows by status
func (r *WorkflowRepository) CountByStatus(status string) (int64, error) {
	var count int64
	err := r.BaseRepository.db.Model(&models.Workflow{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// SearchByName searches workflows by name (case insensitive)
func (r *WorkflowRepository) SearchByName(name string) ([]*models.Workflow, error) {
	var workflows []*models.Workflow
	err := r.BaseRepository.db.Preload("Nodes").Where("LOWER(name) LIKE LOWER(?)", "%"+name+"%").Find(&workflows).Error
	if err != nil {
		return nil, err
	}
	return workflows, nil
}

// GetRecentlyCreated retrieves recently created workflows
func (r *WorkflowRepository) GetRecentlyCreated(limit int) ([]*models.Workflow, error) {
	var workflows []*models.Workflow
	err := r.BaseRepository.db.Preload("Nodes").Order("created_at DESC").Limit(limit).Find(&workflows).Error
	if err != nil {
		return nil, err
	}
	return workflows, nil
}