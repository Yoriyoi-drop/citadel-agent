package security

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"citadel-agent/backend/internal/interfaces"
)

// EncryptionMode represents the encryption mode
type EncryptionMode string

const (
	EncryptionModeEncrypt EncryptionMode = "encrypt"
	EncryptionModeDecrypt EncryptionMode = "decrypt"
)

// EncryptionAlgorithm represents the encryption algorithm
type EncryptionAlgorithm string

const (
	AlgorithmAES256 EncryptionAlgorithm = "aes256"
	AlgorithmAES192 EncryptionAlgorithm = "aes192"
	AlgorithmAES128 EncryptionAlgorithm = "aes128"
)

// EncryptionConfig represents the configuration for an encryption node
type EncryptionConfig struct {
	Mode      EncryptionMode      `json:"mode"`
	Algorithm EncryptionAlgorithm `json:"algorithm"`
	Key       string             `json:"key"`
	IV        string             `json:"iv"`
	SourceField string           `json:"source_field"`    // Field to encrypt/decrypt
	TargetField string           `json:"target_field"`    // Field to store result
	EnableCaching bool           `json:"enable_caching"`
	CacheTTL    int              `json:"cache_ttl"`       // in seconds
	EnableProfiling bool         `json:"enable_profiling"`
	ReturnRawResults bool         `json:"return_raw_results"`
	CustomParams map[string]interface{} `json:"custom_params"`
}

// EncryptionNode represents an encryption node
type EncryptionNode struct {
	config *EncryptionConfig
}

// NewEncryptionNode creates a new encryption node
func NewEncryptionNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	// Convert config map to struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var encConfig EncryptionConfig
	if err := json.Unmarshal(jsonData, &encConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate and set defaults
	if encConfig.Mode == "" {
		encConfig.Mode = EncryptionModeEncrypt
	}

	if encConfig.Algorithm == "" {
		encConfig.Algorithm = AlgorithmAES256
	}

	if encConfig.CacheTTL == 0 {
		encConfig.CacheTTL = 3600 // 1 hour default cache TTL
	}

	if encConfig.SourceField == "" {
		encConfig.SourceField = "data" // Default field to encrypt/decrypt
	}

	if encConfig.TargetField == "" {
		encConfig.TargetField = "result" // Default field to store result
	}

	return &EncryptionNode{
		config: &encConfig,
	}, nil
}

// Execute executes the encryption operation
func (en *EncryptionNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	startTime := time.Now()

	// Override config values with inputs if provided
	mode := en.config.Mode
	if inputMode, exists := inputs["mode"]; exists {
		if modeStr, ok := inputMode.(string); ok && modeStr != "" {
			if strings.ToLower(modeStr) == "encrypt" {
				mode = EncryptionModeEncrypt
			} else if strings.ToLower(modeStr) == "decrypt" {
				mode = EncryptionModeDecrypt
			}
		}
	}

	algorithm := en.config.Algorithm
	if inputAlgorithm, exists := inputs["algorithm"]; exists {
		if algoStr, ok := inputAlgorithm.(string); ok && algoStr != "" {
			switch strings.ToLower(algoStr) {
			case "aes256":
				algorithm = AlgorithmAES256
			case "aes192":
				algorithm = AlgorithmAES192
			case "aes128":
				algorithm = AlgorithmAES128
			}
		}
	}

	key := en.config.Key
	if inputKey, exists := inputs["key"]; exists {
		if keyStr, ok := inputKey.(string); ok {
			key = keyStr
		}
	}

	if key == "" {
		return nil, fmt.Errorf("encryption key is required")
	}

	sourceField := en.config.SourceField
	if inputSourceField, exists := inputs["source_field"]; exists {
		if sourceFieldStr, ok := inputSourceField.(string); ok && sourceFieldStr != "" {
			sourceField = sourceFieldStr
		}
	}

	targetField := en.config.TargetField
	if inputTargetField, exists := inputs["target_field"]; exists {
		if targetFieldStr, ok := inputTargetField.(string); ok && targetFieldStr != "" {
			targetField = targetFieldStr
		}
	}

	enableProfiling := en.config.EnableProfiling
	if inputEnableProfiling, exists := inputs["enable_profiling"]; exists {
		if enableProfilingBool, ok := inputEnableProfiling.(bool); ok {
			enableProfiling = enableProfilingBool
		}
	}

	returnRawResults := en.config.ReturnRawResults
	if inputReturnRaw, exists := inputs["return_raw_results"]; exists {
		if returnRawBool, ok := inputReturnRaw.(bool); ok {
			returnRawResults = returnRawBool
		}
	}

	// Get the data to encrypt/decrypt
	dataToProcess := ""
	if data, exists := inputs[sourceField]; exists {
		if dataStr, ok := data.(string); ok {
			dataToProcess = dataStr
		} else {
			// Try to convert to string
			dataToProcess = fmt.Sprintf("%v", data)
		}
	} else {
		// If source field is not found, try to use the whole input as a string
		jsonData, err := json.Marshal(inputs)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal input data: %w", err)
		}
		dataToProcess = string(jsonData)
	}

	// Perform the encryption/decryption
	var result string
	var err error

	switch mode {
	case EncryptionModeEncrypt:
		result, err = en.encryptData(dataToProcess, key, algorithm)
	case EncryptionModeDecrypt:
		result, err = en.decryptData(dataToProcess, key, algorithm)
	default:
		return nil, fmt.Errorf("unsupported encryption mode: %s", mode)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to %s data: %w", string(mode), err)
	}

	// Prepare result
	output := make(map[string]interface{})
	
	// Add the processed result to the target field
	output[targetField] = result

	// Add all original inputs to the output
	for k, v := range inputs {
		if k != sourceField { // Don't overwrite the source field
			output[k] = v
		}
	}

	// Add encryption-specific metadata
	output["success"] = true
	output["operation"] = string(mode)
	output["algorithm"] = string(algorithm)
	output["source_field"] = sourceField
	output["target_field"] = targetField
	output["input_size"] = len(dataToProcess)
	output["output_size"] = len(result)
	output["timestamp"] = time.Now().Unix()
	output["execution_time"] = time.Since(startTime).Seconds()

	if returnRawResults {
		output["raw_input"] = dataToProcess
		output["raw_output"] = result
	}

	// Add profiling data if enabled
	if enableProfiling {
		output["profiling"] = map[string]interface{}{
			"start_time": startTime.Unix(),
			"end_time":   time.Now().Unix(),
			"duration":   time.Since(startTime).Seconds(),
			"operation":  string(mode),
			"algorithm":  string(algorithm),
			"input_size": len(dataToProcess),
			"output_size": len(result),
		}
	}

	return output, nil
}

