package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func WriteJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")

	if authHeader == "" {
		return "", fmt.Errorf("the authorization header cannot be found")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("does not contain bearer token")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	return strings.TrimSpace(token), nil
}

func ValidateJWT(tokenString, tokenSecret string) (uint, error) {

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		userID, err := strconv.ParseUint(claims.Subject, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid user id: %w", err)
		}
		return uint(userID), nil
	}

	return 0, fmt.Errorf("invalid token")
}

func GetUserIDFromToken(r *http.Request) uint {
	userID, _ := GetUserIDFromTokenWithError(r)
	return userID
}

func GetUserIDFromTokenWithError(r *http.Request) (uint, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return 0, fmt.Errorf("error reading env variables: %w", err)
	}
	tokenSecret := os.Getenv("JWTSECRET")
	if tokenSecret == "" {
		return 0, fmt.Errorf("JWTSECRET not configured")
	}
	
	bearerToken, err := GetBearerToken(r.Header)
	if err != nil {
		return 0, fmt.Errorf("error getting auth header: %w", err)
	}
	
	userID, err := ValidateJWT(bearerToken, tokenSecret)
	if err != nil {
		return 0, fmt.Errorf("error validating token: %w", err)
	}
	
	return userID, nil
}

// RequireEmailVerification middleware checks if the user's email is verified
// If not verified, it returns a 403 Forbidden error
func RequireEmailVerification(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := GetUserIDFromTokenWithError(r)
			if err != nil || userID == 0 {
				WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			// Check if user's email is verified
			var emailVerified bool
			result := db.Table("users").
				Select("email_verified").
				Where("id = ?", userID).
				Row().
				Scan(&emailVerified)

			if result != nil {
				WriteJSONError(w, http.StatusInternalServerError, "failed to check email verification status")
				return
			}

			if !emailVerified {
				WriteJSONError(w, http.StatusForbidden, "email verification required. Please verify your email address to perform this action")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
