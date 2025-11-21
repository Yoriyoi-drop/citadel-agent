// backend/internal/nodes/integrations/telegram_node.go
package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"citadel-agent/backend/internal/workflow/core/engine"
)

// TelegramNodeConfig represents the configuration for a Telegram node
type TelegramNodeConfig struct {
	Token       string `json:"telegram_token"`
	ChatID      string `json:"chat_id"`
	MessageText string `json:"message_text"`
	ParseMode   string `json:"parse_mode"` // "HTML", "Markdown", "MarkdownV2"
	DisableWebPagePreview bool `json:"disable_web_page_preview"`
	DisableNotification   bool `json:"disable_notification"`
	ReplyToMessageID      *int `json:"reply_to_message_id,omitempty"`
	InlineKeyboard        *[][]TelegramInlineKeyboardButton `json:"inline_keyboard,omitempty"`
	ReplyMarkup           interface{} `json:"reply_markup,omitempty"` // Can be inline keyboard or reply keyboard
	MessageType           string `json:"message_type"` // "text", "photo", "document", etc.
	FileURL               string `json:"file_url,omitempty"`
	Caption               string `json:"caption,omitempty"`
}

// TelegramInlineKeyboardButton represents a button in an inline keyboard
type TelegramInlineKeyboardButton struct {
	Text              string `json:"text"`
	URL               string `json:"url,omitempty"`
	CallbackData      string `json:"callback_data,omitempty"`
	SwitchInlineQuery string `json:"switch_inline_query,omitempty"`
}

// TelegramNode represents a Telegram integration node
type TelegramNode struct {
	config *TelegramNodeConfig
}

// NewTelegramNode creates a new Telegram node
func NewTelegramNode(config *TelegramNodeConfig) *TelegramNode {
	return &TelegramNode{
		config: config,
	}
}

// Execute executes the Telegram operation
func (tn *TelegramNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Override config values with inputs if provided
	token := tn.config.Token
	if tok, exists := inputs["telegram_token"]; exists {
		if tokStr, ok := tok.(string); ok {
			token = tokStr
		}
	}

	chatID := tn.config.ChatID
	if id, exists := inputs["chat_id"]; exists {
		if idStr, ok := id.(string); ok {
			chatID = idStr
		}
	}

	messageText := tn.config.MessageText
	if text, exists := inputs["message_text"]; exists {
		if textStr, ok := text.(string); ok {
			messageText = textStr
		}
	}

	parseMode := tn.config.ParseMode
	if mode, exists := inputs["parse_mode"]; exists {
		if modeStr, ok := mode.(string); ok {
			parseMode = modeStr
		}
	}

	disableWebPagePreview := tn.config.DisableWebPagePreview
	if disable, exists := inputs["disable_web_page_preview"]; exists {
		if disableBool, ok := disable.(bool); ok {
			disableWebPagePreview = disableBool
		}
	}

	disableNotification := tn.config.DisableNotification
	if disable, exists := inputs["disable_notification"]; exists {
		if disableBool, ok := disable.(bool); ok {
			disableNotification = disableBool
		}
	}

	messageType := tn.config.MessageType
	if typ, exists := inputs["message_type"]; exists {
		if typStr, ok := typ.(string); ok {
			messageType = typStr
		}
	}

	fileURL := tn.config.FileURL
	if url, exists := inputs["file_url"]; exists {
		if urlStr, ok := url.(string); ok {
			fileURL = urlStr
		}
	}

	caption := tn.config.Caption
	if cap, exists := inputs["caption"]; exists {
		if capStr, ok := cap.(string); ok {
			caption = capStr
		}
	}

	// Validate required fields
	if token == "" {
		return nil, fmt.Errorf("telegram token is required")
	}

	if chatID == "" {
		return nil, fmt.Errorf("chat ID is required")
	}

	var result map[string]interface{}
	var err error

	// Send message based on type
	switch messageType {
	case "photo", "image":
		result, err = tn.sendPhoto(ctx, token, chatID, fileURL, messageText, caption, parseMode, disableNotification)
	case "document":
		result, err = tn.sendDocument(ctx, token, chatID, fileURL, messageText, caption, parseMode, disableNotification)
	case "audio":
		result, err = tn.sendAudio(ctx, token, chatID, fileURL, messageText, caption, parseMode, disableNotification)
	case "video":
		result, err = tn.sendVideo(ctx, token, chatID, fileURL, messageText, caption, parseMode, disableNotification)
	default:
		result, err = tn.sendMessage(ctx, token, chatID, messageText, parseMode, disableWebPagePreview, disableNotification)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to send Telegram message: %w", err)
	}

	return result, nil
}

