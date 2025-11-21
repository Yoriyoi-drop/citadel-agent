// backend/internal/services/notification_service.go
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeEmail    NotificationType = "email"
	NotificationTypeSlack    NotificationType = "slack"
	NotificationTypeWebhook  NotificationType = "webhook"
	NotificationTypePush     NotificationType = "push"
	NotificationTypeSMS      NotificationType = "sms"
)

// NotificationPriority represents the priority level of notification
type NotificationPriority string

const (
	PriorityLow    NotificationPriority = "low"
	PriorityMedium NotificationPriority = "medium"
	PriorityHigh   NotificationPriority = "high"
	PriorityUrgent NotificationPriority = "urgent"
)

// NotificationChannel represents where the notification is sent
type NotificationChannel string

const (
	ChannelUser     NotificationChannel = "user"
	ChannelTeam     NotificationChannel = "team"
	ChannelWebhook  NotificationChannel = "webhook"
	ChannelEmail    NotificationChannel = "email"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	StatusQueued    NotificationStatus = "queued"
	StatusSending   NotificationStatus = "sending"
	StatusSent      NotificationStatus = "sent"
	StatusFailed    NotificationStatus = "failed"
	StatusDelivered NotificationStatus = "delivered"
)

// Notification represents a notification to be sent
type Notification struct {
	ID          string                 `json:"id"`
	Type        NotificationType       `json:"type"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Channel     NotificationChannel    `json:"channel"`
	Recipient   string                 `json:"recipient"` // user ID, email, webhook URL, etc.
	Priority    NotificationPriority   `json:"priority"`
	Status      NotificationStatus     `json:"status"`
	Payload     map[string]interface{} `json:"payload"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	SentAt      *time.Time             `json:"sent_at,omitempty"`
	Error       *string                `json:"error,omitempty"`
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"max_attempts"`
	Tags        []string               `json:"tags"`
}

// NotificationService handles notification operations
type NotificationService struct {
	db          *pgxpool.Pool
	emailSender EmailSender
	slackSender SlackSender
	webhookSender WebhookSender
	pushSender  PushSender
	smsSender   SMSSender
}

// EmailSender interface for sending emails
type EmailSender interface {
	Send(ctx context.Context, recipient, subject, body string) error
}

// SlackSender interface for sending Slack notifications
type SlackSender interface {
	Send(ctx context.Context, webhookURL, message string) error
}

// WebhookSender interface for sending webhook notifications
type WebhookSender interface {
	Send(ctx context.Context, webhookURL string, payload map[string]interface{}) error
}

// PushSender interface for sending push notifications
type PushSender interface {
	Send(ctx context.Context, recipient, title, message string) error
}

// SMSSender interface for sending SMS
type SMSSender interface {
	Send(ctx context.Context, recipient, message string) error
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	db *pgxpool.Pool,
	emailSender EmailSender,
	slackSender SlackSender,
	webhookSender WebhookSender,
	pushSender PushSender,
	smsSender SMSSender,
) *NotificationService {
	return &NotificationService{
		db:            db,
		emailSender:   emailSender,
		slackSender:   slackSender,
		webhookSender: webhookSender,
		pushSender:    pushSender,
		smsSender:     smsSender,
	}
}

// CreateNotification creates a new notification
func (ns *NotificationService) CreateNotification(ctx context.Context, notification *Notification) (*Notification, error) {
	notification.ID = uuid.New().String()
	notification.Status = StatusQueued
	notification.CreatedAt = time.Now()
	notification.Attempts = 0
	notification.MaxAttempts = 3 // Default max attempts

	// If no scheduled time, set to now
	if notification.ScheduledAt == nil {
		now := time.Now()
		notification.ScheduledAt = &now
	}

	// Validate notification
	if err := ns.validateNotification(notification); err != nil {
		return nil, fmt.Errorf("invalid notification: %w", err)
	}

	// Save to database
	if err := ns.saveNotification(ctx, notification); err != nil {
		return nil, fmt.Errorf("failed to save notification: %w", err)
	}

	// If the notification is scheduled for now, add to send queue
	if notification.ScheduledAt.Before(time.Now()) || notification.ScheduledAt.Equal(time.Now()) {
		go ns.processNotification(notification)
	}

	return notification, nil
}

// SendNotification sends a notification immediately
func (ns *NotificationService) SendNotification(ctx context.Context, notification *Notification) error {
	notification.ID = uuid.New().String()
	notification.Status = StatusQueued
	notification.CreatedAt = time.Now()
	notification.Attempts = 0
	notification.MaxAttempts = 3

	// Validate notification
	if err := ns.validateNotification(notification); err != nil {
		return fmt.Errorf("invalid notification: %w", err)
	}

	// Save to database
	if err := ns.saveNotification(ctx, notification); err != nil {
		return fmt.Errorf("failed to save notification: %w", err)
	}

	// Process the notification
	return ns.processNotification(notification)
}

// processNotification processes a single notification
func (ns *NotificationService) processNotification(notification *Notification) error {
	ctx := context.Background()
	
	// Update status to sending
	notification.Status = StatusSending
	if err := ns.updateNotificationStatus(ctx, notification.ID, StatusSending); err != nil {
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	var sendErr error
	switch notification.Type {
	case NotificationTypeEmail:
		sendErr = ns.sendEmailNotification(ctx, notification)
	case NotificationTypeSlack:
		sendErr = ns.sendSlackNotification(ctx, notification)
	case NotificationTypeWebhook:
		sendErr = ns.sendWebhookNotification(ctx, notification)
	case NotificationTypePush:
		sendErr = ns.sendPushNotification(ctx, notification)
	case NotificationTypeSMS:
		sendErr = ns.sendSMSNotification(ctx, notification)
	default:
		return fmt.Errorf("unsupported notification type: %s", notification.Type)
	}

	// Update status based on result
	if sendErr != nil {
		notification.Attempts++
		notification.Error = &sendErr.Error()
		
		if notification.Attempts >= notification.MaxAttempts {
			notification.Status = StatusFailed
		} else {
			// Re-queue for retry
			notification.Status = StatusQueued
			// Schedule retry after some time
			retryTime := time.Now().Add(time.Duration(notification.Attempts) * time.Minute)
			notification.ScheduledAt = &retryTime
			go func() {
				time.Sleep(time.Duration(notification.Attempts) * time.Minute)
				ns.processNotification(notification)
			}()
		}
	} else {
		notification.Status = StatusSent
		sentAt := time.Now()
		notification.SentAt = &sentAt
	}

	// Update notification in database
	if err := ns.updateNotification(ctx, notification); err != nil {
		return fmt.Errorf("failed to update notification: %w", err)
	}

	return sendErr
}

// sendEmailNotification sends an email notification
func (ns *NotificationService) sendEmailNotification(ctx context.Context, notification *Notification) error {
	if ns.emailSender == nil {
		return fmt.Errorf("email sender not configured")
	}

	return ns.emailSender.Send(ctx, notification.Recipient, notification.Title, notification.Message)
}

// sendSlackNotification sends a Slack notification
func (ns *NotificationService) sendSlackNotification(ctx context.Context, notification *Notification) error {
	if ns.slackSender == nil {
		return fmt.Errorf("slack sender not configured")
	}

	return ns.slackSender.Send(ctx, notification.Recipient, notification.Message)
}

// sendWebhookNotification sends a webhook notification
func (ns *NotificationService) sendWebhookNotification(ctx context.Context, notification *Notification) error {
	if ns.webhookSender == nil {
		return fmt.Errorf("webhook sender not configured")
	}

	payload := notification.Payload
	if payload == nil {
		payload = make(map[string]interface{})
	}
	
	payload["title"] = notification.Title
	payload["message"] = notification.Message
	payload["type"] = notification.Type
	payload["priority"] = notification.Priority

	return ns.webhookSender.Send(ctx, notification.Recipient, payload)
}

// sendPushNotification sends a push notification
func (ns *NotificationService) sendPushNotification(ctx context.Context, notification *Notification) error {
	if ns.pushSender == nil {
		return fmt.Errorf("push sender not configured")
	}

	return ns.pushSender.Send(ctx, notification.Recipient, notification.Title, notification.Message)
}

// sendSMSNotification sends an SMS notification
func (ns *NotificationService) sendSMSNotification(ctx context.Context, notification *Notification) error {
	if ns.smsSender == nil {
		return fmt.Errorf("SMS sender not configured")
	}

	return ns.smsSender.Send(ctx, notification.Recipient, notification.Message)
}

// GetNotification retrieves a notification by ID
func (ns *NotificationService) GetNotification(ctx context.Context, id string) (*Notification, error) {
	query := `
		SELECT id, type, title, message, channel, recipient, priority, status, 
		       payload, metadata, created_at, scheduled_at, sent_at, error, 
		       attempts, max_attempts, tags
		FROM notifications
		WHERE id = $1
	`

	var notification Notification
	var payloadJSON, metadataJSON, tagsJSON []byte
	var scheduledAt, sentAt *time.Time
	var errorStr *string

	err := ns.db.QueryRow(ctx, query, id).Scan(
		&notification.ID,
		&notification.Type,
		&notification.Title,
		&notification.Message,
		&notification.Channel,
		&notification.Recipient,
		&notification.Priority,
		&notification.Status,
		&payloadJSON,
		&metadataJSON,
		&notification.CreatedAt,
		&scheduledAt,
		&sentAt,
		&errorStr,
		&notification.Attempts,
		&notification.MaxAttempts,
		&tagsJSON,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	// Parse JSON fields
	if payloadJSON != nil {
		if err := json.Unmarshal(payloadJSON, &notification.Payload); err != nil {
			return nil, fmt.Errorf("failed to parse payload: %w", err)
		}
	}
	
	if metadataJSON != nil {
		if err := json.Unmarshal(metadataJSON, &notification.Metadata); err != nil {
			return nil, fmt.Errorf("failed to parse metadata: %w", err)
		}
	}
	
	if tagsJSON != nil {
		var tags []string
		if err := json.Unmarshal(tagsJSON, &tags); err != nil {
			return nil, fmt.Errorf("failed to parse tags: %w", err)
		}
		notification.Tags = tags
	}
	
	notification.ScheduledAt = scheduledAt
	notification.SentAt = sentAt
	notification.Error = errorStr

	return &notification, nil
}

// GetNotificationsByUser retrieves notifications for a specific user
func (ns *NotificationService) GetNotificationsByUser(ctx context.Context, userID string, status *NotificationStatus, limit, offset int) ([]*Notification, error) {
	query := `
		SELECT id, type, title, message, channel, recipient, priority, status, 
		       payload, metadata, created_at, scheduled_at, sent_at, error, 
		       attempts, max_attempts, tags
		FROM notifications
		WHERE recipient = $1
	`
	
	args := []interface{}{userID}
	argCount := 2
	
	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, *status)
		argCount++
	}
	
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)
	
	rows, err := ns.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()
	
	var notifications []*Notification
	for rows.Next() {
		var notification Notification
		var payloadJSON, metadataJSON, tagsJSON []byte
		var scheduledAt, sentAt *time.Time
		var errorStr *string

		err := rows.Scan(
			&notification.ID,
			&notification.Type,
			&notification.Title,
			&notification.Message,
			&notification.Channel,
			&notification.Recipient,
			&notification.Priority,
			&notification.Status,
			&payloadJSON,
			&metadataJSON,
			&notification.CreatedAt,
			&scheduledAt,
			&sentAt,
			&errorStr,
			&notification.Attempts,
			&notification.MaxAttempts,
			&tagsJSON,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		// Parse JSON fields
		if payloadJSON != nil {
			if err := json.Unmarshal(payloadJSON, &notification.Payload); err != nil {
				return nil, fmt.Errorf("failed to parse payload: %w", err)
			}
		}
		
		if metadataJSON != nil {
			if err := json.Unmarshal(metadataJSON, &notification.Metadata); err != nil {
				return nil, fmt.Errorf("failed to parse metadata: %w", err)
			}
		}
		
		if tagsJSON != nil {
			var tags []string
			if err := json.Unmarshal(tagsJSON, &tags); err != nil {
				return nil, fmt.Errorf("failed to parse tags: %w", err)
			}
			notification.Tags = tags
		}
		
		notification.ScheduledAt = scheduledAt
		notification.SentAt = sentAt
		notification.Error = errorStr
		
		notifications = append(notifications, &notification)
	}
	
	return notifications, nil
}

// GetNotificationsByStatus retrieves notifications with a specific status
func (ns *NotificationService) GetNotificationsByStatus(ctx context.Context, status NotificationStatus, limit, offset int) ([]*Notification, error) {
	query := `
		SELECT id, type, title, message, channel, recipient, priority, status, 
		       payload, metadata, created_at, scheduled_at, sent_at, error, 
		       attempts, max_attempts, tags
		FROM notifications
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := ns.db.Query(ctx, query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()
	
	var notifications []*Notification
	for rows.Next() {
		var notification Notification
		var payloadJSON, metadataJSON, tagsJSON []byte
		var scheduledAt, sentAt *time.Time
		var errorStr *string

		err := rows.Scan(
			&notification.ID,
			&notification.Type,
			&notification.Title,
			&notification.Message,
			&notification.Channel,
			&notification.Recipient,
			&notification.Priority,
			&notification.Status,
			&payloadJSON,
			&metadataJSON,
			&notification.CreatedAt,
			&scheduledAt,
			&sentAt,
			&errorStr,
			&notification.Attempts,
			&notification.MaxAttempts,
			&tagsJSON,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		// Parse JSON fields
		if payloadJSON != nil {
			if err := json.Unmarshal(payloadJSON, &notification.Payload); err != nil {
				return nil, fmt.Errorf("failed to parse payload: %w", err)
			}
		}
		
		if metadataJSON != nil {
			if err := json.Unmarshal(metadataJSON, &notification.Metadata); err != nil {
				return nil, fmt.Errorf("failed to parse metadata: %w", err)
			}
		}
		
		if tagsJSON != nil {
			var tags []string
			if err := json.Unmarshal(tagsJSON, &tags); err != nil {
				return nil, fmt.Errorf("failed to parse tags: %w", err)
			}
			notification.Tags = tags
		}
		
		notification.ScheduledAt = scheduledAt
		notification.SentAt = sentAt
		notification.Error = errorStr
		
		notifications = append(notifications, &notification)
	}
	
	return notifications, nil
}

// validateNotification validates a notification
func (ns *NotificationService) validateNotification(notification *Notification) error {
	if notification.Title == "" {
		return fmt.Errorf("title is required")
	}
	
	if notification.Message == "" {
		return fmt.Errorf("message is required")
	}
	
	if notification.Recipient == "" {
		return fmt.Errorf("recipient is required")
	}
	
	switch notification.Type {
	case NotificationTypeEmail:
		// Additional email validation could be added
	case NotificationTypeSlack:
		// Additional Slack validation could be added
	case NotificationTypeWebhook:
		// Additional webhook validation could be added
	case NotificationTypePush:
		// Additional push validation could be added
	case NotificationTypeSMS:
		// Additional SMS validation could be added
	default:
		return fmt.Errorf("invalid notification type: %s", notification.Type)
	}
	
	return nil
}

// saveNotification saves a notification to the database
func (ns *NotificationService) saveNotification(ctx context.Context, notification *Notification) error {
	payloadJSON, err := json.Marshal(notification.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	metadataJSON, err := json.Marshal(notification.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	
	tagsJSON, err := json.Marshal(notification.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := `
		INSERT INTO notifications (
			id, type, title, message, channel, recipient, priority, status,
			payload, metadata, created_at, scheduled_at, sent_at, error,
			attempts, max_attempts, tags
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	_, err = ns.db.Exec(ctx, query,
		notification.ID,
		notification.Type,
		notification.Title,
		notification.Message,
		notification.Channel,
		notification.Recipient,
		notification.Priority,
		notification.Status,
		payloadJSON,
		metadataJSON,
		notification.CreatedAt,
		notification.ScheduledAt,
		notification.SentAt,
		notification.Error,
		notification.Attempts,
		notification.MaxAttempts,
		tagsJSON,
	)
	
	if err != nil {
		return fmt.Errorf("failed to insert notification: %w", err)
	}
	
	return nil
}

// updateNotification updates a notification in the database
func (ns *NotificationService) updateNotification(ctx context.Context, notification *Notification) error {
	payloadJSON, err := json.Marshal(notification.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	metadataJSON, err := json.Marshal(notification.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	
	tagsJSON, err := json.Marshal(notification.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := `
		UPDATE notifications
		SET type = $2, title = $3, message = $4, channel = $5, recipient = $6,
		    priority = $7, status = $8, payload = $9, metadata = $10,
		    scheduled_at = $11, sent_at = $12, error = $13, attempts = $14
		WHERE id = $1
	`

	_, err = ns.db.Exec(ctx, query,
		notification.ID,
		notification.Type,
		notification.Title,
		notification.Message,
		notification.Channel,
		notification.Recipient,
		notification.Priority,
		notification.Status,
		payloadJSON,
		metadataJSON,
		notification.ScheduledAt,
		notification.SentAt,
		notification.Error,
		notification.Attempts,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update notification: %w", err)
	}
	
	return nil
}

// updateNotificationStatus updates only the status of a notification
func (ns *NotificationService) updateNotificationStatus(ctx context.Context, id string, status NotificationStatus) error {
	query := "UPDATE notifications SET status = $2 WHERE id = $1"
	
	_, err := ns.db.Exec(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update notification status: %w", err)
	}
	
	return nil
}

// DeleteNotification deletes a notification
func (ns *NotificationService) DeleteNotification(ctx context.Context, id string) error {
	query := "DELETE FROM notifications WHERE id = $1"
	
	_, err := ns.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}
	
	return nil
}