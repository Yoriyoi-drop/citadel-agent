package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/workflow/core/engine"
)

// NotificationConfig mewakili konfigurasi untuk node Notification
type NotificationConfig struct {
	Channel         string                   `json:"channel"`           // Channel notifikasi (email, sms, push, slack, etc.)
	Recipient       string                   `json:"recipient"`         // Penerima notifikasi
	Subject         string                   `json:"subject"`           // Subjek pesan (terutama untuk email)
	Message         string                   `json:"message"`           // Isi pesan
	Template        string                   `json:"template"`          // Template pesan
	Attachments     []string                 `json:"attachments"`       // Lampiran
	Priority        string                   `json:"priority"`          // Prioritas (low, normal, high, urgent)
	ScheduleTime    *int64                   `json:"schedule_time"`     // Waktu penjadwalan (jika dijadwalkan)
	EnableTracking  bool                     `json:"enable_tracking"`   // Apakah mengaktifkan pelacakan
	TrackOpens      bool                     `json:"track_opens"`       // Apakah melacak pembukaan
	TrackClicks     bool                     `json:"track_clicks"`      // Apakah melacak klik
	RetryAttempts   int                      `json:"retry_attempts"`    // Jumlah percobaan ulang
	RetryInterval   int                      `json:"retry_interval"`    // Interval retry dalam detik
	Timeout         int                      `json:"timeout"`           // Waktu timeout dalam detik
	EnableCaching   bool                     `json:"enable_caching"`    // Apakah mengaktifkan caching
	CacheTTL        int                      `json:"cache_ttl"`         // Waktu cache dalam detik
	EnableProfiling bool                     `json:"enable_profiling"`  // Apakah mengaktifkan profiling
	ReturnRawResults bool                    `json:"return_raw_results"` // Apakah mengembalikan hasil mentah
	CustomParams    map[string]interface{}   `json:"custom_params"`     // Parameter khusus untuk notifikasi
	Preprocessing   PreprocessingConfig      `json:"preprocessing"`     // Konfigurasi pra-pemrosesan
	Postprocessing  PostprocessingConfig     `json:"postprocessing"`    // Konfigurasi pasca-pemrosesan
	ChannelConfig   map[string]interface{}   `json:"channel_config"`    // Konfigurasi khusus channel
}

// PreprocessingConfig mewakili konfigurasi pra-pemrosesan
type PreprocessingConfig struct {
	NormalizeInput bool                   `json:"normalize_input"` // Apakah menormalkan input
	ValidateInput  bool                   `json:"validate_input"`  // Apakah memvalidasi input
	TransformInput bool                   `json:"transform_input"` // Apakah mentransformasi input
	TransformRules map[string]interface{} `json:"transform_rules"` // Aturan transformasi
}

// PostprocessingConfig mewakili konfigurasi pasca-pemrosesan
type PostprocessingConfig struct {
	FilterOutput  bool              `json:"filter_output"`   // Apakah memfilter output
	OutputMapping map[string]string `json:"output_mapping"`  // Pemetaan field output
	TransformOutput bool            `json:"transform_output"` // Apakah mentransformasi output
}

// NotificationNode mewakili node yang mengirim notifikasi
type NotificationNode struct {
	config *NotificationConfig
}

// NewNotificationNode membuat node Notification baru
func NewNotificationNode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Konversi map interface{} ke JSON lalu ke struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("gagal mengubah konfig menjadi json: %v", err)
	}

	var notificationConfig NotificationConfig
	err = json.Unmarshal(jsonData, &notificationConfig)
	if err != nil {
		return nil, fmt.Errorf("gagal menguraikan konfig: %v", err)
	}

	// Validasi dan atur default
	if notificationConfig.Channel == "" {
		notificationConfig.Channel = "email"
	}

	if notificationConfig.Priority == "" {
		notificationConfig.Priority = "normal"
	}

	if notificationConfig.RetryAttempts == 0 {
		notificationConfig.RetryAttempts = 3
	}

	if notificationConfig.RetryInterval == 0 {
		notificationConfig.RetryInterval = 30 // default 30 detik
	}

	if notificationConfig.Timeout == 0 {
		notificationConfig.Timeout = 60 // default timeout 60 detik
	}

	if notificationConfig.ChannelConfig == nil {
		notificationConfig.ChannelConfig = make(map[string]interface{})
	}

	return &NotificationNode{
		config: &notificationConfig,
	}, nil
}

