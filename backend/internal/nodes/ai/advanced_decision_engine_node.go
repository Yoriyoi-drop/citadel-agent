package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
	"github.com/citadel-agent/backend/internal/nodes/utils"
	"github.com/citadel-agent/backend/internal/workflow/core/engine"
)

// AdvancedDecisionEngineConfig mewakili konfigurasi untuk node Advanced Decision Engine
type AdvancedDecisionEngineConfig struct {
	Provider           string                 `json:"provider"`            // Penyedia engine keputusan
	DecisionType       string                 `json:"decision_type"`       // Jenis keputusan (approval, classification, routing, dll.)
	ModelName          string                 `json:"model_name"`          // Nama model keputusan
	APIKey             string                 `json:"api_key"`             // Kunci API untuk layanan
	Rules              []DecisionRule         `json:"rules"`               // Kumpulan aturan keputusan
	EnableMachineLearning bool                `json:"enable_machine_learning"` // Apakah mengaktifkan ML
	MLModel            string                 `json:"ml_model"`            // Model ML untuk keputusan hybrid
	ConfidenceThreshold float64               `json:"confidence_threshold"` // Ambang batas kepercayaan
	FallbackStrategy   string                 `json:"fallback_strategy"`   // Strategi cadangan (rule_based, manual, default_action)
	EnableExplainability bool                `json:"enable_explainability"` // Apakah mengaktifkan penjelasan
	TraceID            string                 `json:"trace_id"`            // ID untuk melacak keputusan
	EnableCaching      bool                   `json:"enable_caching"`      // Apakah mengaktifkan caching keputusan
	CacheTTL           int                    `json:"cache_ttl"`           // Waktu cache dalam detik
	EnableProfiling    bool                   `json:"enable_profiling"`    // Apakah mengaktifkan profiling
	Timeout            int                    `json:"timeout"`             // Waktu timeout dalam detik
	ReturnRawResults   bool                   `json:"return_raw_results"`  // Apakah mengembalikan hasil mentah
	CustomParams       map[string]interface{} `json:"custom_params"`       // Parameter khusus untuk engine keputusan
	Preprocessing      PreprocessingConfig    `json:"preprocessing"`       // Konfigurasi pra-pemrosesan
	Postprocessing     PostprocessingConfig   `json:"postprocessing"`      // Konfigurasi pasca-pemrosesan
	DecisionWeights    map[string]float64     `json:"decision_weights"`    // Bobot untuk berbagai faktor keputusan
}

// DecisionRule mewakili aturan keputusan
type DecisionRule struct {
	ID          string      `json:"id"`           // ID aturan
	Name        string      `json:"name"`         // Nama aturan
	Description string      `json:"description"`  // Deskripsi aturan
	Conditions  []Condition `json:"conditions"`   // Kondisi untuk aturan
	Action      string      `json:"action"`       // Aksi untuk diambil jika kondisi terpenuhi
	Priority    int         `json:"priority"`     // Prioritas aturan
	Enabled     bool        `json:"enabled"`      // Apakah aturan diaktifkan
}

// Condition mewakili kondisi dalam aturan keputusan
type Condition struct {
	Field     string      `json:"field"`      // Bidang untuk dicek
	Operator  string      `json:"operator"`   // Operator (>, <, =, !=, contains, etc.)
	Value     interface{} `json:"value"`      // Nilai untuk dibandingkan
	Threshold float64     `json:"threshold"`  // Ambang batas numerik
}

// DecisionResult mewakili hasil keputusan
type DecisionResult struct {
	Decision     string      `json:"decision"`      // Keputusan yang diambil
	Confidence   float64     `json:"confidence"`    // Tingkat kepercayaan
	Explanation  string      `json:"explanation"`   // Penjelasan keputusan
	RuleMatched  string      `json:"rule_matched"`  // Aturan yang cocok
	Timestamp    int64       `json:"timestamp"`     // Waktu keputusan dibuat
	ProcessingTime float64   `json:"processing_time"` // Waktu pemrosesan
	FeaturesUsed []string    `json:"features_used"` // Fitur yang digunakan
	TraceID      string      `json:"trace_id"`      // ID pelacakan
}

// AdvancedDecisionEngineNode mewakili node yang membuat keputusan canggih
type AdvancedDecisionEngineNode struct {
	config *AdvancedDecisionEngineConfig
}

