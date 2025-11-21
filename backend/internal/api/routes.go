// backend/internal/api/routes.go
package api

import (
	"citadel-agent/backend/internal/api/handlers"
	"citadel-agent/backend/internal/services"

	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(app *fiber.App, 
	monitoringService *services.MonitoringService,
	tenantService *services.TenantService,
	apiKeyService *services.APIKeyService,
	notificationService *services.NotificationService) {
	
	// Create handlers
	monitoringHandler := handlers.NewMonitoringHandler(monitoringService)
	tenantHandler := handlers.NewTenantHandler(tenantService)
	apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)

	// Base API v1 routes
	v1 := app.Group("/api/v1")

	// Monitoring routes
	monitoringHandler.RegisterRoutes(v1.Group("/monitoring"))

	// Tenant routes
	tenantHandler.RegisterRoutes(v1.Group("/tenants"))

	// API Key routes
	apiKeyHandler.RegisterRoutes(v1.Group("/api-keys"))

	// Notification routes
	notificationHandler.RegisterRoutes(v1.Group("/notifications"))

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "healthy",
			"service":   "citadel-agent",
			"timestamp": c.Context().Time().Unix(),
		})
	})
}