package orders

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"time"

	"github.com/gorilla/mux"
)

// UpdateOrderStatus handles updating the status of an order
func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get order ID from URL
	vars := mux.Vars(r)
	orderID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid order ID")
		return
	}

	// Parse request
	var req UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get order
	var order models.Order
	if err := h.db.Preload("Event").First(&order, orderID).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "order not found")
		return
	}

	// Check permissions
	if user.Role != models.RoleAdmin && user.Role != models.RoleOrganizer {
		middleware.WriteJSONError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	// If organizer, verify they own the event
	if user.Role == models.RoleOrganizer && order.Event.AccountID != user.AccountID {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Validate status transition
	if !isValidStatusTransition(order.Status, req.Status) {
		middleware.WriteJSONError(w, http.StatusBadRequest,
			fmt.Sprintf("cannot transition from %s to %s", order.Status, req.Status))
		return
	}

	// Update order status
	oldStatus := order.Status
	order.Status = req.Status
	if req.Status == models.OrderFulfilled && order.CompletedAt == nil {
		now := time.Now()
		order.CompletedAt = &now
	}

	if err := h.db.Save(&order).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update order status")
		return
	}

	// Track metrics for order completion
	if h.metrics != nil && req.Status == models.OrderFulfilled && oldStatus != models.OrderFulfilled {
		// Track revenue
		h.metrics.TrackRevenue(float64(order.Amount), order.Currency, fmt.Sprintf("%d", order.EventID), "")

		// Track completed order
		paymentMethod := "unknown"
		h.metrics.TrackOrderCompleted(paymentMethod, float64(order.Amount), order.Currency, time.Since(order.CreatedAt))
	}

	response := map[string]interface{}{
		"message": "Order status updated successfully",
		"order":   convertToOrderResponse(order),
	}

	json.NewEncoder(w).Encode(response)
}

// CancelOrder handles cancelling an order
func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get order ID from URL
	vars := mux.Vars(r)
	orderID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid order ID")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get order
	var order models.Order
	if err := h.db.Preload("OrderItems").First(&order, orderID).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "order not found")
		return
	}

	// Check ownership or admin
	if user.Role != models.RoleAdmin && order.AccountID != user.AccountID {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Check if order can be cancelled
	if order.Status == models.OrderCancelled {
		middleware.WriteJSONError(w, http.StatusBadRequest, "order is already cancelled")
		return
	}
	if order.Status == models.OrderFulfilled {
		middleware.WriteJSONError(w, http.StatusBadRequest, "cannot cancel fulfilled orders")
		return
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

	// Update order status
	order.Status = models.OrderCancelled
	order.IsCancelled = true
	now := time.Now()
	order.CancelledAt = &now

	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to cancel order")
		return
	}

	// Cancel all tickets for this order
	// First get the order item IDs
	var orderItemIDs []uint
	for _, item := range order.OrderItems {
		orderItemIDs = append(orderItemIDs, item.ID)
	}

	// Update tickets status to cancelled
	if len(orderItemIDs) > 0 {
		if err := tx.Model(&models.Ticket{}).
			Where("order_item_id IN ?", orderItemIDs).
			Update("status", models.TicketCancelled).Error; err != nil {
			tx.Rollback()
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to cancel tickets")
			return
		}
	}

	// Return tickets to inventory
	for _, item := range order.OrderItems {
		var ticketClass models.TicketClass
		if err := tx.First(&ticketClass, item.TicketClassID).Error; err != nil {
			tx.Rollback()
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update inventory")
			return
		}

		ticketClass.QuantitySold -= item.Quantity
		if err := tx.Save(&ticketClass).Error; err != nil {
			tx.Rollback()
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update inventory")
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to cancel order")
		return
	}
	committed = true

	// Track metrics
	if h.metrics != nil {
		h.metrics.OrdersFailed.WithLabelValues("cancelled").Inc()
	}

	response := map[string]interface{}{
		"message": "Order cancelled successfully",
		"order":   convertToOrderResponse(order),
	}

	json.NewEncoder(w).Encode(response)
}

// RefundOrder handles refunding an order
func (h *OrderHandler) RefundOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get order ID from URL
	vars := mux.Vars(r)
	orderID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid order ID")
		return
	}

	// Parse request
	var req RefundOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Only admins and organizers can process refunds
	if user.Role != models.RoleAdmin && user.Role != models.RoleOrganizer {
		middleware.WriteJSONError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	// Get order
	var order models.Order
	if err := h.db.Preload("Event").Preload("OrderItems").First(&order, orderID).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "order not found")
		return
	}

	// If organizer, verify they own the event
	if user.Role == models.RoleOrganizer && order.Event.AccountID != user.AccountID {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Check if order can be refunded
	if order.Status != models.OrderPaid && order.Status != models.OrderFulfilled {
		middleware.WriteJSONError(w, http.StatusBadRequest, "only paid or fulfilled orders can be refunded")
		return
	}

	// Calculate refund amount
	refundAmount := req.Amount
	if refundAmount == 0 || refundAmount > float64(order.Amount) {
		refundAmount = float64(order.Amount) // Full refund
	}

	// Create RefundRecord before processing
	refundNumber := fmt.Sprintf("REF-%d-%d", order.ID, time.Now().Unix())
	refundRecord := models.RefundRecord{
		RefundNumber:    refundNumber,
		RefundType:      models.RefundFull,
		RefundReason:    models.RefundReason(req.Reason),
		Status:          models.RefundProcessing,
		OrderID:         order.ID,
		EventID:         order.EventID,
		AccountID:       order.AccountID,
		OrganizerID:     order.Event.OrganizerID,
		OriginalAmount:  models.Money(order.TotalAmount),
		RefundAmount:    models.Money(refundAmount * 100), // Convert to cents
		OrganizerImpact: models.Money(refundAmount * 100),
		Currency:        order.Currency,
		RequestedBy:     &userID,
		RequestedAt:     time.Now(),
		Description:     req.Reason,
	}

	// Save refund record
	if err := h.db.Create(&refundRecord).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to create refund record")
		return
	}

	// Get payment record with external transaction ID (must be done before updating order)
	var paymentRecord models.PaymentRecord
	if err := h.db.Where("order_id = ? AND status = ?", order.ID, models.RecordCompleted).
		First(&paymentRecord).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "no completed payment found for order")
		return
	}

	if paymentRecord.ExternalTransactionID == nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "no external transaction ID for refund")
		return
	}

	// Initiate refund through Intasend
	if h.paymentHandler != nil {
		amountInCents := int64(refundAmount * 100)
		intasendResp, err := h.paymentHandler.InitiateIntasendRefund(
			*paymentRecord.ExternalTransactionID,
			amountInCents,
			req.Reason,
		)
		if err != nil {
			// Mark refund as failed
			refundRecord.Status = models.RefundFailed
			failedAt := time.Now()
			refundRecord.FailedAt = &failedAt
			h.db.Save(&refundRecord)

			middleware.WriteJSONError(w, http.StatusInternalServerError,
				fmt.Sprintf("failed to initiate refund with payment provider: %v", err))
			return
		}

		// Update refund record with external ID
		if intasendResp != nil && intasendResp.ID != "" {
			refundRecord.ExternalRefundID = &intasendResp.ID
		}
	}

	// Mark refund as completed
	now := time.Now()
	refundRecord.Status = models.RefundCompleted
	refundRecord.ProcessedAt = &now
	refundRecord.CompletedAt = &now
	refundRecord.ApprovedBy = &userID
	refundRecord.ApprovedAt = &now
	h.db.Save(&refundRecord)

	// Update order status after successful refund initiation
	order.Status = models.OrderRefunded
	amountRefunded := float32(refundAmount)
	order.AmountRefunded = &amountRefunded
	refundedAt := time.Now()
	order.RefundedAt = &refundedAt

	if refundAmount < float64(order.Amount) {
		order.Status = models.OrderPartialRefund
		order.IsPartiallyRefunded = true
	}

	if err := h.db.Save(&order).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update order after refund")
		return
	}

	// Update all tickets for this order to refunded status
	// First get the order item IDs
	var orderItemIDs []uint
	for _, item := range order.OrderItems {
		orderItemIDs = append(orderItemIDs, item.ID)
	}

	// Update tickets status to refunded
	if len(orderItemIDs) > 0 {
		if err := h.db.Model(&models.Ticket{}).
			Where("order_item_id IN ?", orderItemIDs).
			Update("status", models.TicketRefunded).Error; err != nil {
			// Log error but don't fail the refund
			fmt.Printf("Warning: Failed to update ticket statuses: %v\n", err)
		}
	}

	// Track metrics
	if h.metrics != nil {
		h.metrics.RefundsTotal.WithLabelValues(order.Currency, req.Reason).Add(refundAmount)
	}

	// Send refund notification email to customer
	if h.notificationService != nil && req.NotifyCustomer {
		go h.sendRefundNotificationEmail(&order, refundAmount, req.Reason)
	}

	response := map[string]interface{}{
		"message":       "Order refunded successfully",
		"order":         convertToOrderResponse(order),
		"refund_amount": refundAmount,
	}

	json.NewEncoder(w).Encode(response)
}

