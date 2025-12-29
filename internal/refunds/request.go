package refunds

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ticketing_system/internal/models"

	"github.com/gorilla/mux"
)

// RequestRefund handles customer-initiated refund requests
func (h *RefundHandler) RequestRefund(w http.ResponseWriter, r *http.Request) {
	// Get account ID from context (set by auth middleware)
	accountID, ok := r.Context().Value("account_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req RefundRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.OrderID == 0 {
		writeError(w, http.StatusBadRequest, "Order ID is required")
		return
	}

	if req.RefundType == "" {
		writeError(w, http.StatusBadRequest, "Refund type is required")
		return
	}

	if req.RefundReason == "" {
		writeError(w, http.StatusBadRequest, "Refund reason is required")
		return
	}

	if req.Description == "" {
		writeError(w, http.StatusBadRequest, "Description is required")
		return
	}

	// For partial refunds, require line items
	if req.RefundType == models.RefundPartial || req.RefundType == models.RefundTicket {
		if len(req.LineItems) == 0 {
			writeError(w, http.StatusBadRequest, "Line items required for partial/ticket refunds")
			return
		}
	}

	// Fetch the order and verify ownership
	var order models.Order
	if err := h.db.Preload("OrderItems").Preload("Event").First(&order, req.OrderID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Order not found")
		return
	}

	if order.AccountID != accountID {
		writeError(w, http.StatusForbidden, "You don't have permission to refund this order")
		return
	}

	// Check if order is eligible for refund
	if order.Status != "paid" && order.Status != "completed" {
		writeError(w, http.StatusBadRequest, "Order must be paid or completed to request refund")
		return
	}

	// Check if order already has a pending/approved refund
	var existingRefund models.RefundRecord
	err := h.db.Where("order_id = ? AND status IN ?", order.ID, []string{"requested", "approved", "processing"}).First(&existingRefund).Error
	if err == nil {
		writeError(w, http.StatusConflict, "A refund is already pending for this order")
		return
	}

	// Anti-spam: Check for duplicate refund requests within 24 hours
	var recentRefund models.RefundRecord
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)
	err = h.db.Where("order_id = ? AND account_id = ? AND created_at > ?",
		order.ID, accountID, twentyFourHoursAgo).First(&recentRefund).Error
	if err == nil {
		writeError(w, http.StatusTooManyRequests,
			"You have already requested a refund for this order recently. Please wait before trying again.")
		return
	}

	// Prevent refunds for events that have already occurred
	if order.Event != nil && order.Event.EndDate.Before(time.Now()) {
		writeError(w, http.StatusBadRequest,
			"Cannot request refund for an event that has already ended")
		return
	}

	// Check if any tickets have been checked in (attended)
	var checkedInCount int64
	h.db.Model(&models.Ticket{}).
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Where("order_items.order_id = ? AND tickets.checked_in_at IS NOT NULL", order.ID).
		Count(&checkedInCount)

	if checkedInCount > 0 {
		writeError(w, http.StatusForbidden,
			"Cannot request refund because one or more tickets have been used/checked in at the event")
		return
	}

	// Calculate refund amount
	refundAmount := int64(0)
	if req.RefundType == models.RefundFull {
		refundAmount = int64(order.TotalAmount)
	} else {
		// Calculate from line items
		for _, item := range req.LineItems {
			refundAmount += item.Amount
		}
	}

	// Validate refund amount doesn't exceed order total
	if refundAmount > int64(order.TotalAmount) {
		writeError(w, http.StatusBadRequest, "Refund amount exceeds order total")
		return
	}

	// Create refund record
	refundNumber := fmt.Sprintf("REF-%d-%d", order.ID, time.Now().Unix())
	refund := models.RefundRecord{
		RefundNumber:    refundNumber,
		RefundType:      req.RefundType,
		RefundReason:    req.RefundReason,
		Status:          models.RefundRequested,
		OrderID:         order.ID,
		EventID:         order.EventID,
		AccountID:       accountID,
		OrganizerID:     order.Event.OrganizerID,
		OriginalAmount:  models.Money(order.TotalAmount),
		RefundAmount:    models.Money(refundAmount),
		OrganizerImpact: models.Money(refundAmount), // Will be adjusted during approval
		Currency:        order.Currency,
		RequestedBy:     &accountID,
		RequestedAt:     time.Now(),
		Description:     req.Description,
	}

	// Start transaction
	tx := h.db.Begin()
	committed := false
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		} else if !committed {
			tx.Rollback()
		}
	}()

	// Save refund record
	if err := tx.Create(&refund).Error; err != nil {
		tx.Rollback()
		writeError(w, http.StatusInternalServerError, "Failed to create refund request")
		return
	}

	// Create line items for partial refunds
	if req.RefundType == models.RefundPartial || req.RefundType == models.RefundTicket {
		for _, item := range req.LineItems {
			// Verify order item exists
			var orderItem models.OrderItem
			if err := tx.First(&orderItem, item.OrderItemID).Error; err != nil {
				tx.Rollback()
				writeError(w, http.StatusBadRequest, fmt.Sprintf("Order item %d not found", item.OrderItemID))
				return
			}

			lineItem := models.RefundLineItem{
				RefundRecordID: refund.ID,
				OrderItemID:    item.OrderItemID,
				TicketID:       item.TicketID,
				Quantity:       item.Quantity,
				RefundAmount:   models.Money(item.Amount),
				Reason:         &item.Reason,
				Description:    fmt.Sprintf("Refund for %d ticket(s)", item.Quantity),
			}

			if err := tx.Create(&lineItem).Error; err != nil {
				tx.Rollback()
				writeError(w, http.StatusInternalServerError, "Failed to create refund line items")
				return
			}
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to complete refund request")
		return
	}
	committed = true

	// Send notification to customer about refund request
	go h.sendRefundRequestedEmail(&refund, &order)

	// Send notification to organizer about pending refund
	go h.sendOrganizerRefundPendingEmail(&refund, &order)

	writeJSON(w, http.StatusCreated, RefundResponse{
		Success:      true,
		RefundID:     refund.ID,
		RefundNumber: refund.RefundNumber,
		Status:       string(refund.Status),
		Amount:       int64(refund.RefundAmount),
		Message:      "Refund request submitted successfully. It will be reviewed by the organizer.",
	})
}