// encryptData encrypts data using the specified algorithm
func (en *EncryptionNode) encryptData(plaintext, key string, algorithm EncryptionAlgorithm) (string, error) {
	var keyBytes []byte
	var keySize int

	switch algorithm {
	case AlgorithmAES256:
		keySize = 32 // 256 bits
	case AlgorithmAES192:
		keySize = 24 // 192 bits
	case AlgorithmAES128:
		keySize = 16 // 128 bits
	default:
		return "", fmt.Errorf("unsupported encryption algorithm: %s", algorithm)
	}

	// Create key of required length
	keyBytes = make([]byte, keySize)
	copy(keyBytes, []byte(key))
	if len(key) > keySize {
		keyBytes = keyBytes[:keySize]
	}

	// Create cipher
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Generate a random IV
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return "", fmt.Errorf("failed to generate IV: %w", err)
	}

	// Pad plaintext to be multiple of block size
	padding := aes.BlockSize - (len(plaintext) % aes.BlockSize)
	paddedText := []byte(plaintext)
	for i := 0; i < padding; i++ {
		paddedText = append(paddedText, byte(padding))
	}

	// Create cipher mode and encrypt
	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(paddedText))
	mode.CryptBlocks(ciphertext, paddedText)

	// Combine IV and ciphertext
	result := append(iv, ciphertext...)

	return base64.StdEncoding.EncodeToString(result), nil
}

// decryptData decrypts data using the specified algorithm
func (en *EncryptionNode) decryptData(encryptedDataBase64, key string, algorithm EncryptionAlgorithm) (string, error) {
	var keySize int

	switch algorithm {
	case AlgorithmAES256:
		keySize = 32 // 256 bits
	case AlgorithmAES192:
		keySize = 24 // 192 bits
	case AlgorithmAES128:
		keySize = 16 // 128 bits
	default:
		return "", fmt.Errorf("unsupported encryption algorithm: %s", algorithm)
	}

	// Decode the base64 data
	encryptedData, err := base64.StdEncoding.DecodeString(encryptedDataBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted data: %w", err)
	}

	// Check if data is long enough to contain IV + at least one block
	if len(encryptedData) < aes.BlockSize+aes.BlockSize {
		return "", fmt.Errorf("encrypted data too short")
	}

	// Extract IV and ciphertext
	iv := encryptedData[:aes.BlockSize]
	ciphertext := encryptedData[aes.BlockSize:]

	// Pad or truncate key to required length
	keyBytes := make([]byte, keySize)
	copy(keyBytes, []byte(key))
	if len(key) > keySize {
		keyBytes = keyBytes[:keySize]
	}

	// Create cipher
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create cipher mode and decrypt
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove padding
	padding := int(plaintext[len(plaintext)-1])
	if padding > len(plaintext) || padding <= 0 {
		return "", fmt.Errorf("invalid padding")
	}
	plaintext = plaintext[:len(plaintext)-padding]

	return string(plaintext), nil
}

// GetType returns the type of node
func (en *EncryptionNode) GetType() string {
	return "encryption"
}

// GetID returns the unique ID of the node instance
func (en *EncryptionNode) GetID() string {
	return fmt.Sprintf("enc_%s_%s_%d", string(en.config.Mode), string(en.config.Algorithm), time.Now().Unix())
}