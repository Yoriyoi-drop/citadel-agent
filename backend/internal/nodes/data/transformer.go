// backend/internal/nodes/data/transformer.go
package data

import (
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/citadel-agent/backend/internal/workflow/core/engine"
)

// TransformOperationType represents the type of data transformation operation
type TransformOperationType string

const (
	TransformJSONPath     TransformOperationType = "json_path"
	TransformXPath       TransformOperationType = "xpath"  // Simplified for JSON
	TransformFilter      TransformOperationType = "filter"
	TransformMap         TransformOperationType = "map"
	TransformReduce      TransformOperationType = "reduce"
	TransformSort        TransformOperationType = "sort"
	TransformJoin        TransformOperationType = "join"
	TransformSplit       TransformOperationType = "split"
	TransformReplace     TransformOperationType = "replace"
	TransformCaseConvert TransformOperationType = "case_convert"
	TransformBase64      TransformOperationType = "base64"
	TransformCSV         TransformOperationType = "csv"
	TransformFlatten     TransformOperationType = "flatten"
	TransformMerge       TransformOperationType = "merge"
	TransformValidate    TransformOperationType = "validate"
)

// DataTransformConfig represents the configuration for a data transformation node
type DataTransformConfig struct {
	Operation    TransformOperationType `json:"operation"`
	Path         string                `json:"path"`
	Expression   string                `json:"expression"`
	Filter       string                `json:"filter"`
	Mapping      map[string]string     `json:"mapping"`
	SortField    string                `json:"sort_field"`
	SortOrder    string                `json:"sort_order"` // "asc" or "desc"
	JoinKey      string                `json:"join_key"`
	Separator    string                `json:"separator"`
	OldValue     string                `json:"old_value"`
	NewValue     string                `json:"new_value"`
	CaseType     string                `json:"case_type"` // "upper", "lower", "title"
	Base64Action string                `json:"base64_action"` // "encode" or "decode"
	CSVHeaders   []string              `json:"csv_headers"`
	Schema       map[string]interface{} `json:"schema"`
	DefaultValue interface{}           `json:"default_value"`
	ValidateRule string                `json:"validate_rule"`
}

// DataTransformNode represents a data transformation node
type DataTransformNode struct {
	config *DataTransformConfig
}

// NewDataTransformNode creates a new data transformation node
func NewDataTransformNode(config *DataTransformConfig) *DataTransformNode {
	// Set defaults
	if config.Separator == "" {
		config.Separator = ","
	}
	if config.SortOrder == "" {
		config.SortOrder = "asc"
	}

	return &DataTransformNode{
		config: config,
	}
}

// Execute executes the data transformation node
func (dtn *DataTransformNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Get the input data
	inputData, exists := inputs["data"]
	if !exists {
		return nil, fmt.Errorf("input data is required")
	}

	// Perform the transformation based on operation type
	switch dtn.config.Operation {
	case TransformJSONPath:
		return dtn.executeJSONPath(inputData)
	case TransformXPath:
		return dtn.executeXPath(inputData)
	case TransformFilter:
		return dtn.executeFilter(inputData)
	case TransformMap:
		return dtn.executeMap(inputData)
	case TransformReduce:
		return dtn.executeReduce(inputData)
	case TransformSort:
		return dtn.executeSort(inputData)
	case TransformJoin:
		return dtn.executeJoin(inputData)
	case TransformSplit:
		return dtn.executeSplit(inputData)
	case TransformReplace:
		return dtn.executeReplace(inputData)
	case TransformCaseConvert:
		return dtn.executeCaseConvert(inputData)
	case TransformBase64:
		return dtn.executeBase64(inputData)
	case TransformCSV:
		return dtn.executeCSV(inputData)
	case TransformFlatten:
		return dtn.executeFlatten(inputData)
	case TransformMerge:
		return dtn.executeMerge(inputData)
	case TransformValidate:
		return dtn.executeValidate(inputData)
	default:
		return nil, fmt.Errorf("unsupported transform operation: %s", dtn.config.Operation)
	}
}

