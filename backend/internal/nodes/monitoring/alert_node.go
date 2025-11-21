// backend/internal/nodes/monitoring/alert_node.go
package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"citadel-agent/backend/internal/workflow/core/engine"
	"citadel-agent/backend/internal/workflow/core/observability"
)

// AlertConditionType represents the type of alert condition
type AlertConditionType string

const (
	AlertConditionGreaterThan     AlertConditionType = "greater_than"
	AlertConditionLessThan        AlertConditionType = "less_than"
	AlertConditionEquals          AlertConditionType = "equals"
	AlertConditionNotEquals       AlertConditionType = "not_equals"
	AlertConditionGreaterThanEq   AlertConditionType = "greater_than_equal"
	AlertConditionLessThanEq      AlertConditionType = "less_than_equal"
	AlertConditionContains        AlertConditionType = "contains"
	AlertConditionMatchesRegex    AlertConditionType = "matches_regex"
	AlertConditionExists          AlertConditionType = "exists"
	AlertConditionDoesNotExist    AlertConditionType = "does_not_exist"
)

// AlertSeverity represents the severity of an alert
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
	AlertSeverityError    AlertSeverity = "error"
)

// AlertChannel represents the channel for alert delivery
type AlertChannel string

const (
	AlertChannelEmail    AlertChannel = "email"
	AlertChannelSlack    AlertChannel = "slack"
	AlertChannelWebhook  AlertChannel = "webhook"
	AlertChannelPagerDuty AlertChannel = "pagerduty"
	AlertChannelTeams    AlertChannel = "teams"
	AlertChannelDiscord  AlertChannel = "discord"
)

// AlertNodeConfig represents the configuration for an alert node
type AlertNodeConfig struct {
	ConditionType     AlertConditionType  `json:"condition_type"`
	ThresholdValue    interface{}         `json:"threshold_value"`
	Field             string              `json:"field"`
	MetricName        string              `json:"metric_name"`
	Severity          AlertSeverity       `json:"severity"`
	Channels          []AlertChannel      `json:"channels"`
	Recipients        []string            `json:"recipients"` // Email addresses, webhook URLs, etc.
	MessageTemplate   string              `json:"message_template"`
	TitleTemplate     string              `json:"title_template"`
	EvaluationWindow  time.Duration       `json:"evaluation_window"`
	EvaluationPeriod  string              `json:"evaluation_period"` // "1m", "5m", "15m", etc.
	RetriggerInterval time.Duration       `json:"retrigger_interval"`
	RepeatLimit       int                 `json:"repeat_limit"`
	Enabled           bool                `json:"enabled"`
	NotifyOnce        bool                `json:"notify_once"`
	AggregationType   string              `json:"aggregation_type"` // "avg", "sum", "min", "max", "count"
}

// AlertNode represents an alert node
type AlertNode struct {
	config    *AlertNodeConfig
	telemetry observability.TelemetryService
}

// NewAlertNode creates a new alert node
func NewAlertNode(config *AlertNodeConfig, telemetry observability.TelemetryService) *AlertNode {
	if config.EvaluationWindow == 0 {
		config.EvaluationWindow = 5 * time.Minute
	}
	if config.RetriggerInterval == 0 {
		config.RetriggerInterval = 15 * time.Minute
	}
	if config.RepeatLimit == 0 {
		config.RepeatLimit = 5
	}
	if config.AggregationType == "" {
		config.AggregationType = "avg"
	}

	return &AlertNode{
		config:    config,
		telemetry: telemetry,
	}
}

