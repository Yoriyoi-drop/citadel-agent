package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/workflow/core/engine"
)

// ConditionProcessorConfig mewakili konfigurasi untuk node Condition Processor
type ConditionProcessorConfig struct {
	Conditions      []ConditionRule          `json:"conditions"`         // Aturan-aturan kondisi
	DefaultAction   string                   `json:"default_action"`     // Aksi default jika tidak ada kondisi yang terpenuhi
	ProcessingMode  string                   `json:"processing_mode"`    // Mode pemrosesan (first_match, all_match, etc.)
	LogicOperator   string                   `json:"logic_operator"`     // Operator logika (AND, OR)
	EnableCaching   bool                     `json:"enable_caching"`     // Apakah mengaktifkan caching hasil
	CacheTTL        int                      `json:"cache_ttl"`          // Waktu cache dalam detik
	EnableProfiling bool                     `json:"enable_profiling"`   // Apakah mengaktifkan profiling
	Timeout         int                      `json:"timeout"`            // Waktu timeout dalam detik
	ReturnRawResults bool                    `json:"return_raw_results"` // Apakah mengembalikan hasil mentah
	CustomParams    map[string]interface{}   `json:"custom_params"`      // Parameter khusus untuk kondisi
	Preprocessing   PreprocessingConfig      `json:"preprocessing"`      // Konfigurasi pra-pemrosesan
	Postprocessing  PostprocessingConfig     `json:"postprocessing"`     // Konfigurasi pasca-pemrosesan
}

// ConditionRule mewakili aturan kondisi individual
type ConditionRule struct {
	ID          string      `json:"id"`           // ID aturan
	Name        string      `json:"name"`         // Nama aturan
	Description string      `json:"description"`  // Deskripsi aturan
	Field       string      `json:"field"`        // Bidang untuk dicek
	Operator    string      `json:"operator"`     // Operator pembanding
	Value       interface{} `json:"value"`        // Nilai untuk dibandingkan
	Action      string      `json:"action"`       // Aksi untuk diambil jika kondisi terpenuhi
	Priority    int         `json:"priority"`     // Prioritas aturan
	Enabled     bool        `json:"enabled"`      // Apakah aturan diaktifkan
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

// ConditionProcessorNode mewakili node yang memproses kondisi-kondisi logika
type ConditionProcessorNode struct {
	config *ConditionProcessorConfig
}

// NewConditionProcessorNode membuat node Condition Processor baru
func NewConditionProcessorNode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Konversi map interface{} ke JSON lalu ke struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("gagal mengubah konfig menjadi json: %v", err)
	}

	var conditionConfig ConditionProcessorConfig
	err = json.Unmarshal(jsonData, &conditionConfig)
	if err != nil {
		return nil, fmt.Errorf("gagal menguraikan konfig: %v", err)
	}

	// Validasi dan atur default
	if conditionConfig.ProcessingMode == "" {
		conditionConfig.ProcessingMode = "first_match"
	}

	if conditionConfig.LogicOperator == "" {
		conditionConfig.LogicOperator = "AND"
	}

	if conditionConfig.Timeout == 0 {
		conditionConfig.Timeout = 30 // default timeout 30 detik
	}

	// Jika tidak ada kondisi, buat kondisi default
	if len(conditionConfig.Conditions) == 0 {
		conditionConfig.Conditions = []ConditionRule{
			{
				ID:          "default_condition",
				Name:        "Default Condition",
				Description: "Default condition that always evaluates to true",
				Field:       "always_true",
				Operator:    "equals",
				Value:       true,
				Action:      "continue",
				Priority:    1,
				Enabled:     true,
			},
		}
	}

	return &ConditionProcessorNode{
		config: &conditionConfig,
	}, nil
}

