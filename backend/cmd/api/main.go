package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/citadel-agent/backend/internal/config"
	"github.com/citadel-agent/backend/internal/database"
	"github.com/citadel-agent/backend/internal/auth"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// App represents the main application
type App struct {
	config     *config.Config
	db         *pgxpool.Pool
	authService *auth.AuthService
	httpServer *http.Server
}

// NewApp creates a new application instance
func NewApp() *App {
	cfg := config.LoadConfig()
	
	// Initialize database
	db, err := database.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize auth service
	authService := auth.NewAuthService(db)
	
	return &App{
		config:     cfg,
		db:         db,
		authService: authService,
	}
}

// SetupRoutes configures the HTTP routes
func (a *App) SetupRoutes(app *fiber.App) {
	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"timestamp": time.Now().Unix(),
		})
	})

	// Register auth routes
	a.authService.RegisterFiberRoutes(app)

	// Register other routes...
}

// Start starts the HTTP server
func (a *App) Start() error {
	// Create Fiber app
	fiberApp := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Printf("Error: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Internal server error",
			})
		},
	})

	// Setup routes
	a.SetupRoutes(fiberApp)

	// Create HTTP server
	a.httpServer = &http.Server{
		Addr:    ":" + a.config.ServerPort,
		Handler: fiberApp,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", a.config.ServerPort)
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown the server
	if err := a.httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return err
	}

	// Close database connection
	if a.db != nil {
		a.db.Close()
	}

	log.Println("Server exited")
	return nil
}

// Run starts the application
func (a *App) Run() {
	if err := a.Start(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}
}

func main() {
	app := NewApp()
	app.Run()
}