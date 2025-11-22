package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/engine"
)

// TextToSpeechNodeConfig represents the configuration for a Text-to-Speech node
type TextToSpeechNodeConfig struct {
	Provider       string                 `json:"provider"`        // AI provider (openai, google, azure, etc.)
	ApiKey         string                 `json:"api_key"`         // API key for the TTS service
	Model          string                 `json:"model"`           // TTS model to use
	Text           string                 `json:"text"`            // Text to convert to speech
	Language       string                 `json:"language"`        // Language code (en, es, fr, etc.)
	Voice          string                 `json:"voice"`           // Voice to use (male, female, specific voice id)
	Speed          float64                `json:"speed"`           // Speech speed (0.5-2.0)
	Pitch          float64                `json:"pitch"`           // Speech pitch (0.5-2.0)
	Format         string                 `json:"format"`          // Audio format (mp3, wav, flac, etc.)
	SampleRate     int                    `json:"sample_rate"`     // Audio sample rate
	Bitrate        int                    `json:"bitrate"`         // Audio bitrate
	Emotion        string                 `json:"emotion"`         // Emotion to apply to speech (neutral, happy, sad, etc.)
	WordBreak      bool                   `json:"word_break"`      // Whether to handle word breaks
	CustomParams   map[string]interface{} `json:"custom_params"`   // Custom parameters for the TTS service
	Timeout        int                    `json:"timeout"`         // Request timeout in seconds
	Enabled        bool                   `json:"enabled"`         // Whether the node is enabled
}

// TextToSpeechNode represents a node that converts text to speech using AI services
type TextToSpeechNode struct {
	config *TextToSpeechNodeConfig
}

// NewTextToSpeechNode creates a new Text-to-Speech node
func NewTextToSpeechNode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Convert interface{} map to JSON and back to struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}

	var ttsConfig TextToSpeechNodeConfig
	err = json.Unmarshal(jsonData, &ttsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// Validate required fields
	if ttsConfig.Provider == "" {
		ttsConfig.Provider = "openai" // default provider
	}

	if ttsConfig.Model == "" {
		ttsConfig.Model = "tts-1" // default model
	}

	if ttsConfig.Language == "" {
		ttsConfig.Language = "en" // default language
	}

	if ttsConfig.Voice == "" {
		ttsConfig.Voice = "alloy" // default voice
	}

	if ttsConfig.Timeout == 0 {
		ttsConfig.Timeout = 30 // default timeout of 30 seconds
	}

	if ttsConfig.Speed == 0 {
		ttsConfig.Speed = 1.0 // default speed
	}

	if ttsConfig.Pitch == 0 {
		ttsConfig.Pitch = 1.0 // default pitch
	}

	return &TextToSpeechNode{
		config: &ttsConfig,
	}, nil
}

// Execute implements the NodeInstance interface
func (t *TextToSpeechNode) Execute(ctx context.Context, input map[string]interface{}) (*engine.ExecutionResult, error) {
	// Override configuration with input values if provided
	provider := t.config.Provider
	if inputProvider, ok := input["provider"].(string); ok && inputProvider != "" {
		provider = inputProvider
	}

	apiKey := t.config.ApiKey
	if inputApiKey, ok := input["api_key"].(string); ok && inputApiKey != "" {
		apiKey = inputApiKey
	}

	model := t.config.Model
	if inputModel, ok := input["model"].(string); ok && inputModel != "" {
		model = inputModel
	}

	text := t.config.Text
	if inputText, ok := input["text"].(string); ok && inputText != "" {
		text = inputText
	}

	language := t.config.Language
	if inputLanguage, ok := input["language"].(string); ok && inputLanguage != "" {
		language = inputLanguage
	}

	voice := t.config.Voice
	if inputVoice, ok := input["voice"].(string); ok && inputVoice != "" {
		voice = inputVoice
	}

	speed := t.config.Speed
	if inputSpeed, ok := input["speed"].(float64); ok {
		speed = inputSpeed
	}

	pitch := t.config.Pitch
	if inputPitch, ok := input["pitch"].(float64); ok {
		pitch = inputPitch
	}

	format := t.config.Format
	if inputFormat, ok := input["format"].(string); ok && inputFormat != "" {
		format = inputFormat
	}

	sampleRate := t.config.SampleRate
	if inputSampleRate, ok := input["sample_rate"].(float64); ok {
		sampleRate = int(inputSampleRate)
	}

	bitrate := t.config.Bitrate
	if inputBitrate, ok := input["bitrate"].(float64); ok {
		bitrate = int(inputBitrate)
	}

	emotion := t.config.Emotion
	if inputEmotion, ok := input["emotion"].(string); ok && inputEmotion != "" {
		emotion = inputEmotion
	}

	wordBreak := t.config.WordBreak
	if inputWordBreak, ok := input["word_break"].(bool); ok {
		wordBreak = inputWordBreak
	}

	customParams := t.config.CustomParams
	if inputCustomParams, ok := input["custom_params"].(map[string]interface{}); ok {
		customParams = inputCustomParams
	}

	timeout := t.config.Timeout
	if inputTimeout, ok := input["timeout"].(float64); ok {
		timeout = int(inputTimeout)
	}

	enabled := t.config.Enabled
	if inputEnabled, ok := input["enabled"].(bool); ok {
		enabled = inputEnabled
	}

	// Check if node should be enabled
	if !enabled {
		return &engine.ExecutionResult{
			Status: "success",
			Data: map[string]interface{}{
				"message": "text-to-speech processor disabled, not executed",
				"enabled": false,
			},
			Timestamp: time.Now(),
		}, nil
	}

	// Validate required input
	if text == "" {
		return &engine.ExecutionResult{
			Status:    "error",
			Error:     "text is required for text-to-speech processing",
			Timestamp: time.Now(),
		}, nil
	}

	if apiKey == "" {
		return &engine.ExecutionResult{
			Status:    "error",
			Error:     "api_key is required for text-to-speech processing",
			Timestamp: time.Now(),
		}, nil
	}

	// In a real implementation, this would call the actual text-to-speech service
	// For now, we'll simulate the response
	result := map[string]interface{}{
		"audio_url": "https://example.com/generated-audio.mp3", // Simulated audio URL
		"audio_data": "", // In a real implementation, this might contain base64 encoded audio data
		"text_processed": text,
		"language": language,
		"voice": voice,
		"speed": speed,
		"pitch": pitch,
		"format": format,
		"duration": 5.2, // in seconds
		"file_size": 45000, // in bytes
		"processing_time": time.Since(time.Now().Add(-time.Duration(timeout) * time.Second)).Seconds(),
	}

	return &engine.ExecutionResult{
		Status: "success",
		Data: map[string]interface{}{
			"message":        "text-to-speech processing completed",
			"result":         result,
			"provider":       provider,
			"model":          model,
			"language":       language,
			"voice":          voice,
			"speed":          speed,
			"pitch":          pitch,
			"format":         format,
			"timestamp":      time.Now().Unix(),
		},
		Timestamp: time.Now(),
	}, nil
}

// GetType returns the type of the node
func (t *TextToSpeechNode) GetType() string {
	return "text_to_speech"
}

// GetID returns a unique ID for the node instance
func (t *TextToSpeechNode) GetID() string {
	return "tts_" + fmt.Sprintf("%d", time.Now().Unix())
}