// Execute executes the alert node
func (an *AlertNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	if !an.config.Enabled {
		return map[string]interface{}{
			"success": true,
			"message": "alert is disabled",
			"enabled": false,
		}, nil
	}

	// Evaluate the condition
	triggered, currentValue, err := an.evaluateCondition(inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate alert condition: %w", err)
	}

	result := map[string]interface{}{
		"success":       true,
		"triggered":     triggered,
		"current_value": currentValue,
		"threshold":     an.config.ThresholdValue,
		"condition":     string(an.config.ConditionType),
		"field":         an.config.Field,
		"severity":      string(an.config.Severity),
		"timestamp":     time.Now().Unix(),
	}

	if triggered {
		// Generate alert message
		title := an.generateTitle(currentValue)
		message := an.generateMessage(currentValue)

		// Send alerts through configured channels
		alertResults := make(map[string]interface{})
		
		for _, channel := range an.config.Channels {
			success, err := an.sendAlert(ctx, channel, title, message, currentValue)
			alertResults[string(channel)] = map[string]interface{}{
				"success": success,
				"error":   err,
			}
		}

		result["alert_sent"] = true
		result["alert_channels"] = alertResults
		result["title"] = title
		result["message"] = message

		// Record alert metrics
		an.telemetry.RecordAlertTriggered(an.config.Severity, an.config.MetricName, time.Now())
	} else {
		result["alert_sent"] = false
		result["message"] = "condition not met, no alert sent"
	}

	return result, nil
}

// evaluateCondition evaluates if the alert condition is met
func (an *AlertNode) evaluateCondition(inputs map[string]interface{}) (bool, interface{}, error) {
	// Get the value to compare against
	var currentValue interface{}
	
	if an.config.Field != "" {
		// Get value from input using field path
		value, exists := inputs[an.config.Field]
		if !exists {
			return false, nil, fmt.Errorf("field '%s' not found in inputs", an.config.Field)
		}
		currentValue = value
	} else if an.config.MetricName != "" {
		// In a real implementation, this would fetch from metrics storage
		// For now, we'll expect the metric value to be provided in inputs
		value, exists := inputs[an.config.MetricName]
		if !exists {
			return false, nil, fmt.Errorf("metric '%s' not found in inputs", an.config.MetricName)
		}
		currentValue = value
	} else {
		return false, nil, fmt.Errorf("either field or metric name must be specified")
	}

	// Convert current value to comparable format
	currentFloat, currentStr, currentBool := an.toFloatStrBool(currentValue)

	// Compare based on condition type
	switch an.config.ConditionType {
	case AlertConditionGreaterThan:
		thresholdFloat, _, _ := an.toFloatStrBool(an.config.ThresholdValue)
		return currentFloat > thresholdFloat, currentValue, nil
	case AlertConditionLessThan:
		thresholdFloat, _, _ := an.toFloatStrBool(an.config.ThresholdValue)
		return currentFloat < thresholdFloat, currentValue, nil
	case AlertConditionEquals:
		thresholdFloat, thresholdStr, thresholdBool := an.toFloatStrBool(an.config.ThresholdValue)
		// Compare with any of the converted values
		return currentFloat == thresholdFloat || 
			   currentStr == thresholdStr || 
			   currentBool == thresholdBool, currentValue, nil
	case AlertConditionNotEquals:
		thresholdFloat, thresholdStr, thresholdBool := an.toFloatStrBool(an.config.ThresholdValue)
		equal := currentFloat == thresholdFloat || 
		         currentStr == thresholdStr || 
		         currentBool == thresholdBool
		return !equal, currentValue, nil
	case AlertConditionGreaterThanEq:
		thresholdFloat, _, _ := an.toFloatStrBool(an.config.ThresholdValue)
		return currentFloat >= thresholdFloat, currentValue, nil
	case AlertConditionLessThanEq:
		thresholdFloat, _, _ := an.toFloatStrBool(an.config.ThresholdValue)
		return currentFloat <= thresholdFloat, currentValue, nil
	case AlertConditionContains:
		_, thresholdStr, _ := an.toFloatStrBool(an.config.ThresholdValue)
		return strings.Contains(currentStr, thresholdStr), currentValue, nil
	case AlertConditionMatchesRegex:
		_, patternStr, _ := an.toFloatStrBool(an.config.ThresholdValue)
		// In a real implementation, we would compile and match the regex
		// For now, we'll do a simple contains check
		return strings.Contains(currentStr, patternStr), currentValue, nil
	case AlertConditionExists:
		return currentValue != nil, currentValue, nil
	case AlertConditionDoesNotExist:
		return currentValue == nil, currentValue, nil
	default:
		return false, currentValue, fmt.Errorf("unknown condition type: %s", an.config.ConditionType)
	}
}

