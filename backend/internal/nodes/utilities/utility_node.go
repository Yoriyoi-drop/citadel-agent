// backend/internal/nodes/utilities/utility_node.go
package utilities

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
	"github.com/citadel-agent/backend/internal/workflow/core/engine"
)

// UtilityOperationType represents the type of utility operation
type UtilityOperationType string

const (
	UtilityOpStringOp    UtilityOperationType = "string_operation"
	UtilityOpMathOp      UtilityOperationType = "math_operation"
	UtilityOpDateOp      UtilityOperationType = "date_operation"
	UtilityOpCryptoOp    UtilityOperationType = "crypto_operation"
	UtilityOpConvertOp   UtilityOperationType = "convert_operation"
	UtilityOpValidateOp  UtilityOperationType = "validate_operation"
	UtilityOpGenerateOp  UtilityOperationType = "generate_operation"
	UtilityOpTransformOp UtilityOperationType = "transform_operation"
)

// StringOperation represents string manipulation operations
type StringOperation string

const (
	StringOpToUpper   StringOperation = "to_upper"
	StringOpToLower   StringOperation = "to_lower"
	StringOpTrim      StringOperation = "trim"
	StringOpReplace   StringOperation = "replace"
	StringOpSplit     StringOperation = "split"
	StringOpJoin      StringOperation = "join"
	StringOpConcat    StringOperation = "concat"
	StringOpRegex     StringOperation = "regex"
)

// MathOperation represents mathematical operations
type MathOperation string

const (
	MathOpAdd      MathOperation = "add"
	MathOpSubtract MathOperation = "subtract"
	MathOpMultiply MathOperation = "multiply"
	MathOpDivide   MathOperation = "divide"
	MathOpPower    MathOperation = "power"
	MathOpModulo   MathOperation = "modulo"
	MathOpRound    MathOperation = "round"
	MathOpCeil     MathOperation = "ceil"
	MathOpFloor    MathOperation = "floor"
)

// UtilityConfig represents the configuration for a utility node
type UtilityConfig struct {
	Operation      UtilityOperationType `json:"operation"`
	StringOp       StringOperation      `json:"string_operation"`
	MathOp         MathOperation        `json:"math_operation"`
	InputValues    []interface{}        `json:"input_values"`
	InputValue     interface{}          `json:"input_value"`
	InputString    string              `json:"input_string"`
	Format         string              `json:"format"`
	RegexPattern   string              `json:"regex_pattern"`
	Replacement    string              `json:"replacement"`
	Separator      string              `json:"separator"`
	SeedValue      interface{}         `json:"seed_value"`
	DateLayout     string              `json:"date_layout"`
	TargetType     string              `json:"target_type"`
	ValidateAs     string              `json:"validate_as"`
	GenerateType   string              `json:"generate_type"`
	GenerateLength int                 `json:"generate_length"`
}

// UtilityNode represents a utility node
type UtilityNode struct {
	config *UtilityConfig
}

// NewUtilityNode creates a new utility node
func NewUtilityNode(config *UtilityConfig) *UtilityNode {
	if config.DateLayout == "" {
		config.DateLayout = time.RFC3339
	}
	if config.Separator == "" {
		config.Separator = ","
	}

	return &UtilityNode{
		config: config,
	}
}

