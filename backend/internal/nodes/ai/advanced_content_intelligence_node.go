package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
	"github.com/citadel-agent/backend/internal/workflow/core/engine"
)

// AdvancedContentIntelligenceConfig mewakili konfigurasi untuk node Advanced Content Intelligence
type AdvancedContentIntelligenceConfig struct {
	Provider          string                 `json:"provider"`           // Penyedia layanan
	ContentType       string                 `json:"content_type"`       // Jenis konten (text, image, video, audio, document)
	AnalysisMode      string                 `json:"analysis_mode"`      // Mode analisis (classification, sentiment, extraction)
	ModelName         string                 `json:"model_name"`         // Nama model yang digunakan
	APIKey            string                 `json:"api_key"`            // Kunci API untuk layanan
	MaxResults        int                    `json:"max_results"`        // Jumlah maksimum hasil
	EnableCategorization bool                `json:"enable_categorization"` // Apakah mengaktifkan kategorisasi
	EnableSentiment   bool                   `json:"enable_sentiment"`   // Apakah mengaktifkan analisis sentimen
	EnableSummarization bool                 `json:"enable_summarization"` // Apakah mengaktifkan ringkasan
	EnableTopicModeling bool                 `json:"enable_topic_modeling"` // Apakah mengaktifkan pemodelan topik
	EnableEntityExtraction bool             `json:"enable_entity_extraction"` // Apakah mengaktifkan ekstraksi entitas
	EnableQualityAssessment bool            `json:"enable_quality_assessment"` // Apakah mengaktifkan penilaian kualitas
	QualityThreshold  float64                `json:"quality_threshold"`  // Ambang batas kualitas konten
	Language          string                 `json:"language"`           // Bahasa konten
	TargetLanguages   []string               `json:"target_languages"`   // Bahasa target untuk analisis
	Categories        []string               `json:"categories"`         // Kategori yang didukung
	EnableAutoModeration bool                `json:"enable_auto_moderation"` // Apakah mengaktifkan moderasi otomatis
	FlaggedCategories []string               `json:"flagged_categories"` // Kategori yang perlu ditandai
	EnableCaching     bool                   `json:"enable_caching"`     // Apakah mengaktifkan caching hasil
	CacheTTL          int                    `json:"cache_ttl"`          // Waktu cache dalam detik
	EnableProfiling   bool                   `json:"enable_profiling"`   // Apakah mengaktifkan profiling
	Timeout           int                    `json:"timeout"`            // Waktu timeout dalam detik
	ReturnRawResults  bool                   `json:"return_raw_results"` // Apakah mengembalikan hasil mentah
	CustomParams      map[string]interface{} `json:"custom_params"`      // Parameter khusus untuk analisis
	Preprocessing     PreprocessingConfig    `json:"preprocessing"`      // Konfigurasi pra-pemrosesan
	Postprocessing    PostprocessingConfig   `json:"postprocessing"`     // Konfigurasi pasca-pemrosesan
	ContentFilters    ContentFilters         `json:"content_filters"`    // Filter untuk konten
}

// ContentFilters mewakili filter untuk analisis konten
type ContentFilters struct {
	ProfanityFilter   bool     `json:"profanity_filter"`     // Filter kata-kata tidak pantas
	SpamFilter        bool     `json:"spam_filter"`          // Filter spam
	PlagiarismCheck   bool     `json:"plagiarism_check"`     // Pemeriksa plagiarisme
	CopyrightCheck    bool     `json:"copyright_check"`      // Pemeriksaan hak cipta
	RelevanceThreshold float64 `json:"relevance_threshold"`  // Ambang batas relevansi
	MinLength         int      `json:"min_length"`          // Panjang minimum konten
	MaxLength         int      `json:"max_length"`          // Panjang maksimum konten
}

// ContentAnalysisResult mewakili hasil analisis konten
type ContentAnalysisResult struct {
	ContentID        string                 `json:"content_id"`
	ContentType      string                 `json:"content_type"`
	AnalysisMode     string                 `json:"analysis_mode"`
	Categories       []CategoryResult       `json:"categories"`
	Sentiment        SentimentResult        `json:"sentiment"`
	Summary          string                 `json:"summary"`
	Topics           []TopicResult          `json:"topics"`
	Entities         []NEResult             `json:"entities"`
	QualityScore     float64                `json:"quality_score"`
	IsAppropriate    bool                   `json:"is_appropriate"`
	FlaggedIssues    []string               `json:"flagged_issues"`
	ProcessingTime   float64                `json:"processing_time"`
	Timestamp        int64                  `json:"timestamp"`
	RawResults       interface{}            `json:"raw_results,omitempty"`
}

