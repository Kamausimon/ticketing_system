package payments

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

// InitiateRefund initiates a refund for a payment
func (h *PaymentHandler) InitiateRefund(w http.ResponseWriter, r *http.Request) {
	var req RefundPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get payment record
	var paymentRecord models.PaymentRecord
	if err := h.db.First(&paymentRecord, req.PaymentRecordID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Payment record not found")
		return
	}

	if paymentRecord.Status != models.RecordCompleted {
		writeError(w, http.StatusBadRequest, "Can only refund completed payments")
		return
	}

	if paymentRecord.ExternalTransactionID == nil {
		writeError(w, http.StatusBadRequest, "No external transaction ID for refund")
		return
	}

	// Get order from payment record
	var order models.Order
	if err := h.db.Preload("Event").First(&order, paymentRecord.OrderID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Order not found")
		return
	}

	// Initiate refund via Intasend
	refundResp, err := h.InitiateIntasendRefund(*paymentRecord.ExternalTransactionID, req.Amount, req.Reason)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Refund initiation failed: %v", err))
		return
	}

	// Create refund record in database
	refundNumber := fmt.Sprintf("REF-%d-%s", time.Now().Unix(), refundResp.ID[:8])
	refundRecord := models.RefundRecord{
		RefundNumber:     refundNumber,
		RefundType:       models.RefundPartial,
		RefundReason:     models.RefundCustomerRequest,
		Status:           models.RefundProcessing,
		OrderID:          order.ID,
		EventID:          order.EventID,
		AccountID:        order.AccountID,
		OrganizerID:      order.Event.OrganizerID,
		OriginalAmount:   models.Money(paymentRecord.Amount),
		RefundAmount:     models.Money(req.Amount * 100),
		OrganizerImpact:  models.Money(req.Amount * 100),
		Currency:         "KSH",
		ExternalRefundID: &refundResp.ID,
		RequestedAt:      time.Now(),
		Description:      req.Reason,
	}

	if err := h.db.Create(&refundRecord).Error; err != nil {
		// Log error but don't fail the request since refund was initiated
		fmt.Printf("Warning: Failed to create refund record: %v\n", err)
	}

	response := RefundResponse{
		Success:  true,
		RefundID: refundResp.ID,
		Amount:   req.Amount,
		Status:   refundResp.State,
		Message:  "Refund initiated successfully",
	}

	writeJSON(w, http.StatusOK, response)
}

// GetRefundStatus gets status of a refund
func (h *PaymentHandler) GetRefundStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	refundID := vars["id"]

	// Check database for refund record
	var refundRecord models.RefundRecord
	if err := h.db.Where("external_refund_id = ? OR refund_number = ?", refundID, refundID).First(&refundRecord).Error; err != nil {
		writeError(w, http.StatusNotFound, "refund not found")
		return
	}

	// If refund is still processing, check with Intasend API
	if refundRecord.Status == models.RefundProcessing || refundRecord.Status == models.RefundRequested {
		// In production, implement actual Intasend API status check
		// For now, return current status from database
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"refund_id": refundRecord.RefundNumber,
		"status":    "pending",
		"message":   "Refund status check not yet implemented",
	})
}

// ListRefunds lists refunds for an order
func (h *PaymentHandler) ListRefunds(w http.ResponseWriter, r *http.Request) {
	orderIDStr := r.URL.Query().Get("order_id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	var refunds []models.RefundRecord
	if err := h.db.Where("order_id = ?", orderID).
		Order("created_at DESC").
		Find(&refunds).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch refunds")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"refunds": refunds,
		"total":   len(refunds),
	})
}

// ApproveRefund approves a refund request (admin only)
func (h *PaymentHandler) ApproveRefund(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	refundID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid refund ID")
		return
	}

	var refundRecord models.RefundRecord
	if err := h.db.First(&refundRecord, refundID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Refund record not found")
		return
	}

	if refundRecord.Status != models.RefundRequested {
		writeError(w, http.StatusBadRequest, "Refund already processed")
		return
	}

	// Get authenticated user ID
	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Verify user is admin
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		writeError(w, http.StatusNotFound, "User not found")
		return
	}

	if user.Role != models.RoleAdmin {
		writeError(w, http.StatusForbidden, "Admin access required")
		return
	}

	// Update refund record
	now := time.Now()
	refundRecord.Status = models.RefundApproved
	refundRecord.ApprovedAt = &now
	refundRecord.ApprovedBy = &userID

	if err := h.db.Save(&refundRecord).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to approve refund")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"refund_id": refundID,
		"status":    refundRecord.Status,
		"message":   "Refund approved successfully",
	})
}
