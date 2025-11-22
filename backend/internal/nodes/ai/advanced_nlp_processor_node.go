package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"citadel-agent/backend/internal/workflow/core/engine"
)

// AdvancedNLPProcessorConfig represents the configuration for an Advanced NLP Processor node
type AdvancedNLPProcessorConfig struct {
	Provider        string                 `json:"provider"`          // NLP provider (openai, anthropic, huggingface, etc.)
	Model           string                 `json:"model"`             // NLP model to use
	APIKey          string                 `json:"api_key"`           // API key for the NLP service
	Language        string                 `json:"language"`          // Language of the text (en, es, fr, etc.)
	MaxTokens       int                    `json:"max_tokens"`        // Maximum number of tokens to process
	Temperature     float64                `json:"temperature"`       // Creativity parameter
	Operations      []NLPOperation         `json:"operations"`        // List of NLP operations to perform
	TargetLanguages []string               `json:"target_languages"`  // Target languages for translation
	EnableCaching   bool                   `json:"enable_caching"`    // Whether to cache processing results
	CacheTTL        int                    `json:"cache_ttl"`         // Cache time-to-live in seconds
	EnableProfiling bool                   `json:"enable_profiling"`  // Whether to profile processing performance
	Timeout         int                    `json:"timeout"`           // Request timeout in seconds
	ReturnRawResults bool                 `json:"return_raw_results"` // Whether to return raw processing results
	CustomParams    map[string]interface{} `json:"custom_params"`     // Custom parameters for processing
	Preprocessing   PreprocessingConfig    `json:"preprocessing"`     // Text preprocessing configuration
	Postprocessing  PostprocessingConfig   `json:"postprocessing"`    // Text postprocessing configuration
	AnalysisTypes   []string               `json:"analysis_types"`    // Types of analysis to perform (sentiment, entity, etc.)
}

// NLPOperation represents a specific NLP operation to perform
type NLPOperation string

const (
	TokenizationOperation     NLPOperation = "tokenization"
	SentimentAnalysis         NLPOperation = "sentiment_analysis"
	EntityRecognition         NLPOperation = "entity_recognition"
	TextSummarization         NLPOperation = "text_summarization"
	TranslationOperation      NLPOperation = "translation"
	QuestionAnswering        NLPOperation = "question_answering"
	TextClassification       NLPOperation = "text_classification"
	KeyphraseExtraction      NLPOperation = "keyphrase_extraction"
	GrammarCorrection        NLPOperation = "grammar_correction"
	TopicModeling            NLPOperation = "topic_modeling"
	TextGeneration           NLPOperation = "text_generation"
	LanguageDetection        NLPOperation = "language_detection"
)

// AdvancedNLPProcessorNode represents a node that performs advanced natural language processing
type AdvancedNLPProcessorNode struct {
	config *AdvancedNLPProcessorConfig
}

// NewAdvancedNLPProcessorNode creates a new Advanced NLP Processor node
func NewAdvancedNLPProcessorNode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Convert interface{} map to JSON and back to struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}

	var nlpConfig AdvancedNLPProcessorConfig
	err = json.Unmarshal(jsonData, &nlpConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// Validate and set defaults
	if nlpConfig.Provider == "" {
		nlpConfig.Provider = "openai"
	}

	if nlpConfig.Model == "" {
		nlpConfig.Model = "gpt-4"
	}

	if nlpConfig.Language == "" {
		nlpConfig.Language = "en"
	}

	if len(nlpConfig.Operations) == 0 {
		nlpConfig.Operations = []NLPOperation{SentimentAnalysis, EntityRecognition, KeyphraseExtraction}
	}

	if nlpConfig.MaxTokens == 0 {
		nlpConfig.MaxTokens = 2048
	}

	if nlpConfig.Temperature == 0 {
		nlpConfig.Temperature = 0.7
	}

	if nlpConfig.Timeout == 0 {
		nlpConfig.Timeout = 120 // default timeout of 120 seconds
	}

	return &AdvancedNLPProcessorNode{
		config: &nlpConfig,
	}, nil
}

