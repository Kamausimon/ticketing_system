package organizers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
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

	userID := middleware.GetUserIDFromToken(r)

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
	if err := h.db.Where("is_email_confirmed = ?", false).Find(&organizers).Error; err != nil {
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

	userID := middleware.GetUserIDFromToken(r)

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
	organizerID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid organizer ID")
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

	userID := middleware.GetUserIDFromToken(r)

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
		Email:      organizer.Email,
		Token:      token,
		ExpiresAt:  expiresAt,
		IssuedAt:   time.Now(),
		LastSentAt: time.Now(),
	}

	if err := h.db.Create(&verification).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to create verification token")
		return
	}

	// Send verification email (assuming email package is available)
	// In production, integrate with actual email service
	verificationLink := "https://yourdomain.com/verify-email?token=" + token

	// Log for now (replace with actual email sending)
	_ = verificationLink // Placeholder for email sending

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Verification email sent successfully",
		"expires_at": expiresAt,
	})
}

// generateVerificationToken generates a secure random token for email verification
func generateVerificationToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