// CategoryResult mewakili hasil kategorisasi
type CategoryResult struct {
	Category string  `json:"category"`
	Score    float64 `json:"score"`
	Primary  bool    `json:"primary"`
}

// SentimentResult mewakili hasil analisis sentimen
type SentimentResult struct {
	Sentiment string  `json:"sentiment"`
	Score     float64 `json:"score"`
	Confidence float64 `json:"confidence"`
}

// TopicResult mewakili hasil pemodelan topik
type TopicResult struct {
	TopicID   int     `json:"topic_id"`
	Name      string  `json:"name"`
	Score     float64 `json:"score"`
	Keywords  []string `json:"keywords"`
}

// NEResult mewakili hasil ekstraksi entitas nama
type NEResult struct {
	Entity  string  `json:"entity"`
	Type    string  `json:"type"`
	Confidence float64 `json:"confidence"`
	Start   int     `json:"start"`
	End     int     `json:"end"`
}

// AdvancedContentIntelligenceNode mewakili node yang melakukan analisis kecerdasan konten canggih
type AdvancedContentIntelligenceNode struct {
	config *AdvancedContentIntelligenceConfig
}

// NewAdvancedContentIntelligenceNode membuat node Advanced Content Intelligence baru
func NewAdvancedContentIntelligenceNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	// Konversi map interface{} ke JSON lalu ke struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("gagal mengubah konfig menjadi json: %v", err)
	}

	var contentConfig AdvancedContentIntelligenceConfig
	err = json.Unmarshal(jsonData, &contentConfig)
	if err != nil {
		return nil, fmt.Errorf("gagal menguraikan konfig: %v", err)
	}

	// Validasi dan atur default
	if contentConfig.AnalysisMode == "" {
		contentConfig.AnalysisMode = "classification"
	}

	if contentConfig.ContentType == "" {
		contentConfig.ContentType = "text"
	}

	if contentConfig.MaxResults == 0 {
		contentConfig.MaxResults = 10
	}

	if contentConfig.Language == "" {
		contentConfig.Language = "en"
	}

	if contentConfig.QualityThreshold == 0 {
		contentConfig.QualityThreshold = 0.7
	}

	if contentConfig.Timeout == 0 {
		contentConfig.Timeout = 120 // default timeout 120 detik
	}

	if len(contentConfig.Categories) == 0 {
		contentConfig.Categories = []string{"technology", "business", "health", "entertainment", "sports", "education"}
	}

	return &AdvancedContentIntelligenceNode{
		config: &contentConfig,
	}, nil
}