// Execute mengimplementasikan interface NodeInstance
func (n *NotificationNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Timpa konfigurasi dengan nilai input jika disediakan
	channel := n.config.Channel
	if inputChannel, ok := input["channel"].(string); ok && inputChannel != "" {
		channel = inputChannel
	}

	recipient := n.config.Recipient
	if inputRecipient, ok := input["recipient"].(string); ok && inputRecipient != "" {
		recipient = inputRecipient
	}

	subject := n.config.Subject
	if inputSubject, ok := input["subject"].(string); ok && inputSubject != "" {
		subject = inputSubject
	}

	message := n.config.Message
	if inputMessage, ok := input["message"].(string); ok && inputMessage != "" {
		message = inputMessage
	}

	template := n.config.Template
	if inputTemplate, ok := input["template"].(string); ok && inputTemplate != "" {
		template = inputTemplate
	}

	attachments := n.config.Attachments
	if inputAttachments, ok := input["attachments"].([]interface{}); ok {
		attachments = make([]string, len(inputAttachments))
		for i, val := range inputAttachments {
			if valStr, ok := val.(string); ok {
				attachments[i] = valStr
			}
		}
	}

	priority := n.config.Priority
	if inputPriority, ok := input["priority"].(string); ok && inputPriority != "" {
		priority = inputPriority
	}

	enableTracking := n.config.EnableTracking
	if inputEnableTracking, ok := input["enable_tracking"].(bool); ok {
		enableTracking = inputEnableTracking
	}

	trackOpens := n.config.TrackOpens
	if inputTrackOpens, ok := input["track_opens"].(bool); ok {
		trackOpens = inputTrackOpens
	}

	trackClicks := n.config.TrackClicks
	if inputTrackClicks, ok := input["track_clicks"].(bool); ok {
		trackClicks = inputTrackClicks
	}

	retryAttempts := n.config.RetryAttempts
	if inputRetryAttempts, ok := input["retry_attempts"].(float64); ok {
		retryAttempts = int(inputRetryAttempts)
	}

	retryInterval := n.config.RetryInterval
	if inputRetryInterval, ok := input["retry_interval"].(float64); ok {
		retryInterval = int(inputRetryInterval)
	}

	timeout := n.config.Timeout
	if inputTimeout, ok := input["timeout"].(float64); ok {
		timeout = int(inputTimeout)
	}

	enableCaching := n.config.EnableCaching
	if inputEnableCaching, ok := input["enable_caching"].(bool); ok {
		enableCaching = inputEnableCaching
	}

	cacheTTL := n.config.CacheTTL
	if inputCacheTTL, ok := input["cache_ttl"].(float64); ok {
		cacheTTL = int(inputCacheTTL)
	}

	enableProfiling := n.config.EnableProfiling
	if inputEnableProfiling, ok := input["enable_profiling"].(bool); ok {
		enableProfiling = inputEnableProfiling
	}

	returnRawResults := n.config.ReturnRawResults
	if inputReturnRaw, ok := input["return_raw_results"].(bool); ok {
		returnRawResults = inputReturnRaw
	}

	customParams := n.config.CustomParams
	if inputCustomParams, ok := input["custom_params"].(map[string]interface{}); ok {
		customParams = inputCustomParams
	}

	channelConfig := n.config.ChannelConfig
	if inputChannelConfig, ok := input["channel_config"].(map[string]interface{}); ok {
		channelConfig = inputChannelConfig
	}

	// Validasi input yang diperlukan
	if recipient == "" {
		if rec, exists := input["recipient"]; exists {
			if recStr, ok := rec.(string); ok {
				recipient = recStr
			}
		}
	}

	if recipient == "" {
		return map[string]interface{}{
			"success":   false,
			"error":     "recipient diperlukan untuk mengirim notifikasi",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	if message == "" && subject == "" && template == "" {
		if msg, exists := input["message"]; exists {
			if msgStr, ok := msg.(string); ok {
				message = msgStr
			}
		}
		if message == "" {
			if subj, exists := input["subject"]; exists {
				if subjStr, ok := subj.(string); ok {
					subject = subjStr
				}
			}
			if subject != "" && message == "" {
				message = subject // Gunakan subject sebagai message jika message kosong
			}
		}
	}

	if message == "" && template == "" {
		return map[string]interface{}{
			"success":   false,
			"error":     "message atau template diperlukan untuk mengirim notifikasi",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Buat konteks operasi dengan timeout
	notificationCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Kirim notifikasi
	notificationResult, err := n.sendNotification(notificationCtx, channel, recipient, input)
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
		"channel":              channel,
		"recipient":            recipient,
		"subject":              subject,
		"message_preview":      truncateString(message, 100), // Tampilkan preview pesan
		"template":             template,
		"attachments_count":    len(attachments),
		"priority":             priority,
		"enable_tracking":      enableTracking,
		"track_opens":          trackOpens,
		"track_clicks":         trackClicks,
		"retry_attempts":       retryAttempts,
		"retry_interval":       retryInterval,
		"notification_result":  notificationResult,
		"enable_caching":       enableCaching,
		"enable_profiling":     enableProfiling,
		"return_raw_results":   returnRawResults,
		"timestamp":            time.Now().Unix(),
		"input_data":           input,
		"config":               n.config,
		"channel_config":       channelConfig,
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

// sendNotification mengirim notifikasi ke channel yang ditentukan
func (n *NotificationNode) sendNotification(ctx context.Context, channel, recipient string, input map[string]interface{}) (map[string]interface{}, error) {
	// Simulasikan waktu pemrosesan
	time.Sleep(100 * time.Millisecond)

	// Dapatkan pesan dari input atau konfigurasi
	message := n.config.Message
	if inputMessage, exists := input["message"]; exists {
		if msgStr, ok := inputMessage.(string); ok {
			message = msgStr
		}
	} else if n.config.Template != "" {
		// Jika template disediakan, gunakan template
		message = n.processTemplate(n.config.Template, input)
	} else {
		// Gunakan subject sebagai fallback
		message = n.config.Subject
	}

	// Simulasikan hasil pengiriman berdasarkan channel
	var deliveryStatus string
	var deliveryId string
	var trackingEnabled bool

	switch channel {
	case "email":
		deliveryStatus = "delivered"
		deliveryId = fmt.Sprintf("email_%d_%s", time.Now().UnixNano(), recipient)
		trackingEnabled = n.config.EnableTracking
	case "sms":
		deliveryStatus = "sent"
		deliveryId = fmt.Sprintf("sms_%d_%s", time.Now().UnixNano(), recipient)
		trackingEnabled = false // SMS biasanya tidak bisa dilacak secara rinci
	case "push":
		deliveryStatus = "delivered"
		deliveryId = fmt.Sprintf("push_%d_%s", time.Now().UnixNano(), recipient)
		trackingEnabled = true
	case "slack":
		deliveryStatus = "posted"
		deliveryId = fmt.Sprintf("slack_%d_%s", time.Now().UnixNano(), recipient)
		trackingEnabled = false
	default:
		deliveryStatus = "delivered"
		deliveryId = fmt.Sprintf("%s_%d_%s", channel, time.Now().UnixNano(), recipient)
		trackingEnabled = n.config.EnableTracking
	}

	result := map[string]interface{}{
		"channel_used":    channel,
		"recipient":      recipient,
		"message_sent":   true,
		"delivery_status": deliveryStatus,
		"delivery_id":    deliveryId,
		"message":        message,
		"timestamp":      time.Now().Unix(),
		"processing_time": time.Since(time.Now().Add(-100 * time.Millisecond)).Seconds(),
		"tracking_enabled": trackingEnabled,
		"tracking_available": map[string]bool{
			"opens":  n.config.TrackOpens,
			"clicks": n.config.TrackClicks,
		},
		"retry_info": map[string]interface{}{
			"attempts": n.config.RetryAttempts,
			"interval": n.config.RetryInterval,
		},
		"notification_type": "standard",
		"priority":          n.config.Priority,
		"provider":          fmt.Sprintf("%s_provider", channel),
		"delivery_receipt":   deliveryId,
	}

	// Simulasikan kemungkinan kegagalan berdasarkan channel
	if channel == "sms" && recipient == "invalid_number" {
		result["delivery_status"] = "failed"
		result["error"] = "Invalid phone number"
	}

	return result, nil
}

// processTemplate memproses template dengan data input
func (n *NotificationNode) processTemplate(template string, input map[string]interface{}) string {
	// Dalam implementasi nyata, ini akan menggunakan template engine seperti Go's text/template
	// Untuk simulasi, kita hanya akan mengganti placeholder sederhana
	
	result := template
	
	// Ganti placeholder dengan nilai dari input
	for key, value := range input {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = replaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	
	// Ganti placeholder umum
	result = replaceAll(result, "{{timestamp}}", fmt.Sprintf("%d", time.Now().Unix()))
	result = replaceAll(result, "{{date}}", time.Now().Format("2006-01-02"))
	
	return result
}

// replaceAll adalah fungsi helper untuk mengganti semua kemunculan string
func replaceAll(str, old, new string) string {
	// Fungsi sederhana untuk mengganti semua kemunculan old dengan new
	result := ""
	i := 0
	for i < len(str) {
		if i <= len(str)-len(old) && str[i:i+len(old)] == old {
			result += new
			i += len(old)
		} else {
			result += string(str[i])
			i++
		}
	}
	return result
}

// truncateString memotong string ke panjang maksimum
func truncateString(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen] + "..."
}

// GetType mengembalikan jenis node
func (n *NotificationNode) GetType() string {
	return "notification"
}

// GetID mengembalikan ID unik untuk instance node
func (n *NotificationNode) GetID() string {
	return "notification_" + fmt.Sprintf("%d", time.Now().Unix())
}

// RegisterNotificationNode mendaftarkan node Notification dengan engine
func RegisterNotificationNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("notification", func(config map[string]interface{}) (engine.NodeInstance, error) {
		return NewNotificationNode(config)
	})
}