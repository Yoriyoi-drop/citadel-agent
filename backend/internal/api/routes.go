package api

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/contrib/websocket"

	"citadel-agent/backend/internal/api/controllers"
	"citadel-agent/backend/internal/services"
)

// SetupRoutes configures all API routes
func SetupRoutes(
	app *fiber.App,
	workflowService *services.WorkflowService,
	nodeService *services.NodeService,
	executionService *services.ExecutionService,
	userService *services.UserService,
	executionManagerService *services.ExecutionManagerService,
) {
	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	// Add middleware to inject current time into context
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("now", time.Now().Unix())
		return c.Next()
	})

	// Create controllers
	workflowController := controllers.NewWorkflowController(workflowService, executionService, executionManagerService)
	nodeController := controllers.NewNodeController(nodeService)
	executionController := controllers.NewExecutionController(executionService, executionManagerService)
	authController := controllers.NewAuthController(userService)
	userController := controllers.NewUserController(userService)

	// API versioning
	api := app.Group("/api/v1")

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "OK",
			"time":   time.Now().Unix(),
		})
	})

	// Public routes (no authentication required)
	public := api.Group("/public")
	public.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "API is running",
			"version": "v1",
			"timestamp": time.Now().Unix(),
		})
	})

	// Authentication routes
	auth := api.Group("/auth")
	auth.Post("/register", authController.Register)
	auth.Post("/login", authController.Login)
	auth.Get("/profile", authController.Profile) // Requires authentication

	// User routes (requires authentication)
	users := api.Group("/users")
	users.Get("/", userController.GetUsers)
	users.Get("/:id", userController.GetUser)
	users.Put("/:id", userController.UpdateUser)
	users.Delete("/:id", userController.DeleteUser)

	// Workflow routes
	workflows := api.Group("/workflows")
	workflows.Post("/", workflowController.CreateWorkflow)
	workflows.Get("/", workflowController.GetWorkflows)
	workflows.Get("/:id", workflowController.GetWorkflow)
	workflows.Put("/:id", workflowController.UpdateWorkflow)
	workflows.Delete("/:id", workflowController.DeleteWorkflow)

	// Workflow execution routes
	workflows.Post("/:id/execute", workflowController.ExecuteWorkflow)
	workflows.Get("/:id/executions", workflowController.GetWorkflowExecutions)
	workflows.Get("/:id/stats", workflowController.GetWorkflowStats)

	// Node routes (child of workflows)
	workflowNodes := api.Group("/workflows/:workflowId/nodes")
	workflowNodes.Post("/", nodeController.CreateNode)
	workflowNodes.Get("/", nodeController.GetNodes)
	workflowNodes.Get("/:id", nodeController.GetNode)
	workflowNodes.Put("/:id", nodeController.UpdateNode)
	workflowNodes.Delete("/:id", nodeController.DeleteNode)

	// Execution routes
	executions := api.Group("/executions")
	executions.Get("/", executionController.GetExecutions)
	executions.Get("/:id", executionController.GetExecution)
	executions.Get("/:id/results", executionController.GetExecutionResults)
	executions.Put("/:id/cancel", executionController.CancelExecution)
	executions.Post("/:id/retry", executionController.RetryExecution)

	// Execution management routes
	executionManagement := api.Group("/execution-manager")
	executionManagement.Post("/execute", executionController.ExecuteWorkflow) // For ad-hoc executions
	executionManagement.Get("/recent", executionController.GetRecentExecutions)
	executionManagement.Get("/running", executionController.GetRunningExecutions)
	executionManagement.Get("/stats/:workflowId", executionController.GetExecutionStats)

	// WebSocket route for real-time updates
	app.Get("/ws", func(c *fiber.Ctx) error {
		// WebSocket upgrade
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return websocket.New(func(c *websocket.Conn) {
				// TODO: Implement WebSocket handler for real-time workflow updates
				// This is where clients can connect to receive live updates about workflow execution
				for {
					// Read message from browser
					mt, msg, err := c.ReadMessage()
					if err != nil {
						log.Println("read:", err)
						break
					}

					// Write message back to browser
					if err := c.WriteMessage(mt, msg); err != nil {
						log.Println("write:", err)
						break
					}
				}
			})(c)
		}
		return fiber.ErrUpgradeRequired
	})
}