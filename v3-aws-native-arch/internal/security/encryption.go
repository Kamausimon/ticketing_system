package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

var (
	// ErrInvalidKey indicates the encryption key is invalid
	ErrInvalidKey = errors.New("encryption key must be 16, 24, or 32 bytes")
	// ErrInvalidCiphertext indicates the ciphertext is invalid
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
)

// EncryptionService handles encryption and decryption of sensitive data
type EncryptionService struct {
	key []byte
}

// NewEncryptionService creates a new encryption service with the given key
// Key must be 16, 24, or 32 bytes for AES-128, AES-192, or AES-256
func NewEncryptionService(key string) (*EncryptionService, error) {
	keyBytes := []byte(key)

	// Validate key length
	keyLen := len(keyBytes)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		return nil, ErrInvalidKey
	}

	return &EncryptionService{
		key: keyBytes,
	}, nil
}

// Encrypt encrypts plaintext using AES-GCM and returns base64-encoded ciphertext
// Returns empty string if plaintext is empty
func (s *EncryptionService) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	// Create AES cipher
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt and prepend nonce
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Encode to base64 for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64-encoded ciphertext and returns plaintext
// Returns empty string if ciphertext is empty
func (s *EncryptionService) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	// Create AES cipher
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Validate ciphertext length
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	// Extract nonce and ciphertext
	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// EncryptBankDetails encrypts sensitive bank account information
func (s *EncryptionService) EncryptBankDetails(accountNumber, bankCode string) (encryptedNumber, encryptedCode string, err error) {
	encryptedNumber, err = s.Encrypt(accountNumber)
	if err != nil {
		return "", "", err
	}

	encryptedCode, err = s.Encrypt(bankCode)
	if err != nil {
		return "", "", err
	}

	return encryptedNumber, encryptedCode, nil
}

// DecryptBankDetails decrypts sensitive bank account information
func (s *EncryptionService) DecryptBankDetails(encryptedNumber, encryptedCode string) (accountNumber, bankCode string, err error) {
	accountNumber, err = s.Decrypt(encryptedNumber)
	if err != nil {
		return "", "", err
	}

	bankCode, err = s.Decrypt(encryptedCode)
	if err != nil {
		return "", "", err
	}

	return accountNumber, bankCode, nil
}

// MaskSensitiveData masks sensitive data for display purposes
// Shows only the last 4 characters
func MaskSensitiveData(data string) string {
	if data == "" {
		return ""
	}

	length := len(data)
	if length <= 4 {
		return "****"
	}

	return "****" + data[length-4:]
}

// MaskBankAccountNumber masks bank account number for display
func MaskBankAccountNumber(accountNumber string) string {
	return MaskSensitiveData(accountNumber)
}
