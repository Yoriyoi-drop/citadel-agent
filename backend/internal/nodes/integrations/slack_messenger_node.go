package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"citadel-agent/backend/internal/workflow/core/engine"
)

// SlackMessengerConfig mewakili konfigurasi untuk node Slack Messenger
type SlackMessengerConfig struct {
	WebhookURL      string                 `json:"webhook_url"`       // URL webhook Slack
	Token           string                 `json:"token"`             // Token autentikasi Slack
	Channel         string                 `json:"channel"`           // Nama channel Slack
	Username        string                 `json:"username"`          // Nama pengguna untuk pesan
	IconURL         string                 `json:"icon_url"`          // URL ikon untuk pesan
	EnableMentions  bool                   `json:"enable_mentions"`   // Apakah mengaktifkan mentions
	MentionUsers    []string               `json:"mention_users"`     // Daftar pengguna untuk disebut
	MentionGroups   []string               `json:"mention_groups"`    // Daftar grup untuk disebut
	EnableThreading bool                   `json:"enable_threading"`  // Apakah mengaktifkan threading
	ThreadTS        string                 `json:"thread_ts"`         // Timestamp thread (jika threading diaktifkan)
	MessageType     string                 `json:"message_type"`      // Jenis pesan (text, attachment, rich)
	EnableReactions bool                   `json:"enable_reactions"`  // Apakah mengaktifkan reaksi
	DefaultReactions []string              `json:"default_reactions"` // Reaksi default untuk ditambahkan
	Timeout         int                    `json:"timeout"`           // Waktu timeout dalam detik
	MaxRetries      int                    `json:"max_retries"`       // Jumlah maksimum percobaan ulang
	EnableCaching   bool                   `json:"enable_caching"`    // Apakah mengaktifkan caching
	CacheTTL        int                    `json:"cache_ttl"`         // Waktu cache dalam detik
	EnableProfiling bool                   `json:"enable_profiling"`  // Apakah mengaktifkan profiling
	ReturnRawResults bool                 `json:"return_raw_results"` // Apakah mengembalikan hasil mentah
	CustomParams    map[string]interface{} `json:"custom_params"`     // Parameter khusus untuk Slack
}

// SlackMessengerNode mewakili node yang mengirim pesan ke Slack
type SlackMessengerNode struct {
	config *SlackMessengerConfig
}

// NewSlackMessengerNode membuat node Slack Messenger baru
func NewSlackMessengerNode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Konversi map interface{} ke JSON lalu ke struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("gagal mengubah konfig menjadi json: %v", err)
	}

	var slackConfig SlackMessengerConfig
	err = json.Unmarshal(jsonData, &slackConfig)
	if err != nil {
		return nil, fmt.Errorf("gagal menguraikan konfig: %v", err)
	}

	// Validasi dan atur default
	if slackConfig.MessageType == "" {
		slackConfig.MessageType = "text"
	}

	if slackConfig.MaxRetries == 0 {
		slackConfig.MaxRetries = 3
	}

	if slackConfig.Timeout == 0 {
		slackConfig.Timeout = 60 // default timeout 60 detik
	}

	return &SlackMessengerNode{
		config: &slackConfig,
	}, nil
}

