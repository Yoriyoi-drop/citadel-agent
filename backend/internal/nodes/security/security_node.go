// backend/internal/nodes/security/security_node.go
package security

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/citadel-agent/backend/internal/interfaces"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/scrypt"
)

// SecurityOperationType represents the type of security operation
type SecurityOperationType string

const (
	SecurityOpHash          SecurityOperationType = "hash"
	SecurityOpEncrypt       SecurityOperationType = "encrypt"
	SecurityOpDecrypt       SecurityOperationType = "decrypt"
	SecurityOpSign          SecurityOperationType = "sign"
	SecurityOpVerify        SecurityOperationType = "verify"
	SecurityOpValidate      SecurityOperationType = "validate"
	SecurityOpMask          SecurityOperationType = "mask"
	SecurityOpGenerateToken SecurityOperationType = "generate_token"
	SecurityOpValidateToken SecurityOperationType = "validate_token"
)

// SecurityAlgorithm represents the cryptographic algorithm to use
type SecurityAlgorithm string

const (
	AlgorithmSHA256   SecurityAlgorithm = "sha256"
	AlgorithmSHA512   SecurityAlgorithm = "sha512"
	AlgorithmAES256   SecurityAlgorithm = "aes256"
	AlgorithmHMACSHA256 SecurityAlgorithm = "hmac_sha256"
	AlgorithmBCrypt   SecurityAlgorithm = "bcrypt"
	AlgorithmScrypt   SecurityAlgorithm = "scrypt"
	AlgorithmRSA      SecurityAlgorithm = "rsa"
	AlgorithmECDSA    SecurityAlgorithm = "ecdsa"
)

// SecurityNodeConfig represents the configuration for a security node
type SecurityNodeConfig struct {
	Operation       SecurityOperationType `json:"operation"`
	Algorithm       SecurityAlgorithm     `json:"algorithm"`
	SecretKey       string              `json:"secret_key"`
	PublicKey       string              `json:"public_key"`
	PrivateKey      string              `json:"private_key"`
	Iterations      int                 `json:"iterations"`
	KeyLength       int                 `json:"key_length"`
	Salt            string              `json:"salt"`
	IV              string              `json:"iv"`
	IncludeSalt     bool                `json:"include_salt"`
	ExcludeChars    []string            `json:"exclude_chars"`
	TokenExpiry     time.Duration       `json:"token_expiry"`
	ValidateRules   []ValidationRule      `json:"validate_rules"`
	MaskPattern     string              `json:"mask_pattern"`
	MaskCustomPattern string            `json:"mask_custom_pattern"`
}

// ValidationRule represents a validation rule for data
type ValidationRule struct {
	Type    string      `json:"type"`    // email, phone, url, credit_card, etc.
	Pattern string      `json:"pattern"` // Regex pattern
	Min     interface{} `json:"min"`     // Min length/value
	Max     interface{} `json:"max"`     // Max length/value
	Required bool       `json:"required"`
	Message  string      `json:"message"`
}

// SecurityNode represents a security operation node
type SecurityNode struct {
	config *SecurityNodeConfig
}

// NewSecurityNode creates a new security node
func NewSecurityNode(config *SecurityNodeConfig) *SecurityNode {
	if config.Iterations == 0 {
		switch config.Algorithm {
		case AlgorithmBCrypt:
			config.Iterations = 12
		case AlgorithmScrypt:
			config.Iterations = 32768
		default:
			config.Iterations = 10000
		}
	}

	if config.KeyLength == 0 {
		config.KeyLength = 32 // Default to 256 bits
	}

	if config.TokenExpiry == 0 {
		config.TokenExpiry = 24 * time.Hour // Default to 24 hours
	}

	return &SecurityNode{
		config: config,
	}
}

