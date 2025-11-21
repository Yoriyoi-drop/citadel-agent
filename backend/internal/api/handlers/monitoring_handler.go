// backend/internal/api/handlers/monitoring_handler.go
package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"citadel-agent/backend/internal/observability"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// MonitoringHandler handles monitoring and observability endpoints
type MonitoringHandler struct {
	monitoringService *observability.MonitoringService
}

// NewMonitoringHandler creates a new monitoring handler
func NewMonitoringHandler(monitoringService *observability.MonitoringService) *MonitoringHandler {
	return &MonitoringHandler{
		monitoringService: monitoringService,
	}
}

// GetSystemHealth returns system health metrics
func (mh *MonitoringHandler) GetSystemHealth(c *fiber.Ctx) error {
	health, err := mh.monitoringService.GetSystemHealth(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get system health: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    health,
		"timestamp": time.Now().Unix(),
	})
}

// GetWorkflowExecutionTimeline returns execution timeline for a workflow
func (mh *MonitoringHandler) GetWorkflowExecutionTimeline(c *fiber.Ctx) error {
	workflowID := c.Params("workflowId")
	
	// Validate workflow ID
	if _, err := uuid.Parse(workflowID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid workflow ID format",
		})
	}

	events, err := mh.monitoringService.GetWorkflowExecutionTimeline(c.Context(), workflowID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get workflow timeline: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    events,
		"count":   len(events),
		"workflow_id": workflowID,
	})
}

// GetNodeExecutionStats returns statistics for node executions
func (mh *MonitoringHandler) GetNodeExecutionStats(c *fiber.Ctx) error {
	nodeType := c.Params("nodeType")
	workflowID := c.Params("workflowId")
	
	stats, err := mh.monitoringService.GetNodeExecutionStats(c.Context(), nodeType, workflowID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get node stats: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    stats,
		"node_type": nodeType,
		"workflow_id": workflowID,
	})
}

// GetTenantActivity returns activity metrics for a tenant
func (mh *MonitoringHandler) GetTenantActivity(c *fiber.Ctx) error {
	tenantID := c.Params("tenantId")
	daysParam := c.Query("days", "7")
	
	// Validate tenant ID
	if _, err := uuid.Parse(tenantID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid tenant ID format",
		})
	}

	days, err := strconv.Atoi(daysParam)
	if err != nil || days <= 0 {
		days = 7 // Default to 7 days
	}

	activity, err := mh.monitoringService.GetTenantActivity(c.Context(), tenantID, days)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get tenant activity: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    activity,
		"tenant_id": tenantID,
		"days":    days,
	})
}

// GetPerformanceMetrics returns performance metrics for specific components
func (mh *MonitoringHandler) GetPerformanceMetrics(c *fiber.Ctx) error {
	component := c.Params("component")
	period := c.Query("period", "24h")
	
	metrics, err := mh.monitoringService.GetPerformanceMetrics(c.Context(), component, period)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get performance metrics: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    metrics,
		"component": component,
		"period": period,
	})
}

// GetErrorRate returns error rate for a specific period
func (mh *MonitoringHandler) GetErrorRate(c *fiber.Ctx) error {
	service := c.Params("service")
	resource := c.Query("resource", "*")
	hoursParam := c.Query("hours", "24")
	
	hours, err := strconv.Atoi(hoursParam)
	if err != nil || hours <= 0 {
		hours = 24
	}

	rate, err := mh.monitoringService.GetErrorRate(c.Context(), service, resource, hours)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get error rate: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"error_rate": rate,
		"service": service,
		"resource": resource,
		"hours": hours,
	})
}

// GetRecentEvents returns recent system events
func (mh *MonitoringHandler) GetRecentEvents(c *fiber.Ctx) error {
	eventType := c.Query("type", "*")
	limitParam := c.Query("limit", "50")
	offsetParam := c.Query("offset", "0")
	
	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 50
	}
	
	offset, err := strconv.Atoi(offsetParam)
	if err != nil || offset < 0 {
		offset = 0
	}
	
	// In a real implementation, we would query recent events from the monitoring service
	// For now, we'll return an empty list
	
	events := make([]*observability.Event, 0)
	
	return c.JSON(fiber.Map{
		"success": true,
		"data":    events,
		"count":   len(events),
		"limit":   limit,
		"offset":  offset,
		"type":    eventType,
	})
}

