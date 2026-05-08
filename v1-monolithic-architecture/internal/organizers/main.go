package organizers

import (
	"encoding/json"
	"net/http"
	"strings"
	"ticketing_system/internal/analytics"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"ticketing_system/internal/notifications"
	"ticketing_system/internal/security"
	"ticketing_system/internal/storage"
	"time"

	"gorm.io/gorm"
)

type OrganizerHandler struct {
	db            *gorm.DB
	_metrics      *analytics.PrometheusMetrics // Reserved for future instrumentation
	notifications *notifications.NotificationService
	encryption    *security.EncryptionService
	storage       *storage.StorageService
}

func NewOrganizerHandler(db *gorm.DB, metrics *analytics.PrometheusMetrics, notificationService *notifications.NotificationService, encryptionService *security.EncryptionService, storageService *storage.StorageService) *OrganizerHandler {
	return &OrganizerHandler{
		db:            db,
		_metrics:      metrics,
		notifications: notificationService,
		encryption:    encryptionService,
		storage:       storageService,
	}
}

// Request/Response structures
type OrganizerApplicationRequest struct {
	Name                string `json:"name"`
	About               string `json:"about"`
	Email               string `json:"email"`
	Phone               string `json:"phone"`
	Facebook            string `json:"facebook"`
	Twitter             string `json:"twitter"`
	TaxName             string `json:"tax_name"`
	TaxPin              string `json:"tax_pin"`
	PageHeaderBgColor   string `json:"page_header_bg_color"`
	PageBgColor         string `json:"page_bg_color"`
	PageTextColor       string `json:"page_text_color"`
	EnableOrganizerPage bool   `json:"enable_organizer_page"`
}

type OrganizerApplicationResponse struct {
	Message          string `json:"message"`
	OrganizerID      uint   `json:"organizer_id"`
	Status           string `json:"status"`
	RequiresApproval bool   `json:"requires_approval"`
}

