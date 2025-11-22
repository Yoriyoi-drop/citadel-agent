// backend/internal/nodes/utils/helpers.go
package utils

import (
	"strconv"
	"strings"
)

// GetStringValue safely extracts a string value with default fallback
func GetStringValue(v interface{}, defaultValue string) string {
	if v == nil {
		return defaultValue
	}
	if s, ok := v.(string); ok {
		return s
	}
	if i, ok := v.(int); ok {
		return strconv.Itoa(i)
	}
	if f, ok := v.(float64); ok {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	if b, ok := v.(bool); ok {
		return strconv.FormatBool(b)
	}
	return defaultValue
}

// GetFloat64Value safely extracts a float64 value with default fallback
func GetFloat64Value(v interface{}, defaultValue float64) float64 {
	if v == nil {
		return defaultValue
	}
	if f, ok := v.(float64); ok {
		return f
	}
	if s, ok := v.(string); ok {
		// Parse string to float64
		if val, err := strconv.ParseFloat(s, 64); err == nil {
			return val
		}
		return defaultValue
	}
	if i, ok := v.(int); ok {
		return float64(i)
	}
	if b, ok := v.(bool); ok {
		if b {
			return 1.0
		}
		return 0.0
	}
	return defaultValue
}

// GetIntValue safely extracts an int value with default fallback
func GetIntValue(v interface{}, defaultValue int) int {
	if v == nil {
		return defaultValue
	}
	if i, ok := v.(int); ok {
		return i
	}
	if f, ok := v.(float64); ok {
		return int(f)
	}
	if s, ok := v.(string); ok {
		// Parse string to int
		if val, err := strconv.Atoi(s); err == nil {
			return val
		}
		return defaultValue
	}
	return defaultValue
}

// GetBoolValue safely extracts a bool value with default fallback
func GetBoolValue(v interface{}, defaultValue bool) bool {
	if v == nil {
		return defaultValue
	}
	if b, ok := v.(bool); ok {
		return b
	}
	if s, ok := v.(string); ok {
		// Parse string to bool using standard package
		if val, err := strconv.ParseBool(strings.ToLower(s)); err == nil {
			return val
		}
		// Check for common truthy values
		s = strings.ToLower(s)
		return s == "true" || s == "1" || s == "yes" || s == "on"
	}
	if i, ok := v.(int); ok {
		return i != 0
	}
	if f, ok := v.(float64); ok {
		return f != 0.0
	}
	return defaultValue
}

// GetMapValue safely extracts a map[string]interface{} value with default fallback
func GetMapValue(v interface{}, defaultValue map[string]interface{}) map[string]interface{} {
	if v == nil {
		return defaultValue
	}
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return defaultValue
}

// GetSliceValue safely extracts a []interface{} value with default fallback
func GetSliceValue(v interface{}, defaultValue []interface{}) []interface{} {
	if v == nil {
		return defaultValue
	}
	if s, ok := v.([]interface{}); ok {
		return s
	}
	if sliceV, ok := v.([]map[string]interface{}); ok {
		// Convert []map[string]interface{} to []interface{}
		result := make([]interface{}, len(sliceV))
		for i, item := range sliceV {
			result[i] = item
		}
		return result
	}
	return defaultValue
}