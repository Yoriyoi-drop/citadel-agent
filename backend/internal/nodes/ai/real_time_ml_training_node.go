package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
)

// RealTimeMLTrainingConfig mewakili konfigurasi untuk node Real-Time ML Training
type RealTimeMLTrainingConfig struct {
	ModelType       string                 `json:"model_type"`        // Jenis model (classification, regression, dll.)
	ModelName       string                 `json:"model_name"`        // Nama model untuk dilatih
	StreamingData   string                 `json:"streaming_data"`    // Sumber data streaming
	Features        []string               `json:"features"`          // Daftar fitur untuk pelatihan
	TargetVariable  string                 `json:"target_variable"`   // Variabel target untuk pembelajaran terawasi
	Hyperparameters map[string]interface{} `json:"hyperparameters"`   // Hyperparameter model
	TrainingParams  RealTimeTrainingParams `json:"training_params"`   // Parameter pelatihan real-time
	WindowSize      int                    `json:"window_size"`       // Ukuran jendela untuk pelatihan real-time
	UpdateInterval  int                    `json:"update_interval"`   // Interval pembaruan model dalam detik
	MaxTrainingTime int                    `json:"max_training_time"` // Waktu pelatihan maksimum dalam detik
	EnableAdaptation bool                  `json:"enable_adaptation"` // Apakah mengaktifkan adaptasi real-time
	PerformanceThreshold float64           `json:"performance_threshold"` // Ambang batas kinerja
	CheckpointPath  string                 `json:"checkpoint_path"`   // Jalur untuk menyimpan checkpoint model
	ModelSavePath   string                 `json:"model_save_path"`   // Jalur untuk menyimpan model akhir
	EnableLogging   bool                   `json:"enable_logging"`    // Apakah mengaktifkan logging pelatihan
	DebugMode       bool                   `json:"debug_mode"`        // Apakah berjalan dalam mode debug
	CustomParams    map[string]interface{} `json:"custom_params"`     // Parameter khusus untuk proses pelatihan
}

// RealTimeTrainingParams mewakili parameter spesifik untuk proses pelatihan real-time
type RealTimeTrainingParams struct {
	BatchSize         int     `json:"batch_size"`         // Ukuran batch pelatihan
	LearningRate      float64 `json:"learning_rate"`      // Tingkat pembelajaran untuk pelatihan
	MaxSamples        int     `json:"max_samples"`        // Jumlah maksimum sampel untuk disimpan
	ForgettingFactor  float64 `json:"forgetting_factor"`  // Faktor pelupaan untuk adaptasi konsep
	DriftDetection    bool    `json:"drift_detection"`    // Apakah mendeteksi pergeseran data
	DriftThreshold    float64 `json:"drift_threshold"`    // Ambang batas deteksi pergeseran
	EarlyStopping     bool    `json:"early_stopping"`     // Apakah menggunakan pemberhentian awal
	MinDelta          float64 `json:"min_delta"`          // Perubahan minimum untuk pemberhentian awal
	Patience          int     `json:"patience"`           // Kesabaran untuk pemberhentian awal
	LossFunction      string  `json:"loss_function"`      // Fungsi kerugian untuk digunakan
	Optimizer         string  `json:"optimizer"`          // Algoritma optimasi untuk digunakan
	AdaptationMethod  string  `json:"adaptation_method"`  // Metode adaptasi model (incremental, online, dll.)
}

// RealTimeMLTrainingNode mewakili node yang melatih model ML secara real-time
type RealTimeMLTrainingNode struct {
	config *RealTimeMLTrainingConfig
}

// NewRealTimeMLTrainingNode membuat node Real-Time ML Training baru
func NewRealTimeMLTrainingNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	// Konversi map interface{} ke JSON lalu ke struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("gagal mengubah konfig menjadi json: %v", err)
	}

	var trainingConfig RealTimeMLTrainingConfig
	err = json.Unmarshal(jsonData, &trainingConfig)
	if err != nil {
		return nil, fmt.Errorf("gagal menguraikan konfig: %v", err)
	}

	// Validasi dan atur default
	if trainingConfig.ModelType == "" {
		trainingConfig.ModelType = "classification"
	}

	if trainingConfig.TrainingParams.BatchSize == 0 {
		trainingConfig.TrainingParams.BatchSize = 1
	}

	if trainingConfig.TrainingParams.MaxSamples == 0 {
		trainingConfig.TrainingParams.MaxSamples = 1000
	}

	if trainingConfig.TrainingParams.LearningRate == 0 {
		trainingConfig.TrainingParams.LearningRate = 0.001
	}

	if trainingConfig.UpdateInterval == 0 {
		trainingConfig.UpdateInterval = 10 // default ke 10 detik
	}

	if trainingConfig.MaxTrainingTime == 0 {
		trainingConfig.MaxTrainingTime = 3600 // default ke 1 jam
	}

	if trainingConfig.WindowSize == 0 {
		trainingConfig.WindowSize = 100 // window size default
	}

	return &RealTimeMLTrainingNode{
		config: &trainingConfig,
	}, nil
}

