package utils

import (
	"strconv"
)

// GetString safely extracts a string value from interface{}
func GetString(v interface{}, defaultValue string) string {
	if v == nil {
		return defaultValue
	}
	if s, ok := v.(string); ok {
		return s
	}
	return defaultValue
}

// GetFloat64 safely extracts a float64 value from interface{}
func GetFloat64(v interface{}, defaultValue float64) float64 {
	if v == nil {
		return defaultValue
	}
	if f, ok := v.(float64); ok {
		return f
	}
	if s, ok := v.(string); ok {
		// Simple conversion from string
		if val, err := strconv.ParseFloat(s, 64); err == nil {
			return val
		}
	}
	return defaultValue
}

// GetInt safely extracts an int value from interface{}
func GetInt(v interface{}, defaultValue int) int {
	if v == nil {
		return defaultValue
	}
	if f, ok := v.(float64); ok {
		return int(f)
	}
	if s, ok := v.(string); ok {
		// Simple conversion from string
		if val, err := strconv.Atoi(s); err == nil {
			return val
		}
	}
	return defaultValue
}

// GetBool safely extracts a bool value from interface{}
func GetBool(v interface{}, defaultValue bool) bool {
	if v == nil {
		return defaultValue
	}
	if b, ok := v.(bool); ok {
		return b
	}
	if s, ok := v.(string); ok {
		return s == "true" || s == "1" || s == "yes" || s == "on"
	}
	return defaultValue
}