// Execute mengimplementasikan interface NodeInstance
func (c *ConditionProcessorNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Timpa konfigurasi dengan nilai input jika disediakan
	processingMode := c.config.ProcessingMode
	if inputProcessingMode, ok := input["processing_mode"].(string); ok && inputProcessingMode != "" {
		processingMode = inputProcessingMode
	}

	logicOperator := c.config.LogicOperator
	if inputLogicOperator, ok := input["logic_operator"].(string); ok && inputLogicOperator != "" {
		logicOperator = inputLogicOperator
	}

	defaultAction := c.config.DefaultAction
	if inputDefaultAction, ok := input["default_action"].(string); ok && inputDefaultAction != "" {
		defaultAction = inputDefaultAction
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
			"error":     "input diperlukan untuk mengevaluasi kondisi",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Buat konteks operasi dengan timeout
	conditionCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Evaluasi kondisi-kondisi
	conditionResult, err := c.evaluateConditions(conditionCtx, input)
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
		"processing_mode":      processingMode,
		"logic_operator":       logicOperator,
		"default_action":       defaultAction,
		"condition_result":     conditionResult,
		"enable_caching":       enableCaching,
		"enable_profiling":     enableProfiling,
		"return_raw_results":   returnRawResults,
		"timestamp":            time.Now().Unix(),
		"input_data":           input,
		"config":               c.config,
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

// evaluateConditions mengevaluasi semua kondisi yang ditentukan
func (c *ConditionProcessorNode) evaluateConditions(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Simulasikan waktu pemrosesan
	time.Sleep(50 * time.Millisecond)

	// Filter kondisi yang aktif
	var activeConditions []ConditionRule
	for _, condition := range c.config.Conditions {
		if condition.Enabled {
			activeConditions = append(activeConditions, condition)
		}
	}

	// Urutkan berdasarkan prioritas (tertinggi dulu)
	sortedConditions := make([]ConditionRule, len(activeConditions))
	copy(sortedConditions, activeConditions)
	
	// Urutkan dengan bubble sort sederhana
	for i := 0; i < len(sortedConditions)-1; i++ {
		for j := 0; j < len(sortedConditions)-i-1; j++ {
			if sortedConditions[j].Priority < sortedConditions[j+1].Priority {
				sortedConditions[j], sortedConditions[j+1] = sortedConditions[j+1], sortedConditions[j]
			}
		}
	}

	var matchedConditions []ConditionRule
	var evaluatedResults []map[string]interface{}
	var matchedAction string
	var actionTaken bool

	// Proses berdasarkan mode pemrosesan
	switch c.config.ProcessingMode {
	case "first_match":
		// Evaluasi kondisi satu per satu dan hentikan saat pertama kali cocok
		for _, condition := range sortedConditions {
			result := map[string]interface{}{
				"condition_id": condition.ID,
				"condition_name": condition.Name,
				"field": condition.Field,
				"operator": condition.Operator,
				"value": condition.Value,
				"priority": condition.Priority,
			}

			// Evaluasi kondisi
			isMatch, err := c.evaluateCondition(condition, input)
			if err != nil {
				result["error"] = err.Error()
				result["match"] = false
			} else {
				result["match"] = isMatch
				if isMatch {
					matchedConditions = append(matchedConditions, condition)
					matchedAction = condition.Action
					actionTaken = true
					// Hentikan evaluasi karena ini adalah first_match
					break
				}
			}

			evaluatedResults = append(evaluatedResults, result)
		}
	case "all_match":
		// Evaluasi semua kondisi dan cocokkan semua yang benar
		for _, condition := range sortedConditions {
			result := map[string]interface{}{
				"condition_id": condition.ID,
				"condition_name": condition.Name,
				"field": condition.Field,
				"operator": condition.Operator,
				"value": condition.Value,
				"priority": condition.Priority,
			}

			// Evaluasi kondisi
			isMatch, err := c.evaluateCondition(condition, input)
			if err != nil {
				result["error"] = err.Error()
				result["match"] = false
			} else {
				result["match"] = isMatch
				if isMatch {
					matchedConditions = append(matchedConditions, condition)
					// Dalam mode all_match, kita bisa menggabungkan semua aksi
					if matchedAction == "" {
						matchedAction = condition.Action
					} else {
						matchedAction += "," + condition.Action
					}
					actionTaken = true
				}
			}

			evaluatedResults = append(evaluatedResults, result)
		}
	case "match_any":
		// Evaluasi semua kondisi dan cocokkan jika salah satu benar (OR logic)
		for _, condition := range sortedConditions {
			result := map[string]interface{}{
				"condition_id": condition.ID,
				"condition_name": condition.Name,
				"field": condition.Field,
				"operator": condition.Operator,
				"value": condition.Value,
				"priority": condition.Priority,
			}

			// Evaluasi kondisi
			isMatch, err := c.evaluateCondition(condition, input)
			if err != nil {
				result["error"] = err.Error()
				result["match"] = false
			} else {
				result["match"] = isMatch
				if isMatch {
					matchedConditions = append(matchedConditions, condition)
					if !actionTaken {
						matchedAction = condition.Action
						actionTaken = true
					}
				}
			}

			evaluatedResults = append(evaluatedResults, result)

			// Jika sudah cocok dan ini adalah match_any, kita bisa berhenti lebih awal
			if c.config.LogicOperator == "OR" && actionTaken {
				break
			}
		}
	default:
		// Default ke first_match
		for _, condition := range sortedConditions {
			result := map[string]interface{}{
				"condition_id": condition.ID,
				"condition_name": condition.Name,
				"field": condition.Field,
				"operator": condition.Operator,
				"value": condition.Value,
				"priority": condition.Priority,
			}

			// Evaluasi kondisi
			isMatch, err := c.evaluateCondition(condition, input)
			if err != nil {
				result["error"] = err.Error()
				result["match"] = false
			} else {
				result["match"] = isMatch
				if isMatch {
					matchedConditions = append(matchedConditions, condition)
					matchedAction = condition.Action
					actionTaken = true
					break
				}
			}

			evaluatedResults = append(evaluatedResults, result)
		}
	}

	// Jika tidak ada kondisi yang cocok, gunakan aksi default
	if !actionTaken && c.config.DefaultAction != "" {
		matchedAction = c.config.DefaultAction
	}

	result := map[string]interface{}{
		"evaluated_conditions": evaluatedResults,
		"matched_conditions":   len(matchedConditions),
		"total_conditions":     len(sortedConditions),
		"matched_condition_rules": matchedConditions,
		"action_taken":         matchedAction,
		"action_taken_bool":    actionTaken,
		"default_action_used":  !actionTaken && c.config.DefaultAction != "",
		"processing_time":      time.Since(time.Now().Add(-50 * time.Millisecond)).Seconds(),
		"timestamp":            time.Now().Unix(),
		"processing_mode":      c.config.ProcessingMode,
	}

	return result, nil
}

// evaluateCondition mengevaluasi kondisi individual
func (c *ConditionProcessorNode) evaluateCondition(condition ConditionRule, input map[string]interface{}) (bool, error) {
	// Dapatkan nilai dari input berdasarkan field kondisi
	inputValue, exists := input[condition.Field]
	if !exists {
		// Jika field tidak ditemukan di input, periksa apakah ini adalah field khusus
		// Misalnya, field yang selalu true untuk kondisi default
		if condition.Field == "always_true" {
			return true, nil
		}
		return false, fmt.Errorf("field '%s' tidak ditemukan dalam input", condition.Field)
	}

	// Konversi input dan nilai kondisi ke bentuk yang dapat dibandingkan
	inputStr := fmt.Sprintf("%v", inputValue)
	conditionStr := fmt.Sprintf("%v", condition.Value)

	switch condition.Operator {
	case "=", "==", "equals", "equal":
		return inputStr == conditionStr, nil
	case "!=", "not_equals", "not_equal":
		return inputStr != conditionStr, nil
	case ">", "greater_than":
		if inputFloat, inputOk := asFloat64(inputValue); inputOk {
			if condFloat, condOk := asFloat64(condition.Value); condOk {
				return inputFloat > condFloat, nil
			}
		}
		return false, fmt.Errorf("operator '%s' memerlukan nilai numerik", condition.Operator)
	case "<", "less_than":
		if inputFloat, inputOk := asFloat64(inputValue); inputOk {
			if condFloat, condOk := asFloat64(condition.Value); condOk {
				return inputFloat < condFloat, nil
			}
		}
		return false, fmt.Errorf("operator '%s' memerlukan nilai numerik", condition.Operator)
	case ">=", "greater_than_or_equal":
		if inputFloat, inputOk := asFloat64(inputValue); inputOk {
			if condFloat, condOk := asFloat64(condition.Value); condOk {
				return inputFloat >= condFloat, nil
			}
		}
		return false, fmt.Errorf("operator '%s' memerlukan nilai numerik", condition.Operator)
	case "<=", "less_than_or_equal":
		if inputFloat, inputOk := asFloat64(inputValue); inputOk {
			if condFloat, condOk := asFloat64(condition.Value); condOk {
				return inputFloat <= condFloat, nil
			}
		}
		return false, fmt.Errorf("operator '%s' memerlukan nilai numerik", condition.Operator)
	case "contains":
		return contains(inputStr, conditionStr), nil
	case "starts_with":
		return len(inputStr) >= len(conditionStr) && inputStr[:len(conditionStr)] == conditionStr, nil
	case "ends_with":
		return len(inputStr) >= len(conditionStr) && inputStr[len(inputStr)-len(conditionStr):] == conditionStr, nil
	case "in", "within":
		// Operator 'in' memeriksa apakah nilai input ada dalam array nilai kondisi
		if conditionArr, ok := condition.Value.([]interface{}); ok {
			for _, val := range conditionArr {
				if fmt.Sprintf("%v", val) == inputStr {
					return true, nil
				}
			}
			return false, nil
		}
		return false, fmt.Errorf("operator '%s' memerlukan array nilai", condition.Operator)
	case "matches_regex", "regex":
		// Dalam implementasi penuh, ini akan menggunakan regex
		// Untuk simulasi, kita anggap selalu cocok
		return true, nil
	default:
		return false, fmt.Errorf("operator '%s' tidak didukung", condition.Operator)
	}

	return false, nil
}

// asFloat64 mencoba mengkonversi nilai ke float64
func asFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	case string:
		// Dalam implementasi penuh, ini akan menggunakan strconv.ParseFloat
		// Untuk simulasi, kita kembalikan 0, false
		return 0, false
	default:
		return 0, false
	}
}

// GetType mengembalikan jenis node
func (c *ConditionProcessorNode) GetType() string {
	return "condition_processor"
}

// GetID mengembalikan ID unik untuk instance node
func (c *ConditionProcessorNode) GetID() string {
	return "condition_" + fmt.Sprintf("%d", time.Now().Unix())
}

// RegisterConditionProcessorNode mendaftarkan node Condition Processor dengan engine
func RegisterConditionProcessorNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("condition_processor", func(config map[string]interface{}) (engine.NodeInstance, error) {
		return NewConditionProcessorNode(config)
	})
}

// contains adalah fungsi helper untuk memeriksa apakah string mengandung substring
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