// OrganizerApply handles organizer application/registration
func (h *OrganizerHandler) OrganizerApply(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get user ID from JWT token
	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request body
	var req OrganizerApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields
	if req.Name == "" || req.Email == "" || req.Phone == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "name, email, and phone are required")
		return
	}

	// Check if user exists and get their account
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Check if user already has an organizer profile
	var existingOrganizer models.Organizer
	if err := h.db.Where("account_id = ?", user.AccountID).First(&existingOrganizer).Error; err == nil {
		middleware.WriteJSONError(w, http.StatusConflict, "organizer profile already exists")
		return
	}

	// Check for duplicate email or business name
	var duplicateCheck models.Organizer
	if err := h.db.Where("email = ? OR name = ?", req.Email, req.Name).First(&duplicateCheck).Error; err == nil {
		middleware.WriteJSONError(w, http.StatusConflict, "email or business name already registered")
		return
	}

	// Create organizer profile
	organizer := models.Organizer{
		AccountID:           user.AccountID,
		Name:                strings.TrimSpace(req.Name),
		About:               strings.TrimSpace(req.About),
		Email:               strings.ToLower(strings.TrimSpace(req.Email)),
		Phone:               strings.TrimSpace(req.Phone),
		Facebook:            req.Facebook,
		Twitter:             req.Twitter,
		IsEmailConfirmed:    false, // Will be confirmed later
		ShowTwitterWidget:   false,
		ShowFacebookWidget:  false,
		TaxName:             strings.TrimSpace(req.TaxName),
		TaxPin:              strings.TrimSpace(req.TaxPin),
		PageHeaderBgColor:   req.PageHeaderBgColor,
		PageBgColor:         req.PageBgColor,
		PageTextColor:       req.PageTextColor,
		EnableOrganizerPage: req.EnableOrganizerPage,
	}

	// Save organizer to database
	if err := h.db.Create(&organizer).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to create organizer profile")
		return
	}

	// Update user role to organizer
	if err := h.db.Model(&user).Update("role", models.RoleOrganizer).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update user role")
		return
	}

	// Send confirmation email to organizer
	if h.notifications != nil {
		emailData := notifications.OrganizerApplicationConfirmationData{
			Name:  organizer.Name,
			Email: organizer.Email,
		}
		if err := h.notifications.SendOrganizerApplicationConfirmation(organizer.Email, emailData); err != nil {
			// Log error but don't fail the request
			middleware.WriteJSONError(w, http.StatusInternalServerError, "organizer created but failed to send confirmation email")
			return
		}
	}

	// Notify admins for approval
	if h.notifications != nil {
		var admins []models.User
		if err := h.db.Where("role = ?", models.RoleAdmin).Find(&admins).Error; err == nil {
			for _, admin := range admins {
				adminNotificationData := notifications.AdminOrganizerNotificationData{
					AdminName:      admin.FirstName + " " + admin.LastName,
					OrganizerName:  organizer.Name,
					OrganizerEmail: organizer.Email,
					OrganizerPhone: organizer.Phone,
					TaxName:        organizer.TaxName,
					TaxPin:         organizer.TaxPin,
					AppliedDate:    organizer.CreatedAt.Format("January 2, 2006"),
					ReviewURL:      h.notifications.GetAdminReviewURL(organizer.ID),
				}
				h.notifications.SendAdminOrganizerNotification(admin.Email, adminNotificationData)
			}
		}
	}

	response := OrganizerApplicationResponse{
		Message:          "Organizer application submitted successfully",
		OrganizerID:      organizer.ID,
		Status:           "pending_confirmation",
		RequiresApproval: true, // Set to true for manual approval workflow
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UpdateKYCStatusRequest for updating KYC status and notes
type UpdateKYCStatusRequest struct {
	KYCStatus string `json:"kyc_status"` // "pending", "scheduled", "completed", "failed"
	KYCNotes  string `json:"kyc_notes"`  // Admin notes from KYC process
}

// UpdateKYCStatus allows admins to update KYC status and add notes
func (h *OrganizerHandler) UpdateKYCStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get organizer ID from URL path
	organizerID := r.URL.Query().Get("id")
	if organizerID == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "organizer ID is required")
		return
	}

	// Parse request body
	var req UpdateKYCStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate KYC status
	validStatuses := map[string]bool{
		"pending":   true,
		"scheduled": true,
		"completed": true,
		"failed":    true,
	}
	if req.KYCStatus != "" && !validStatuses[req.KYCStatus] {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid KYC status")
		return
	}

	// Find organizer
	var organizer models.Organizer
	if err := h.db.Where("id = ?", organizerID).First(&organizer).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "organizer not found")
		return
	}

	// Update KYC fields
	updates := map[string]interface{}{}
	if req.KYCStatus != "" {
		updates["kyc_status"] = req.KYCStatus

		// Update verification status based on KYC status
		switch req.KYCStatus {
		case "completed":
			// KYC is done, admin will then use the verify endpoint to approve or reject
			updates["verification_status"] = "kyc_completed"
			now := time.Now().Format(time.RFC3339)
			updates["kyc_completed_at"] = now
		case "scheduled":
			updates["verification_status"] = "kyc_scheduled"
		case "failed":
			updates["verification_status"] = "kyc_failed"
		case "pending":
			updates["verification_status"] = "pending"
		}
	}
	if req.KYCNotes != "" {
		// Append to existing notes with separator
		existingNotes := organizer.KYCNotes
		if existingNotes != "" {
			existingNotes += "\n\n---\n\n"
		}
		updates["kyc_notes"] = existingNotes + req.KYCNotes
	}

	// Update in database
	if err := h.db.Model(&organizer).Updates(updates).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update KYC status")
		return
	}

	response := map[string]interface{}{
		"message":      "KYC status updated successfully",
		"kyc_status":   req.KYCStatus,
		"organizer_id": organizer.ID,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
