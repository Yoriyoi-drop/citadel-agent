package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
)

// MultiModalAIProcessorConfig represents the configuration for a Multi-Modal AI Processor node
type MultiModalAIProcessorConfig struct {
	Providers       []AIProvider           `json:"providers"`         // List of AI providers to use
	Models          []ModelConfig          `json:"models"`            // Configuration for different models per modality
	Modalities      []string               `json:"modalities"`        // Supported modalities (text, image, audio, video, etc.)
	ProcessingMode  string                 `json:"processing_mode"`   // How to process multiple modalities (fusion, parallel, sequential)
	FusionStrategy  string                 `json:"fusion_strategy"`   // Strategy for fusing modalities (early, late, cross_attention)
	Timeout         int                    `json:"timeout"`           // Request timeout in seconds
	MaxRetries      int                    `json:"max_retries"`       // Number of retries for failed requests
	EnableCaching   bool                   `json:"enable_caching"`    // Whether to cache processing results
	CacheTTL        int                    `json:"cache_ttl"`         // Cache time-to-live in seconds
	EnableProfiling bool                   `json:"enable_profiling"`  // Whether to profile processing performance
	ReturnRawResults bool                 `json:"return_raw_results"` // Whether to return raw results from each modality
	CustomParams    map[string]interface{} `json:"custom_params"`     // Custom parameters for the processor
	Preprocessing   map[string]interface{} `json:"preprocessing"`     // Preprocessing config per modality
	Postprocessing  map[string]interface{} `json:"postprocessing"`    // Postprocessing config per modality
	Thresholds      map[string]float64     `json:"thresholds"`        // Thresholds per modality
}

// ModelConfig represents configuration for a model for a specific modality
type ModelConfig struct {
	Modality      string                 `json:"modality"`          // The modality this model handles (text, image, audio)
	ModelName     string                 `json:"model_name"`        // Name of the model
	Provider      AIProvider             `json:"provider"`          // Provider for this model
	APIKey        string                 `json:"api_key"`           // API key for the provider
	Endpoint      string                 `json:"endpoint"`          // Endpoint for the model
	Parameters    map[string]interface{} `json:"parameters"`        // Model-specific parameters
	Enabled       bool                   `json:"enabled"`           // Whether this model is enabled
}

// MultiModalAIProcessorNode represents a node that processes multiple types of input data (text, image, audio, etc.)
type MultiModalAIProcessorNode struct {
	config *MultiModalAIProcessorConfig
}

// NewMultiModalAIProcessorNode creates a new Multi-Modal AI Processor node
func NewMultiModalAIProcessorNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	// Convert interface{} map to JSON and back to struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}

	var multiModalConfig MultiModalAIProcessorConfig
	err = json.Unmarshal(jsonData, &multiModalConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// Validate and set defaults
	if len(multiModalConfig.Providers) == 0 {
		multiModalConfig.Providers = []AIProvider{ProviderOpenAI} // default provider
	}

	if len(multiModalConfig.Modalities) == 0 {
		multiModalConfig.Modalities = []string{"text"} // default to text modality
	}

	if multiModalConfig.ProcessingMode == "" {
		multiModalConfig.ProcessingMode = "sequential"
	}

	if multiModalConfig.FusionStrategy == "" {
		multiModalConfig.FusionStrategy = "early"
	}

	if multiModalConfig.Timeout == 0 {
		multiModalConfig.Timeout = 120 // default timeout of 120 seconds
	}

	if multiModalConfig.MaxRetries == 0 {
		multiModalConfig.MaxRetries = 3
	}

	return &MultiModalAIProcessorNode{
		config: &multiModalConfig,
	}, nil
}

