// backend/internal/auth/encryption.go
package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// EncryptionService provides encryption and decryption capabilities
type EncryptionService struct {
	key []byte
}

// NewEncryptionService creates a new encryption service with the provided key
func NewEncryptionService(key string) (*EncryptionService, error) {
	if len(key) == 0 {
		return nil, errors.New("encryption key cannot be empty")
	}

	// Ensure key is 32 bytes for AES-256
	keyBytes := []byte(key)
	if len(keyBytes) < 32 {
		// Pad the key if it's too short
		paddedKey := make([]byte, 32)
		copy(paddedKey, keyBytes)
		keyBytes = paddedKey
	} else if len(keyBytes) > 32 {
		// Truncate the key if it's too long
		keyBytes = keyBytes[:32]
	}

	return &EncryptionService{
		key: keyBytes,
	}, nil
}

// Encrypt encrypts the provided data and returns a base64 encoded string
func (es *EncryptionService) Encrypt(data []byte) (string, error) {
	block, err := aes.NewCipher(es.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts the provided base64 encoded data
func (es *EncryptionService) Decrypt(encryptedData string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(es.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// EncryptString encrypts a string and returns base64 encoded result
func (es *EncryptionService) EncryptString(data string) (string, error) {
	return es.Encrypt([]byte(data))
}

// DecryptString decrypts a base64 encoded string
func (es *EncryptionService) DecryptString(encryptedData string) (string, error) {
	data, err := es.Decrypt(encryptedData)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// EncryptWorkflowData encrypts sensitive workflow data
func (es *EncryptionService) EncryptWorkflowData(data map[string]interface{}) (map[string]interface{}, error) {
	encryptedData := make(map[string]interface{})
	
	for key, value := range data {
		switch v := value.(type) {
		case string:
			// Encrypt string values that might contain sensitive data
			if es.isSensitiveKey(key) {
				encryptedValue, err := es.EncryptString(v)
				if err != nil {
					return nil, err
				}
				encryptedData[key] = encryptedValue
			} else {
				encryptedData[key] = v
			}
		case map[string]interface{}:
			// Recursively encrypt nested maps
			encryptedNested, err := es.EncryptWorkflowData(v)
			if err != nil {
				return nil, err
			}
			encryptedData[key] = encryptedNested
		default:
			// For other types, add as is
			encryptedData[key] = v
		}
	}
	
	return encryptedData, nil
}

// DecryptWorkflowData decrypts sensitive workflow data
func (es *EncryptionService) DecryptWorkflowData(data map[string]interface{}) (map[string]interface{}, error) {
	decryptedData := make(map[string]interface{})
	
	for key, value := range data {
		switch v := value.(type) {
		case string:
			// Try to decrypt if it looks like encrypted data
			if es.mightBeEncrypted(v) {
				decryptedValue, err := es.DecryptString(v)
				if err != nil {
					// If decryption fails, keep the original value
					decryptedData[key] = v
				} else {
					decryptedData[key] = decryptedValue
				}
			} else {
				decryptedData[key] = v
			}
		case map[string]interface{}:
			// Recursively decrypt nested maps
			decryptedNested, err := es.DecryptWorkflowData(v)
			if err != nil {
				return nil, err
			}
			decryptedData[key] = decryptedNested
		default:
			// For other types, add as is
			decryptedData[key] = v
		}
	}
	
	return decryptedData, nil
}

// isSensitiveKey checks if a key might contain sensitive data
func (es *EncryptionService) isSensitiveKey(key string) bool {
	sensitiveKeys := []string{
		"password", "secret", "token", "key", "api_key", "access_token",
		"refresh_token", "auth", "credential", "credentials", "private", 
		"oauth", "client_secret", "webhook", "hook", "cert", "certificate",
	}
	
	keyLower := lowercase(key)
	for _, sensitiveKey := range sensitiveKeys {
		if contains(keyLower, sensitiveKey) {
			return true
		}
	}
	return false
}

// mightBeEncrypted checks if a string looks like encrypted data
func (es *EncryptionService) mightBeEncrypted(data string) bool {
	// Check if it's a base64 encoded string (common in encrypted data)
	_, err := base64.StdEncoding.DecodeString(data)
	return err == nil
}

// Helper functions
func lowercase(s string) string {
	// Simple lowercase implementation (in production, use strings.ToLower)
	result := make([]byte, len(s))
	for i, c := range []byte(s) {
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}