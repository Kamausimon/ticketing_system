package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ticketing_system/internal/accounts"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"ticketing_system/pkg/qrcode"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// TwoFactorHandler handles 2FA-related requests
type TwoFactorHandler struct {
	db             *gorm.DB
	config         *TOTPConfig
	issuer         string // App name for TOTP
	activityLogger *accounts.ActivityLogger
}

// NewTwoFactorHandler creates a new 2FA handler
func NewTwoFactorHandler(db *gorm.DB, issuer string) *TwoFactorHandler {
	return &TwoFactorHandler{
		db:             db,
		config:         DefaultTOTPConfig(),
		issuer:         issuer,
		activityLogger: accounts.NewActivityLogger(db),
	}
}

// SetupRequest represents the request to start 2FA setup
type SetupRequest struct {
	Password string `json:"password"` // Require password confirmation for security
}

// SetupResponse represents the response for 2FA setup
type SetupResponse struct {
	Secret      string   `json:"secret"`       // TOTP secret (show once)
	QRCodeData  string   `json:"qr_code_data"` // Base64 encoded QR code image
	QRCodeURL   string   `json:"qr_code_url"`  // otpauth:// URL
	BackupCodes []string `json:"backup_codes"` // Recovery codes (show once)
}

// VerifySetupRequest represents the request to verify and enable 2FA
type VerifySetupRequest struct {
	Code string `json:"code"` // TOTP code to verify
}

// VerifyLoginRequest represents the request to verify 2FA during login
type VerifyLoginRequest struct {
	Code           string `json:"code"`             // TOTP code or recovery code
	TrustDevice    bool   `json:"trust_device"`     // Remember this device (future enhancement)
	IsRecoveryCode bool   `json:"is_recovery_code"` // Indicates if using recovery code
}

// DisableRequest represents the request to disable 2FA
type DisableRequest struct {
	Password string `json:"password"` // Require password confirmation
	Code     string `json:"code"`     // Current TOTP code
}