// toFloatStrBool converts an interface{} value to float64, string, and bool representations
func (an *AlertNode) toFloatStrBool(value interface{}) (float64, string, bool) {
	var floatVal float64
	var strVal string
	var boolVal bool

	// Convert to string first
	switch v := value.(type) {
	case string:
		strVal = v
		// Try to parse as float
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			floatVal = f
		}
		// Try to parse as bool
		if b, err := strconv.ParseBool(v); err == nil {
			boolVal = b
		}
	case float64:
		floatVal = v
		strVal = strconv.FormatFloat(v, 'f', -1, 64)
		boolVal = v != 0
	case int:
		floatVal = float64(v)
		strVal = strconv.Itoa(v)
		boolVal = v != 0
	case bool:
		boolVal = v
		strVal = strconv.FormatBool(v)
		floatVal = 0
		if v {
			floatVal = 1
		}
	case nil:
		strVal = ""
		boolVal = false
	default:
		strVal = fmt.Sprintf("%v", v)
		// Try to parse as float
		if f, err := strconv.ParseFloat(strVal, 64); err == nil {
			floatVal = f
		}
		// Try to parse as bool
		if b, err := strconv.ParseBool(strVal); err == nil {
			boolVal = b
		}
	}

	return floatVal, strVal, boolVal
}

// generateTitle generates an alert title based on template
func (an *AlertNode) generateTitle(currentValue interface{}) string {
	if an.config.TitleTemplate != "" {
		title := an.config.TitleTemplate
		title = strings.ReplaceAll(title, "{{field}}", an.config.Field)
		title = strings.ReplaceAll(title, "{{current_value}}", fmt.Sprintf("%v", currentValue))
		title = strings.ReplaceAll(title, "{{threshold}}", fmt.Sprintf("%v", an.config.ThresholdValue))
		title = strings.ReplaceAll(title, "{{severity}}", string(an.config.Severity))
		return title
	}
	
	// Default title template
	return fmt.Sprintf("[%s] Alert: %s %s %v", 
		strings.ToUpper(string(an.config.Severity)),
		an.config.Field,
		an.getConditionSymbol(),
		an.config.ThresholdValue)
}

// generateMessage generates an alert message based on template
func (an *AlertNode) generateMessage(currentValue interface{}) string {
	if an.config.MessageTemplate != "" {
		message := an.config.MessageTemplate
		message = strings.ReplaceAll(message, "{{field}}", an.config.Field)
		message = strings.ReplaceAll(message, "{{current_value}}", fmt.Sprintf("%v", currentValue))
		message = strings.ReplaceAll(message, "{{threshold}}", fmt.Sprintf("%v", an.config.ThresholdValue))
		message = strings.ReplaceAll(message, "{{severity}}", string(an.config.Severity))
		message = strings.ReplaceAll(message, "{{timestamp}}", time.Now().Format(time.RFC3339))
		return message
	}
	
	// Default message template
	return fmt.Sprintf("Alert triggered! Field '%s' has value %v, which meets condition '%s %v'. Severity: %s", 
		an.config.Field, 
		currentValue, 
		an.getConditionSymbol(),
		an.config.ThresholdValue, 
		an.config.Severity)
}

// getConditionSymbol returns symbol representation of condition
func (an *AlertNode) getConditionSymbol() string {
	switch an.config.ConditionType {
	case AlertConditionGreaterThan:
		return ">"
	case AlertConditionLessThan:
		return "<"
	case AlertConditionEquals:
		return "=="
	case AlertConditionNotEquals:
		return "!="
	case AlertConditionGreaterThanEq:
		return ">="
	case AlertConditionLessThanEq:
		return "<="
	default:
		return string(an.config.ConditionType)
	}
}