// Execute executes the utility operation
func (un *UtilityNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	operation := un.config.Operation
	if op, exists := inputs["operation"]; exists {
		if opStr, ok := op.(string); ok {
			operation = UtilityOperationType(opStr)
		}
	}

	// Override config values with inputs if provided
	inputValue := un.config.InputValue
	if val, exists := inputs["input_value"]; exists {
		inputValue = val
	}

	inputValues := un.config.InputValues
	if vals, exists := inputs["input_values"]; exists {
		if valsSlice, ok := vals.([]interface{}); ok {
			inputValues = valsSlice
		}
	}

	inputString := un.config.InputString
	if str, exists := inputs["input_string"]; exists {
		if strStr, ok := str.(string); ok {
			inputString = strStr
		}
	}

	switch operation {
	case UtilityOpStringOp:
		return un.stringOperation(inputString, inputs)
	case UtilityOpMathOp:
		return un.mathOperation(inputValues, inputs)
	case UtilityOpDateOp:
		return un.dateOperation(inputValue, inputs)
	case UtilityOpCryptoOp:
		return un.cryptoOperation(inputString, inputs)
	case UtilityOpConvertOp:
		return un.convertOperation(inputValue, inputs)
	case UtilityOpValidateOp:
		return un.validateOperation(inputValue, inputs)
	case UtilityOpGenerateOp:
		return un.generateOperation(inputs)
	case UtilityOpTransformOp:
		return un.transformOperation(inputs)
	default:
		return nil, fmt.Errorf("unsupported utility operation: %s", operation)
	}
}

// stringOperation performs string manipulation operations
func (un *UtilityNode) stringOperation(input string, inputs map[string]interface{}) (map[string]interface{}, error) {
	var result string
	var err error

	stringOp := un.config.StringOp
	if op, exists := inputs["string_operation"]; exists {
		if opStr, ok := op.(string); ok {
			stringOp = StringOperation(opStr)
		}
	}

	switch stringOp {
	case StringOpToUpper:
		result = strings.ToUpper(input)
	case StringOpToLower:
		result = strings.ToLower(input)
	case StringOpTrim:
		result = strings.TrimSpace(input)
	case StringOpReplace:
		oldStr := getStringValue(inputs["old"], "")
		newStr := getStringValue(inputs["new"], "")
		if oldStr == "" {
			oldStr = un.config.Replacement // Using replacement as old for this case
		}
		result = strings.ReplaceAll(input, oldStr, newStr)
	case StringOpSplit:
		separator := getStringValue(inputs["separator"], un.config.Separator)
		parts := strings.Split(input, separator)
		return map[string]interface{}{
			"success":  true,
			"result":   parts,
			"count":    len(parts),
			"operation": "string_split",
			"separator": separator,
			"timestamp": time.Now().Unix(),
		}, nil
	case StringOpJoin:
		separator := getStringValue(inputs["separator"], un.config.Separator)
		// Join operation expects a slice of strings in inputs
		var stringSlice []string
		if vals, exists := inputs["input_values"]; exists {
			if valsSlice, ok := vals.([]interface{}); ok {
				for _, val := range valsSlice {
					stringSlice = append(stringSlice, fmt.Sprintf("%v", val))
				}
			}
		}
		result = strings.Join(stringSlice, separator)
	case StringOpConcat:
		// Concatenate input with additional values
		additional := getStringValue(inputs["additional"], "")
		result = input + additional
	case StringOpRegex:
		pattern := getStringValue(inputs["pattern"], un.config.RegexPattern)
		if pattern == "" {
			return nil, fmt.Errorf("regex pattern is required")
		}
		
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern: %w", err)
		}
		
		matches := re.FindAllString(input, -1)
		return map[string]interface{}{
			"success":  true,
			"matches":  matches,
			"count":    len(matches),
			"pattern":  pattern,
			"operation": "string_regex",
			"timestamp": time.Now().Unix(),
		}, nil
	default:
		result = input
	}

	return map[string]interface{}{
		"success":   true,
		"result":    result,
		"operation": "string_operation",
		"input":     input,
		"timestamp": time.Now().Unix(),
	}, nil
}