// executeJSONPath extracts data using a simplified JSON path
func (dtn *DataTransformNode) executeJSONPath(inputData interface{}) (map[string]interface{}, error) {
	// In a real implementation, we would use a proper JSON path library
	// For now, we'll implement a simple path resolution
	result, err := dtn.extractByPath(inputData, dtn.config.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to extract by path: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"result":  result,
		"operation": string(TransformJSONPath),
		"path":    dtn.config.Path,
		"input":   inputData,
	}, nil
}

// executeXPath is similar to JSONPath but for XML processing (simplified for JSON)
func (dtn *DataTransformNode) executeXPath(inputData interface{}) (map[string]interface{}, error) {
	return dtn.executeJSONPath(inputData) // For now, treat the same as JSONPath
}

// executeFilter filters data based on a condition
func (dtn *DataTransformNode) executeFilter(inputData interface{}) (map[string]interface{}, error) {
	// For now, we'll implement a simple filter for arrays of objects
	inputSlice, ok := inputData.([]interface{})
	if !ok {
		return nil, fmt.Errorf("input data must be an array for filter operation")
	}

	var filtered []interface{}
	for _, item := range inputSlice {
		// Apply simple filter based on config
		if dtn.matchFilter(item) {
			filtered = append(filtered, item)
		}
	}

	return map[string]interface{}{
		"success": true,
		"result":  filtered,
		"operation": string(TransformFilter),
		"filter":  dtn.config.Filter,
		"input":   inputData,
		"count":   len(filtered),
	}, nil
}

