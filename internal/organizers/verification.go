package organizers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"ticketing_system/internal/auth"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"ticketing_system/internal/notifications"
	"time"

	"github.com/gorilla/mux"
)

// Verification-related structures
type VerificationRequest struct {
	Action string `json:"action"` // "approve" or "reject"
	Reason string `json:"reason"` // Optional reason for rejection
}

type VerificationResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

type PendingOrganizerResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	TaxName   string `json:"tax_name"`
	TaxPin    string `json:"tax_pin"`
	CreatedAt string `json:"created_at"`
}

// GetPendingOrganizers returns organizers awaiting verification (Admin only)
func (h *OrganizerHandler) GetPendingOrganizers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Check if user is admin
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	if user.Role != models.RoleAdmin {
		middleware.WriteJSONError(w, http.StatusForbidden, "admin access required")
		return
	}

	// Get organizers with unconfirmed emails (pending verification)
	var organizers []models.Organizer
	if err := h.db.Where("is_verified = ?", false).Find(&organizers).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch pending organizers")
		return
	}

	// Convert to response format
	var pendingOrganizers []PendingOrganizerResponse
	for _, org := range organizers {
		pendingOrganizers = append(pendingOrganizers, PendingOrganizerResponse{
			ID:        org.ID,
			Name:      org.Name,
			Email:     org.Email,
			Phone:     org.Phone,
			TaxName:   org.TaxName,
			TaxPin:    org.TaxPin,
			CreatedAt: org.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	json.NewEncoder(w).Encode(pendingOrganizers)
}

// VerifyOrganizer handles organizer verification by admin
func (h *OrganizerHandler) VerifyOrganizer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Check if user is admin
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	if user.Role != models.RoleAdmin {
		middleware.WriteJSONError(w, http.StatusForbidden, "admin access required")
		return
	}

	// Get organizer ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "organizer ID is required")
		return
	}

	organizerID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid organizer ID format")
		return
	}

	// Parse request
	var req VerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate action
	if req.Action != "approve" && req.Action != "reject" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "action must be 'approve' or 'reject'")
		return
	}

	// Find organizer
	var organizer models.Organizer
	if err := h.db.Where("id = ?", uint(organizerID)).First(&organizer).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "organizer not found")
		return
	}

	var response VerificationResponse

	if req.Action == "approve" {
		// Approve organizer
		updates := map[string]interface{}{
			"is_email_confirmed":  true,
			"is_verified":         true,
			"verification_status": "approved",
		}
		if err := h.db.Model(&organizer).Updates(updates).Error; err != nil {
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to approve organizer")
			return
		}

		response = VerificationResponse{
			Message: "Organizer approved successfully",
			Status:  "approved",
		}

		// Send approval email to organizer
		approvalData := notifications.OrganizerApprovalData{
			OrganizerName:  organizer.Name,
			OrganizerEmail: organizer.Email,
		}
		if err := h.notifications.SendOrganizerApprovalEmail(organizer.Email, approvalData); err != nil {
			// Log the error but don't fail the approval
			log.Printf("❌ Failed to send approval email: %v", err)
		}

	} else {
		// Reject organizer
		updates := map[string]interface{}{
			"is_verified":         false,
			"verification_status": "rejected",
			"rejection_reason":    req.Reason,
		}
		if err := h.db.Model(&organizer).Updates(updates).Error; err != nil {
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to reject organizer")
			return
		}

		response = VerificationResponse{
			Message: "Organizer rejected. Reason: " + req.Reason,
			Status:  "rejected",
		}

		// Send rejection email to organizer
		rejectionData := notifications.OrganizerRejectionData{
			OrganizerName:   organizer.Name,
			OrganizerEmail:  organizer.Email,
			RejectionReason: req.Reason,
		}
		if err := h.notifications.SendOrganizerRejectionEmail(organizer.Email, rejectionData); err != nil {
			// Log the error but don't fail the rejection
			log.Printf("❌ Failed to send rejection email: %v", err)
		}
	}

	json.NewEncoder(w).Encode(response)

}

// SendVerificationEmail sends verification email to organizer
func (h *OrganizerHandler) SendVerificationEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get organizer
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	var organizer models.Organizer
	if err := h.db.Where("account_id = ?", user.AccountID).First(&organizer).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "organizer profile not found")
		return
	}

	if organizer.IsEmailConfirmed {
		middleware.WriteJSONError(w, http.StatusBadRequest, "email already verified")
		return
	}

	// Generate verification token
	token := generateVerificationToken()
	expiresAt := time.Now().Add(24 * time.Hour)

	// Store verification token
	verification := models.EmailVerification{
		UserID:     userID,
		Email:      organizer.Email,
		Token:      token,
		ExpiresAt:  expiresAt,
		IssuedAt:   time.Now(),
		LastSentAt: time.Now(),
		Status:     models.VerificationPending,
		IPAddress:  auth.GetClientIP(r),
		UserAgent:  r.UserAgent(),
	}

	if err := h.db.Create(&verification).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to create verification token")
		return
	}

	// Send verification email using notification service
	if h.notifications != nil {
		emailData := notifications.EmailVerificationData{
			Name:             organizer.Name,
			Email:            organizer.Email,
			VerificationLink: "https://yourdomain.com/organizers/verify-email?token=" + token,
			ExpiresAt:        expiresAt.Format("January 2, 2006 at 3:04 PM"),
		}
		if err := h.notifications.SendOrganizerEmailVerification(organizer.Email, emailData); err != nil {
			log.Printf("❌ Failed to send verification email: %v", err)
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to send verification email")
			return
		}
	} else {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "email service not configured")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Verification email sent successfully",
		"expires_at": expiresAt,
	})
}

// VerifyOrganizerEmail verifies organizer email using token
func (h *OrganizerHandler) VerifyOrganizerEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "verification token is required")
		return
	}

	// Find verification record
	var verification models.EmailVerification
	if err := h.db.Where("token = ?", token).First(&verification).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "invalid or expired verification token")
		return
	}

	// Check if token is expired
	if time.Now().After(verification.ExpiresAt) {
		middleware.WriteJSONError(w, http.StatusBadRequest, "verification token has expired")
		return
	}

	// Check if already verified
	if verification.VerifiedAt != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "email already verified")
		return
	}

	// Find organizer by email
	var organizer models.Organizer
	if err := h.db.Where("email = ?", verification.Email).First(&organizer).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "organizer not found")
		return
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Mark email as verified
	now := time.Now()
	if err := tx.Model(&verification).Updates(map[string]interface{}{
		"verified_at": now,
	}).Error; err != nil {
		tx.Rollback()
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update verification status")
		return
	}

	// Update organizer email confirmation status
	if err := tx.Model(&organizer).Updates(map[string]interface{}{
		"is_email_confirmed":  true,
		"verification_status": "email_verified",
	}).Error; err != nil {
		tx.Rollback()
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update organizer status")
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to complete verification")
		return
	}

	// Send welcome/confirmation email
	if h.notifications != nil {
		welcomeData := notifications.OrganizerWelcomeData{
			OrganizerName:  organizer.Name,
			OrganizerEmail: organizer.Email,
		}
		if err := h.notifications.SendOrganizerWelcome(organizer.Email, welcomeData); err != nil {
			log.Printf("❌ Failed to send welcome email: %v", err)
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Email verified successfully",
		"status":  "email_verified",
	})
}

// generateVerificationToken generates a secure random token for email verification
func generateVerificationToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
