package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"citadel-agent/backend/internal/workflow/core/engine"
)

// AdvancedRecommendationEngineConfig mewakili konfigurasi untuk node Advanced Recommendation Engine
type AdvancedRecommendationEngineConfig struct {
	Provider        string                 `json:"provider"`          // Penyedia rekomendasi (custom, openai, dll.)
	Algorithm       string                 `json:"algorithm"`         // Algoritma rekomendasi (collaborative_filtering, content_based, dll.)
	ModelName       string                 `json:"model_name"`        // Nama model rekomendasi
	APIKey          string                 `json:"api_key"`           // Kunci API untuk layanan rekomendasi
	UserID          string                 `json:"user_id"`           // ID pengguna untuk rekomendasi personalisasi
	ItemType        string                 `json:"item_type"`         // Jenis item untuk direkomendasikan (product, content, service)
	ContextFeatures []string               `json:"context_features"`  // Fitur kontekstual (lokasi, waktu, perangkat)
	NumRecommendations int                 `json:"num_recommendations"` // Jumlah rekomendasi yang akan dihasilkan
	ScoreThreshold  float64                `json:"score_threshold"`   // Ambang batas skor untuk rekomendasi
	DiversityFactor float64                `json:"diversity_factor"`  // Faktor keragaman dalam rekomendasi
	FreshnessFactor float64                `json:"freshness_factor"`  // Faktor kesegaran dalam rekomendasi
	ExclusionList   []string               `json:"exclusion_list"`    // Daftar item untuk dikecualikan
	Whitelist       []string               `json:"whitelist"`         // Daftar item hanya untuk disertakan
	EnableCaching   bool                   `json:"enable_caching"`    // Apakah mengaktifkan caching rekomendasi
	CacheTTL        int                    `json:"cache_ttl"`         // Waktu cache dalam detik
	EnableProfiling bool                   `json:"enable_profiling"`  // Apakah mengaktifkan profiling
	Timeout         int                    `json:"timeout"`           // Waktu timeout dalam detik
	ReturnRawResults bool                 `json:"return_raw_results"` // Apakah mengembalikan hasil mentah
	CustomParams    map[string]interface{} `json:"custom_params"`     // Parameter khusus untuk rekomendasi
	Preprocessing   PreprocessingConfig    `json:"preprocessing"`     // Konfigurasi pra-pemrosesan
	Postprocessing  PostprocessingConfig   `json:"postprocessing"`    // Konfigurasi pasca-pemrosesan
	FeatureWeights  map[string]float64     `json:"feature_weights"`   // Bobot untuk fitur-fitur berbeda
}

// RecommendationItem mewakili item yang direkomendasikan
type RecommendationItem struct {
	ItemID     string      `json:"item_id"`
	Name       string      `json:"name"`
	Category   string      `json:"category"`
	Score      float64     `json:"score"`
	Reason     string      `json:"reason"`
	Metadata   interface{} `json:"metadata,omitempty"`
	Contextual bool        `json:"contextual"`
}

// RecommendationResult mewakili hasil rekomendasi
type RecommendationResult struct {
	Items         []RecommendationItem `json:"items"`
	UserID        string              `json:"user_id"`
	Algorithm     string              `json:"algorithm"`
	Timestamp     int64               `json:"timestamp"`
	ProcessingTime float64             `json:"processing_time"`
}

// AdvancedRecommendationEngineNode mewakili node yang menghasilkan rekomendasi canggih
type AdvancedRecommendationEngineNode struct {
	config *AdvancedRecommendationEngineConfig
}

// NewAdvancedRecommendationEngineNode membuat node Advanced Recommendation Engine baru
func NewAdvancedRecommendationEngineNode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Konversi map interface{} ke JSON lalu ke struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("gagal mengubah konfig menjadi json: %v", err)
	}

	var recConfig AdvancedRecommendationEngineConfig
	err = json.Unmarshal(jsonData, &recConfig)
	if err != nil {
		return nil, fmt.Errorf("gagal menguraikan konfig: %v", err)
	}

	// Validasi dan atur default
	if recConfig.Provider == "" {
		recConfig.Provider = "citadel"
	}

	if recConfig.Algorithm == "" {
		recConfig.Algorithm = "hybrid"
	}

	if recConfig.NumRecommendations == 0 {
		recConfig.NumRecommendations = 10
	}

	if recConfig.ScoreThreshold == 0 {
		recConfig.ScoreThreshold = 0.5
	}

	if recConfig.DiversityFactor == 0 {
		recConfig.DiversityFactor = 0.3
	}

	if recConfig.FreshnessFactor == 0 {
		recConfig.FreshnessFactor = 0.2
	}

	if recConfig.Timeout == 0 {
		recConfig.Timeout = 60 // default timeout 60 detik
	}

	return &AdvancedRecommendationEngineNode{
		config: &recConfig,
	}, nil
}