// Setup2FA initiates 2FA setup for the authenticated user
func (h *TwoFactorHandler) Setup2FA(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request
	var req SetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get user and verify password
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "invalid password")
		return
	}

	// Check if 2FA already enabled
	var existing models.TwoFactorAuth
	err := h.db.Where("user_id = ? AND enabled = ?", userID, true).First(&existing).Error
	if err == nil {
		middleware.WriteJSONError(w, http.StatusConflict, "2FA is already enabled")
		return
	}

	// Generate new secret
	secret, err := GenerateTOTPSecret()
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to generate secret")
		return
	}

	// Generate recovery codes
	recoveryCodes, err := GenerateRecoveryCodes(10)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to generate recovery codes")
		return
	}

	// Create or update 2FA session (temporary until verified)
	session := models.TwoFactorSession{
		UserID:    userID,
		Secret:    secret,
		Verified:  false,
		ExpiresAt: time.Now().Add(15 * time.Minute),
		IPAddress: GetClientIP(r),
		UserAgent: r.UserAgent(),
	}

	// Delete any existing sessions
	h.db.Where("user_id = ?", userID).Delete(&models.TwoFactorSession{})

	if err := h.db.Create(&session).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to create setup session")
		return
	}

	// Generate provisioning URI for QR code
	accountName := user.Email
	provisioningURI := GenerateProvisioningURI(secret, h.issuer, accountName, h.config)

	// Generate QR code as base64 PNG
	qrCodeData, err := qrcode.GenerateQRCodeBase64(provisioningURI, 256)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to generate QR code")
		return
	}

	response := SetupResponse{
		Secret:      secret,
		QRCodeData:  qrCodeData,
		QRCodeURL:   provisioningURI,
		BackupCodes: recoveryCodes,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// VerifySetup verifies the TOTP code and enables 2FA
func (h *TwoFactorHandler) VerifySetup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request
	var req VerifySetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Code == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "verification code is required")
		return
	}

	// Get the setup session
	var session models.TwoFactorSession
	err := h.db.Where("user_id = ? AND verified = ? AND expires_at > ?",
		userID, false, time.Now()).First(&session).Error
	if err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "no active setup session found")
		return
	}

	// Validate the TOTP code
	valid, err := ValidateTOTPCode(session.Secret, req.Code, h.config)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to validate code")
		return
	}

	if !valid {
		// Log failed attempt
		h.logAttempt(userID, false, "invalid_code", r)
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid verification code")
		return
	}

	// Code is valid - enable 2FA
	now := time.Now()
	twoFactorAuth := models.TwoFactorAuth{
		UserID:     userID,
		Enabled:    true,
		Secret:     session.Secret,
		VerifiedAt: &now,
		Method:     "totp",
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Save 2FA config
	if err := tx.Create(&twoFactorAuth).Error; err != nil {
		tx.Rollback()
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to enable 2FA")
		return
	}

	// Mark session as verified and delete it
	if err := tx.Delete(&session).Error; err != nil {
		tx.Rollback()
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to complete setup")
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to commit changes")
		return
	}

	// Log successful setup
	h.logAttempt(userID, true, "setup_completed", r)

	// Get user for activity logging
	var user models.User
	if err := h.db.First(&user, userID).Error; err == nil {
		if h.activityLogger != nil {
			h.activityLogger.Log2FAEnabled(user.AccountID, &user.ID, r.RemoteAddr, r.UserAgent())
		}
	}

	response := map[string]interface{}{
		"message": "Two-factor authentication has been successfully enabled",
		"enabled": true,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// VerifyLogin verifies 2FA code during login and issues full access token
func (h *TwoFactorHandler) VerifyLogin(w http.ResponseWriter, r *http.Request, tokenSecret string) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request
	var req VerifyLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Code == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "verification code is required")
		return
	}

	// Get user details
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get 2FA config
	var twoFactorAuth models.TwoFactorAuth
	err := h.db.Where("user_id = ? AND enabled = ?", userID, true).First(&twoFactorAuth).Error
	if err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "2FA not enabled for this user")
		return
	}

	var valid bool

	if req.IsRecoveryCode {
		// Validate recovery code
		valid = h.validateRecoveryCode(twoFactorAuth.ID, req.Code, r)
	} else {
		// Validate TOTP code
		valid, err = ValidateTOTPCode(twoFactorAuth.Secret, req.Code, h.config)
		if err != nil {
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to validate code")
			return
		}
	}

	if !valid {
		h.logAttempt(userID, false, "invalid_code", r)
		middleware.WriteJSONError(w, http.StatusUnauthorized, "invalid verification code")
		return
	}

	// Update last used time
	now := time.Now()
	twoFactorAuth.LastUsedAt = &now
	h.db.Save(&twoFactorAuth)

	// Log successful verification
	h.logAttempt(userID, true, "login_verified", r)

	// Log 2FA verification activity
	if h.activityLogger != nil {
		h.activityLogger.Log2FAVerified(user.AccountID, &user.ID, r.RemoteAddr, r.UserAgent())
	}

	// Generate full access token
	token, err := MakeJWT(userID, tokenSecret, time.Hour)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "error generating access token")
		return
	}

	// Generate refresh token
	refreshToken, err := MakeRefreshToken()
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "error creating refresh token")
		return
	}

	refreshExpiresAt := time.Now().Add(60 * 24 * time.Hour).Unix()

	// Update user session
	if err = h.db.Model(&user).Updates(map[string]interface{}{
		"refresh_token":     &refreshToken,
		"refresh_token_exp": &refreshExpiresAt,
		"last_login_at":     time.Now().Unix(),
	}).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update user session")
		return
	}

	response := map[string]interface{}{
		"message":  "Two-factor authentication verified successfully",
		"verified": true,
		"token":    token,
		"user_id":  user.ID,
		"role":     user.Role,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Disable2FA disables 2FA for the authenticated user
func (h *TwoFactorHandler) Disable2FA(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request
	var req DisableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get user and verify password
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "invalid password")
		return
	}

	// Get 2FA config
	var twoFactorAuth models.TwoFactorAuth
	err := h.db.Where("user_id = ? AND enabled = ?", userID, true).First(&twoFactorAuth).Error
	if err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "2FA is not enabled")
		return
	}

	// Verify current TOTP code
	valid, err := ValidateTOTPCode(twoFactorAuth.Secret, req.Code, h.config)
	if err != nil || !valid {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "invalid verification code")
		return
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete all recovery codes
	if err := tx.Where("two_factor_auth_id = ?", twoFactorAuth.ID).Delete(&models.RecoveryCode{}).Error; err != nil {
		tx.Rollback()
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to remove recovery codes")
		return
	}

	// Delete 2FA config
	if err := tx.Delete(&twoFactorAuth).Error; err != nil {
		tx.Rollback()
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to disable 2FA")
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to commit changes")
		return
	}

	// Log 2FA disabled activity
	if h.activityLogger != nil {
		h.activityLogger.Log2FADisabled(user.AccountID, &user.ID, r.RemoteAddr, r.UserAgent())
	}

	response := map[string]interface{}{
		"message": "Two-factor authentication has been disabled",
		"enabled": false,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetStatus returns the 2FA status for the authenticated user
func (h *TwoFactorHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var twoFactorAuth models.TwoFactorAuth
	err := h.db.Where("user_id = ? AND enabled = ?", userID, true).First(&twoFactorAuth).Error

	enabled := err == nil

	response := map[string]interface{}{
		"enabled": enabled,
	}

	if enabled {
		response["verified_at"] = twoFactorAuth.VerifiedAt
		response["last_used_at"] = twoFactorAuth.LastUsedAt
		response["method"] = twoFactorAuth.Method

		// Count unused recovery codes
		var unusedCount int64
		h.db.Model(&models.RecoveryCode{}).Where("two_factor_auth_id = ? AND used = ?",
			twoFactorAuth.ID, false).Count(&unusedCount)
		response["recovery_codes_remaining"] = unusedCount
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// RegenerateRecoveryCodes generates new recovery codes for the user
func (h *TwoFactorHandler) RegenerateRecoveryCodes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request
	var req struct {
		Password string `json:"password"`
		Code     string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Verify password
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "invalid password")
		return
	}

	// Get 2FA config
	var twoFactorAuth models.TwoFactorAuth
	err := h.db.Where("user_id = ? AND enabled = ?", userID, true).First(&twoFactorAuth).Error
	if err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "2FA is not enabled")
		return
	}

	// Verify TOTP code
	valid, err := ValidateTOTPCode(twoFactorAuth.Secret, req.Code, h.config)
	if err != nil || !valid {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "invalid verification code")
		return
	}

	// Generate new recovery codes
	recoveryCodes, err := GenerateRecoveryCodes(10)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to generate recovery codes")
		return
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete old recovery codes
	if err := tx.Where("two_factor_auth_id = ?", twoFactorAuth.ID).Delete(&models.RecoveryCode{}).Error; err != nil {
		tx.Rollback()
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to remove old codes")
		return
	}

	// Save new recovery codes (hashed)
	for _, code := range recoveryCodes {
		hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
		if err != nil {
			tx.Rollback()
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to hash recovery codes")
			return
		}

		recoveryCode := models.RecoveryCode{
			TwoFactorAuthID: twoFactorAuth.ID,
			CodeHash:        string(hash),
			Used:            false,
		}

		if err := tx.Create(&recoveryCode).Error; err != nil {
			tx.Rollback()
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to save recovery codes")
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to commit changes")
		return
	}

	// Log recovery codes regeneration
	if h.activityLogger != nil {
		h.activityLogger.LogRecoveryCodesRegenerated(user.AccountID, &user.ID, r.RemoteAddr, r.UserAgent())
	}

	response := map[string]interface{}{
		"message":        "Recovery codes regenerated successfully",
		"recovery_codes": recoveryCodes,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// validateRecoveryCode validates and marks a recovery code as used
func (h *TwoFactorHandler) validateRecoveryCode(twoFactorAuthID uint, code string, r *http.Request) bool {
	// Get all unused recovery codes for this 2FA config
	var recoveryCodes []models.RecoveryCode
	if err := h.db.Where("two_factor_auth_id = ? AND used = ?", twoFactorAuthID, false).
		Find(&recoveryCodes).Error; err != nil {
		return false
	}

	// Try each code
	for _, rc := range recoveryCodes {
		if err := bcrypt.CompareHashAndPassword([]byte(rc.CodeHash), []byte(code)); err == nil {
			// Mark as used
			now := time.Now()
			ip := GetClientIP(r)
			rc.Used = true
			rc.UsedAt = &now
			rc.UsedFromIP = &ip

			h.db.Save(&rc)
			return true
		}
	}

	return false
}

// logAttempt logs a 2FA attempt
func (h *TwoFactorHandler) logAttempt(userID uint, success bool, failureType string, r *http.Request) {
	attempt := models.TwoFactorAttempt{
		UserID:      userID,
		Success:     success,
		IPAddress:   GetClientIP(r),
		UserAgent:   r.UserAgent(),
		FailureType: failureType,
		AttemptedAt: time.Now(),
	}

	h.db.Create(&attempt)
}

// GetRecentAttempts returns recent 2FA attempts for the user (admin/debugging)
func (h *TwoFactorHandler) GetRecentAttempts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var attempts []models.TwoFactorAttempt
	if err := h.db.Where("user_id = ?", userID).
		Order("attempted_at DESC").
		Limit(20).
		Find(&attempts).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch attempts")
		return
	}

	response := map[string]interface{}{
		"attempts": attempts,
		"count":    len(attempts),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// SaveRecoveryCodes saves hashed recovery codes during initial setup
func (h *TwoFactorHandler) SaveRecoveryCodes(twoFactorAuthID uint, codes []string) error {
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, code := range codes {
		hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to hash recovery code: %w", err)
		}

		recoveryCode := models.RecoveryCode{
			TwoFactorAuthID: twoFactorAuthID,
			CodeHash:        string(hash),
			Used:            false,
		}

		if err := tx.Create(&recoveryCode).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to save recovery code: %w", err)
		}
	}

	return tx.Commit().Error
}
