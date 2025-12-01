package refunds

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ticketing_system/internal/models"

	"github.com/gorilla/mux"
)

// ListPendingRefunds returns all refunds awaiting approval (admin/organizer only)
func (h *RefundHandler) ListPendingRefunds(w http.ResponseWriter, r *http.Request) {
	// Get user ID and role from context (set by auth middleware)
	userRole, ok := r.Context().Value("role").(string)
	if !ok || (userRole != "admin" && userRole != "organizer") {
		writeError(w, http.StatusForbidden, "Access denied")
		return
	}

	organizerID, _ := r.Context().Value("organizer_id").(uint)

	// Query pending refunds
	query := h.db.Where("status = ?", models.RefundRequested)

	// If organizer, only show their events' refunds
	if userRole == "organizer" && organizerID != 0 {
		query = query.Where("organizer_id = ?", organizerID)
	}

	var refunds []models.RefundRecord
	if err := query.Preload("Order").Preload("Account").Preload("Event").Order("requested_at ASC").Find(&refunds).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch pending refunds")
		return
	}

	summaries := []RefundSummary{}
	for _, refund := range refunds {
		summaries = append(summaries, RefundSummary{
			ID:           refund.ID,
			RefundNumber: refund.RefundNumber,
			OrderID:      refund.OrderID,
			Status:       string(refund.Status),
			RefundType:   string(refund.RefundType),
			Amount:       int64(refund.RefundAmount),
			Currency:     refund.Currency,
			RequestedAt:  refund.RequestedAt.Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, RefundListResponse{
		Refunds: summaries,
		Total:   len(summaries),
	})
}

// ApproveRefund allows admin/organizer to approve or reject a refund request
func (h *RefundHandler) ApproveRefund(w http.ResponseWriter, r *http.Request) {
	// Get user ID and role from context
	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userRole, ok := r.Context().Value("role").(string)
	if !ok || (userRole != "admin" && userRole != "organizer") {
		writeError(w, http.StatusForbidden, "Access denied")
		return
	}

	organizerID, _ := r.Context().Value("organizer_id").(uint)

	// Get refund ID from URL
	vars := mux.Vars(r)
	refundID := vars["id"]

	var req RefundApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// If rejecting, require rejection reason
	if !req.Approved && (req.RejectionReason == nil || *req.RejectionReason == "") {
		writeError(w, http.StatusBadRequest, "Rejection reason is required when rejecting refund")
		return
	}

	// Fetch refund
	var refund models.RefundRecord
	if err := h.db.Preload("Order").Preload("Event").First(&refund, refundID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Refund not found")
		return
	}

	// Verify organizer can only approve their own events' refunds
	if userRole == "organizer" && refund.OrganizerID != organizerID {
		writeError(w, http.StatusForbidden, "You can only approve refunds for your own events")
		return
	}

	// Check if refund is still pending
	if refund.Status != models.RefundRequested {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Refund is already %s", refund.Status))
		return
	}

	// Update refund based on approval decision
	now := time.Now()
	refund.ApprovedBy = &userID
	refund.ApprovedAt = &now
	refund.InternalNotes = req.InternalNotes

	if req.Approved {
		refund.Status = models.RefundApproved

		// Calculate organizer impact by deducting platform fees
		// Organizer only loses the net amount they would have received
		// Platform fees are not returned to organizer since they were platform's revenue
		var platformFeeAmount models.Money

		// Query platform fees from payment records for this order
		if err := h.db.Model(&models.PaymentRecord{}).
			Where("order_id = ? AND type = ?", refund.OrderID, models.RecordPlatformFee).
			Select("COALESCE(SUM(amount), 0)").
			Scan(&platformFeeAmount).Error; err != nil {
			// If can't determine platform fees, use a conservative estimate (5% default)
			platformFeeAmount = models.Money(float64(refund.RefundAmount) * 0.05)
		}

		// Calculate proportional platform fee for this refund
		// If refunding part of the order, calculate proportional fee
		proportionalPlatformFee := models.Money(0)
		if refund.OriginalAmount > 0 {
			// Proportional fee = (refund amount / original amount) * total platform fees
			proportion := float64(refund.RefundAmount) / float64(refund.OriginalAmount)
			proportionalPlatformFee = models.Money(float64(platformFeeAmount) * proportion)
		}

		// Organizer impact = refund amount - platform fees
		// This means organizer only loses what they would have received (net amount)
		refund.OrganizerImpact = refund.RefundAmount - proportionalPlatformFee

		// Ensure organizer impact is not negative
		if refund.OrganizerImpact < 0 {
			refund.OrganizerImpact = 0
		}
	} else {
		refund.Status = models.RefundRejected
		refund.RejectionReason = req.RejectionReason
		refund.FailedAt = &now
	}

	if err := h.db.Save(&refund).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update refund status")
		return
	}

	// Send notification to customer about approval/rejection
	if req.Approved {
		go h.sendRefundApprovedEmail(&refund)
	} else {
		go h.sendRefundRejectedEmail(&refund)
	}

	message := "Refund request approved successfully"
	if !req.Approved {
		message = "Refund request rejected"
	}

	writeJSON(w, http.StatusOK, RefundResponse{
		Success:      true,
		RefundID:     refund.ID,
		RefundNumber: refund.RefundNumber,
		Status:       string(refund.Status),
		Amount:       int64(refund.RefundAmount),
		Message:      message,
	})
}

