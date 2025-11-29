package refunds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ticketing_system/internal/models"

	"github.com/gorilla/mux"
)

const (
	IntasendAPIBaseURL     = "https://api.intasend.com/api/v1"
	IntasendSandboxBaseURL = "https://sandbox.intasend.com/api/v1"
)

// ProcessRefund processes an approved refund through the payment gateway
func (h *RefundHandler) ProcessRefund(w http.ResponseWriter, r *http.Request) {
	// Get user role from context (admin/organizer only)
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
	if err := h.db.Preload("Order").Preload("Order.PaymentRecords").First(&refund, refundID).Error; err != nil {
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

	// Find the original payment record
	var paymentRecord models.PaymentRecord
	found := false
	for _, pr := range refund.Order.PaymentRecords {
		if pr.Status == models.RecordCompleted {
			paymentRecord = pr
			found = true
			break
		}
	}

	if !found {
		writeError(w, http.StatusBadRequest, "No completed payment found for this order")
		return
	}

	// Update refund status to processing
	now := time.Now()
	refund.Status = models.RefundProcessing
	refund.ProcessedAt = &now

	if err := h.db.Save(&refund).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update refund status")
		return
	}

	// Process refund through Intasend
	externalRefundID, err := h.initiateIntasendRefund(&paymentRecord, &refund)
	if err != nil {
		// Mark refund as failed
		refund.Status = models.RefundFailed
		failedAt := time.Now()
		refund.FailedAt = &failedAt
		h.db.Save(&refund)

		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to process refund: %v", err))
		return
	}

	// Update refund with external ID and mark as completed
	refund.ExternalRefundID = &externalRefundID
	refund.Status = models.RefundCompleted
	completedAt := time.Now()
	refund.CompletedAt = &completedAt

	if err := h.db.Save(&refund).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update refund status")
		return
	}

	// Update order status
	refund.Order.Status = "refunded"
	h.db.Save(&refund.Order)

	// Send notification to customer about refund completion
	go h.sendRefundCompletedEmail(&refund)

	writeJSON(w, http.StatusOK, RefundResponse{
		Success:      true,
		RefundID:     refund.ID,
		RefundNumber: refund.RefundNumber,
		Status:       string(refund.Status),
		Amount:       int64(refund.RefundAmount),
		Message:      "Refund processed successfully",
	})
}

// initiateIntasendRefund sends refund request to Intasend API
func (h *RefundHandler) initiateIntasendRefund(payment *models.PaymentRecord, refund *models.RefundRecord) (string, error) {
	// Determine API URL based on test mode
	baseURL := IntasendAPIBaseURL
	if h.IntasendTestMode {
		baseURL = IntasendSandboxBaseURL
	}

	// Prepare refund request
	refundData := map[string]interface{}{
		"invoice_id": payment.ExternalTransactionID, // Intasend transaction ID
		"amount":     float64(refund.RefundAmount) / 100.0,
		"reason":     string(refund.RefundReason),
	}

	jsonData, err := json.Marshal(refundData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal refund data: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/refunds/", baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.IntasendSecretKey))

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("refund failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract refund ID
	refundID, ok := result["id"].(string)
	if !ok {
		// Try tracking_id or transaction_id as fallback
		if trackingID, ok := result["tracking_id"].(string); ok {
			refundID = trackingID
		} else {
			refundID = fmt.Sprintf("REF-%d", time.Now().Unix())
		}
	}

	return refundID, nil
}

// RetryFailedRefund allows retrying a failed refund
func (h *RefundHandler) RetryFailedRefund(w http.ResponseWriter, r *http.Request) {
	// Get user role from context (admin only)
	userRole, ok := r.Context().Value("role").(string)
	if !ok || userRole != "admin" {
		writeError(w, http.StatusForbidden, "Access denied")
		return
	}

	// Get refund ID from URL
	vars := mux.Vars(r)
	refundID := vars["id"]

	var refund models.RefundRecord
	if err := h.db.Preload("Order").Preload("Order.PaymentRecords").First(&refund, refundID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Refund not found")
		return
	}

	// Check if refund is failed
	if refund.Status != models.RefundFailed {
		writeError(w, http.StatusBadRequest, "Can only retry failed refunds")
		return
	}

	// Find the original payment record
	var paymentRecord models.PaymentRecord
	found := false
	for _, pr := range refund.Order.PaymentRecords {
		if pr.Status == models.RecordCompleted {
			paymentRecord = pr
			found = true
			break
		}
	}

	if !found {
		writeError(w, http.StatusBadRequest, "No completed payment found for this order")
		return
	}

	// Reset status to processing
	now := time.Now()
	refund.Status = models.RefundProcessing
	refund.ProcessedAt = &now
	refund.FailedAt = nil

	if err := h.db.Save(&refund).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update refund status")
		return
	}

	// Process refund through Intasend
	externalRefundID, err := h.initiateIntasendRefund(&paymentRecord, &refund)
	if err != nil {
		// Mark refund as failed again
		refund.Status = models.RefundFailed
		failedAt := time.Now()
		refund.FailedAt = &failedAt
		h.db.Save(&refund)

		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to process refund: %v", err))
		return
	}

	// Update refund with external ID and mark as completed
	refund.ExternalRefundID = &externalRefundID
	refund.Status = models.RefundCompleted
	completedAt := time.Now()
	refund.CompletedAt = &completedAt

	if err := h.db.Save(&refund).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update refund status")
		return
	}

	writeJSON(w, http.StatusOK, RefundResponse{
		Success:      true,
		RefundID:     refund.ID,
		RefundNumber: refund.RefundNumber,
		Status:       string(refund.Status),
		Amount:       int64(refund.RefundAmount),
		Message:      "Refund processed successfully",
	})
}

// GetRefundStatistics returns refund statistics for an organizer or admin
func (h *RefundHandler) GetRefundStatistics(w http.ResponseWriter, r *http.Request) {
	// Get user role from context
	userRole, ok := r.Context().Value("role").(string)
	if !ok || (userRole != "admin" && userRole != "organizer") {
		writeError(w, http.StatusForbidden, "Access denied")
		return
	}

	organizerID, _ := r.Context().Value("organizer_id").(uint)

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
