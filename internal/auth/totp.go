package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	// TOTP configuration
	TOTPPeriod    = 30 // 30 seconds period
	TOTPDigits    = 6  // 6 digit codes
	TOTPSkewSteps = 1  // Allow 1 time step before/after for clock skew
	SecretLength  = 20 // 20 bytes = 160 bits for secret
)

// TOTPConfig holds TOTP configuration
type TOTPConfig struct {
	Period    int
	Digits    int
	SkewSteps int
}

// DefaultTOTPConfig returns the default TOTP configuration
func DefaultTOTPConfig() *TOTPConfig {
	return &TOTPConfig{
		Period:    TOTPPeriod,
		Digits:    TOTPDigits,
		SkewSteps: TOTPSkewSteps,
	}
}

// GenerateTOTPSecret generates a random secret for TOTP
func GenerateTOTPSecret() (string, error) {
	secret := make([]byte, SecretLength)
	_, err := rand.Read(secret)
	if err != nil {
		return "", fmt.Errorf("failed to generate random secret: %w", err)
	}

	// Encode to base32 without padding
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret)
	return encoded, nil
}

// GenerateTOTPCode generates a TOTP code for the given secret at the current time
func GenerateTOTPCode(secret string, config *TOTPConfig) (string, error) {
	if config == nil {
		config = DefaultTOTPConfig()
	}

	counter := time.Now().Unix() / int64(config.Period)
	return generateHOTP(secret, counter, config.Digits)
}

// ValidateTOTPCode validates a TOTP code against the secret
// It allows for clock skew by checking adjacent time steps
func ValidateTOTPCode(secret, code string, config *TOTPConfig) (bool, error) {
	if config == nil {
		config = DefaultTOTPConfig()
	}

	// Clean the code (remove spaces, dashes)
	code = strings.ReplaceAll(code, " ", "")
	code = strings.ReplaceAll(code, "-", "")

	if len(code) != config.Digits {
		return false, nil
	}

	currentCounter := time.Now().Unix() / int64(config.Period)

	// Check current time and allow for clock skew
	for i := -config.SkewSteps; i <= config.SkewSteps; i++ {
		counter := currentCounter + int64(i)
		expectedCode, err := generateHOTP(secret, counter, config.Digits)
		if err != nil {
			return false, err
		}

		if code == expectedCode {
			return true, nil
		}
	}

	return false, nil
}

// generateHOTP generates an HMAC-based One-Time Password
func generateHOTP(secret string, counter int64, digits int) (string, error) {
	// Decode base32 secret
	secret = strings.ToUpper(secret)
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return "", fmt.Errorf("invalid secret: %w", err)
	}

	// Convert counter to bytes (big-endian)
	counterBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBytes, uint64(counter))

	// Calculate HMAC-SHA1
	h := hmac.New(sha1.New, key)
	h.Write(counterBytes)
	hash := h.Sum(nil)

	// Dynamic truncation
	offset := hash[len(hash)-1] & 0x0f
	truncatedHash := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff

	// Generate code
	code := truncatedHash % uint32(math.Pow10(digits))

	// Format with leading zeros
	format := fmt.Sprintf("%%0%dd", digits)
	return fmt.Sprintf(format, code), nil
}

// GenerateProvisioningURI generates a provisioning URI for QR code generation
// Format: otpauth://totp/ISSUER:ACCOUNT?secret=SECRET&issuer=ISSUER&algorithm=SHA1&digits=6&period=30
func GenerateProvisioningURI(secret, issuer, accountName string, config *TOTPConfig) string {
	if config == nil {
		config = DefaultTOTPConfig()
	}

	return fmt.Sprintf(
		"otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=%d&period=%d",
		issuer,
		accountName,
		secret,
		issuer,
		config.Digits,
		config.Period,
	)
}

// GenerateRecoveryCodes generates a set of single-use recovery codes
func GenerateRecoveryCodes(count int) ([]string, error) {
	codes := make([]string, count)

	for i := 0; i < count; i++ {
		// Generate 8 random bytes
		b := make([]byte, 8)
		_, err := rand.Read(b)
		if err != nil {
			return nil, fmt.Errorf("failed to generate recovery code: %w", err)
		}

		// Format as XXXX-XXXX (8 hex characters with dash)
		code := fmt.Sprintf("%X-%X", b[:4], b[4:])
		codes[i] = code
	}

	return codes, nil
}

// ValidateRecoveryCode validates a recovery code
// This should be used with database lookup to ensure the code exists and hasn't been used
func ValidateRecoveryCode(inputCode, storedCode string) bool {
	// Clean both codes
	inputCode = strings.ToUpper(strings.TrimSpace(inputCode))
	storedCode = strings.ToUpper(strings.TrimSpace(storedCode))

	// Constant-time comparison to prevent timing attacks
	return hmac.Equal([]byte(inputCode), []byte(storedCode))
}
