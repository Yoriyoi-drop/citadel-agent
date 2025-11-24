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
	ErrInvalidToken   = errors.New("invalid bot token")
	ErrInvalidChannel = errors.New("invalid channel ID")
	ErrSendFailed     = errors.New("failed to send message")
)

// DiscordClient handles Discord API operations
type DiscordClient struct {
	botToken   string
	webhookURL string
	httpClient *http.Client
}

// DiscordConfig holds Discord configuration
type DiscordConfig struct {
	BotToken   string
	WebhookURL string
}

// DiscordMessage represents a Discord message
type DiscordMessage struct {
	Content   string         `json:"content,omitempty"`
	Username  string         `json:"username,omitempty"`
	AvatarURL string         `json:"avatar_url,omitempty"`
	Embeds    []DiscordEmbed `json:"embeds,omitempty"`
}

// DiscordEmbed represents a Discord embed
type DiscordEmbed struct {
	Title       string              `json:"title,omitempty"`
	Description string              `json:"description,omitempty"`
	URL         string              `json:"url,omitempty"`
	Color       int                 `json:"color,omitempty"`
	Fields      []DiscordEmbedField `json:"fields,omitempty"`
	Footer      *DiscordEmbedFooter `json:"footer,omitempty"`
	Timestamp   string              `json:"timestamp,omitempty"`
}

// DiscordEmbedField represents an embed field
type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// DiscordEmbedFooter represents an embed footer
type DiscordEmbedFooter struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url,omitempty"`
}

// NewDiscordClient creates a new Discord client
func NewDiscordClient(config DiscordConfig) (*DiscordClient, error) {
	if config.BotToken == "" && config.WebhookURL == "" {
		return nil, errors.New("either bot token or webhook URL is required")
	}

	return &DiscordClient{
		botToken:   config.BotToken,
		webhookURL: config.WebhookURL,
		httpClient: &http.Client{},
	}, nil
}

// SendMessage sends a message to a Discord channel
func (c *DiscordClient) SendMessage(channelID string, content string) error {
	if c.botToken == "" {
		return ErrInvalidToken
	}

	payload := map[string]interface{}{
		"content": content,
	}

	return c.sendToChannel(channelID, payload)
}

// SendEmbed sends an embed message to a Discord channel
func (c *DiscordClient) SendEmbed(channelID string, embed DiscordEmbed) error {
	if c.botToken == "" {
		return ErrInvalidToken
	}

	payload := map[string]interface{}{
		"embeds": []DiscordEmbed{embed},
	}

	return c.sendToChannel(channelID, payload)
}

// SendWebhook sends a message via webhook
func (c *DiscordClient) SendWebhook(message DiscordMessage) error {
	if c.webhookURL == "" {
		return errors.New("webhook URL not configured")
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%w: %s", ErrSendFailed, string(body))
	}

	return nil
}

// SendFile sends a file to a Discord channel
func (c *DiscordClient) SendFile(channelID string, filePath string, comment string) error {
	if c.botToken == "" {
		return ErrInvalidToken
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	part, err := writer.CreateFormFile("file", filePath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, file); err != nil {
		return err
	}

	// Add comment if provided
	if comment != "" {
		if err := writer.WriteField("content", comment); err != nil {
			return err
		}
	}

	if err := writer.Close(); err != nil {
		return err
	}

	url := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", channelID)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bot "+c.botToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%w: %s", ErrSendFailed, string(respBody))
	}

	return nil
}

// sendToChannel sends a payload to a Discord channel
func (c *DiscordClient) sendToChannel(channelID string, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", channelID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bot "+c.botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%w: %s", ErrSendFailed, string(body))
	}

	return nil
}

// CreateEmbed creates a Discord embed
func CreateEmbed(title, description string, color int) DiscordEmbed {
	return DiscordEmbed{
		Title:       title,
		Description: description,
		Color:       color,
	}
}

// AddField adds a field to an embed
func (e *DiscordEmbed) AddField(name, value string, inline bool) {
	e.Fields = append(e.Fields, DiscordEmbedField{
		Name:   name,
		Value:  value,
		Inline: inline,
	})
}

// SetFooter sets the footer of an embed
func (e *DiscordEmbed) SetFooter(text string, iconURL string) {
	e.Footer = &DiscordEmbedFooter{
		Text:    text,
		IconURL: iconURL,
	}
}