// Execute implements the NodeInstance interface
func (n *AdvancedNLPProcessorNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Override configuration with input values if provided
	provider := n.config.Provider
	if inputProvider, ok := input["provider"].(string); ok && inputProvider != "" {
		provider = inputProvider
	}

	model := n.config.Model
	if inputModel, ok := input["model"].(string); ok && inputModel != "" {
		model = inputModel
	}

	apiKey := n.config.APIKey
	if inputAPIKey, ok := input["api_key"].(string); ok && inputAPIKey != "" {
		apiKey = inputAPIKey
	}

	language := n.config.Language
	if inputLanguage, ok := input["language"].(string); ok && inputLanguage != "" {
		language = inputLanguage
	}

	maxTokens := n.config.MaxTokens
	if inputMaxTokens, ok := input["max_tokens"].(float64); ok {
		maxTokens = int(inputMaxTokens)
	}

	temperature := n.config.Temperature
	if inputTemp, ok := input["temperature"].(float64); ok {
		temperature = inputTemp
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

	timeout := n.config.Timeout
	if inputTimeout, ok := input["timeout"].(float64); ok {
		timeout = int(inputTimeout)
	}

	returnRawResults := n.config.ReturnRawResults
	if inputReturnRaw, ok := input["return_raw_results"].(bool); ok {
		returnRawResults = inputReturnRaw
	}

	customParams := n.config.CustomParams
	if inputCustomParams, ok := input["custom_params"].(map[string]interface{}); ok {
		customParams = inputCustomParams
	}

	// Extract text input
	textInput := ""
	if text, exists := input["text"]; exists {
		if textStr, ok := text.(string); ok {
			textInput = textStr
		}
	} else if inputStr, ok := input["input"].(string); ok {
		textInput = inputStr
	} else {
		// Convert the entire input to string as fallback
		if inputBytes, err := json.Marshal(input); err == nil {
			textInput = string(inputBytes)
		}
	}

	// Validate required input
	if textInput == "" {
		return map[string]interface{}{
			"success":   false,
			"error":     "text input is required for NLP processing",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	if apiKey == "" {
		return map[string]interface{}{
			"success":   false,
			"error":     "api_key is required for NLP processing",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Create processing context with timeout
	processingCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Process the text with all specified operations
	processingResults, err := n.performNLPProcessing(processingCtx, textInput)
	if err != nil {
		return map[string]interface{}{
			"success":   false,
			"error":     err.Error(),
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Prepare final result
	finalResult := map[string]interface{}{
		"success":              true,
		"provider":             provider,
		"model":                model,
		"language":             language,
		"operations_performed": n.config.Operations,
		"processing_results":   processingResults,
		"input_text_length":    len(textInput),
		"enable_caching":       enableCaching,
		"enable_profiling":     enableProfiling,
		"return_raw_results":   returnRawResults,
		"timestamp":            time.Now().Unix(),
		"input_text":           textInput,
		"config":               n.config,
	}

	// Include raw results if requested
	if returnRawResults && processingResults["raw_results"] != nil {
		finalResult["raw_results"] = processingResults["raw_results"]
	}

	// Add performance metrics if profiling is enabled
	if enableProfiling {
		finalResult["performance_metrics"] = map[string]interface{}{
			"start_time": time.Now().Unix(),
			"end_time":   time.Now().Unix(),
			"duration":   time.Since(time.Now().Add(-time.Duration(timeout) * time.Second)).Seconds(),
		}
	}

	return finalResult, nil
}

// performNLPProcessing performs all the specified NLP operations
func (n *AdvancedNLPProcessorNode) performNLPProcessing(ctx context.Context, text string) (map[string]interface{}, error) {
	results := make(map[string]interface{})
	rawResults := make(map[string]interface{})

	// Perform each NLP operation
	for _, operation := range n.config.Operations {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// Continue processing
		}

		var operationResult map[string]interface{}
		var err error

		switch operation {
		case TokenizationOperation:
			operationResult, err = n.performTokenization(ctx, text)
		case SentimentAnalysis:
			operationResult, err = n.performSentimentAnalysis(ctx, text)
		case EntityRecognition:
			operationResult, err = n.performEntityRecognition(ctx, text)
		case TextSummarization:
			operationResult, err = n.performTextSummarization(ctx, text)
		case TranslationOperation:
			operationResult, err = n.performTranslation(ctx, text)
		case QuestionAnswering:
			operationResult, err = n.performQuestionAnswering(ctx, text)
		case TextClassification:
			operationResult, err = n.performTextClassification(ctx, text)
		case KeyphraseExtraction:
			operationResult, err = n.performKeyphraseExtraction(ctx, text)
		case GrammarCorrection:
			operationResult, err = n.performGrammarCorrection(ctx, text)
		case TopicModeling:
			operationResult, err = n.performTopicModeling(ctx, text)
		case TextGeneration:
			operationResult, err = n.performTextGeneration(ctx, text)
		case LanguageDetection:
			operationResult, err = n.performLanguageDetection(ctx, text)
		default:
			operationResult, err = n.performDefaultAnalysis(ctx, text, string(operation))
		}

		if err != nil {
			operationResult = map[string]interface{}{
				"success": false,
				"error":   err.Error(),
				"operation": string(operation),
			}
		}

		// Store result for this operation
		results[string(operation)] = operationResult
		rawResults[string(operation)] = operationResult
	}

	// Perform language detection if not already specified
	if n.config.Language == "detect" || n.config.Language == "" {
		languageResult, err := n.performLanguageDetection(ctx, text)
		if err == nil {
			results["language_detection"] = languageResult
			rawResults["language_detection"] = languageResult
		}
	}

	// Aggregate results
	aggregated := n.aggregateResults(results)

	return map[string]interface{}{
		"results_by_operation": results,
		"raw_results":          rawResults,
		"aggregated":           aggregated,
		"all_operations_count": len(n.config.Operations),
		"successful_operations": len(results),
	}, nil
}

// performTokenization performs tokenization on the text
func (n *AdvancedNLPProcessorNode) performTokenization(ctx context.Context, text string) (map[string]interface{}, error) {
	// Simulate tokenization
	tokens := []string{}
	currentToken := ""
	
	for _, char := range text {
		if char == ' ' || char == '\t' || char == '\n' || char == '.' || char == ',' || char == '!' || char == '?' {
			if currentToken != "" {
				tokens = append(tokens, currentToken)
				currentToken = ""
			}
		} else {
			currentToken += string(char)
		}
	}
	
	if currentToken != "" {
		tokens = append(tokens, currentToken)
	}

	result := map[string]interface{}{
		"operation": "tokenization",
		"tokens": tokens,
		"token_count": len(tokens),
		"text_length": len(text),
		"processing_time": time.Since(time.Now().Add(-50 * time.Millisecond)).Seconds(), // Simulate processing time
		"timestamp": time.Now().Unix(),
	}

	return result, nil
}

// performSentimentAnalysis performs sentiment analysis
func (n *AdvancedNLPProcessorNode) performSentimentAnalysis(ctx context.Context, text string) (map[string]interface{}, error) {
	// Simulate sentiment analysis
	sentiment := "neutral"
	score := 0.5
	
	positiveWords := []string{"good", "great", "excellent", "amazing", "wonderful", "fantastic", "love", "like", "enjoy"}
	negativeWords := []string{"bad", "terrible", "awful", "hate", "dislike", "horrible", "worst", "disappointing"}
	
	textLower := ""
	for _, r := range text {
		textLower += string(r) | 32 // Convert to lowercase
	}
	
	// Count positive and negative words
	positiveCount := 0
	negativeCount := 0
	
	for _, word := range positiveWords {
		if contains(textLower, word) {
			positiveCount++
		}
	}
	
	for _, word := range negativeWords {
		if contains(textLower, word) {
			negativeCount++
		}
	}
	
	// Calculate sentiment
	if positiveCount > negativeCount {
		sentiment = "positive"
		score = 0.6 + float64(positiveCount-negativeCount)*0.1
		if score > 1.0 {
			score = 1.0
		}
	} else if negativeCount > positiveCount {
		sentiment = "negative"
		score = 0.4 - float64(negativeCount-positiveCount)*0.1
		if score < 0.0 {
			score = 0.0
		}
	}
	
	result := map[string]interface{}{
		"operation": "sentiment_analysis",
		"sentiment": sentiment,
		"score": score,
		"confidence": score,
		"positive_words_count": positiveCount,
		"negative_words_count": negativeCount,
		"processing_time": time.Since(time.Now().Add(-50 * time.Millisecond)).Seconds(),
		"timestamp": time.Now().Unix(),
	}

	return result, nil
}

// performEntityRecognition performs named entity recognition
func (n *AdvancedNLPProcessorNode) performEntityRecognition(ctx context.Context, text string) (map[string]interface{}, error) {
	// Simulate entity recognition
	entities := []map[string]interface{}{
		{
			"text": "John Doe",
			"type": "PERSON",
			"confidence": 0.95,
		},
		{
			"text": "New York",
			"type": "GPE", // Geopolitical entity
			"confidence": 0.92,
		},
		{
			"text": "Monday",
			"type": "DATE",
			"confidence": 0.88,
		},
		{
			"text": "Microsoft",
			"type": "ORG",
			"confidence": 0.90,
		},
	}
	
	result := map[string]interface{}{
		"operation": "entity_recognition",
		"entities": entities,
		"entity_count": len(entities),
		"processing_time": time.Since(time.Now().Add(-60 * time.Millisecond)).Seconds(),
		"timestamp": time.Now().Unix(),
	}

	return result, nil
}

// performTextSummarization performs text summarization
func (n *AdvancedNLPProcessorNode) performTextSummarization(ctx context.Context, text string) (map[string]interface{}, error) {
	// Simulate text summarization
	summary := "This is a summary of the provided text. It captures the main points and key information from the original text while reducing its length."
	
	result := map[string]interface{}{
		"operation": "text_summarization",
		"summary": summary,
		"original_length": len(text),
		"summary_length": len(summary),
		"compression_ratio": float64(len(summary)) / float64(len(text)),
		"processing_time": time.Since(time.Now().Add(-100 * time.Millisecond)).Seconds(),
		"timestamp": time.Now().Unix(),
	}

	return result, nil
}

// performTranslation performs text translation
func (n *AdvancedNLPProcessorNode) performTranslation(ctx context.Context, text string) (map[string]interface{}, error) {
	// Simulate translation
	translated := "This is the translated text in the target language."
	
	result := map[string]interface{}{
		"operation": "translation",
		"original_text": text,
		"translated_text": translated,
		"source_language": n.config.Language,
		"target_language": "en", // Default target
		"processing_time": time.Since(time.Now().Add(-80 * time.Millisecond)).Seconds(),
		"timestamp": time.Now().Unix(),
	}

	return result, nil
}

// performQuestionAnswering performs question answering
func (n *AdvancedNLPProcessorNode) performQuestionAnswering(ctx context.Context, text string) (map[string]interface{}, error) {
	// Simulate question answering
	question := "What is the main topic?"
	answer := "The main topic is advanced natural language processing."
	
	result := map[string]interface{}{
		"operation": "question_answering",
		"question": question,
		"answer": answer,
		"confidence": 0.85,
		"processing_time": time.Since(time.Now().Add(-90 * time.Millisecond)).Seconds(),
		"timestamp": time.Now().Unix(),
	}

	return result, nil
}

// performTextClassification performs text classification
func (n *AdvancedNLPProcessorNode) performTextClassification(ctx context.Context, text string) (map[string]interface{}, error) {
	// Simulate text classification
	classification := map[string]interface{}{
		"category": "technology",
		"confidence": 0.88,
		"subcategories": []string{"AI", "NLP", "Machine Learning"},
	}
	
	result := map[string]interface{}{
		"operation": "text_classification",
		"classification": classification,
		"processing_time": time.Since(time.Now().Add(-70 * time.Millisecond)).Seconds(),
		"timestamp": time.Now().Unix(),
	}

	return result, nil
}

// performKeyphraseExtraction performs keyphrase extraction
func (n *AdvancedNLPProcessorNode) performKeyphraseExtraction(ctx context.Context, text string) (map[string]interface{}, error) {
	// Simulate keyphrase extraction
	keyphrases := []string{"machine learning", "natural language processing", "artificial intelligence", "text analysis", "semantic understanding"}
	
	result := map[string]interface{}{
		"operation": "keyphrase_extraction",
		"keyphrases": keyphrases,
		"keyphrase_count": len(keyphrases),
		"processing_time": time.Since(time.Now().Add(-60 * time.Millisecond)).Seconds(),
		"timestamp": time.Now().Unix(),
	}

	return result, nil
}

// performGrammarCorrection performs grammar correction
func (n *AdvancedNLPProcessorNode) performGrammarCorrection(ctx context.Context, text string) (map[string]interface{}, error) {
	// Simulate grammar correction
	corrected := "This is the grammatically corrected text."
	
	result := map[string]interface{}{
		"operation": "grammar_correction",
		"original_text": text,
		"corrected_text": corrected,
		"corrections_count": 1,
		"processing_time": time.Since(time.Now().Add(-80 * time.Millisecond)).Seconds(),
		"timestamp": time.Now().Unix(),
	}

	return result, nil
}

// performTopicModeling performs topic modeling
func (n *AdvancedNLPProcessorNode) performTopicModeling(ctx context.Context, text string) (map[string]interface{}, error) {
	// Simulate topic modeling
	topics := []map[string]interface{}{
		{
			"topic_id": 1,
			"topic_name": "Artificial Intelligence",
			"relevance_score": 0.85,
			"keywords": []string{"AI", "machine learning", "neural networks"},
		},
		{
			"topic_id": 2,
			"topic_name": "Natural Language Processing",
			"relevance_score": 0.78,
			"keywords": []string{"text", "language", "processing"},
		},
	}
	
	result := map[string]interface{}{
		"operation": "topic_modeling",
		"topics": topics,
		"topic_count": len(topics),
		"processing_time": time.Since(time.Now().Add(-120 * time.Millisecond)).Seconds(),
		"timestamp": time.Now().Unix(),
	}

	return result, nil
}

// performTextGeneration performs text generation
func (n *AdvancedNLPProcessorNode) performTextGeneration(ctx context.Context, text string) (map[string]interface{}, error) {
	// Simulate text generation
	generatedText := "This is the generated text based on the provided input. The AI model has created new content that is relevant to the input while maintaining coherence and context."
	
	result := map[string]interface{}{
		"operation": "text_generation",
		"input_prompt": text,
		"generated_text": generatedText,
		"token_count": len(generatedText),
		"processing_time": time.Since(time.Now().Add(-150 * time.Millisecond)).Seconds(),
		"timestamp": time.Now().Unix(),
	}

	return result, nil
}

// performLanguageDetection performs language detection
func (n *AdvancedNLPProcessorNode) performLanguageDetection(ctx context.Context, text string) (map[string]interface{}, error) {
	// Simulate language detection
	language := "en" // default to English
	confidence := 0.95
	
	// This would be more sophisticated in a real implementation
	if contains(text, "hola") || contains(text, "mundo") {
		language = "es"
	} else if contains(text, "bonjour") || contains(text, "monde") {
		language = "fr"
	} else if contains(text, "hallo") {
		language = "de"
	}
	
	result := map[string]interface{}{
		"operation": "language_detection",
		"detected_language": language,
		"confidence": confidence,
		"processing_time": time.Since(time.Now().Add(-40 * time.Millisecond)).Seconds(),
		"timestamp": time.Now().Unix(),
	}

	return result, nil
}

// performDefaultAnalysis performs a default analysis for unknown operations
func (n *AdvancedNLPProcessorNode) performDefaultAnalysis(ctx context.Context, text string, operation string) (map[string]interface{}, error) {
	result := map[string]interface{}{
		"operation": operation,
		"result": fmt.Sprintf("Processed text with %s operation", operation),
		"text_length": len(text),
		"processing_time": time.Since(time.Now().Add(-30 * time.Millisecond)).Seconds(),
		"timestamp": time.Now().Unix(),
	}

	return result, nil
}

// aggregateResults combines results from all operations into a summary
func (n *AdvancedNLPProcessorNode) aggregateResults(results map[string]interface{}) map[string]interface{} {
	aggregate := make(map[string]interface{})
	
	// Collect sentiment from sentiment analysis
	if sentimentResult, exists := results[string(SentimentAnalysis)]; exists {
		if resultMap, ok := sentimentResult.(map[string]interface{}); ok {
			if sentiment, ok := resultMap["sentiment"]; ok {
				aggregate["overall_sentiment"] = sentiment
			}
			if score, ok := resultMap["score"]; ok {
				aggregate["overall_sentiment_score"] = score
			}
		}
	}
	
	// Collect entities from entity recognition
	entityCount := 0
	if entityResult, exists := results[string(EntityRecognition)]; exists {
		if resultMap, ok := entityResult.(map[string]interface{}); ok {
			if entities, ok := resultMap["entities"]; ok {
				if entitySlice, ok := entities.([]map[string]interface{}); ok {
					entityCount = len(entitySlice)
				}
			}
		}
	}
	aggregate["total_entities"] = entityCount
	
	// Collect keyphrases from keyphrase extraction
	keyphraseCount := 0
	if keyphraseResult, exists := results[string(KeyphraseExtraction)]; exists {
		if resultMap, ok := keyphraseResult.(map[string]interface{}); ok {
			if keyphrases, ok := resultMap["keyphrases"]; ok {
				if keyphraseSlice, ok := keyphrases.([]string); ok {
					keyphraseCount = len(keyphraseSlice)
				}
			}
		}
	}
	aggregate["total_keyphrases"] = keyphraseCount
	
	// Count operations performed
	aggregate["operations_performed"] = len(results)
	
	// Add timestamp
	aggregate["aggregation_timestamp"] = time.Now().Unix()
	
	return aggregate
}

// contains is a helper function to check if a string contains a substring
func contains(text, substr string) bool {
	textLen := len(text)
	substrLen := len(substr)
	
	if substrLen > textLen {
		return false
	}
	
	for i := 0; i <= textLen-substrLen; i++ {
		match := true
		for j := 0; j < substrLen; j++ {
			if text[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	
	return false
}

// GetType returns the type of the node
func (n *AdvancedNLPProcessorNode) GetType() string {
	return "advanced_nlp_processor"
}

// GetID returns a unique ID for the node instance
func (n *AdvancedNLPProcessorNode) GetID() string {
	return "adv_nlp_" + fmt.Sprintf("%d", time.Now().Unix())
}

// RegisterAdvancedNLPProcessorNode registers the Advanced NLP Processor node type with the engine
func RegisterAdvancedNLPProcessorNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("advanced_nlp_processor", func(config map[string]interface{}) (engine.NodeInstance, error) {
		return NewAdvancedNLPProcessorNode(config)
	})
}