// GetRefundDetails returns detailed information about a refund (admin/organizer)
func (h *RefundHandler) GetRefundDetails(w http.ResponseWriter, r *http.Request) {
	// Get user role from context
	userRole, ok := r.Context().Value("role").(string)
	if !ok || (userRole != "admin" && userRole != "organizer") {
		writeError(w, http.StatusForbidden, "Access denied")
		return
	}

	organizerID, _ := r.Context().Value("organizer_id").(uint)

	// Get refund ID from URL
	vars := mux.Vars(r)
	refundID := vars["id"]

	var refund models.RefundRecord
	if err := h.db.Preload("RefundLineItems").Preload("Order").Preload("Account").Preload("Event").First(&refund, refundID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Refund not found")
		return
	}

	// Verify organizer can only see their own events' refunds
	if userRole == "organizer" && refund.OrganizerID != organizerID {
		writeError(w, http.StatusForbidden, "Access denied")
		return
	}

	// Build line items response
	lineItems := []RefundLineItemDetail{}
	for _, item := range refund.RefundLineItems {
		lineItems = append(lineItems, RefundLineItemDetail{
			OrderItemID: item.OrderItemID,
			TicketID:    item.TicketID,
			Quantity:    item.Quantity,
			Amount:      int64(item.RefundAmount),
			Description: item.Description,
		})
	}

	response := RefundStatusResponse{
		RefundNumber:     refund.RefundNumber,
		Status:           string(refund.Status),
		RefundType:       string(refund.RefundType),
		RefundReason:     string(refund.RefundReason),
		OriginalAmount:   int64(refund.OriginalAmount),
		RefundAmount:     int64(refund.RefundAmount),
		Currency:         refund.Currency,
		RequestedAt:      refund.RequestedAt.Format(time.RFC3339),
		ExternalRefundID: refund.ExternalRefundID,
		LineItems:        lineItems,
	}

	if refund.ApprovedAt != nil {
		approvedAt := refund.ApprovedAt.Format(time.RFC3339)
		response.ApprovedAt = &approvedAt
	}

	if refund.ProcessedAt != nil {
		processedAt := refund.ProcessedAt.Format(time.RFC3339)
		response.ProcessedAt = &processedAt
	}

	if refund.CompletedAt != nil {
		completedAt := refund.CompletedAt.Format(time.RFC3339)
		response.CompletedAt = &completedAt
	}

	writeJSON(w, http.StatusOK, response)
}

// ListRefundsByOrganizer returns all refunds for an organizer's events
func (h *RefundHandler) ListRefundsByOrganizer(w http.ResponseWriter, r *http.Request) {
	// Get organizer ID from context
	organizerID, ok := r.Context().Value("organizer_id").(uint)
	if !ok {
		writeError(w, http.StatusForbidden, "Access denied")
		return
	}

	// Optional status filter
	status := r.URL.Query().Get("status")

	query := h.db.Where("organizer_id = ?", organizerID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var refunds []models.RefundRecord
	if err := query.Order("requested_at DESC").Find(&refunds).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch refunds")
		return
	}

	summaries := []RefundSummary{}
	for _, refund := range refunds {
		summaries = append(summaries, RefundSummary{
			ID:           refund.ID,
			RefundNumber: refund.RefundNumber,
			OrderID:      refund.OrderID,
			Status:       string(refund.Status),
			RefundType:   string(refund.RefundType),
			Amount:       int64(refund.RefundAmount),
			Currency:     refund.Currency,
			RequestedAt:  refund.RequestedAt.Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, RefundListResponse{
		Refunds: summaries,
		Total:   len(summaries),
	})
}