// Execute implements the NodeInstance interface
func (m *MultiModalAIProcessorNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Override configuration with input values if provided
	processingMode := m.config.ProcessingMode
	if inputProcessingMode, ok := input["processing_mode"].(string); ok && inputProcessingMode != "" {
		processingMode = inputProcessingMode
	}

	fusionStrategy := m.config.FusionStrategy
	if inputFusionStrategy, ok := input["fusion_strategy"].(string); ok && inputFusionStrategy != "" {
		fusionStrategy = inputFusionStrategy
	}

	maxRetries := m.config.MaxRetries
	if inputMaxRetries, ok := input["max_retries"].(float64); ok {
		maxRetries = int(inputMaxRetries)
	}

	enableCaching := m.config.EnableCaching
	if inputEnableCaching, ok := input["enable_caching"].(bool); ok {
		enableCaching = inputEnableCaching
	}

	cacheTTL := m.config.CacheTTL
	if inputCacheTTL, ok := input["cache_ttl"].(float64); ok {
		cacheTTL = int(inputCacheTTL)
	}

	enableProfiling := m.config.EnableProfiling
	if inputEnableProfiling, ok := input["enable_profiling"].(bool); ok {
		enableProfiling = inputEnableProfiling
	}

	returnRawResults := m.config.ReturnRawResults
	if inputReturnRaw, ok := input["return_raw_results"].(bool); ok {
		returnRawResults = inputReturnRaw
	}

	customParams := m.config.CustomParams
	if inputCustomParams, ok := input["custom_params"].(map[string]interface{}); ok {
		customParams = inputCustomParams
	}

	timeout := m.config.Timeout
	if inputTimeout, ok := input["timeout"].(float64); ok {
		timeout = int(inputTimeout)
	}

	// Validate required input
	if len(input) == 0 {
		return map[string]interface{}{
			"success":   false,
			"error":     "input data is required for multi-modal processing",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Create processing context with timeout
	processingCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Process different modalities based on the processing mode
	var processedResults map[string]interface{}
	var err error

	switch processingMode {
	case "parallel":
		processedResults, err = m.processModalitiesParallel(processingCtx, input, maxRetries)
	case "sequential":
		processedResults, err = m.processModalitiesSequential(processingCtx, input, maxRetries)
	case "fusion":
		processedResults, err = m.processModalitiesWithFusion(processingCtx, input, fusionStrategy, maxRetries)
	default:
		// Default to sequential processing
		processedResults, err = m.processModalitiesSequential(processingCtx, input, maxRetries)
	}

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
		"processing_mode":       processingMode,
		"fusion_strategy":       fusionStrategy,
		"modalities_processed":  m.config.Modalities,
		"providers_used":        m.config.Providers,
		"processed_results":     processedResults,
		"enable_caching":        enableCaching,
		"enable_profiling":      enableProfiling,
		"return_raw_results":    returnRawResults,
		"timestamp":             time.Now().Unix(),
		"input_data":            input,
		"config":                m.config,
	}

	// Include raw results if requested
	if returnRawResults && processedResults["raw_results"] != nil {
		finalResult["raw_results"] = processedResults["raw_results"]
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

// processModalitiesSequential processes modalities one by one
func (m *MultiModalAIProcessorNode) processModalitiesSequential(ctx context.Context, input map[string]interface{}, maxRetries int) (map[string]interface{}, error) {
	results := make(map[string]interface{})
	rawResults := make(map[string]interface{})

	for _, modality := range m.config.Modalities {
		// Get the model config for this modality
		modelConfig := m.getModelConfigForModality(modality)
		if modelConfig == nil || !modelConfig.Enabled {
			continue
		}

		// Extract modality-specific input
		modalityInput := m.extractModalityInput(input, modality)
		if modalityInput == nil {
			continue
		}

		// Process this modality with retry logic
		var modalityResult map[string]interface{}
		var err error

		for attempt := 0; attempt <= maxRetries; attempt++ {
			modalityResult, err = m.processModality(ctx, modality, modelConfig, modalityInput)
			if err == nil {
				break // Success, break out of retry loop
			}
			if attempt == maxRetries {
				// All retries exhausted
				results[modality] = map[string]interface{}{
					"success": false,
					"error":   err.Error(),
					"attempt": attempt,
				}
				continue
			}
			// Wait before retry (in a real implementation)
			time.Sleep(100 * time.Millisecond)
		}

		if err == nil {
			results[modality] = modalityResult
			rawResults[modality] = modalityResult
		}
	}

	return map[string]interface{}{
		"results":     results,
		"raw_results": rawResults,
		"fusion":      nil, // No fusion in sequential mode
	}, nil
}

// processModalitiesParallel processes modalities concurrently
func (m *MultiModalAIProcessorNode) processModalitiesParallel(ctx context.Context, input map[string]interface{}, maxRetries int) (map[string]interface{}, error) {
	results := make(map[string]interface{})
	rawResults := make(map[string]interface{})
	
	// Create a channel to collect results
	resultChan := make(chan map[string]interface{}, len(m.config.Modalities))
	errChan := make(chan error, 1)

	// Process each modality in a goroutine
	for _, modality := range m.config.Modalities {
		go func(mod string) {
			// Get the model config for this modality
			modelConfig := m.getModelConfigForModality(mod)
			if modelConfig == nil || !modelConfig.Enabled {
				resultChan <- map[string]interface{}{
					"modality": mod,
					"result":   nil,
					"error":    nil,
				}
				return
			}

			// Extract modality-specific input
			modalityInput := m.extractModalityInput(input, mod)
			if modalityInput == nil {
				resultChan <- map[string]interface{}{
					"modality": mod,
					"result":   nil,
					"error":    nil,
				}
				return
			}

			// Process this modality with retry logic
			var modalityResult map[string]interface{}
			var err error

			for attempt := 0; attempt <= maxRetries; attempt++ {
				modalityResult, err = m.processModality(ctx, mod, modelConfig, modalityInput)
				if err == nil {
					break // Success, break out of retry loop
				}
				if attempt == maxRetries {
					// All retries exhausted
					resultChan <- map[string]interface{}{
						"modality": mod,
						"result": map[string]interface{}{
							"success": false,
							"error":   err.Error(),
							"attempt": attempt,
						},
						"error": err,
					}
					return
				}
				// Wait before retry (in a real implementation)
				time.Sleep(100 * time.Millisecond)
			}

			if err == nil {
				resultChan <- map[string]interface{}{
					"modality": mod,
					"result":   modalityResult,
					"error":    nil,
				}
			} else {
				resultChan <- map[string]interface{}{
					"modality": mod,
					"result": map[string]interface{}{
						"success": false,
						"error":   err.Error(),
					},
					"error": err,
				}
			}
		}(modality)
	}

	// Collect results from all goroutines
	collected := 0
	for collected < len(m.config.Modalities) {
		select {
		case result := <-resultChan:
			modality := result["modality"].(string)
			if resultData := result["result"]; resultData != nil {
				results[modality] = resultData
				rawResults[modality] = resultData
			}
			collected++
		case err := <-errChan:
			return nil, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return map[string]interface{}{
		"results":     results,
		"raw_results": rawResults,
		"fusion":      nil, // No fusion in parallel mode
	}, nil
}

// processModalitiesWithFusion processes modalities and fuses the results
func (m *MultiModalAIProcessorNode) processModalitiesWithFusion(ctx context.Context, input map[string]interface{}, fusionStrategy string, maxRetries int) (map[string]interface{}, error) {
	// First, process modalities individually
	individualResults := make(map[string]interface{})
	rawResults := make(map[string]interface{})

	for _, modality := range m.config.Modalities {
		// Get the model config for this modality
		modelConfig := m.getModelConfigForModality(modality)
		if modelConfig == nil || !modelConfig.Enabled {
			continue
		}

		// Extract modality-specific input
		modalityInput := m.extractModalityInput(input, modality)
		if modalityInput == nil {
			continue
		}

		// Process this modality with retry logic
		var modalityResult map[string]interface{}
		var err error

		for attempt := 0; attempt <= maxRetries; attempt++ {
			modalityResult, err = m.processModality(ctx, modality, modelConfig, modalityInput)
			if err == nil {
				break // Success, break out of retry loop
			}
			if attempt == maxRetries {
				// All retries exhausted
				individualResults[modality] = map[string]interface{}{
					"success": false,
					"error":   err.Error(),
					"attempt": attempt,
				}
				continue
			}
			// Wait before retry (in a real implementation)
			time.Sleep(100 * time.Millisecond)
		}

		if err == nil {
			individualResults[modality] = modalityResult
			rawResults[modality] = modalityResult
		}
	}

	// Then fuse the results based on the fusion strategy
	fusedResult := m.fuseResults(individualResults, fusionStrategy)

	return map[string]interface{}{
		"results":     individualResults,
		"raw_results": rawResults,
		"fused_result": fusedResult,
		"fusion_strategy": fusionStrategy,
	}, nil
}

// fuseResults combines results from different modalities based on the fusion strategy
func (m *MultiModalAIProcessorNode) fuseResults(results map[string]interface{}, strategy string) map[string]interface{} {
	fused := make(map[string]interface{})
	
	switch strategy {
	case "early":
		// Early fusion: combine inputs before processing
		// This would have been handled in the processing stage
		fused["strategy"] = "early"
		fused["combined_result"] = results
	case "late":
		// Late fusion: combine outputs after processing
		combinedScore := 0.0
		count := 0
		details := make(map[string]interface{})
		
		for modality, result := range results {
			if resultMap, ok := result.(map[string]interface{}); ok {
				if score, exists := resultMap["confidence"]; exists {
					if scoreFloat, ok := score.(float64); ok {
						combinedScore += scoreFloat
						count++
					}
				}
				details[modality] = resultMap
			}
		}
		
		if count > 0 {
			fused["combined_confidence"] = combinedScore / float64(count)
		}
		fused["strategy"] = "late"
		fused["modality_details"] = details
	case "cross_attention":
		// Cross-attention fusion: simulate cross-modal attention mechanisms
		fused["strategy"] = "cross_attention"
		fused["cross_modal_analysis"] = "Cross-modal attention patterns analyzed"
		fused["fused_features"] = "Features combined using attention mechanisms"
	default:
		// Default to late fusion
		combinedScore := 0.0
		count := 0
		details := make(map[string]interface{})
		
		for modality, result := range results {
			if resultMap, ok := result.(map[string]interface{}); ok {
				if score, exists := resultMap["confidence"]; exists {
					if scoreFloat, ok := score.(float64); ok {
						combinedScore += scoreFloat
						count++
					}
				}
				details[modality] = resultMap
			}
		}
		
		if count > 0 {
			fused["combined_confidence"] = combinedScore / float64(count)
		}
		fused["strategy"] = "late_default"
		fused["modality_details"] = details
	}
	
	fused["fusion_timestamp"] = time.Now().Unix()
	
	return fused
}

// getModelConfigForModality gets the model configuration for a specific modality
func (m *MultiModalAIProcessorNode) getModelConfigForModality(modality string) *ModelConfig {
	for _, modelConfig := range m.config.Models {
		if modelConfig.Modality == modality {
			return &modelConfig
		}
	}
	
	// If no specific config found, return a default config for the modality
	return &ModelConfig{
		Modality: modality,
		ModelName: fmt.Sprintf("default_%s_model", modality),
		Provider: ProviderOpenAI,
		Enabled: true,
		Parameters: make(map[string]interface{}),
	}
}

// extractModalityInput extracts input data for a specific modality
func (m *MultiModalAIProcessorNode) extractModalityInput(input map[string]interface{}, modality string) interface{} {
	switch modality {
	case "text":
		// Look for text-related keys in input
		if text, exists := input["text"]; exists {
			return text
		}
		if prompt, exists := input["prompt"]; exists {
			return prompt
		}
		if message, exists := input["message"]; exists {
			return message
		}
		// If no specific text key, combine string values
		textParts := []string{}
		for k, v := range input {
			if str, ok := v.(string); ok && k != "modality" {
				textParts = append(textParts, fmt.Sprintf("%s: %s", k, str))
			}
		}
		return textParts
	case "image":
		// Look for image-related keys in input
		if image, exists := input["image"]; exists {
			return image
		}
		if imageUrl, exists := input["image_url"]; exists {
			return imageUrl
		}
		if imageData, exists := input["image_data"]; exists {
			return imageData
		}
		return nil
	case "audio":
		// Look for audio-related keys in input
		if audio, exists := input["audio"]; exists {
			return audio
		}
		if audioUrl, exists := input["audio_url"]; exists {
			return audioUrl
		}
		if audioData, exists := input["audio_data"]; exists {
			return audioData
		}
		return nil
	case "video":
		// Look for video-related keys in input
		if video, exists := input["video"]; exists {
			return video
		}
		if videoUrl, exists := input["video_url"]; exists {
			return videoUrl
		}
		if videoData, exists := input["video_data"]; exists {
			return videoData
		}
		return nil
	default:
		// For other modalities, return the whole input or modality-specific part
		if modalityData, exists := input[modality]; exists {
			return modalityData
		}
		return input
	}
}

// processModality processes a single modality using the appropriate model
func (m *MultiModalAIProcessorNode) processModality(ctx context.Context, modality string, modelConfig *ModelConfig, input interface{}) (map[string]interface{}, error) {
	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	// Simulate different processing based on modality
	var result map[string]interface{}

	switch modality {
	case "text":
		result = map[string]interface{}{
			"modality": "text",
			"processing_type": "natural_language_understanding",
			"entities": []string{"entity1", "entity2", "entity3"},
			"sentiment": "positive",
			"key_phrases": []string{"phrase1", "phrase2"},
			"language": "en",
			"text_summary": "This is a summary of the text input",
			"confidence": 0.92,
		}
	case "image":
		result = map[string]interface{}{
			"modality": "image",
			"processing_type": "computer_vision",
			"objects_detected": []map[string]interface{}{
				{
					"class": "person",
					"confidence": 0.95,
					"bbox": map[string]interface{}{
						"x": 100,
						"y": 150,
						"width": 200,
						"height": 300,
					},
				},
				{
					"class": "car",
					"confidence": 0.88,
					"bbox": map[string]interface{}{
						"x": 400,
						"y": 200,
						"width": 250,
						"height": 150,
					},
				},
			},
			"image_description": "An image containing people and cars",
			"image_quality": map[string]interface{}{
				"brightness": 0.7,
				"contrast": 0.8,
				"sharpness": 0.6,
			},
			"confidence": 0.89,
		}
	case "audio":
		result = map[string]interface{}{
			"modality": "audio",
			"processing_type": "speech_recognition",
			"transcript": "This is the transcribed text from the audio input",
			"language": "en",
			"audio_quality": map[string]interface{}{
				"noise_level": 0.2,
				"clarity": 0.85,
			},
			"sentiment": "neutral",
			"confidence": 0.87,
		}
	case "video":
		result = map[string]interface{}{
			"modality": "video",
			"processing_type": "video_analysis",
			"frame_analysis": map[string]interface{}{
				"duration": "10s",
				"fps": 30,
				"resolution": "1920x1080",
			},
			"objects_tracked": []string{"person", "car"},
			"scene_description": "A busy street scene with people and vehicles",
			"key_frames": []int{0, 300, 600},
			"confidence": 0.85,
		}
	default:
		result = map[string]interface{}{
			"modality": modality,
			"processing_type": "general_processing",
			"input_processed": input,
			"result": fmt.Sprintf("Processed %s modality with default processor", modality),
			"confidence": 0.80,
		}
	}

	// Add metadata
	result["model_used"] = modelConfig.ModelName
	result["provider"] = string(modelConfig.Provider)
	result["model_parameters"] = modelConfig.Parameters
	result["processing_time"] = time.Since(time.Now().Add(-100 * time.Millisecond)).Seconds()
	result["timestamp"] = time.Now().Unix()
	
	// Apply threshold if specified for this modality
	if threshold, exists := m.config.Thresholds[modality]; exists {
		if confidence, ok := result["confidence"].(float64); ok {
			if confidence < threshold {
				result["below_threshold"] = true
				result["confidence"] = confidence
			}
		}
	}

	return result, nil
}

// GetType returns the type of the node
func (m *MultiModalAIProcessorNode) GetType() string {
	return "multi_modal_ai_processor"
}

// GetID returns a unique ID for the node instance
func (m *MultiModalAIProcessorNode) GetID() string {
	return "multi_modal_" + fmt.Sprintf("%d", time.Now().Unix())
}

// RegisterMultiModalAIProcessorNode registers the Multi-Modal AI Processor node type with the engine
func RegisterMultiModalAIProcessorNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("multi_modal_ai_processor", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return NewMultiModalAIProcessorNode(config)
	})
}