package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"citadel-agent/backend/internal/workflow/core/engine"
)

// AdvancedPredictiveAnalyticsConfig mewakili konfigurasi untuk node Advanced Predictive Analytics
type AdvancedPredictiveAnalyticsConfig struct {
	Provider          string                 `json:"provider"`           // Penyedia analitik prediktif
	AnalysisType      string                 `json:"analysis_type"`      // Jenis analitik (forecasting, classification, regression)
	ModelName         string                 `json:"model_name"`         // Nama model analitik
	APIKey            string                 `json:"api_key"`            // Kunci API untuk layanan
	TargetVariable    string                 `json:"target_variable"`    // Variabel target untuk prediksi
	Features          []string               `json:"features"`           // Fitur-fitur untuk analitik
	TimeHorizon       int                    `json:"time_horizon"`       // Horizon waktu untuk prediksi (dalam hari)
	ConfidenceLevel   float64                `json:"confidence_level"`   // Tingkat kepercayaan (0.0-1.0)
	EnableEnsemble    bool                   `json:"enable_ensemble"`    // Apakah mengaktifkan ensemble modeling
	EnsembleModels    []string               `json:"ensemble_models"`    // Model-model dalam ensemble
	EnableAnomalyDetection bool              `json:"enable_anomaly_detection"` // Apakah mengaktifkan deteksi anomali
	AnomalyThreshold  float64                `json:"anomaly_threshold"`  // Ambang batas deteksi anomali
	SeasonalPatterns  []string               `json:"seasonal_patterns"`  // Pola musiman untuk dipertimbangkan
	ExternalFactors   []string               `json:"external_factors"`   // Faktor eksternal yang mempengaruhi prediksi
	EnableExplainability bool                `json:"enable_explainability"` // Apakah mengaktifkan penjelasan
	EnableCaching     bool                   `json:"enable_caching"`     // Apakah mengaktifkan caching hasil
	CacheTTL          int                    `json:"cache_ttl"`          // Waktu cache dalam detik
	EnableProfiling   bool                   `json:"enable_profiling"`   // Apakah mengaktifkan profiling
	Timeout           int                    `json:"timeout"`            // Waktu timeout dalam detik
	ReturnRawResults  bool                   `json:"return_raw_results"` // Apakah mengembalikan hasil mentah
	CustomParams      map[string]interface{} `json:"custom_params"`      // Parameter khusus untuk analitik
	Preprocessing     PreprocessingConfig    `json:"preprocessing"`      // Konfigurasi pra-pemrosesan
	Postprocessing    PostprocessingConfig   `json:"postprocessing"`     // Konfigurasi pasca-pemrosesan
	FeatureWeights    map[string]float64     `json:"feature_weights"`    // Bobot untuk fitur-fitur berbeda
	ValidationParams  ValidationParameters   `json:"validation_params"`  // Parameter validasi
}

// ValidationParameters mewakili parameter untuk validasi model
type ValidationParameters struct {
	TestSize         float64 `json:"test_size"`          // Ukuran data uji (0.0-1.0)
	CrossValidation  bool    `json:"cross_validation"`   // Apakah menggunakan validasi silang
	CVFolds          int     `json:"cv_folds"`          // Jumlah fold validasi silang
	Metrics          []string `json:"metrics"`           // Metrik evaluasi
	EarlyStopping    bool    `json:"early_stopping"`    // Apakah menggunakan early stopping
	ValidationWindow int     `json:"validation_window"` // Jendela validasi untuk data waktu
}

// PredictionResult mewakili hasil prediksi
type PredictionResult struct {
	Predictions      []PredictionPoint `json:"predictions"`
	Confidence       float64          `json:"confidence"`
	ModelPerformance map[string]float64 `json:"model_performance"`
	FeatureImportance map[string]float64 `json:"feature_importance"`
	Anomalies        []AnomalyPoint   `json:"anomalies,omitempty"`
	Timestamp        int64            `json:"timestamp"`
	ProcessingTime   float64          `json:"processing_time"`
}

// PredictionPoint mewakili titik prediksi individual
type PredictionPoint struct {
	Date        int64   `json:"date"`         // Timestamp untuk prediksi
	Value       float64 `json:"value"`        // Nilai prediksi
	LowerBound  float64 `json:"lower_bound"`  // Batas bawah interval kepercayaan
	UpperBound  float64 `json:"upper_bound"`  // Batas atas interval kepercayaan
	Probability float64 `json:"probability"`  // Probabilitas (untuk klasifikasi)
	Anomaly     bool    `json:"anomaly"`      // Apakah ini anomali
}

