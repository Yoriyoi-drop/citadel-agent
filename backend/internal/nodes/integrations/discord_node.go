// backend/internal/nodes/integrations/discord_node.go
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

// DiscordNodeConfig represents the configuration for a Discord node
type DiscordNodeConfig struct {
	Token        string                 `json:"discord_token"`
	WebhookURL   string                 `json:"webhook_url"`
	ChannelID    string                 `json:"channel_id"`
	Username     string                 `json:"username"`
	AvatarURL    string                 `json:"avatar_url"`
	Content      string                 `json:"content"`
	Embeds       []DiscordEmbed         `json:"embeds"`
	Components   []DiscordComponent     `json:"components"`
	TTS          bool                   `json:"tts"`
	AllowedMentions *DiscordMention     `json:"allowed_mentions"`
}

// DiscordEmbed represents a Discord embed
type DiscordEmbed struct {
	Title       string            `json:"title,omitempty"`
	Type        string            `json:"type,omitempty"`
	Description string            `json:"description,omitempty"`
	URL         string            `json:"url,omitempty"`
	Timestamp   string            `json:"timestamp,omitempty"`
	Color       int               `json:"color,omitempty"`
	Footer      *DiscordFooter    `json:"footer,omitempty"`
	Image       *DiscordImage     `json:"image,omitempty"`
	Thumbnail   *DiscordThumbnail `json:"thumbnail,omitempty"`
	Video       *DiscordVideo     `json:"video,omitempty"`
	Author      *DiscordAuthor    `json:"author,omitempty"`
	Fields      []DiscordField    `json:"fields,omitempty"`
}

// DiscordField represents a field in a Discord embed
type DiscordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// DiscordFooter represents a footer in a Discord embed
type DiscordFooter struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url,omitempty"`
}

// DiscordImage represents an image in a Discord embed
type DiscordImage struct {
	URL string `json:"url"`
}

// DiscordThumbnail represents a thumbnail in a Discord embed
type DiscordThumbnail struct {
	URL string `json:"url"`
}

// DiscordVideo represents a video in a Discord embed
type DiscordVideo struct {
	URL string `json:"url"`
}

// DiscordAuthor represents an author in a Discord embed
type DiscordAuthor struct {
	Name    string `json:"name"`
	URL     string `json:"url,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

// DiscordComponent represents a Discord message component
type DiscordComponent struct {
	Type       int                    `json:"type"`
	Style      int                    `json:"style,omitempty"`
	Label      string                 `json:"label,omitempty"`
	Emoji      *DiscordEmoji          `json:"emoji,omitempty"`
	CustomID   string                 `json:"custom_id,omitempty"`
	URL        string                 `json:"url,omitempty"`
	Disabled   bool                   `json:"disabled,omitempty"`
	Components []DiscordComponent     `json:"components,omitempty"`
}

// DiscordEmoji represents a Discord emoji
type DiscordEmoji struct {
	Name     string `json:"name,omitempty"`
	ID       string `json:"id,omitempty"`
	Animated bool   `json:"animated,omitempty"`
}

// DiscordMention represents mention settings
type DiscordMention struct {
	Parse       []string `json:"parse,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Users       []string `json:"users,omitempty"`
	Everyone    bool     `json:"replied_user,omitempty"`
}

// DiscordNode represents a Discord integration node
type DiscordNode struct {
	config *DiscordNodeConfig
}

// NewDiscordNode creates a new Discord node
func NewDiscordNode(config *DiscordNodeConfig) *DiscordNode {
	return &DiscordNode{
		config: config,
	}
}