// Execute mengimplementasikan interface NodeInstance
func (c *AdvancedContentIntelligenceNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Timpa konfigurasi dengan nilai input jika disediakan
	provider := c.config.Provider
	if inputProvider, ok := input["provider"].(string); ok && inputProvider != "" {
		provider = inputProvider
	}

	contentType := c.config.ContentType
	if inputContentType, ok := input["content_type"].(string); ok && inputContentType != "" {
		contentType = inputContentType
	}

	analysisMode := c.config.AnalysisMode
	if inputAnalysisMode, ok := input["analysis_mode"].(string); ok && inputAnalysisMode != "" {
		analysisMode = inputAnalysisMode
	}

	modelName := c.config.ModelName
	if inputModelName, ok := input["model_name"].(string); ok && inputModelName != "" {
		modelName = inputModelName
	}

	apiKey := c.config.APIKey
	if inputAPIKey, ok := input["api_key"].(string); ok && inputAPIKey != "" {
		apiKey = inputAPIKey
	}

	maxResults := c.config.MaxResults
	if inputMaxResults, ok := input["max_results"].(float64); ok {
		maxResults = int(inputMaxResults)
	}

	enableCategorization := c.config.EnableCategorization
	if inputEnableCategorization, ok := input["enable_categorization"].(bool); ok {
		enableCategorization = inputEnableCategorization
	}

	enableSentiment := c.config.EnableSentiment
	if inputEnableSentiment, ok := input["enable_sentiment"].(bool); ok {
		enableSentiment = inputEnableSentiment
	}

	enableSummarization := c.config.EnableSummarization
	if inputEnableSummarization, ok := input["enable_summarization"].(bool); ok {
		enableSummarization = inputEnableSummarization
	}

	enableTopicModeling := c.config.EnableTopicModeling
	if inputEnableTopicModeling, ok := input["enable_topic_modeling"].(bool); ok {
		enableTopicModeling = inputEnableTopicModeling
	}

	enableEntityExtraction := c.config.EnableEntityExtraction
	if inputEnableEntityExtraction, ok := input["enable_entity_extraction"].(bool); ok {
		enableEntityExtraction = inputEnableEntityExtraction
	}

	enableQualityAssessment := c.config.EnableQualityAssessment
	if inputEnableQualityAssessment, ok := input["enable_quality_assessment"].(bool); ok {
		enableQualityAssessment = inputEnableQualityAssessment
	}

	qualityThreshold := c.config.QualityThreshold
	if inputQualityThreshold, ok := input["quality_threshold"].(float64); ok {
		qualityThreshold = inputQualityThreshold
	}

	language := c.config.Language
	if inputLanguage, ok := input["language"].(string); ok && inputLanguage != "" {
		language = inputLanguage
	}

	targetLanguages := c.config.TargetLanguages
	if inputTargetLanguages, ok := input["target_languages"].([]interface{}); ok {
		targetLanguages = make([]string, len(inputTargetLanguages))
		for i, val := range inputTargetLanguages {
			targetLanguages[i] = fmt.Sprintf("%v", val)
		}
	}

	categories := c.config.Categories
	if inputCategories, ok := input["categories"].([]interface{}); ok {
		categories = make([]string, len(inputCategories))
		for i, val := range inputCategories {
			categories[i] = fmt.Sprintf("%v", val)
		}
	}

	enableAutoModeration := c.config.EnableAutoModeration
	if inputEnableAutoModeration, ok := input["enable_auto_moderation"].(bool); ok {
		enableAutoModeration = inputEnableAutoModeration
	}

	flaggedCategories := c.config.FlaggedCategories
	if inputFlaggedCategories, ok := input["flagged_categories"].([]interface{}); ok {
		flaggedCategories = make([]string, len(inputFlaggedCategories))
		for i, val := range inputFlaggedCategories {
			flaggedCategories[i] = fmt.Sprintf("%v", val)
		}
	}

	enableCaching := c.config.EnableCaching
	if inputEnableCaching, ok := input["enable_caching"].(bool); ok {
		enableCaching = inputEnableCaching
	}

	cacheTTL := c.config.CacheTTL
	if inputCacheTTL, ok := input["cache_ttl"].(float64); ok {
		cacheTTL = int(inputCacheTTL)
	}

	enableProfiling := c.config.EnableProfiling
	if inputEnableProfiling, ok := input["enable_profiling"].(bool); ok {
		enableProfiling = inputEnableProfiling
	}

	timeout := c.config.Timeout
	if inputTimeout, ok := input["timeout"].(float64); ok {
		timeout = int(inputTimeout)
	}

	returnRawResults := c.config.ReturnRawResults
	if inputReturnRaw, ok := input["return_raw_results"].(bool); ok {
		returnRawResults = inputReturnRaw
	}

	customParams := c.config.CustomParams
	if inputCustomParams, ok := input["custom_params"].(map[string]interface{}); ok {
		customParams = inputCustomParams
	}

	// Validasi input
	if len(input) == 0 {
		return map[string]interface{}{
			"success":   false,
			"error":     "input diperlukan untuk analisis kecerdasan konten",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	if apiKey == "" {
		return map[string]interface{}{
			"success":   false,
			"error":     "api_key diperlukan untuk analisis kecerdasan konten",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Buat konteks analisis dengan timeout
	contentCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Lakukan analisis kecerdasan konten
	contentResult, err := c.performContentIntelligence(contentCtx, input)
	if err != nil {
		return map[string]interface{}{
			"success":   false,
			"error":     err.Error(),
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Siapkan hasil akhir
	finalResult := map[string]interface{}{
		"success":                    true,
		"provider":                   provider,
		"content_type":               contentType,
		"analysis_mode":              analysisMode,
		"model_name":                 modelName,
		"max_results":                maxResults,
		"enable_categorization":      enableCategorization,
		"enable_sentiment":           enableSentiment,
		"enable_summarization":       enableSummarization,
		"enable_topic_modeling":      enableTopicModeling,
		"enable_entity_extraction":   enableEntityExtraction,
		"enable_quality_assessment":  enableQualityAssessment,
		"quality_threshold":          qualityThreshold,
		"language":                   language,
		"target_languages":           targetLanguages,
		"categories":                 categories,
		"enable_auto_moderation":     enableAutoModeration,
		"flagged_categories":         flaggedCategories,
		"content_result":             contentResult,
		"enable_caching":             enableCaching,
		"enable_profiling":           enableProfiling,
		"return_raw_results":         returnRawResults,
		"timestamp":                  time.Now().Unix(),
		"input_data":                 input,
		"config":                     c.config,
	}

	// Tambahkan hasil mentah jika diminta
	if returnRawResults && contentResult.RawResults != nil {
		finalResult["raw_results"] = contentResult.RawResults
	}

	// Tambahkan metrik kinerja jika profiling diaktifkan
	if enableProfiling {
		finalResult["performance_metrics"] = map[string]interface{}{
			"start_time": time.Now().Unix(),
			"end_time":   time.Now().Unix(),
			"duration":   contentResult.ProcessingTime,
		}
	}

	return finalResult, nil
}

// performContentIntelligence melakukan analisis kecerdasan konten
func (c *AdvancedContentIntelligenceNode) performContentIntelligence(ctx context.Context, input map[string]interface{}) (*ContentAnalysisResult, error) {
	// Simulasikan waktu pemrosesan
	startTime := time.Now()
	time.Sleep(200 * time.Millisecond)

	// Ambil konten dari input
	contentText := ""
	if text, exists := input["text"]; exists {
		if textStr, ok := text.(string); ok {
			contentText = textStr
		}
	} else if content, exists := input["content"]; exists {
		if contentStr, ok := content.(string); ok {
			contentText = contentStr
		}
	} else if title, exists := input["title"]; exists {
		if titleStr, ok := title.(string); ok {
			contentText = titleStr
		}
	}

	// Buat ID konten berdasarkan timestamp jika tidak ada
	contentID := fmt.Sprintf("content_%d", time.Now().UnixNano())
	if id, exists := input["content_id"]; exists {
		if idStr, ok := id.(string); ok {
			contentID = idStr
		}
	}

	// Simulasikan kategorisasi
	var categories []CategoryResult
	if c.config.EnableCategorization {
		// Pilih kategori berdasarkan panjang teks atau kata kunci
		selectedCategories := []string{"technology", "business", "health"}
		for i, cat := range selectedCategories {
			score := 0.8 - (float64(i) * 0.1)
			isPrimary := i == 0
			categories = append(categories, CategoryResult{
				Category: cat,
				Score:    score,
				Primary:  isPrimary,
			})
		}
	}

	// Simulasikan analisis sentimen
	sentiment := SentimentResult{
		Sentiment: "positive",
		Score:     0.75,
		Confidence: 0.88,
	}
	if c.config.EnableSentiment {
		// Analisis sentimen berdasarkan kata kunci
		contentLower := ""
		for _, r := range contentText {
			contentLower += string(r | 32) // Konversi ke huruf kecil
		}
		
		positiveWords := []string{"good", "great", "excellent", "love", "amazing"}
		negativeWords := []string{"bad", "terrible", "awful", "hate", "poor"}
		
		positiveCount := 0
		negativeCount := 0
		
		for _, word := range positiveWords {
			if contains(contentLower, word) {
				positiveCount++
			}
		}
		
		for _, word := range negativeWords {
			if contains(contentLower, word) {
				negativeCount++
			}
		}
		
		if positiveCount > negativeCount {
			sentiment.Sentiment = "positive"
			sentiment.Score = 0.6 + float64(positiveCount-negativeCount)*0.1
		} else if negativeCount > positiveCount {
			sentiment.Sentiment = "negative"
			sentiment.Score = 0.4 - float64(negativeCount-positiveCount)*0.1
		} else {
			sentiment.Sentiment = "neutral"
			sentiment.Score = 0.5
		}
		
		if sentiment.Score > 1.0 {
			sentiment.Score = 1.0
		}
		if sentiment.Score < 0.0 {
			sentiment.Score = 0.0
		}
	}

	// Simulasikan ringkasan
	summary := "Ini adalah ringkasan dari konten yang dianalisis. Konten ini membahas berbagai topik penting yang relevan dengan kategori yang ditentukan."
	if c.config.EnableSummarization {
		if len(contentText) > 0 {
			// Buat ringkasan berdasarkan beberapa kalimat pertama
			summary = contentText
			if len(summary) > 200 {
				// Potong ke 200 karakter dan tambahkan elipsis
				summary = summary[:200] + "..."
			}
		}
	}

	// Simulasikan pemodelan topik
	var topics []TopicResult
	if c.config.EnableTopicModeling {
		topics = []TopicResult{
			{
				TopicID:  1,
				Name:     "Artificial Intelligence",
				Score:    0.78,
				Keywords: []string{"AI", "machine learning", "neural networks", "algorithms"},
			},
			{
				TopicID:  2,
				Name:     "Technology Trends",
				Score:    0.65,
				Keywords: []string{"innovation", "digital transformation", "emerging tech"},
			},
		}
	}

	// Simulasikan ekstraksi entitas
	var entities []NEResult
	if c.config.EnableEntityExtraction {
		entities = []NEResult{
			{
				Entity:     "John Doe",
				Type:       "PERSON",
				Confidence: 0.92,
				Start:      0,
				End:        8,
			},
			{
				Entity:     "New York",
				Type:       "GPE",
				Confidence: 0.88,
				Start:      25,
				End:        33,
			},
			{
				Entity:     "Microsoft",
				Type:       "ORG",
				Confidence: 0.91,
				Start:      50,
				End:        59,
			},
		}
	}

	// Simulasikan penilaian kualitas
	qualityScore := 0.8
	if c.config.EnableQualityAssessment {
		// Hitung skor kualitas berdasarkan panjang konten dan struktur
		if len(contentText) < 100 {
			qualityScore = 0.4
		} else if len(contentText) < 500 {
			qualityScore = 0.7
		} else {
			qualityScore = 0.9
		}
	}

	// Simulasikan apakah konten sesuai
	isAppropriate := qualityScore >= c.config.QualityThreshold

	// Simulasikan masalah yang ditandai
	var flaggedIssues []string
	if c.config.EnableAutoModeration {
		// Cek apakah konten memiliki kategori yang ditandai
		if qualityScore < c.config.QualityThreshold {
			flaggedIssues = append(flaggedIssues, "low_quality_content")
		}
		
		// Cek filter konten
		if c.config.ContentFilters.ProfanityFilter {
			if contains(contentLower, "bad") || contains(contentLower, "terrible") {
				flaggedIssues = append(flaggedIssues, "profanity_detected")
			}
		}
	}

	result := &ContentAnalysisResult{
		ContentID:      contentID,
		ContentType:    c.config.ContentType,
		AnalysisMode:   c.config.AnalysisMode,
		Categories:     categories,
		Sentiment:      sentiment,
		Summary:        summary,
		Topics:         topics,
		Entities:       entities,
		QualityScore:   qualityScore,
		IsAppropriate:  isAppropriate,
		FlaggedIssues:  flaggedIssues,
		ProcessingTime: time.Since(startTime).Seconds(),
		Timestamp:      time.Now().Unix(),
	}

	if c.config.ReturnRawResults {
		result.RawResults = map[string]interface{}{
			"original_content_length": len(contentText),
			"analysis_details": map[string]interface{}{
				"content_id": contentID,
				"analyzer_version": "1.0.0",
				"analysis_timestamp": time.Now().Unix(),
			},
		}
	}

	return result, nil
}

// GetType mengembalikan jenis node
func (c *AdvancedContentIntelligenceNode) GetType() string {
	return "advanced_content_intelligence"
}

// GetID mengembalikan ID unik untuk instance node
func (c *AdvancedContentIntelligenceNode) GetID() string {
	return "adv_content_intel_" + fmt.Sprintf("%d", time.Now().Unix())
}

// RegisterAdvancedContentIntelligenceNode mendaftarkan node Advanced Content Intelligence dengan engine
func RegisterAdvancedContentIntelligenceNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("advanced_content_intelligence", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return NewAdvancedContentIntelligenceNode(config)
	})
}