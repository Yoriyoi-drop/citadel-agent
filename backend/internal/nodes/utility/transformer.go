package utility

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
)

// TransformerConfig represents the configuration for a data transformer node
type TransformerConfig struct {
	TransformType string                 `json:"transform_type"`  // "json_to_xml", "json_to_csv", "template", etc.
	Mapping       map[string]interface{} `json:"mapping"`         // Field mappings
	Template      string                 `json:"template"`        // Template string
	TransformRules []TransformRule       `json:"transform_rules"` // Transformation rules
	SourceFormat  string                 `json:"source_format"`   // "json", "xml", "csv", etc.
	TargetFormat  string                 `json:"target_format"`   // "json", "xml", "csv", etc.
	EnableCaching bool                   `json:"enable_caching"`
	CacheTTL      int                    `json:"cache_ttl"`       // in seconds
	EnableProfiling bool                 `json:"enable_profiling"`
	ReturnRawResults bool                 `json:"return_raw_results"`
	CustomParams  map[string]interface{} `json:"custom_params"`
}

// TransformRule represents a single transformation rule
type TransformRule struct {
	Field       string      `json:"field"`       // Field to transform
	Operation   string      `json:"operation"`   // Operation type (e.g., "uppercase", "lowercase", "trim")
	Value       interface{} `json:"value"`       // Value for operation (if needed)
	Condition   string      `json:"condition"`   // Condition for applying rule
	TargetField string      `json:"target_field"` // Target field (if different from source)
}

// TransformerNode represents a data transformer node
type TransformerNode struct {
	config *TransformerConfig
}

// NewTransformerNode creates a new transformer node
func NewTransformerNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	// Convert config map to struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var transConfig TransformerConfig
	if err := json.Unmarshal(jsonData, &transConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set defaults
	if transConfig.TransformType == "" {
		transConfig.TransformType = "mapping"
	}

	if transConfig.SourceFormat == "" {
		transConfig.SourceFormat = "json"
	}

	if transConfig.TargetFormat == "" {
		transConfig.TargetFormat = "json"
	}

	if transConfig.CacheTTL == 0 {
		transConfig.CacheTTL = 3600 // 1 hour default cache TTL
	}

	return &TransformerNode{
		config: &transConfig,
	}, nil
}

// Execute executes the data transformation
func (tn *TransformerNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	startTime := time.Now()

	// Override config values with inputs if provided
	transformType := tn.config.TransformType
	if inputType, exists := inputs["transform_type"]; exists {
		if typeStr, ok := inputType.(string); ok && typeStr != "" {
			transformType = typeStr
		}
	}

	mapping := tn.config.Mapping
	if inputMapping, exists := inputs["mapping"]; exists {
		if mappingMap, ok := inputMapping.(map[string]interface{}); ok {
			mapping = mappingMap
		}
	}

	template := tn.config.Template
	if inputTemplate, exists := inputs["template"]; exists {
		if tplStr, ok := inputTemplate.(string); ok {
			template = tplStr
		}
	}

	transformRules := tn.config.TransformRules
	if inputRules, exists := inputs["transform_rules"]; exists {
		if rulesSlice, ok := inputRules.([]interface{}); ok {
			transformRules = make([]TransformRule, len(rulesSlice))
			for i, rule := range rulesSlice {
				if ruleMap, ok := rule.(map[string]interface{}); ok {
					var rule TransformRule
					
					if field, exists := ruleMap["field"]; exists {
						if fieldStr, ok := field.(string); ok {
							rule.Field = fieldStr
						}
					}
					
					if operation, exists := ruleMap["operation"]; exists {
						if opStr, ok := operation.(string); ok {
							rule.Operation = opStr
						}
					}
					
					if value, exists := ruleMap["value"]; exists {
						rule.Value = value
					}
					
					if condition, exists := ruleMap["condition"]; exists {
						if condStr, ok := condition.(string); ok {
							rule.Condition = condStr
						}
					}
					
					if targetField, exists := ruleMap["target_field"]; exists {
						if tgtStr, ok := targetField.(string); ok {
							rule.TargetField = tgtStr
						}
					}
					
					transformRules[i] = rule
				}
			}
		}
	}

	sourceFormat := tn.config.SourceFormat
	if inputSourceFormat, exists := inputs["source_format"]; exists {
		if formatStr, ok := inputSourceFormat.(string); ok {
			sourceFormat = formatStr
		}
	}

	targetFormat := tn.config.TargetFormat
	if inputTargetFormat, exists := inputs["target_format"]; exists {
		if formatStr, ok := inputTargetFormat.(string); ok {
			targetFormat = formatStr
		}
	}

	enableProfiling := tn.config.EnableProfiling
	if inputEnableProfiling, exists := inputs["enable_profiling"]; exists {
		if prof, ok := inputEnableProfiling.(bool); ok {
			enableProfiling = prof
		}
	}

	returnRawResults := tn.config.ReturnRawResults
	if inputReturnRaw, exists := inputs["return_raw_results"]; exists {
		if raw, ok := inputReturnRaw.(bool); ok {
			returnRawResults = raw
		}
	}

	// Perform transformation based on type
	var result interface{}
	var err error

	switch transformType {
	case "mapping":
		result, err = tn.applyMapping(inputs, mapping)
	case "template":
		result, err = tn.applyTemplate(inputs, template)
	case "transform_rules":
		result, err = tn.applyTransformRules(inputs, transformRules)
	case "json_to_csv", "json_to_xml", "xml_to_json", "csv_to_json":
		result, err = tn.applyFormatConversion(inputs, sourceFormat, targetFormat)
	default:
		// Default to mapping if transform type is unknown
		result, err = tn.applyMapping(inputs, mapping)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to apply transformation: %w", err)
	}

	// Prepare output
	output := make(map[string]interface{})
	
	// Add transformed data
	output["transformed_data"] = result
	output["success"] = true
	output["transform_type"] = transformType
	output["source_format"] = sourceFormat
	output["target_format"] = targetFormat
	output["input_size"] = len(inputs)
	output["timestamp"] = time.Now().Unix()
	output["execution_time"] = time.Since(startTime).Seconds()
	
	// Add original inputs unless specifically told not to
	if !returnRawResults {
		output["input_data"] = inputs
	} else {
		output["raw_input"] = inputs
		output["raw_output"] = result
	}

	// Add profiling data if enabled
	if enableProfiling {
		output["profiling"] = map[string]interface{}{
			"start_time": startTime.Unix(),
			"end_time":   time.Now().Unix(),
			"duration":   time.Since(startTime).Seconds(),
			"operation":  transformType,
			"input_size": len(inputs),
			"transform_type": transformType,
		}
	}

	return output, nil
}