// Execute executes the Discord operation
func (dn *DiscordNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Override config values with inputs if provided
	token := dn.config.Token
	if tok, exists := inputs["discord_token"]; exists {
		if tokStr, ok := tok.(string); ok {
			token = tokStr
		}
	}

	webhookURL := dn.config.WebhookURL
	if url, exists := inputs["webhook_url"]; exists {
		if urlStr, ok := url.(string); ok {
			webhookURL = urlStr
		}
	}

	channelID := dn.config.ChannelID
	if cid, exists := inputs["channel_id"]; exists {
		if cidStr, ok := cid.(string); ok {
			channelID = cidStr
		}
	}

	content := dn.config.Content
	if cont, exists := inputs["content"]; exists {
		if contStr, ok := cont.(string); ok {
			content = contStr
		}
	}

	username := dn.config.Username
	if name, exists := inputs["username"]; exists {
		if nameStr, ok := name.(string); ok {
			username = nameStr
		}
	}

	avatarURL := dn.config.AvatarURL
	if url, exists := inputs["avatar_url"]; exists {
		if urlStr, ok := url.(string); ok {
			avatarURL = urlStr
		}
	}

	tts := dn.config.TTS
	if ttsVal, exists := inputs["tts"]; exists {
		if ttsBool, ok := ttsVal.(bool); ok {
			tts = ttsBool
		}
	}

	// Validate required fields
	if webhookURL == "" && (token == "" || channelID == "") {
		return nil, fmt.Errorf("either webhook URL or both token and channel ID must be provided")
	}

	// Prepare the message payload
	payload := map[string]interface{}{
		"content": content,
		"tts":     tts,
	}

	// Add optional fields
	if username != "" {
		payload["username"] = username
	}

	if avatarURL != "" {
		payload["avatar_url"] = avatarURL
	}

	if len(dn.config.Embeds) > 0 {
		payload["embeds"] = dn.config.Embeds
	}

	if len(dn.config.Components) > 0 {
		payload["components"] = dn.config.Components
	}

	if dn.config.AllowedMentions != nil {
		payload["allowed_mentions"] = dn.config.AllowedMentions
	}

	// Override with inputs if provided
	if embeds, exists := inputs["embeds"]; exists {
		if embedsList, ok := embeds.([]interface{}); ok {
			embedList := make([]DiscordEmbed, len(embedsList))
			for i, emb := range embedsList {
				if embMap, ok := emb.(map[string]interface{}); ok {
					jsonData, _ := json.Marshal(embMap)
					json.Unmarshal(jsonData, &embedList[i])
				}
			}
			payload["embeds"] = embedList
		}
	}

	if components, exists := inputs["components"]; exists {
		if compList, ok := components.([]interface{}); ok {
			// Conversion dari interface{} ke DiscordComponent adalah proses kompleks
			// Untuk implementasi penuh, kita perlu fungsi konversi khusus
			payload["components"] = compList
		}
	}

	// Send the message
	var result map[string]interface{}
	var err error

	if webhookURL != "" {
		result, err = dn.sendViaWebhook(ctx, webhookURL, payload)
	} else {
		result, err = dn.sendViaAPI(ctx, token, channelID, payload)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to send Discord message: %w", err)
	}

	return result, nil
}

// sendViaWebhook sends a message via Discord webhook
func (dn *DiscordNode) sendViaWebhook(ctx context.Context, webhookURL string, payload map[string]interface{}) (map[string]interface{}, error) {
	// Marshal the payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
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
	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	
	result := map[string]interface{}{
		"success":     success,
		"status_code": resp.StatusCode,
		"response":    string(responseBody),
		"method":      "webhook",
		"timestamp":   time.Now().Unix(),
	}

	if !success {
		result["error"] = fmt.Sprintf("Discord webhook failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	return result, nil
}

// sendViaAPI sends a message via Discord API
func (dn *DiscordNode) sendViaAPI(ctx context.Context, token, channelID string, payload map[string]interface{}) (map[string]interface{}, error) {
	// Prepare the API URL
	apiURL := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", channelID)

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
	req.Header.Set("Authorization", "Bot "+token)

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
		// If JSON parsing fails, return the raw response
		apiResp = map[string]interface{}{
			"raw_response": string(responseBody),
		}
	}

	// Check if the request was successful
	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	
	result := map[string]interface{}{
		"success":     success,
		"status_code": resp.StatusCode,
		"response":    apiResp,
		"method":      "api",
		"timestamp":   time.Now().Unix(),
	}

	if !success {
		result["error"] = fmt.Sprintf("Discord API returned status %d", resp.StatusCode)
		if errMsg, exists := apiResp["message"]; exists {
			result["error"] = fmt.Sprintf("%s: %v", result["error"], errMsg)
		}
	}

	return result, nil
}

// RegisterDiscordNode registers the Discord node type with the engine
func RegisterDiscordNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("discord_integration", func(config map[string]interface{}) (engine.NodeInstance, error) {
		var token string
		if tok, exists := config["discord_token"]; exists {
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

		var channelID string
		if id, exists := config["channel_id"]; exists {
			if idStr, ok := id.(string); ok {
				channelID = idStr
			}
		}

		var username string
		if name, exists := config["username"]; exists {
			if nameStr, ok := name.(string); ok {
				username = nameStr
			}
		}

		var avatarURL string
		if url, exists := config["avatar_url"]; exists {
			if urlStr, ok := url.(string); ok {
				avatarURL = urlStr
			}
		}

		var content string
		if cont, exists := config["content"]; exists {
			if contStr, ok := cont.(string); ok {
				content = contStr
			}
		}

		var tts bool
		if ttsVal, exists := config["tts"]; exists {
			if ttsBool, ok := ttsVal.(bool); ok {
				tts = ttsBool
			}
		}

		// Parse embeds if provided
		var embeds []DiscordEmbed
		if embList, exists := config["embeds"]; exists {
			if embSlice, ok := embList.([]interface{}); ok {
				for _, emb := range embSlice {
					if embMap, ok := emb.(map[string]interface{}); ok {
						jsonData, _ := json.Marshal(embMap)
						var embed DiscordEmbed
						json.Unmarshal(jsonData, &embed)
						embeds = append(embeds, embed)
					}
				}
			}
		}

		// For components and mentions, we would implement similar parsing
		// Due to complexity, we'll leave those as exercises for full implementation

		nodeConfig := &DiscordNodeConfig{
			Token:       token,
			WebhookURL:  webhookURL,
			ChannelID:   channelID,
			Username:    username,
			AvatarURL:   avatarURL,
			Content:     content,
			Embeds:      embeds,
			TTS:         tts,
		}

		return NewDiscordNode(nodeConfig), nil
	})
}