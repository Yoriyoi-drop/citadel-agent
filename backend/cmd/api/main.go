package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	// "github.com/spf13/viper" // TODO: Re-enable when config is implemented

	"github.com/citadel-agent/backend/internal/config"
	"github.com/citadel-agent/backend/internal/nodes"
	"github.com/citadel-agent/backend/internal/workflow/core/engine"
)

func main() {
	// Initialize configuration
	cfg := config.LoadConfig()

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Custom error handling
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSAllowedOrigins,
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Initialize node factory and register all node types
	nodeFactory := nodes.GetNodeFactory()

	// Initialize workflow engine
	// Note: For now, we'll use a simple in-memory implementation
	_ = engine.NewEngine(&engine.Config{
		Parallelism:  10,
		Logger:       nil, // Initialize logger here
		Storage:      nil, // Initialize storage here
		NodeRegistry: nodeFactory,
	}) // TODO: Use workflowEngine when workflow routes are implemented

	// API Routes
	api := app.Group("/api/v1")

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":                "ok",
			"service":               "citadel-api",
			"version":               "1.0.0",
			"timestamp":             time.Now().Unix(),
			"node_types_registered": len(nodeFactory.ListNodeTypes()),
		})
	})

	// Simple workflow execution route
	api.Post("/workflows/execute", func(c *fiber.Ctx) error {
		var req struct {
			WorkflowID string                 `json:"workflow_id"`
			Inputs     map[string]interface{} `json:"inputs"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
		}

		// For now, return a mock response
		return c.JSON(fiber.Map{
			"success":      true,
			"execution_id": fmt.Sprintf("exec_%d", time.Now().Unix()),
			"message":      "Workflow execution started",
			"timestamp":    time.Now().Unix(),
		})
	})

	// Simple nodes route
	api.Get("/nodes", func(c *fiber.Ctx) error {
		nodeTypes := nodeFactory.ListNodeTypes()
		return c.JSON(fiber.Map{
			"success":    true,
			"node_types": nodeTypes,
			"count":      len(nodeTypes),
			"timestamp":  time.Now().Unix(),
		})
	})

	// Root route
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Welcome to Citadel Agent API",
			"status":  "running",
			"docs":    "/api/v1/docs", // Placeholder for future docs
		})
	})

	// Auto-open browser if not in production
	if os.Getenv("APP_ENV") != "production" {
		go func() {
			// Open frontend URL instead of backend
			startBrowser("http://localhost:5173")
		}()
	}

	// Start server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting Citadel API server on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

// startBrowser opens the default browser to the given URL
func startBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin": // macOS
		err = exec.Command("open", url).Start()
	default:
		fmt.Printf("Unsupported platform. Open manually: %s\n", url)
	}

	if err != nil {
		fmt.Printf("Could not open browser: %v\n", err)
	}
}
