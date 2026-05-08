package notifications

import (
	"encoding/json"
	"net/http"
	"ticketing_system/internal/middleware"
)

// Handler provides HTTP endpoints for notification operations
type Handler struct {
	service *NotificationService
}

// NewHandler creates a new notification handler
func NewHandler(service *NotificationService) *Handler {
	return &Handler{
		service: service,
	}
}

// TestEmailRequest represents a test email request
type TestEmailRequest struct {
	Email string `json:"email"`
}

// TestEmail handles testing email configuration
func (h *Handler) TestEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req TestEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "email is required")
		return
	}

	// Test email configuration
	if err := h.service.TestEmailConfiguration(req.Email); err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to send test email: "+err.Error())
		return
	}

	response := map[string]interface{}{
		"message": "Test email sent successfully",
		"email":   req.Email,
	}

	json.NewEncoder(w).Encode(response)
}

// SendWelcomeEmailRequest represents a welcome email request
type SendWelcomeEmailRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// SendWelcomeEmail handles sending welcome emails
func (h *Handler) SendWelcomeEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req SendWelcomeEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Name == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "email and name are required")
		return
	}

	if err := h.service.SendWelcomeEmail(req.Email, req.Name); err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to send welcome email")
		return
	}

	response := map[string]interface{}{
		"message": "Welcome email sent successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// SendVerificationEmailRequest represents a verification email request
type SendVerificationEmailRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Code  string `json:"code"`
}

// SendVerificationEmail handles sending verification emails
func (h *Handler) SendVerificationEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req SendVerificationEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Name == "" || req.Code == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "email, name, and code are required")
		return
	}

	if err := h.service.SendVerificationEmail(req.Email, req.Name, req.Code); err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to send verification email")
		return
	}

	response := map[string]interface{}{
		"message": "Verification email sent successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// SendPasswordResetRequest represents a password reset email request
type SendPasswordResetRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Token string `json:"token"`
}

// SendPasswordReset handles sending password reset emails
func (h *Handler) SendPasswordReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req SendPasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Name == "" || req.Token == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "email, name, and token are required")
		return
	}

	if err := h.service.SendPasswordResetEmail(req.Email, req.Name, req.Token); err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to send password reset email")
		return
	}

	response := map[string]interface{}{
		"message": "Password reset email sent successfully",
	}

	json.NewEncoder(w).Encode(response)
}