// Execute mengimplementasikan interface NodeInstance
func (r *AdvancedRecommendationEngineNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Timpa konfigurasi dengan nilai input jika disediakan
	provider := r.config.Provider
	if inputProvider, ok := input["provider"].(string); ok && inputProvider != "" {
		provider = inputProvider
	}

	algorithm := r.config.Algorithm
	if inputAlgorithm, ok := input["algorithm"].(string); ok && inputAlgorithm != "" {
		algorithm = inputAlgorithm
	}

	modelName := r.config.ModelName
	if inputModelName, ok := input["model_name"].(string); ok && inputModelName != "" {
		modelName = inputModelName
	}

	apiKey := r.config.APIKey
	if inputAPIKey, ok := input["api_key"].(string); ok && inputAPIKey != "" {
		apiKey = inputAPIKey
	}

	userID := r.config.UserID
	if inputUserID, ok := input["user_id"].(string); ok && inputUserID != "" {
		userID = inputUserID
	}

	itemType := r.config.ItemType
	if inputItemType, ok := input["item_type"].(string); ok && inputItemType != "" {
		itemType = inputItemType
	}

	contextFeatures := r.config.ContextFeatures
	if inputContextFeatures, ok := input["context_features"].([]interface{}); ok {
		contextFeatures = make([]string, len(inputContextFeatures))
		for i, val := range inputContextFeatures {
			contextFeatures[i] = fmt.Sprintf("%v", val)
		}
	}

	numRecommendations := r.config.NumRecommendations
	if inputNumRecs, ok := input["num_recommendations"].(float64); ok {
		numRecommendations = int(inputNumRecs)
	}

	scoreThreshold := r.config.ScoreThreshold
	if inputScoreThreshold, ok := input["score_threshold"].(float64); ok {
		scoreThreshold = inputScoreThreshold
	}

	diversityFactor := r.config.DiversityFactor
	if inputDiversityFactor, ok := input["diversity_factor"].(float64); ok {
		diversityFactor = inputDiversityFactor
	}

	freshnessFactor := r.config.FreshnessFactor
	if inputFreshnessFactor, ok := input["freshness_factor"].(float64); ok {
		freshnessFactor = inputFreshnessFactor
	}

	exclusionList := r.config.ExclusionList
	if inputExclusionList, ok := input["exclusion_list"].([]interface{}); ok {
		exclusionList = make([]string, len(inputExclusionList))
		for i, val := range inputExclusionList {
			exclusionList[i] = fmt.Sprintf("%v", val)
		}
	}

	whitelist := r.config.Whitelist
	if inputWhitelist, ok := input["whitelist"].([]interface{}); ok {
		whitelist = make([]string, len(inputWhitelist))
		for i, val := range inputWhitelist {
			whitelist[i] = fmt.Sprintf("%v", val)
		}
	}

	enableCaching := r.config.EnableCaching
	if inputEnableCaching, ok := input["enable_caching"].(bool); ok {
		enableCaching = inputEnableCaching
	}

	cacheTTL := r.config.CacheTTL
	if inputCacheTTL, ok := input["cache_ttl"].(float64); ok {
		cacheTTL = int(inputCacheTTL)
	}

	enableProfiling := r.config.EnableProfiling
	if inputEnableProfiling, ok := input["enable_profiling"].(bool); ok {
		enableProfiling = inputEnableProfiling
	}

	timeout := r.config.Timeout
	if inputTimeout, ok := input["timeout"].(float64); ok {
		timeout = int(inputTimeout)
	}

	returnRawResults := r.config.ReturnRawResults
	if inputReturnRaw, ok := input["return_raw_results"].(bool); ok {
		returnRawResults = inputReturnRaw
	}

	customParams := r.config.CustomParams
	if inputCustomParams, ok := input["custom_params"].(map[string]interface{}); ok {
		customParams = inputCustomParams
	}

	// Validasi input yang diperlukan
	if userID == "" {
		// Coba ambil userID dari input
		if id, exists := input["user_id"].(string); exists && id != "" {
			userID = id
		} else if id, exists := input["userId"].(string); exists && id != "" {
			userID = id
		} else if id, exists := input["id"].(string); exists && id != "" {
			userID = id
		}
	}

	if userID == "" {
		return map[string]interface{}{
			"success":   false,
			"error":     "user_id diperlukan untuk rekomendasi personalisasi",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	if apiKey == "" {
		return map[string]interface{}{
			"success":   false,
			"error":     "api_key diperlukan untuk engine rekomendasi",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Buat konteks rekomendasi dengan timeout
	recCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Hasilkan rekomendasi
	recommendationResult, err := r.generateRecommendations(recCtx, userID, input)
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
		"provider":             provider,
		"algorithm":            algorithm,
		"model_name":           modelName,
		"user_id":              userID,
		"item_type":            itemType,
		"context_features":     contextFeatures,
		"num_recommendations":  numRecommendations,
		"score_threshold":      scoreThreshold,
		"diversity_factor":     diversityFactor,
		"freshness_factor":     freshnessFactor,
		"exclusion_list":       exclusionList,
		"whitelist":            whitelist,
		"recommendations":      recommendationResult.Items,
		"recommendation_count": len(recommendationResult.Items),
		"enable_caching":       enableCaching,
		"enable_profiling":     enableProfiling,
		"return_raw_results":   returnRawResults,
		"timestamp":            time.Now().Unix(),
		"input_data":           input,
		"config":               r.config,
	}

	// Tambahkan metrik kinerja jika profiling diaktifkan
	if enableProfiling {
		finalResult["performance_metrics"] = map[string]interface{}{
			"start_time": time.Now().Unix(),
			"end_time":   time.Now().Unix(),
			"duration":   recommendationResult.ProcessingTime,
		}
	}

	return finalResult, nil
}

// generateRecommendations menghasilkan rekomendasi untuk pengguna
func (r *AdvancedRecommendationEngineNode) generateRecommendations(ctx context.Context, userID string, input map[string]interface{}) (*RecommendationResult, error) {
	// Simulasikan waktu pemrosesan
	startTime := time.Now()
	time.Sleep(150 * time.Millisecond) // Simulasikan waktu pemrosesan

	// Simulasikan item-item yang direkomendasikan
	var recommendedItems []RecommendationItem
	
	// Buat beberapa item rekomendasi berdasarkan tipe item
	for i := 0; i < r.config.NumRecommendations; i++ {
		score := 0.9 - (float64(i) * 0.08) // Skor menurun sedikit untuk tiap item
		if score < r.config.ScoreThreshold {
			score = r.config.ScoreThreshold
		}

		item := RecommendationItem{
			ItemID:   fmt.Sprintf("item_%d_%d", i+1, time.Now().UnixNano()%1000),
			Name:     fmt.Sprintf("Rekomendasi Item %d", i+1),
			Category: "general",
			Score:    score,
			Reason:   fmt.Sprintf("Berdasarkan kesamaan dengan preferensi pengguna %s", userID),
			Contextual: i%3 == 0, // 1 dari 3 item adalah kontekstual
		}

		// Tambahkan metadata berdasarkan tipe item
		switch r.config.ItemType {
		case "product":
			item.Category = "produk"
			item.Metadata = map[string]interface{}{
				"harga":      fmt.Sprintf("%d000", 100+i*10),
				"kategori":   []string{"elektronik", "fashion", "rumah tangga"}[i%3],
				"penilaian":  float64(4 + i%2),
				"tersedia":   true,
			}
		case "content":
			item.Category = "konten"
			item.Metadata = map[string]interface{}{
				"jenis":      []string{"artikel", "video", "podcast"}[i%3],
				"kategori":   []string{"teknologi", "olahraga", "hiburan"}[i%3],
				"durasi":     fmt.Sprintf("%d menit", 5+i*2),
				"penulis":    fmt.Sprintf("Penulis %d", i+1),
			}
		case "service":
			item.Category = "layanan"
			item.Metadata = map[string]interface{}{
				"jenis":      []string{"layanan", "langganan", "konsultasi"}[i%3],
				"kategori":   []string{"bisnis", "kesehatan", "pendidikan"}[i%3],
				"harga":      fmt.Sprintf("%d000", 50+i*5),
				"rating":     float64(4 + i%2),
			}
		default:
			item.Metadata = map[string]interface{}{
				"kategori": []string{"kategori_1", "kategori_2", "kategori_3"}[i%3],
				"atribut":  fmt.Sprintf("atribut_%d", i+1),
			}
		}

		// Filter berdasarkan exclusion dan whitelist jika ada
		shouldInclude := true
		
		// Cek exclusion list
		for _, excluded := range r.config.ExclusionList {
			if item.ItemID == excluded || item.Name == excluded {
				shouldInclude = false
				break
			}
		}
		
		// Cek whitelist jika ada
		if shouldInclude && len(r.config.Whitelist) > 0 {
			shouldInclude = false
			for _, allowed := range r.config.Whitelist {
				if item.ItemID == allowed || item.Name == allowed {
					shouldInclude = true
					break
				}
			}
		}
		
		if shouldInclude && score >= r.config.ScoreThreshold {
			recommendedItems = append(recommendedItems, item)
		}
	}

	// Filter jumlah item sesuai konfigurasi jika perlu
	if len(recommendedItems) > r.config.NumRecommendations {
		recommendedItems = recommendedItems[:r.config.NumRecommendations]
	}

	result := &RecommendationResult{
		Items:         recommendedItems,
		UserID:        userID,
		Algorithm:     r.config.Algorithm,
		Timestamp:     time.Now().Unix(),
		ProcessingTime: time.Since(startTime).Seconds(),
	}

	return result, nil
}

// GetType mengembalikan jenis node
func (r *AdvancedRecommendationEngineNode) GetType() string {
	return "advanced_recommendation_engine"
}

// GetID mengembalikan ID unik untuk instance node
func (r *AdvancedRecommendationEngineNode) GetID() string {
	return "adv_rec_" + fmt.Sprintf("%d", time.Now().Unix())
}

// RegisterAdvancedRecommendationEngineNode mendaftarkan node Advanced Recommendation Engine dengan engine
func RegisterAdvancedRecommendationEngineNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("advanced_recommendation_engine", func(config map[string]interface{}) (engine.NodeInstance, error) {
		return NewAdvancedRecommendationEngineNode(config)
	})
}