// AnomalyPoint mewakili titik anomali
type AnomalyPoint struct {
	Date    int64   `json:"date"`    // Timestamp anomali
	Value   float64 `json:"value"`   // Nilai yang dianggap anomali
	Score   float64 `json:"score"`   // Skor anomali
	Reason  string  `json:"reason"`  // Alasan mengapa ini anomali
}

// AdvancedPredictiveAnalyticsNode mewakili node yang melakukan analitik prediktif canggih
type AdvancedPredictiveAnalyticsNode struct {
	config *AdvancedPredictiveAnalyticsConfig
}

// NewAdvancedPredictiveAnalyticsNode membuat node Advanced Predictive Analytics baru
func NewAdvancedPredictiveAnalyticsNode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Konversi map interface{} ke JSON lalu ke struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("gagal mengubah konfig menjadi json: %v", err)
	}

	var analyticsConfig AdvancedPredictiveAnalyticsConfig
	err = json.Unmarshal(jsonData, &analyticsConfig)
	if err != nil {
		return nil, fmt.Errorf("gagal menguraikan konfig: %v", err)
	}

	// Validasi dan atur default
	if analyticsConfig.AnalysisType == "" {
		analyticsConfig.AnalysisType = "forecasting"
	}

	if analyticsConfig.TimeHorizon == 0 {
		analyticsConfig.TimeHorizon = 30 // default 30 hari
	}

	if analyticsConfig.ConfidenceLevel == 0 {
		analyticsConfig.ConfidenceLevel = 0.95
	}

	if analyticsConfig.AnomalyThreshold == 0 {
		analyticsConfig.AnomalyThreshold = 2.0 // 2 standar deviasi
	}

	if analyticsConfig.Timeout == 0 {
		analyticsConfig.Timeout = 120 // default timeout 120 detik
	}

	if len(analyticsConfig.EnsembleModels) == 0 {
		analyticsConfig.EnsembleModels = []string{"linear_regression", "random_forest", "gradient_boosting"}
	}

	return &AdvancedPredictiveAnalyticsNode{
		config: &analyticsConfig,
	}, nil
}

