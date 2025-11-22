package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"citadel-agent/backend/internal/workflow/core/engine"
)

// LoopProcessorConfig mewakili konfigurasi untuk node Loop Processor
type LoopProcessorConfig struct {
	ItemsSource     string                   `json:"items_source"`      // Sumber item untuk diiterasi (nama field dalam input)
	MaxIterations   int                      `json:"max_iterations"`    // Jumlah maksimum iterasi
	MaxConcurrency  int                      `json:"max_concurrency"`   // Jumlah maksimum eksekusi concurrently
	EnableParallel  bool                     `json:"enable_parallel"`   // Apakah mengaktifkan eksekusi paralel
	TimeoutPerItem  int                      `json:"timeout_per_item"`  // Waktu timeout per item dalam detik
	EnableBreak     bool                     `json:"enable_break"`      // Apakah mengaktifkan kemampuan untuk break loop
	BreakCondition  string                   `json:"break_condition"`   // Kondisi untuk break loop
	EnableContinue  bool                     `json:"enable_continue"`   // Apakah mengaktifkan kemampuan untuk continue ke iterasi berikutnya
	ContinueCondition string                 `json:"continue_condition"` // Kondisi untuk continue ke iterasi berikutnya
	AccumulateResults bool                   `json:"accumulate_results"` // Apakah mengumpulkan hasil dari semua iterasi
	ResultKey       string                   `json:"result_key"`       // Nama field untuk menyimpan hasil
	EnableCaching   bool                     `json:"enable_caching"`   // Apakah mengaktifkan caching hasil
	CacheTTL        int                      `json:"cache_ttl"`        // Waktu cache dalam detik
	EnableProfiling bool                     `json:"enable_profiling"` // Apakah mengaktifkan profiling
	Timeout         int                      `json:"timeout"`          // Waktu timeout total dalam detik
	ReturnRawResults bool                    `json:"return_raw_results"` // Apakah mengembalikan hasil mentah
	CustomParams    map[string]interface{}   `json:"custom_params"`    // Parameter khusus untuk loop
	Preprocessing   PreprocessingConfig      `json:"preprocessing"`    // Konfigurasi pra-pemrosesan
	Postprocessing  PostprocessingConfig     `json:"postprocessing"`   // Konfigurasi pasca-pemrosesan
}

// LoopIterationResult mewakili hasil dari satu iterasi loop
type LoopIterationResult struct {
	Index       int         `json:"index"`        // Indeks iterasi
	Item        interface{} `json:"item"`         // Item yang diproses
	Result      interface{} `json:"result"`       // Hasil pemrosesan
	Success     bool        `json:"success"`      // Apakah iterasi berhasil
	Error       string      `json:"error"`        // Kesalahan jika ada
	ProcessingTime float64  `json:"processing_time"` // Waktu pemrosesan
	Timestamp   int64       `json:"timestamp"`    // Waktu eksekusi
}

// LoopProcessorNode mewakili node yang memproses perulangan
type LoopProcessorNode struct {
	config *LoopProcessorConfig
}

// NewLoopProcessorNode membuat node Loop Processor baru
func NewLoopProcessorNode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Konversi map interface{} ke JSON lalu ke struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("gagal mengubah konfig menjadi json: %v", err)
	}

	var loopConfig LoopProcessorConfig
	err = json.Unmarshal(jsonData, &loopConfig)
	if err != nil {
		return nil, fmt.Errorf("gagal menguraikan konfig: %v", err)
	}

	// Validasi dan atur default
	if loopConfig.MaxIterations == 0 {
		loopConfig.MaxIterations = 100
	}

	if loopConfig.MaxConcurrency == 0 {
		loopConfig.MaxConcurrency = 1 // default ke eksekusi serial
	}

	if loopConfig.Timeout == 0 {
		loopConfig.Timeout = 300 // default timeout 300 detik
	}

	if loopConfig.TimeoutPerItem == 0 {
		loopConfig.TimeoutPerItem = 30 // default 30 detik per item
	}

	if loopConfig.ResultKey == "" {
		loopConfig.ResultKey = "loop_results"
	}

	return &LoopProcessorNode{
		config: &loopConfig,
	}, nil
}

