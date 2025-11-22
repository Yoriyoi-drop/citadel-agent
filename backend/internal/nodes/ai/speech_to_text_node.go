package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/engine"
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
func NewSpeechToTextNode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Convert interface{} map to JSON and back to struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}

	var sttConfig SpeechToTextNodeConfig
	err = json.Unmarshal(jsonData, &sttConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// Validate required fields
	if sttConfig.Provider == "" {
		sttConfig.Provider = "openai" // default provider
	}

	if sttConfig.Model == "" {
		sttConfig.Model = "whisper-1" // default model
	}

	if sttConfig.Language == "" {
		sttConfig.Language = "en" // default language
	}

	if sttConfig.Timeout == 0 {
		sttConfig.Timeout = 60 // default timeout of 60 seconds, as audio processing may take longer
	}

	return &SpeechToTextNode{
		config: &sttConfig,
	}, nil
}

// Execute implements the NodeInstance interface
func (s *SpeechToTextNode) Execute(ctx context.Context, input map[string]interface{}) (*engine.ExecutionResult, error) {
	// Override configuration with input values if provided
	provider := s.config.Provider
	if inputProvider, ok := input["provider"].(string); ok && inputProvider != "" {
		provider = inputProvider
	}

	apiKey := s.config.ApiKey
	if inputApiKey, ok := input["api_key"].(string); ok && inputApiKey != "" {
		apiKey = inputApiKey
	}

	model := s.config.Model
	if inputModel, ok := input["model"].(string); ok && inputModel != "" {
		model = inputModel
	}

	audioURL := s.config.AudioURL
	if inputAudioURL, ok := input["audio_url"].(string); ok && inputAudioURL != "" {
		audioURL = inputAudioURL
	}

	audioData := s.config.AudioData
	if inputAudioData, ok := input["audio_data"].(string); ok && inputAudioData != "" {
		audioData = inputAudioData
	}

	language := s.config.Language
	if inputLanguage, ok := input["language"].(string); ok && inputLanguage != "" {
		language = inputLanguage
	}

	sampleRate := s.config.SampleRate
	if inputSampleRate, ok := input["sample_rate"].(float64); ok {
		sampleRate = int(inputSampleRate)
	}

	channels := s.config.Channels
	if inputChannels, ok := input["channels"].(float64); ok {
		channels = int(inputChannels)
	}

	format := s.config.Format
	if inputFormat, ok := input["format"].(string); ok && inputFormat != "" {
		format = inputFormat
	}

	punctuate := s.config.Punctuate
	if inputPunctuate, ok := input["punctuate"].(bool); ok {
		punctuate = inputPunctuate
	}

	profanity := s.config.Profanity
	if inputProfanity, ok := input["profanity"].(string); ok && inputProfanity != "" {
		profanity = inputProfanity
	}

	timestamps := s.config.Timestamps
	if inputTimestamps, ok := input["timestamps"].(bool); ok {
		timestamps = inputTimestamps
	}

	diarization := s.config.Diarization
	if inputDiarization, ok := input["diarization"].(bool); ok {
		diarization = inputDiarization
	}

	customParams := s.config.CustomParams
	if inputCustomParams, ok := input["custom_params"].(map[string]interface{}); ok {
		customParams = inputCustomParams
	}

	timeout := s.config.Timeout
	if inputTimeout, ok := input["timeout"].(float64); ok {
		timeout = int(inputTimeout)
	}

	enabled := s.config.Enabled
	if inputEnabled, ok := input["enabled"].(bool); ok {
		enabled = inputEnabled
	}

	// Check if node should be enabled
	if !enabled {
		return &engine.ExecutionResult{
			Status: "success",
			Data: map[string]interface{}{
				"message": "speech-to-text processor disabled, not executed",
				"enabled": false,
			},
			Timestamp: time.Now(),
		}, nil
	}

	// Validate required input
	if audioURL == "" && audioData == "" {
		return &engine.ExecutionResult{
			Status:    "error",
			Error:     "either audio_url or audio_data is required for speech-to-text processing",
			Timestamp: time.Now(),
		}, nil
	}

	if apiKey == "" {
		return &engine.ExecutionResult{
			Status:    "error",
			Error:     "api_key is required for speech-to-text processing",
			Timestamp: time.Now(),
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

	return &engine.ExecutionResult{
		Status: "success",
		Data: map[string]interface{}{
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
		},
		Timestamp: time.Now(),
	}, nil
}

// GetType returns the type of the node
func (s *SpeechToTextNode) GetType() string {
	return "speech_to_text"
}

// GetID returns a unique ID for the node instance
func (s *SpeechToTextNode) GetID() string {
	return "stt_" + fmt.Sprintf("%d", time.Now().Unix())
}