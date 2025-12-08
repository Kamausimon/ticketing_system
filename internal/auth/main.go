package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"ticketing_system/internal/accounts"
	"ticketing_system/internal/analytics"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"ticketing_system/internal/notifications"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db                  *gorm.DB
	metrics             *analytics.PrometheusMetrics
	notificationService *notifications.NotificationService
	activityLogger      *accounts.ActivityLogger
}

func NewAuthHandler(db *gorm.DB, metrics *analytics.PrometheusMetrics) *AuthHandler {
	return &AuthHandler{
		db:             db,
		metrics:        metrics,
		activityLogger: accounts.NewActivityLogger(db),
	}
}

// NewAuthHandlerWithNotifications creates a new AuthHandler with notification service
func NewAuthHandlerWithNotifications(db *gorm.DB, metrics *analytics.PrometheusMetrics, notifService *notifications.NotificationService) *AuthHandler {
	return &AuthHandler{
		db:                  db,
		metrics:             metrics,
		notificationService: notifService,
		activityLogger:      accounts.NewActivityLogger(db),
	}
}

type RegisterRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type RegisterReponse struct {
	Message string `json:"message"`
	UserId  uint   `json:"user_id"`
	Email   string `json:"email"`
}

func (h *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "make sure to fill in all the fields")
		return
	}

	// Normalize email and username for checking
	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))
	normalizedUsername := strings.ToLower(strings.TrimSpace(req.Username))

	// Check using Count instead of First to avoid locking issues
	var count int64
	err := h.db.Model(&models.User{}).Where("LOWER(email) = ? OR LOWER(username) = ?", normalizedEmail, normalizedUsername).Count(&count).Error

	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "database error checking for existing user")
		return
	}

	if count > 0 {
		middleware.WriteJSONError(w, http.StatusConflict, "user already exists")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "password hashing failed")
		return
	}

	// Create account first
	account := models.Account{
		FirstName: strings.TrimSpace(req.FirstName),
		LastName:  strings.TrimSpace(req.LastName),
		Email:     normalizedEmail,
		IsActive:  true,
		IsBanned:  false,
	}

	if err := h.db.Create(&account).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to create account")
		return
	}
	user := models.User{
		AccountID:     account.ID,
		FirstName:     strings.TrimSpace(req.FirstName),
		LastName:      strings.TrimSpace(req.LastName),
		Username:      normalizedUsername,
		Phone:         req.Phone,
		Email:         normalizedEmail,
		Password:      string(hashedPassword),
		Role:          models.RoleCustomer,
		IsActive:      true,
		Isconfirmed:   false,
		EmailVerified: false, // Email not verified on signup
	}

	if err := h.db.Create(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	// Generate verification token
	verificationToken, err := GenerateSecureToken(32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to generate verification token")
		return
	}

	expiresAt := time.Now().Add(24 * time.Hour) // Token valid for 24 hours

	// Create email verification record
	emailVerification := models.EmailVerification{
		UserID:     user.ID,
		Token:      verificationToken,
		Email:      user.Email,
		Status:     models.VerificationPending,
		ExpiresAt:  expiresAt,
		LastSentAt: time.Now(),
		IPAddress:  r.RemoteAddr,
		UserAgent:  r.Header.Get("User-Agent"),
		IssuedAt:   time.Now(),
	}

	if err := h.db.Create(&emailVerification).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to create verification token")
		return
	}

	// Send verification email if notification service is available
	if h.notificationService != nil {
		fullName := user.FirstName + " " + user.LastName
		go func() {
			if err := h.notificationService.SendVerificationEmail(user.Email, fullName, verificationToken); err != nil {
				// Log error but don't fail registration
				middleware.WriteJSONError(w, http.StatusInternalServerError, "user registered but failed to send verification email")
			}
		}()
	}

	// Track user registration
	if h.metrics != nil {
		h.metrics.UsersRegistered.Inc()
	}

	// Log registration activity
	if h.activityLogger != nil {
		h.activityLogger.LogRegistration(user.AccountID, &user.ID, r.RemoteAddr, r.Header.Get("User-Agent"))
	}

	response := RegisterReponse{
		Message: "user registered successfully. Please check your email to verify your account",
		UserId:  user.ID,
		Email:   user.Email,
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Do not load env per request; read JWT secret directly
	tokenSecret := os.Getenv("JWTSECRET")
	if tokenSecret == "" {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "server configuration error: missing JWTSECRET")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "email and password must be provided")
		return
	}

	// normalize email
	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))

	var user models.User
	// Use GORM with Select and timeout; default scope excludes soft-deleted rows
	{
		ctx, cancel := context.WithTimeout(r.Context(), 1500*time.Millisecond)
		defer cancel()
		err := h.db.WithContext(ctx).
			Model(&models.User{}).
			Select("id", "account_id", "email", "password", "role", "is_active", "email_verified").
			Where("LOWER(email) = ?", normalizedEmail).
			First(&user).Error

		if err != nil && err != gorm.ErrRecordNotFound {
			middleware.WriteJSONError(w, http.StatusInternalServerError, "database error")
			return
		}
	}

	if user.ID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		// Track failed login attempt
		if h.metrics != nil {
			h.metrics.TrackLoginAttempt("failed")
		}
		// Log failed login
		if h.activityLogger != nil {
			h.activityLogger.LogLoginFailed(user.AccountID, r.RemoteAddr, r.Header.Get("User-Agent"), "invalid password")
		}
		middleware.WriteJSONError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Check if 2FA is enabled for this user with timeout and lightweight existence check
	has2FA := false
	{
		ctx, cancel := context.WithTimeout(r.Context(), 800*time.Millisecond)
		defer cancel()
		var count int64
		if err := h.db.WithContext(ctx).
			Model(&models.TwoFactorAuth{}).
			Where("user_id = ? AND enabled = ?", user.ID, true).
			Count(&count).Error; err == nil {
			has2FA = count > 0
		}
	}

	if has2FA {
		// 2FA is enabled - return partial token and require 2FA verification
		// Create a temporary short-lived token for 2FA verification only
		tempToken, err := MakeJWT(user.ID, tokenSecret, 15*time.Minute) // 15 min temp token
		if err != nil {
			middleware.WriteJSONError(w, http.StatusInternalServerError, "error generating verification token")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":      "Two-factor authentication required",
			"requires_2fa": true,
			"temp_token":   tempToken, // Temporary token for 2FA verification
			"user_id":      user.ID,
		})
		return
	}

	// Track successful login attempt
	if h.metrics != nil {
		h.metrics.TrackLoginAttempt("success")
	}

	// Log successful login
	if h.activityLogger != nil {
		h.activityLogger.LogLogin(user.AccountID, &user.ID, r.RemoteAddr, r.Header.Get("User-Agent"))
	}

	expirationDuration := time.Hour
	token, err := MakeJWT(user.ID, tokenSecret, expirationDuration)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "error generating jwt token")
		return
	}

	refreshToken, err := MakeRefreshToken()
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, " error creating refreshToken")
		return
	}

	refreshExpiresAt := time.Now().Add(60 * 24 * time.Hour).Unix()

	if err = h.db.Model(&user).Updates(map[string]interface{}{
		"refresh_token":     &refreshToken,
		"refresh_token_exp": &refreshExpiresAt,
		"last_login_at":     time.Now().Unix(),
	}).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update user session")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"user_id": user.ID,
		"role":    user.Role,
		"token":   token,
	})
}

