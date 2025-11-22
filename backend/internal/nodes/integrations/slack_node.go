// backend/internal/nodes/integrations/slack_node.go
package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/citadel-agent/backend/internal/workflow/core/engine"
)

// SlackNodeConfig represents the configuration for a Slack node
type SlackNodeConfig struct {
	SlackToken     string                 `json:"slack_token"`
	WebhookURL     string                 `json:"webhook_url"`
	Channel        string                 `json:"channel"`
	Username       string                 `json:"username"`
	IconEmoji      string                 `json:"icon_emoji"`
	IconURL        string                 `json:"icon_url"`
	ThreadTS       string                 `json:"thread_ts,omitempty"`
	Attachments    []SlackAttachment      `json:"attachments,omitempty"`
	Blocks         []map[string]interface{} `json:"blocks,omitempty"`
}

// SlackAttachment represents a Slack attachment
type SlackAttachment struct {
	Color       string                 `json:"color,omitempty"`
	Fallback    string                 `json:"fallback"`
	Text        string                 `json:"text"`
	Title       string                 `json:"title,omitempty"`
	TitleLink   string                 `json:"title_link,omitempty"`
	Pretext     string                 `json:"pretext,omitempty"`
	ImageURL    string                 `json:"image_url,omitempty"`
	ThumbURL    string                 `json:"thumb_url,omitempty"`
	Footer      string                 `json:"footer,omitempty"`
	FooterIcon  string                 `json:"footer_icon,omitempty"`
	Timestamp   int64                  `json:"ts,omitempty"`
	Actions     []SlackAction          `json:"actions,omitempty"`
	Fields      []SlackField           `json:"fields,omitempty"`
}

