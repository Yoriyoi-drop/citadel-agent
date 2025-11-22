// citadel-agent/backend/internal/repositories/node_repository.go
package repositories

import (
	"github.com/citadel-agent/backend/internal/models"
	"gorm.io/gorm"
)

// NodeRepository handles node database operations
type NodeRepository struct {
	BaseRepository
}

// NewNodeRepository creates a new node repository instance
func NewNodeRepository(db *gorm.DB) *NodeRepository {
	return &NodeRepository{
		BaseRepository: *NewBaseRepository(db),
	}
}

// Create creates a new node
func (r *NodeRepository) Create(node *models.Node) error {
	return r.BaseRepository.db.Create(node).Error
}

// GetByID retrieves a node by ID
func (r *NodeRepository) GetByID(id string) (*models.Node, error) {
	var node models.Node
	err := r.BaseRepository.db.Where("id = ?", id).First(&node).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// GetByWorkflowID retrieves all nodes for a workflow
func (r *NodeRepository) GetByWorkflowID(workflowID string) ([]*models.Node, error) {
	var nodes []*models.Node
	err := r.BaseRepository.db.Where("workflow_id = ?", workflowID).Find(&nodes).Error
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

// Update updates a node
func (r *NodeRepository) Update(node *models.Node) error {
	return r.BaseRepository.db.Save(node).Error
}

// Delete deletes a node by ID
func (r *NodeRepository) Delete(id string) error {
	return r.BaseRepository.db.Delete(&models.Node{}, "id = ?", id).Error
}

// DeleteByWorkflowID deletes all nodes for a specific workflow
func (r *NodeRepository) DeleteByWorkflowID(workflowID string) error {
	return r.BaseRepository.db.Where("workflow_id = ?", workflowID).Delete(&models.Node{}).Error
}

// GetByType retrieves nodes by type
func (r *NodeRepository) GetByType(nodeType string) ([]*models.Node, error) {
	var nodes []*models.Node
	err := r.BaseRepository.db.Where("type = ?", nodeType).Find(&nodes).Error
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

// GetByTypes retrieves nodes by multiple types
func (r *NodeRepository) GetByTypes(nodeTypes []string) ([]*models.Node, error) {
	var nodes []*models.Node
	err := r.BaseRepository.db.Where("type IN ?", nodeTypes).Find(&nodes).Error
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

// GetWithPagination retrieves nodes with pagination
func (r *NodeRepository) GetWithPagination(offset, limit int) ([]*models.Node, error) {
	var nodes []*models.Node
	err := r.BaseRepository.db.Offset(offset).Limit(limit).Find(&nodes).Error
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

// Count counts all nodes
func (r *NodeRepository) Count() (int64, error) {
	var count int64
	err := r.BaseRepository.db.Model(&models.Node{}).Count(&count).Error
	return count, err
}

// CountByWorkflowID counts nodes for a specific workflow
func (r *NodeRepository) CountByWorkflowID(workflowID string) (int64, error) {
	var count int64
	err := r.BaseRepository.db.Model(&models.Node{}).Where("workflow_id = ?", workflowID).Count(&count).Error
	return count, err
}

// GetByWorkflowIDAndStatus is a helper function to get nodes by workflow and status (if needed for execution logic)
func (r *NodeRepository) GetByWorkflowIDAndStatus(workflowID, status string) ([]*models.Node, error) {
	// This is for node status, but nodes don't have a status field in the model
	// This would be useful if we had a node execution status
	// For now, just return all nodes for the workflow
	return r.GetByWorkflowID(workflowID)
}