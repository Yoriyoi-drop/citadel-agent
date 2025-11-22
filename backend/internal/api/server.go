package api

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"github.com/citadel-agent/backend/internal/temporal"
	"github.com/citadel-agent/backend/internal/workflow/core/engine"
	"github.com/citadel-agent/backend/internal/plugins"
)

var serverStartTime time.Time

// Server represents the API server
type Server struct {
	app             *fiber.App
	temporalService *temporal.TemporalWorkflowService
	pluginManager   *plugins.NodeManager
	engine          *engine.Engine
	config          *Config
}

// Config holds the configuration for the API server
type Config struct {
	Port             string
	Host             string
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	IdleTimeout      time.Duration
	ShutdownTimeout  time.Duration
	EnableCORS       bool
	EnableLogger     bool
	EnableRecover    bool
	BasePath         string
	DebugMode        bool
	RequestIDHeader  string
}

// NewServer creates a new API server
func NewServer(temporalService *temporal.TemporalWorkflowService, pluginManager *plugins.NodeManager, baseEngine *engine.Engine, config *Config) *Server {
	// Set default config if not provided
	if config == nil {
		config = GetDefaultConfig()
	}

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:         "Citadel Agent API Server",
		ReadTimeout:     config.ReadTimeout,
		WriteTimeout:    config.WriteTimeout,
		IdleTimeout:     config.IdleTimeout,
		DisableStartupMessage: false,
		EnablePrintRoutes:     config.DebugMode,
	})

	// Set up middleware
	serverStartTime = time.Now()

	// Global middleware to set server start time
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("server_start_time", serverStartTime)
		c.Locals("request_start_time", time.Now())
		return c.Next()
	})

	if config.EnableLogger {
		app.Use(logger.New())
	}

	if config.EnableRecover {
		app.Use(recover.New())
	}

	app.Use(requestid.New(requestid.Config{
		Header: config.RequestIDHeader,
	}))

	if config.EnableCORS {
		app.Use(cors.New(cors.Config{
			AllowOrigins: getEnv("CORS_ORIGINS", "*"),
			AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
			AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Request-ID",
		}))
	}

	server := &Server{
		app:             app,
		temporalService: temporalService,
		pluginManager:   pluginManager,
		engine:          baseEngine,
		config:          config,
	}

	// Set up routes
	server.setupRoutes()

	return server
}

// GetDefaultConfig returns a configuration with default values
func GetDefaultConfig() *Config {
	return &Config{
		Port:             getEnv("PORT", "3000"),
		Host:             getEnv("HOST", "0.0.0.0"),
		ReadTimeout:      30 * time.Second,
		WriteTimeout:     30 * time.Second,
		IdleTimeout:      60 * time.Second,
		ShutdownTimeout:  10 * time.Second,
		EnableCORS:       true,
		EnableLogger:     true,
		EnableRecover:    true,
		BasePath:         "/api/v1",
		DebugMode:        getEnv("DEBUG", "false") == "true",
		RequestIDHeader:  "X-Request-ID",
	}
}

// setupRoutes sets up all API routes
func (s *Server) setupRoutes() {
	// Health check endpoint
	s.app.Get("/health", s.healthCheck)

	// API version group
	api := s.app.Group(s.config.BasePath)

	// Workflow endpoints
	workflow := api.Group("/workflows")
	workflow.Post("/", s.createWorkflow)
	workflow.Get("/", s.listWorkflows)
	workflow.Get("/:id", s.getWorkflow)
	workflow.Post("/:id/execute", s.executeWorkflow)
	workflow.Get("/:id/status", s.getWorkflowStatus)
	workflow.Post("/:id/cancel", s.cancelWorkflow)
	workflow.Post("/:id/terminate", s.terminateWorkflow)

	// Node endpoints
	nodes := api.Group("/nodes")
	nodes.Get("/", s.listNodeTypes)
	nodes.Post("/register", s.registerNodeType)
	nodes.Get("/plugins", s.listPluginNodes)

	// Plugin endpoints
	plugins := api.Group("/plugins")
	plugins.Get("/", s.listPlugins)
	plugins.Post("/register", s.registerPlugin)
	plugins.Get("/:id", s.getPlugin)
	plugins.Delete("/:id", s.unregisterPlugin)
	plugins.Post("/:id/execute", s.executePlugin)

	// Engine endpoints
	engineGroup := api.Group("/engine")
	engineGroup.Get("/status", s.getEngineStatus)
	engineGroup.Get("/stats", s.getEngineStats)
}

// Start starts the API server
func (s *Server) Start() error {
	address := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)
	log.Printf("Starting Citadel Agent API server on %s", address)
	
	return s.app.Listen(address)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}

// Helper function to get environment variables with default values
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// healthCheck returns the health status of the server
func (s *Server) healthCheck(c *fiber.Ctx) error {
	requestStartTime, ok := c.Locals("request_start_time").(time.Time)
	if !ok {
		requestStartTime = time.Now()
	}

	return c.JSON(fiber.Map{
		"status":  "healthy",
		"message": "Citadel Agent API server is running",
		"timestamp": time.Now().Unix(),
		"uptime": time.Since(serverStartTime).String(),
		"request_processing_time": time.Since(requestStartTime).String(),
	})
}