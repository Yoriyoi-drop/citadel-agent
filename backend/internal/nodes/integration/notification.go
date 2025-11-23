package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
)

// NotificationChannel represents the type of notification channel
type NotificationChannel string

const (
	EmailChannel    NotificationChannel = "email"
	SMSChannel      NotificationChannel = "sms"
	SlackChannel    NotificationChannel = "slack"
	WebhookChannel  NotificationChannel = "webhook"
	DiscordChannel  NotificationChannel = "discord"
	TelegramChannel NotificationChannel = "telegram"
)

// NotificationPriority represents the priority of the notification
type NotificationPriority string

const (
	PriorityLow    NotificationPriority = "low"
	PriorityNormal NotificationPriority = "normal"
	PriorityHigh   NotificationPriority = "high"
	PriorityUrgent NotificationPriority = "urgent"
)

// NotificationConfig represents the configuration for a notification node
type NotificationConfig struct {
	Channel          NotificationChannel    `json:"channel"`
	Recipients       []string               `json:"recipients"`
	Title            string                 `json:"title"`
	Message          string                 `json:"message"`
	Priority         NotificationPriority   `json:"priority"`
	Template         string                 `json:"template"`
	ChannelConfig    map[string]interface{} `json:"channel_config"`
	Sender           string                 `json:"sender"`
	Attachments      []string               `json:"attachments"` // URLs or file paths
	EnableCaching    bool                   `json:"enable_caching"`
	CacheTTL         int                    `json:"cache_ttl"` // in seconds
	EnableProfiling  bool                   `json:"enable_profiling"`
	ReturnRawResults bool                   `json:"return_raw_results"`
	CustomParams     map[string]interface{} `json:"custom_params"`
	Timeout          int                    `json:"timeout"` // in seconds
}

// NotificationNode represents a notification sending node
type NotificationNode struct {
	config *NotificationConfig
	client *http.Client
}

// NewNotificationNode creates a new notification node
func NewNotificationNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	// Convert config map to struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var notifConfig NotificationConfig
	if err := json.Unmarshal(jsonData, &notifConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate and set defaults
	if notifConfig.Channel == "" {
		notifConfig.Channel = EmailChannel
	}

	if notifConfig.Priority == "" {
		notifConfig.Priority = PriorityNormal
	}

	if notifConfig.Timeout == 0 {
		notifConfig.Timeout = 30 // 30 seconds default
	}

	if notifConfig.CacheTTL == 0 {
		notifConfig.CacheTTL = 3600 // 1 hour default cache TTL
	}

	if notifConfig.ChannelConfig == nil {
		notifConfig.ChannelConfig = make(map[string]interface{})
	}

	// Initialize HTTP client
	client := &http.Client{
		Timeout: time.Duration(notifConfig.Timeout) * time.Second,
	}

	return &NotificationNode{
		config: &notifConfig,
		client: client,
	}, nil
}