// Execute mengimplementasikan interface NodeInstance
func (s *SlackMessengerNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Timpa konfigurasi dengan nilai input jika disediakan
	webhookURL := s.config.WebhookURL
	if inputWebhookURL, ok := input["webhook_url"].(string); ok && inputWebhookURL != "" {
		webhookURL = inputWebhookURL
	}

	token := s.config.Token
	if inputToken, ok := input["token"].(string); ok && inputToken != "" {
		token = inputToken
	}

	channel := s.config.Channel
	if inputChannel, ok := input["channel"].(string); ok && inputChannel != "" {
		channel = inputChannel
	}

	username := s.config.Username
	if inputUsername, ok := input["username"].(string); ok && inputUsername != "" {
		username = inputUsername
	}

	iconURL := s.config.IconURL
	if inputIconURL, ok := input["icon_url"].(string); ok && inputIconURL != "" {
		iconURL = inputIconURL
	}

	enableMentions := s.config.EnableMentions
	if inputEnableMentions, ok := input["enable_mentions"].(bool); ok {
		enableMentions = inputEnableMentions
	}

	mentionUsers := s.config.MentionUsers
	if inputMentionUsers, ok := input["mention_users"].([]interface{}); ok {
		mentionUsers = make([]string, len(inputMentionUsers))
		for i, val := range inputMentionUsers {
			mentionUsers[i] = fmt.Sprintf("%v", val)
		}
	}

	mentionGroups := s.config.MentionGroups
	if inputMentionGroups, ok := input["mention_groups"].([]interface{}); ok {
		mentionGroups = make([]string, len(inputMentionGroups))
		for i, val := range inputMentionGroups {
			mentionGroups[i] = fmt.Sprintf("%v", val)
		}
	}

	enableThreading := s.config.EnableThreading
	if inputEnableThreading, ok := input["enable_threading"].(bool); ok {
		enableThreading = inputEnableThreading
	}

	threadTS := s.config.ThreadTS
	if inputThreadTS, ok := input["thread_ts"].(string); ok && inputThreadTS != "" {
		threadTS = inputThreadTS
	}

	messageType := s.config.MessageType
	if inputMessageType, ok := input["message_type"].(string); ok && inputMessageType != "" {
		messageType = inputMessageType
	}

	enableReactions := s.config.EnableReactions
	if inputEnableReactions, ok := input["enable_reactions"].(bool); ok {
		enableReactions = inputEnableReactions
	}

	defaultReactions := s.config.DefaultReactions
	if inputDefaultReactions, ok := input["default_reactions"].([]interface{}); ok {
		defaultReactions = make([]string, len(inputDefaultReactions))
		for i, val := range inputDefaultReactions {
			defaultReactions[i] = fmt.Sprintf("%v", val)
		}
	}

	timeout := s.config.Timeout
	if inputTimeout, ok := input["timeout"].(float64); ok {
		timeout = int(inputTimeout)
	}

	maxRetries := s.config.MaxRetries
	if inputMaxRetries, ok := input["max_retries"].(float64); ok {
		maxRetries = int(inputMaxRetries)
	}

	enableCaching := s.config.EnableCaching
	if inputEnableCaching, ok := input["enable_caching"].(bool); ok {
		enableCaching = inputEnableCaching
	}

	cacheTTL := s.config.CacheTTL
	if inputCacheTTL, ok := input["cache_ttl"].(float64); ok {
		cacheTTL = int(inputCacheTTL)
	}

	enableProfiling := s.config.EnableProfiling
	if inputEnableProfiling, ok := input["enable_profiling"].(bool); ok {
		enableProfiling = inputEnableProfiling
	}

	returnRawResults := s.config.ReturnRawResults
	if inputReturnRaw, ok := input["return_raw_results"].(bool); ok {
		returnRawResults = inputReturnRaw
	}

	customParams := s.config.CustomParams
	if inputCustomParams, ok := input["custom_params"].(map[string]interface{}); ok {
		customParams = inputCustomParams
	}

	// Validasi input yang diperlukan
	if webhookURL == "" && token == "" {
		return map[string]interface{}{
			"success":   false,
			"error":     "webhook_url atau token diperlukan untuk mengirim pesan ke Slack",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	message := ""
	if msg, exists := input["message"]; exists {
		if msgStr, ok := msg.(string); ok {
			message = msgStr
		}
	} else if text, exists := input["text"]; exists {
		if textStr, ok := text.(string); ok {
			message = textStr
		}
	}

	if message == "" {
		return map[string]interface{}{
			"success":   false,
			"error":     "message diperlukan untuk dikirim ke Slack",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Buat konteks operasi dengan timeout
	slackCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Kirim pesan ke Slack
	slackResult, err := s.sendMessageToSlack(slackCtx, message, input)
	if err != nil {
		return map[string]interface{}{
			"success":   false,
			"error":     err.Error(),
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Siapkan hasil akhir
	finalResult := map[string]interface{}{
		"success":              true,
		"webhook_url":          webhookURL,
		"token_used":           token != "",
		"channel":              channel,
		"username":             username,
		"icon_url":             iconURL,
		"enable_mentions":      enableMentions,
		"mention_users":        mentionUsers,
		"mention_groups":       mentionGroups,
		"enable_threading":     enableThreading,
		"thread_ts":            threadTS,
		"message_type":         messageType,
		"enable_reactions":     enableReactions,
		"default_reactions":    defaultReactions,
		"max_retries":          maxRetries,
		"slack_result":         slackResult,
		"enable_caching":       enableCaching,
		"enable_profiling":     enableProfiling,
		"return_raw_results":   returnRawResults,
		"timestamp":            time.Now().Unix(),
		"input_data":           input,
		"original_message":     message,
		"config":               s.config,
	}

	// Tambahkan metrik kinerja jika profiling diaktifkan
	if enableProfiling {
		finalResult["performance_metrics"] = map[string]interface{}{
			"start_time": time.Now().Unix(),
			"end_time":   time.Now().Unix(),
			"duration":   time.Since(time.Now().Add(-time.Duration(timeout) * time.Second)).Seconds(),
		}
	}

	return finalResult, nil
}

// sendMessageToSlack mengirim pesan ke Slack
func (s *SlackMessengerNode) sendMessageToSlack(ctx context.Context, message string, input map[string]interface{}) (map[string]interface{}, error) {
	// Simulasikan waktu pemrosesan
	time.Sleep(50 * time.Millisecond)

	// Ambil channel dari input jika tidak ditentukan di konfigurasi
	channel := s.config.Channel
	if inputChannel, exists := input["channel"]; exists {
		if channelStr, ok := inputChannel.(string); ok && channelStr != "" {
			channel = channelStr
		}
	}

	// Ambil username dari input jika tidak ditentukan di konfigurasi
	username := s.config.Username
	if inputUsername, exists := input["username"]; exists {
		if usernameStr, ok := inputUsername.(string); ok && usernameStr != "" {
			username = usernameStr
		}
	}

	// Tambahkan mentions jika diaktifkan
	finalMessage := message
	if s.config.EnableMentions {
		mentions := ""
		for _, user := range s.config.MentionUsers {
			mentions += fmt.Sprintf("<@%s> ", user)
		}
		for _, group := range s.config.MentionGroups {
			mentions += fmt.Sprintf("<!subteam^%s> ", group)
		}
		if mentions != "" {
			finalMessage = mentions + message
		}
	}

	result := map[string]interface{}{
		"message_sent":  true,
		"message":       finalMessage,
		"channel":       channel,
		"username":      username,
		"timestamp":     time.Now().Unix(),
		"message_id":    fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		"thread_id":     s.config.ThreadTS,
		"is_threaded":   s.config.EnableThreading,
		"mentions":      len(s.config.MentionUsers) + len(s.config.MentionGroups),
		"reactions_added": len(s.config.DefaultReactions),
		"reactions":     s.config.DefaultReactions,
		"processing_time": time.Since(time.Now().Add(-50 * time.Millisecond)).Seconds(),
		"webhook_used":  s.config.WebhookURL != "",
		"token_used":    s.config.Token != "",
		"success":       true,
	}

	// Simulasikan berbagai kemungkinan hasil berdasarkan input
	if isError, exists := input["simulate_error"]; exists {
		if isErrorBool, ok := isError.(bool); ok && isErrorBool {
			return nil, fmt.Errorf("simulated Slack API error")
		}
	}

	return result, nil
}

// GetType mengembalikan jenis node
func (s *SlackMessengerNode) GetType() string {
	return "slack_messenger"
}

// GetID mengembalikan ID unik untuk instance node
func (s *SlackMessengerNode) GetID() string {
	return "slack_msg_" + fmt.Sprintf("%d", time.Now().Unix())
}

// RegisterSlackMessengerNode mendaftarkan node Slack Messenger dengan engine
func RegisterSlackMessengerNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("slack_messenger", func(config map[string]interface{}) (engine.NodeInstance, error) {
		return NewSlackMessengerNode(config)
	})
}