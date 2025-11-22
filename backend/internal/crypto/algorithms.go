package crypto

// EncryptionAlgorithm represents the encryption algorithm
type EncryptionAlgorithm string

const (
	AES256 EncryptionAlgorithm = "aes256"
	AES192 EncryptionAlgorithm = "aes192"
	AES128 EncryptionAlgorithm = "aes128"
)

// SecurityAlgorithm represents the security algorithm
type SecurityAlgorithm string

const (
	AlgorithmSHA256 SecurityAlgorithm = "sha256"
	AlgorithmSHA512 SecurityAlgorithm = "sha512"
	AlgorithmAES256 SecurityAlgorithm = "aes256"
	AlgorithmHMACSHA256 SecurityAlgorithm = "hmac_sha256"
	AlgorithmBCrypt SecurityAlgorithm = "bcrypt"
	AlgorithmScrypt SecurityAlgorithm = "scrypt"
	AlgorithmRSA SecurityAlgorithm = "rsa"
	AlgorithmECDSA SecurityAlgorithm = "ecdsa"
)