// Execute executes the security operation
func (sn *SecurityNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	operation := sn.config.Operation
	if op, exists := inputs["operation"]; exists {
		if opStr, ok := op.(string); ok {
			operation = SecurityOperationType(opStr)
		}
	}

	algorithm := sn.config.Algorithm
	if algo, exists := inputs["algorithm"]; exists {
		if algoStr, ok := algo.(string); ok {
			algorithm = SecurityAlgorithm(algoStr)
		}
	}

	secretKey := sn.config.SecretKey
	if key, exists := inputs["secret_key"]; exists {
		if keyStr, ok := key.(string); ok {
			secretKey = keyStr
		}
	}

	data := ""
	if d, exists := inputs["data"]; exists {
		if dStr, ok := d.(string); ok {
			data = dStr
		}
	}

	// Perform the security operation based on type
	switch operation {
	case SecurityOpHash:
		return sn.hashData(data, algorithm, secretKey)
	case SecurityOpEncrypt:
		return sn.encryptData(data, algorithm, secretKey)
	case SecurityOpDecrypt:
		return sn.decryptData(data, algorithm, secretKey)
	case SecurityOpSign:
		return sn.signData(data, algorithm, secretKey)
	case SecurityOpVerify:
		signature := ""
		if sig, exists := inputs["signature"]; exists {
			if sigStr, ok := sig.(string); ok {
				signature = sigStr
			}
		}
		return sn.verifyData(data, signature, algorithm, secretKey)
	case SecurityOpValidate:
		return sn.validateData(data, inputs)
	case SecurityOpMask:
		pattern := sn.config.MaskPattern
		if pat, exists := inputs["mask_pattern"]; exists {
			if patStr, ok := pat.(string); ok {
				pattern = patStr
			}
		}
		return sn.maskData(data, pattern)
	case SecurityOpGenerateToken:
		expiry := sn.config.TokenExpiry
		if exp, exists := inputs["expiry_seconds"]; exists {
			if expFloat, ok := exp.(float64); ok {
				expiry = time.Duration(expFloat) * time.Second
			}
		}
		return sn.generateToken(data, secretKey, expiry)
	case SecurityOpValidateToken:
		return sn.validateToken(data, secretKey)
	default:
		return nil, fmt.Errorf("unsupported security operation: %s", operation)
	}
}

// hashData performs hashing operation
func (sn *SecurityNode) hashData(data string, algorithm SecurityAlgorithm, secretKey string) (map[string]interface{}, error) {
	var hashResult string
	var err error

	switch algorithm {
	case AlgorithmSHA256:
		h := sha256.New()
		h.Write([]byte(data))
		hashResult = hex.EncodeToString(h.Sum(nil))
	case AlgorithmSHA512:
		h := sha512.New()
		h.Write([]byte(data))
		hashResult = hex.EncodeToString(h.Sum(nil))
	case AlgorithmHMACSHA256:
		if secretKey == "" {
			return nil, fmt.Errorf("secret key is required for HMAC")
		}
		h := hmac.New(sha256.New, []byte(secretKey))
		h.Write([]byte(data))
		hashResult = hex.EncodeToString(h.Sum(nil))
	case AlgorithmBCrypt:
		hashBytes, err := bcrypt.GenerateFromPassword([]byte(data), sn.config.Iterations)
		if err != nil {
			return nil, fmt.Errorf("bcrypt hash failed: %w", err)
		}
		hashResult = string(hashBytes)
	case AlgorithmScrypt:
		salt := []byte(sn.config.Salt)
		if len(salt) == 0 {
			salt = make([]byte, 32)
			if _, err := rand.Read(salt); err != nil {
				return nil, fmt.Errorf("failed to generate salt: %w", err)
			}
		}

		dk, err := scrypt.Key([]byte(data), salt, sn.config.Iterations, 8, 1, sn.config.KeyLength)
		if err != nil {
			return nil, fmt.Errorf("scrypt hash failed: %w", err)
		}

		if sn.config.IncludeSalt {
			hashResult = base64.StdEncoding.EncodeToString(salt) + ":" + base64.StdEncoding.EncodeToString(dk)
		} else {
			hashResult = base64.StdEncoding.EncodeToString(dk)
		}
	default:
		return nil, fmt.Errorf("unsupported hashing algorithm: %s", algorithm)
	}

	return map[string]interface{}{
		"success":     true,
		"result":      hashResult,
		"algorithm":   string(algorithm),
		"operation":   string(SecurityOpHash),
		"input_size":  len(data),
		"timestamp":   time.Now().Unix(),
	}, nil
}

// encryptData performs encryption operation
func (sn *SecurityNode) encryptData(data, key string, algorithm SecurityAlgorithm) (map[string]interface{}, error) {
	switch algorithm {
	case AlgorithmAES256:
		return sn.aes256Encrypt(data, key)
	default:
		return nil, fmt.Errorf("unsupported encryption algorithm: %s", algorithm)
	}
}

