// backend/internal/api/handlers/monitoring_handler.go
package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"citadel-agent/backend/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// MonitoringHandler handles monitoring-related HTTP requests
type MonitoringHandler struct {
	monitoringService *services.MonitoringService
}

// NewMonitoringHandler creates a new monitoring handler
func NewMonitoringHandler(monitoringService *services.MonitoringService) *MonitoringHandler {
	return &MonitoringHandler{
		monitoringService: monitoringService,
	}
}

// GetSystemMetrics returns system-level metrics
func (mh *MonitoringHandler) GetSystemMetrics(c *fiber.Ctx) error {
	metrics, err := mh.monitoringService.GetSystemMetrics(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get system metrics: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"metrics": metrics,
	})
}

// GetMetrics retrieves metrics based on filters
func (mh *MonitoringHandler) GetMetrics(c *fiber.Ctx) error {
	name := c.Query("name")
	service := c.Query("service")
	
	startStr := c.Query("start")
	endStr := c.Query("end")
	
	limitStr := c.Query("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 50
	}

	var start, end *time.Time
	
	if startStr != "" {
		parsed, err := time.Parse(time.RFC3339, startStr)
		if err == nil {
			start = &parsed
		}
	}
	
	if endStr != "" {
		parsed, err := time.Parse(time.RFC3339, endStr)
		if err == nil {
			end = &parsed
		}
	}

	metrics, err := mh.monitoringService.GetMetrics(c.Context(), name, service, start, end, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get metrics: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    metrics,
		"count":   len(metrics),
	})
}

// RecordMetric records a new metric
func (mh *MonitoringHandler) RecordMetric(c *fiber.Ctx) error {
	var req struct {
		Name   string                 `json:"name"`
		Value  float64                `json:"value"`
		Labels map[string]string      `json:"labels"`
		Tags   []string               `json:"tags"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Metric name is required",
		})
	}

	err := mh.monitoringService.RecordMetric(c.Context(), req.Name, req.Value, req.Labels, req.Tags)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to record metric: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Metric recorded successfully",
	})
}

// GetAlerts retrieves alerts based on filters
func (mh *MonitoringHandler) GetAlerts(c *fiber.Ctx) error {
	statusStr := c.Query("status")
	severityStr := c.Query("severity")
	
	limitStr := c.Query("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 50
	}

	var status *services.AlertStatus
	if statusStr != "" {
		s := services.AlertStatus(statusStr)
		status = &s
	}

	var severity *services.AlertSeverity
	if severityStr != "" {
		s := services.AlertSeverity(severityStr)
		severity = &s
	}

	alerts, err := mh.monitoringService.GetAlerts(c.Context(), status, severity, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get alerts: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    alerts,
		"count":   len(alerts),
	})
}

// CreateAlert creates a new alert
func (mh *MonitoringHandler) CreateAlert(c *fiber.Ctx) error {
	var req struct {
		Name        string            `json:"name"`
		Severity    string            `json:"severity"`
		Message     string            `json:"message"`
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
		Condition   string            `json:"condition"`
		Threshold   *float64          `json:"threshold"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" || req.Condition == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Alert name and condition are required",
		})
	}

	alert := &services.Alert{
		Name:        req.Name,
		Severity:    services.AlertSeverity(req.Severity),
		Message:     req.Message,
		Labels:      req.Labels,
		Annotations: req.Annotations,
		Condition:   req.Condition,
		Threshold:   req.Threshold,
	}

	createdAlert, err := mh.monitoringService.CreateAlert(c.Context(), alert)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create alert: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    createdAlert,
	})
}

// ResolveAlert resolves an active alert
func (mh *MonitoringHandler) ResolveAlert(c *fiber.Ctx) error {
	alertID := c.Params("alertId")
	
	// Validate UUID format
	if _, err := uuid.Parse(alertID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid alert ID format",
		})
	}

	err := mh.monitoringService.ResolveAlert(c.Context(), alertID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to resolve alert: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Alert resolved successfully",
	})
}

// GetLogEntries retrieves log entries based on filters
func (mh *MonitoringHandler) GetLogEntries(c *fiber.Ctx) error {
	level := c.Query("level")
	service := c.Query("service")
	
	startStr := c.Query("start")
	endStr := c.Query("end")
	
	limitStr := c.Query("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 50
	}

	var start, end *time.Time
	
	if startStr != "" {
		parsed, err := time.Parse(time.RFC3339, startStr)
		if err == nil {
			start = &parsed
		}
	}
	
	if endStr != "" {
		parsed, err := time.Parse(time.RFC3339, endStr)
		if err == nil {
			end = &parsed
		}
	}

	logEntries, err := mh.monitoringService.GetLogEntries(c.Context(), level, service, start, end, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get log entries: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    logEntries,
		"count":   len(logEntries),
	})
}

// LogMessage records a log entry
func (mh *MonitoringHandler) LogMessage(c *fiber.Ctx) error {
	var req struct {
		Level   string                 `json:"level"`
		Message string                 `json:"message"`
		Service string                 `json:"service"`
		Fields  map[string]interface{} `json:"fields"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Level == "" || req.Message == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Log level and message are required",
		})
	}

	// Add service to fields for consistency
	if req.Fields == nil {
		req.Fields = make(map[string]interface{})
	}
	req.Fields["service"] = req.Service

	err := mh.monitoringService.Log(c.Context(), req.Level, req.Message, req.Fields)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to log message: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Log recorded successfully",
	})
}

// HealthCheck performs a health check of the monitoring system
func (mh *MonitoringHandler) HealthCheck(c *fiber.Ctx) error {
	healthy, err := mh.monitoringService.HealthCheck(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   fmt.Sprintf("Health check failed: %v", err),
		})
	}

	status := "healthy"
	if !healthy {
		status = "unhealthy"
		c.Status(503)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"status":  status,
		"timestamp": time.Now().Unix(),
	})
}

// RegisterRoutes registers monitoring handler routes
func (mh *MonitoringHandler) RegisterRoutes(router fiber.Router) {
	// Public metrics and health endpoints
	router.Get("/health", mh.HealthCheck)
	router.Get("/metrics/system", mh.GetSystemMetrics)
	
	// Authenticated endpoints
	authGroup := router.Use() // This would use the auth middleware in a real implementation
	{
		authGroup.Get("/metrics", mh.GetMetrics)
		authGroup.Post("/metrics", mh.RecordMetric)
		
		authGroup.Get("/alerts", mh.GetAlerts)
		authGroup.Post("/alerts", mh.CreateAlert)
		authGroup.Post("/alerts/:alertId/resolve", mh.ResolveAlert)
		
		authGroup.Get("/logs", mh.GetLogEntries)
		authGroup.Post("/logs", mh.LogMessage)
	}
}