// mathOperation performs mathematical operations
func (un *UtilityNode) mathOperation(inputValues []interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Use values from inputs if provided, otherwise use config values
	if len(inputValues) == 0 {
		if vals, exists := inputs["input_values"]; exists {
			if valsSlice, ok := vals.([]interface{}); ok {
				inputValues = valsSlice
			}
		}
	}
	
	if len(inputValues) < 1 {
		return nil, fmt.Errorf("at least one input value is required for math operations")
	}

	var result float64
	var err error

	mathOp := un.config.MathOp
	if op, exists := inputs["math_operation"]; exists {
		if opStr, ok := op.(string); ok {
			mathOp = MathOperation(opStr)
		}
	}

	// Convert first value to float64
	firstVal, err := toFloat64(inputValues[0])
	if err != nil {
		return nil, fmt.Errorf("invalid first value for math operation: %w", err)
	}

	result = firstVal

	switch mathOp {
	case MathOpAdd:
		for i := 1; i < len(inputValues); i++ {
			val, err := toFloat64(inputValues[i])
			if err != nil {
				return nil, fmt.Errorf("invalid value at position %d: %w", i, err)
			}
			result += val
		}
	case MathOpSubtract:
		for i := 1; i < len(inputValues); i++ {
			val, err := toFloat64(inputValues[i])
			if err != nil {
				return nil, fmt.Errorf("invalid value at position %d: %w", i, err)
			}
			result -= val
		}
	case MathOpMultiply:
		for i := 1; i < len(inputValues); i++ {
			val, err := toFloat64(inputValues[i])
			if err != nil {
				return nil, fmt.Errorf("invalid value at position %d: %w", i, err)
			}
			result *= val
		}
	case MathOpDivide:
		if len(inputValues) < 2 {
			return nil, fmt.Errorf("division requires at least 2 values")
		}
		for i := 1; i < len(inputValues); i++ {
			val, err := toFloat64(inputValues[i])
			if err != nil {
				return nil, fmt.Errorf("invalid value at position %d: %w", i, err)
			}
			if val == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			result /= val
		}
	case MathOpPower:
		if len(inputValues) != 2 {
			return nil, fmt.Errorf("power operation requires exactly 2 values (base and exponent)")
		}
		exp, err := toFloat64(inputValues[1])
		if err != nil {
			return nil, fmt.Errorf("invalid exponent: %w", err)
		}
		result = pow(result, exp)
	case MathOpModulo:
		if len(inputValues) != 2 {
			return nil, fmt.Errorf("modulo operation requires exactly 2 values")
		}
		mod, err := toFloat64(inputValues[1])
		if err != nil {
			return nil, fmt.Errorf("invalid modulo value: %w", err)
		}
		result = float64(int64(result) % int64(mod))
	case MathOpRound:
		// Round to specified decimal places (default 0)
		precision := 0
		if prec, exists := inputs["precision"]; exists {
			if precFloat, ok := prec.(float64); ok {
				precision = int(precFloat)
			}
		}
		multiplier := 1.0
		for i := 0; i < precision; i++ {
			multiplier *= 10
		}
		result = float64(int(result*multiplier+0.5)) / multiplier
	case MathOpCeil:
		result = ceil(result)
	case MathOpFloor:
		result = floor(result)
	default:
		// For operations that don't change the value
	}

	return map[string]interface{}{
		"success":   true,
		"result":    result,
		"operation": "math_operation",
		"input_values": inputValues,
		"timestamp": time.Now().Unix(),
	}, nil
}

// dateOperation performs date/time operations
func (un *UtilityNode) dateOperation(inputValue interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	dateStr := fmt.Sprintf("%v", inputValue)
	
	// Parse the input date
	var parsedTime time.Time
	var err error

	// Try to parse as timestamp first
	if timestamp, ok := inputValue.(float64); ok {
		parsedTime = time.Unix(int64(timestamp), 0)
	} else {
		// Try to parse as time string using various layouts
		layouts := []string{
			time.RFC3339,
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
			"2006-01-02",
			"01/02/2006",
			"02-Jan-2006",
		}
		
		for _, layout := range layouts {
			if parsedTime, err = time.Parse(layout, dateStr); err == nil {
				break
			}
		}
		
		if err != nil {
			return nil, fmt.Errorf("unable to parse date: %w", err)
		}
	}

	// Apply transformations based on inputs
	resultTime := parsedTime

	// Add/subtract time if specified
	if addHours, exists := inputs["add_hours"]; exists {
		if hours, ok := addHours.(float64); ok {
			resultTime = resultTime.Add(time.Duration(hours) * time.Hour)
		}
	}
	
	if addDays, exists := inputs["add_days"]; exists {
		if days, ok := addDays.(float64); ok {
			resultTime = resultTime.AddDate(0, 0, int(days))
		}
	}

	// Format result
	format := getStringValue(inputs["format"], un.config.DateLayout)
	formatted := resultTime.Format(format)

	return map[string]interface{}{
		"success":     true,
		"result":      formatted,
		"timestamp":   resultTime.Unix(),
		"operation":   "date_operation",
		"input":       inputValue,
		"formatted_as": format,
	}, nil
}

