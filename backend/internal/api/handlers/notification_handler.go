// backend/internal/api/handlers/notification_handler.go
package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"citadel-agent/backend/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// NotificationHandler handles notification-related HTTP requests
type NotificationHandler struct {
	notificationService *services.NotificationService
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(notificationService *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// CreateNotification creates a new notification
func (nh *NotificationHandler) CreateNotification(c *fiber.Ctx) error {
	var req struct {
		Type        string                 `json:"type"`
		Title       string                 `json:"title"`
		Message     string                 `json:"message"`
		Channel     string                 `json:"channel"`
		Recipient   string                 `json:"recipient"`
		Priority    string                 `json:"priority"`
		Payload     map[string]interface{} `json:"payload"`
		Metadata    map[string]interface{} `json:"metadata"`
		ScheduledAt *string                `json:"scheduled_at,omitempty"` // RFC3339 format
		Tags        []string               `json:"tags"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Title == "" || req.Message == "" || req.Recipient == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Title, message, and recipient are required",
		})
	}

	notification := &services.Notification{
		Type:      services.NotificationType(req.Type),
		Title:     req.Title,
		Message:   req.Message,
		Channel:   services.NotificationChannel(req.Channel),
		Recipient: req.Recipient,
		Priority:  services.NotificationPriority(req.Priority),
		Payload:   req.Payload,
		Metadata:  req.Metadata,
		Tags:      req.Tags,
	}

	// Set default values if not provided
	if notification.Type == "" {
		notification.Type = services.NotificationTypeEmail
	}
	if notification.Priority == "" {
		notification.Priority = services.PriorityMedium
	}
	if notification.Channel == "" {
		notification.Channel = services.ChannelUser
	}

	// Parse scheduled time if provided
	if req.ScheduledAt != nil {
		// In a real implementation, you would parse the RFC3339 time
		// For now, we'll just set a flag to indicate it's scheduled
	}

	// Send the notification
	err := nh.notificationService.SendNotification(c.Context(), notification)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to send notification: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Notification sent successfully",
		"data":    notification,
	})
}

// CreateScheduledNotification creates a new scheduled notification
func (nh *NotificationHandler) CreateScheduledNotification(c *fiber.Ctx) error {
	var req struct {
		Type        string                 `json:"type"`
		Title       string                 `json:"title"`
		Message     string                 `json:"message"`
		Channel     string                 `json:"channel"`
		Recipient   string                 `json:"recipient"`
		Priority    string                 `json:"priority"`
		Payload     map[string]interface{} `json:"payload"`
		Metadata    map[string]interface{} `json:"metadata"`
		ScheduledAt string                 `json:"scheduled_at"` // RFC3339 format
		Tags        []string               `json:"tags"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Title == "" || req.Message == "" || req.Recipient == "" || req.ScheduledAt == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Title, message, recipient, and scheduled_at are required",
		})
	}

	notification := &services.Notification{
		Type:      services.NotificationType(req.Type),
		Title:     req.Title,
		Message:   req.Message,
		Channel:   services.NotificationChannel(req.Channel),
		Recipient: req.Recipient,
		Priority:  services.NotificationPriority(req.Priority),
		Payload:   req.Payload,
		Metadata:  req.Metadata,
		Tags:      req.Tags,
	}

	// Set default values if not provided
	if notification.Type == "" {
		notification.Type = services.NotificationTypeEmail
	}
	if notification.Priority == "" {
		notification.Priority = services.PriorityMedium
	}
	if notification.Channel == "" {
		notification.Channel = services.ChannelUser
	}

	// Create the scheduled notification
	createdNotification, err := nh.notificationService.CreateNotification(c.Context(), notification)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create scheduled notification: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Notification scheduled successfully",
		"data":    createdNotification,
	})
}

// GetNotification retrieves a notification by ID
func (nh *NotificationHandler) GetNotification(c *fiber.Ctx) error {
	notificationID := c.Params("notificationId")

	// Validate UUID format
	if _, err := uuid.Parse(notificationID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid notification ID format",
		})
	}

	notification, err := nh.notificationService.GetNotification(c.Context(), notificationID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get notification: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    notification,
	})
}

// GetNotificationsByUser retrieves notifications for a specific user
func (nh *NotificationHandler) GetNotificationsByUser(c *fiber.Ctx) error {
	userID := c.Params("userId")
	statusStr := c.Query("status")
	limitStr := c.Query("limit", "50")
	offsetStr := c.Query("offset", "0")

	// Validate UUID format
	if _, err := uuid.Parse(userID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	var status *services.NotificationStatus
	if statusStr != "" {
		s := services.NotificationStatus(statusStr)
		status = &s
	}

	notifications, err := nh.notificationService.GetNotificationsByUser(c.Context(), userID, status, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get notifications: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    notifications,
		"count":   len(notifications),
		"limit":   limit,
		"offset":  offset,
	})
}

// GetNotificationsByStatus retrieves notifications with a specific status
func (nh *NotificationHandler) GetNotificationsByStatus(c *fiber.Ctx) error {
	statusStr := c.Params("status")
	limitStr := c.Query("limit", "50")
	offsetStr := c.Query("offset", "0")

	status := services.NotificationStatus(statusStr)
	if status == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Status is required",
		})
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	notifications, err := nh.notificationService.GetNotificationsByStatus(c.Context(), status, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get notifications: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    notifications,
		"count":   len(notifications),
		"limit":   limit,
		"offset":  offset,
	})
}

// DeleteNotification deletes a notification
func (nh *NotificationHandler) DeleteNotification(c *fiber.Ctx) error {
	notificationID := c.Params("notificationId")

	// Validate UUID format
	if _, err := uuid.Parse(notificationID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid notification ID format",
		})
	}

	err := nh.notificationService.DeleteNotification(c.Context(), notificationID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to delete notification: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Notification deleted successfully",
	})
}

// RegisterRoutes registers notification handler routes
func (nh *NotificationHandler) RegisterRoutes(router fiber.Router) {
	// Authenticated endpoints
	authGroup := router.Use() // This would use the auth middleware in a real implementation
	{
		// Notification management
		authGroup.Post("/notifications", nh.CreateNotification)
		authGroup.Post("/notifications/scheduled", nh.CreateScheduledNotification)
		authGroup.Get("/notifications/:notificationId", nh.GetNotification)
		authGroup.Delete("/notifications/:notificationId", nh.DeleteNotification)
		
		// Specific queries
		authGroup.Get("/users/:userId/notifications", nh.GetNotificationsByUser)
		authGroup.Get("/notifications/status/:status", nh.GetNotificationsByStatus)
	}
}