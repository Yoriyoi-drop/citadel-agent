// backend/internal/nodes/utils/helpers.go
package utils

// GetStringVal safely extracts a string value
func GetStringVal(v interface{}, defaultValue string) string {
	if v == nil {
		return defaultValue
	}
	if s, ok := v.(string); ok {
		return s
	}
	return defaultValue
}

// GetBoolVal safely extracts a bool value
func GetBoolVal(v interface{}, defaultValue bool) bool {
	if v == nil {
		return defaultValue
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return defaultValue
}

// GetIntVal safely extracts an int value
func GetIntVal(v interface{}, defaultValue int) int {
	if v == nil {
		return defaultValue
	}
	if f, ok := v.(float64); ok {
		return int(f)
	}
	if i, ok := v.(int); ok {
		return i
	}
	return defaultValue
}

// GetFloat64Val safely extracts a float64 value
func GetFloat64Val(v interface{}, defaultValue float64) float64 {
	if v == nil {
		return defaultValue
	}
	if f, ok := v.(float64); ok {
		return f
	}
	if i, ok := v.(int); ok {
		return float64(i)
	}
	return defaultValue
}

// Contains checks if a slice contains a string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}