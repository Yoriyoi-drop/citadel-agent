package controllers

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"citadel-agent/backend/internal/models"
	"citadel-agent/backend/internal/services"
)

// WorkflowController handles workflow-related HTTP requests
type WorkflowController struct {
	workflowService *services.WorkflowService
}

// NewWorkflowController creates a new instance of WorkflowController
func NewWorkflowController(workflowService *services.WorkflowService) *WorkflowController {
	return &WorkflowController{
		workflowService: workflowService,
	}
}

// CreateWorkflow creates a new workflow
// @Summary Create a new workflow
// @Description Create a new workflow with the provided details
// @Tags workflows
// @Accept json
// @Produce json
// @Param workflow body models.Workflow true "Workflow data"
// @Success 201 {object} models.Workflow
// @Failure 400 {object} map[string]string
// @Router /workflows [post]
func (wc *WorkflowController) CreateWorkflow(c *fiber.Ctx) error {
	// Parse request body
	workflow := new(models.Workflow)
	if err := c.BodyParser(workflow); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if workflow.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Name is required",
		})
	}

	// Generate ID if not provided
	if workflow.ID == "" {
		workflow.ID = uuid.New().String()
	}

	// Set timestamps
	workflow.CreatedAt = c.Locals("now").(int64)
	workflow.UpdatedAt = c.Locals("now").(int64)

	// Create workflow
	createdWorkflow, err := wc.workflowService.CreateWorkflow(c.Context(), workflow)
	if err != nil {
		log.Printf("Error creating workflow: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create workflow",
		})
	}

	return c.Status(201).JSON(createdWorkflow)
}

// GetWorkflow retrieves a workflow by ID
// @Summary Get a workflow by ID
// @Description Get a workflow by its unique ID
// @Tags workflows
// @Produce json
// @Param id path string true "Workflow ID"
// @Success 200 {object} models.Workflow
// @Failure 404 {object} map[string]string
// @Router /workflows/{id} [get]
func (wc *WorkflowController) GetWorkflow(c *fiber.Ctx) error {
	workflowID := c.Params("id")

	workflow, err := wc.workflowService.GetWorkflowByID(c.Context(), workflowID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Workflow not found",
		})
	}

	return c.JSON(workflow)
}

// GetWorkflows retrieves all workflows with optional pagination
// @Summary Get all workflows
// @Description Get all workflows with optional pagination
// @Tags workflows
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {array} models.Workflow
// @Router /workflows [get]
func (wc *WorkflowController) GetWorkflows(c *fiber.Ctx) error {
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	workflows, err := wc.workflowService.GetWorkflows(c.Context(), page, limit)
	if err != nil {
		log.Printf("Error getting workflows: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get workflows",
		})
	}

	return c.JSON(workflows)
}

// UpdateWorkflow updates an existing workflow
// @Summary Update a workflow
// @Description Update an existing workflow by ID
// @Tags workflows
// @Accept json
// @Produce json
// @Param id path string true "Workflow ID"
// @Param workflow body models.Workflow true "Workflow data"
// @Success 200 {object} models.Workflow
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /workflows/{id} [put]
func (wc *WorkflowController) UpdateWorkflow(c *fiber.Ctx) error {
	workflowID := c.Params("id")

	// Check if workflow exists
	_, err := wc.workflowService.GetWorkflowByID(c.Context(), workflowID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Workflow not found",
		})
	}

	// Parse request body
	workflow := new(models.Workflow)
	if err := c.BodyParser(workflow); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update required fields
	workflow.ID = workflowID
	workflow.UpdatedAt = c.Locals("now").(int64)

	updatedWorkflow, err := wc.workflowService.UpdateWorkflow(c.Context(), workflow)
	if err != nil {
		log.Printf("Error updating workflow: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update workflow",
		})
	}

	return c.JSON(updatedWorkflow)
}

// DeleteWorkflow deletes a workflow by ID
// @Summary Delete a workflow
// @Description Delete a workflow by its unique ID
// @Tags workflows
// @Produce json
// @Param id path string true "Workflow ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /workflows/{id} [delete]
func (wc *WorkflowController) DeleteWorkflow(c *fiber.Ctx) error {
	workflowID := c.Params("id")

	err := wc.workflowService.DeleteWorkflow(c.Context(), workflowID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Workflow not found",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Workflow deleted successfully",
	})
}

// ExecuteWorkflow triggers the execution of a workflow
// @Summary Execute a workflow
// @Description Execute a workflow by its unique ID
// @Tags workflows
// @Produce json
// @Param id path string true "Workflow ID"
// @Param variables body map[string]interface{} false "Execution variables"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /workflows/{id}/execute [post]
func (wc *WorkflowController) ExecuteWorkflow(c *fiber.Ctx) error {
	workflowID := c.Params("id")

	// Check if workflow exists
	workflow, err := wc.workflowService.GetWorkflowByID(c.Context(), workflowID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Workflow not found",
		})
	}

	// Parse execution variables from request body
	variables := make(map[string]interface{})
	if err := c.BodyParser(&variables); err != nil && err.Error() != "request body is nil" {
		log.Printf("Error parsing execution variables: %v", err)
	}

	// Execute the workflow
	execution, err := wc.workflowService.ExecuteWorkflow(c.Context(), workflow, variables)
	if err != nil {
		log.Printf("Error executing workflow: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to execute workflow",
		})
	}

	return c.JSON(fiber.Map{
		"execution_id": execution.ID,
		"status":       execution.Status,
		"started_at":   execution.StartedAt,
	})
}