// SlackField represents a field in a Slack attachment
type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// SlackAction represents an action in a Slack message
type SlackAction struct {
	Name  string `json:"name"`
	Text  string `json:"text"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// SlackNode represents a Slack integration node
type SlackNode struct {
	config *SlackNodeConfig
}

// NewSlackNode creates a new Slack node
func NewSlackNode(config *SlackNodeConfig) *SlackNode {
	return &SlackNode{
		config: config,
	}
}

// Execute executes the Slack operation
func (sn *SlackNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Override config values with inputs if provided
	token := sn.config.SlackToken
	if tok, exists := inputs["slack_token"]; exists {
		if tokStr, ok := tok.(string); ok {
			token = tokStr
		}
	}

	channel := sn.config.Channel
	if ch, exists := inputs["channel"]; exists {
		if chStr, ok := ch.(string); ok {
			channel = chStr
		}
	}

	text := ""
	if txt, exists := inputs["text"]; exists {
		if txtStr, ok := txt.(string); ok {
			text = txtStr
		}
	}

	// Check if we have a webhook URL or should use API token
	if sn.config.WebhookURL != "" {
		return sn.sendViaWebhook(ctx, text, inputs)
	} else if token != "" {
		return sn.sendViaAPI(ctx, token, channel, text, inputs)
	} else {
		return nil, fmt.Errorf("either webhook URL or API token must be provided")
	}
}

// sendViaWebhook sends a message via Slack webhook
func (sn *SlackNode) sendViaWebhook(ctx context.Context, text string, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Prepare the payload
	payload := map[string]interface{}{
		"text": text,
	}

	// Override with specific inputs
	if channel, exists := inputs["channel"]; exists {
		if chStr, ok := channel.(string); ok {
			payload["channel"] = chStr
		}
	}

	if username, exists := inputs["username"]; exists {
		if userStr, ok := username.(string); ok {
			payload["username"] = userStr
		}
	}

	if iconEmoji, exists := inputs["icon_emoji"]; exists {
		if emojiStr, ok := iconEmoji.(string); ok {
			payload["icon_emoji"] = emojiStr
		}
	}

	// Marshal the payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "POST", sn.config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check if the request was successful
	success := resp.StatusCode == 200 && string(responseBody) == "ok"
	
	result := map[string]interface{}{
		"success":     success,
		"status_code": resp.StatusCode,
		"response":    string(responseBody),
		"method":      "webhook",
		"timestamp":   time.Now().Unix(),
	}

	if !success {
		result["error"] = fmt.Sprintf("Slack webhook failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	return result, nil
}

// sendViaAPI sends a message via Slack API
func (sn *SlackNode) sendViaAPI(ctx context.Context, token, channel, text string, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Prepare the API URL
	apiURL := "https://slack.com/api/chat.postMessage"

	// Prepare the payload
	payload := map[string]interface{}{
		"channel": channel,
		"text":    text,
	}

	// Add optional fields from config
	if sn.config.Username != "" {
		payload["username"] = sn.config.Username
	}

	if sn.config.IconEmoji != "" {
		payload["icon_emoji"] = sn.config.IconEmoji
	}

	if sn.config.IconURL != "" {
		payload["icon_url"] = sn.config.IconURL
	}

	if sn.config.ThreadTS != "" {
		payload["thread_ts"] = sn.config.ThreadTS
	}

	if len(sn.config.Attachments) > 0 {
		payload["attachments"] = sn.config.Attachments
	}

	if len(sn.config.Blocks) > 0 {
		payload["blocks"] = sn.config.Blocks
	}

	// Override with specific inputs
	if attachments, exists := inputs["attachments"]; exists {
		if attachmentsArr, ok := attachments.([]interface{}); ok {
			attachmentList := make([]SlackAttachment, len(attachmentsArr))
			for i, att := range attachmentsArr {
				if attMap, ok := att.(map[string]interface{}); ok {
					jsonData, _ := json.Marshal(attMap)
					json.Unmarshal(jsonData, &attachmentList[i])
				}
			}
			payload["attachments"] = attachmentList
		}
	}

	// Marshal the payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Execute the request
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse the response
	var apiResp map[string]interface{}
	if err := json.Unmarshal(responseBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check if the request was successful
	ok, exists := apiResp["ok"].(bool)
	success := exists && ok
	
	result := map[string]interface{}{
		"success":     success,
		"status_code": resp.StatusCode,
		"response":    apiResp,
		"method":      "api",
		"timestamp":   time.Now().Unix(),
	}

	if !success {
		result["error"] = apiResp["error"]
	}

	return result, nil
}

// RegisterSlackNode registers the Slack node type with the engine
func RegisterSlackNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("slack_integration", func(config map[string]interface{}) (engine.NodeInstance, error) {
		var token string
		if tok, exists := config["slack_token"]; exists {
			if tokStr, ok := tok.(string); ok {
				token = tokStr
			}
		}

		var webhookURL string
		if url, exists := config["webhook_url"]; exists {
			if urlStr, ok := url.(string); ok {
				webhookURL = urlStr
			}
		}

		var channel string
		if ch, exists := config["channel"]; exists {
			if chStr, ok := ch.(string); ok {
				channel = chStr
			}
		}

		var username string
		if user, exists := config["username"]; exists {
			if userStr, ok := user.(string); ok {
				username = userStr
			}
		}

		var iconEmoji string
		if emoji, exists := config["icon_emoji"]; exists {
			if emojiStr, ok := emoji.(string); ok {
				iconEmoji = emojiStr
			}
		}

		var iconURL string
		if url, exists := config["icon_url"]; exists {
			if urlStr, ok := url.(string); ok {
				iconURL = urlStr
			}
		}

		var threadTS string
		if ts, exists := config["thread_ts"]; exists {
			if tsStr, ok := ts.(string); ok {
				threadTS = tsStr
			}
		}

		var attachments []SlackAttachment
		if atts, exists := config["attachments"]; exists {
			if attsArr, ok := atts.([]interface{}); ok {
				for _, att := range attsArr {
					if attMap, ok := att.(map[string]interface{}); ok {
						jsonData, _ := json.Marshal(attMap)
						var attachment SlackAttachment
						json.Unmarshal(jsonData, &attachment)
						attachments = append(attachments, attachment)
					}
				}
			}
		}

		var blocks []map[string]interface{}
		if blks, exists := config["blocks"]; exists {
			if blksArr, ok := blks.([]interface{}); ok {
				for _, blk := range blksArr {
					if blkMap, ok := blk.(map[string]interface{}); ok {
						blocks = append(blocks, blkMap)
					}
				}
			}
		}

		nodeConfig := &SlackNodeConfig{
			SlackToken:  token,
			WebhookURL:  webhookURL,
			Channel:     channel,
			Username:    username,
			IconEmoji:   iconEmoji,
			IconURL:     iconURL,
			ThreadTS:    threadTS,
			Attachments: attachments,
			Blocks:      blocks,
		}

		return NewSlackNode(nodeConfig), nil
	})
}