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

	"gorm.io/gorm"
)

type OrganizerHandler struct {
	db            *gorm.DB
	_metrics      *analytics.PrometheusMetrics // Reserved for future instrumentation
	notifications *notifications.NotificationService
	encryption    *security.EncryptionService
}

func NewOrganizerHandler(db *gorm.DB, metrics *analytics.PrometheusMetrics, notificationService *notifications.NotificationService, encryptionService *security.EncryptionService) *OrganizerHandler {
	return &OrganizerHandler{
		db:            db,
		_metrics:      metrics,
		notifications: notificationService,
		encryption:    encryptionService,
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
	userID := middleware.GetUserIDFromToken(r)

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
		TaxName:             req.TaxName,
		TaxPin:              req.TaxPin,
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