// Execute executes the notification sending operation
func (nn *NotificationNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	startTime := time.Now()

	// Override config values with inputs if provided
	channel := nn.config.Channel
	if inputChannel, exists := inputs["channel"]; exists {
		if chnl, ok := inputChannel.(string); ok && chnl != "" {
			switch strings.ToLower(chnl) {
			case "email":
				channel = EmailChannel
			case "sms":
				channel = SMSChannel
			case "slack":
				channel = SlackChannel
			case "webhook":
				channel = WebhookChannel
			case "discord":
				channel = DiscordChannel
			case "telegram":
				channel = TelegramChannel
			}
		}
	}

	recipients := nn.config.Recipients
	if inputRecipients, exists := inputs["recipients"]; exists {
		if recSlice, ok := inputRecipients.([]interface{}); ok {
			recipients = make([]string, len(recSlice))
			for i, r := range recSlice {
				if rStr, ok := r.(string); ok {
					recipients[i] = rStr
				} else {
					recipients[i] = fmt.Sprintf("%v", r)
				}
			}
		} else if recStr, ok := inputRecipients.(string); ok {
			recipients = []string{recStr}
		}
	}

	title := nn.config.Title
	if inputTitle, exists := inputs["title"]; exists {
		if t, ok := inputTitle.(string); ok {
			title = t
		}
	}

	message := nn.config.Message
	if inputMessage, exists := inputs["message"]; exists {
		if msg, ok := inputMessage.(string); ok {
			message = msg
		}
	}

	priority := nn.config.Priority
	if inputPriority, exists := inputs["priority"]; exists {
		if pr, ok := inputPriority.(string); ok && pr != "" {
			switch strings.ToLower(pr) {
			case "low":
				priority = PriorityLow
			case "normal":
				priority = PriorityNormal
			case "high":
				priority = PriorityHigh
			case "urgent":
				priority = PriorityUrgent
			}
		}
	}

	template := nn.config.Template
	if inputTemplate, exists := inputs["template"]; exists {
		if tpl, ok := inputTemplate.(string); ok {
			template = tpl
		}
	}

	channelConfig := nn.config.ChannelConfig
	if inputConfig, exists := inputs["channel_config"]; exists {
		if configMap, ok := inputConfig.(map[string]interface{}); ok {
			channelConfig = make(map[string]interface{})
			for k, v := range configMap {
				channelConfig[k] = v
			}
		}
	}

	sender := nn.config.Sender
	if inputSender, exists := inputs["sender"]; exists {
		if snd, ok := inputSender.(string); ok {
			sender = snd
		}
	}
	_ = sender // Use sender variable (will be used in actual implementation)

	attachments := nn.config.Attachments
	if inputAttachments, exists := inputs["attachments"]; exists {
		if attSlice, ok := inputAttachments.([]interface{}); ok {
			attachments = make([]string, len(attSlice))
			for i, att := range attSlice {
				if attStr, ok := att.(string); ok {
					attachments[i] = attStr
				} else {
					attachments[i] = fmt.Sprintf("%v", att)
				}
			}
		}
	}

	enableProfiling := nn.config.EnableProfiling
	if inputEnableProfiling, exists := inputs["enable_profiling"]; exists {
		if prof, ok := inputEnableProfiling.(bool); ok {
			enableProfiling = prof
		}
	}

	returnRawResults := nn.config.ReturnRawResults
	if inputReturnRaw, exists := inputs["return_raw_results"]; exists {
		if raw, ok := inputReturnRaw.(bool); ok {
			returnRawResults = raw
		}
	}

	// Prepare message content
	messageContent := message
	if template != "" {
		// Apply template if provided
		messageContent = nn.applyTemplate(template, inputs)
	}

	// Send notification based on channel
	var result map[string]interface{}
	var err error

	switch channel {
	case EmailChannel:
		result, err = nn.sendEmail(recipients, title, messageContent, channelConfig)
	case SMSChannel:
		result, err = nn.sendSMS(recipients, messageContent, channelConfig)
	case SlackChannel:
		result, err = nn.sendSlackMessage(recipients, title, messageContent, channelConfig)
	case DiscordChannel:
		result, err = nn.sendDiscordMessage(recipients, title, messageContent, channelConfig)
	case TelegramChannel:
		result, err = nn.sendTelegramMessage(recipients, title, messageContent, channelConfig)
	case WebhookChannel:
		result, err = nn.sendWebhook(recipients, title, messageContent, channelConfig)
	default:
		return nil, fmt.Errorf("unsupported notification channel: %s", channel)
	}

	if err != nil {
		return map[string]interface{}{
			"success":   false,
			"error":     err.Error(),
			"channel":   string(channel),
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Prepare final result
	finalResult := map[string]interface{}{
		"success":           true,
		"notification_sent": true,
		"channel":           string(channel),
		"recipients":        recipients,
		"title":             title,
		"message_length":    len(messageContent),
		"priority":          string(priority),
		"result":            result,
		"timestamp":         time.Now().Unix(),
		"execution_time":    time.Since(startTime).Seconds(),
		"input_data":        inputs,
	}

	if returnRawResults {
		finalResult["raw_message"] = messageContent
		finalResult["raw_recipients"] = recipients
		finalResult["raw_channel_config"] = channelConfig
	}

	// Add profiling data if enabled
	if enableProfiling {
		finalResult["profiling"] = map[string]interface{}{
			"start_time": startTime.Unix(),
			"end_time":   time.Now().Unix(),
			"duration":   time.Since(startTime).Seconds(),
			"channel":    string(channel),
			"recipients": len(recipients),
			"priority":   string(priority),
		}
	}

	return finalResult, nil
}

// sendEmail sends an email notification
func (nn *NotificationNode) sendEmail(recipients []string, title, message string, config map[string]interface{}) (map[string]interface{}, error) {
	// In a real implementation, this would use a mail service like SMTP
	// For this example, we'll simulate the call

	// Get SMTP configuration
	smtpHost, _ := config["smtp_host"].(string)
	smtpPort, _ := config["smtp_port"].(float64)
	// smtpUsername, _ := config["smtp_username"].(string)  // Unused variable
	// smtpPassword, _ := config["smtp_password"].(string)  // Unused variable
	fromEmail, _ := config["from_email"].(string)

	// Simulate email sending
	sendTime := time.Now()

	// Create fake response for simulation
	response := map[string]interface{}{
		"status":          "sent",
		"recipients":      recipients,
		"title":           title,
		"message_preview": truncateString(message, 100),
		"timestamp":       sendTime.Unix(),
		"smtp_host":       smtpHost,
		"smtp_port":       int(smtpPort),
		"from_email":      fromEmail,
		"provider":        "email",
		"mock_send":       true, // Indicates this is a simulated send
	}

	// In a real implementation, we would:
	// 1. Establish SMTP connection
	// 2. Authenticate
	// 3. Send email
	// 4. Return actual status

	return response, nil
}

// sendSMS sends an SMS notification
func (nn *NotificationNode) sendSMS(recipients []string, message string, config map[string]interface{}) (map[string]interface{}, error) {
	// In a real implementation, this would use an SMS service like Twilio
	// For this example, we'll simulate the call

	// Get SMS provider configuration
	provider, _ := config["provider"].(string)
	// accountSid, _ := config["account_sid"].(string)  // Unused variable
	// authToken, _ := config["auth_token"].(string)    // Unused variable
	fromNumber, _ := config["from_number"].(string)

	// Simulate SMS sending
	sendTime := time.Now()

	// Create fake response for simulation
	response := map[string]interface{}{
		"status":      "sent",
		"recipients":  recipients,
		"message":     truncateString(message, 160), // SMS limit
		"timestamp":   sendTime.Unix(),
		"provider":    provider,
		"from_number": fromNumber,
		"mock_send":   true, // Indicates this is a simulated send
	}

	// In a real implementation, we would:
	// 1. Call SMS provider API
	// 2. Handle authentication
	// 3. Send SMS messages
	// 4. Return actual status

	return response, nil
}

// sendSlackMessage sends a Slack notification
func (nn *NotificationNode) sendSlackMessage(recipients []string, title, message string, config map[string]interface{}) (map[string]interface{}, error) {
	webhookURL, exists := config["webhook_url"].(string)
	if !exists || webhookURL == "" {
		return nil, fmt.Errorf("Slack webhook URL is required")
	}

	// Prepare the Slack message payload
	slackMsg := map[string]interface{}{
		"text": title,
		"blocks": []map[string]interface{}{
			{
				"type": "section",
				"text": map[string]interface{}{
					"type": "mrkdwn",
					"text": fmt.Sprintf("*%s*\n%s", title, message),
				},
			},
		},
	}

	// Add recipient if provided as channel
	if len(recipients) > 0 {
		slackMsg["channel"] = recipients[0] // Use first recipient as channel
	}

	// Marshal the payload
	jsonData, err := json.Marshal(slackMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Slack message: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create Slack request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := nn.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send Slack message: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody := make([]byte, 0)
	if resp.Body != nil {
		respBody, _ = io.ReadAll(resp.Body)
	}

	// Prepare response
	response := map[string]interface{}{
		"status":       resp.Status,
		"status_code":  resp.StatusCode,
		"sent_to":      recipients,
		"title":        title,
		"message":      truncateString(message, 200),
		"timestamp":    time.Now().Unix(),
		"webhook_used": webhookURL,
		"provider":     "slack",
	}

	// Add response body if available
	if len(respBody) > 0 {
		response["raw_response"] = string(respBody)
	}

	return response, nil
}

// sendDiscordMessage sends a Discord notification
func (nn *NotificationNode) sendDiscordMessage(recipients []string, title, message string, config map[string]interface{}) (map[string]interface{}, error) {
	webhookURL, exists := config["webhook_url"].(string)
	if !exists || webhookURL == "" {
		return nil, fmt.Errorf("Discord webhook URL is required")
	}

	// Prepare the Discord message payload
	discordMsg := map[string]interface{}{
		"content": fmt.Sprintf("**%s**\n%s", title, message),
	}

	// Marshal the payload
	jsonData, err := json.Marshal(discordMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Discord message: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := nn.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send Discord message: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody := make([]byte, 0)
	if resp.Body != nil {
		respBody, _ = io.ReadAll(resp.Body)
	}

	// Prepare response
	response := map[string]interface{}{
		"status":       resp.Status,
		"status_code":  resp.StatusCode,
		"sent_to":      recipients,
		"title":        title,
		"message":      truncateString(message, 200),
		"timestamp":    time.Now().Unix(),
		"webhook_used": webhookURL,
		"provider":     "discord",
	}

	// Add response body if available
	if len(respBody) > 0 {
		response["raw_response"] = string(respBody)
	}

	return response, nil
}

// sendTelegramMessage sends a Telegram notification
func (nn *NotificationNode) sendTelegramMessage(recipients []string, title, message string, config map[string]interface{}) (map[string]interface{}, error) {
	botToken, exists := config["bot_token"].(string)
	if !exists || botToken == "" {
		return nil, fmt.Errorf("Telegram bot token is required")
	}

	webhookURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	// Prepare the Telegram message payload
	telegramMsg := map[string]interface{}{
		"text":       fmt.Sprintf("<b>%s</b>\n%s", title, message),
		"parse_mode": "HTML", // or "MarkdownV2"
	}

	// Add recipient if provided as chat ID
	if len(recipients) > 0 {
		telegramMsg["chat_id"] = recipients[0] // Use first recipient as chat ID
	} else {
		// If no recipient provided, check if chat_id is in config
		if chatID, exists := config["chat_id"].(string); exists && chatID != "" {
			telegramMsg["chat_id"] = chatID
		} else {
			return nil, fmt.Errorf("Telegram chat_id is required")
		}
	}

	// Marshal the payload
	jsonData, err := json.Marshal(telegramMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Telegram message: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := nn.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send Telegram message: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody := make([]byte, 0)
	if resp.Body != nil {
		respBody, _ = io.ReadAll(resp.Body)
	}

	// Prepare response
	response := map[string]interface{}{
		"status":      resp.Status,
		"status_code": resp.StatusCode,
		"sent_to":     recipients,
		"title":       title,
		"message":     truncateString(message, 200),
		"timestamp":   time.Now().Unix(),
		"provider":    "telegram",
	}

	// Add response body if available
	if len(respBody) > 0 {
		response["raw_response"] = string(respBody)
	}

	return response, nil
}

// sendWebhook sends a generic webhook notification
func (nn *NotificationNode) sendWebhook(recipients []string, title, message string, config map[string]interface{}) (map[string]interface{}, error) {
	webhookURL, exists := config["webhook_url"].(string)
	if !exists || webhookURL == "" {
		return nil, fmt.Errorf("webhook URL is required")
	}

	// Prepare the webhook payload
	webhookPayload := map[string]interface{}{
		"title":     title,
		"message":   message,
		"timestamp": time.Now().Unix(),
		"type":      "notification",
		"channel":   "webhook",
	}

	// Add any additional data from config
	for k, v := range config {
		if k != "webhook_url" { // Don't override the webhook URL
			webhookPayload[k] = v
		}
	}

	// Marshal the payload
	jsonData, err := json.Marshal(webhookPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := nn.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody := make([]byte, 0)
	if resp.Body != nil {
		respBody, _ = io.ReadAll(resp.Body)
	}

	// Prepare response
	response := map[string]interface{}{
		"status":       resp.Status,
		"status_code":  resp.StatusCode,
		"title":        title,
		"message":      truncateString(message, 200),
		"timestamp":    time.Now().Unix(),
		"webhook_used": webhookURL,
		"provider":     "webhook",
	}

	// Add response body if available
	if len(respBody) > 0 {
		response["raw_response"] = string(respBody)
	}

	return response, nil
}

// applyTemplate applies a template to the input data
func (nn *NotificationNode) applyTemplate(template string, inputs map[string]interface{}) string {
	result := template

	for k, v := range inputs {
		placeholder := "{{" + k + "}}"
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", v))
	}

	return result
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}

// GetType returns the type of node
func (nn *NotificationNode) GetType() string {
	return "notification"
}

// GetID returns the unique ID of the node instance
func (nn *NotificationNode) GetID() string {
	return fmt.Sprintf("notif_%s_%d", nn.config.Channel, time.Now().Unix())
}