// applyMapping applies field mappings to transform data
func (tn *TransformerNode) applyMapping(input map[string]interface{}, mapping map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	if mapping == nil || len(mapping) == 0 {
		// If no mapping provided, return the input as is
		return input, nil
	}
	
	// Apply mappings
	for newKey, valueSpec := range mapping {
		switch spec := valueSpec.(type) {
		case string:
			// If valueSpec is a string, check if it refers to an input field
			if val, exists := input[spec]; exists {
				result[newKey] = val
			} else {
				// If not found in input, treat as literal value
				result[newKey] = spec
			}
		case map[string]interface{}:
			// For nested mapping
			if nestedInput, exists := input[newKey]; exists {
				if nestedMap, ok := nestedInput.(map[string]interface{}); ok {
					nestedResult, err := tn.applyMapping(nestedMap, spec)
					if err != nil {
						return nil, err
					}
					result[newKey] = nestedResult
				} else {
					result[newKey] = nestedInput
				}
			} else {
				result[newKey] = spec
			}
		case []interface{}:
			// For array mapping
			result[newKey] = spec
		default:
			result[newKey] = spec
		}
	}
	
	// Add unmapped fields if not explicitly mapped
	for k, v := range input {
		if _, exists := result[k]; !exists {
			result[k] = v
		}
	}
	
	return result, nil
}

// applyTemplate applies a template to the input data
func (tn *TransformerNode) applyTemplate(input map[string]interface{}, template string) (interface{}, error) {
	// In a real implementation, this would use a templating engine like Go's text/template
	// For now, we'll do a basic string replacement
	
	result := template
	
	for k, v := range input {
		placeholder := "{{" + k + "}}"
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", v))
	}
	
	return result, nil
}

// applyTransformRules applies transformation rules to the input data
func (tn *TransformerNode) applyTransformRules(input map[string]interface{}, rules []TransformRule) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// Copy original input
	for k, v := range input {
		result[k] = v
	}
	
	// Apply each rule
	for _, rule := range rules {
		// Check if condition is met
		conditionMet := true
		if rule.Condition != "" {
			conditionMet = tn.evaluateCondition(rule.Condition, result)
		}
		
		if !conditionMet {
			continue // Skip rule if condition is not met
		}
		
		// Get source value
		sourceValue, exists := input[rule.Field]
		if !exists {
			continue // Skip if field doesn't exist
		}
		
		// Apply transformation
		transformedValue, err := tn.applyTransformation(sourceValue, rule)
		if err != nil {
			return nil, fmt.Errorf("failed to apply transformation: %w", err)
		}
		
		// Determine target field
		targetField := rule.TargetField
		if targetField == "" {
			targetField = rule.Field // Use same field if no target specified
		}
		
		result[targetField] = transformedValue
	}
	
	return result, nil
}

