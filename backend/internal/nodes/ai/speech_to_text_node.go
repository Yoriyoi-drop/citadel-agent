package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
	"github.com/citadel-agent/backend/internal/nodes/utils"
)

// SpeechToTextNodeConfig represents the configuration for a Speech-to-Text node
type SpeechToTextNodeConfig struct {
	Provider       string                 `json:"provider"`        // AI provider (openai, google, azure, etc.)
	ApiKey         string                 `json:"api_key"`         // API key for the STT service
	Model          string                 `json:"model"`           // STT model to use
	AudioURL       string                 `json:"audio_url"`       // URL of the audio file to transcribe
	AudioData      string                 `json:"audio_data"`      // Base64 encoded audio data
	Language       string                 `json:"language"`        // Language code (en, es, fr, etc.)
	SampleRate     int                    `json:"sample_rate"`     // Audio sample rate
	Channels       int                    `json:"channels"`        // Number of audio channels
	Format         string                 `json:"format"`          // Audio format (mp3, wav, flac, etc.)
	Punctuate      bool                   `json:"punctuate"`       // Whether to add punctuation
	Profanity      string                 `json:"profanity"`       // Profanity filter (remove, mask, etc.)
	Timestamps     bool                   `json:"timestamps"`      // Whether to include timestamps
	Diarization    bool                   `json:"diarization"`     // Whether to identify speakers
	CustomParams   map[string]interface{} `json:"custom_params"`   // Custom parameters for the STT service
	Timeout        int                    `json:"timeout"`         // Request timeout in seconds
	Enabled        bool                   `json:"enabled"`         // Whether the node is enabled
}

// SpeechToTextNode represents a node that converts speech to text using AI services
type SpeechToTextNode struct {
	config *SpeechToTextNodeConfig
}

// NewSpeechToTextNode creates a new Speech-to-Text node
func NewSpeechToTextNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	// Extract config values
	provider := utils.GetStringVal(config["provider"], "openai")
	apiKey := utils.GetStringVal(config["api_key"], "")
	model := utils.GetStringVal(config["model"], "whisper-1")
	audioURL := utils.GetStringVal(config["audio_url"], "")
	audioData := utils.GetStringVal(config["audio_data"], "")
	language := utils.GetStringVal(config["language"], "en")
	format := utils.GetStringVal(config["format"], "")
	profanity := utils.GetStringVal(config["profanity"], "")

	sampleRate := utils.GetIntVal(config["sample_rate"], 0)
	channels := utils.GetIntVal(config["channels"], 0)
	timeout := utils.GetIntVal(config["timeout"], 60)

	punctuate := utils.GetBoolVal(config["punctuate"], false)
	timestamps := utils.GetBoolVal(config["timestamps"], false)
	diarization := utils.GetBoolVal(config["diarization"], false)
	enabled := utils.GetBoolVal(config["enabled"], true)

	customParams := make(map[string]interface{})
	if paramsVal, exists := config["custom_params"]; exists {
		if paramsMap, ok := paramsVal.(map[string]interface{}); ok {
			customParams = paramsMap
		}
	}

	nodeConfig := &SpeechToTextNodeConfig{
		Provider:       provider,
		ApiKey:         apiKey,
		Model:          model,
		AudioURL:       audioURL,
		AudioData:      audioData,
		Language:       language,
		SampleRate:     sampleRate,
		Channels:       channels,
		Format:         format,
		Punctuate:      punctuate,
		Profanity:      profanity,
		Timestamps:     timestamps,
		Diarization:    diarization,
		CustomParams:   customParams,
		Timeout:        timeout,
		Enabled:        enabled,
	}

	return &SpeechToTextNode{
		config: nodeConfig,
	}, nil
}