// NewAdvancedDecisionEngineNode membuat node Advanced Decision Engine baru
func NewAdvancedDecisionEngineNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	// Konversi map interface{} ke JSON lalu ke struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("gagal mengubah konfig menjadi json: %v", err)
	}

	var decisionConfig AdvancedDecisionEngineConfig
	err = json.Unmarshal(jsonData, &decisionConfig)
	if err != nil {
		return nil, fmt.Errorf("gagal menguraikan konfig: %v", err)
	}

	// Validasi dan atur default
	if decisionConfig.DecisionType == "" {
		decisionConfig.DecisionType = "classification"
	}

	if decisionConfig.FallbackStrategy == "" {
		decisionConfig.FallbackStrategy = "rule_based"
	}

	if decisionConfig.ConfidenceThreshold == 0 {
		decisionConfig.ConfidenceThreshold = 0.7
	}

	if decisionConfig.Timeout == 0 {
		decisionConfig.Timeout = 60 // default timeout 60 detik
	}

	// Jika tidak ada aturan dan ML tidak diaktifkan, buat aturan default
	if len(decisionConfig.Rules) == 0 && !decisionConfig.EnableMachineLearning {
		decisionConfig.Rules = []DecisionRule{
			{
				ID:          "default_approval",
				Name:        "Default Approval Rule",
				Description: "Default rule for approval decisions",
				Conditions: []Condition{
					{
						Field:    "amount",
						Operator: "<",
						Value:    1000.0,
					},
				},
				Action:   "approve",
				Priority: 1,
				Enabled:  true,
			},
		}
	}

	return &AdvancedDecisionEngineNode{
		config: &decisionConfig,
	}, nil
}