// Execute mengimplementasikan interface NodeInstance
func (r *RealTimeMLTrainingNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Timpa konfigurasi dengan nilai input jika disediakan
	modelType := r.config.ModelType
	if inputModelType, ok := input["model_type"].(string); ok && inputModelType != "" {
		modelType = inputModelType
	}

	modelName := r.config.ModelName
	if inputModelName, ok := input["model_name"].(string); ok && inputModelName != "" {
		modelName = inputModelName
	}

	streamingData := r.config.StreamingData
	if inputStreamingData, ok := input["streaming_data"].(string); ok && inputStreamingData != "" {
		streamingData = inputStreamingData
	}

	features := r.config.Features
	if inputFeatures, ok := input["features"].([]interface{}); ok {
		features = make([]string, len(inputFeatures))
		for i, val := range inputFeatures {
			features[i] = fmt.Sprintf("%v", val)
		}
	}

	targetVariable := r.config.TargetVariable
	if inputTargetVariable, ok := input["target_variable"].(string); ok && inputTargetVariable != "" {
		targetVariable = inputTargetVariable
	}

	hyperparameters := r.config.Hyperparameters
	if inputHyperparams, ok := input["hyperparameters"].(map[string]interface{}); ok {
		hyperparameters = inputHyperparams
	}

	windowSize := r.config.WindowSize
	if inputWindowSize, ok := input["window_size"].(float64); ok {
		windowSize = int(inputWindowSize)
	}

	updateInterval := r.config.UpdateInterval
	if inputUpdateInterval, ok := input["update_interval"].(float64); ok {
		updateInterval = int(inputUpdateInterval)
	}

	maxTrainingTime := r.config.MaxTrainingTime
	if inputMaxTrainingTime, ok := input["max_training_time"].(float64); ok {
		maxTrainingTime = int(inputMaxTrainingTime)
	}

	enableAdaptation := r.config.EnableAdaptation
	if inputEnableAdaptation, ok := input["enable_adaptation"].(bool); ok {
		enableAdaptation = inputEnableAdaptation
	}

	performanceThreshold := r.config.PerformanceThreshold
	if inputPerfThreshold, ok := input["performance_threshold"].(float64); ok {
		performanceThreshold = inputPerfThreshold
	}

	checkpointPath := r.config.CheckpointPath
	if inputCheckpointPath, ok := input["checkpoint_path"].(string); ok && inputCheckpointPath != "" {
		checkpointPath = inputCheckpointPath
	}

	modelSavePath := r.config.ModelSavePath
	if inputModelSavePath, ok := input["model_save_path"].(string); ok && inputModelSavePath != "" {
		modelSavePath = inputModelSavePath
	}

	enableLogging := r.config.EnableLogging
	if inputEnableLogging, ok := input["enable_logging"].(bool); ok {
		enableLogging = inputEnableLogging
	}

	debugMode := r.config.DebugMode
	if inputDebugMode, ok := input["debug_mode"].(bool); ok {
		debugMode = inputDebugMode
	}

	customParams := r.config.CustomParams
	if inputCustomParams, ok := input["custom_params"].(map[string]interface{}); ok {
		customParams = inputCustomParams
	}

	// Validasi input yang diperlukan
	if streamingData == "" {
		return map[string]interface{}{
			"success":   false,
			"error":     "streaming_data diperlukan untuk pelatihan ML real-time",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Buat konteks pelatihan dengan timeout
	trainingCtx, cancel := context.WithTimeout(ctx, time.Duration(maxTrainingTime)*time.Second)
	defer cancel()

	// Simulasikan proses pelatihan real-time
	trainingResult, err := r.simulateRealTimeTraining(trainingCtx, input)
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
		"model_type":           modelType,
		"model_name":           modelName,
		"streaming_data":       streamingData,
		"features_used":        features,
		"target_variable":      targetVariable,
		"hyperparameters":      hyperparameters,
		"window_size":          windowSize,
		"update_interval":      updateInterval,
		"enable_adaptation":    enableAdaptation,
		"performance_threshold": performanceThreshold,
		"training_completed":   true,
		"training_result":      trainingResult,
		"training_parameters":  r.config.TrainingParams,
		"checkpoint_path":      checkpointPath,
		"model_save_path":      modelSavePath,
		"enable_logging":       enableLogging,
		"debug_mode":           debugMode,
		"timestamp":            time.Now().Unix(),
		"input_data":           input,
	}

	// Jika mode debug diaktifkan, tambahkan informasi lebih detail
	if debugMode {
		finalResult["debug_info"] = map[string]interface{}{
			"config": r.config,
			"input":  input,
		}
	}

	return finalResult, nil
}

// simulateRealTimeTraining menyimulasikan proses pelatihan real-time
func (r *RealTimeMLTrainingNode) simulateRealTimeTraining(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Simulasikan pelatihan dengan beberapa iterasi
	trainingResults := make([]map[string]interface{}, 0)
	
	// Proses data dalam batch real-time
	for batch := 1; batch <= 5; batch++ { // Simulasikan 5 batch
		// Simulasikan waktu pemrosesan batch
		time.Sleep(200 * time.Millisecond)
		
		// Simulasikan metrik untuk batch ini
		batchMetrics := map[string]interface{}{
			"batch":        batch,
			"loss":         1.0 / float64(batch), // Simulasikan penurunan loss
			"accuracy":     0.4 + (float64(batch) * 0.1), // Simulasikan peningkatan akurasi
			"learning_rate": r.config.TrainingParams.LearningRate,
			"timestamp":    time.Now().Unix(),
			"data_samples": 50, // Jumlah sampel per batch
		}
		
		trainingResults = append(trainingResults, batchMetrics)
		
		// Periksa apakah konteks dibatalkan
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// Lanjutkan pelatihan
		}
	}

	// Simulasikan metrik akhir model
	finalMetrics := map[string]interface{}{
		"final_loss":        0.2,
		"final_accuracy":    0.85,
		"total_batches":     5,
		"training_time":     time.Since(time.Now().Add(-time.Duration(5*200) * time.Millisecond)).Seconds(),
		"model_size":        "8.5 MB", // Ukuran model simulasi
		"features_count":    len(r.config.Features),
		"total_samples":     250, // Jumlah total sampel
		"model_architecture": fmt.Sprintf("%s_model_realtime", r.config.ModelType),
		"drift_detected":    false, // Untuk simulasi
	}

	// Simulasikan deteksi pergeseran data
	if r.config.TrainingParams.DriftDetection {
		finalMetrics["drift_detected"] = true
		finalMetrics["drift_magnitude"] = 0.12
		finalMetrics["drift_timestamp"] = time.Now().Unix()
	}

	result := map[string]interface{}{
		"training_history": trainingResults,
		"final_metrics":    finalMetrics,
		"model_path":       r.config.ModelSavePath,
		"checkpoint_path":  r.config.CheckpointPath,
		"training_status":  "completed",
		"real_time_adaptation": r.config.EnableAdaptation,
		"model_card": map[string]interface{}{
			"name":        r.config.ModelName,
			"type":        r.config.ModelType,
			"description": fmt.Sprintf("Model %s real-time terlatih untuk tugas %s", r.config.ModelType, r.config.TargetVariable),
			"version":     "1.0.0",
			"created_at":  time.Now().Unix(),
			"framework":   "simulated_realtime_ml_framework",
			"license":     "MIT",
			"adaptation_method": r.config.TrainingParams.AdaptationMethod,
		},
		"performance_history": []map[string]interface{}{
			{
				"timestamp": time.Now().Unix(),
				"accuracy":  0.72,
				"loss":      0.35,
			},
			{
				"timestamp": time.Now().Unix() + 10,
				"accuracy":  0.78,
				"loss":      0.28,
			},
			{
				"timestamp": time.Now().Unix() + 20,
				"accuracy":  0.81,
				"loss":      0.24,
			},
			{
				"timestamp": time.Now().Unix() + 30,
				"accuracy":  0.84,
				"loss":      0.21,
			},
			{
				"timestamp": time.Now().Unix() + 40,
				"accuracy":  0.85,
				"loss":      0.20,
			},
		},
	}

	return result, nil
}

// GetType mengembalikan jenis node
func (r *RealTimeMLTrainingNode) GetType() string {
	return "real_time_ml_training"
}

// GetID mengembalikan ID unik untuk instance node
func (r *RealTimeMLTrainingNode) GetID() string {
	return "realtime_ml_" + fmt.Sprintf("%d", time.Now().Unix())
}

// RegisterRealTimeMLTrainingNode mendaftarkan node Real-Time ML Training dengan engine
func RegisterRealTimeMLTrainingNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("real_time_ml_training", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return NewRealTimeMLTrainingNode(config)
	})
}