// Execute mengimplementasikan interface NodeInstance
func (l *LoopProcessorNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Timpa konfigurasi dengan nilai input jika disediakan
	itemsSource := l.config.ItemsSource
	if inputItemsSource, ok := input["items_source"].(string); ok && inputItemsSource != "" {
		itemsSource = inputItemsSource
	}

	maxIterations := l.config.MaxIterations
	if inputMaxIterations, ok := input["max_iterations"].(float64); ok {
		maxIterations = int(inputMaxIterations)
	}

	maxConcurrency := l.config.MaxConcurrency
	if inputMaxConcurrency, ok := input["max_concurrency"].(float64); ok {
		maxConcurrency = int(inputMaxConcurrency)
	}

	enableParallel := l.config.EnableParallel
	if inputEnableParallel, ok := input["enable_parallel"].(bool); ok {
		enableParallel = inputEnableParallel
	}

	timeoutPerItem := l.config.TimeoutPerItem
	if inputTimeoutPerItem, ok := input["timeout_per_item"].(float64); ok {
		timeoutPerItem = int(inputTimeoutPerItem)
	}

	enableBreak := l.config.EnableBreak
	if inputEnableBreak, ok := input["enable_break"].(bool); ok {
		enableBreak = inputEnableBreak
	}

	breakCondition := l.config.BreakCondition
	if inputBreakCondition, ok := input["break_condition"].(string); ok && inputBreakCondition != "" {
		breakCondition = inputBreakCondition
	}

	enableContinue := l.config.EnableContinue
	if inputEnableContinue, ok := input["enable_continue"].(bool); ok {
		enableContinue = inputEnableContinue
	}

	continueCondition := l.config.ContinueCondition
	if inputContinueCondition, ok := input["continue_condition"].(string); ok && inputContinueCondition != "" {
		continueCondition = inputContinueCondition
	}

	accumulateResults := l.config.AccumulateResults
	if inputAccumulateResults, ok := input["accumulate_results"].(bool); ok {
		accumulateResults = inputAccumulateResults
	}

	resultKey := l.config.ResultKey
	if inputResultKey, ok := input["result_key"].(string); ok && inputResultKey != "" {
		resultKey = inputResultKey
	}

	enableCaching := l.config.EnableCaching
	if inputEnableCaching, ok := input["enable_caching"].(bool); ok {
		enableCaching = inputEnableCaching
	}

	cacheTTL := l.config.CacheTTL
	if inputCacheTTL, ok := input["cache_ttl"].(float64); ok {
		cacheTTL = int(inputCacheTTL)
	}

	enableProfiling := l.config.EnableProfiling
	if inputEnableProfiling, ok := input["enable_profiling"].(bool); ok {
		enableProfiling = inputEnableProfiling
	}

	timeout := l.config.Timeout
	if inputTimeout, ok := input["timeout"].(float64); ok {
		timeout = int(inputTimeout)
	}

	returnRawResults := l.config.ReturnRawResults
	if inputReturnRaw, ok := input["return_raw_results"].(bool); ok {
		returnRawResults = inputReturnRaw
	}

	customParams := l.config.CustomParams
	if inputCustomParams, ok := input["custom_params"].(map[string]interface{}); ok {
		customParams = inputCustomParams
	}

	// Validasi input
	if len(input) == 0 {
		return map[string]interface{}{
			"success":   false,
			"error":     "input diperlukan untuk loop processor",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Dapatkan items untuk diiterasi
	items := []interface{}{}
	
	if itemsSource != "" {
		// Ambil dari field yang dispesifikasikan
		if sourceItems, exists := input[itemsSource]; exists {
			switch v := sourceItems.(type) {
			case []interface{}:
				items = v
			case []map[string]interface{}:
				// Konversi ke []interface{}
				for _, item := range v {
					items = append(items, item)
				}
			case []string:
				// Konversi ke []interface{}
				for _, item := range v {
					items = append(items, item)
				}
			case []int:
				// Konversi ke []interface{}
				for _, item := range v {
					items = append(items, item)
				}
			case []float64:
				// Konversi ke []interface{}
				for _, item := range v {
					items = append(items, item)
				}
			default:
				// Jika bukan array, perlakukan sebagai satu item
				items = []interface{}{v}
			}
		}
	} else {
		// Jika tidak ada items_source, gunakan seluruh input sebagai satu item
		items = []interface{}{input}
	}

	// Batasi jumlah item ke maxIterations
	if len(items) > maxIterations {
		items = items[:maxIterations]
	}

	// Buat konteks operasi dengan timeout
	loopCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Proses loop
	loopResult, err := l.processLoop(loopCtx, items, input)
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
		"items_source":         itemsSource,
		"total_items":          len(items),
		"max_iterations":       maxIterations,
		"max_concurrency":      maxConcurrency,
		"enable_parallel":      enableParallel,
		"timeout_per_item":     timeoutPerItem,
		"enable_break":         enableBreak,
		"break_condition":      breakCondition,
		"enable_continue":      enableContinue,
		"continue_condition":   continueCondition,
		"accumulate_results":   accumulateResults,
		"result_key":           resultKey,
		"loop_result":          loopResult,
		"enable_caching":       enableCaching,
		"enable_profiling":     enableProfiling,
		"return_raw_results":   returnRawResults,
		"timestamp":            time.Now().Unix(),
		"input_data":           input,
		"config":               l.config,
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

// processLoop memproses loop melalui items
func (l *LoopProcessorNode) processLoop(ctx context.Context, items []interface{}, originalInput map[string]interface{}) (map[string]interface{}, error) {
	var results []LoopIterationResult
	var allResults []interface{}

	if l.config.EnableParallel && l.config.MaxConcurrency > 1 {
		// Eksekusi paralel
		results = l.processItemsParallel(ctx, items, originalInput)
	} else {
		// Eksekusi serial
		results = l.processItemsSerial(ctx, items, originalInput)
	}

	// Kumpulkan hasil
	for _, result := range results {
		if result.Success {
			allResults = append(allResults, result.Result)
		}
	}

	// Hitung statistik
	successfulIterations := 0
	failedIterations := 0
	totalProcessingTime := float64(0)

	for _, result := range results {
		if result.Success {
			successfulIterations++
		} else {
			failedIterations++
		}
		totalProcessingTime += result.ProcessingTime
	}

	loopResult := map[string]interface{}{
		"iterations":           results,
		"successful_iterations": successfulIterations,
		"failed_iterations":    failedIterations,
		"total_iterations":     len(results),
		"all_results":          allResults,
		"average_processing_time": totalProcessingTime / float64(len(results)),
		"total_processing_time": totalProcessingTime,
		"timestamp":            time.Now().Unix(),
	}

	return loopResult, nil
}

// processItemsSerial memproses item-item secara serial
func (l *LoopProcessorNode) processItemsSerial(ctx context.Context, items []interface{}, originalInput map[string]interface{}) []LoopIterationResult {
	var results []LoopIterationResult

	for i, item := range items {
		// Periksa apakah konteks sudah habis
		select {
		case <-ctx.Done():
			// Tandai iterasi yang tidak selesai karena timeout
			results = append(results, LoopIterationResult{
				Index:  i,
				Item:   item,
				Success: false,
				Error:   "context cancelled",
				Timestamp: time.Now().Unix(),
			})
			continue
		default:
			// Lanjutkan eksekusi
		}

		// Buat input untuk iterasi ini
		iterationInput := l.createIterationInput(i, item, originalInput)

		// Proses iterasi
		result := l.processIteration(ctx, i, item, iterationInput)
		results = append(results, result)

		// Periksa kondisi break
		if l.config.EnableBreak && l.evaluateBreakCondition(&result) {
			break
		}
	}

	return results
}

// processItemsParallel memproses item-item secara paralel
func (l *LoopProcessorNode) processItemsParallel(ctx context.Context, items []interface{}, originalInput map[string]interface{}) []LoopIterationResult {
	results := make([]LoopIterationResult, len(items))
	
	// Gunakan semaphore untuk membatasi konkurensi
	semaphore := make(chan struct{}, l.config.MaxConcurrency)
	
	// Channel untuk mengumpulkan hasil
	resultChan := make(chan LoopIterationResult, len(items))
	
	// Proses setiap item di goroutine
	for i, item := range items {
		go func(index int, itemData interface{}) {
			// Ambil slot dari semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }() // Kembalikan slot ke semaphore

			// Periksa apakah konteks sudah habis
			select {
			case <-ctx.Done():
				resultChan <- LoopIterationResult{
					Index:  index,
					Item:   itemData,
					Success: false,
					Error:   "context cancelled",
					Timestamp: time.Now().Unix(),
				}
				return
			default:
				// Lanjutkan eksekusi
			}

			// Buat input untuk iterasi ini
			iterationInput := l.createIterationInput(index, itemData, originalInput)

			// Proses iterasi
			result := l.processIteration(ctx, index, itemData, iterationInput)
			resultChan <- result
		}(i, item)
	}

	// Kumpulkan hasil dari semua goroutine
	for i := 0; i < len(items); i++ {
		result := <-resultChan
		// Masukkan hasil ke dalam slice pada indeks yang benar
		results[result.Index] = result
	}

	return results
}

// createIterationInput membuat input untuk iterasi tertentu
func (l *LoopProcessorNode) createIterationInput(index int, item interface{}, originalInput map[string]interface{}) map[string]interface{} {
	iterationInput := make(map[string]interface{})
	
	// Salin input asli
	for k, v := range originalInput {
		iterationInput[k] = v
	}
	
	// Tambahkan informasi iterasi
	iterationInput["item"] = item
	iterationInput["current_item"] = item
	iterationInput["index"] = index
	iterationInput["iteration"] = index
	iterationInput["loop_index"] = index
	
	return iterationInput
}

// processIteration memproses satu iterasi loop
func (l *LoopProcessorNode) processIteration(ctx context.Context, index int, item interface{}, input map[string]interface{}) LoopIterationResult {
	startTime := time.Now()
	
	// Simulasikan pemrosesan item
	// Dalam implementasi nyata, ini akan memanggil node lain atau melakukan pemrosesan kompleks
	result := map[string]interface{}{
		"processed_item": item,
		"index": index,
		"processed_at": time.Now().Unix(),
		"item_hash": fmt.Sprintf("%v_%d", item, index),
	}
	
	// Simulasikan kemungkinan kesalahan berdasarkan indeks
	errorMsg := ""
	success := true
	
	if isError, exists := input["simulate_error_for_index"]; exists {
		if errorIndex, ok := isError.(float64); ok {
			if int(errorIndex) == index {
				errorMsg = "Simulated error for index " + fmt.Sprintf("%d", index)
				success = false
			}
		}
	}

	iterationResult := LoopIterationResult{
		Index:          index,
		Item:           item,
		Result:         result,
		Success:        success,
		Error:          errorMsg,
		ProcessingTime: time.Since(startTime).Seconds(),
		Timestamp:      time.Now().Unix(),
	}
	
	return iterationResult
}

// evaluateBreakCondition mengevaluasi apakah kondisi break terpenuhi
func (l *LoopProcessorNode) evaluateBreakCondition(result *LoopIterationResult) bool {
	// Dalam implementasi nyata, ini akan mengevaluasi ekspresi break_condition
	// Untuk simulasi, kita hanya akan return false
	return false
}

// GetType mengembalikan jenis node
func (l *LoopProcessorNode) GetType() string {
	return "loop_processor"
}

// GetID mengembalikan ID unik untuk instance node
func (l *LoopProcessorNode) GetID() string {
	return "loop_" + fmt.Sprintf("%d", time.Now().Unix())
}

// RegisterLoopProcessorNode mendaftarkan node Loop Processor dengan engine
func RegisterLoopProcessorNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("loop_processor", func(config map[string]interface{}) (engine.NodeInstance, error) {
		return NewLoopProcessorNode(config)
	})
}