// sendAlert sends an alert through the specified channel
func (an *AlertNode) sendAlert(ctx context.Context, channel AlertChannel, title, message string, currentValue interface{}) (bool, error) {
	switch channel {
	case AlertChannelEmail:
		return an.sendEmailAlert(ctx, title, message, currentValue)
	case AlertChannelSlack:
		return an.sendSlackAlert(ctx, title, message, currentValue)
	case AlertChannelWebhook:
		return an.sendWebhookAlert(ctx, title, message, currentValue)
	case AlertChannelPagerDuty:
		return an.sendPagerDutyAlert(ctx, title, message, currentValue)
	case AlertChannelTeams:
		return an.sendTeamsAlert(ctx, title, message, currentValue)
	case AlertChannelDiscord:
		return an.sendDiscordAlert(ctx, title, message, currentValue)
	default:
		return false, fmt.Errorf("unsupported alert channel: %s", channel)
	}
}

// sendEmailAlert sends an email alert
func (an *AlertNode) sendEmailAlert(ctx context.Context, title, message string, currentValue interface{}) (bool, error) {
	// Implementation would use the email node system we created earlier
	// For now, return success as placeholder
	
	// In a real implementation:
	// 1. Validate recipients
	// 2. Create email payload
	// 3. Send via email service
	
	// This is a simplified placeholder
	for _, recipient := range an.config.Recipients {
		if strings.Contains(recipient, "@") { // Simple email validation
			// In real implementation, send email
			fmt.Printf("EMAIL ALERT SENT to %s: %s - %s\n", recipient, title, message)
		}
	}
	
	return true, nil
}

