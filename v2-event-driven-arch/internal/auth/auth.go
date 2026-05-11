package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func MakeJWT(userID uint, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "ticketing",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   strconv.FormatUint(uint64(userID), 10),
	})

	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func MakeRefreshToken() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key) // We don't need the return value
	if err != nil {
		return "", err
	}

	str := hex.EncodeToString(key) // Use the actual random bytes, not the count
	return str, nil
}

// SHA256Hex returns the hex-encoded SHA-256 hash of s.
// Used to store temp token fingerprints without keeping the raw token.
func SHA256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func GetAPIKey(headers http.Header) (string, error) {
	polkaStr := headers.Get("Authorization")
	if polkaStr == "" {
		return "", fmt.Errorf("the authorization header cannot be found")
	}

	if !strings.HasPrefix(polkaStr, "ApiKey ") {
		return "", fmt.Errorf("does not contain bearer token")
	}

	key := strings.TrimPrefix(polkaStr, "ApiKey ")

	return strings.TrimSpace(key), nil
}