// aes256Encrypt performs AES-256 encryption
func (sn *SecurityNode) aes256Encrypt(plaintext, key string) (map[string]interface{}, error) {
	// Create cipher
	block, err := aes.NewCipher([]byte(createKey(key, 32)))
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create IV
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	// Pad plaintext
	padding := aes.BlockSize - len(plaintext)%aes.BlockSize
	paddedText := []byte(plaintext)
	for i := 0; i < padding; i++ {
		paddedText = append(paddedText, byte(padding))
	}

	// Encrypt
	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(paddedText))
	mode.CryptBlocks(ciphertext, paddedText)

	// Combine IV and ciphertext
	result := append(iv, ciphertext...)

	return map[string]interface{}{
		"success":     true,
		"encrypted_data": base64.StdEncoding.EncodeToString(result),
		"algorithm":   string(AlgorithmAES256),
		"operation":   string(SecurityOpEncrypt),
		"iv":          base64.StdEncoding.EncodeToString(iv),
		"input_size":  len(plaintext),
		"output_size": len(result),
		"timestamp":   time.Now().Unix(),
	}, nil
}

// decryptData performs decryption operation
func (sn *SecurityNode) decryptData(encryptedData, key string, algorithm SecurityAlgorithm) (map[string]interface{}, error) {
	switch algorithm {
	case AlgorithmAES256:
		return sn.aes256Decrypt(encryptedData, key)
	default:
		return nil, fmt.Errorf("unsupported decryption algorithm: %s", algorithm)
	}
}

// aes256Decrypt performs AES-256 decryption
func (sn *SecurityNode) aes256Decrypt(encryptedDataBase64, key string) (map[string]interface{}, error) {
	// Decode the base64 data
	encryptedData, err := base64.StdEncoding.DecodeString(encryptedDataBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted data: %w", err)
	}

	// Extract IV (first 16 bytes) and ciphertext
	if len(encryptedData) < aes.BlockSize {
		return nil, fmt.Errorf("encrypted data too short")
	}

	iv := encryptedData[:aes.BlockSize]
	ciphertext := encryptedData[aes.BlockSize:]

	// Create cipher
	block, err := aes.NewCipher([]byte(createKey(key, 32)))
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Decrypt
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove padding
	padding := int(plaintext[len(plaintext)-1])
	if padding > len(plaintext) {
		return nil, fmt.Errorf("invalid padding")
	}
	plaintext = plaintext[:len(plaintext)-padding]

	return map[string]interface{}{
		"success":     true,
		"decrypted_data": string(plaintext),
		"algorithm":   string(AlgorithmAES256),
		"operation":   string(SecurityOpDecrypt),
		"input_size":  len(encryptedData),
		"output_size": len(plaintext),
		"timestamp":   time.Now().Unix(),
	}, nil
}

// signData signs data with specified algorithm
func (sn *SecurityNode) signData(data, key string, algorithm SecurityAlgorithm) (map[string]interface{}, error) {
	if key == "" {
		return nil, fmt.Errorf("key is required for signing")
	}

	var signature string

	switch algorithm {
	case AlgorithmHMACSHA256:
		h := hmac.New(sha256.New, []byte(key))
		h.Write([]byte(data))
		signature = hex.EncodeToString(h.Sum(nil))
	default:
		return nil, fmt.Errorf("unsupported signing algorithm: %s", algorithm)
	}

	return map[string]interface{}{
		"success":   true,
		"signature": signature,
		"algorithm": string(algorithm),
		"operation": string(SecurityOpSign),
		"input_size": len(data),
		"timestamp":  time.Now().Unix(),
	}, nil
}

// verifyData verifies a signature
func (sn *SecurityNode) verifyData(data, signature, key string, algorithm SecurityAlgorithm) (map[string]interface{}, error) {
	if key == "" {
		return nil, fmt.Errorf("key is required for verification")
	}

	if signature == "" {
		return nil, fmt.Errorf("signature is required")
	}

	var expectedSig string
	var err error

	switch algorithm {
	case AlgorithmHMACSHA256:
		h := hmac.New(sha256.New, []byte(key))
		h.Write([]byte(data))
		expectedSig = hex.EncodeToString(h.Sum(nil))
	default:
		return nil, fmt.Errorf("unsupported verification algorithm: %s", algorithm)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to compute signature: %w", err)
	}

	verified := hmac.Equal([]byte(signature), []byte(expectedSig))

	return map[string]interface{}{
		"success":   true,
		"verified":  verified,
		"algorithm": string(algorithm),
		"operation": string(SecurityOpVerify),
		"input_size": len(data),
		"timestamp":  time.Now().Unix(),
	}, nil
}

