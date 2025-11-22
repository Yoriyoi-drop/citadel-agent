package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/workflow/core/engine"
)

// WorkflowControlConfig mewakili konfigurasi untuk node Workflow Control
type WorkflowControlConfig struct {
	ControlType     string                   `json:"control_type"`      // Jenis kontrol (branch, merge, sync, parallel, etc.)
	NextNodeID      string                   `json:"next_node_id"`      // ID node berikutnya
	Condition       string                   `json:"condition"`         // Kondisi untuk kontrol alur
	Timeout         int                      `json:"timeout"`           // Waktu timeout dalam detik
	MaxRetries      int                      `json:"max_retries"`       // Jumlah maksimum percobaan ulang
	EnableRetry     bool                     `json:"enable_retry"`      // Apakah mengaktifkan retry
	RetryDelay      int                      `json:"retry_delay"`       // Delay antar retry dalam detik
	EnableCaching   bool                     `json:"enable_caching"`    // Apakah mengaktifkan caching
	CacheTTL        int                      `json:"cache_ttl"`         // Waktu cache dalam detik
	EnableProfiling bool                     `json:"enable_profiling"`  // Apakah mengaktifkan profiling
	ReturnRawResults bool                    `json:"return_raw_results"` // Apakah mengembalikan hasil mentah
	CustomParams    map[string]interface{}   `json:"custom_params"`     // Parameter khusus untuk kontrol
	Preprocessing   PreprocessingConfig      `json:"preprocessing"`     // Konfigurasi pra-pemrosesan
	Postprocessing  PostprocessingConfig     `json:"postprocessing"`    // Konfigurasi pasca-pemrosesan
	ErrorHandling   ErrorHandlingConfig      `json:"error_handling"`    // Konfigurasi penanganan error
	ParallelBranches []string                 `json:"parallel_branches"` // Cabang-cabang paralel
	SyncPoints      []string                 `json:"sync_points"`      // Titik sinkronisasi
	WaitConditions  []WaitCondition          `json:"wait_conditions"`  // Kondisi tunggu
}

// ErrorHandlingConfig mewakili konfigurasi penanganan error
type ErrorHandlingConfig struct {
	ErrorType     string `json:"error_type"`      // Jenis error (all, specific, timeout, etc.)
	HandlerAction string `json:"handler_action"`  // Aksi untuk dilakukan (retry, skip, fail, custom)
	FallbackValue interface{} `json:"fallback_value"` // Nilai fallback
	NextNode      string `json:"next_node"`       // Node berikutnya jika terjadi error
}

// WaitCondition mewakili kondisi tunggu
type WaitCondition struct {
	Type     string      `json:"type"`      // Jenis kondisi tunggu (event, time, resource, etc.)
	Resource string      `json:"resource"`  // Sumber daya yang ditunggu
	Timeout  int         `json:"timeout"`   // Timeout untuk menunggu
	Value    interface{} `json:"value"`     // Nilai yang ditunggu
}

// PreprocessingConfig mewakili konfigurasi pra-pemrosesan
type PreprocessingConfig struct {
	NormalizeInput bool                   `json:"normalize_input"` // Apakah menormalkan input
	ValidateInput  bool                   `json:"validate_input"`  // Apakah memvalidasi input
	TransformInput bool                   `json:"transform_input"` // Apakah mentransformasi input
	TransformRules map[string]interface{} `json:"transform_rules"` // Aturan transformasi
}

// PostprocessingConfig mewakili konfigurasi pasca-pemrosesan
type PostprocessingConfig struct {
	FilterOutput  bool              `json:"filter_output"`   // Apakah memfilter output
	OutputMapping map[string]string `json:"output_mapping"`  // Pemetaan field output
	TransformOutput bool            `json:"transform_output"` // Apakah mentransformasi output
}

// WorkflowControlNode mewakili node yang mengontrol alur kerja
type WorkflowControlNode struct {
	config *WorkflowControlConfig
}