// isValidStatusTransition checks if a status transition is valid
func isValidStatusTransition(currentStatus, newStatus models.OrderStatus) bool {
	validTransitions := map[models.OrderStatus][]models.OrderStatus{
		models.OrderPending: {
			models.OrderPaid,
			models.OrderCancelled,
		},
		models.OrderPaid: {
			models.OrderFulfilled,
			models.OrderRefunded,
			models.OrderCancelled,
		},
		models.OrderFulfilled: {
			models.OrderRefunded,
		},
		models.OrderCancelled:     {}, // Terminal state
		models.OrderRefunded:      {}, // Terminal state
		models.OrderPartialRefund: {}, // Terminal state
	}

	allowedStatuses, exists := validTransitions[currentStatus]
	if !exists {
		return false
	}

	for _, status := range allowedStatuses {
		if status == newStatus {
			return true
		}
	}

	return false
}

// sendRefundNotificationEmail sends refund confirmation email to customer
func (h *OrderHandler) sendRefundNotificationEmail(order *models.Order, refundAmount float64, reason string) {
	if h.notificationService == nil {
		return
	}

	// Prepare email body
	emailBody := fmt.Sprintf(`
Dear Customer,

Your refund has been successfully processed.

Order Details:
- Order ID: #%d
- Refund Amount: %.2f %s
- Reason: %s
- Event: %s

The refund will be credited to your original payment method within 3-5 business days.

If you have any questions, please contact support.

Best regards,
The Ticketing Team
`, order.ID, refundAmount, order.Currency, reason, order.Event.Title)

	// Send email
	if err := h.notificationService.SendPlainEmail(
		[]string{order.Email},
		fmt.Sprintf("Refund Processed - Order #%d", order.ID),
		emailBody,
	); err != nil {
		// Log error but don't fail the refund
		fmt.Printf("Failed to send refund notification email: %v\n", err)
	}
}