// validateData validates input data against rules
func (sn *SecurityNode) validateData(data string, inputs map[string]interface{}) (map[string]interface{}, error) {
	results := make(map[string]interface{})
	allValid := true

	// Get validation rules from config or inputs
	rules := sn.config.ValidateRules

	// Override with input rules if provided
	if inputRules, exists := inputs["validation_rules"]; exists {
		if rulesSlice, ok := inputRules.([]interface{}); ok {
			rules = make([]ValidationRule, len(rulesSlice))
			for i, rule := range rulesSlice {
				if ruleMap, ok := rule.(map[string]interface{}); ok {
					var minVal, maxVal interface{}
					if min, exists := ruleMap["min"]; exists {
						minVal = min
					}
					if max, exists := ruleMap["max"]; exists {
						maxVal = max
					}

					rules[i] = ValidationRule{
						Type:     getStringValue(ruleMap["type"]),
						Pattern:  getStringValue(ruleMap["pattern"]),
						Min:      minVal,
						Max:      maxVal,
						Required: getBoolValue(ruleMap["required"]),
						Message:  getStringValue(ruleMap["message"]),
					}
				}
			}
		}
	}

	// Apply each validation rule
	for i, rule := range rules {
		ruleName := fmt.Sprintf("rule_%d", i)

		valid, message := sn.validateRule(data, rule)
		results[ruleName] = map[string]interface{}{
			"valid":   valid,
			"rule":    rule,
			"message": message,
		}

		if !valid {
			allValid = false
		}
	}

	return map[string]interface{}{
		"success":    allValid,
		"valid":      allValid,
		"results":    results,
		"operation":  string(SecurityOpValidate),
		"input_size": len(data),
		"timestamp":  time.Now().Unix(),
	}, nil
}

// validateRule applies a single validation rule to data
func (sn *SecurityNode) validateRule(data string, rule ValidationRule) (bool, string) {
	// Check required
	if rule.Required && data == "" {
		return false, rule.Message
	}

	if data == "" {
		return true, "Valid (data is empty but not required)" // Empty values are OK if not required
	}

	// Check regex pattern
	if rule.Pattern != "" {
		matched, err := regexp.MatchString(rule.Pattern, data)
		if err != nil || !matched {
			return false, fmt.Sprintf("Value does not match pattern: %s", rule.Pattern)
		}
	}

	// Check length/numeric constraints
	if rule.Min != nil || rule.Max != nil {
		switch rule.Type {
		case "string", "text":
			length := len(data)

			if rule.Min != nil {
				if minInt, ok := rule.Min.(float64); ok {
					if float64(length) < minInt {
						return false, fmt.Sprintf("Length (%d) is less than minimum (%.0f)", length, minInt)
					}
				}
			}

			if rule.Max != nil {
				if maxInt, ok := rule.Max.(float64); ok {
					if float64(length) > maxInt {
						return false, fmt.Sprintf("Length (%d) is greater than maximum (%.0f)", length, maxInt)
					}
				}
			}
		case "number", "integer", "float":
			num, err := strconv.ParseFloat(data, 64)
			if err != nil {
				return false, "Value is not a valid number"
			}

			if rule.Min != nil {
				if minFloat, ok := rule.Min.(float64); ok {
					if num < minFloat {
						return false, fmt.Sprintf("Value (%.2f) is less than minimum (%.2f)", num, minFloat)
					}
				}
			}

			if rule.Max != nil {
				if maxFloat, ok := rule.Max.(float64); ok {
					if num > maxFloat {
						return false, fmt.Sprintf("Value (%.2f) is greater than maximum (%.2f)", num, maxFloat)
					}
				}
			}
		}
	}

	return true, "Valid"
}