// NewWorkflowControlNode membuat node Workflow Control baru
func NewWorkflowControlNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	// Konversi map interface{} ke JSON lalu ke struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("gagal mengubah konfig menjadi json: %v", err)
	}

	var controlConfig WorkflowControlConfig
	err = json.Unmarshal(jsonData, &controlConfig)
	if err != nil {
		return nil, fmt.Errorf("gagal menguraikan konfig: %v", err)
	}

	// Validasi dan atur default
	if controlConfig.ControlType == "" {
		controlConfig.ControlType = "sequence"
	}

	if controlConfig.MaxRetries == 0 {
		controlConfig.MaxRetries = 3
	}

	if controlConfig.Timeout == 0 {
		controlConfig.Timeout = 60 // default timeout 60 detik
	}

	if controlConfig.RetryDelay == 0 {
		controlConfig.RetryDelay = 1 // default delay 1 detik
	}

	if controlConfig.ErrorHandling.HandlerAction == "" {
		controlConfig.ErrorHandling.HandlerAction = "retry"
	}

	return &WorkflowControlNode{
		config: &controlConfig,
	}, nil
}

// Execute mengimplementasikan interface NodeInstance
func (w *WorkflowControlNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Timpa konfigurasi dengan nilai input jika disediakan
	controlType := w.config.ControlType
	if inputControlType, ok := input["control_type"].(string); ok && inputControlType != "" {
		controlType = inputControlType
	}

	nextNodeID := w.config.NextNodeID
	if inputNextNodeID, ok := input["next_node_id"].(string); ok && inputNextNodeID != "" {
		nextNodeID = inputNextNodeID
	}

	condition := w.config.Condition
	if inputCondition, ok := input["condition"].(string); ok && inputCondition != "" {
		condition = inputCondition
	}

	timeout := w.config.Timeout
	if inputTimeout, ok := input["timeout"].(float64); ok {
		timeout = int(inputTimeout)
	}

	maxRetries := w.config.MaxRetries
	if inputMaxRetries, ok := input["max_retries"].(float64); ok {
		maxRetries = int(inputMaxRetries)
	}

	enableRetry := w.config.EnableRetry
	if inputEnableRetry, ok := input["enable_retry"].(bool); ok {
		enableRetry = inputEnableRetry
	}

	retryDelay := w.config.RetryDelay
	if inputRetryDelay, ok := input["retry_delay"].(float64); ok {
		retryDelay = int(inputRetryDelay)
	}

	enableCaching := w.config.EnableCaching
	if inputEnableCaching, ok := input["enable_caching"].(bool); ok {
		enableCaching = inputEnableCaching
	}

	cacheTTL := w.config.CacheTTL
	if inputCacheTTL, ok := input["cache_ttl"].(float64); ok {
		cacheTTL = int(inputCacheTTL)
	}

	enableProfiling := w.config.EnableProfiling
	if inputEnableProfiling, ok := input["enable_profiling"].(bool); ok {
		enableProfiling = inputEnableProfiling
	}

	returnRawResults := w.config.ReturnRawResults
	if inputReturnRaw, ok := input["return_raw_results"].(bool); ok {
		returnRawResults = inputReturnRaw
	}

	customParams := w.config.CustomParams
	if inputCustomParams, ok := input["custom_params"].(map[string]interface{}); ok {
		customParams = inputCustomParams
	}

	parallelBranches := w.config.ParallelBranches
	if inputParallelBranches, ok := input["parallel_branches"].([]interface{}); ok {
		parallelBranches = make([]string, len(inputParallelBranches))
		for i, val := range inputParallelBranches {
			if valStr, ok := val.(string); ok {
				parallelBranches[i] = valStr
			}
		}
	}

	syncPoints := w.config.SyncPoints
	if inputSyncPoints, ok := input["sync_points"].([]interface{}); ok {
		syncPoints = make([]string, len(inputSyncPoints))
		for i, val := range inputSyncPoints {
			if valStr, ok := val.(string); ok {
				syncPoints[i] = valStr
			}
		}
	}

	// Validasi input
	if len(input) == 0 {
		return map[string]interface{}{
			"success":   false,
			"error":     "input diperlukan untuk kontrol alur kerja",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Buat konteks operasi dengan timeout
	controlCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Lakukan kontrol alur kerja
	controlResult, err := w.executeControl(controlCtx, controlType, input)
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
		"control_type":         controlType,
		"next_node_id":         nextNodeID,
		"condition":            condition,
		"max_retries":          maxRetries,
		"enable_retry":         enableRetry,
		"retry_delay":          retryDelay,
		"parallel_branches":    parallelBranches,
		"sync_points":         syncPoints,
		"control_result":       controlResult,
		"enable_caching":       enableCaching,
		"enable_profiling":     enableProfiling,
		"return_raw_results":   returnRawResults,
		"timestamp":            time.Now().Unix(),
		"input_data":           input,
		"config":               w.config,
	}

	// Tambahkan informasi kontrol lanjutan
	if nextNodeID != "" {
		finalResult["next_node"] = nextNodeID
	}

	// Tambahkan metrik kinerja jika profiling diaktifkan
	if enableProfiling {
		finalResult["performance_metrics"] = map[string]interface{}{
			"start_time": time.Now().Unix(),
			"end_time":   time.Now().Unix(),
			"duration":   time.Since(time.Now().Add(-time.Duration(timeout) * time.Second)).Seconds(),
		}
	}

	return finalResult, nil
}

// executeControl mengeksekusi kontrol alur kerja sesuai tipe
func (w *WorkflowControlNode) executeControl(ctx context.Context, controlType string, input map[string]interface{}) (map[string]interface{}, error) {
	// Simulasikan waktu pemrosesan
	time.Sleep(80 * time.Millisecond)

	var result map[string]interface{}

	switch controlType {
	case "sequence":
		result = w.handleSequence(input)
	case "branch":
		result = w.handleBranch(input)
	case "merge":
		result = w.handleMerge(input)
	case "parallel":
		result = w.handleParallel(input)
	case "sync":
		result = w.handleSync(input)
	case "wait":
		result = w.handleWait(input)
	case "loop":
		result = w.handleLoop(input)
	case "condition":
		result = w.handleConditional(input)
	default:
		result = w.handleDefault(input)
	}

	result["control_type_executed"] = controlType
	result["processing_time"] = time.Since(time.Now().Add(-80 * time.Millisecond)).Seconds()
	result["timestamp"] = time.Now().Unix()

	return result, nil
}

// handleSequence menangani kontrol urutan
func (w *WorkflowControlNode) handleSequence(input map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{
		"action":         "sequence_control",
		"control_type":   "sequence",
		"description":    "Melanjutkan ke node berikutnya dalam urutan",
		"input_forwarded": true,
		"next_node":      w.config.NextNodeID,
		"sequence_step":  1,
	}

	return result
}

// handleBranch menangani kontrol percabangan
func (w *WorkflowControlNode) handleBranch(input map[string]interface{}) map[string]interface{} {
	conditionMet := true // Simulasikan bahwa kondisi terpenuhi
	
	// Evaluasi kondisi jika ada
	if w.config.Condition != "" {
		conditionMet = w.evaluateCondition(w.config.Condition, input)
	}

	result := map[string]interface{}{
		"action":        "branch_control",
		"control_type":  "branch",
		"condition":     w.config.Condition,
		"condition_met": conditionMet,
		"branch_taken":  "true_branch", // Atau "false_branch" tergantung hasil
		"description":   "Mengarahkan alur kerja berdasarkan kondisi",
		"condition_result": conditionMet,
	}

	if conditionMet {
		result["next_node"] = w.config.NextNodeID
	} else {
		// Dalam implementasi nyata, ini akan mengarah ke node alternatif
		result["next_node"] = "false_branch_node"
	}

	return result
}

// handleMerge menangani kontrol penggabungan
func (w *WorkflowControlNode) handleMerge(input map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{
		"action":       "merge_control",
		"control_type": "merge",
		"description":  "Menggabungkan beberapa alur kerja menjadi satu",
		"merged_data":  input,
		"merge_type":   "sequential", // Atau "concurrent" tergantung implementasi
		"next_node":    w.config.NextNodeID,
	}

	return result
}

// handleParallel menangani kontrol paralel
func (w *WorkflowControlNode) handleParallel(input map[string]interface{}) map[string]interface{} {
	branches := w.config.ParallelBranches
	if len(branches) == 0 {
		branches = []string{"branch_1", "branch_2"} // Default branches
	}

	result := map[string]interface{}{
		"action":          "parallel_control",
		"control_type":    "parallel",
		"description":     "Menjalankan beberapa cabang secara paralel",
		"branches_count":  len(branches),
		"branches":        branches,
		"next_node":       w.config.NextNodeID,
		"parallelism":     true,
	}

	return result
}

// handleSync menangani kontrol sinkronisasi
func (w *WorkflowControlNode) handleSync(input map[string]interface{}) map[string]interface{} {
	syncPoints := w.config.SyncPoints
	if len(syncPoints) == 0 {
		syncPoints = []string{"sync_point_1"} // Default sync point
	}

	result := map[string]interface{}{
		"action":         "sync_control",
		"control_type":   "sync",
		"description":    "Menyinkronkan beberapa cabang eksekusi",
		"sync_points":    syncPoints,
		"sync_achieved":  true,
		"next_node":      w.config.NextNodeID,
		"awaiting_nodes": []string{}, // Dalam implementasi nyata, ini akan berisi node yang ditunggu
		"synchronized_at": time.Now().Unix(),
	}

	return result
}

// handleWait menangani kontrol tunggu
func (w *WorkflowControlNode) handleWait(input map[string]interface{}) map[string]interface{} {
	waitConditions := w.config.WaitConditions
	if len(waitConditions) == 0 {
		waitConditions = []WaitCondition{
			{
				Type:     "time",
				Resource: "duration",
				Timeout:  5,
				Value:    5,
			},
		}
	}

	result := map[string]interface{}{
		"action":           "wait_control",
		"control_type":     "wait",
		"description":      "Menunggu kondisi tertentu sebelum melanjutkan",
		"wait_conditions":  waitConditions,
		"wait_completed":   true,
		"wait_duration":    2.5, // Simulasikan durasi tunggu
		"next_node":        w.config.NextNodeID,
	}

	return result
}

// handleLoop menangani kontrol perulangan
func (w *WorkflowControlNode) handleLoop(input map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{
		"action":         "loop_control",
		"control_type":   "loop",
		"description":    "Mengelola perulangan dalam alur kerja",
		"loop_condition": w.config.Condition,
		"next_node":      w.config.NextNodeID,
		"continue_loop":  false, // Ditentukan oleh kondisi sebenarnya
		"loop_count":     1,     // Jumlah iterasi
	}

	// Evaluasi apakah loop harus dilanjutkan
	continueLoop := true
	if w.config.Condition != "" {
		continueLoop = w.evaluateCondition(w.config.Condition, input)
	}
	result["continue_loop"] = continueLoop

	return result
}

// handleConditional menangani kontrol kondisional
func (w *WorkflowControlNode) handleConditional(input map[string]interface{}) map[string]interface{} {
	conditionMet := true
	if w.config.Condition != "" {
		conditionMet = w.evaluateCondition(w.config.Condition, input)
	}

	result := map[string]interface{}{
		"action":            "conditional_control",
		"control_type":      "condition",
		"evaluated_condition": w.config.Condition,
		"condition_result":  conditionMet,
		"next_node":         w.config.NextNodeID,
		"condition_met":     conditionMet,
		"description":       "Melakukan evaluasi kondisi untuk menentukan alur berikutnya",
	}

	return result
}

// handleDefault menangani kontrol default
func (w *WorkflowControlNode) handleDefault(input map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{
		"action":      "default_control",
		"control_type": "default",
		"description": "Kontrol default ketika tipe kontrol tidak dikenali",
		"next_node":   w.config.NextNodeID,
		"input_passed": true,
	}

	return result
}

// evaluateCondition mengevaluasi ekspresi kondisi
func (w *WorkflowControlNode) evaluateCondition(condition string, input map[string]interface{}) bool {
	// Dalam implementasi nyata, ini akan mengevaluasi ekspresi kondisi kompleks
	// Untuk simulasi, kita buat beberapa kondisi sederhana
	
	// Jika kondisi adalah nama field dalam input, cek apakah nil atau tidak
	if value, exists := input[condition]; exists {
		// Jika nil, kondisi tidak terpenuhi; jika tidak nil, kondisi terpenuhi
		return value != nil
	}
	
	// Jika kondisi adalah "always_true", selalu true
	if condition == "always_true" {
		return true
	}
	
	// Jika kondisi adalah "always_false", selalu false
	if condition == "always_false" {
		return false
	}
	
	// Default: anggap kondisi terpenuhi
	return true
}

// GetType mengembalikan jenis node
func (w *WorkflowControlNode) GetType() string {
	return "workflow_control"
}

// GetID mengembalikan ID unik untuk instance node
func (w *WorkflowControlNode) GetID() string {
	return "wf_ctrl_" + fmt.Sprintf("%d", time.Now().Unix())
}

// RegisterWorkflowControlNode mendaftarkan node Workflow Control dengan engine
func RegisterWorkflowControlNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("workflow_control", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return NewWorkflowControlNode(config)
	})
}