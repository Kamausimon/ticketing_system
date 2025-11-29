package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"ticketing_system/internal/analytics"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"ticketing_system/internal/notifications"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db                  *gorm.DB
	metrics             *analytics.PrometheusMetrics
	notificationService *notifications.NotificationService
}

func NewAuthHandler(db *gorm.DB, metrics *analytics.PrometheusMetrics) *AuthHandler {
	return &AuthHandler{
		db:      db,
		metrics: metrics,
	}
}

// NewAuthHandlerWithNotifications creates a new AuthHandler with notification service
func NewAuthHandlerWithNotifications(db *gorm.DB, metrics *analytics.PrometheusMetrics, notifService *notifications.NotificationService) *AuthHandler {
	return &AuthHandler{
		db:                  db,
		metrics:             metrics,
		notificationService: notifService,
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

	var existingUser models.User
	if err := h.db.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingUser).Error; err != gorm.ErrRecordNotFound {
		middleware.WriteJSONError(w, http.StatusConflict, "user already exists")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "password hashing failed")
		return
	}

	user := models.User{
		FirstName:     strings.TrimSpace(req.FirstName),
		LastName:      strings.TrimSpace(req.LastName),
		Username:      strings.ToLower(strings.TrimSpace(req.Username)),
		Phone:         req.Phone,
		Email:         strings.ToLower(strings.TrimSpace(req.Email)),
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
	err := godotenv.Load(".env")
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "error loading the env variables")
		return
	}
	tokenSecret := os.Getenv("JWTSECRET")
	w.Header().Set("Content-Type", "application/json")
	var req LoginRequest
	json.NewDecoder(r.Body).Decode(&req)

	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		// Track failed login attempt
		if h.metrics != nil {
			h.metrics.TrackLoginAttempt("failed")
		}
		middleware.WriteJSONError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Track successful login attempt
	if h.metrics != nil {
		h.metrics.TrackLoginAttempt("success")
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

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req ResetPasswordRequest
	if req.Password != req.PasswordConfirm {
		middleware.WriteJSONError(w, http.StatusBadRequest, "passwords dont match")
		return
	}

	//check if the token exists in password reset and is valid
	var PasswordResetToken models.PasswordReset
	if err := h.db.Where("token = ?", req.ResetToken).First(&PasswordResetToken).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "invalid or expired token")
		return
	}

	if PasswordResetToken.ExpiresAt.Before(time.Now()) {
		middleware.WriteJSONError(w, http.StatusBadRequest, "reset token has expired")
		return
	}

	// Check if token was already used
	if PasswordResetToken.Status != models.ResetPending {
		middleware.WriteJSONError(w, http.StatusBadRequest, "reset token has already been used or is invalid")
		return
	}

	//if all is okay hash the password and update user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "erro hashing password")
		return
	}
	var user models.User
	if err = h.db.Model(&user).Updates(map[string]interface{}{
		"password": hashedPassword,
	}).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "error updating password")
		return
	}

	//delete the reset token
	if err := h.db.Model(&PasswordResetToken).Updates(map[string]interface{}{
		"token": nil,
	}).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, " error deleteing the reset token")
		return
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

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
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

	// Check if user exists
	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// For security, don't reveal if email exists or not
		json.NewEncoder(w).Encode(ForgotPasswordResponse{
			Message: "If an account with that email exists, a password reset link has been sent",
		})
		return
	}

	// Check rate limiting - prevent spam
	var recentReset models.PasswordReset
	if err := h.db.Where("email = ? AND expires_at > ? AND status = ?",
		req.Email, time.Now(), models.ResetPending).First(&recentReset).Error; err == nil {
		middleware.WriteJSONError(w, http.StatusTooManyRequests, "password reset already requested, please check your email")
		return
	}

	// Generate reset token
	resetToken, err := GeneratePasswordResetToken()
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to generate reset token")
		return
	}

	// Create password reset record
	passwordReset := models.PasswordReset{
		Token:        resetToken,
		Email:        req.Email,
		Status:       models.ResetPending,
		Method:       models.ResetByEmail,
		UserID:       &user.ID,
		AccountID:    &user.AccountID,
		IPAddress:    r.RemoteAddr,
		UserAgent:    r.Header.Get("User-Agent"),
		ExpiresAt:    time.Now().Add(15 * time.Minute), // 15 minutes expiry
		IssuedAt:     time.Now(),
		OriginalIP:   r.RemoteAddr,
		CleanupAfter: time.Now().Add(7 * 24 * time.Hour), // Cleanup after 7 days
	}

	// Save to database
	if err := h.db.Create(&passwordReset).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to create reset request")
		return
	}

	// TODO: Send email with reset token
	// SendPasswordResetEmail(req.Email, resetToken)

	json.NewEncoder(w).Encode(ForgotPasswordResponse{
		Message: "Password reset link has been sent to your email",
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