// maskData masks sensitive data according to specified pattern
func (sn *SecurityNode) maskData(data, pattern string) (map[string]interface{}, error) {
	maskedData := data

	// Apply masking based on pattern
	switch strings.ToLower(pattern) {
	case "email":
		maskedData = sn.maskEmail(data)
	case "phone", "telephone":
		maskedData = sn.maskPhone(data)
	case "credit_card", "card":
		maskedData = sn.maskCreditCard(data)
	case "ssn", "social_security":
		maskedData = sn.maskSSN(data)
	case "custom":
		// Use custom pattern from config/inputs
		if sn.config.MaskCustomPattern != "" {
			maskedData = sn.maskCustom(data, sn.config.MaskCustomPattern)
		}
	default:
		// Default to hiding middle portion
		if len(data) > 6 {
			start := data[:2]
			end := data[len(data)-2:]
			middleLength := len(data) - 4
			if middleLength > 0 {
				maskedData = start + strings.Repeat("*", middleLength) + end
			}
		}
	}

	return map[string]interface{}{
		"success":      true,
		"original":     data,
		"masked":       maskedData,
		"pattern":      pattern,
		"operation":    string(SecurityOpMask),
		"input_size":   len(data),
		"output_size":  len(maskedData),
		"timestamp":    time.Now().Unix(),
	}, nil
}

// maskEmail masks an email address
func (sn *SecurityNode) maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email // Not a valid email, return as is
	}

	localPart := parts[0]
	domain := parts[1]

	if len(localPart) <= 2 {
		maskedLocal := strings.Repeat("*", len(localPart))
		return maskedLocal + "@" + domain
	}

	start := localPart[:1]
	end := localPart[len(localPart)-1:]
	middleLength := len(localPart) - 2
	maskedMiddle := strings.Repeat("*", middleLength)

	return start + maskedMiddle + end + "@" + domain
}

// maskPhone masks a phone number
func (sn *SecurityNode) maskPhone(phone string) string {
	// Remove non-digit characters
	cleaned := ""
	for _, char := range phone {
		if unicode.IsDigit(char) {
			cleaned += string(char)
		}
	}

	if len(cleaned) < 4 {
		return strings.Repeat("*", len(phone))
	}

	// Keep last 4 digits, mask the rest
	prefix := cleaned[:len(cleaned)-4]
	suffix := cleaned[len(cleaned)-4:]

	maskedPrefix := strings.Repeat("*", len(prefix))

	// Preserve original formatting
	result := ""
	phoneIndex := 0
	maskIndex := 0

	for _, char := range phone {
		if unicode.IsDigit(char) {
			if maskIndex < len(maskedPrefix) {
				result += "*"
			} else {
				result += string(suffix[maskIndex-len(maskedPrefix)])
			}
			maskIndex++
		} else {
			result += string(char)
		}
	}

	return result
}

// maskCreditCard masks a credit card number
func (sn *SecurityNode) maskCreditCard(card string) string {
	// Remove non-digit characters
	cleaned := ""
	for _, char := range card {
		if unicode.IsDigit(char) {
			cleaned += string(char)
		}
	}

	if len(cleaned) < 8 {
		return strings.Repeat("*", len(card))
	}

	// Show first 4 and last 4, mask the middle
	prefix := cleaned[:4]
	suffix := cleaned[len(cleaned)-4:]
	middleLength := len(cleaned) - 8
	maskedMiddle := strings.Repeat("*", middleLength)

	maskedCleaned := prefix + maskedMiddle + suffix

	// Preserve original formatting
	result := ""
	cardIndex := 0
	maskIndex := 0

	for _, char := range card {
		if unicode.IsDigit(char) {
			if maskIndex < 4 || maskIndex >= len(cleaned)-4 {
				result += string(maskedCleaned[maskIndex])
			} else {
				result += "*"
			}
			maskIndex++
		} else {
			result += string(char)
		}
	}

	return result
}

// maskSSN masks a social security number
func (sn *SecurityNode) maskSSN(ssn string) string {
	// Remove non-digit characters
	cleaned := ""
	for _, char := range ssn {
		if unicode.IsDigit(char) {
			cleaned += string(char)
		}
	}

	if len(cleaned) != 9 {
		return strings.Repeat("*", len(ssn))
	}

	// Format: XXX-XX-XXXX, mask middle 5 digits
	prefix := cleaned[:3]
	middle := cleaned[3:5]
	suffix := cleaned[5:]

	masked := prefix + strings.Repeat("*", 5) + suffix

	// Preserve original formatting
	result := ""
	ssnIndex := 0
	maskIndex := 0

	for _, char := range ssn {
		if unicode.IsDigit(char) {
			if maskIndex < 3 || maskIndex >= 8 {
				result += string(masked[maskIndex])
			} else {
				result += "*"
			}
			maskIndex++
		} else {
			result += string(char)
		}
	}

	return result
}

