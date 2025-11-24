package utility

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
)

// DataTransformerNode implements a node that transforms data
type DataTransformerNode struct {
	id            string
	nodeType      string
	transformType string
	mapping       map[string]string
	operation     string
	parameters    map[string]interface{}
	config        map[string]interface{}
}

// Initialize sets up the transformer node with configuration
func (dt *DataTransformerNode) Initialize(config map[string]interface{}) error {
	dt.config = config

	if transformType, ok := config["transform_type"]; ok {
		if tt, ok := transformType.(string); ok {
			dt.transformType = tt
		} else {
			return fmt.Errorf("transform_type must be a string")
		}
	}

	if mapping, ok := config["mapping"]; ok {
		if m, ok := mapping.(map[string]interface{}); ok {
			dt.mapping = make(map[string]string)
			for k, v := range m {
				if vStr, ok := v.(string); ok {
					dt.mapping[k] = vStr
				} else {
					dt.mapping[k] = fmt.Sprintf("%v", v)
				}
			}
		}
	}

	if operation, ok := config["operation"]; ok {
		if op, ok := operation.(string); ok {
			dt.operation = op
		}
	}

	if parameters, ok := config["parameters"]; ok {
		if p, ok := parameters.(map[string]interface{}); ok {
			dt.parameters = p
		}
	}

	return nil
}

// Execute transforms the input data
func (dt *DataTransformerNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	var outputData map[string]interface{}

	switch dt.transformType {
	case "mapping":
		outputData = dt.applyMapping(inputs)
	case "filtering":
		outputData = dt.applyFiltering(inputs)
	case "custom":
		outputData = dt.applyCustomOperation(inputs)
	default:
		// Default behavior: return input data unchanged
		outputData = inputs
	}

	return outputData, nil
}

// applyMapping applies field mapping to transform input data
func (dt *DataTransformerNode) applyMapping(inputData map[string]interface{}) map[string]interface{} {
	output := make(map[string]interface{})

	for inputKey, outputKey := range dt.mapping {
		if value, exists := inputData[inputKey]; exists {
			output[outputKey] = value
		}
	}

	return output
}

// applyFiltering filters input data based on criteria
func (dt *DataTransformerNode) applyFiltering(inputData map[string]interface{}) map[string]interface{} {
	output := make(map[string]interface{})

	// For now, we'll just return data that matches a simple filter
	// More complex filtering can be added later
	for key, value := range inputData {
		include := true

		// Apply filters if specified
		if filters, exists := dt.parameters["filters"]; exists {
			if filterList, ok := filters.([]interface{}); ok {
				for _, filter := range filterList {
					if filterMap, isMap := filter.(map[string]interface{}); isMap {
						if field, exists := filterMap["field"]; exists {
							if fieldStr, ok := field.(string); ok {
								if fieldStr == key {
									// Apply the condition
									if condition, exists := filterMap["condition"]; exists {
										if condStr, ok := condition.(string); ok {
											if condStr == "has_value" {
												include = value != nil && value != ""
											} else if condStr == "not_null" {
												include = value != nil
											} else if condStr == "not_empty" {
												include = value != ""
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}

		if include {
			output[key] = value
		}
	}

	return output
}

// applyCustomOperation applies a custom transformation operation
func (dt *DataTransformerNode) applyCustomOperation(inputData map[string]interface{}) map[string]interface{} {
	// For now, implement some basic operations
	switch dt.operation {
	case "json_parse":
		return dt.applyJSONParse(inputData)
	case "json_stringify":
		return dt.applyJSONStringify(inputData)
	case "string_operations":
		return dt.applyStringOperations(inputData)
	default:
		// Return input data unchanged for unknown operations
		return inputData
	}
}

// applyJSONParse parses JSON strings in the data
func (dt *DataTransformerNode) applyJSONParse(inputData map[string]interface{}) map[string]interface{} {
	output := make(map[string]interface{})

	for key, value := range inputData {
		if strValue, ok := value.(string); ok {
			var parsed interface{}
			if err := json.Unmarshal([]byte(strValue), &parsed); err == nil {
				output[key] = parsed
			} else {
				// If parsing fails, keep the original value
				output[key] = value
			}
		} else {
			output[key] = value
		}
	}

	return output
}

// applyJSONStringify converts objects to JSON strings
func (dt *DataTransformerNode) applyJSONStringify(inputData map[string]interface{}) map[string]interface{} {
	output := make(map[string]interface{})

	for key, value := range inputData {
		if strValue, ok := value.(string); ok {
			// If it's already a string, keep as is
			output[key] = strValue
		} else {
			// Convert to JSON string
			if bytes, err := json.Marshal(value); err == nil {
				output[key] = string(bytes)
			} else {
				// If marshaling fails, convert to string using fmt
				output[key] = fmt.Sprintf("%v", value)
			}
		}
	}

	return output
}

// applyStringOperations applies string manipulation operations
func (dt *DataTransformerNode) applyStringOperations(inputData map[string]interface{}) map[string]interface{} {
	output := make(map[string]interface{})

	for key, value := range inputData {
		strValue := fmt.Sprintf("%v", value) // Convert to string first

		// Apply operations specified in parameters
		if ops, exists := dt.parameters["operations"]; exists {
			if opList, ok := ops.([]interface{}); ok {
				for _, op := range opList {
					if opMap, isMap := op.(map[string]interface{}); isMap {
						if opName, exists := opMap["operation"]; exists {
							if opNameStr, ok := opName.(string); ok {
								switch opNameStr {
								case "trim":
									strValue = strings.TrimSpace(strValue)
								case "lowercase":
									strValue = strings.ToLower(strValue)
								case "uppercase":
									strValue = strings.ToUpper(strValue)
								case "replace":
									if from, exists := opMap["from"]; exists {
										if to, exists := opMap["to"]; exists {
											fromStr := fmt.Sprintf("%v", from)
											toStr := fmt.Sprintf("%v", to)
											strValue = strings.ReplaceAll(strValue, fromStr, toStr)
										}
									}
								}
							}
						}
					}
				}
			}
		}

		output[key] = strValue
	}

	return output
}

// GetType returns the type of the node
func (dt *DataTransformerNode) GetType() string {
	return dt.nodeType
}

// GetID returns the unique identifier for this node instance
func (dt *DataTransformerNode) GetID() string {
	return dt.id
}

// NewTransformerNode creates a new data transformer node constructor for the registry
func NewTransformerNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	node := &DataTransformerNode{
		id:       fmt.Sprintf("transformer_%d", time.Now().UnixNano()),
		nodeType: "data_transformer",
	}

	if err := node.Initialize(config); err != nil {
		return nil, err
	}

	return node, nil
}