// executeMap transforms data using a mapping
func (dtn *DataTransformNode) executeMap(inputData interface{}) (map[string]interface{}, error) {
	// For simple mapping, we'll transform keys based on the mapping config
	result, err := dtn.applyMapping(inputData)
	if err != nil {
		return nil, fmt.Errorf("failed to apply mapping: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"result":  result,
		"operation": string(TransformMap),
		"mapping": dtn.config.Mapping,
		"input":   inputData,
	}, nil
}

// executeReduce reduces an array to a single value
func (dtn *DataTransformNode) executeReduce(inputData interface{}) (map[string]interface{}, error) {
	// For now, we'll implement a simple sum operation
	inputSlice, ok := inputData.([]interface{})
	if !ok {
		return nil, fmt.Errorf("input data must be an array for reduce operation")
	}

	var sum float64
	for _, item := range inputSlice {
		switch v := item.(type) {
		case float64:
			sum += v
		case int:
			sum += float64(v)
		case string:
			if num, err := strconv.ParseFloat(v, 64); err == nil {
				sum += num
			}
		}
	}

	return map[string]interface{}{
		"success": true,
		"result":  sum,
		"operation": string(TransformReduce),
		"input":   inputData,
	}, nil
}

// executeSort sorts an array of objects
func (dtn *DataTransformNode) executeSort(inputData interface{}) (map[string]interface{}, error) {
	// For now, we'll return the input as-is with sort metadata
	// In a real implementation, we would sort the array

	return map[string]interface{}{
		"success": true,
		"result":  inputData,
		"operation": string(TransformSort),
		"field":   dtn.config.SortField,
		"order":   dtn.config.SortOrder,
		"input":   inputData,
	}, nil
}

// executeJoin joins arrays or objects
func (dtn *DataTransformNode) executeJoin(inputData interface{}) (map[string]interface{}, error) {
	// For now, return mock result
	return map[string]interface{}{
		"success": true,
		"result":  inputData,
		"operation": string(TransformJoin),
		"key":     dtn.config.JoinKey,
		"input":   inputData,
	}, nil
}

// executeSplit splits a string or array
func (dtn *DataTransformNode) executeSplit(inputData interface{}) (map[string]interface{}, error) {
	inputStr, ok := inputData.(string)
	if !ok {
		return nil, fmt.Errorf("input data must be a string for split operation")
	}

	parts := strings.Split(inputStr, dtn.config.Separator)

	return map[string]interface{}{
		"success": true,
		"result":  parts,
		"operation": string(TransformSplit),
		"separator": dtn.config.Separator,
		"input":   inputData,
		"count":   len(parts),
	}, nil
}

// executeReplace replaces substrings
func (dtn *DataTransformNode) executeReplace(inputData interface{}) (map[string]interface{}, error) {
	inputStr, ok := inputData.(string)
	if !ok {
		return nil, fmt.Errorf("input data must be a string for replace operation")
	}

	result := strings.ReplaceAll(inputStr, dtn.config.OldValue, dtn.config.NewValue)

	return map[string]interface{}{
		"success": true,
		"result":  result,
		"operation": string(TransformReplace),
		"old_value": dtn.config.OldValue,
		"new_value": dtn.config.NewValue,
		"input":   inputData,
	}, nil
}

// executeCaseConvert converts case of strings
func (dtn *DataTransformNode) executeCaseConvert(inputData interface{}) (map[string]interface{}, error) {
	inputStr, ok := inputData.(string)
	if !ok {
		return nil, fmt.Errorf("input data must be a string for case conversion")
	}

	var result string
	switch strings.ToLower(dtn.config.CaseType) {
	case "upper":
		result = strings.ToUpper(inputStr)
	case "lower":
		result = strings.ToLower(inputStr)
	case "title":
		result = strings.Title(strings.ToLower(inputStr))
	default:
		result = inputStr // default to no change
	}

	return map[string]interface{}{
		"success": true,
		"result":  result,
		"operation": string(TransformCaseConvert),
		"case_type": dtn.config.CaseType,
		"input":   inputData,
	}, nil
}

// executeBase64 encodes or decodes base64
func (dtn *DataTransformNode) executeBase64(inputData interface{}) (map[string]interface{}, error) {
	inputStr, ok := inputData.(string)
	if !ok {
		return nil, fmt.Errorf("input data must be a string for base64 operation")
	}

	var result string
	var err error

	switch dtn.config.Base64Action {
	case "encode":
		result = base64.StdEncoding.EncodeToString([]byte(inputStr))
	case "decode":
		decoded, decodeErr := base64.StdEncoding.DecodeString(inputStr)
		if decodeErr != nil {
			return nil, fmt.Errorf("failed to decode base64: %w", decodeErr)
		}
		result = string(decoded)
	default:
		result = inputStr
	}

	return map[string]interface{}{
		"success": true,
		"result":  result,
		"operation": string(TransformBase64),
		"action":  dtn.config.Base64Action,
		"input":   inputData,
	}, nil
}

// executeCSV converts between CSV and JSON
func (dtn *DataTransformNode) executeCSV(inputData interface{}) (map[string]interface{}, error) {
	switch dtn.config.Base64Action { // Using Base64Action field for CSV direction
	case "to_json":
		// Convert CSV string to JSON array
		csvStr, ok := inputData.(string)
		if !ok {
			return nil, fmt.Errorf("input data must be a string for CSV to JSON conversion")
		}
		
		records, err := csv.NewReader(strings.NewReader(csvStr)).ReadAll()
		if err != nil {
			return nil, fmt.Errorf("failed to parse CSV: %w", err)
		}
		
		var result []map[string]interface{}
		if len(records) > 0 {
			headers := records[0]
			for i, record := range records {
				if i == 0 { // Skip header row
					continue
				}
				
				row := make(map[string]interface{})
				for j, value := range record {
					if j < len(headers) {
						row[headers[j]] = value
					} else {
						row[fmt.Sprintf("column_%d", j)] = value
					}
				}
				result = append(result, row)
			}
		}
		
		return map[string]interface{}{
			"success": true,
			"result":  result,
			"operation": string(TransformCSV),
			"action":  "to_json",
			"input":   inputData,
		}, nil
		
	case "to_csv":
		// Convert JSON array to CSV string
		inputSlice, ok := inputData.([]interface{})
		if !ok {
			return nil, fmt.Errorf("input data must be an array for JSON to CSV conversion")
		}
		
		var records [][]string
		var headers []string
		
		// Extract headers from first object if it's a map
		if len(inputSlice) > 0 {
			if firstObj, ok := inputSlice[0].(map[string]interface{}); ok {
				for k := range firstObj {
					headers = append(headers, k)
				}
				records = append(records, headers)
				
				// Add each object as a row
				for _, item := range inputSlice {
					if obj, ok := item.(map[string]interface{}); ok {
						var row []string
						for _, header := range headers {
							if val, exists := obj[header]; exists {
								row = append(row, fmt.Sprintf("%v", val))
							} else {
								row = append(row, "")
							}
						}
						records = append(records, row)
					}
				}
			}
		}
		
		var csvBuilder strings.Builder
		for _, record := range records {
			csvBuilder.WriteString(strings.Join(record, dtn.config.Separator))
			csvBuilder.WriteString("\n")
		}
		
		return map[string]interface{}{
			"success": true,
			"result":  csvBuilder.String(),
			"operation": string(TransformCSV),
			"action":  "to_csv",
			"input":   inputData,
		}, nil
		
	default:
		return nil, fmt.Errorf("invalid CSV action: %s", dtn.config.Base64Action)
	}
}

// executeFlatten flattens nested objects/arrays
func (dtn *DataTransformNode) executeFlatten(inputData interface{}) (map[string]interface{}, error) {
	flattened := dtn.flatten(inputData, "", make(map[string]interface{}))
	
	return map[string]interface{}{
		"success": true,
		"result":  flattened,
		"operation": string(TransformFlatten),
		"input":   inputData,
	}, nil
}

// executeMerge merges objects/arrays
func (dtn *DataTransformNode) executeMerge(inputData interface{}) (map[string]interface{}, error) {
	// For now, return the input as-is
	// In a real implementation, we would merge multiple objects/arrays

	return map[string]interface{}{
		"success": true,
		"result":  inputData,
		"operation": string(TransformMerge),
		"input":   inputData,
	}, nil
}

// executeValidate validates data against schema/rules
func (dtn *DataTransformNode) executeValidate(inputData interface{}) (map[string]interface{}, error) {
	// Simple validation - check if required fields exist
	var errors []string

	if dtn.config.Schema != nil {
		if obj, ok := inputData.(map[string]interface{}); ok {
			for key, required := range dtn.config.Schema {
				if _, exists := obj[key]; !exists {
					errors = append(errors, fmt.Sprintf("missing required field: %s", key))
				} else {
					// Type validation could go here
					if required == "required" {
						if obj[key] == nil {
							errors = append(errors, fmt.Sprintf("required field %s is null", key))
						}
					}
				}
			}
		}
	}

	return map[string]interface{}{
		"success": true,
		"valid":   len(errors) == 0,
		"errors":  errors,
		"operation": string(TransformValidate),
		"input":   inputData,
	}, nil
}

// extractByPath extracts data by a simple dot notation path
func (dtn *DataTransformNode) extractByPath(data interface{}, path string) (interface{}, error) {
	if path == "" {
		return data, nil
	}

	// Split the path by dots
	parts := strings.Split(path, ".")

	current := data
	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		case map[string]string:
			current = v[part]
		default:
			return nil, fmt.Errorf("cannot navigate path on type %T", current)
		}

		if current == nil {
			return dtn.config.DefaultValue, nil
		}
	}

	return current, nil
}