// maskCustom applies custom masking pattern
func (sn *SecurityNode) maskCustom(data, pattern string) string {
	// Pattern format: "start:length:end"
	// e.g., "2:4:2" means show first 2, mask next 4, show last 2
	parts := strings.Split(pattern, ":")
	if len(parts) != 3 {
		return strings.Repeat("*", len(data))
	}

	start, err1 := strconv.Atoi(parts[0])
	middle, err2 := strconv.Atoi(parts[1])
	end, err3 := strconv.Atoi(parts[2])

	if err1 != nil || err2 != nil || err3 != nil {
		return strings.Repeat("*", len(data))
	}

	if start < 0 || middle < 0 || end < 0 || start+middle+end > len(data) {
		return strings.Repeat("*", len(data))
	}

	result := data[:start] + strings.Repeat("*", middle) + data[len(data)-end:]

	return result
}

// generateToken generates a secure token
func (sn *SecurityNode) generateToken(payload, secretKey string, expiry time.Duration) (map[string]interface{}, error) {
	if secretKey == "" {
		return nil, fmt.Errorf("secret key is required for token generation")
	}

	// Create a simple token by combining payload, timestamp, and signature
	timestamp := time.Now().Unix()
	expireAt := time.Now().Add(expiry).Unix()

	tokenData := fmt.Sprintf("%s|%d|%d", payload, timestamp, expireAt)

	// Sign the token data
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(tokenData))
	signature := hex.EncodeToString(h.Sum(nil))

	// Combine token data and signature
	fullToken := tokenData + "|" + signature

	encodedToken := base64.URLEncoding.EncodeToString([]byte(fullToken))

	return map[string]interface{}{
		"success":   true,
		"token":     encodedToken,
		"payload":   payload,
		"issued_at": timestamp,
		"expires_at": expireAt,
		"operation": string(SecurityOpGenerateToken),
		"timestamp": time.Now().Unix(),
	}, nil
}

