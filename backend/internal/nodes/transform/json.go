package transform

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var (
	ErrInvalidJSON   = errors.New("invalid JSON")
	ErrInvalidPath   = errors.New("invalid JSON path")
	ErrInvalidSchema = errors.New("invalid JSON schema")
)

// JSONTransformer handles JSON transformation operations
type JSONTransformer struct{}

// NewJSONTransformer creates a new JSON transformer
func NewJSONTransformer() *JSONTransformer {
	return &JSONTransformer{}
}

// Parse parses JSON string to object
func (t *JSONTransformer) Parse(jsonStr string) (interface{}, error) {
	var result interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}
	return result, nil
}

// Stringify converts object to JSON string
func (t *JSONTransformer) Stringify(obj interface{}) (string, error) {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// PrettyPrint formats JSON with indentation
func (t *JSONTransformer) PrettyPrint(jsonStr string) (string, error) {
	var obj interface{}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	bytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Minify removes whitespace from JSON
func (t *JSONTransformer) Minify(jsonStr string) (string, error) {
	var obj interface{}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	bytes, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Get retrieves value at JSON path using gjson
func (t *JSONTransformer) Get(jsonStr string, path string) (interface{}, error) {
	if !gjson.Valid(jsonStr) {
		return nil, ErrInvalidJSON
	}

	result := gjson.Get(jsonStr, path)
	if !result.Exists() {
		return nil, nil
	}

	return result.Value(), nil
}

// Set sets value at JSON path using sjson
func (t *JSONTransformer) Set(jsonStr string, path string, value interface{}) (string, error) {
	result, err := sjson.Set(jsonStr, path, value)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidPath, err)
	}
	return result, nil
}

// Delete deletes value at JSON path
func (t *JSONTransformer) Delete(jsonStr string, path string) (string, error) {
	result, err := sjson.Delete(jsonStr, path)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidPath, err)
	}
	return result, nil
}

// Merge merges two JSON objects
func (t *JSONTransformer) Merge(json1, json2 string) (string, error) {
	var obj1, obj2 map[string]interface{}

	if err := json.Unmarshal([]byte(json1), &obj1); err != nil {
		return "", fmt.Errorf("invalid first JSON: %w", err)
	}

	if err := json.Unmarshal([]byte(json2), &obj2); err != nil {
		return "", fmt.Errorf("invalid second JSON: %w", err)
	}

	// Merge obj2 into obj1
	for k, v := range obj2 {
		obj1[k] = v
	}

	bytes, err := json.Marshal(obj1)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// Filter filters array based on condition
func (t *JSONTransformer) Filter(jsonStr string, arrayPath string, filterFn func(interface{}) bool) (string, error) {
	result := gjson.Get(jsonStr, arrayPath)
	if !result.IsArray() {
		return "", errors.New("path does not point to an array")
	}

	var filtered []interface{}
	result.ForEach(func(key, value gjson.Result) bool {
		if filterFn(value.Value()) {
			filtered = append(filtered, value.Value())
		}
		return true
	})

	return t.Set(jsonStr, arrayPath, filtered)
}

// Map transforms array elements
func (t *JSONTransformer) Map(jsonStr string, arrayPath string, mapFn func(interface{}) interface{}) (string, error) {
	result := gjson.Get(jsonStr, arrayPath)
	if !result.IsArray() {
		return "", errors.New("path does not point to an array")
	}

	var mapped []interface{}
	result.ForEach(func(key, value gjson.Result) bool {
		mapped = append(mapped, mapFn(value.Value()))
		return true
	})

	return t.Set(jsonStr, arrayPath, mapped)
}

// Validate validates JSON against a simple schema
func (t *JSONTransformer) Validate(jsonStr string, requiredFields []string) error {
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	for _, field := range requiredFields {
		if _, exists := obj[field]; !exists {
			return fmt.Errorf("%w: missing required field '%s'", ErrInvalidSchema, field)
		}
	}

	return nil
}

// ToMap converts JSON string to map
func (t *JSONTransformer) ToMap(jsonStr string) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}
	return result, nil
}

// ToArray converts JSON string to array
func (t *JSONTransformer) ToArray(jsonStr string) ([]interface{}, error) {
	var result []interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}
	return result, nil
}

// FromMap converts map to JSON string
func (t *JSONTransformer) FromMap(m map[string]interface{}) (string, error) {
	bytes, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FromArray converts array to JSON string
func (t *JSONTransformer) FromArray(arr []interface{}) (string, error) {
	bytes, err := json.Marshal(arr)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
