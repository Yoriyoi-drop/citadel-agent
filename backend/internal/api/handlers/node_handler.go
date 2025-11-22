// backend/internal/api/handlers/node_handler.go
package handlers

import (
	"fmt"

	"github.com/citadel-agent/backend/internal/workflow/nodes"
	"github.com/gofiber/fiber/v2"
)

// NodeHandler handles node-related HTTP requests
type NodeHandler struct {
	nodeRegistry *nodes.NodeRegistry
}

// NewNodeHandler creates a new node handler
func NewNodeHandler() *NodeHandler {
	return &NodeHandler{
		nodeRegistry: nodes.GetNodeRegistry(),
	}
}

// GetNodeTypes handles GET /api/v1/node-types
func (nh *NodeHandler) GetNodeTypes(c *fiber.Ctx) error {
	nodeTypes := nh.nodeRegistry.ListNodes()
	
	// Convert to response format
	response := make([]map[string]interface{}, 0, len(nodeTypes))
	
	for name, node := range nodeTypes {
		nodeInfo := map[string]interface{}{
			"name":        name,
			"description": node.GetDescription(),
			"inputs":      node.GetInputs(),
			"outputs":     node.GetOutputs(),
		}
		response = append(response, nodeInfo)
	}
	
	return c.JSON(response)
}

// GetNodeType handles GET /api/v1/node-types/:type
func (nh *NodeHandler) GetNodeType(c *fiber.Ctx) error {
	nodeType := c.Params("type")
	
	node, err := nh.nodeRegistry.GetNode(nodeType)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("Node type '%s' not found", nodeType),
		})
	}
	
	response := map[string]interface{}{
		"name":        nodeType,
		"description": node.GetDescription(),
		"inputs":      node.GetInputs(),
		"outputs":     node.GetOutputs(),
	}
	
	return c.JSON(response)
}

// RegisterRoutes registers node handler routes
func (nh *NodeHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/node-types", nh.GetNodeTypes)
	router.Get("/node-types/:type", nh.GetNodeType)
}