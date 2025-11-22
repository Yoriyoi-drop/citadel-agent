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

// NodeService handles node-related business logic
type NodeService struct {
	repo *repositories.NodeRepository
	repositoryFactory *repositories.RepositoryFactory
}

// NewNodeService creates a new node service
func NewNodeService(db *gorm.DB) *NodeService {
	repositoryFactory := repositories.NewRepositoryFactory(db)

	return &NodeService{
		repo: repositoryFactory.GetNodeRepository(),
		repositoryFactory: repositoryFactory,
	}
}

// CreateNode creates a new node with validation
func (s *NodeService) CreateNode(node *models.Node) error {
	// Validate input
	if node.WorkflowID == "" {
		return errors.New("workflow ID is required")
	}
	if node.Type == "" {
		return errors.New("node type is required")
	}
	if node.Name == "" {
		return errors.New("node name is required")
	}

	// Generate ID if not provided
	if node.ID == "" {
		node.ID = uuid.New().String()
	}

	// Set timestamps
	node.CreatedAt = time.Now()
	node.UpdatedAt = time.Now()

	return s.repo.Create(node)
}

// GetNode retrieves a node by ID
func (s *NodeService) GetNode(id string) (*models.Node, error) {
	if id == "" {
		return nil, errors.New("node ID is required")
	}

	return s.repo.GetByID(id)
}

// GetNodes retrieves all nodes for a workflow
func (s *NodeService) GetNodes(workflowID string) ([]*models.Node, error) {
	if workflowID == "" {
		return nil, errors.New("workflow ID is required")
	}

	return s.repo.GetByWorkflowID(workflowID)
}

// GetNodesByType retrieves nodes by type
func (s *NodeService) GetNodesByType(nodeType string) ([]*models.Node, error) {
	if nodeType == "" {
		return nil, errors.New("node type is required")
	}

	return s.repo.GetByType(nodeType)
}

// GetNodesByTypes retrieves nodes by multiple types
func (s *NodeService) GetNodesByTypes(nodeTypes []string) ([]*models.Node, error) {
	if len(nodeTypes) == 0 {
		return nil, errors.New("at least one node type is required")
	}

	return s.repo.GetByTypes(nodeTypes)
}

// GetNodesWithPagination retrieves nodes with pagination
func (s *NodeService) GetNodesWithPagination(offset, limit int) ([]*models.Node, error) {
	if offset < 0 || limit <= 0 || limit > 100 {
		return nil, errors.New("invalid pagination parameters")
	}

	return s.repo.GetWithPagination(offset, limit)
}

// UpdateNode updates a node with validation
func (s *NodeService) UpdateNode(node *models.Node) error {
	// Validate input
	if node.ID == "" {
		return errors.New("node ID is required")
	}
	if node.Type == "" {
		return errors.New("node type is required")
	}
	if node.Name == "" {
		return errors.New("node name is required")
	}

	// Check if node exists
	existing, err := s.repo.GetByID(node.ID)
	if err != nil {
		return fmt.Errorf("node not found: %w", err)
	}

	// Update allowed fields
	existing.Type = node.Type
	existing.Name = node.Name
	existing.Description = node.Description
	existing.PositionX = node.PositionX
	existing.PositionY = node.PositionY
	existing.Settings = node.Settings
	existing.UpdatedAt = time.Now()

	return s.repo.Update(existing)
}

// DeleteNode deletes a node by ID
func (s *NodeService) DeleteNode(id string) error {
	if id == "" {
		return errors.New("node ID is required")
	}

	return s.repo.Delete(id)
}

// DeleteNodesByWorkflow deletes all nodes for a specific workflow
func (s *NodeService) DeleteNodesByWorkflow(workflowID string) error {
	if workflowID == "" {
		return errors.New("workflow ID is required")
	}

	return s.repo.DeleteByWorkflowID(workflowID)
}

// CountNodes counts all nodes
func (s *NodeService) CountNodes() (int64, error) {
	return s.repo.Count()
}

// CountNodesByWorkflow counts nodes for a specific workflow
func (s *NodeService) CountNodesByWorkflow(workflowID string) (int64, error) {
	if workflowID == "" {
		return 0, errors.New("workflow ID is required")
	}

	return s.repo.CountByWorkflowID(workflowID)
}

// ValidateNode validates a node's configuration
func (s *NodeService) ValidateNode(node *models.Node) error {
	if node == nil {
		return errors.New("node cannot be nil")
	}

	// Validate required fields
	if node.Type == "" {
		return errors.New("node type is required")
	}

	if node.Name == "" {
		return errors.New("node name is required")
	}

	// Validate workflow relationship
	if node.WorkflowID == "" {
		return errors.New("workflow ID is required")
	}

	// Specific validations based on node type
	switch node.Type {
	case "http_request":
		if node.Settings != nil {
			if url, exists := node.Settings["url"]; exists {
				if urlStr, ok := url.(string); !ok || urlStr == "" {
					return errors.New("url is required for HTTP request node")
				}
			}
		}
	case "function":
		if node.Settings != nil {
			if code, exists := node.Settings["code"]; exists {
				if codeStr, ok := code.(string); !ok || codeStr == "" {
					return errors.New("code is required for function node")
				}
			}
		}
	}

	return nil
}

// BatchCreateNodes creates multiple nodes at once
func (s *NodeService) BatchCreateNodes(nodes []*models.Node) error {
	if len(nodes) == 0 {
		return errors.New("no nodes to create")
	}

	for _, node := range nodes {
		if err := s.ValidateNode(node); err != nil {
			return fmt.Errorf("invalid node at index: %w", err)
		}

		// Generate ID if not provided
		if node.ID == "" {
			node.ID = uuid.New().String()
		}

		// Set timestamps
		node.CreatedAt = time.Now()
		node.UpdatedAt = time.Now()

		if err := s.repo.Create(node); err != nil {
			return fmt.Errorf("failed to create node %s: %w", node.Name, err)
		}
	}

	return nil
}

// GetNodeEdges retrieves the input/output edges for a node (simulated - actual implementation would be more complex)
func (s *NodeService) GetNodeEdges(nodeID string) (inputs []string, outputs []string, err error) {
	if nodeID == "" {
		return nil, nil, errors.New("node ID is required")
	}

	// In a real implementation, this would query a separate edges table
	// For now, we'll just return empty slices
	return []string{}, []string{}, nil
}