// Execute mengimplementasikan interface NodeInstance
func (d *AdvancedDecisionEngineNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Timpa konfigurasi dengan nilai input jika disediakan
	provider := d.config.Provider
	if inputProvider, ok := input["provider"].(string); ok && inputProvider != "" {
		provider = inputProvider
	}

	decisionType := d.config.DecisionType
	if inputDecisionType, ok := input["decision_type"].(string); ok && inputDecisionType != "" {
		decisionType = inputDecisionType
	}

	modelName := d.config.ModelName
	if inputModelName, ok := input["model_name"].(string); ok && inputModelName != "" {
		modelName = inputModelName
	}

	apiKey := d.config.APIKey
	if inputAPIKey, ok := input["api_key"].(string); ok && inputAPIKey != "" {
		apiKey = inputAPIKey
	}

	enableML := d.config.EnableMachineLearning
	if inputEnableML, ok := input["enable_machine_learning"].(bool); ok {
		enableML = inputEnableML
	}

	mlModel := d.config.MLModel
	if inputMLModel, ok := input["ml_model"].(string); ok && inputMLModel != "" {
		mlModel = inputMLModel
	}

	confidenceThreshold := d.config.ConfidenceThreshold
	if inputConfThreshold, ok := input["confidence_threshold"].(float64); ok {
		confidenceThreshold = inputConfThreshold
	}

	fallbackStrategy := d.config.FallbackStrategy
	if inputFallback, ok := input["fallback_strategy"].(string); ok && inputFallback != "" {
		fallbackStrategy = inputFallback
	}

	enableExplainability := d.config.EnableExplainability
	if inputEnableExplain, ok := input["enable_explainability"].(bool); ok {
		enableExplainability = inputEnableExplain
	}

	traceID := d.config.TraceID
	if inputTraceID, ok := input["trace_id"].(string); ok && inputTraceID != "" {
		traceID = inputTraceID
	}

	enableCaching := d.config.EnableCaching
	if inputEnableCaching, ok := input["enable_caching"].(bool); ok {
		enableCaching = inputEnableCaching
	}

	cacheTTL := d.config.CacheTTL
	if inputCacheTTL, ok := input["cache_ttl"].(float64); ok {
		cacheTTL = int(inputCacheTTL)
	}

	enableProfiling := d.config.EnableProfiling
	if inputEnableProfiling, ok := input["enable_profiling"].(bool); ok {
		enableProfiling = inputEnableProfiling
	}

	timeout := d.config.Timeout
	if inputTimeout, ok := input["timeout"].(float64); ok {
		timeout = int(inputTimeout)
	}

	returnRawResults := d.config.ReturnRawResults
	if inputReturnRaw, ok := input["return_raw_results"].(bool); ok {
		returnRawResults = inputReturnRaw
	}

	customParams := d.config.CustomParams
	if inputCustomParams, ok := input["custom_params"].(map[string]interface{}); ok {
		customParams = inputCustomParams
	}

	// Validasi input
	if len(input) == 0 {
		return map[string]interface{}{
			"success":   false,
			"error":     "input diperlukan untuk membuat keputusan",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	if apiKey == "" && !enableML {
		return map[string]interface{}{
			"success":   false,
			"error":     "api_key diperlukan untuk engine keputusan",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Buat konteks keputusan dengan timeout
	decisionCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Buat ID pelacakan jika belum ada
	if traceID == "" {
		traceID = fmt.Sprintf("decision_%d", time.Now().UnixNano())
	}

	// Buat keputusan
	decisionResult, err := d.makeDecision(decisionCtx, input, traceID)
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
		"decision_type":        decisionType,
		"model_name":           modelName,
		"enable_machine_learning": enableML,
		"ml_model":             mlModel,
		"confidence_threshold": confidenceThreshold,
		"fallback_strategy":    fallbackStrategy,
		"enable_explainability": enableExplainability,
		"decision_result":      decisionResult,
		"enable_caching":       enableCaching,
		"enable_profiling":     enableProfiling,
		"return_raw_results":   returnRawResults,
		"trace_id":             traceID,
		"timestamp":            time.Now().Unix(),
		"input_data":           input,
		"config":               d.config,
	}

	// Tambahkan penjelasan jika diaktifkan
	if enableExplainability {
		finalResult["explanation"] = decisionResult.Explanation
		finalResult["rule_explanation"] = d.generateRuleExplanation(decisionResult.RuleMatched, input)
	}

	// Tambahkan metrik kinerja jika profiling diaktifkan
	if enableProfiling {
		finalResult["performance_metrics"] = map[string]interface{}{
			"start_time": time.Now().Unix(),
			"end_time":   time.Now().Unix(),
			"duration":   decisionResult.ProcessingTime,
		}
	}

	return finalResult, nil
}

// makeDecision membuat keputusan berdasarkan input dan konfigurasi
func (d *AdvancedDecisionEngineNode) makeDecision(ctx context.Context, input map[string]interface{}, traceID string) (*DecisionResult, error) {
	// Simulasikan waktu pemrosesan
	startTime := time.Now()
	time.Sleep(80 * time.Millisecond)

	var decision string
	var confidence float64
	var explanation string
	var ruleMatched string
	var featuresUsed []string

	// Jika ML diaktifkan, gunakan model ML, jika tidak gunakan aturan
	if d.config.EnableMachineLearning {
		decision, confidence, explanation = d.makeMLBasedDecision(input)
		ruleMatched = "ml_model"
	} else {
		// Cocokkan aturan untuk membuat keputusan
		matchedRule, decisionConfidence := d.matchRules(input)
		if matchedRule != nil {
			decision = matchedRule.Action
			confidence = decisionConfidence
			explanation = fmt.Sprintf("Aturan '%s' cocok dengan input", matchedRule.Name)
			ruleMatched = matchedRule.ID
		} else {
			// Gunakan strategi fallback jika tidak ada aturan yang cocok
			decision, confidence, explanation = d.applyFallbackStrategy(input)
			ruleMatched = "fallback_" + d.config.FallbackStrategy
		}
	}

	// Ambil nama-nama fitur yang digunakan dari input
	for key := range input {
		featuresUsed = append(featuresUsed, key)
	}

	result := &DecisionResult{
		Decision:       decision,
		Confidence:     confidence,
		Explanation:    explanation,
		RuleMatched:    ruleMatched,
		Timestamp:      time.Now().Unix(),
		ProcessingTime: time.Since(startTime).Seconds(),
		FeaturesUsed:   featuresUsed,
		TraceID:        traceID,
	}

	return result, nil
}

// makeMLBasedDecision membuat keputusan berdasarkan model ML
func (d *AdvancedDecisionEngineNode) makeMLBasedDecision(input map[string]interface{}) (string, float64, string) {
	// Ini adalah simulasi, dalam implementasi nyata akan memanggil model ML
	decision := "approve"
	confidence := 0.85
	explanation := "Keputusan dibuat berdasarkan model ML dengan fitur-fitur yang dianalisis"

	// Buat keputusan berdasarkan input
	if amount, exists := input["amount"]; exists {
		if amountFloat, ok := amount.(float64); ok {
			if amountFloat > 10000 {
				decision = "reject"
				confidence = 0.9
				explanation = "Jumlah melebihi ambang batas otomatis"
			} else if amountFloat > 5000 {
				decision = "review"
				confidence = 0.75
				explanation = "Jumlah tinggi memerlukan tinjauan manual"
			}
		}
	}

	return decision, confidence, explanation
}

// matchRules mencocokkan input dengan aturan-aturan yang ada
func (d *AdvancedDecisionEngineNode) matchRules(input map[string]interface{}) (*DecisionRule, float64) {
	// Urutkan aturan berdasarkan prioritas (tertinggi dulu)
	sortedRules := make([]DecisionRule, len(d.config.Rules))
	copy(sortedRules, d.config.Rules)
	
	// Urutkan dengan bubble sort sederhana
	for i := 0; i < len(sortedRules)-1; i++ {
		for j := 0; j < len(sortedRules)-i-1; j++ {
			if sortedRules[j].Priority < sortedRules[j+1].Priority {
				sortedRules[j], sortedRules[j+1] = sortedRules[j+1], sortedRules[j]
			}
		}
	}

	// Cek setiap aturan
	for _, rule := range sortedRules {
		if !rule.Enabled {
			continue
		}

		if d.evaluateRuleConditions(&rule, input) {
			confidence := 0.8 + (float64(rule.Priority) * 0.05) // Beri kepercayaan berdasarkan prioritas
			if confidence > 1.0 {
				confidence = 1.0
			}
			return &rule, confidence
		}
	}

	return nil, 0.0
}

// evaluateRuleConditions mengevaluasi apakah kondisi-kondisi dalam aturan cocok dengan input
func (d *AdvancedDecisionEngineNode) evaluateRuleConditions(rule *DecisionRule, input map[string]interface{}) bool {
	for _, condition := range rule.Conditions {
		inputValue, exists := input[condition.Field]
		if !exists {
			return false
		}

		if !d.evaluateCondition(inputValue, &condition) {
			return false
		}
	}

	return true
}

// evaluateCondition mengevaluasi kondisi individual
func (d *AdvancedDecisionEngineNode) evaluateCondition(inputValue interface{}, condition *Condition) bool {
	// Konversi input dan nilai kondisi ke bentuk yang dapat dibandingkan
	inputStr := fmt.Sprintf("%v", inputValue)
	conditionStr := fmt.Sprintf("%v", condition.Value)

	switch condition.Operator {
	case "=", "==", "equals":
		return inputStr == conditionStr
	case "!=", "not_equals":
		return inputStr != conditionStr
	case ">", "greater_than":
		if inputFloat, inputOk := inputValue.(float64); inputOk {
			if condFloat, condOk := condition.Value.(float64); condOk {
				return inputFloat > condFloat
			}
		}
	case "<", "less_than":
		if inputFloat, inputOk := inputValue.(float64); inputOk {
			if condFloat, condOk := condition.Value.(float64); condOk {
				return inputFloat < condFloat
			}
		}
	case ">=", "greater_than_or_equal":
		if inputFloat, inputOk := inputValue.(float64); inputOk {
			if condFloat, condOk := condition.Value.(float64); condOk {
				return inputFloat >= condFloat
			}
		}
	case "<=", "less_than_or_equal":
		if inputFloat, inputOk := inputValue.(float64); inputOk {
			if condFloat, condOk := condition.Value.(float64); condOk {
				return inputFloat <= condFloat
			}
		}
	case "contains":
		return strings.Contains(inputStr, conditionStr)
	case "starts_with":
		return len(inputStr) >= len(conditionStr) && inputStr[:len(conditionStr)] == conditionStr
	case "ends_with":
		return len(inputStr) >= len(conditionStr) && inputStr[len(inputStr)-len(conditionStr):] == conditionStr
	}

	return false
}

// applyFallbackStrategy menerapkan strategi fallback ketika tidak ada aturan yang cocok
func (d *AdvancedDecisionEngineNode) applyFallbackStrategy(input map[string]interface{}) (string, float64, string) {
	switch d.config.FallbackStrategy {
	case "manual":
		return "manual_review", 0.5, "Tidak ada aturan yang cocok, perlu peninjauan manual"
	case "default_action":
		return "approve", 0.6, "Menggunakan aksi default karena tidak ada aturan yang cocok"
	default: // rule_based
		// Ini adalah fallback default
		return "approve", 0.7, "Menggunakan pendekatan berbasis aturan karena tidak ada aturan spesifik yang cocok"
	}
}

// generateRuleExplanation menghasilkan penjelasan untuk keputusan berdasarkan aturan
func (d *AdvancedDecisionEngineNode) generateRuleExplanation(ruleID string, input map[string]interface{}) string {
	for _, rule := range d.config.Rules {
		if rule.ID == ruleID {
			return fmt.Sprintf("Keputusan '%s' diambil berdasarkan aturan '%s': %s. Input '%v' memenuhi kondisi.", 
				rule.Action, rule.Name, rule.Description, input)
		}
	}
	return "Penjelasan tidak tersedia untuk aturan ini"
}

// GetType mengembalikan jenis node
func (d *AdvancedDecisionEngineNode) GetType() string {
	return "advanced_decision_engine"
}

// GetID mengembalikan ID unik untuk instance node
func (d *AdvancedDecisionEngineNode) GetID() string {
	return "adv_decision_" + fmt.Sprintf("%d", time.Now().Unix())
}

// RegisterAdvancedDecisionEngineNode mendaftarkan node Advanced Decision Engine dengan engine
func RegisterAdvancedDecisionEngineNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("advanced_decision_engine", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return NewAdvancedDecisionEngineNode(config)
	})
}

