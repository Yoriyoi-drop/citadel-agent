package communication

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

var (
	ErrInvalidBotToken = errors.New("invalid bot token")
	ErrInvalidChatID   = errors.New("invalid chat ID")
	ErrTelegramAPI     = errors.New("telegram API error")
)

// TelegramClient handles Telegram Bot API operations
type TelegramClient struct {
	botToken   string
	apiURL     string
	httpClient *http.Client
}

// TelegramConfig holds Telegram configuration
type TelegramConfig struct {
	BotToken string
}

// TelegramMessage represents a Telegram message
type TelegramMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"` // Markdown, HTML, MarkdownV2
}

// TelegramResponse represents Telegram API response
type TelegramResponse struct {
	OK          bool            `json:"ok"`
	Result      json.RawMessage `json:"result,omitempty"`
	Description string          `json:"description,omitempty"`
}

// NewTelegramClient creates a new Telegram client
func NewTelegramClient(config TelegramConfig) (*TelegramClient, error) {
	if config.BotToken == "" {
		return nil, ErrInvalidBotToken
	}

	return &TelegramClient{
		botToken:   config.BotToken,
		apiURL:     fmt.Sprintf("https://api.telegram.org/bot%s", config.BotToken),
		httpClient: &http.Client{},
	}, nil
}

// SendMessage sends a text message
func (c *TelegramClient) SendMessage(chatID string, text string) error {
	return c.SendMessageWithOptions(chatID, text, "")
}

// SendMessageWithOptions sends a message with parse mode
func (c *TelegramClient) SendMessageWithOptions(chatID string, text string, parseMode string) error {
	payload := TelegramMessage{
		ChatID:    chatID,
		Text:      text,
		ParseMode: parseMode,
	}

	return c.callAPI("sendMessage", payload)
}

// SendPhoto sends a photo
func (c *TelegramClient) SendPhoto(chatID string, photoPath string, caption string) error {
	file, err := os.Open(photoPath)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add chat_id
	if err := writer.WriteField("chat_id", chatID); err != nil {
		return err
	}

	// Add caption if provided
	if caption != "" {
		if err := writer.WriteField("caption", caption); err != nil {
			return err
		}
	}

	// Add photo
	part, err := writer.CreateFormFile("photo", photoPath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, file); err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	return c.callAPIMultipart("sendPhoto", body, writer.FormDataContentType())
}

// SendDocument sends a document
func (c *TelegramClient) SendDocument(chatID string, documentPath string, caption string) error {
	file, err := os.Open(documentPath)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add chat_id
	if err := writer.WriteField("chat_id", chatID); err != nil {
		return err
	}

	// Add caption if provided
	if caption != "" {
		if err := writer.WriteField("caption", caption); err != nil {
			return err
		}
	}

	// Add document
	part, err := writer.CreateFormFile("document", documentPath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, file); err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	return c.callAPIMultipart("sendDocument", body, writer.FormDataContentType())
}

// SendLocation sends a location
func (c *TelegramClient) SendLocation(chatID string, latitude, longitude float64) error {
	payload := map[string]interface{}{
		"chat_id":   chatID,
		"latitude":  latitude,
		"longitude": longitude,
	}

	return c.callAPI("sendLocation", payload)
}

// SendInlineKeyboard sends a message with inline keyboard
func (c *TelegramClient) SendInlineKeyboard(chatID string, text string, buttons [][]InlineKeyboardButton) error {
	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
		"reply_markup": map[string]interface{}{
			"inline_keyboard": buttons,
		},
	}

	return c.callAPI("sendMessage", payload)
}

// InlineKeyboardButton represents an inline keyboard button
type InlineKeyboardButton struct {
	Text         string `json:"text"`
	URL          string `json:"url,omitempty"`
	CallbackData string `json:"callback_data,omitempty"`
}

// DeleteMessage deletes a message
func (c *TelegramClient) DeleteMessage(chatID string, messageID int) error {
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"message_id": messageID,
	}

	return c.callAPI("deleteMessage", payload)
}

// EditMessage edits a message
func (c *TelegramClient) EditMessage(chatID string, messageID int, newText string) error {
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"message_id": messageID,
		"text":       newText,
	}

	return c.callAPI("editMessageText", payload)
}

// GetMe gets bot information
func (c *TelegramClient) GetMe() (map[string]interface{}, error) {
	var response TelegramResponse
	if err := c.callAPIWithResponse("getMe", nil, &response); err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// callAPI makes an API call with JSON payload
func (c *TelegramClient) callAPI(method string, payload interface{}) error {
	return c.callAPIWithResponse(method, payload, nil)
}

// callAPIWithResponse makes an API call and returns response
func (c *TelegramClient) callAPIWithResponse(method string, payload interface{}, response *TelegramResponse) error {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewBuffer(jsonData)
	}

	url := fmt.Sprintf("%s/%s", c.apiURL, method)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp TelegramResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return err
	}

	if !apiResp.OK {
		return fmt.Errorf("%w: %s", ErrTelegramAPI, apiResp.Description)
	}

	if response != nil {
		*response = apiResp
	}

	return nil
}

// callAPIMultipart makes an API call with multipart data
func (c *TelegramClient) callAPIMultipart(method string, body *bytes.Buffer, contentType string) error {
	url := fmt.Sprintf("%s/%s", c.apiURL, method)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", contentType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp TelegramResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return err
	}

	if !apiResp.OK {
		return fmt.Errorf("%w: %s", ErrTelegramAPI, apiResp.Description)
	}

	return nil
}
