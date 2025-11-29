package middleware

import (
	"net/http"
	"ticketing_system/internal/models"

	"gorm.io/gorm"
)

// Require2FA is a middleware that ensures the user has 2FA enabled
// Use this for high-security endpoints that should only be accessible with 2FA
func Require2FA(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			// Get user ID from token (assumes AuthMiddleware has already run)
			userID := GetUserIDFromToken(r)
			if userID == 0 {
				WriteJSONError(w, http.StatusUnauthorized, "authentication required")
				return
			}

			// Check if user has 2FA enabled
			var twoFactorAuth models.TwoFactorAuth
			err := db.Where("user_id = ? AND enabled = ?", userID, true).First(&twoFactorAuth).Error
			if err != nil {
				WriteJSONError(w, http.StatusForbidden, "two-factor authentication is required for this action")
				return
			}

			// 2FA is enabled, allow request to continue
			next.ServeHTTP(w, r)
		})
	}
}

// Recommend2FA is a middleware that checks if user should enable 2FA
// It adds a header suggesting 2FA but doesn't block the request
func Recommend2FA(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user ID from token
			userID := GetUserIDFromToken(r)
			if userID != 0 {
				// Check if user has 2FA enabled
				var twoFactorAuth models.TwoFactorAuth
				err := db.Where("user_id = ? AND enabled = ?", userID, true).First(&twoFactorAuth).Error
				if err != nil {
					// User doesn't have 2FA - add recommendation header
					w.Header().Set("X-2FA-Recommended", "true")
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireOrganizerWith2FA ensures the user is an organizer and has 2FA enabled
// This is useful for high-value organizer operations
func RequireOrganizerWith2FA(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			userID := GetUserIDFromToken(r)
			if userID == 0 {
				WriteJSONError(w, http.StatusUnauthorized, "authentication required")
				return
			}

			// Get user and check role
			var user models.User
			if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
				WriteJSONError(w, http.StatusNotFound, "user not found")
				return
			}

			// Check if user is an organizer
			if user.Role != models.RoleOrganizer && user.Role != models.RoleAdmin {
				WriteJSONError(w, http.StatusForbidden, "organizer access required")
				return
			}

			// Check if 2FA is enabled
			var twoFactorAuth models.TwoFactorAuth
			err := db.Where("user_id = ? AND enabled = ?", userID, true).First(&twoFactorAuth).Error
			if err != nil {
				WriteJSONError(w, http.StatusForbidden, "two-factor authentication is required for organizer operations. Please enable 2FA in your security settings")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