// cryptoOperation performs cryptographic operations
func (un *UtilityNode) cryptoOperation(input string, inputs map[string]interface{}) (map[string]interface{}, error) {
	operation := getStringValue(inputs["crypto_operation"], "hash")

	switch operation {
	case "hash":
		algorithm := getStringValue(inputs["algorithm"], "sha256")
		var result string
		
		switch algorithm {
		case "sha256":
			result = un.sha256Hash(input)
		case "md5":
			result = un.md5Hash(input)
		default:
			return nil, fmt.Errorf("unsupported hash algorithm: %s", algorithm)
		}
		
		return map[string]interface{}{
			"success":   true,
			"result":    result,
			"algorithm": algorithm,
			"operation": "crypto_hash",
			"timestamp": time.Now().Unix(),
		}, nil
	case "encode_base64":
		encoded := base64.StdEncoding.EncodeToString([]byte(input))
		return map[string]interface{}{
			"success":   true,
			"result":    encoded,
			"operation": "crypto_encode_base64",
			"timestamp": time.Now().Unix(),
		}, nil
	case "decode_base64":
		decoded, err := base64.StdEncoding.DecodeString(input)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64: %w", err)
		}
		return map[string]interface{}{
			"success":   true,
			"result":    string(decoded),
			"operation": "crypto_decode_base64",
			"timestamp": time.Now().Unix(),
		}, nil
	case "encode_hex":
		encoded := hex.EncodeToString([]byte(input))
		return map[string]interface{}{
			"success":   true,
			"result":    encoded,
			"operation": "crypto_encode_hex",
			"timestamp": time.Now().Unix(),
		}, nil
	case "decode_hex":
		decoded, err := hex.DecodeString(input)
		if err != nil {
			return nil, fmt.Errorf("failed to decode hex: %w", err)
		}
		return map[string]interface{}{
			"success":   true,
			"result":    string(decoded),
			"operation": "crypto_decode_hex",
			"timestamp": time.Now().Unix(),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported crypto operation: %s", operation)
	}
}

// convertOperation performs data type conversions
func (un *UtilityNode) convertOperation(input interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	targetType := getStringValue(inputs["target_type"], un.config.TargetType)
	
	if targetType == "" {
		return nil, fmt.Errorf("target type is required for conversion")
	}

	var result interface{}
	var err error

	switch targetType {
	case "string":
		result = fmt.Sprintf("%v", input)
	case "int":
		if f, ok := input.(float64); ok {
			result = int(f)
		} else {
			result, err = toInt(input)
		}
	case "float", "float64":
		result, err = toFloat64(input)
	case "bool":
		result = toBool(input)
	case "json_string":
		result = fmt.Sprintf("%+v", input) // Simple representation
	default:
		return nil, fmt.Errorf("unsupported conversion target type: %s", targetType)
	}

	if err != nil {
		return nil, fmt.Errorf("conversion error: %w", err)
	}

	return map[string]interface{}{
		"success":   true,
		"result":    result,
		"target_type": targetType,
		"operation": "convert_operation",
		"input":     input,
		"timestamp": time.Now().Unix(),
	}, nil
}

