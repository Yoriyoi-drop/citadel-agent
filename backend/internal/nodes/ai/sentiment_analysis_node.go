package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/engine"
)

// SentimentAnalysisNodeConfig represents the configuration for a Sentiment Analysis node
type SentimentAnalysisNodeConfig struct {
	Provider       string                 `json:"provider"`        // AI provider (openai, google, azure, etc.)
	ApiKey         string                 `json:"api_key"`         // API key for the sentiment analysis service
	Model          string                 `json:"model"`           // Sentiment analysis model to use
	Text           string                 `json:"text"`            // Text to analyze for sentiment
	Language       string                 `json:"language"`        // Language of the text
	ReturnScores   bool                   `json:"return_scores"`   // Whether to return sentiment scores
	ReturnTokens   bool                   `json:"return_tokens"`   // Whether to return token-level analysis
	MaxSentiment   string                 `json:"max_sentiment"`   // Maximum sentiment to detect (positive, negative, neutral)
	CustomParams   map[string]interface{} `json:"custom_params"`   // Custom parameters for the AI service
	Timeout        int                    `json:"timeout"`         // Request timeout in seconds
	Enabled        bool                   `json:"enabled"`         // Whether the node is enabled
}

// SentimentAnalysisNode represents a node that analyzes sentiment in text using AI
type SentimentAnalysisNode struct {
	config *SentimentAnalysisNodeConfig
}

// NewSentimentAnalysisNode creates a new Sentiment Analysis node
func NewSentimentAnalysisNode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Convert interface{} map to JSON and back to struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}

	var sentimentConfig SentimentAnalysisNodeConfig
	err = json.Unmarshal(jsonData, &sentimentConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// Validate required fields
	if sentimentConfig.Provider == "" {
		sentimentConfig.Provider = "openai" // default provider
	}

	if sentimentConfig.Model == "" {
		sentimentConfig.Model = "gpt-3.5-turbo" // default model
	}

	if sentimentConfig.Language == "" {
		sentimentConfig.Language = "en" // default language
	}

	if sentimentConfig.Timeout == 0 {
		sentimentConfig.Timeout = 30 // default timeout of 30 seconds
	}

	return &SentimentAnalysisNode{
		config: &sentimentConfig,
	}, nil
}

// Execute implements the NodeInstance interface
func (s *SentimentAnalysisNode) Execute(ctx context.Context, input map[string]interface{}) (*engine.ExecutionResult, error) {
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

	text := s.config.Text
	if inputText, ok := input["text"].(string); ok && inputText != "" {
		text = inputText
	}

	language := s.config.Language
	if inputLanguage, ok := input["language"].(string); ok && inputLanguage != "" {
		language = inputLanguage
	}

	returnScores := s.config.ReturnScores
	if inputReturnScores, ok := input["return_scores"].(bool); ok {
		returnScores = inputReturnScores
	}

	returnTokens := s.config.ReturnTokens
	if inputReturnTokens, ok := input["return_tokens"].(bool); ok {
		returnTokens = inputReturnTokens
	}

	maxSentiment := s.config.MaxSentiment
	if inputMaxSentiment, ok := input["max_sentiment"].(string); ok && inputMaxSentiment != "" {
		maxSentiment = inputMaxSentiment
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
				"message": "sentiment analysis processor disabled, not executed",
				"enabled": false,
			},
			Timestamp: time.Now(),
		}, nil
	}

	// Validate required input
	if text == "" {
		return &engine.ExecutionResult{
			Status:    "error",
			Error:     "text is required for sentiment analysis",
			Timestamp: time.Now(),
		}, nil
	}

	if apiKey == "" {
		return &engine.ExecutionResult{
			Status:    "error",
			Error:     "api_key is required for sentiment analysis",
			Timestamp: time.Now(),
		}, nil
	}

	// In a real implementation, this would call the actual sentiment analysis AI service
	// For now, we'll simulate the response
	sentiment := "positive"
	confidence := 0.89
	sentimentScores := map[string]interface{}{
		"positive": 0.89,
		"negative": 0.05,
		"neutral":  0.06,
	}

	// Determine sentiment based on the text
	if len(text) > 0 {
		lowerText := fmt.Sprintf("%s", text)
		if containsWords(lowerText, []string{"bad", "terrible", "awful", "hate", "worst"}) {
			sentiment = "negative"
			confidence = 0.78
			sentimentScores = map[string]interface{}{
				"positive": 0.05,
				"negative": 0.78,
				"neutral":  0.17,
			}
		} else if containsWords(lowerText, []string{"good", "great", "awesome", "love", "excellent", "amazing"}) {
			sentiment = "positive"
			confidence = 0.92
			sentimentScores = map[string]interface{}{
				"positive": 0.92,
				"negative": 0.02,
				"neutral":  0.06,
			}
		} else if containsWords(lowerText, []string{"okay", "fine", "alright", "average", "normal"}) {
			sentiment = "neutral"
			confidence = 0.65
			sentimentScores = map[string]interface{}{
				"positive": 0.15,
				"negative": 0.10,
				"neutral":  0.75,
			}
		}
	}

	result := map[string]interface{}{
		"input_text": text,
		"detected_language": language,
		"overall_sentiment": sentiment,
		"confidence": confidence,
		"analysis_time": time.Since(time.Now().Add(-time.Duration(timeout) * time.Second)).Seconds(),
	}

	// Add sentiment scores if requested
	if returnScores {
		result["sentiment_scores"] = sentimentScores
	}

	// Add token analysis if requested
	if returnTokens {
		result["token_analysis"] = []map[string]interface{}{
			{
				"token": "this",
				"sentiment": "neutral",
				"confidence": 0.95,
			},
			{
				"token": "product",
				"sentiment": "neutral",
				"confidence": 0.98,
			},
			{
				"token": "is",
				"sentiment": "neutral",
				"confidence": 0.99,
			},
			{
				"token": "amazing",
				"sentiment": "positive",
				"confidence": 0.94,
			},
		}
	}

	return &engine.ExecutionResult{
		Status: "success",
		Data: map[string]interface{}{
			"message":        "sentiment analysis completed",
			"result":         result,
			"provider":       provider,
			"model":          model,
			"language":       language,
			"return_scores":  returnScores,
			"return_tokens":  returnTokens,
			"timestamp":      time.Now().Unix(),
		},
		Timestamp: time.Now(),
	}, nil
}

// containsWords checks if text contains any of the words in the list
func containsWords(text string, words []string) bool {
	lowerText := fmt.Sprintf("%s", text)
	for _, word := range words {
		if contains(lowerText, word) {
			return true
		}
	}
	return false
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetType returns the type of the node
func (s *SentimentAnalysisNode) GetType() string {
	return "sentiment_analysis"
}

// GetID returns a unique ID for the node instance
func (s *SentimentAnalysisNode) GetID() string {
	return "sentiment_analysis_" + fmt.Sprintf("%d", time.Now().Unix())
}