package transform

import (
	"encoding/json"
	"fmt"
)

// JSONTransformer provides methods to manipulate JSON data
type JSONTransformer struct{}

// Parse parses a JSON string into a map
func (t *JSONTransformer) Parse(input string) (interface{}, error) {
	var result interface{}
	err := json.Unmarshal([]byte(input), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return result, nil
}

// Stringify converts an object to a JSON string
func (t *JSONTransformer) Stringify(input interface{}) (string, error) {
	bytes, err := json.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("failed to stringify object: %w", err)
	}
	return string(bytes), nil
}

// GetValue extracts a value from a JSON object using a key path
// Note: This is a simplified implementation. A real one would use something like GJSON or a proper path parser.
func (t *JSONTransformer) GetValue(input map[string]interface{}, key string) (interface{}, bool) {
	val, ok := input[key]
	return val, ok
}