// GetRefundStatus returns the status of a specific refund
func (h *RefundHandler) GetRefundStatus(w http.ResponseWriter, r *http.Request) {
	// Get account ID from context
	accountID, ok := r.Context().Value("account_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get refund ID from URL
	vars := mux.Vars(r)
	refundID := vars["id"]

	var refund models.RefundRecord
	if err := h.db.Preload("RefundLineItems").First(&refund, refundID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Refund not found")
		return
	}

	// Verify ownership (customer can only see their own refunds)
	if refund.AccountID != accountID {
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

// ListRefunds returns all refunds for the current account
func (h *RefundHandler) ListRefunds(w http.ResponseWriter, r *http.Request) {
	// Get account ID from context
	accountID, ok := r.Context().Value("account_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var refunds []models.RefundRecord
	if err := h.db.Where("account_id = ?", accountID).Order("created_at DESC").Find(&refunds).Error; err != nil {
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

// CancelRefundRequest allows customer to cancel a pending refund request
func (h *RefundHandler) CancelRefundRequest(w http.ResponseWriter, r *http.Request) {
	// Get account ID from context
	accountID, ok := r.Context().Value("account_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
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

	// Verify ownership
	if refund.AccountID != accountID {
		writeError(w, http.StatusForbidden, "Access denied")
		return
	}

	// Can only cancel if still in requested status
	if refund.Status != models.RefundRequested {
		writeError(w, http.StatusBadRequest, "Can only cancel refund requests that are still pending approval")
		return
	}

	// Update status to rejected with reason
	reason := "Cancelled by customer"
	now := time.Now()
	refund.Status = models.RefundRejected
	refund.RejectionReason = &reason
	refund.FailedAt = &now

	if err := h.db.Save(&refund).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to cancel refund request")
		return
	}

	writeJSON(w, http.StatusOK, RefundResponse{
		Success:      true,
		RefundID:     refund.ID,
		RefundNumber: refund.RefundNumber,
		Status:       string(refund.Status),
		Amount:       int64(refund.RefundAmount),
		Message:      "Refund request cancelled successfully",
	})
}
