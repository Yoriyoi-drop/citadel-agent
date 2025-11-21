// backend/internal/app/app.go
package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"citadel-agent/backend/config"
	"citadel-agent/backend/internal/ai"
	"citadel-agent/backend/internal/api"
	"citadel-agent/backend/internal/repositories"
	"citadel-agent/backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
)

// App represents the main application
type App struct {
	config           *config.EngineConfig
	db               *pgxpool.Pool
	server           *fiber.App
	monitoringSvc    *services.MonitoringService
	tenantSvc        *services.TenantService
	apiKeySvc        *services.APIKeyService
	notificationSvc  *services.NotificationService
	aiRuntimeSvc     *ai.AIRuntimeManager
}

// NewApp creates a new application instance
func NewApp(cfg *config.EngineConfig) (*App, error) {
	app := &App{
		config: cfg,
	}

	// Initialize database connection
	if err := app.initDatabase(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize services
	if err := app.initServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	// Initialize HTTP server
	if err := app.initServer(); err != nil {
		return nil, fmt.Errorf("failed to initialize server: %w", err)
	}

	return app, nil
}

// initDatabase initializes the database connection
func (a *App) initDatabase() error {
	var err error
	a.db, err = pgxpool.New(context.Background(), fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		a.config.DatabaseConfig.Username,
		a.config.DatabaseConfig.Password,
		a.config.DatabaseConfig.Host,
		a.config.DatabaseConfig.Port,
		a.config.DatabaseConfig.Name,
		a.config.DatabaseConfig.SSLMode,
	))
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.db.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established successfully")
	return nil
}

// initServices initializes all application services
func (a *App) initServices() error {
	// Initialize repositories
	tenantRepo := repositories.NewTenantRepository(a.db)
	userRepo := repositories.NewUserRepository(a.db) // Assuming this exists
	teamRepo := repositories.NewTeamRepository(a.db) // Assuming this exists
	apiKeyRepo := repositories.NewAPIKeyRepository(a.db)
	monitoringRepo := repositories.NewMonitoringRepository(a.db) // New monitoring repo

	// Initialize services
	a.monitoringSvc = services.NewMonitoringService(a.db, &loggerAdapter{})
	a.tenantSvc = services.NewTenantService(a.db, tenantRepo, userRepo, teamRepo, nil)
	a.apiKeySvc = services.NewAPIKeyService(a.db, apiKeyRepo, userRepo, teamRepo)
	a.notificationSvc = services.NewNotificationService(
		a.db,
		&emailSender{},   // Implementasi email sender
		&slackSender{},   // Implementasi slack sender
		&webhookSender{}, // Implementasi webhook sender
		&pushSender{},    // Implementasi push sender
		&smsSender{},     // Implementasi sms sender
	)
	a.aiRuntimeSvc = ai.NewAIRuntimeManager(a.db)

	log.Println("All services initialized successfully")
	return nil
}

// initServer initializes the HTTP server
func (a *App) initServer() error {
	// Create Fiber app
	a.server = fiber.New(fiber.Config{
		ReadTimeout:      a.config.APIConfig.ReadTimeout,
		WriteTimeout:     a.config.APIConfig.WriteTimeout,
		IdleTimeout:      a.config.APIConfig.IdleTimeout,
		MaxRequestBodySize: int(a.config.APIConfig.MaxRequestBodySize),
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Log error
			log.Printf("Error: %v", err)
			
			// Return JSON error response
			return c.Status(500).JSON(fiber.Map{
				"error": "Internal server error",
			})
		},
	})

	// Register middleware
	a.registerMiddleware()

	// Register routes
	api.RegisterRoutes(a.server, a.monitoringSvc, a.tenantSvc, a.apiKeySvc, a.notificationSvc)

	log.Println("HTTP server initialized successfully")
	return nil
}

// registerMiddleware registers application middleware
func (a *App) registerMiddleware() {
	// Recovery middleware
	a.server.Use(recover.New())

	// Logger middleware
	if a.config.LoggingConfig.Level == "debug" {
		a.server.Use(logger.New())
	}

	// CORS middleware
	if a.config.APIConfig.EnableCORS {
		corsConfig := cors.Config{
			AllowOrigins:     fmt.Sprintf("%s", a.config.APIConfig.CORSOrigins),
			AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
			AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID",
			ExposeHeaders:    "X-Request-ID",
			AllowCredentials: true,
		}
		a.server.Use(cors.New(corsConfig))
	}

	// Rate limiting will be implemented per route as needed
}

// Start starts the application
func (a *App) Start() error {
	log.Printf("Starting Citadel Agent server on %s:%d", a.config.APIConfig.Host, a.config.APIConfig.Port)

	// Start the server
	return a.server.Listen(fmt.Sprintf("%s:%d", a.config.APIConfig.Host, a.config.APIConfig.Port))
}

// Stop stops the application
func (a *App) Stop() error {
	log.Println("Shutting down Citadel Agent server...")

	// Close database connection
	if a.db != nil {
		a.db.Close()
	}

	// Close server gracefully
	if a.server != nil {
		if err := a.server.Shutdown(); err != nil {
			log.Printf("Error shutting down server: %v", err)
			return err
		}
	}

	log.Println("Citadel Agent server stopped successfully")
	return nil
}

// loggerAdapter adapts the application logger to the services interface
type loggerAdapter struct{}

func (l *loggerAdapter) Debug(msg string, args ...interface{}) {
	log.Printf("[DEBUG] "+msg, args...)
}

func (l *loggerAdapter) Info(msg string, args ...interface{}) {
	log.Printf("[INFO] "+msg, args...)
}

func (l *loggerAdapter) Warn(msg string, args ...interface{}) {
	log.Printf("[WARN] "+msg, args...)
}

func (l *loggerAdapter) Error(msg string, args ...interface{}) {
	log.Printf("[ERROR] "+msg, args...)
}

// emailSender implements the email sender interface
type emailSender struct{}

func (e *emailSender) Send(ctx context.Context, recipient, subject, body string) error {
	// Implementation would go here
	// For now, we'll just log it
	log.Printf("Email sent to %s: %s", recipient, subject)
	return nil
}

// slackSender implements the slack sender interface
type slackSender struct{}

func (s *slackSender) Send(ctx context.Context, webhookURL, message string) error {
	// Implementation would go here
	// For now, we'll just log it
	log.Printf("Slack message sent to %s", webhookURL)
	return nil
}

// webhookSender implements the webhook sender interface
type webhookSender struct{}

func (w *webhookSender) Send(ctx context.Context, webhookURL string, payload map[string]interface{}) error {
	// Implementation would go here
	// For now, we'll just log it
	log.Printf("Webhook sent to %s", webhookURL)
	return nil
}

// pushSender implements the push sender interface
type pushSender struct{}

func (p *pushSender) Send(ctx context.Context, recipient, title, message string) error {
	// Implementation would go here
	// For now, we'll just log it
	log.Printf("Push notification sent to %s", recipient)
	return nil
}

// smsSender implements the SMS sender interface
type smsSender struct{}

func (s *smsSender) Send(ctx context.Context, recipient, message string) error {
	// Implementation would go here
	// For now, we'll just log it
	log.Printf("SMS sent to %s", recipient)
	return nil
}