// Execute mengimplementasikan interface NodeInstance
func (p *AdvancedPredictiveAnalyticsNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Timpa konfigurasi dengan nilai input jika disediakan
	provider := p.config.Provider
	if inputProvider, ok := input["provider"].(string); ok && inputProvider != "" {
		provider = inputProvider
	}

	analysisType := p.config.AnalysisType
	if inputAnalysisType, ok := input["analysis_type"].(string); ok && inputAnalysisType != "" {
		analysisType = inputAnalysisType
	}

	modelName := p.config.ModelName
	if inputModelName, ok := input["model_name"].(string); ok && inputModelName != "" {
		modelName = inputModelName
	}

	apiKey := p.config.APIKey
	if inputAPIKey, ok := input["api_key"].(string); ok && inputAPIKey != "" {
		apiKey = inputAPIKey
	}

	targetVariable := p.config.TargetVariable
	if inputTargetVar, ok := input["target_variable"].(string); ok && inputTargetVar != "" {
		targetVariable = inputTargetVar
	}

	features := p.config.Features
	if inputFeatures, ok := input["features"].([]interface{}); ok {
		features = make([]string, len(inputFeatures))
		for i, val := range inputFeatures {
			features[i] = fmt.Sprintf("%v", val)
		}
	}

	timeHorizon := p.config.TimeHorizon
	if inputTimeHorizon, ok := input["time_horizon"].(float64); ok {
		timeHorizon = int(inputTimeHorizon)
	}

	confidenceLevel := p.config.ConfidenceLevel
	if inputConfidence, ok := input["confidence_level"].(float64); ok {
		confidenceLevel = inputConfidence
	}

	enableEnsemble := p.config.EnableEnsemble
	if inputEnableEnsemble, ok := input["enable_ensemble"].(bool); ok {
		enableEnsemble = inputEnableEnsemble
	}

	ensembleModels := p.config.EnsembleModels
	if inputEnsembleModels, ok := input["ensemble_models"].([]interface{}); ok {
		ensembleModels = make([]string, len(inputEnsembleModels))
		for i, val := range inputEnsembleModels {
			ensembleModels[i] = fmt.Sprintf("%v", val)
		}
	}

	enableAnomalyDetection := p.config.EnableAnomalyDetection
	if inputEnableAnomaly, ok := input["enable_anomaly_detection"].(bool); ok {
		enableAnomalyDetection = inputEnableAnomaly
	}

	anomalyThreshold := p.config.AnomalyThreshold
	if inputAnomalyThreshold, ok := input["anomaly_threshold"].(float64); ok {
		anomalyThreshold = inputAnomalyThreshold
	}

	seasonalPatterns := p.config.SeasonalPatterns
	if inputSeasonalPatterns, ok := input["seasonal_patterns"].([]interface{}); ok {
		seasonalPatterns = make([]string, len(inputSeasonalPatterns))
		for i, val := range inputSeasonalPatterns {
			seasonalPatterns[i] = fmt.Sprintf("%v", val)
		}
	}

	externalFactors := p.config.ExternalFactors
	if inputExternalFactors, ok := input["external_factors"].([]interface{}); ok {
		externalFactors = make([]string, len(inputExternalFactors))
		for i, val := range inputExternalFactors {
			externalFactors[i] = fmt.Sprintf("%v", val)
		}
	}

	enableExplainability := p.config.EnableExplainability
	if inputEnableExplain, ok := input["enable_explainability"].(bool); ok {
		enableExplainability = inputEnableExplain
	}

	enableCaching := p.config.EnableCaching
	if inputEnableCaching, ok := input["enable_caching"].(bool); ok {
		enableCaching = inputEnableCaching
	}

	cacheTTL := p.config.CacheTTL
	if inputCacheTTL, ok := input["cache_ttl"].(float64); ok {
		cacheTTL = int(inputCacheTTL)
	}

	enableProfiling := p.config.EnableProfiling
	if inputEnableProfiling, ok := input["enable_profiling"].(bool); ok {
		enableProfiling = inputEnableProfiling
	}

	timeout := p.config.Timeout
	if inputTimeout, ok := input["timeout"].(float64); ok {
		timeout = int(inputTimeout)
	}

	returnRawResults := p.config.ReturnRawResults
	if inputReturnRaw, ok := input["return_raw_results"].(bool); ok {
		returnRawResults = inputReturnRaw
	}

	customParams := p.config.CustomParams
	if inputCustomParams, ok := input["custom_params"].(map[string]interface{}); ok {
		customParams = inputCustomParams
	}

	// Validasi input
	if len(input) == 0 {
		return map[string]interface{}{
			"success":   false,
			"error":     "input diperlukan untuk analitik prediktif",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	if apiKey == "" {
		return map[string]interface{}{
			"success":   false,
			"error":     "api_key diperlukan untuk analitik prediktif",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Buat konteks analitik dengan timeout
	analyticsCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Lakukan analitik prediktif
	analyticsResult, err := p.performPredictiveAnalytics(analyticsCtx, input)
	if err != nil {
		return map[string]interface{}{
			"success":   false,
			"error":     err.Error(),
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Siapkan hasil akhir
	finalResult := map[string]interface{}{
		"success":                  true,
		"provider":                 provider,
		"analysis_type":            analysisType,
		"model_name":               modelName,
		"target_variable":          targetVariable,
		"features":                 features,
		"time_horizon":             timeHorizon,
		"confidence_level":         confidenceLevel,
		"enable_ensemble":          enableEnsemble,
		"ensemble_models":          ensembleModels,
		"enable_anomaly_detection": enableAnomalyDetection,
		"anomaly_threshold":        anomalyThreshold,
		"seasonal_patterns":        seasonalPatterns,
		"external_factors":         externalFactors,
		"analytics_result":         analyticsResult,
		"enable_explainability":    enableExplainability,
		"enable_caching":           enableCaching,
		"enable_profiling":         enableProfiling,
		"return_raw_results":       returnRawResults,
		"timestamp":                time.Now().Unix(),
		"input_data":               input,
		"config":                   p.config,
	}

	// Tambahkan penjelasan jika diaktifkan
	if enableExplainability {
		finalResult["explanation"] = p.generateExplanation(analyticsResult, input)
		finalResult["feature_importance"] = analyticsResult.FeatureImportance
	}

	// Tambahkan metrik kinerja jika profiling diaktifkan
	if enableProfiling {
		finalResult["performance_metrics"] = map[string]interface{}{
			"start_time": time.Now().Unix(),
			"end_time":   time.Now().Unix(),
			"duration":   analyticsResult.ProcessingTime,
		}
	}

	return finalResult, nil
}

// performPredictiveAnalytics melakukan analitik prediktif
func (p *AdvancedPredictiveAnalyticsNode) performPredictiveAnalytics(ctx context.Context, input map[string]interface{}) (*PredictionResult, error) {
	// Simulasikan waktu pemrosesan
	startTime := time.Now()
	time.Sleep(150 * time.Millisecond)

	// Simulasikan hasil prediksi
	var predictions []PredictionPoint
	var anomalies []AnomalyPoint

	// Buat prediksi berdasarkan horizon waktu
	currentTime := time.Now().Unix()
	for i := 1; i <= p.config.TimeHorizon; i++ {
		// Simulasikan nilai prediksi berdasarkan tren
		baseValue := 100.0
		if amount, exists := input["base_value"]; exists {
			if amountFloat, ok := amount.(float64); ok {
				baseValue = amountFloat
			}
		}

		// Tambahkan tren dan noise
		trend := float64(i) * 0.5
		noise := (float64(i%7) - 3.0) * 2.0 // Fluktuasi mingguan
		predictedValue := baseValue + trend + noise

		// Hitung interval kepercayaan (2.5% dan 97.5%)
		margin := 10.0
		lowerBound := predictedValue - margin
		upperBound := predictedValue + margin

		prediction := PredictionPoint{
			Date:        currentTime + int64(i*86400), // Tambah 1 hari dalam detik
			Value:       predictedValue,
			LowerBound:  lowerBound,
			UpperBound:  upperBound,
			Probability: 0.85, // Untuk kasus klasifikasi
			Anomaly:     false,
		}

		// Simulasikan deteksi anomali
		if p.config.EnableAnomalyDetection {
			// Misalkan setiap 7 hari adalah anomali
			if i%7 == 0 {
				prediction.Anomaly = true
				anomalies = append(anomalies, AnomalyPoint{
					Date:   prediction.Date,
					Value:  prediction.Value,
					Score:  p.config.AnomalyThreshold + float64(i%5),
					Reason: fmt.Sprintf("Nilai pada hari ke-%d melampaui ambang batas anomali", i),
				})
			}
		}

		predictions = append(predictions, prediction)
	}

	// Simulasikan kinerja model
	modelPerformance := map[string]float64{
		"r_squared":    0.87,
		"mae":          5.2,
		"mape":         0.08,
		"rmse":         7.1,
		"accuracy":     0.92,
	}

	// Simulasikan pentingnya fitur
	featureImportance := make(map[string]float64)
	for _, feature := range p.config.Features {
		// Berikan pentingnya acak antara 0.1 dan 1.0
		featureImportance[feature] = 0.1 + (float64(len(feature)) * 0.1)
	}
	// Pastikan jumlah pentingnya fitur <= 1.0
	total := 0.0
	for _, imp := range featureImportance {
		total += imp
	}
	if total > 1.0 {
		for k, v := range featureImportance {
			featureImportance[k] = v / total
		}
	}

	result := &PredictionResult{
		Predictions:      predictions,
		Confidence:       p.config.ConfidenceLevel,
		ModelPerformance: modelPerformance,
		FeatureImportance: featureImportance,
		Timestamp:        time.Now().Unix(),
		ProcessingTime:   time.Since(startTime).Seconds(),
	}

	if p.config.EnableAnomalyDetection {
		result.Anomalies = anomalies
	}

	return result, nil
}

// generateExplanation menghasilkan penjelasan untuk hasil analitik
func (p *AdvancedPredictiveAnalyticsNode) generateExplanation(result *PredictionResult, input map[string]interface{}) string {
	explanation := fmt.Sprintf(
		"Analisis prediktif dilakukan untuk variabel '%s' selama %d hari ke depan. ",
		p.config.TargetVariable, p.config.TimeHorizon)

	if len(result.Anomalies) > 0 {
		explanation += fmt.Sprintf(
			"Teridentifikasi %d anomali potensial dalam prediksi. ",
			len(result.Anomalies))
	}

	explanation += fmt.Sprintf(
		"Model menunjukkan tingkat kepercayaan %.2f%% dengan kinerja R² %.2f dan akurasi %.2f%%. ",
		result.Confidence*100, result.ModelPerformance["r_squared"], result.ModelPerformance["accuracy"]*100)

	if p.config.EnableEnsemble {
		explanation += fmt.Sprintf(
			"Analisis menggunakan ensemble dari %d model: %v. ",
			len(p.config.EnsembleModels), p.config.EnsembleModels)
	}

	return explanation
}

// GetType mengembalikan jenis node
func (p *AdvancedPredictiveAnalyticsNode) GetType() string {
	return "advanced_predictive_analytics"
}

// GetID mengembalikan ID unik untuk instance node
func (p *AdvancedPredictiveAnalyticsNode) GetID() string {
	return "adv_pred_analytics_" + fmt.Sprintf("%d", time.Now().Unix())
}

// RegisterAdvancedPredictiveAnalyticsNode mendaftarkan node Advanced Predictive Analytics dengan engine
func RegisterAdvancedPredictiveAnalyticsNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("advanced_predictive_analytics", func(config map[string]interface{}) (engine.NodeInstance, error) {
		return NewAdvancedPredictiveAnalyticsNode(config)
	})
}