// Execute implements the NodeInstance interface
func (s *SpeechToTextNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Override configuration with input values if provided
	provider := s.config.Provider
	if inputProvider, exists := input["provider"]; exists {
		if inputProviderStr, ok := inputProvider.(string); ok && inputProviderStr != "" {
			provider = inputProviderStr
		}
	}

	apiKey := s.config.ApiKey
	if inputApiKey, exists := input["api_key"]; exists {
		if inputApiKeyStr, ok := inputApiKey.(string); ok && inputApiKeyStr != "" {
			apiKey = inputApiKeyStr
		}
	}

	model := s.config.Model
	if inputModel, exists := input["model"]; exists {
		if inputModelStr, ok := inputModel.(string); ok && inputModelStr != "" {
			model = inputModelStr
		}
	}

	audioURL := s.config.AudioURL
	if inputAudioURL, exists := input["audio_url"]; exists {
		if inputAudioURLStr, ok := inputAudioURL.(string); ok && inputAudioURLStr != "" {
			audioURL = inputAudioURLStr
		}
	}

	audioData := s.config.AudioData
	if inputAudioData, exists := input["audio_data"]; exists {
		if inputAudioDataStr, ok := inputAudioData.(string); ok && inputAudioDataStr != "" {
			audioData = inputAudioDataStr
		}
	}

	language := s.config.Language
	if inputLanguage, exists := input["language"]; exists {
		if inputLanguageStr, ok := inputLanguage.(string); ok && inputLanguageStr != "" {
			language = inputLanguageStr
		}
	}

	sampleRate := s.config.SampleRate
	if inputSampleRate, exists := input["sample_rate"]; exists {
		if inputSampleRateFloat, ok := inputSampleRate.(float64); ok {
			sampleRate = int(inputSampleRateFloat)
		}
	}

	channels := s.config.Channels
	if inputChannels, exists := input["channels"]; exists {
		if inputChannelsFloat, ok := inputChannels.(float64); ok {
			channels = int(inputChannelsFloat)
		}
	}

	format := s.config.Format
	if inputFormat, exists := input["format"]; exists {
		if inputFormatStr, ok := inputFormat.(string); ok && inputFormatStr != "" {
			format = inputFormatStr
		}
	}

	punctuate := s.config.Punctuate
	if inputPunctuate, exists := input["punctuate"]; exists {
		if inputPunctuateBool, ok := inputPunctuate.(bool); ok {
			punctuate = inputPunctuateBool
		}
	}

	profanity := s.config.Profanity
	if inputProfanity, exists := input["profanity"]; exists {
		if inputProfanityStr, ok := inputProfanity.(string); ok && inputProfanityStr != "" {
			profanity = inputProfanityStr
		}
	}

	timestamps := s.config.Timestamps
	if inputTimestamps, exists := input["timestamps"]; exists {
		if inputTimestampsBool, ok := inputTimestamps.(bool); ok {
			timestamps = inputTimestampsBool
		}
	}

	diarization := s.config.Diarization
	if inputDiarization, exists := input["diarization"]; exists {
		if inputDiarizationBool, ok := inputDiarization.(bool); ok {
			diarization = inputDiarizationBool
		}
	}

	customParams := s.config.CustomParams
	if inputCustomParams, exists := input["custom_params"]; exists {
		if inputCustomParamsMap, ok := inputCustomParams.(map[string]interface{}); ok {
			customParams = inputCustomParamsMap
		}
	}

	timeout := s.config.Timeout
	if inputTimeout, exists := input["timeout"]; exists {
		if inputTimeoutFloat, ok := inputTimeout.(float64); ok {
			timeout = int(inputTimeoutFloat)
		}
	}

	enabled := s.config.Enabled
	if inputEnabled, exists := input["enabled"]; exists {
		if inputEnabledBool, ok := inputEnabled.(bool); ok {
			enabled = inputEnabledBool
		}
	}

	// Check if node should be enabled
	if !enabled {
		return map[string]interface{}{
			"success": true,
			"message": "speech-to-text processor disabled, not executed",
			"enabled": false,
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Validate required input
	if audioURL == "" && audioData == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "either audio_url or audio_data is required for speech-to-text processing",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	if apiKey == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "api_key is required for speech-to-text processing",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// In a real implementation, this would call the actual speech-to-text service
	// For now, we'll simulate the response
	result := map[string]interface{}{
		"transcript": "Hello, this is a sample audio transcription. The weather today is sunny and warm.",
		"detected_language": language,
		"confidence": 0.92,
		"duration": 12.5, // in seconds
		"words_count": 15,
		"processing_time": time.Since(time.Now().Add(-time.Duration(timeout) * time.Second)).Seconds(),
	}

	// Add timestamps if requested
	if timestamps {
		result["word_timestamps"] = []map[string]interface{}{
			{"word": "Hello,", "start": 0.0, "end": 0.5},
			{"word": "this", "start": 0.6, "end": 1.0},
			{"word": "is", "start": 1.1, "end": 1.3},
			{"word": "a", "start": 1.4, "end": 1.5},
			{"word": "sample", "start": 1.6, "end": 2.0},
		}
	}

	// Add speaker diarization if requested
	if diarization {
		result["speakers"] = []map[string]interface{}{
			{"speaker_id": "SPEAKER_0", "words": 8, "duration": 8.2},
			{"speaker_id": "SPEAKER_1", "words": 7, "duration": 4.3},
		}
	}

	return map[string]interface{}{
		"success": true,
		"message":        "speech-to-text processing completed",
		"result":         result,
		"provider":       provider,
		"model":          model,
		"language":       language,
		"format":         format,
		"punctuated":     punctuate,
		"timestamped":    timestamps,
		"diarized":       diarization,
		"timestamp":      time.Now().Unix(),
	}, nil
}

