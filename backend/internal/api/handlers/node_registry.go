package handlers

import (
	"github.com/citadel-agent/backend/internal/nodes/loader"
	"github.com/citadel-agent/backend/internal/nodes/registry"
	"github.com/gofiber/fiber/v2"
)

// NodeRegistryHandler handles new node registry API
type NodeRegistryHandler struct {
	registry *registry.Registry
}

// NewNodeRegistryHandler creates handler for new registry
func NewNodeRegistryHandler() *NodeRegistryHandler {
	// Load all nodes
	if err := loader.LoadAllNodes(); err != nil {
		// Log error but don't fail
		println("Warning: Failed to load some nodes:", err.Error())
	}

	return &NodeRegistryHandler{
		registry: registry.GetRegistry(),
	}
}

// ListNodes returns all registered nodes
func (h *NodeRegistryHandler) ListNodes(c *fiber.Ctx) error {
	nodes := h.registry.List()

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"nodes": nodes,
			"count": len(nodes),
		},
	})
}

// GetNode returns specific node metadata
func (h *NodeRegistryHandler) GetNode(c *fiber.Ctx) error {
	nodeID := c.Params("id")

	reg, err := h.registry.Get(nodeID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Node not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    reg.Metadata,
	})
}

// ListByCategory returns nodes by category
func (h *NodeRegistryHandler) ListByCategory(c *fiber.Ctx) error {
	category := c.Params("category")

	nodes := h.registry.ListByCategory(category)

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"category": category,
			"nodes":    nodes,
			"count":    len(nodes),
		},
	})
}

// SearchNodes searches for nodes
func (h *NodeRegistryHandler) SearchNodes(c *fiber.Ctx) error {
	query := c.Query("q")

	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Search query required",
		})
	}

	nodes := h.registry.Search(query)

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"query": query,
			"nodes": nodes,
			"count": len(nodes),
		},
	})
}

// GetCategories returns all categories
func (h *NodeRegistryHandler) GetCategories(c *fiber.Ctx) error {
	categories := h.registry.Categories()

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"categories": categories,
			"count":      len(categories),
		},
	})
}

// GetStats returns registry statistics
func (h *NodeRegistryHandler) GetStats(c *fiber.Ctx) error {
	nodes := h.registry.List()
	categories := h.registry.Categories()

	// Count by category
	categoryCount := make(map[string]int)
	for _, node := range nodes {
		categoryCount[node.Category]++
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"total_nodes":      h.registry.Count(),
			"total_categories": len(categories),
			"categories":       categories,
			"by_category":      categoryCount,
		},
	})
}
