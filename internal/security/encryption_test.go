package security

import (
	"testing"
)

func TestEncryptionService(t *testing.T) {
	// Test with 32-byte key (AES-256)
	key := "12345678901234567890123456789012" // Exactly 32 bytes
	service, err := NewEncryptionService(key)
	if err != nil {
		t.Fatalf("Failed to create encryption service: %v", err)
	}

	// Test encryption and decryption
	plaintext := "1234567890"
	encrypted, err := service.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Encrypted text should be different from plaintext
	if encrypted == plaintext {
		t.Error("Encrypted text should differ from plaintext")
	}

	// Encrypted text should be longer due to nonce and auth tag
	if len(encrypted) <= len(plaintext) {
		t.Error("Encrypted text should be longer than plaintext")
	}

	// Decrypt and verify
	decrypted, err := service.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Decrypted text doesn't match. Expected: %s, Got: %s", plaintext, decrypted)
	}
}

func TestEncryptionServiceInvalidKey(t *testing.T) {
	// Test with invalid key length
	invalidKey := "short"
	_, err := NewEncryptionService(invalidKey)
	if err == nil {
		t.Error("Expected error for invalid key length")
	}
	if err != ErrInvalidKey {
		t.Errorf("Expected ErrInvalidKey, got: %v", err)
	}
}

func TestEncryptDecryptBankDetails(t *testing.T) {
	key := "12345678901234567890123456789012" // Exactly 32 bytes
	service, _ := NewEncryptionService(key)

	accountNumber := "9876543210"
	bankCode := "TESTSWIFT123"

	// Encrypt
	encNumber, encCode, err := service.EncryptBankDetails(accountNumber, bankCode)
	if err != nil {
		t.Fatalf("Failed to encrypt bank details: %v", err)
	}

	// Verify encrypted
	if encNumber == accountNumber || encCode == bankCode {
		t.Error("Bank details should be encrypted")
	}

	// Decrypt
	decNumber, decCode, err := service.DecryptBankDetails(encNumber, encCode)
	if err != nil {
		t.Fatalf("Failed to decrypt bank details: %v", err)
	}

	// Verify decrypted matches original
	if decNumber != accountNumber {
		t.Errorf("Account number mismatch. Expected: %s, Got: %s", accountNumber, decNumber)
	}
	if decCode != bankCode {
		t.Errorf("Bank code mismatch. Expected: %s, Got: %s", bankCode, decCode)
	}
}

func TestEncryptEmptyString(t *testing.T) {
	key := "12345678901234567890123456789012" // Exactly 32 bytes
	service, _ := NewEncryptionService(key)

	// Test empty string
	encrypted, err := service.Encrypt("")
	if err != nil {
		t.Fatalf("Encryption of empty string failed: %v", err)
	}
	if encrypted != "" {
		t.Error("Encrypted empty string should be empty")
	}

	// Test decryption of empty string
	decrypted, err := service.Decrypt("")
	if err != nil {
		t.Fatalf("Decryption of empty string failed: %v", err)
	}
	if decrypted != "" {
		t.Error("Decrypted empty string should be empty")
	}
}

func TestMaskBankAccountNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1234567890", "****7890"},
		{"12345", "****2345"},
		{"123", "****"},
		{"", ""},
	}

	for _, test := range tests {
		result := MaskBankAccountNumber(test.input)
		if result != test.expected {
			t.Errorf("Masking failed for %s. Expected: %s, Got: %s",
				test.input, test.expected, result)
		}
	}
}

func TestEncryptionDeterminism(t *testing.T) {
	key := "12345678901234567890123456789012" // Exactly 32 bytes
	service, _ := NewEncryptionService(key)

	plaintext := "1234567890"

	// Encrypt same plaintext twice
	encrypted1, _ := service.Encrypt(plaintext)
	encrypted2, _ := service.Encrypt(plaintext)

	// They should be different due to random nonce
	if encrypted1 == encrypted2 {
		t.Error("Encrypting same plaintext twice should produce different ciphertext (random nonce)")
	}

	// But both should decrypt to same plaintext
	decrypted1, _ := service.Decrypt(encrypted1)
	decrypted2, _ := service.Decrypt(encrypted2)

	if decrypted1 != plaintext || decrypted2 != plaintext {
		t.Error("Both encryptions should decrypt to original plaintext")
	}
}

func TestDecryptInvalidData(t *testing.T) {
	key := "12345678901234567890123456789012" // Exactly 32 bytes
	service, _ := NewEncryptionService(key)

	// Test with invalid base64
	_, err := service.Decrypt("invalid-base64-!!!!")
	if err == nil {
		t.Error("Expected error when decrypting invalid base64")
	}

	// Test with valid base64 but invalid ciphertext
	_, err = service.Decrypt("YWJjZGVmZ2g=") // "abcdefgh" in base64
	if err == nil {
		t.Error("Expected error when decrypting invalid ciphertext")
	}
}

func TestDifferentKeysCannotDecrypt(t *testing.T) {
	key1 := "12345678901234567890123456789012" // 32 bytes
	key2 := "abcdefghijklmnopqrstuvwxyz123456" // 32 bytes

	service1, _ := NewEncryptionService(key1)
	service2, _ := NewEncryptionService(key2)

	plaintext := "secret-data"

	// Encrypt with first key
	encrypted, _ := service1.Encrypt(plaintext)

	// Try to decrypt with different key
	_, err := service2.Decrypt(encrypted)
	if err == nil {
		t.Error("Different key should not be able to decrypt data")
	}
}
