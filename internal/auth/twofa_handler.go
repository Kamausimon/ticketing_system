package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ticketing_system/internal/accounts"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"ticketing_system/internal/notifications"
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
	emailService   *notifications.EmailService
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

// SetEmailService sets the email service for sending 2FA setup emails
func (h *TwoFactorHandler) SetEmailService(emailService *notifications.EmailService) {
	h.emailService = emailService
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

	userID, authErr := middleware.GetUserIDFromTokenWithError(r)
	if authErr != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
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

	// Send QR code via email if email service is configured
	if h.emailService != nil {
		go h.sendSetupEmail(user.Email, user.Username, qrCodeData, secret, recoveryCodes)
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

	userID, authErr := middleware.GetUserIDFromTokenWithError(r)
	if authErr != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
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

	// Delete any existing 2FA records for this user (including soft-deleted ones)
	// This prevents unique constraint violations when re-enabling 2FA
	var existingAuth models.TwoFactorAuth
	if err := tx.Unscoped().Where("user_id = ?", userID).First(&existingAuth).Error; err == nil {
		// Delete associated recovery codes first
		tx.Unscoped().Where("two_factor_auth_id = ?", existingAuth.ID).Delete(&models.RecoveryCode{})
		// Delete the 2FA auth record
		tx.Unscoped().Delete(&existingAuth)
	}

	// Save new 2FA config
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

	userID, authErr := middleware.GetUserIDFromTokenWithError(r)
	if authErr != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
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

	userID, authErr := middleware.GetUserIDFromTokenWithError(r)
	if authErr != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
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

	// Hard delete all recovery codes (permanently remove, not soft delete)
	if err := tx.Unscoped().Where("two_factor_auth_id = ?", twoFactorAuth.ID).Delete(&models.RecoveryCode{}).Error; err != nil {
		tx.Rollback()
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to remove recovery codes")
		return
	}

	// Hard delete 2FA config (permanently remove to allow re-enabling)
	if err := tx.Unscoped().Delete(&twoFactorAuth).Error; err != nil {
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

	userID, authErr := middleware.GetUserIDFromTokenWithError(r)
	if authErr != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
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

	userID, authErr := middleware.GetUserIDFromTokenWithError(r)
	if authErr != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
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

	// Hard delete old recovery codes (permanently remove)
	if err := tx.Unscoped().Where("two_factor_auth_id = ?", twoFactorAuth.ID).Delete(&models.RecoveryCode{}).Error; err != nil {
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

	// Send new recovery codes via email if email service is configured
	if h.emailService != nil {
		go h.sendRecoveryCodesEmail(user.Email, user.Username, recoveryCodes)
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

	userID, authErr := middleware.GetUserIDFromTokenWithError(r)
	if authErr != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
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

// sendSetupEmail sends the 2FA setup email with QR code
func (h *TwoFactorHandler) sendSetupEmail(email, username, qrCodeBase64, secret string, recoveryCodes []string) {
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #4CAF50; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border: 1px solid #ddd; border-radius: 0 0 5px 5px; }
        .qr-code { text-align: center; margin: 20px 0; }
        .qr-code img { max-width: 300px; border: 2px solid #4CAF50; padding: 10px; background: white; }
        .secret-box { background: #fff; border: 2px dashed #4CAF50; padding: 15px; margin: 20px 0; text-align: center; }
        .secret-key { font-size: 18px; font-weight: bold; color: #4CAF50; letter-spacing: 2px; }
        .backup-codes { background: #fff; border: 2px solid #ff9800; padding: 15px; margin: 20px 0; }
        .backup-codes h3 { color: #ff9800; margin-top: 0; }
        .code-list { display: grid; grid-template-columns: repeat(2, 1fr); gap: 10px; font-family: monospace; }
        .warning { background: #fff3cd; border: 1px solid #ffc107; padding: 15px; margin: 20px 0; border-radius: 5px; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
        .step { margin: 15px 0; padding: 10px; background: white; border-left: 4px solid #4CAF50; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔐 Two-Factor Authentication Setup</h1>
        </div>
        <div class="content">
            <p>Hi %s,</p>
            <p>You've enabled Two-Factor Authentication (2FA) for your account. Follow these steps to complete the setup:</p>
            
            <div class="step">
                <strong>Step 1:</strong> Download an authenticator app if you don't have one:
                <ul>
                    <li>Google Authenticator (iOS/Android)</li>
                    <li>Authy (iOS/Android/Desktop)</li>
                    <li>Microsoft Authenticator (iOS/Android)</li>
                </ul>
            </div>
            
            <div class="step">
                <strong>Step 2:</strong> Scan this QR code with your authenticator app:
            </div>
            
            <div class="qr-code">
                <img src="data:image/png;base64,%s" alt="2FA QR Code" />
            </div>
            
            <div class="step">
                <strong>Step 3 (Alternative):</strong> Or manually enter this secret key:
            </div>
            
            <div class="secret-box">
                <p>Secret Key:</p>
                <div class="secret-key">%s</div>
                <p style="font-size: 12px; color: #666; margin-top: 10px;">
                    (Use this if you can't scan the QR code)
                </p>
            </div>
            
            <div class="warning">
                <strong>⚠️ Important:</strong> Save your backup codes! These are the ONLY way to access your account if you lose your device.
            </div>
            
            <div class="backup-codes">
                <h3>🔑 Backup Recovery Codes</h3>
                <p>Each code can only be used once. Store them in a safe place!</p>
                <div class="code-list">
                    %s
                </div>
            </div>
            
            <div class="warning">
                <strong>Security Tips:</strong>
                <ul>
                    <li>Never share your secret key or recovery codes with anyone</li>
                    <li>Store recovery codes in a secure location (password manager, safe, etc.)</li>
                    <li>This email contains sensitive information - delete it after saving your codes</li>
                </ul>
            </div>
            
            <p style="margin-top: 30px;">If you didn't request this, please contact support immediately.</p>
        </div>
        <div class="footer">
            <p>This is an automated message from Ticketing System</p>
            <p>&copy; 2025 Ticketing System. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, username, qrCodeBase64, secret, h.formatBackupCodes(recoveryCodes))

	emailData := notifications.EmailData{
		To:       []string{email},
		Subject:  "🔐 Two-Factor Authentication Setup",
		HTMLBody: htmlBody,
		Body: fmt.Sprintf(`Two-Factor Authentication Setup

Hi %s,

You've enabled Two-Factor Authentication for your account.

Secret Key: %s

Backup Codes:
%s

Please scan the QR code in the HTML version of this email or manually enter the secret key in your authenticator app.

Keep your backup codes in a safe place!

If you didn't request this, please contact support immediately.
`, username, secret, h.formatBackupCodesPlainText(recoveryCodes)),
	}

	if err := h.emailService.Send(emailData); err != nil {
		// Log error but don't fail the setup process
		fmt.Printf("Failed to send 2FA setup email to %s: %v\n", email, err)
	}
}

// formatBackupCodes formats backup codes for HTML display
func (h *TwoFactorHandler) formatBackupCodes(codes []string) string {
	var html string
	for _, code := range codes {
		html += fmt.Sprintf("<div>%s</div>", code)
	}
	return html
}

// formatBackupCodesPlainText formats backup codes for plain text display
func (h *TwoFactorHandler) formatBackupCodesPlainText(codes []string) string {
	var text string
	for i, code := range codes {
		text += fmt.Sprintf("%d. %s\n", i+1, code)
	}
	return text
}

// sendRecoveryCodesEmail sends the regenerated recovery codes via email
func (h *TwoFactorHandler) sendRecoveryCodesEmail(email, username string, recoveryCodes []string) {
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #ff9800; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border: 1px solid #ddd; border-radius: 0 0 5px 5px; }
        .backup-codes { background: #fff; border: 2px solid #ff9800; padding: 20px; margin: 20px 0; border-radius: 5px; }
        .backup-codes h3 { color: #ff9800; margin-top: 0; text-align: center; }
        .code-list { display: grid; grid-template-columns: repeat(2, 1fr); gap: 15px; font-family: monospace; font-size: 14px; margin-top: 20px; }
        .code-item { background: #f5f5f5; padding: 10px; border-radius: 3px; text-align: center; font-weight: bold; }
        .warning { background: #fff3cd; border: 1px solid #ffc107; padding: 15px; margin: 20px 0; border-radius: 5px; }
        .alert { background: #f8d7da; border: 1px solid #f5c2c7; padding: 15px; margin: 20px 0; border-radius: 5px; color: #842029; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
        .timestamp { background: #e3f2fd; padding: 10px; border-left: 4px solid #2196F3; margin: 20px 0; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔑 New Recovery Codes Generated</h1>
        </div>
        <div class="content">
            <p>Hi %s,</p>
            <p>Your Two-Factor Authentication recovery codes have been regenerated as requested.</p>
            
            <div class="alert">
                <strong>⚠️ Important Security Notice:</strong>
                <ul style="margin: 10px 0;">
                    <li>Your old recovery codes are now <strong>INVALID</strong></li>
                    <li>Each new code below can only be used <strong>ONCE</strong></li>
                    <li>Store these codes in a secure location immediately</li>
                </ul>
            </div>
            
            <div class="timestamp">
                <strong>Generated:</strong> %s
            </div>
            
            <div class="backup-codes">
                <h3>🔐 Your New Recovery Codes</h3>
                <p style="text-align: center; color: #666; margin-bottom: 20px;">
                    Save these codes now - you won't be able to see them again!
                </p>
                <div class="code-list">
                    %s
                </div>
            </div>
            
            <div class="warning">
                <strong>What are recovery codes?</strong>
                <p>Recovery codes allow you to access your account if you lose your authentication device. Each code can only be used once.</p>
            </div>
            
            <div class="warning">
                <strong>Security Best Practices:</strong>
                <ul>
                    <li>✅ Store codes in a password manager</li>
                    <li>✅ Keep a printed copy in a safe place</li>
                    <li>✅ Never share these codes with anyone</li>
                    <li>✅ Regenerate codes if you suspect they've been compromised</li>
                    <li>✅ Delete this email after saving the codes securely</li>
                </ul>
            </div>
            
            <div class="alert">
                <strong>Didn't request this?</strong><br>
                If you didn't regenerate your recovery codes, your account may be compromised. 
                Please contact support immediately and change your password.
            </div>
        </div>
        <div class="footer">
            <p>This is an automated security notification from Ticketing System</p>
            <p>&copy; 2025 Ticketing System. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, username, time.Now().Format("January 2, 2006 at 3:04 PM MST"), h.formatBackupCodesWithStyle(recoveryCodes))

	emailData := notifications.EmailData{
		To:       []string{email},
		Subject:  "🔑 New Two-Factor Authentication Recovery Codes",
		HTMLBody: htmlBody,
		Body: fmt.Sprintf(`New Recovery Codes Generated

Hi %s,

Your Two-Factor Authentication recovery codes have been regenerated.

IMPORTANT: Your old recovery codes are now INVALID.

New Recovery Codes (each can only be used once):
%s

Security Tips:
- Store these codes in a secure location (password manager, safe, etc.)
- Each code can only be used once
- Never share these codes with anyone
- Delete this email after saving the codes securely

If you didn't request this, please contact support immediately and change your password.

Generated: %s

Ticketing System
`, username, h.formatBackupCodesPlainText(recoveryCodes), time.Now().Format("January 2, 2006 at 3:04 PM MST")),
	}

	if err := h.emailService.Send(emailData); err != nil {
		// Log error but don't fail the process
		fmt.Printf("Failed to send recovery codes email to %s: %v\n", email, err)
	}
}

// formatBackupCodesWithStyle formats backup codes with HTML styling
func (h *TwoFactorHandler) formatBackupCodesWithStyle(codes []string) string {
	var html string
	for _, code := range codes {
		html += fmt.Sprintf("<div class=\"code-item\">%s</div>", code)
	}
	return html
}