// matchFilter checks if an item matches the filter
func (dtn *DataTransformNode) matchFilter(item interface{}) bool {
	// A simple string-based filter for now
	itemStr := fmt.Sprintf("%v", item)
	return strings.Contains(strings.ToLower(itemStr), strings.ToLower(dtn.config.Filter))
}

// applyMapping applies a mapping to transform data
func (dtn *DataTransformNode) applyMapping(data interface{}) (interface{}, error) {
	if dtn.config.Mapping == nil || len(dtn.config.Mapping) == 0 {
		return data, nil
	}

	// Handle mapping for objects
	if obj, ok := data.(map[string]interface{}); ok {
		result := make(map[string]interface{})
		for key, value := range obj {
			if newKey, exists := dtn.config.Mapping[key]; exists {
				result[newKey] = value
			} else {
				result[key] = value
			}
		}
		return result, nil
	}

	return data, nil
}

// flatten recursively flattens nested structures
func (dtn *DataTransformNode) flatten(data interface{}, prefix string, result map[string]interface{}) map[string]interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		for k, val := range v {
			newKey := k
			if prefix != "" {
				newKey = prefix + "." + k
			}
			dtn.flatten(val, newKey, result)
		}
	case []interface{}:
		for i, val := range v {
			newKey := fmt.Sprintf("%s[%d]", prefix, i)
			dtn.flatten(val, newKey, result)
		}
	default:
		result[prefix] = v
	}

	return result
}