// applyTransformation applies a single transformation to a value
func (tn *TransformerNode) applyTransformation(value interface{}, rule TransformRule) (interface{}, error) {
	switch rule.Operation {
	case "uppercase":
		if str, ok := value.(string); ok {
			return strings.ToUpper(str), nil
		}
		return value, nil
	case "lowercase":
		if str, ok := value.(string); ok {
			return strings.ToLower(str), nil
		}
		return value, nil
	case "trim":
		if str, ok := value.(string); ok {
			return strings.TrimSpace(str), nil
		}
		return value, nil
	case "replace":
		if str, ok := value.(string); ok {
			fromStr := ""
			toStr := ""
			
			if rule.Value != nil {
				if valueMap, ok := rule.Value.(map[string]interface{}); ok {
					if from, exists := valueMap["from"]; exists {
						if fromStrVal, fromOk := from.(string); fromOk {
							fromStr = fromStrVal
						}
					}
					if to, exists := valueMap["to"]; exists {
						if toStrVal, toOk := to.(string); toOk {
							toStr = toStrVal
						}
					}
				}
			}
			
			return strings.ReplaceAll(str, fromStr, toStr), nil
		}
		return value, nil
	case "substring":
		if str, ok := value.(string); ok {
			start := 0
			end := len(str)
			
			if rule.Value != nil {
				if valueMap, ok := rule.Value.(map[string]interface{}); ok {
					if startVal, exists := valueMap["start"]; exists {
						if startFloat, startOk := startVal.(float64); startOk {
							start = int(startFloat)
						}
					}
					if endVal, exists := valueMap["end"]; exists {
						if endFloat, endOk := endVal.(float64); endOk {
							end = int(endFloat)
						}
					}
				}
			}
			
			if start < 0 {
				start = 0
			}
			if end > len(str) {
				end = len(str)
			}
			if start > end {
				start, end = end, start
			}
			
			return str[start:end], nil
		}
		return value, nil
	case "split":
		if str, ok := value.(string); ok {
			separator := ","
			
			if rule.Value != nil {
				if sepStr, ok := rule.Value.(string); ok {
					separator = sepStr
				}
			}
			
			return strings.Split(str, separator), nil
		}
		return value, nil
	case "parse_number":
		if str, ok := value.(string); ok {
			if num, err := strconv.ParseFloat(str, 64); err == nil {
				return num, nil
			}
		} else if num, ok := value.(float64); ok {
			return num, nil
		} else if num, ok := value.(int); ok {
			return float64(num), nil
		}
		return value, nil
	case "to_string":
		return fmt.Sprintf("%v", value), nil
	default:
		// For unknown operations, return the value unchanged
		return value, nil
	}
}

// evaluateCondition evaluates a simple condition against the input
func (tn *TransformerNode) evaluateCondition(condition string, input map[string]interface{}) bool {
	// For simplicity, we'll implement a basic condition evaluation
	// In a real implementation, this would be more sophisticated
	
	// Example: "field_name equals value" or "field_name exists"
	parts := strings.Fields(condition)
	if len(parts) < 3 {
		return false
	}
	
	fieldName := parts[0]
	operator := parts[1]
	expectedValue := strings.Join(parts[2:], " ")
	
	switch operator {
	case "equals":
		if actualValue, exists := input[fieldName]; exists {
			return fmt.Sprintf("%v", actualValue) == expectedValue
		}
		return false
	case "exists":
		_, exists := input[fieldName]
		return exists
	case "contains":
		if actualValue, exists := input[fieldName]; exists {
			return strings.Contains(fmt.Sprintf("%v", actualValue), expectedValue)
		}
		return false
	default:
		return false
	}
}

// applyFormatConversion performs format conversion
func (tn *TransformerNode) applyFormatConversion(input map[string]interface{}, sourceFormat, targetFormat string) (interface{}, error) {
	if sourceFormat == targetFormat {
		return input, nil
	}
	
	// For now, we'll just return the input as-is
	// In a real implementation, this would convert between formats
	return input, nil
}

// GetType returns the type of node
func (tn *TransformerNode) GetType() string {
	return "data_transformer"
}

// GetID returns the unique ID of the node instance
func (tn *TransformerNode) GetID() string {
	return fmt.Sprintf("transform_%s_%d", tn.config.TransformType, time.Now().Unix())
}