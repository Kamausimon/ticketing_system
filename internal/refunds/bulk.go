package refunds

import (
	"encoding/json"
	"fmt"
	"net/http"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"time"
)

// BulkRefundRequest represents a bulk refund request
type BulkRefundRequest struct {
	RefundIDs []uint `json:"refund_ids"`
	Action    string `json:"action"` // "approve" or "reject"
	Reason    string `json:"reason,omitempty"`
}

// BulkRefundResponse represents the result of a bulk refund operation
type BulkRefundResponse struct {
	TotalProcessed int                `json:"total_processed"`
	TotalSucceeded int                `json:"total_succeeded"`
	TotalFailed    int                `json:"total_failed"`
	Results        []RefundBulkResult `json:"results"`
	Message        string             `json:"message"`
}

// RefundBulkResult represents the result of a single refund operation
type RefundBulkResult struct {
	RefundID     uint    `json:"refund_id"`
	OrderID      uint    `json:"order_id"`
	Status       string  `json:"status"` // "success" or "failed"
	Error        string  `json:"error,omitempty"`
	RefundAmount float64 `json:"refund_amount,omitempty"`
}

// ProcessBulkRefunds processes multiple refunds at once
func (h *RefundHandler) ProcessBulkRefunds(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request
	var req BulkRefundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate input
	if len(req.RefundIDs) == 0 {
		middleware.WriteJSONError(w, http.StatusBadRequest, "refund_ids is required")
		return
	}

	if req.Action != "approve" && req.Action != "reject" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "action must be 'approve' or 'reject'")
		return
	}

	// Get user and verify organizer status
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Process each refund
	var results []RefundBulkResult
	totalSucceeded := 0
	totalFailed := 0

	for _, refundID := range req.RefundIDs {
		result := h.processSingleRefund(refundID, user.AccountID, userID, req.Action, req.Reason)
		results = append(results, result)

		if result.Status == "success" {
			totalSucceeded++
		} else {
			totalFailed++
		}
	}

	response := BulkRefundResponse{
		TotalProcessed: len(req.RefundIDs),
		TotalSucceeded: totalSucceeded,
		TotalFailed:    totalFailed,
		Results:        results,
		Message:        fmt.Sprintf("Processed %d refunds: %d succeeded, %d failed", len(req.RefundIDs), totalSucceeded, totalFailed),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// processSingleRefund processes a single refund as part of bulk operation
func (h *RefundHandler) processSingleRefund(refundID uint, accountID uint, userID uint, action string, reason string) RefundBulkResult {
	result := RefundBulkResult{
		RefundID: refundID,
		Status:   "failed",
	}

	// Get refund with related data
	var refund models.RefundRecord
	if err := h.db.Preload("Order.Event").
		Where("id = ?", refundID).
		First(&refund).Error; err != nil {
		result.Error = "refund not found"
		return result
	}

	result.OrderID = refund.OrderID
	result.RefundAmount = float64(refund.RefundAmount) / 100 // Convert from cents

	// Verify ownership
	if refund.Order.Event.AccountID != accountID {
		result.Error = "access denied"
		return result
	}

	// Check if refund is in pending status
	if refund.Status != models.RefundRequested {
		result.Error = fmt.Sprintf("refund status is '%s', not pending", refund.Status)
		return result
	}

	// Process based on action
	switch action {
	case "approve":
		// Update refund status
		refund.Status = models.RefundApproved
		refund.ApprovedAt = timePtr(time.Now())
		refund.ApprovedBy = &userID

		if err := h.db.Save(&refund).Error; err != nil {
			result.Error = "failed to update refund status"
			return result
		}

		// Send approval email
		if h.notificationService != nil {
			h.sendRefundApprovedEmail(&refund)
		}

		result.Status = "success"

	case "reject":
		// Update refund status
		refund.Status = models.RefundRejected
		refund.ProcessedAt = timePtr(time.Now())
		if reason != "" {
			refund.RejectionReason = &reason
		}

		if err := h.db.Save(&refund).Error; err != nil {
			result.Error = "failed to update refund status"
			return result
		}

		// Send rejection email
		if h.notificationService != nil {
			h.sendRefundRejectedEmail(&refund)
		}

		result.Status = "success"

	default:
		result.Error = fmt.Sprintf("invalid action: %s", action)
	}

	return result
}

// BulkAutoApproveRefunds automatically approves eligible refunds
type BulkAutoApproveRequest struct {
	EventID         uint    `json:"event_id"`
	MaxRefundAmount float64 `json:"max_refund_amount,omitempty"`
	DaysBeforeEvent int     `json:"days_before_event,omitempty"` // Only auto-approve if event is X days away
}

// AutoApproveBulkRefunds automatically approves eligible refunds based on criteria
func (h *RefundHandler) AutoApproveBulkRefunds(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request
	var req BulkAutoApproveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate input
	if req.EventID == 0 {
		middleware.WriteJSONError(w, http.StatusBadRequest, "event_id is required")
		return
	}

	// Get user and verify organizer status
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Verify user owns the event
	var event models.Event
	if err := h.db.Where("id = ? AND account_id = ?", req.EventID, user.AccountID).First(&event).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied or event not found")
		return
	}

	// Build query for eligible refunds
	query := h.db.Model(&models.RefundRecord{}).
		Joins("JOIN orders ON orders.id = refund_records.order_id").
		Where("orders.event_id = ? AND refund_records.status = ?", req.EventID, models.RefundRequested)

	// Apply max refund amount filter
	if req.MaxRefundAmount > 0 {
		query = query.Where("refund_records.amount <= ?", int64(req.MaxRefundAmount*100))
	}

	// Apply days before event filter
	if req.DaysBeforeEvent > 0 {
		cutoffDate := event.StartDate.AddDate(0, 0, -req.DaysBeforeEvent)
		if time.Now().Before(cutoffDate) {
			// Event is still far enough away
			query = query.Where("refund_records.created_at <= ?", time.Now())
		} else {
			// Event is too close, don't auto-approve any
			middleware.WriteJSONError(w, http.StatusBadRequest,
				fmt.Sprintf("event is within %d days, auto-approval not allowed", req.DaysBeforeEvent))
			return
		}
	}

	// Get eligible refunds
	var refunds []models.RefundRecord
	if err := query.Preload("Order.Event").Find(&refunds).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch refunds")
		return
	}

	if len(refunds) == 0 {
		response := BulkRefundResponse{
			TotalProcessed: 0,
			TotalSucceeded: 0,
			TotalFailed:    0,
			Message:        "no eligible refunds found for auto-approval",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Process each refund
	var results []RefundBulkResult
	totalSucceeded := 0
	totalFailed := 0

	for _, refund := range refunds {
		result := h.processSingleRefund(refund.ID, user.AccountID, userID, "approve", "Auto-approved")
		results = append(results, result)

		if result.Status == "success" {
			totalSucceeded++
		} else {
			totalFailed++
		}
	}

	response := BulkRefundResponse{
		TotalProcessed: len(refunds),
		TotalSucceeded: totalSucceeded,
		TotalFailed:    totalFailed,
		Results:        results,
		Message:        fmt.Sprintf("Auto-approved %d refunds, %d failed", totalSucceeded, totalFailed),
	}

	json.NewEncoder(w).Encode(response)
}

// BulkRefundStats provides statistics about refunds for an event
type BulkRefundStats struct {
	EventID             uint    `json:"event_id"`
	TotalRefunds        int     `json:"total_refunds"`
	PendingRefunds      int     `json:"pending_refunds"`
	ApprovedRefunds     int     `json:"approved_refunds"`
	RejectedRefunds     int     `json:"rejected_refunds"`
	CompletedRefunds    int     `json:"completed_refunds"`
	TotalRefundAmount   float64 `json:"total_refund_amount"`
	PendingRefundAmount float64 `json:"pending_refund_amount"`
}

// GetBulkRefundStats returns statistics about refunds for an event
func (h *RefundHandler) GetBulkRefundStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get event ID from query
	eventIDStr := r.URL.Query().Get("event_id")
	if eventIDStr == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "event_id is required")
		return
	}

	var eventID uint
	if _, err := fmt.Sscanf(eventIDStr, "%d", &eventID); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid event_id")
		return
	}

	// Get user and verify organizer status
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Verify user owns the event
	var event models.Event
	if err := h.db.Where("id = ? AND account_id = ?", eventID, user.AccountID).First(&event).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied or event not found")
		return
	}

	// Get refund statistics
	var stats BulkRefundStats
	stats.EventID = eventID

	// Total refunds
	var total int64
	h.db.Model(&models.RefundRecord{}).
		Joins("JOIN orders ON orders.id = refund_records.order_id").
		Where("orders.event_id = ?", eventID).
		Count(&total)
	stats.TotalRefunds = int(total)

	// Count by status
	var pending, approved, rejected, completed int64
	h.db.Model(&models.RefundRecord{}).
		Joins("JOIN orders ON orders.id = refund_records.order_id").
		Where("orders.event_id = ? AND refund_records.status = ?", eventID, models.RefundRequested).
		Count(&pending)
	stats.PendingRefunds = int(pending)

	h.db.Model(&models.RefundRecord{}).
		Joins("JOIN orders ON orders.id = refund_records.order_id").
		Where("orders.event_id = ? AND refund_records.status = ?", eventID, models.RefundApproved).
		Count(&approved)
	stats.ApprovedRefunds = int(approved)

	h.db.Model(&models.RefundRecord{}).
		Joins("JOIN orders ON orders.id = refund_records.order_id").
		Where("orders.event_id = ? AND refund_records.status = ?", eventID, models.RefundRejected).
		Count(&rejected)
	stats.RejectedRefunds = int(rejected)

	h.db.Model(&models.RefundRecord{}).
		Joins("JOIN orders ON orders.id = refund_records.order_id").
		Where("orders.event_id = ? AND refund_records.status = ?", eventID, models.RefundCompleted).
		Count(&completed)
	stats.CompletedRefunds = int(completed)

	// Total refund amounts (convert from cents to currency units)
	var totalCents, pendingCents int64
	h.db.Model(&models.RefundRecord{}).
		Joins("JOIN orders ON orders.id = refund_records.order_id").
		Where("orders.event_id = ?", eventID).
		Select("COALESCE(SUM(refund_records.amount), 0)").
		Scan(&totalCents)
	stats.TotalRefundAmount = float64(totalCents) / 100

	h.db.Model(&models.RefundRecord{}).
		Joins("JOIN orders ON orders.id = refund_records.order_id").
		Where("orders.event_id = ? AND refund_records.status = ?", eventID, models.RefundRequested).
		Select("COALESCE(SUM(refund_records.amount), 0)").
		Scan(&pendingCents)
	stats.PendingRefundAmount = float64(pendingCents) / 100

	json.NewEncoder(w).Encode(stats)
}

// timePtr returns a pointer to a time value
func timePtr(t time.Time) *time.Time {
	return &t
}