func (h *AuthHandler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := middleware.GetUserIDFromToken(r)
	var user models.User
	if err := h.db.First(&user, userID).Error; err == nil {
		// Log logout activity
		if h.activityLogger != nil {
			h.activityLogger.LogLogout(user.AccountID, &user.ID, r.RemoteAddr, r.Header.Get("User-Agent"))
		}
	}

	if err := h.db.Model(&models.User{}).Where("id =  ?", userID).Updates(map[string]interface{}{
		"refresh_token":     nil,
		"refresh_token_exp": nil,
	}).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "logout failed")
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "logged out successfully",
	})
}

type ResetPasswordRequest struct {
	ResetToken      string `json:"token"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"passwordConfirm"`
}

type ResetResponse struct {
	Message string `json:"message"`
}

// ResetPassword handles password reset with comprehensive validation and security tracking
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	// Add security headers
	AddSecurityHeaders(w)
	w.Header().Set("Content-Type", "application/json")

	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request fields
	if req.ResetToken == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "reset token is required")
		return
	}
	if req.Password == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "password is required")
		return
	}
	if req.Password != req.PasswordConfirm {
		middleware.WriteJSONError(w, http.StatusBadRequest, "passwords do not match")
		return
	}
	if len(req.Password) < 8 {
		middleware.WriteJSONError(w, http.StatusBadRequest, "password must be at least 8 characters long")
		return
	}

	// Retrieve password reset token with comprehensive validation
	var passwordReset models.PasswordReset
	if err := h.db.Where("token = ?", req.ResetToken).First(&passwordReset).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid or expired token")
		return
	}

	// Record attempt for security tracking
	attempt := models.PasswordResetAttempt{
		PasswordResetID: passwordReset.ID,
		IPAddress:       GetClientIP(r),
		UserAgent:       r.Header.Get("User-Agent"),
		AttemptedAt:     time.Now(),
	}

	// VALIDATION 1: Check token expiration
	if time.Now().After(passwordReset.ExpiresAt) {
		attempt.FailureReason = stringPtr("Token expired")
		attempt.ErrorCode = stringPtr("TOKEN_EXPIRED")
		h.db.Create(&attempt)

		// Mark token as expired
		h.db.Model(&passwordReset).Update("status", models.ResetExpired)

		middleware.WriteJSONError(w, http.StatusBadRequest, "reset token has expired. Please request a new password reset")
		return
	}
	attempt.NotExpired = true

	// VALIDATION 2: Check if token was already used
	if passwordReset.Status != models.ResetPending {
		attempt.FailureReason = stringPtr(fmt.Sprintf("Token already %s", passwordReset.Status))
		attempt.ErrorCode = stringPtr("TOKEN_ALREADY_USED")
		h.db.Create(&attempt)

		middleware.WriteJSONError(w, http.StatusConflict, "this reset token has already been used. Please request a new one")
		return
	}

	// VALIDATION 3: Check attempt count
	if passwordReset.AttemptCount >= passwordReset.MaxAttempts {
		attempt.FailureReason = stringPtr("Max attempts exceeded")
		attempt.ErrorCode = stringPtr("MAX_ATTEMPTS_EXCEEDED")
		h.db.Create(&attempt)

		// Mark token as invalid for security
		h.db.Model(&passwordReset).Update("status", models.ResetInvalid)

		middleware.WriteJSONError(w, http.StatusForbidden, "maximum reset attempts exceeded. Please request a new password reset link")
		return
	}

	// VALIDATION 4: IP consistency check (optional but recommended)
	if passwordReset.SameIPRequired && passwordReset.OriginalIP != GetClientIP(r) {
		attempt.IPMatched = false
		attempt.FailureReason = stringPtr("IP mismatch")
		attempt.ErrorCode = stringPtr("IP_MISMATCH")
		h.db.Create(&attempt)

		middleware.WriteJSONError(w, http.StatusForbidden, "password reset attempted from different IP address")
		return
	}
	attempt.IPMatched = true

	// Get associated user
	var user models.User
	if err := h.db.First(&user, passwordReset.UserID).Error; err != nil {
		attempt.FailureReason = stringPtr("User not found")
		attempt.ErrorCode = stringPtr("USER_NOT_FOUND")
		h.db.Create(&attempt)

		middleware.WriteJSONError(w, http.StatusNotFound, "user account not found")
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		attempt.FailureReason = stringPtr("Password hashing failed")
		attempt.ErrorCode = stringPtr("HASH_ERROR")
		h.db.Create(&attempt)

		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to process password reset")
		return
	}

	// Update password and mark token as used
	now := time.Now()
	if err := h.db.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		attempt.FailureReason = stringPtr("Password update failed")
		attempt.ErrorCode = stringPtr("UPDATE_ERROR")
		h.db.Create(&attempt)

		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update password")
		return
	}

	// Mark password reset as used
	if err := h.db.Model(&passwordReset).Updates(map[string]interface{}{
		"status":       models.ResetUsed,
		"used_at":      now,
		"used_from_ip": GetClientIP(r),
	}).Error; err != nil {
		attempt.FailureReason = stringPtr("Failed to mark token as used")
		attempt.ErrorCode = stringPtr("MARK_USED_ERROR")
		h.db.Create(&attempt)

		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to finalize password reset")
		return
	}

	// Record successful attempt
	attempt.WasSuccessful = true
	attempt.TokenValid = true
	h.db.Create(&attempt)

	// Log password reset activity
	if h.activityLogger != nil {
		h.activityLogger.LogPasswordReset(user.AccountID, &user.ID, r.RemoteAddr, r.Header.Get("User-Agent"))
	}

	// Send confirmation email if notification service available
	if h.notificationService != nil {
		h.notificationService.SendPlainEmail([]string{user.Email}, "Password Reset Successful", "Your password has been successfully reset.")
	}

	response := ResetResponse{
		Message: "password reset successfully",
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ForgotPasswordResponse struct {
	Message string `json:"message"`
}

// ForgotPassword handles password reset request with rate limiting and attempt tracking
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	// Add security headers
	AddSecurityHeaders(w)
	w.Header().Set("Content-Type", "application/json")

	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "email is required")
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	clientIP := GetClientIP(r)

	// Check if user exists
	var user models.User
	userExists := true
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		userExists = false
	}

	// RATE LIMITING: Check if user has requested too many resets
	// Get configuration with fallback defaults
	var config models.ResetConfiguration
	if err := h.db.First(&config).Error; err != nil {
		// No config in database, use sensible defaults
		config = models.ResetConfiguration{
			TokenExpiryMinutes:  15,
			MaxAttemptsPerToken: 3,
			MaxRequestsPerHour:  5,
			MaxRequestsPerIP:    10,
			CleanupAfterDays:    7,
		}
	}

	if userExists {
		// Count recent requests from this user
		var recentCount int64
		h.db.Model(&models.PasswordReset{}).
			Where("email = ? AND issued_at > ? AND status IN (?)",
				req.Email,
				time.Now().Add(-1*time.Hour),
				[]models.ResetStatus{models.ResetPending, models.ResetUsed}).
			Count(&recentCount)

		if int(recentCount) >= config.MaxRequestsPerHour {
			middleware.WriteJSONError(w, http.StatusTooManyRequests,
				"too many password reset requests. Please try again later")
			return
		}

		// Check for ongoing pending reset
		var pendingReset models.PasswordReset
		if err := h.db.Where("email = ? AND status = ? AND expires_at > ?",
			req.Email, models.ResetPending, time.Now()).First(&pendingReset).Error; err == nil {
			// Existing pending reset, don't create another
			json.NewEncoder(w).Encode(ForgotPasswordResponse{
				Message: "If an account with that email exists, a password reset link has been sent",
			})
			return
		}
	}

	// Rate limiting per IP
	var ipCount int64
	h.db.Model(&models.PasswordReset{}).
		Where("ip_address = ? AND issued_at > ?",
			clientIP, time.Now().Add(-1*time.Hour)).
		Count(&ipCount)

	if int(ipCount) >= config.MaxRequestsPerIP {
		middleware.WriteJSONError(w, http.StatusTooManyRequests,
			"too many password reset requests from your IP address")
		return
	}

	// Generate reset token
	resetToken, err := GeneratePasswordResetToken()
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to generate reset token")
		return
	}

	now := time.Now()
	passwordReset := models.PasswordReset{
		Token:        resetToken,
		Email:        req.Email,
		Status:       models.ResetPending,
		Method:       models.ResetByEmail,
		UserID:       nil,
		AccountID:    nil,
		IPAddress:    clientIP,
		UserAgent:    r.Header.Get("User-Agent"),
		ExpiresAt:    now.Add(time.Duration(config.TokenExpiryMinutes) * time.Minute),
		IssuedAt:     now,
		OriginalIP:   clientIP,
		MaxAttempts:  config.MaxAttemptsPerToken,
		AttemptCount: 0,
		CleanupAfter: now.Add(time.Duration(config.CleanupAfterDays) * 24 * time.Hour),
	}

	// Assign user if exists
	if userExists {
		passwordReset.UserID = &user.ID
		if user.AccountID != 0 {
			passwordReset.AccountID = &user.AccountID
		}
	}

	// Save to database
	if err := h.db.Create(&passwordReset).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to create reset request")
		return
	}

	// Log password reset request
	if userExists && h.activityLogger != nil {
		h.activityLogger.LogPasswordResetRequest(user.AccountID, &user.ID, r.RemoteAddr, r.Header.Get("User-Agent"))
	}

	// Send email with reset token if user exists
	if userExists && h.notificationService != nil {
		h.notificationService.SendPasswordResetEmail(user.Email, user.FirstName, resetToken)

	}
	// Generic response (don't reveal if email exists)
	json.NewEncoder(w).Encode(ForgotPasswordResponse{
		Message: "If an account with that email exists, a password reset link has been sent",
	})
}

