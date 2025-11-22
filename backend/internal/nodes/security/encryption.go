// backend/internal/nodes/security/encryption.go
package security

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
	"github.com/citadel-agent/backend/internal/workflow/core/engine"
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
}

// EncryptionNode represents an encryption node
type EncryptionNode struct {
	config *EncryptionConfig
}

// NewEncryptionNode creates a new encryption node
func NewEncryptionNode(config *EncryptionConfig) *EncryptionNode {
	return &EncryptionNode{
		config: config,
	}
}

// Execute executes the encryption operation
func (en *EncryptionNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	mode := en.config.Mode
	if m, exists := inputs["mode"]; exists {
		if mStr, ok := m.(string); ok {
			mode = EncryptionMode(mStr)
		}
	}

	algorithm := en.config.Algorithm
	if algo, exists := inputs["algorithm"]; exists {
		if algoStr, ok := algo.(string); ok {
			algorithm = EncryptionAlgorithm(algoStr)
		}
	}

	key := en.config.Key
	if k, exists := inputs["key"]; exists {
		if kStr, ok := k.(string); ok {
			key = kStr
		}
	}

	data := ""
	if d, exists := inputs["data"]; exists {
		if dStr, ok := d.(string); ok {
			data = dStr
		}
	}

	switch mode {
	case EncryptionModeEncrypt:
		result, err := en.encryptData(data, key, algorithm)
		if err != nil {
			return nil, fmt.Errorf("encryption failed: %w", err)
		}
		return result, nil
	case EncryptionModeDecrypt:
		result, err := en.decryptData(data, key, algorithm)
		if err != nil {
			return nil, fmt.Errorf("decryption failed: %w", err)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported encryption mode: %s", mode)
	}
}

// encryptData encrypts data using the specified algorithm
func (en *EncryptionNode) encryptData(plaintext, key string, algorithm EncryptionAlgorithm) (map[string]interface{}, error) {
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
		return nil, fmt.Errorf("unsupported encryption algorithm: %s", algorithm)
	}

	// Pad or truncate key to required length
	keyBytes = make([]byte, keySize)
	copy(keyBytes, []byte(key))
	if len(key) > keySize {
		keyBytes = keyBytes[:keySize]
	}

	// Create cipher
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Generate a random IV
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
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

	return map[string]interface{}{
		"success":        true,
		"encrypted_data": base64.StdEncoding.EncodeToString(result),
		"iv":            base64.StdEncoding.EncodeToString(iv),
		"algorithm":     string(algorithm),
		"operation":     "encrypt",
		"input_size":    len(plaintext),
		"output_size":   len(result),
		"timestamp":     time.Now().Unix(),
	}, nil
}

// decryptData decrypts data using the specified algorithm
func (en *EncryptionNode) decryptData(encryptedDataBase64, key string, algorithm EncryptionAlgorithm) (map[string]interface{}, error) {
	var keySize int

	switch algorithm {
	case AlgorithmAES256:
		keySize = 32 // 256 bits
	case AlgorithmAES192:
		keySize = 24 // 192 bits
	case AlgorithmAES128:
		keySize = 16 // 128 bits
	default:
		return nil, fmt.Errorf("unsupported encryption algorithm: %s", algorithm)
	}

	// Decode the base64 data
	encryptedData, err := base64.StdEncoding.DecodeString(encryptedDataBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted data: %w", err)
	}

	// Check if data is long enough to contain IV + at least one block
	if len(encryptedData) < aes.BlockSize+aes.BlockSize {
		return nil, fmt.Errorf("encrypted data too short")
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
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create cipher mode and decrypt
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove padding
	padding := int(plaintext[len(plaintext)-1])
	if padding > len(plaintext) || padding <= 0 {
		return nil, fmt.Errorf("invalid padding")
	}
	plaintext = plaintext[:len(plaintext)-padding]

	return map[string]interface{}{
		"success":        true,
		"decrypted_data": string(plaintext),
		"algorithm":      string(algorithm),
		"operation":      "decrypt",
		"input_size":     len(encryptedData),
		"output_size":    len(plaintext),
		"timestamp":      time.Now().Unix(),
	}, nil
}

// EncryptionNodeFromConfig creates a new encryption node from a configuration map
func EncryptionNodeFromConfig(config map[string]interface{}) (interfaces.NodeInstance, error) {
	var mode EncryptionMode
	if m, exists := config["mode"]; exists {
		if mStr, ok := m.(string); ok {
			mode = EncryptionMode(mStr)
		}
	}

	var algorithm EncryptionAlgorithm
	if algo, exists := config["algorithm"]; exists {
		if algoStr, ok := algo.(string); ok {
			algorithm = EncryptionAlgorithm(algoStr)
		}
	}

	var key string
	if k, exists := config["key"]; exists {
		if kStr, ok := k.(string); ok {
			key = kStr
		}
	}

	nodeConfig := &EncryptionConfig{
		Mode:      mode,
		Algorithm: algorithm,
		Key:       key,
	}

	return NewEncryptionNode(nodeConfig), nil
}

// RegisterEncryptionNode registers the encryption node type with the engine
func RegisterEncryptionNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("encryption", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return EncryptionNodeFromConfig(config)
	})
}