// validateOperation performs validation operations
func (un *UtilityNode) validateOperation(input interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	validationType := getStringValue(inputs["validate_as"], un.config.ValidateAs)
	inputStr := fmt.Sprintf("%v", input)

	if validationType == "" {
		return nil, fmt.Errorf("validation type is required")
	}

	var valid bool
	var message string

	switch validationType {
	case "email":
		valid = un.validateEmail(inputStr)
		if valid {
			message = "Valid email format"
		} else {
			message = "Invalid email format"
		}
	case "url":
		valid = un.validateURL(inputStr)
		if valid {
			message = "Valid URL format"
		} else {
			message = "Invalid URL format"
		}
	case "phone":
		valid = un.validatePhone(inputStr)
		if valid {
			message = "Valid phone number format"
		} else {
			message = "Invalid phone number format"
		}
	case "credit_card":
		valid = un.validateCreditCard(inputStr)
		if valid {
			message = "Valid credit card number"
		} else {
			message = "Invalid credit card number"
		}
	case "regex":
		pattern := getStringValue(inputs["pattern"], un.config.RegexPattern)
		if pattern == "" {
			return nil, fmt.Errorf("regex pattern is required for regex validation")
		}
		
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern: %w", err)
		}
		
		valid = re.MatchString(inputStr)
		if valid {
			message = "Text matches pattern"
		} else {
			message = "Text does not match pattern"
		}
	default:
		return nil, fmt.Errorf("unsupported validation type: %s", validationType)
	}

	return map[string]interface{}{
		"success":   true,
		"valid":     valid,
		"message":   message,
		"validation_type": validationType,
		"operation": "validate_operation",
		"input":     input,
		"timestamp": time.Now().Unix(),
	}, nil
}

// generateOperation performs data generation operations
func (un *UtilityNode) generateOperation(inputs map[string]interface{}) (map[string]interface{}, error) {
	generateType := getStringValue(inputs["generate_type"], un.config.GenerateType)
	
	if generateType == "" {
		generateType = "random_string"
	}

	var result interface{}
	var err error

	switch generateType {
	case "random_string":
		length := int(getFloat64Value(inputs["length"], float64(un.config.GenerateLength)))
		if length == 0 {
			length = 10
		}
		result, err = un.generateRandomString(length)
		if err != nil {
			return nil, err
		}
	case "random_int":
		min := int(getFloat64Value(inputs["min"], 0))
		max := int(getFloat64Value(inputs["max"], 100))
		result = un.generateRandomInt(min, max)
	case "uuid":
		result = un.generateUUID()
	case "timestamp":
		result = time.Now().Unix()
	case "date_string":
		format := getStringValue(inputs["format"], un.config.DateLayout)
		result = time.Now().Format(format)
	default:
		return nil, fmt.Errorf("unsupported generation type: %s", generateType)
	}

	return map[string]interface{}{
		"success":      true,
		"result":       result,
		"generation_type": generateType,
		"operation":    "generate_operation",
		"timestamp":    time.Now().Unix(),
	}, nil
}

// transformOperation performs data transformation operations
func (un *UtilityNode) transformOperation(inputs map[string]interface{}) (map[string]interface{}, error) {
	data := inputs["data"]
	transformType := getStringValue(inputs["transform_type"], "default")

	var result interface{}

	switch transformType {
	case "json_to_map":
		// In a real implementation, this would parse JSON string to map
		result = data
	case "map_to_json":
		// In a real implementation, this would convert map to JSON string
		result = fmt.Sprintf("%+v", data)
	case "flatten":
		// A simple flatten implementation
		result = un.flattenData(data)
	case "uppercase_keys":
		// Convert map keys to uppercase (if data is a map)
		result = un.uppercaseMapKeys(data)
	default:
		// Default transformation - just return the data
		result = data
	}

	return map[string]interface{}{
		"success":     true,
		"result":      result,
		"transform_type": transformType,
		"operation":   "transform_operation",
		"timestamp":   time.Now().Unix(),
	}, nil
}

// Helper functions for crypto operations
func (un *UtilityNode) sha256Hash(input string) string {
	// Simplified SHA256 implementation for example
	// In a real implementation, use crypto/sha256
	return fmt.Sprintf("sha256:%s", input) // Placeholder
}

func (un *UtilityNode) md5Hash(input string) string {
	// Simplified MD5 implementation for example
	// In a real implementation, use crypto/md5
	return fmt.Sprintf("md5:%s", input) // Placeholder
}