// GeneratePasswordResetToken creates a secure random token for password resets
func GeneratePasswordResetToken() (string, error) {
	return GenerateSecureToken(32) // 32 characters
}

// GenerateSecureToken creates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length/2) // hex encoding doubles the length
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// VerifyEmailRequest represents an email verification request
type VerifyEmailRequest struct {
	Token string `json:"token"`
}

// VerifyEmailResponse represents an email verification response
type VerifyEmailResponse struct {
	Message string `json:"message"`
	Email   string `json:"email"`
}

// VerifyEmail verifies a user's email address
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Token == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "verification token is required")
		return
	}

	// Find verification record
	var verification models.EmailVerification
	if err := h.db.Where("token = ?", req.Token).First(&verification).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid or expired verification token")
		return
	}

	// Check if token is expired
	if verification.ExpiresAt.Before(time.Now()) {
		verification.Status = models.VerificationExpired
		h.db.Save(&verification)
		middleware.WriteJSONError(w, http.StatusBadRequest, "verification token has expired")
		return
	}

	// Check if already verified
	if verification.Status == models.VerificationVerified {
		middleware.WriteJSONError(w, http.StatusConflict, "email already verified")
		return
	}

	// Update user
	if err := h.db.Model(&models.User{}).Where("id = ?", verification.UserID).Updates(map[string]interface{}{
		"email_verified":    true,
		"email_verified_at": time.Now(),
	}).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to verify email")
		return
	}

	// Update verification record
	now := time.Now()
	verification.Status = models.VerificationVerified
	verification.VerifiedAt = &now
	if err := h.db.Save(&verification).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update verification record")
		return
	}

	// Get user for activity logging
	var user models.User
	if err := h.db.First(&user, verification.UserID).Error; err == nil {
		if h.activityLogger != nil {
			h.activityLogger.LogEmailVerified(user.AccountID, &user.ID, r.RemoteAddr, r.Header.Get("User-Agent"))
		}
	}

	response := VerifyEmailResponse{
		Message: "email verified successfully",
		Email:   verification.Email,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ResendVerificationRequest represents a resend verification request
type ResendVerificationRequest struct {
	Email string `json:"email"`
}

// ResendVerification resends the verification email
func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req ResendVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "email is required")
		return
	}

	// Find user
	var user models.User
	if err := h.db.Where("email = ?", strings.ToLower(req.Email)).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Check if already verified
	if user.EmailVerified {
		middleware.WriteJSONError(w, http.StatusConflict, "email already verified")
		return
	}

	// Find pending verification
	var verification models.EmailVerification
	if err := h.db.Where("user_id = ? AND status = ?", user.ID, models.VerificationPending).First(&verification).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "no pending verification found")
		return
	}

	// Check if max resends reached
	if verification.ResendCount >= verification.MaxResends {
		middleware.WriteJSONError(w, http.StatusTooManyRequests, "maximum resend attempts reached. Please contact support")
		return
	}

	// Check rate limiting - don't allow resend within 5 minutes
	if time.Since(verification.LastSentAt) < 5*time.Minute {
		middleware.WriteJSONError(w, http.StatusTooManyRequests, "please wait before requesting a new verification email")
		return
	}

	// Generate new token
	newToken, err := GenerateSecureToken(32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to generate new verification token")
		return
	}

	// Update verification record
	verification.Token = newToken
	verification.ExpiresAt = time.Now().Add(24 * time.Hour)
	verification.LastSentAt = time.Now()
	verification.ResendCount++

	if err := h.db.Save(&verification).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update verification token")
		return
	}

	// Resend verification email
	if h.notificationService != nil {
		fullName := user.FirstName + " " + user.LastName
		go func() {
			if err := h.notificationService.SendVerificationEmail(user.Email, fullName, newToken); err != nil {
				// Log but don't fail the request
			}
		}()
	}

	response := map[string]interface{}{
		"message": "verification email resent successfully",
		"email":   user.Email,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CheckEmailVerificationStatus checks the email verification status
func (h *AuthHandler) CheckEmailVerificationStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	response := map[string]interface{}{
		"email_verified": user.EmailVerified,
		"email":          user.Email,
		"verified_at":    user.EmailVerifiedAt,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