// sendMessage sends a text message via Telegram API
func (tn *TelegramNode) sendMessage(ctx context.Context, token, chatID, text, parseMode string, disableWebPagePreview, disableNotification bool) (map[string]interface{}, error) {
	// Prepare the API URL
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)

	// Prepare form data
	formData := url.Values{}
	formData.Set("chat_id", chatID)
	formData.Set("text", text)

	if parseMode != "" {
		formData.Set("parse_mode", parseMode)
	}

	if disableWebPagePreview {
		formData.Set("disable_web_page_preview", "true")
	}

	if disableNotification {
		formData.Set("disable_notification", "true")
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
		"method":      "sendMessage",
		"timestamp":   time.Now().Unix(),
		"chat_id":     chatID,
	}

	if !success {
		result["error"] = fmt.Sprintf("Telegram API returned error: %v", apiResp["description"])
	}

	return result, nil
}

// sendPhoto sends a photo via Telegram API
func (tn *TelegramNode) sendPhoto(ctx context.Context, token, chatID, photoURL, text, caption, parseMode string, disableNotification bool) (map[string]interface{}, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", token)

	// Prepare form data
	formData := url.Values{}
	formData.Set("chat_id", chatID)
	if caption != "" {
		formData.Set("caption", caption)
	}
	if parseMode != "" {
		formData.Set("parse_mode", parseMode)
	}
	if disableNotification {
		formData.Set("disable_notification", "true")
	}

	// If photo is a URL, download it first and send as form data
	if strings.HasPrefix(photoURL, "http://") || strings.HasPrefix(photoURL, "https://") {
		// For URL, we can pass the URL directly
		formData.Set("photo", photoURL)
	} else {
		return nil, fmt.Errorf("only URL-based photos are supported in this implementation")
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute the request
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send photo: %w", err)
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
		"method":      "sendPhoto",
		"timestamp":   time.Now().Unix(),
		"chat_id":     chatID,
	}

	if !success {
		result["error"] = fmt.Sprintf("Telegram API returned error: %v", apiResp["description"])
	}

	return result, nil
}

// sendDocument sends a document via Telegram API
func (tn *TelegramNode) sendDocument(ctx context.Context, token, chatID, documentURL, text, caption, parseMode string, disableNotification bool) (map[string]interface{}, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendDocument", token)

	// Prepare form data
	formData := url.Values{}
	formData.Set("chat_id", chatID)
	if caption != "" {
		formData.Set("caption", caption)
	}
	if parseMode != "" {
		formData.Set("parse_mode", parseMode)
	}
	if disableNotification {
		formData.Set("disable_notification", "true")
	}

	// If document is a URL, pass it directly
	if strings.HasPrefix(documentURL, "http://") || strings.HasPrefix(documentURL, "https://") {
		formData.Set("document", documentURL)
	} else {
		return nil, fmt.Errorf("only URL-based documents are supported in this implementation")
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute the request
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send document: %w", err)
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
		"method":      "sendDocument",
		"timestamp":   time.Now().Unix(),
		"chat_id":     chatID,
	}

	if !success {
		result["error"] = fmt.Sprintf("Telegram API returned error: %v", apiResp["description"])
	}

	return result, nil
}

