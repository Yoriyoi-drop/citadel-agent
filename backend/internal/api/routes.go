package api

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v1"

	"citadel-agent/backend/internal/api/controllers"
	"citadel-agent/backend/internal/services"
)

// SetupRoutes configures all API routes
func SetupRoutes(app *fiber.App, workflowService *services.WorkflowService, nodeService *services.NodeService) {
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
	workflowController := controllers.NewWorkflowController(workflowService)
	nodeController := controllers.NewNodeController(nodeService)

	// API versioning
	api := app.Group("/api/v1")

	// Workflow routes
	workflows := api.Group("/workflows")
	workflows.Post("/", workflowController.CreateWorkflow)
	workflows.Get("/", workflowController.GetWorkflows)
	workflows.Get("/:id", workflowController.GetWorkflow)
	workflows.Put("/:id", workflowController.UpdateWorkflow)
	workflows.Delete("/:id", workflowController.DeleteWorkflow)
	workflows.Post("/:id/execute", workflowController.ExecuteWorkflow)

	// Node routes
	nodes := api.Group("/nodes")
	nodes.Post("/", nodeController.CreateNode)
	nodes.Get("/", nodeController.GetNodes)
	nodes.Get("/:id", nodeController.GetNode)
	nodes.Put("/:id", nodeController.UpdateNode)
	nodes.Delete("/:id", nodeController.DeleteNode)

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

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "OK",
			"time":   time.Now().Unix(),
		})
	})
}