// RegisterDataTransformNode registers the data transform node type with the engine
func RegisterDataTransformNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("data_transform", func(config map[string]interface{}) (engine.NodeInstance, error) {
		var operation TransformOperationType
		if opVal, exists := config["operation"]; exists {
			if opStr, ok := opVal.(string); ok {
				operation = TransformOperationType(opStr)
			}
		}

		var path string
		if pathVal, exists := config["path"]; exists {
			if pathStr, ok := pathVal.(string); ok {
				path = pathStr
			}
		}

		var expression string
		if exprVal, exists := config["expression"]; exists {
			if exprStr, ok := exprVal.(string); ok {
				expression = exprStr
			}
		}

		var filter string
		if filterVal, exists := config["filter"]; exists {
			if filterStr, ok := filterVal.(string); ok {
				filter = filterStr
			}
		}

		var mapping map[string]string
		if mapVal, exists := config["mapping"]; exists {
			if mapObj, ok := mapVal.(map[string]interface{}); ok {
				mapping = make(map[string]string)
				for k, v := range mapObj {
					if vStr, ok := v.(string); ok {
						mapping[k] = vStr
					}
				}
			}
		}

		var sortField string
		if fieldVal, exists := config["sort_field"]; exists {
			if fieldStr, ok := fieldVal.(string); ok {
				sortField = fieldStr
			}
		}

		var sortOrder string
		if orderVal, exists := config["sort_order"]; exists {
			if orderStr, ok := orderVal.(string); ok {
				sortOrder = orderStr
			}
		}

		var joinKey string
		if keyVal, exists := config["join_key"]; exists {
			if keyStr, ok := keyVal.(string); ok {
				joinKey = keyStr
			}
		}

		var separator string
		if sepVal, exists := config["separator"]; exists {
			if sepStr, ok := sepVal.(string); ok {
				separator = sepStr
			}
		}

		var oldValue string
		if oldVal, exists := config["old_value"]; exists {
			if oldStr, ok := oldVal.(string); ok {
				oldValue = oldStr
			}
		}

		var newValue string
		if newVal, exists := config["new_value"]; exists {
			if newStr, ok := newVal.(string); ok {
				newValue = newStr
			}
		}

		var caseType string
		if caseVal, exists := config["case_type"]; exists {
			if caseStr, ok := caseVal.(string); ok {
				caseType = caseStr
			}
		}

		var base64Action string
		if actionVal, exists := config["base64_action"]; exists {
			if actionStr, ok := actionVal.(string); ok {
				base64Action = actionStr
			}
		}

		var schema map[string]interface{}
		if schemaVal, exists := config["schema"]; exists {
			if schemaObj, ok := schemaVal.(map[string]interface{}); ok {
				schema = schemaObj
			}
		}

		var validateRule string
		if ruleVal, exists := config["validate_rule"]; exists {
			if ruleStr, ok := ruleVal.(string); ok {
				validateRule = ruleStr
			}
		}

		var csvHeaders []string
		if headersVal, exists := config["csv_headers"]; exists {
			if headersSlice, ok := headersVal.([]interface{}); ok {
				for _, header := range headersSlice {
					if headerStr, ok := header.(string); ok {
						csvHeaders = append(csvHeaders, headerStr)
					}
				}
			}
		}

		var defaultValue interface{}
		if defaultVal, exists := config["default_value"]; exists {
			defaultValue = defaultVal
		}

		nodeConfig := &DataTransformConfig{
			Operation:    operation,
			Path:         path,
			Expression:   expression,
			Filter:       filter,
			Mapping:      mapping,
			SortField:    sortField,
			SortOrder:    sortOrder,
			JoinKey:      joinKey,
			Separator:    separator,
			OldValue:     oldValue,
			NewValue:     newValue,
			CaseType:     caseType,
			Base64Action: base64Action,
			CSVHeaders:   csvHeaders,
			Schema:       schema,
			DefaultValue: defaultValue,
			ValidateRule: validateRule,
		}

		return NewDataTransformNode(nodeConfig), nil
	})
}