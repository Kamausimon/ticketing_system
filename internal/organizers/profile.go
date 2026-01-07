package organizers

import (
	"encoding/json"
	"net/http"
	"strings"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"time"
)

// Profile-related request/response structures
type OrganizerProfileResponse struct {
	ID                  uint      `json:"id"`
	Name                string    `json:"name"`
	About               string    `json:"about"`
	Email               string    `json:"email"`
	Phone               string    `json:"phone"`
	Facebook            string    `json:"facebook"`
	Twitter             string    `json:"twitter"`
	LogoPath            *string   `json:"logo_path"`
	IsEmailConfirmed    bool      `json:"is_email_confirmed"`
	ShowTwitterWidget   bool      `json:"show_twitter_widget"`
	ShowFacebookWidget  bool      `json:"show_facebook_widget"`
	TaxName             string    `json:"tax_name"`
	TaxPin              string    `json:"tax_pin"`
	PageHeaderBgColor   string    `json:"page_header_bg_color"`
	PageBgColor         string    `json:"page_bg_color"`
	PageTextColor       string    `json:"page_text_color"`
	EnableOrganizerPage bool      `json:"enable_organizer_page"`
	CreatedAt           time.Time `json:"created_at"`
}

type UpdateProfileRequest struct {
	Name                string `json:"name"`
	About               string `json:"about"`
	Phone               string `json:"phone"`
	Facebook            string `json:"facebook"`
	Twitter             string `json:"twitter"`
	TaxName             string `json:"tax_name"`
	TaxPin              string `json:"tax_pin"`
	PageHeaderBgColor   string `json:"page_header_bg_color"`
	PageBgColor         string `json:"page_bg_color"`
	PageTextColor       string `json:"page_text_color"`
	EnableOrganizerPage bool   `json:"enable_organizer_page"`
	ShowTwitterWidget   bool   `json:"show_twitter_widget"`
	ShowFacebookWidget  bool   `json:"show_facebook_widget"`
}

// GetOrganizerProfile returns the organizer's profile information
func (h *OrganizerHandler) GetOrganizerProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)

	// Get user and their organizer profile
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

	response := OrganizerProfileResponse{
		ID:                  organizer.ID,
		Name:                organizer.Name,
		About:               organizer.About,
		Email:               organizer.Email,
		Phone:               organizer.Phone,
		Facebook:            organizer.Facebook,
		Twitter:             organizer.Twitter,
		LogoPath:            organizer.LogoPath,
		IsEmailConfirmed:    organizer.IsEmailConfirmed,
		ShowTwitterWidget:   organizer.ShowTwitterWidget,
		ShowFacebookWidget:  organizer.ShowFacebookWidget,
		TaxName:             organizer.TaxName,
		TaxPin:              organizer.TaxPin,
		PageHeaderBgColor:   organizer.PageHeaderBgColor,
		PageBgColor:         organizer.PageBgColor,
		PageTextColor:       organizer.PageTextColor,
		EnableOrganizerPage: organizer.EnableOrganizerPage,
		CreatedAt:           organizer.CreatedAt,
	}

	json.NewEncoder(w).Encode(response)
}

// UpdateOrganizerProfile updates organizer profile information
func (h *OrganizerHandler) UpdateOrganizerProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get user and their organizer profile
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

	// Validate required fields
	if req.Name == "" || req.Phone == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "name and phone are required")
		return
	}

	// Update organizer fields
	updates := map[string]interface{}{
		"name":                  strings.TrimSpace(req.Name),
		"about":                 strings.TrimSpace(req.About),
		"phone":                 strings.TrimSpace(req.Phone),
		"facebook":              req.Facebook,
		"twitter":               req.Twitter,
		"tax_name":              req.TaxName,
		"tax_pin":               req.TaxPin,
		"page_header_bg_color":  req.PageHeaderBgColor,
		"page_bg_color":         req.PageBgColor,
		"page_text_color":       req.PageTextColor,
		"enable_organizer_page": req.EnableOrganizerPage,
		"show_twitter_widget":   req.ShowTwitterWidget,
		"show_facebook_widget":  req.ShowFacebookWidget,
	}

	if err := h.db.Model(&organizer).Updates(updates).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update organizer profile")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Profile updated successfully",
	})
}

// UploadOrganizerLogo handles logo upload for organizer
func (h *OrganizerHandler) UploadOrganizerLogo(w http.ResponseWriter, r *http.Request) {
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

	// Parse multipart form (10 MB max)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "failed to parse form data")
		return
	}

	file, header, err := r.FormFile("logo")
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "logo file is required")
		return
	}
	defer file.Close()

	// Validate file type (only images)
	contentType := header.Header.Get("Content-Type")
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	if !allowedTypes[contentType] {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid file type. Allowed: jpg, png, gif, webp")
		return
	}

	// Validate file size (max 5MB)
	if header.Size > 5*1024*1024 {
		middleware.WriteJSONError(w, http.StatusBadRequest, "file size exceeds 5MB limit")
		return
	}

	// Upload file using storage service
	var logoPath string
	if h.storage != nil {
		result, err := h.storage.UploadFile(file, header, "logos")
		if err != nil {
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to upload logo: "+err.Error())
			return
		}
		logoPath = result.URL
	} else {
		// Fallback to placeholder if storage service not available
		logoPath = "/uploads/logos/placeholder_" + header.Filename
	}

	// Update organizer logo path
	if err := h.db.Model(&organizer).Update("logo_path", logoPath).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update logo")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"logo_url":  logoPath,
		"message":   "Logo uploaded successfully",
		"file_size": header.Size,
	})
}