// SearchEvents searches for events based on criteria
func (mh *MonitoringHandler) SearchEvents(c *fiber.Ctx) error {
	var req struct {
		Query     string `json:"query"`
		Type      string `json:"type"`
		Service   string `json:"service"`
		TenantID  string `json:"tenant_id"`
		UserID    string `json:"user_id"`
		Resource  string `json:"resource"`
		Action    string `json:"action"`
		Status    string `json:"status"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
		Limit     int    `json:"limit"`
		Offset    int    `json:"offset"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	
	if req.Limit <= 0 || req.Limit > 1000 {
		req.Limit = 50
	}
	
	if req.Offset < 0 {
		req.Offset = 0
	}
	
	// In a real implementation, we would search for events based on criteria
	// For now, we'll return an empty list
	
	events := make([]*observability.Event, 0)
	
	return c.JSON(fiber.Map{
		"success": true,
		"data":    events,
		"count":   len(events),
		"limit":   req.Limit,
		"offset":  req.Offset,
		"query":   req.Query,
	})
}

// GetMetricsDashboard returns aggregated metrics for dashboard
func (mh *MonitoringHandler) GetMetricsDashboard(c *fiber.Ctx) error {
	period := c.Query("period", "24h")
	granularity := c.Query("granularity", "hour")
	
	// Validate period
	allowedPeriods := map[string]bool{
		"1h":  true,
		"6h":  true,
		"12h": true,
		"24h": true,
		"7d":  true,
		"30d": true,
	}
	
	if !allowedPeriods[period] {
		period = "24h"
	}
	
	// Validate granularity
	allowedGranularities := map[string]bool{
		"minute": true,
		"hour":   true,
		"day":    true,
	}
	
	if !allowedGranularities[granularity] {
		granularity = "hour"
	}
	
	// Get dashboard metrics
	metrics := map[string]interface{}{
		"period":      period,
		"granularity": granularity,
		"timestamp":   time.Now().Unix(),
		"summary": map[string]interface{}{
			"total_requests":       0,
			"successful_requests":  0,
			"failed_requests":      0,
			"total_workflows":      0,
			"successful_workflows": 0,
			"failed_workflows":     0,
			"active_tenants":       0,
			"active_users":         0,
			"total_nodes":          0,
			"error_rate":           0.0,
			"avg_response_time":    0.0,
		},
		"trend_data": map[string]interface{}{},
		"top_items": map[string]interface{}{
			"slowest_endpoints": []string{},
			"highest_error_rates": []string{},
			"most_used_workflows": []string{},
		},
		"alerts": map[string]interface{}{
			"critical_count": 0,
			"warning_count":  0,
			"info_count":     0,
		},
	}
	
	return c.JSON(fiber.Map{
		"success": true,
		"data":    metrics,
	})
}

// RegisterRoutes registers monitoring handler routes
func (mh *MonitoringHandler) RegisterRoutes(router fiber.Router) {
	// Health and status endpoints (public)
	router.Get("/health", mh.GetSystemHealth)
	router.Get("/metrics/dashboard", mh.GetMetricsDashboard)
	
	// Authenticated monitoring endpoints
	authRouter := router // In a real implementation, this would be wrapped with auth middleware
	{
		authRouter.Get("/workflow/:workflowId/timeline", mh.GetWorkflowExecutionTimeline)
		authRouter.Get("/node/:nodeType/workflow/:workflowId/stats", mh.GetNodeExecutionStats)
		authRouter.Get("/tenant/:tenantId/activity", mh.GetTenantActivity)
		authRouter.Get("/performance/:component", mh.GetPerformanceMetrics)
		authRouter.Get("/error-rate/:service", mh.GetErrorRate)
		authRouter.Get("/events", mh.GetRecentEvents)
		authRouter.Post("/events/search", mh.SearchEvents)
	}
}