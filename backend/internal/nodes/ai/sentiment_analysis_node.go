package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
	"github.com/citadel-agent/backend/internal/nodes/utils"
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
func NewSentimentAnalysisNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	// Extract config values
	provider := utils.GetStringVal(config["provider"], "openai")
	apiKey := utils.GetStringVal(config["api_key"], "")
	model := utils.GetStringVal(config["model"], "gpt-3.5-turbo")
	text := utils.GetStringVal(config["text"], "")
	language := utils.GetStringVal(config["language"], "en")
	maxSentiment := utils.GetStringVal(config["max_sentiment"], "")

	returnScores := utils.GetBoolVal(config["return_scores"], false)
	returnTokens := utils.GetBoolVal(config["return_tokens"], false)
	enabled := utils.GetBoolVal(config["enabled"], true)
	timeout := utils.GetIntVal(config["timeout"], 30)

	customParams := make(map[string]interface{})
	if paramsVal, exists := config["custom_params"]; exists {
		if paramsMap, ok := paramsVal.(map[string]interface{}); ok {
			customParams = paramsMap
		}
	}

	nodeConfig := &SentimentAnalysisNodeConfig{
		Provider:       provider,
		ApiKey:         apiKey,
		Model:          model,
		Text:           text,
		Language:       language,
		ReturnScores:   returnScores,
		ReturnTokens:   returnTokens,
		MaxSentiment:   maxSentiment,
		CustomParams:   customParams,
		Timeout:        timeout,
		Enabled:        enabled,
	}

	return &SentimentAnalysisNode{
		config: nodeConfig,
	}, nil
}

// Execute implements the NodeInstance interface
func (s *SentimentAnalysisNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
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

	text := s.config.Text
	if inputText, exists := input["text"]; exists {
		if inputTextStr, ok := inputText.(string); ok && inputTextStr != "" {
			text = inputTextStr
		}
	}

	language := s.config.Language
	if inputLanguage, exists := input["language"]; exists {
		if inputLanguageStr, ok := inputLanguage.(string); ok && inputLanguageStr != "" {
			language = inputLanguageStr
		}
	}

	returnScores := s.config.ReturnScores
	if inputReturnScores, exists := input["return_scores"]; exists {
		if inputReturnScoresBool, ok := inputReturnScores.(bool); ok {
			returnScores = inputReturnScoresBool
		}
	}

	returnTokens := s.config.ReturnTokens
	if inputReturnTokens, exists := input["return_tokens"]; exists {
		if inputReturnTokensBool, ok := inputReturnTokens.(bool); ok {
			returnTokens = inputReturnTokensBool
		}
	}

	maxSentiment := s.config.MaxSentiment
	if inputMaxSentiment, exists := input["max_sentiment"]; exists {
		if inputMaxSentimentStr, ok := inputMaxSentiment.(string); ok && inputMaxSentimentStr != "" {
			maxSentiment = inputMaxSentimentStr
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
			"message": "sentiment analysis processor disabled, not executed",
			"enabled": false,
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Validate required input
	if text == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "text is required for sentiment analysis",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	if apiKey == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "api_key is required for sentiment analysis",
			"timestamp": time.Now().Unix(),
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

	return map[string]interface{}{
		"success": true,
		"message":        "sentiment analysis completed",
		"result":         result,
		"provider":       provider,
		"model":          model,
		"language":       language,
		"return_scores":  returnScores,
		"return_tokens":  returnTokens,
		"timestamp":      time.Now().Unix(),
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

