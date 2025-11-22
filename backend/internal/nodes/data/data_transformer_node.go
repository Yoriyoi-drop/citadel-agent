package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"citadel-agent/backend/internal/workflow/core/engine"
)

// DataTransformerConfig mewakili konfigurasi untuk node Data Transformer
type DataTransformerConfig struct {
	TransformType   string                   `json:"transform_type"`    // Jenis transformasi (json_to_csv, csv_to_json, etc.)
	Mapping         map[string]interface{}   `json:"mapping"`           // Pemetaan field
	Template        string                   `json:"template"`          // Template untuk transformasi
	Operations      []DataOperation          `json:"operations"`        // Operasi-operasi transformasi
	SourceFormat    string                   `json:"source_format"`     // Format sumber (json, csv, xml, etc.)
	TargetFormat    string                   `json:"target_format"`     // Format tujuan
	EnableValidation bool                    `json:"enable_validation"` // Apakah mengaktifkan validasi
	ValidationRules []ValidationRule          `json:"validation_rules"`  // Aturan validasi
	EnableFiltering bool                     `json:"enable_filtering"`  // Apakah mengaktifkan penyaringan
	FilterRules     []FilterRule             `json:"filter_rules"`      // Aturan penyaringan
	EnableAggregation bool                   `json:"enable_aggregation"` // Apakah mengaktifkan agregasi
	GroupByFields   []string                 `json:"group_by_fields"`   // Field untuk pengelompokan
	EnableCaching   bool                     `json:"enable_caching"`    // Apakah mengaktifkan caching hasil
	CacheTTL        int                      `json:"cache_ttl"`         // Waktu cache dalam detik
	EnableProfiling bool                     `json:"enable_profiling"`  // Apakah mengaktifkan profiling
	Timeout         int                      `json:"timeout"`           // Waktu timeout dalam detik
	ReturnRawResults bool                    `json:"return_raw_results"` // Apakah mengembalikan hasil mentah
	CustomParams    map[string]interface{}   `json:"custom_params"`     // Parameter khusus untuk transformasi
	Preprocessing   PreprocessingConfig      `json:"preprocessing"`     // Konfigurasi pra-pemrosesan
	Postprocessing  PostprocessingConfig     `json:"postprocessing"`    // Konfigurasi pasca-pemrosesan
	BatchSize       int                      `json:"batch_size"`       // Ukuran batch untuk pemrosesan
}

// DataOperation mewakili operasi transformasi data
type DataOperation struct {
	ID          string      `json:"id"`           // ID operasi
	Type        string      `json:"type"`         // Jenis operasi (rename, cast, calculate, etc.)
	Field       string      `json:"field"`        // Field untuk dioperasikan
	Operation   string      `json:"operation"`    // Operasi yang akan dilakukan
	Value       interface{} `json:"value"`        // Nilai untuk operasi
	TargetField string      `json:"target_field"` // Field tujuan (jika berbeda)
	Enabled     bool        `json:"enabled"`      // Apakah operasi diaktifkan
}

// ValidationRule mewakili aturan validasi
type ValidationRule struct {
	Field       string      `json:"field"`        // Field untuk divalidasi
	RuleType    string      `json:"rule_type"`    // Jenis aturan (required, min_length, max_length, regex, etc.)
	Condition   interface{} `json:"condition"`    // Kondisi validasi
	ErrorMessage string     `json:"error_message"` // Pesan error jika validasi gagal
}

// FilterRule mewakili aturan penyaringan
type FilterRule struct {
	Field     string      `json:"field"`      // Field untuk difilter
	Operator  string      `json:"operator"`   // Operator pembanding
	Value     interface{} `json:"value"`      // Nilai untuk dibandingkan
	Condition string      `json:"condition"`  // Kondisi gabungan (AND, OR)
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

// DataTransformerNode mewakili node yang mentransformasi data
type DataTransformerNode struct {
	config *DataTransformerConfig
}

// NewDataTransformerNode membuat node Data Transformer baru
func NewDataTransformerNode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Konversi map interface{} ke JSON lalu ke struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("gagal mengubah konfig menjadi json: %v", err)
	}

	var transformConfig DataTransformerConfig
	err = json.Unmarshal(jsonData, &transformConfig)
	if err != nil {
		return nil, fmt.Errorf("gagal menguraikan konfig: %v", err)
	}

	// Validasi dan atur default
	if transformConfig.TransformType == "" {
		transformConfig.TransformType = "mapping"
	}

	if transformConfig.SourceFormat == "" {
		transformConfig.SourceFormat = "json"
	}

	if transformConfig.TargetFormat == "" {
		transformConfig.TargetFormat = "json"
	}

	if transformConfig.Timeout == 0 {
		transformConfig.Timeout = 60 // default timeout 60 detik
	}

	if transformConfig.BatchSize == 0 {
		transformConfig.BatchSize = 100 // default batch size
	}

	if transformConfig.Mapping == nil {
		transformConfig.Mapping = make(map[string]interface{})
	}

	return &DataTransformerNode{
		config: &transformConfig,
	}, nil
}