// sendVideo sends a video via Telegram API
func (tn *TelegramNode) sendVideo(ctx context.Context, token, chatID, videoURL, text, caption, parseMode string, disableNotification bool) (map[string]interface{}, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendVideo", token)

	// Prepare form data
	formData := url.Values{}
	formData.Set("chat_id", chatID)
	if caption != "" {
		formData.Set("caption", caption)
	}
	if parseMode != "" {
		formData.Set("parse_mode", parseMode)
	}
	if disableNotification {
		formData.Set("disable_notification", "true")
	}

	// If video is a URL, pass it directly
	if strings.HasPrefix(videoURL, "http://") || strings.HasPrefix(videoURL, "https://") {
		formData.Set("video", videoURL)
	} else {
		return nil, fmt.Errorf("only URL-based videos are supported in this implementation")
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute the request
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send video: %w", err)
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
		"method":      "sendVideo",
		"timestamp":   time.Now().Unix(),
		"chat_id":     chatID,
	}

	if !success {
		result["error"] = fmt.Sprintf("Telegram API returned error: %v", apiResp["description"])
	}

	return result, nil
}

// sendAudio sends an audio file via Telegram API
func (tn *TelegramNode) sendAudio(ctx context.Context, token, chatID, audioURL, text, caption, parseMode string, disableNotification bool) (map[string]interface{}, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendAudio", token)

	// Prepare form data
	formData := url.Values{}
	formData.Set("chat_id", chatID)
	if caption != "" {
		formData.Set("caption", caption)
	}
	if parseMode != "" {
		formData.Set("parse_mode", parseMode)
	}
	if disableNotification {
		formData.Set("disable_notification", "true")
	}

	// If audio is a URL, pass it directly
	if strings.HasPrefix(audioURL, "http://") || strings.HasPrefix(audioURL, "https://") {
		formData.Set("audio", audioURL)
	} else {
		return nil, fmt.Errorf("only URL-based audio is supported in this implementation")
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute the request
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send audio: %w", err)
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
		"method":      "sendAudio",
		"timestamp":   time.Now().Unix(),
		"chat_id":     chatID,
	}

	if !success {
		result["error"] = fmt.Sprintf("Telegram API returned error: %v", apiResp["description"])
	}

	return result, nil
}

// RegisterTelegramNode registers the Telegram node type with the engine
func RegisterTelegramNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("telegram_integration", func(config map[string]interface{}) (engine.NodeInstance, error) {
		var token string
		if tok, exists := config["telegram_token"]; exists {
			if tokStr, ok := tok.(string); ok {
				token = tokStr
			}
		}

		var chatID string
		if id, exists := config["chat_id"]; exists {
			if idStr, ok := id.(string); ok {
				chatID = idStr
			}
		}

		var messageText string
		if text, exists := config["message_text"]; exists {
			if textStr, ok := text.(string); ok {
				messageText = textStr
			}
		}

		var parseMode string
		if mode, exists := config["parse_mode"]; exists {
			if modeStr, ok := mode.(string); ok {
				parseMode = modeStr
			}
		}

		var disableWebPagePreview bool
		if disable, exists := config["disable_web_page_preview"]; exists {
			if disableBool, ok := disable.(bool); ok {
				disableWebPagePreview = disableBool
			}
		}

		var disableNotification bool
		if disable, exists := config["disable_notification"]; exists {
			if disableBool, ok := disable.(bool); ok {
				disableNotification = disableBool
			}
		}

		var messageType string
		if typ, exists := config["message_type"]; exists {
			if typStr, ok := typ.(string); ok {
				messageType = typStr
			}
		}

		var fileURL string
		if url, exists := config["file_url"]; exists {
			if urlStr, ok := url.(string); ok {
				fileURL = urlStr
			}
		}

		var caption string
		if cap, exists := config["caption"]; exists {
			if capStr, ok := cap.(string); ok {
				caption = capStr
			}
		}

		// Convert reply_to_message_id if exists
		var replyToMessageID *int
		if id, exists := config["reply_to_message_id"]; exists {
			if idFloat, ok := id.(float64); ok {
				intVal := int(idFloat)
				replyToMessageID = &intVal
			}
		}

		nodeConfig := &TelegramNodeConfig{
			Token:       token,
			ChatID:      chatID,
			MessageText: messageText,
			ParseMode:   parseMode,
			DisableWebPagePreview: disableWebPagePreview,
			DisableNotification:   disableNotification,
			ReplyToMessageID:      replyToMessageID,
			MessageType:           messageType,
			FileURL:               fileURL,
			Caption:               caption,
		}

		return NewTelegramNode(nodeConfig), nil
	})
}