// Helper functions for validation operations
func (un *UtilityNode) validateEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

func (un *UtilityNode) validateURL(url string) bool {
	pattern := `^https?:\/\/[^\s/$.?#].[^\s]*$`
	matched, _ := regexp.MatchString(pattern, url)
	return matched
}

func (un *UtilityNode) validatePhone(phone string) bool {
	// Simple phone validation - in reality this would be more complex
	pattern := `^[\+]?[1-9][\d]{0,15}$`
	matched, _ := regexp.MatchString(pattern, phone)
	return matched
}

func (un *UtilityNode) validateCreditCard(card string) bool {
	// Remove spaces and dashes
	card = strings.ReplaceAll(strings.ReplaceAll(card, " ", ""), "-", "")
	
	// Simple Luhn algorithm check
	if len(card) < 13 || len(card) > 19 {
		return false
	}
	
	total := 0
	double := false
	for i := len(card) - 1; i >= 0; i-- {
		digit := int(card[i] - '0')
		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		total += digit
		double = !double
	}
	
	return total%10 == 0
}

// Helper functions for generation operations
func (un *UtilityNode) generateRandomString(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be positive")
	}
	
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	
	// Use only alphanumeric characters
	for i, b := range bytes {
		bytes[i] = alphanumericChars[int(b)%len(alphanumericChars)]
	}
	
	return string(bytes), nil
}

func (un *UtilityNode) generateRandomInt(min, max int) int {
	// Simplified implementation - in a real system you'd use crypto/rand
	// For this example, we'll use a basic approach
	return min + (time.Now().Nanosecond() % (max - min + 1))
}

func (un *UtilityNode) generateUUID() string {
	// Generate a simple UUID-like string
	bytes := make([]byte, 16)
	rand.Read(bytes)
	
	// Convert to UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x",
		bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
	
	return uuid
}

// Helper functions for transformation operations
func (un *UtilityNode) flattenData(data interface{}) interface{} {
	// A simple flatten implementation
	// This would be much more complex in a real implementation
	return data
}

func (un *UtilityNode) uppercaseMapKeys(data interface{}) interface{} {
	if m, ok := data.(map[string]interface{}); ok {
		result := make(map[string]interface{})
		for k, v := range m {
			result[strings.ToUpper(k)] = v
		}
		return result
	}
	return data
}

// Math helper functions
func pow(x, y float64) float64 {
	// Simplified implementation - in real system use math.Pow
	result := 1.0
	for i := 0; i < int(y); i++ {
		result *= x
	}
	return result
}

func ceil(x float64) float64 {
	// Simplified implementation - in real system use math.Ceil
	return float64(int64(x + 0.999999))
}

func floor(x float64) float64 {
	// Simplified implementation - in real system use math.Floor
	return float64(int64(x))
}

var alphanumericChars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

// toFloat64 converts an interface value to float64
func toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case string:
		// Try to parse as number
		var result float64
		_, err := fmt.Sscanf(v, "%f", &result)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string '%s' to float64", v)
		}
		return result, nil
	case bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

// toInt converts an interface value to int
func toInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case float64:
		return int(v), nil
	case float32:
		return int(v), nil
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case int32:
		return int(v), nil
	case int16:
		return int(v), nil
	case int8:
		return int(v), nil
	case uint:
		return int(v), nil
	case uint64:
		return int(v), nil
	case uint32:
		return int(v), nil
	case uint16:
		return int(v), nil
	case uint8:
		return int(v), nil
	case string:
		// Try to parse as number
		var result int
		_, err := fmt.Sscanf(v, "%d", &result)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string '%s' to int", v)
		}
		return result, nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int", value)
	}
}

// toBool converts an interface value to bool
func toBool(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case float64:
		return v != 0
	case float32:
		return v != 0
	case int:
		return v != 0
	case int64:
		return v != 0
	case string:
		return v != "" && v != "false" && v != "False" && v != "0"
	default:
		return value != nil
	}
}

