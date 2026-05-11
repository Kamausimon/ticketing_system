package refunds

import (
	"encoding/json"
	"net/http"
	"time"

	apievents "ticketing_system/internal/api_events"
	kafkatopics "ticketing_system/internal/kafka"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"

	"github.com/gorilla/mux"
)

const (
	IntasendAPIBaseURL     = "https://api.intasend.com/api/v1"
	IntasendSandboxBaseURL = "https://sandbox.intasend.com/api/v1"
)

// ProcessRefund processes an approved refund through the payment gateway
func (h *RefundHandler) ProcessRefund(w http.ResponseWriter, r *http.Request) {
	// Get user ID from token
	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if userID == 0 {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get user details to check role
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		writeError(w, http.StatusNotFound, "User not found")
		return
	}

	userRole := string(user.Role)
	if userRole != "admin" && userRole != "organizer" {
		writeError(w, http.StatusForbidden, "Access denied")
		return
	}

	var organizerID uint
	if userRole == "organizer" {
		// Get organizer profile for this user
		var organizer models.Organizer
		if err := h.db.Where("user_id = ?", userID).First(&organizer).Error; err != nil {
			writeError(w, http.StatusForbidden, "Organizer profile not found")
			return
		}
		organizerID = organizer.ID
	}

	// Get refund ID from URL
	vars := mux.Vars(r)
	refundID := vars["id"]

	var refund models.RefundRecord
	if err := h.db.First(&refund, refundID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Refund not found")
		return
	}

	// Verify organizer can only process their own events' refunds
	if userRole == "organizer" && refund.OrganizerID != organizerID {
		writeError(w, http.StatusForbidden, "Access denied")
		return
	}

	// Check if refund is approved
	if refund.Status != models.RefundApproved {
		writeError(w, http.StatusBadRequest, "Refund must be approved before processing")
		return
	}

	// Publish refund.approved via outbox so the refund-processor-worker calls
	// Intasend asynchronously. The HTTP response is no longer held hostage by
	// a 30-second external API call.
	evt := apievents.RefundApprovedEvent{
		RefundID:     refund.ID,
		RefundNumber: refund.RefundNumber,
		Timestamp:    time.Now(),
	}
	payload, err := json.Marshal(evt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to queue refund")
		return
	}
	if err := h.outboxRepo.Save(h.db, models.OutboxEvent{
		Topic:   kafkatopics.RefundApprovedTopic,
		Payload: string(payload),
		Status:  "pending",
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to queue refund")
		return
	}

	writeJSON(w, http.StatusAccepted, RefundResponse{
		Success:      true,
		RefundID:     refund.ID,
		RefundNumber: refund.RefundNumber,
		Status:       string(refund.Status),
		Amount:       int64(refund.RefundAmount),
		Message:      "Refund queued for processing — the customer will be notified when complete",
	})
}


// RetryFailedRefund allows retrying a failed refund
func (h *RefundHandler) RetryFailedRefund(w http.ResponseWriter, r *http.Request) {
	// Get user ID from token
	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if userID == 0 {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get user details to check role (admin only)
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		writeError(w, http.StatusNotFound, "User not found")
		return
	}

	userRole := string(user.Role)
	if userRole != "admin" {
		writeError(w, http.StatusForbidden, "Access denied")
		return
	}

	// Get refund ID from URL
	vars := mux.Vars(r)
	refundID := vars["id"]

	var refund models.RefundRecord
	if err := h.db.First(&refund, refundID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Refund not found")
		return
	}

	if refund.Status != models.RefundFailed {
		writeError(w, http.StatusBadRequest, "Can only retry failed refunds")
		return
	}

	// Reset to approved so the worker picks it up cleanly.
	refund.Status = models.RefundApproved
	refund.FailedAt = nil
	if err := h.db.Save(&refund).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to reset refund status")
		return
	}

	evt := apievents.RefundApprovedEvent{
		RefundID:     refund.ID,
		RefundNumber: refund.RefundNumber,
		Timestamp:    time.Now(),
	}
	payload, err := json.Marshal(evt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to queue retry")
		return
	}
	if err := h.outboxRepo.Save(h.db, models.OutboxEvent{
		Topic:   kafkatopics.RefundApprovedTopic,
		Payload: string(payload),
		Status:  "pending",
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to queue retry")
		return
	}

	writeJSON(w, http.StatusAccepted, RefundResponse{
		Success:      true,
		RefundID:     refund.ID,
		RefundNumber: refund.RefundNumber,
		Status:       string(refund.Status),
		Amount:       int64(refund.RefundAmount),
		Message:      "Refund retry queued",
	})
}

// GetRefundStatistics returns refund statistics for an organizer or admin
func (h *RefundHandler) GetRefundStatistics(w http.ResponseWriter, r *http.Request) {
	// Get user ID from token
	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if userID == 0 {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get user details to check role
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		writeError(w, http.StatusNotFound, "User not found")
		return
	}

	userRole := string(user.Role)
	if userRole != "admin" && userRole != "organizer" {
		writeError(w, http.StatusForbidden, "Access denied")
		return
	}

	var organizerID uint
	if userRole == "organizer" {
		// Get organizer profile for this user
		var organizer models.Organizer
		if err := h.db.Where("user_id = ?", userID).First(&organizer).Error; err != nil {
			writeError(w, http.StatusForbidden, "Organizer profile not found")
			return
		}
		organizerID = organizer.ID
	}

	query := h.db.Model(&models.RefundRecord{})
	if userRole == "organizer" && organizerID != 0 {
		query = query.Where("organizer_id = ?", organizerID)
	}

	// Count by status
	var stats struct {
		TotalRequested  int64
		TotalApproved   int64
		TotalProcessing int64
		TotalCompleted  int64
		TotalRejected   int64
		TotalFailed     int64
		TotalAmount     int64
	}

	query.Where("status = ?", models.RefundRequested).Count(&stats.TotalRequested)
	query.Where("status = ?", models.RefundApproved).Count(&stats.TotalApproved)
	query.Where("status = ?", models.RefundProcessing).Count(&stats.TotalProcessing)
	query.Where("status = ?", models.RefundCompleted).Count(&stats.TotalCompleted)
	query.Where("status = ?", models.RefundRejected).Count(&stats.TotalRejected)
	query.Where("status = ?", models.RefundFailed).Count(&stats.TotalFailed)

	// Calculate total refunded amount
	var result struct {
		Total int64
	}
	query.Where("status = ?", models.RefundCompleted).Select("COALESCE(SUM(refund_amount), 0) as total").Scan(&result)
	stats.TotalAmount = result.Total

	writeJSON(w, http.StatusOK, stats)
}
