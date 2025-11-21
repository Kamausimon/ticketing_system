package payments

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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
	if err := h.DB.First(&paymentRecord, req.PaymentRecordID).Error; err != nil {
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

	// Initiate refund via Intasend
	refundResp, err := h.InitiateIntasendRefund(*paymentRecord.ExternalTransactionID, req.Amount, req.Reason)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Refund initiation failed: %v", err))
		return
	}

	// Create refund record (TODO: implement refund record creation)
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

	// TODO: Implement refund status check via Intasend API
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"refund_id": refundID,
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
	if err := h.DB.Where("order_id = ?", orderID).
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
	if err := h.DB.First(&refundRecord, refundID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Refund record not found")
		return
	}

	if refundRecord.Status != models.RefundRequested {
		writeError(w, http.StatusBadRequest, "Refund already processed")
		return
	}

	now := time.Now()
	refundRecord.Status = models.RefundApproved
	refundRecord.ApprovedAt = &now
	// TODO: Set ApprovedBy from authenticated user

	if err := h.DB.Save(&refundRecord).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to approve refund")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"refund_id": refundID,
		"status":    refundRecord.Status,
		"message":   "Refund approved successfully",
	})
}
