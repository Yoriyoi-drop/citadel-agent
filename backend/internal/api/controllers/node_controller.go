package controllers

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"citadel-agent/backend/internal/models"
	"citadel-agent/backend/internal/services"
)

// NodeController handles node-related HTTP requests
type NodeController struct {
	service *services.NodeService
}

// NewNodeController creates a new node controller
func NewNodeController(nodeService *services.NodeService) *NodeController {
	return &NodeController{
		service: nodeService,
	}
}

// CreateNode creates a new node
func (c *NodeController) CreateNode(ctx *fiber.Ctx) error {
	node := new(models.Node)
	if err := ctx.BodyParser(node); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	if err := c.service.CreateNode(node); err != nil {
		log.Printf("Error creating node: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot create node",
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(node)
}

// GetNode retrieves a node by ID
func (c *NodeController) GetNode(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	node, err := c.service.GetNode(id)
	if err != nil {
		log.Printf("Error getting node %s: %v", id, err)
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Node not found",
		})
	}

	return ctx.JSON(node)
}

// GetNodes retrieves all nodes for a workflow
func (c *NodeController) GetNodes(ctx *fiber.Ctx) error {
	workflowID := ctx.Params("workflowId") // Or get from query param

	nodes, err := c.service.GetNodes(workflowID)
	if err != nil {
		log.Printf("Error getting nodes for workflow %s: %v", workflowID, err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot retrieve nodes",
		})
	}

	return ctx.JSON(nodes)
}

// UpdateNode updates a node
func (c *NodeController) UpdateNode(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	node, err := c.service.GetNode(id)
	if err != nil {
		log.Printf("Error getting node %s: %v", id, err)
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Node not found",
		})
	}

	if err := ctx.BodyParser(node); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}
	// Keep the original ID
	node.ID = id

	if err := c.service.UpdateNode(node); err != nil {
		log.Printf("Error updating node %s: %v", id, err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot update node",
		})
	}

	return ctx.JSON(node)
}

// DeleteNode deletes a node by ID
func (c *NodeController) DeleteNode(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	if err := c.service.DeleteNode(id); err != nil {
		log.Printf("Error deleting node %s: %v", id, err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot delete node",
		})
	}

	return ctx.Status(fiber.StatusNoContent).JSON(fiber.Map{
		"message": "Node deleted successfully",
	})
}