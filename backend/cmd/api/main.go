package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"citadel-agent/backend/internal/api"
	"citadel-agent/backend/internal/config"
	"citadel-agent/backend/internal/database"
	"citadel-agent/backend/internal/engine"
	"citadel-agent/backend/internal/services"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database connection
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize services
	workflowService := services.NewWorkflowService(db.GormDB)
	nodeService := services.NewNodeService(db.GormDB)
	executionService := services.NewExecutionService(db.GormDB)
	userService := services.NewUserService(db.GormDB)

	// Initialize engine components for execution manager
	nodeRegistry := engine.NewNodeRegistry()
	executor := engine.NewExecutor()
	runner := engine.NewRunner(executor)

	executionManagerService := services.NewExecutionManagerService(
		workflowService,
		nodeService,
		executionService,
		runner,
		nodeRegistry,
	)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Citadel Agent API",
		ServerHeader: "Citadel-Agent",
	})

	// Setup middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Setup routes
	api.SetupRoutes(app, workflowService, nodeService, executionService, userService, executionManagerService)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Starting Citadel Agent API on port %s", port)
	log.Fatal(app.Listen(":" + port))
}