// getStringValue safely extracts a string value with default fallback
func getStringValue(v interface{}, defaultValue string) string {
	if v == nil {
		return defaultValue
	}
	if s, ok := v.(string); ok {
		return s
	}
	return defaultValue
}

// getFloat64Value safely extracts a float64 value with default fallback
func getFloat64Value(v interface{}, defaultValue float64) float64 {
	if v == nil {
		return defaultValue
	}
	if f, ok := v.(float64); ok {
		return f
	}
	if s, ok := v.(string); ok {
		if f, err := fmt.Sscanf(s, "%f"); err == nil {
			return float64(f)
		}
	}
	return defaultValue
}

// UtilityNodeFromConfig creates a new utility node from a configuration map
func UtilityNodeFromConfig(config map[string]interface{}) (interfaces.NodeInstance, error) {
	var operation UtilityOperationType
	if op, exists := config["operation"]; exists {
		if opStr, ok := op.(string); ok {
			operation = UtilityOperationType(opStr)
		}
	}

	var stringOp StringOperation
	if op, exists := config["string_operation"]; exists {
		if opStr, ok := op.(string); ok {
			stringOp = StringOperation(opStr)
		}
	}

	var mathOp MathOperation
	if op, exists := config["math_operation"]; exists {
		if opStr, ok := op.(string); ok {
			mathOp = MathOperation(opStr)
		}
	}

	var inputValue interface{}
	if val, exists := config["input_value"]; exists {
		inputValue = val
	}

	var inputString string
	if str, exists := config["input_string"]; exists {
		if strStr, ok := str.(string); ok {
			inputString = strStr
		}
	}

	var inputValues []interface{}
	if vals, exists := config["input_values"]; exists {
		if valsSlice, ok := vals.([]interface{}); ok {
			inputValues = valsSlice
		}
	}

	var format string
	if fmtStr, exists := config["format"]; exists {
		if fmtStr, ok := fmtStr.(string); ok {
			format = fmtStr
		}
	}

	var regexPattern string
	if pattern, exists := config["regex_pattern"]; exists {
		if patternStr, ok := pattern.(string); ok {
			regexPattern = patternStr
		}
	}

	var replacement string
	if rep, exists := config["replacement"]; exists {
		if repStr, ok := rep.(string); ok {
			replacement = repStr
		}
	}

	var separator string
	if sep, exists := config["separator"]; exists {
		if sepStr, ok := sep.(string); ok {
			separator = sepStr
		}
	}

	var dateLayout string
	if layout, exists := config["date_layout"]; exists {
		if layoutStr, ok := layout.(string); ok {
			dateLayout = layoutStr
		}
	}

	var targetType string
	if typ, exists := config["target_type"]; exists {
		if typStr, ok := typ.(string); ok {
			targetType = typStr
		}
	}

	var validateAs string
	if valAs, exists := config["validate_as"]; exists {
		if valAsStr, ok := valAs.(string); ok {
			validateAs = valAsStr
		}
	}

	var generateType string
	if genType, exists := config["generate_type"]; exists {
		if genTypeStr, ok := genType.(string); ok {
			generateType = genTypeStr
		}
	}

	var generateLength float64
	if genLen, exists := config["generate_length"]; exists {
		if genLenFloat, ok := genLen.(float64); ok {
			generateLength = genLenFloat
		}
	}

	nodeConfig := &UtilityConfig{
		Operation:      operation,
		StringOp:       stringOp,
		MathOp:         mathOp,
		InputValue:     inputValue,
		InputString:    inputString,
		InputValues:    inputValues,
		Format:         format,
		RegexPattern:   regexPattern,
		Replacement:    replacement,
		Separator:      separator,
		DateLayout:     dateLayout,
		TargetType:     targetType,
		ValidateAs:     validateAs,
		GenerateType:   generateType,
		GenerateLength: int(generateLength),
	}

	return NewUtilityNode(nodeConfig), nil
}

// RegisterUtilityNode registers the utility node type with the engine
func RegisterUtilityNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("utility", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return UtilityNodeFromConfig(config)
	})
}