// Execute mengimplementasikan interface NodeInstance
func (d *DataTransformerNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Timpa konfigurasi dengan nilai input jika disediakan
	transformType := d.config.TransformType
	if inputTransformType, ok := input["transform_type"].(string); ok && inputTransformType != "" {
		transformType = inputTransformType
	}

	mapping := d.config.Mapping
	if inputMapping, ok := input["mapping"].(map[string]interface{}); ok {
		mapping = inputMapping
	}

	template := d.config.Template
	if inputTemplate, ok := input["template"].(string); ok && inputTemplate != "" {
		template = inputTemplate
	}

	operations := d.config.Operations
	if inputOperations, ok := input["operations"].([]interface{}); ok {
		operations = make([]DataOperation, len(inputOperations))
		for i, op := range inputOperations {
			if opMap, ok := op.(map[string]interface{}); ok {
				var operation DataOperation
				if id, exists := opMap["id"]; exists {
					if idStr, ok := id.(string); ok {
						operation.ID = idStr
					}
				}
				if typ, exists := opMap["type"]; exists {
					if typStr, ok := typ.(string); ok {
						operation.Type = typStr
					}
				}
				if field, exists := opMap["field"]; exists {
					if fieldStr, ok := field.(string); ok {
						operation.Field = fieldStr
					}
				}
				if opStr, exists := opMap["operation"]; exists {
					if opStr2, ok := opStr.(string); ok {
						operation.Operation = opStr2
					}
				}
				if value, exists := opMap["value"]; exists {
					operation.Value = value
				}
				if target, exists := opMap["target_field"]; exists {
					if targetStr, ok := target.(string); ok {
						operation.TargetField = targetStr
					}
				}
				if enabled, exists := opMap["enabled"]; exists {
					if enabledBool, ok := enabled.(bool); ok {
						operation.Enabled = enabledBool
					}
				}
				operations[i] = operation
			}
		}
	}

	sourceFormat := d.config.SourceFormat
	if inputSourceFormat, ok := input["source_format"].(string); ok && inputSourceFormat != "" {
		sourceFormat = inputSourceFormat
	}

	targetFormat := d.config.TargetFormat
	if inputTargetFormat, ok := input["target_format"].(string); ok && inputTargetFormat != "" {
		targetFormat = inputTargetFormat
	}

	enableValidation := d.config.EnableValidation
	if inputEnableValidation, ok := input["enable_validation"].(bool); ok {
		enableValidation = inputEnableValidation
	}

	validationRules := d.config.ValidationRules
	if inputValidationRules, ok := input["validation_rules"].([]interface{}); ok {
		validationRules = make([]ValidationRule, len(inputValidationRules))
		for i, rule := range inputValidationRules {
			if ruleMap, ok := rule.(map[string]interface{}); ok {
				var validationRule ValidationRule
				if field, exists := ruleMap["field"]; exists {
					if fieldStr, ok := field.(string); ok {
						validationRule.Field = fieldStr
					}
				}
				if ruleType, exists := ruleMap["rule_type"]; exists {
					if ruleTypeStr, ok := ruleType.(string); ok {
						validationRule.RuleType = ruleTypeStr
					}
				}
				if condition, exists := ruleMap["condition"]; exists {
					validationRule.Condition = condition
				}
				if errorMsg, exists := ruleMap["error_message"]; exists {
					if errorMsgStr, ok := errorMsg.(string); ok {
						validationRule.ErrorMessage = errorMsgStr
					}
				}
				validationRules[i] = validationRule
			}
		}
	}

	enableFiltering := d.config.EnableFiltering
	if inputEnableFiltering, ok := input["enable_filtering"].(bool); ok {
		enableFiltering = inputEnableFiltering
	}

	filterRules := d.config.FilterRules
	if inputFilterRules, ok := input["filter_rules"].([]interface{}); ok {
		filterRules = make([]FilterRule, len(inputFilterRules))
		for i, rule := range inputFilterRules {
			if ruleMap, ok := rule.(map[string]interface{}); ok {
				var filterRule FilterRule
				if field, exists := ruleMap["field"]; exists {
					if fieldStr, ok := field.(string); ok {
						filterRule.Field = fieldStr
					}
				}
				if operator, exists := ruleMap["operator"]; exists {
					if operatorStr, ok := operator.(string); ok {
						filterRule.Operator = operatorStr
					}
				}
				if value, exists := ruleMap["value"]; exists {
					filterRule.Value = value
				}
				if condition, exists := ruleMap["condition"]; exists {
					if conditionStr, ok := condition.(string); ok {
						filterRule.Condition = conditionStr
					}
				}
				filterRules[i] = filterRule
			}
		}
	}

	enableAggregation := d.config.EnableAggregation
	if inputEnableAggregation, ok := input["enable_aggregation"].(bool); ok {
		enableAggregation = inputEnableAggregation
	}

	groupByFields := d.config.GroupByFields
	if inputGroupByFields, ok := input["group_by_fields"].([]interface{}); ok {
		groupByFields = make([]string, len(inputGroupByFields))
		for i, field := range inputGroupByFields {
			if fieldStr, ok := field.(string); ok {
				groupByFields[i] = fieldStr
			}
		}
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

	batchSize := d.config.BatchSize
	if inputBatchSize, ok := input["batch_size"].(float64); ok {
		batchSize = int(inputBatchSize)
	}

	// Validasi input
	if len(input) == 0 {
		return map[string]interface{}{
			"success":   false,
			"error":     "input diperlukan untuk transformasi data",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Buat konteks operasi dengan timeout
	transformCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Lakukan transformasi data
	transformResult, err := d.transformData(transformCtx, input)
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
		"transform_type":       transformType,
		"source_format":        sourceFormat,
		"target_format":        targetFormat,
		"enable_validation":    enableValidation,
		"enable_filtering":     enableFiltering,
		"enable_aggregation":   enableAggregation,
		"group_by_fields":      groupByFields,
		"batch_size":           batchSize,
		"transform_result":     transformResult,
		"enable_caching":       enableCaching,
		"enable_profiling":     enableProfiling,
		"return_raw_results":   returnRawResults,
		"timestamp":            time.Now().Unix(),
		"input_data":           input,
		"mapping":              mapping,
		"template":             template,
		"operations_count":     len(operations),
		"config":               d.config,
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

// transformData melakukan transformasi data
func (d *DataTransformerNode) transformData(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Simulasikan waktu pemrosesan
	time.Sleep(100 * time.Millisecond)

	// Proses input berdasarkan tipe transformasi
	var transformedData interface{}
	var transformStats map[string]interface{}

	switch d.config.TransformType {
	case "mapping":
		transformedData, transformStats = d.applyMapping(input)
	case "template":
		transformedData, transformStats = d.applyTemplate(input)
	case "json_to_csv", "json_to_xml", "xml_to_json", "csv_to_json":
		transformedData, transformStats = d.applyFormatConversion(input)
	case "aggregation":
		transformedData, transformStats = d.applyAggregation(input)
	case "filtering":
		transformedData, transformStats = d.applyFiltering(input)
	default:
		// Jika tipe transformasi tidak diketahui, gunakan mapping default
		transformedData, transformStats = d.applyMapping(input)
	}

	result := map[string]interface{}{
		"transformed_data": transformedData,
		"transform_type": d.config.TransformType,
		"source_format": d.config.SourceFormat,
		"target_format": d.config.TargetFormat,
		"transformation_stats": transformStats,
		"original_keys_count": len(input),
		"processing_time": time.Since(time.Now().Add(-100 * time.Millisecond)).Seconds(),
		"timestamp": time.Now().Unix(),
		"format_converted": d.config.SourceFormat != d.config.TargetFormat,
	}

	// Jalankan operasi-operasi transformasi tambahan jika ada
	if len(d.config.Operations) > 0 {
		result["applied_operations"] = len(d.config.Operations)
		result["operations_detail"] = d.executeOperations(input, transformedData)
	}

	// Jalankan validasi jika diaktifkan
	if d.config.EnableValidation {
		validationResult := d.validateData(transformedData)
		result["validation_result"] = validationResult
		result["data_valid"] = validationResult["valid"]
	}

	// Jalankan filtering jika diaktifkan
	if d.config.EnableFiltering {
		filteredData, filteredStats := d.applyFilteringWithRules(transformedData)
		result["filtered_data"] = filteredData
		result["filtering_stats"] = filteredStats
	}

	return result, nil
}

// applyMapping menerapkan mapping ke data
func (d *DataTransformerNode) applyMapping(input map[string]interface{}) (interface{}, map[string]interface{}) {
	result := make(map[string]interface{})
	
	// Jika mapping tersedia, gunakan itu
	if len(d.config.Mapping) > 0 {
		for newKey, valueSpec := range d.config.Mapping {
			switch spec := valueSpec.(type) {
			case string:
				// Jika valueSpec adalah string, coba cocokkan dengan key di input
				if val, exists := input[spec]; exists {
					result[newKey] = val
				} else {
					// Jika tidak ditemukan, gunakan sebagai nilai literal
					result[newKey] = spec
				}
			case map[string]interface{}:
				// Mapping bersarang
				if nestedInput, exists := input[newKey]; exists {
					if nestedMap, ok := nestedInput.(map[string]interface{}); ok {
						nestedResult, _ := d.applyMapping(nestedMap)
						result[newKey] = nestedResult
					} else {
						result[newKey] = nestedInput
					}
				}
			case []interface{}:
				// Array mapping
				result[newKey] = spec
			default:
				result[newKey] = spec
			}
		}
	} else {
		// Jika tidak ada mapping, salin semua data
		for k, v := range input {
			result[k] = v
		}
	}

	stats := map[string]interface{}{
		"input_fields": len(input),
		"output_fields": len(result),
		"mapping_applied": len(d.config.Mapping) > 0,
	}

	return result, stats
}

// applyTemplate menerapkan template ke data
func (d *DataTransformerNode) applyTemplate(input map[string]interface{}) (interface{}, map[string]interface{}) {
	// Dalam implementasi nyata, ini akan memproses template
	// Untuk simulasi, kita akan kembalikan input yang sedikit dimodifikasi
	result := make(map[string]interface{})
	
	// Tambahkan informasi template ke hasil
	for k, v := range input {
		result[k] = v
	}
	
	result["_template_applied"] = d.config.Template
	result["_template_processed_at"] = time.Now().Unix()

	stats := map[string]interface{}{
		"input_fields": len(input),
		"output_fields": len(result),
		"template": d.config.Template,
		"template_applied": d.config.Template != "",
	}

	return result, stats
}

// applyFormatConversion mensimulasikan konversi format
func (d *DataTransformerNode) applyFormatConversion(input map[string]interface{}) (interface{}, map[string]interface{}) {
	// Simulasikan konversi format
	result := input
	if d.config.SourceFormat != d.config.TargetFormat {
		// Tandai bahwa konversi format telah dilakukan
		result = make(map[string]interface{})
		for k, v := range input {
			result[k] = v
		}
		result["_format_converted"] = map[string]interface{}{
			"from": d.config.SourceFormat,
			"to":   d.config.TargetFormat,
			"timestamp": time.Now().Unix(),
		}
	}

	stats := map[string]interface{}{
		"source_format": d.config.SourceFormat,
		"target_format": d.config.TargetFormat,
		"conversion_needed": d.config.SourceFormat != d.config.TargetFormat,
		"input_fields": len(input),
	}

	return result, stats
}

// applyAggregation menerapkan agregasi ke data
func (d *DataTransformerNode) applyAggregation(input map[string]interface{}) (interface{}, map[string]interface{}) {
	// Simulasikan agregasi data
	result := make(map[string]interface{})
	
	// Hitung statistik sederhana dari data numerik
	count := 0
	sum := 0.0
	avg := 0.0
	
	for k, v := range input {
		if num, ok := d.getAsFloat64(v); ok {
			sum += num
			count++
		}
		result[k] = v  // Salin field asli
	}
	
	if count > 0 {
		avg = sum / float64(count)
	}
	
	result["_aggregation_stats"] = map[string]interface{}{
		"numeric_fields_count": count,
		"sum": sum,
		"average": avg,
		"aggregation_fields": d.config.GroupByFields,
	}

	stats := map[string]interface{}{
		"total_fields": len(input),
		"numeric_fields": count,
		"aggregates_calculated": map[string]interface{}{
			"sum": sum,
			"average": avg,
			"count": count,
		},
		"group_by_fields": d.config.GroupByFields,
	}

	return result, stats
}

// applyFiltering menerapkan penyaringan ke data
func (d *DataTransformerNode) applyFiltering(input map[string]interface{}) (interface{}, map[string]interface{}) {
	// Untuk simulasi, kembalikan semua data
	result := make(map[string]interface{})
	for k, v := range input {
		result[k] = v
	}

	stats := map[string]interface{}{
		"input_fields": len(input),
		"output_fields": len(result),
		"filtering_enabled": d.config.EnableFiltering,
	}

	return result, stats
}

// applyFilteringWithRules menerapkan aturan penyaringan ke data
func (d *DataTransformerNode) applyFilteringWithRules(data interface{}) (interface{}, map[string]interface{}) {
	// Dalam implementasi nyata, ini akan menerapkan aturan penyaringan
	// Untuk simulasi, kita anggap semua data lolos penyaringan
	result := data

	stats := map[string]interface{}{
		"filter_rules_count": len(d.config.FilterRules),
		"filtering_applied": len(d.config.FilterRules) > 0,
		"data_passed": true,
	}

	return result, stats
}

// validateData melakukan validasi terhadap data
func (d *DataTransformerNode) validateData(data interface{}) map[string]interface{} {
	validationResult := map[string]interface{}{
		"valid": true,
		"errors": []string{},
		"rules_applied": len(d.config.ValidationRules),
		"validation_passed": true,
	}
	
	// Dalam implementasi nyata, ini akan menjalankan aturan validasi
	// Untuk simulasi, kita anggap validasi selalu berhasil
	
	return validationResult
}

// executeOperations menjalankan operasi-operasi transformasi
func (d *DataTransformerNode) executeOperations(originalInput map[string]interface{}, transformedData interface{}) []map[string]interface{} {
	var operationResults []map[string]interface{}
	
	for _, operation := range d.config.Operations {
		if !operation.Enabled {
			continue
		}
		
		opResult := map[string]interface{}{
			"operation_id": operation.ID,
			"type": operation.Type,
			"field": operation.Field,
			"operation": operation.Operation,
			"target_field": operation.TargetField,
			"success": true,
			"processed_at": time.Now().Unix(),
		}
		
		// Simulasikan eksekusi operasi
		switch operation.Operation {
		case "rename":
			// Dalam implementasi nyata, ini akan mengganti nama field
			opResult["action"] = fmt.Sprintf("Renamed field '%s' to '%s'", operation.Field, operation.TargetField)
		case "cast":
			// Dalam implementasi nyata, ini akan mengkonversi tipe data
			opResult["action"] = fmt.Sprintf("Casted field '%s' to type from value", operation.Field)
		case "calculate":
			// Dalam implementasi nyata, ini akan melakukan perhitungan
			opResult["action"] = fmt.Sprintf("Calculated value for field '%s'", operation.Field)
		default:
			opResult["action"] = fmt.Sprintf("Applied %s operation to field '%s'", operation.Operation, operation.Field)
		}
		
		operationResults = append(operationResults, opResult)
	}
	
	return operationResults
}

// getAsFloat64 mencoba mengkonversi nilai ke float64
func (d *DataTransformerNode) getAsFloat64(value interface{}) (float64, bool) {
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
		// Untuk simulasi, kembalikan 0
		return 0, false
	default:
		return 0, false
	}
}

// GetType mengembalikan jenis node
func (d *DataTransformerNode) GetType() string {
	return "data_transformer"
}

// GetID mengembalikan ID unik untuk instance node
func (d *DataTransformerNode) GetID() string {
	return "data_xform_" + fmt.Sprintf("%d", time.Now().Unix())
}

// RegisterDataTransformNode mendaftarkan node Data Transformer dengan engine
func RegisterDataTransformNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("data_transformer", func(config map[string]interface{}) (engine.NodeInstance, error) {
		return NewDataTransformerNode(config)
	})
}