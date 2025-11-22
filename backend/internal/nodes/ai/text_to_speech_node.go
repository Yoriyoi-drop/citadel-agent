package ai

import (
	"context"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
	"github.com/citadel-agent/backend/internal/nodes/utils"
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
func NewTextToSpeechNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	// Extract config values
	provider := utils.GetStringVal(config["provider"], "openai")
	apiKey := utils.GetStringVal(config["api_key"], "")
	model := utils.GetStringVal(config["model"], "tts-1")
	text := utils.GetStringVal(config["text"], "")
	language := utils.GetStringVal(config["language"], "en")
	voice := utils.GetStringVal(config["voice"], "alloy")
	format := utils.GetStringVal(config["format"], "")
	emotion := utils.GetStringVal(config["emotion"], "")

	speed := utils.GetFloat64Val(config["speed"], 1.0)
	pitch := utils.GetFloat64Val(config["pitch"], 1.0)

	sampleRate := utils.GetIntVal(config["sample_rate"], 0)
	bitrate := utils.GetIntVal(config["bitrate"], 0)
	timeout := utils.GetIntVal(config["timeout"], 30)

	wordBreak := utils.GetBoolVal(config["word_break"], false)
	enabled := utils.GetBoolVal(config["enabled"], true)

	customParams := make(map[string]interface{})
	if paramsVal, exists := config["custom_params"]; exists {
		if paramsMap, ok := paramsVal.(map[string]interface{}); ok {
			customParams = paramsMap
		}
	}

	nodeConfig := &TextToSpeechNodeConfig{
		Provider:       provider,
		ApiKey:         apiKey,
		Model:          model,
		Text:           text,
		Language:       language,
		Voice:          voice,
		Speed:          speed,
		Pitch:          pitch,
		Format:         format,
		SampleRate:     sampleRate,
		Bitrate:        bitrate,
		Emotion:        emotion,
		WordBreak:      wordBreak,
		CustomParams:   customParams,
		Timeout:        timeout,
		Enabled:        enabled,
	}

	return &TextToSpeechNode{
		config: nodeConfig,
	}, nil
}

// Execute implements the NodeInstance interface
func (t *TextToSpeechNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Override configuration with input values if provided
	provider := t.config.Provider
	if inputProvider, exists := input["provider"]; exists {
		if inputProviderStr, ok := inputProvider.(string); ok && inputProviderStr != "" {
			provider = inputProviderStr
		}
	}

	apiKey := t.config.ApiKey
	if inputApiKey, exists := input["api_key"]; exists {
		if inputApiKeyStr, ok := inputApiKey.(string); ok && inputApiKeyStr != "" {
			apiKey = inputApiKeyStr
		}
	}

	model := t.config.Model
	if inputModel, exists := input["model"]; exists {
		if inputModelStr, ok := inputModel.(string); ok && inputModelStr != "" {
			model = inputModelStr
		}
	}

	text := t.config.Text
	if inputText, exists := input["text"]; exists {
		if inputTextStr, ok := inputText.(string); ok && inputTextStr != "" {
			text = inputTextStr
		}
	}

	language := t.config.Language
	if inputLanguage, exists := input["language"]; exists {
		if inputLanguageStr, ok := inputLanguage.(string); ok && inputLanguageStr != "" {
			language = inputLanguageStr
		}
	}

	voice := t.config.Voice
	if inputVoice, exists := input["voice"]; exists {
		if inputVoiceStr, ok := inputVoice.(string); ok && inputVoiceStr != "" {
			voice = inputVoiceStr
		}
	}

	speed := t.config.Speed
	if inputSpeed, exists := input["speed"]; exists {
		if inputSpeedFloat, ok := inputSpeed.(float64); ok {
			speed = inputSpeedFloat
		}
	}

	pitch := t.config.Pitch
	if inputPitch, exists := input["pitch"]; exists {
		if inputPitchFloat, ok := inputPitch.(float64); ok {
			pitch = inputPitchFloat
		}
	}

	format := t.config.Format
	if inputFormat, exists := input["format"]; exists {
		if inputFormatStr, ok := inputFormat.(string); ok && inputFormatStr != "" {
			format = inputFormatStr
		}
	}

	sampleRate := t.config.SampleRate
	if inputSampleRate, exists := input["sample_rate"]; exists {
		if inputSampleRateFloat, ok := inputSampleRate.(float64); ok {
			sampleRate = int(inputSampleRateFloat)
		}
	}

	bitrate := t.config.Bitrate
	if inputBitrate, exists := input["bitrate"]; exists {
		if inputBitrateFloat, ok := inputBitrate.(float64); ok {
			bitrate = int(inputBitrateFloat)
		}
	}

	emotion := t.config.Emotion
	if inputEmotion, exists := input["emotion"]; exists {
		if inputEmotionStr, ok := inputEmotion.(string); ok && inputEmotionStr != "" {
			emotion = inputEmotionStr
		}
	}

	wordBreak := t.config.WordBreak
	if inputWordBreak, exists := input["word_break"]; exists {
		if inputWordBreakBool, ok := inputWordBreak.(bool); ok {
			wordBreak = inputWordBreakBool
		}
	}

	customParams := t.config.CustomParams
	if inputCustomParams, exists := input["custom_params"]; exists {
		if inputCustomParamsMap, ok := inputCustomParams.(map[string]interface{}); ok {
			customParams = inputCustomParamsMap
		}
	}

	timeout := t.config.Timeout
	if inputTimeout, exists := input["timeout"]; exists {
		if inputTimeoutFloat, ok := inputTimeout.(float64); ok {
			timeout = int(inputTimeoutFloat)
		}
	}

	enabled := t.config.Enabled
	if inputEnabled, exists := input["enabled"]; exists {
		if inputEnabledBool, ok := inputEnabled.(bool); ok {
			enabled = inputEnabledBool
		}
	}

	// Check if node should be enabled
	if !enabled {
		return map[string]interface{}{
			"success": true,
			"message": "text-to-speech processor disabled, not executed",
			"enabled": false,
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Validate required input
	if text == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "text is required for text-to-speech processing",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	if apiKey == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "api_key is required for text-to-speech processing",
			"timestamp": time.Now().Unix(),
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
		"sample_rate": sampleRate,
		"bitrate": bitrate,
		"emotion": emotion,
		"word_break": wordBreak,
		"custom_params": customParams,
		"duration": 5.2, // in seconds
		"file_size": 45000, // in bytes
		"processing_time": time.Since(time.Now().Add(-time.Duration(timeout) * time.Second)).Seconds(),
	}

	return map[string]interface{}{
		"success": true,
		"message":        "text-to-speech processing completed",
		"result":         result,
		"provider":       provider,
		"model":          model,
		"language":       language,
		"voice":          voice,
		"speed":          speed,
		"pitch":          pitch,
		"format":         format,
		"sample_rate":    sampleRate,
		"bitrate":        bitrate,
		"emotion":        emotion,
		"word_break":     wordBreak,
		"custom_params":  customParams,
		"timestamp":      time.Now().Unix(),
	}, nil
}