// validateToken validates a token
func (sn *SecurityNode) validateToken(token, secretKey string) (map[string]interface{}, error) {
	if secretKey == "" {
		return nil, fmt.Errorf("secret key is required for token validation")
	}

	// Decode the token
	decoded, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token format")
	}

	tokenStr := string(decoded)
	parts := strings.Split(tokenStr, "|")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid token format")
	}

	payload := parts[0]
	issuedAt, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid issued timestamp")
	}

	expiresAt, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid expiry timestamp")
	}

	signature := parts[3]

	// Verify expiration
	if time.Now().Unix() > expiresAt {
		return map[string]interface{}{
			"success": false,
			"valid":   false,
			"error":   "token expired",
			"operation": string(SecurityOpValidateToken),
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Verify signature
	tokenData := fmt.Sprintf("%s|%d|%d", payload, issuedAt, expiresAt)
	expectedH := hmac.New(sha256.New, []byte(secretKey))
	expectedH.Write([]byte(tokenData))
	expectedSig := hex.EncodeToString(expectedH.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
		return map[string]interface{}{
			"success": false,
			"valid":   false,
			"error":   "invalid token signature",
			"operation": string(SecurityOpValidateToken),
			"timestamp": time.Now().Unix(),
		}, nil
	}

	return map[string]interface{}{
		"success":   true,
		"valid":     true,
		"payload":   payload,
		"issued_at": issuedAt,
		"expires_at": expiresAt,
		"operation": string(SecurityOpValidateToken),
		"timestamp": time.Now().Unix(),
	}, nil
}

// createKey creates a key of specified length from input
func createKey(key string, length int) []byte {
	result := make([]byte, length)
	copy(result, []byte(key))

	// If key is shorter than required length, extend it
	if len(key) < length {
		for i := len(key); i < length; i++ {
			result[i] = 0 // Fill with zeros
		}
	}

	return result
}

// getStringValue safely gets a string value from interface{}
func getStringValue(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// getBoolValue safely gets a boolean value from interface{}
func getBoolValue(v interface{}) bool {
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	if s, ok := v.(string); ok {
		b, _ := strconv.ParseBool(s)
		return b
	}
	if f, ok := v.(float64); ok {
		return f != 0
	}
	return false
}

// NewSecurityNodeFromConfig creates a new security node from a configuration map
func NewSecurityNodeFromConfig(config map[string]interface{}) (interfaces.NodeInstance, error) {
	var operation SecurityOperationType
	if op, exists := config["operation"]; exists {
		if opStr, ok := op.(string); ok {
			operation = SecurityOperationType(opStr)
		}
	}

	var algorithm SecurityAlgorithm
	if algo, exists := config["algorithm"]; exists {
		if algoStr, ok := algo.(string); ok {
			algorithm = SecurityAlgorithm(algoStr)
		}
	}

	var secretKey string
	if key, exists := config["secret_key"]; exists {
		if keyStr, ok := key.(string); ok {
			secretKey = keyStr
		}
	}

	var publicKey string
	if key, exists := config["public_key"]; exists {
		if keyStr, ok := key.(string); ok {
			publicKey = keyStr
		}
	}

	var privateKey string
	if key, exists := config["private_key"]; exists {
		if keyStr, ok := key.(string); ok {
			privateKey = keyStr
		}
	}

	var iterations float64
	if iter, exists := config["iterations"]; exists {
		if iterFloat, ok := iter.(float64); ok {
			iterations = iterFloat
		}
	}

	var keyLength float64
	if length, exists := config["key_length"]; exists {
		if lengthFloat, ok := length.(float64); ok {
			keyLength = lengthFloat
		}
	}

	var salt string
	if s, exists := config["salt"]; exists {
		if sStr, ok := s.(string); ok {
			salt = sStr
		}
	}

	var includeSalt bool
	if inc, exists := config["include_salt"]; exists {
		if incBool, ok := inc.(bool); ok {
			includeSalt = incBool
		}
	}

	var tokenExpiry float64
	if exp, exists := config["token_expiry_seconds"]; exists {
		if expFloat, ok := exp.(float64); ok {
			tokenExpiry = expFloat
		}
	}

	var validateRules []ValidationRule
	if rules, exists := config["validate_rules"]; exists {
		if rulesSlice, ok := rules.([]interface{}); ok {
			validateRules = make([]ValidationRule, len(rulesSlice))
			for i, rule := range rulesSlice {
				if ruleMap, ok := rule.(map[string]interface{}); ok {
					var minVal, maxVal interface{}
					if min, exists := ruleMap["min"]; exists {
						minVal = min
					}
					if max, exists := ruleMap["max"]; exists {
						maxVal = max
					}

					validateRules[i] = ValidationRule{
						Type:     getStringValue(ruleMap["type"]),
						Pattern:  getStringValue(ruleMap["pattern"]),
						Min:      minVal,
						Max:      maxVal,
						Required: getBoolValue(ruleMap["required"]),
						Message:  getStringValue(ruleMap["message"]),
					}
				}
			}
		}
	}

	var maskPattern string
	if pattern, exists := config["mask_pattern"]; exists {
		if patternStr, ok := pattern.(string); ok {
			maskPattern = patternStr
		}
	}

	var maskCustomPattern string
	if pattern, exists := config["mask_custom_pattern"]; exists {
		if patternStr, ok := pattern.(string); ok {
			maskCustomPattern = patternStr
		}
	}

	nodeConfig := &SecurityNodeConfig{
		Operation:       operation,
		Algorithm:       algorithm,
		SecretKey:       secretKey,
		PublicKey:       publicKey,
		PrivateKey:      privateKey,
		Iterations:      int(iterations),
		KeyLength:       int(keyLength),
		Salt:            salt,
		IncludeSalt:     includeSalt,
		TokenExpiry:     time.Duration(tokenExpiry) * time.Second,
		ValidateRules:   validateRules,
		MaskPattern:     maskPattern,
		MaskCustomPattern: maskCustomPattern,
	}

	return NewSecurityNode(nodeConfig), nil
}

// RegisterSecurityNode registers the security node type with the engine
func RegisterSecurityNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("security_operation", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return NewSecurityNodeFromConfig(config)
	})
}