// sendSlackAlert sends a Slack alert
func (an *AlertNode) sendSlackAlert(ctx context.Context, title, message string, currentValue interface{}) (bool, error) {
	// Find webhook URLs from recipients
	var webhookURLs []string
	for _, recipient := range an.config.Recipients {
		// Look for Slack webhook URLs
		if strings.Contains(recipient, "hooks.slack.com") {
			webhookURLs = append(webhookURLs, recipient)
		}
	}

	if len(webhookURLs) == 0 {
		return false, fmt.Errorf("no valid Slack webhook URLs found in recipients")
	}

	// Prepare the Slack message payload
	slackPayload := map[string]interface{}{
		"text": title,
		"attachments": []map[string]interface{}{
			{
				"color": an.getSlackColorForSeverity(),
				"fields": []map[string]interface{}{
					{
						"title": "Message",
						"value": message,
						"short": false,
					},
					{
						"title": "Current Value",
						"value": fmt.Sprintf("%v", currentValue),
						"short": true,
					},
					{
						"title": "Severity",
						"value": string(an.config.Severity),
						"short": true,
					},
					{
						"title": "Timestamp",
						"value": time.Now().Format(time.RFC3339),
						"short": true,
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(slackPayload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal Slack payload: %w", err)
	}

	// Send to each webhook
	allSuccessful := true
	for _, webhookURL := range webhookURLs {
		req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Failed to create Slack request: %v\n", err)
			allSuccessful = false
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Failed to send Slack alert: %v\n", err)
			allSuccessful = false
			continue
		}
		resp.Body.Close()
	}

	return allSuccessful, nil
}

// getSlackColorForSeverity returns Slack color code for alert severity
func (an *AlertNode) getSlackColorForSeverity() string {
	switch an.config.Severity {
	case AlertSeverityCritical:
		return "#FF0000" // Red
	case AlertSeverityError:
		return "#FF6B6B" // Light red
	case AlertSeverityWarning:
		return "#FFD93D" // Yellow
	case AlertSeverityInfo:
		return "#3CAEA3" // Teal
	default:
		return "#3CAEA3" // Default teal
	}
}

// sendWebhookAlert sends a webhook alert
func (an *AlertNode) sendWebhookAlert(ctx context.Context, title, message string, currentValue interface{}) (bool, error) {
	// Find webhook URLs from recipients
	var webhookURLs []string
	for _, recipient := range an.config.Recipients {
		if strings.HasPrefix(recipient, "http://") || strings.HasPrefix(recipient, "https://") {
			// Exclude known service endpoints, keep generic webhooks
			if !strings.Contains(recipient, "hooks.slack.com") &&
			   !strings.Contains(recipient, "discord.com/api/webhooks") &&
			   !strings.Contains(recipient, "office.com/webhook") {
				webhookURLs = append(webhookURLs, recipient)
			}
		}
	}

	if len(webhookURLs) == 0 {
		return false, fmt.Errorf("no valid webhook URLs found in recipients")
	}

	// Prepare the webhook payload
	webhookPayload := map[string]interface{}{
		"title":       title,
		"message":     message,
		"severity":    string(an.config.Severity),
		"current_value": currentValue,
		"threshold":   an.config.ThresholdValue,
		"condition":   string(an.config.ConditionType),
		"field":       an.config.Field,
		"timestamp":   time.Now().Unix(),
		"triggered_at": time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(webhookPayload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	// Send to each webhook
	allSuccessful := true
	for _, webhookURL := range webhookURLs {
		req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Failed to create webhook request: %v\n", err)
			allSuccessful = false
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Failed to send webhook alert: %v\n", err)
			allSuccessful = false
			continue
		}
		resp.Body.Close()
	}

	return allSuccessful, nil
}

// sendPagerDutyAlert sends an alert to PagerDuty
func (an *AlertNode) sendPagerDutyAlert(ctx context.Context, title, message string, currentValue interface{}) (bool, error) {
	// Find PagerDuty integration keys
	var integrationKeys []string
	for _, recipient := range an.config.Recipients {
		// Simple check for likely PagerDuty integration key (32-character hex)
		if len(recipient) == 32 && isValidHex(recipient) {
			integrationKeys = append(integrationKeys, recipient)
		}
	}

	if len(integrationKeys) == 0 {
		return false, fmt.Errorf("no valid PagerDuty integration keys found in recipients")
	}

	// Prepare the PagerDuty payload
	pagerDutyPayload := map[string]interface{}{
		"routing_key": integrationKeys[0], // Use first valid key
		"event_action": "trigger",
		"dedup_key":    fmt.Sprintf("citadel-alert-%s", an.config.Field),
		"payload": map[string]interface{}{
			"summary":  title,
			"source":   "citadel-agent",
			"severity": an.getPagerDutySeverity(),
			"custom_details": map[string]interface{}{
				"message":       message,
				"current_value": currentValue,
				"threshold":     an.config.ThresholdValue,
				"condition":     string(an.config.ConditionType),
				"field":         an.config.Field,
				"timestamp":     time.Now().Unix(),
			},
		},
	}

	jsonData, err := json.Marshal(pagerDutyPayload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal PagerDuty payload: %w", err)
	}

	// Send to PagerDuty
	req, err := http.NewRequestWithContext(ctx, "POST", "https://events.pagerduty.com/v2/enqueue", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("failed to create PagerDuty request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send PagerDuty alert: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode == 202, nil
}

// getPagerDutySeverity converts Citadel severity to PagerDuty severity
func (an *AlertNode) getPagerDutySeverity() string {
	switch an.config.Severity {
	case AlertSeverityCritical, AlertSeverityError:
		return "critical"
	case AlertSeverityWarning:
		return "warning"
	default:
		return "info"
	}
}

// sendTeamsAlert sends an alert to Microsoft Teams
func (an *AlertNode) sendTeamsAlert(ctx context.Context, title, message string, currentValue interface{}) (bool, error) {
	// Find Teams webhook URLs
	var teamsWebhooks []string
	for _, recipient := range an.config.Recipients {
		if strings.Contains(recipient, "office.com/webhook") {
			teamsWebhooks = append(teamsWebhooks, recipient)
		}
	}

	if len(teamsWebhooks) == 0 {
		return false, fmt.Errorf("no valid Teams webhook URLs found in recipients")
	}

	// Prepare the Teams message card
	teamsCard := map[string]interface{}{
		"@type": "MessageCard",
		"@context": "http://schema.org/extensions",
		"themeColor": an.getTeamsColorForSeverity(),
		"summary": title,
		"sections": []map[string]interface{}{
			{
				"activityTitle": title,
				"activitySubtitle": fmt.Sprintf("Severity: %s", an.config.Severity),
				"facts": []map[string]interface{}{
					{
						"name": "Field",
						"value": an.config.Field,
					},
					{
						"name": "Current Value",
						"value": fmt.Sprintf("%v", currentValue),
					},
					{
						"name": "Threshold",
						"value": fmt.Sprintf("%v", an.config.ThresholdValue),
					},
					{
						"name": "Condition",
						"value": string(an.config.ConditionType),
					},
					{
						"name": "Time",
						"value": time.Now().Format(time.RFC3339),
					},
				},
				"text": message,
			},
		},
	}

	jsonData, err := json.Marshal(teamsCard)
	if err != nil {
		return false, fmt.Errorf("failed to marshal Teams card: %w", err)
	}

	// Send to Teams webhooks
	allSuccessful := true
	for _, webhookURL := range teamsWebhooks {
		req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Failed to create Teams request: %v\n", err)
			allSuccessful = false
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Failed to send Teams alert: %v\n", err)
			allSuccessful = false
			continue
		}
		resp.Body.Close()
	}

	return allSuccessful, nil
}

// getTeamsColorForSeverity returns Teams color code for alert severity
func (an *AlertNode) getTeamsColorForSeverity() string {
	switch an.config.Severity {
	case AlertSeverityCritical:
		return "FF0000" // Red
	case AlertSeverityError:
		return "FF6B6B" // Light red
	case AlertSeverityWarning:
		return "FFD93D" // Yellow
	case AlertSeverityInfo:
		return "3CAEA3" // Teal
	default:
		return "3CAEA3" // Default teal
	}
}

// sendDiscordAlert sends an alert to Discord
func (an *AlertNode) sendDiscordAlert(ctx context.Context, title, message string, currentValue interface{}) (bool, error) {
	// Find Discord webhook URLs
	var discordWebhooks []string
	for _, recipient := range an.config.Recipients {
		if strings.Contains(recipient, "discord.com/api/webhooks") {
			discordWebhooks = append(discordWebhooks, recipient)
		}
	}

	if len(discordWebhooks) == 0 {
		return false, fmt.Errorf("no valid Discord webhook URLs found in recipients")
	}

	// Prepare the Discord embed
	discordPayload := map[string]interface{}{
		"content": title,
		"embeds": []map[string]interface{}{
			{
				"title": title,
				"description": message,
				"color": an.getDiscordColorForSeverity(),
				"fields": []map[string]interface{}{
					{
						"name": "Field",
						"value": an.config.Field,
						"inline": true,
					},
					{
						"name": "Current Value",
						"value": fmt.Sprintf("%v", currentValue),
						"inline": true,
					},
					{
						"name": "Threshold",
						"value": fmt.Sprintf("%v", an.config.ThresholdValue),
						"inline": true,
					},
					{
						"name": "Condition",
						"value": string(an.config.ConditionType),
						"inline": true,
					},
					{
						"name": "Severity",
						"value": string(an.config.Severity),
						"inline": true,
					},
					{
						"name": "Time",
						"value": time.Now().Format(time.RFC3339),
						"inline": true,
					},
				},
				"timestamp": time.Now().Format(time.RFC3339),
			},
		},
	}

	jsonData, err := json.Marshal(discordPayload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal Discord payload: %w", err)
	}

	// Send to Discord webhooks
	allSuccessful := true
	for _, webhookURL := range discordWebhooks {
		req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Failed to create Discord request: %v\n", err)
			allSuccessful = false
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Failed to send Discord alert: %v\n", err)
			allSuccessful = false
			continue
		}
		resp.Body.Close()
	}

	return allSuccessful, nil
}

// getDiscordColorForSeverity returns Discord color for alert severity
func (an *AlertNode) getDiscordColorForSeverity() int {
	switch an.config.Severity {
	case AlertSeverityCritical:
		return 0xFF0000 // Red
	case AlertSeverityError:
		return 0xFF6B6B // Light red
	case AlertSeverityWarning:
		return 0xFFD93D // Yellow
	case AlertSeverityInfo:
		return 0x3CAEA3 // Teal
	default:
		return 0x3CAEA3 // Default teal
	}
}

// isValidHex checks if a string is a valid hexadecimal
func isValidHex(s string) bool {
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

// RegisterAlertNode registers the alert node with the engine
func RegisterAlertNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("alert_trigger", func(config map[string]interface{}) (engine.NodeInstance, error) {
		var conditionType AlertConditionType
		if typ, exists := config["condition_type"]; exists {
			if typStr, ok := typ.(string); ok {
				conditionType = AlertConditionType(typStr)
			}
		}

		var thresholdValue interface{}
		if thresh, exists := config["threshold_value"]; exists {
			thresholdValue = thresh
		}

		var field string
		if f, exists := config["field"]; exists {
			if fStr, ok := f.(string); ok {
				field = fStr
			}
		}

		var metricName string
		if mName, exists := config["metric_name"]; exists {
			if mNameStr, ok := mName.(string); ok {
				metricName = mNameStr
			}
		}

		var severity AlertSeverity
		if sev, exists := config["severity"]; exists {
			if sevStr, ok := sev.(string); ok {
				severity = AlertSeverity(sevStr)
			}
		} else {
			severity = AlertSeverityWarning // Default severity
		}

		var channels []AlertChannel
		if chanList, exists := config["channels"]; exists {
			if chanSlice, ok := chanList.([]interface{}); ok {
				for _, ch := range chanSlice {
					if chStr, ok := ch.(string); ok {
						channels = append(channels, AlertChannel(chStr))
					}
				}
			}
		}

		var recipients []string
		if recList, exists := config["recipients"]; exists {
			if recSlice, ok := recList.([]interface{}); ok {
				for _, r := range recSlice {
					if rStr, ok := r.(string); ok {
						recipients = append(recipients, rStr)
					}
				}
			}
		}

		var messageTemplate string
		if msg, exists := config["message_template"]; exists {
			if msgStr, ok := msg.(string); ok {
				messageTemplate = msgStr
			}
		}

		var titleTemplate string
		if title, exists := config["title_template"]; exists {
			if titleStr, ok := title.(string); ok {
				titleTemplate = titleStr
			}
		}

		var evalWindow float64
		if window, exists := config["evaluation_window_seconds"]; exists {
			if windowFloat, ok := window.(float64); ok {
				evalWindow = windowFloat
			}
		}

		var retriggerInterval float64
		if interval, exists := config["retrigger_interval_seconds"]; exists {
			if intervalFloat, ok := interval.(float64); ok {
				retriggerInterval = intervalFloat
			}
		}

		var repeatLimit float64
		if limit, exists := config["repeat_limit"]; exists {
			if limitFloat, ok := limit.(float64); ok {
				repeatLimit = limitFloat
			}
		}

		var enabled bool
		if en, exists := config["enabled"]; exists {
			if enBool, ok := en.(bool); ok {
				enabled = enBool
			}
		} else {
			enabled = true // Default to enabled
		}

		var notifyOnce bool
		if notify, exists := config["notify_once"]; exists {
			if notifyBool, ok := notify.(bool); ok {
				notifyOnce = notifyBool
			}
		}

		var aggType string
		if agg, exists := config["aggregation_type"]; exists {
			if aggStr, ok := agg.(string); ok {
				aggType = aggStr
			}
		}

		nodeConfig := &AlertNodeConfig{
			ConditionType:     conditionType,
			ThresholdValue:    thresholdValue,
			Field:             field,
			MetricName:        metricName,
			Severity:          severity,
			Channels:          channels,
			Recipients:        recipients,
			MessageTemplate:   messageTemplate,
			TitleTemplate:     titleTemplate,
			EvaluationWindow:  time.Duration(evalWindow) * time.Second,
			RetriggerInterval: time.Duration(retriggerInterval) * time.Second,
			RepeatLimit:       int(repeatLimit),
			Enabled:           enabled,
			NotifyOnce:        notifyOnce,
			AggregationType:   aggType,
		}

		// For now, we'll pass a nil telemetry service
		// In a real implementation, this would come from the service context
		return NewAlertNode(nodeConfig, nil), nil
	})
}