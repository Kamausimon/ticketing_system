package settlement

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"ticketing_system/internal/models"

	"github.com/gorilla/mux"
)

// SettlementHandler handles HTTP requests for settlements
type SettlementHandler struct {
	service *Service
}

// NewSettlementHandler creates a new settlement handler
func NewSettlementHandler(service *Service) *SettlementHandler {
	return &SettlementHandler{service: service}
}

// CalculateEventSettlement calculates settlement for a specific event
func (h *SettlementHandler) CalculateEventSettlement(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	calculation, err := h.service.CalculateEventSettlement(uint(eventID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(calculation)
}

// GetSettlementPreview provides preview of settlement amounts
func (h *SettlementHandler) GetSettlementPreview(w http.ResponseWriter, r *http.Request) {
	organizerIDStr := r.URL.Query().Get("organizer_id")
	eventIDStr := r.URL.Query().Get("event_id")

	organizerID, err := strconv.ParseUint(organizerIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid organizer ID", http.StatusBadRequest)
		return
	}

	var eventID *uint
	if eventIDStr != "" {
		eid, err := strconv.ParseUint(eventIDStr, 10, 32)
		if err != nil {
			http.Error(w, "Invalid event ID", http.StatusBadRequest)
			return
		}
		eidUint := uint(eid)
		eventID = &eidUint
	}

	preview, err := h.service.GetSettlementPreview(uint(organizerID), eventID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(preview)
}

// ValidateSettlementEligibility checks if event is eligible for settlement
func (h *SettlementHandler) ValidateSettlementEligibility(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	holdingPeriodDays := 7 // Default
	if hpd := r.URL.Query().Get("holding_period_days"); hpd != "" {
		if days, err := strconv.Atoi(hpd); err == nil {
			holdingPeriodDays = days
		}
	}

	err = h.service.ValidateSettlementEligibility(uint(eventID), holdingPeriodDays)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"eligible": false,
			"reason":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"eligible": true,
	})
}

// CreateSettlementBatch creates a new settlement batch
func (h *SettlementHandler) CreateSettlementBatch(w http.ResponseWriter, r *http.Request) {
	var req CreateSettlementBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	settlement, err := h.service.CreateSettlementBatch(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(settlement)
}

// GetSettlement retrieves a specific settlement
func (h *SettlementHandler) GetSettlement(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	settlementID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid settlement ID", http.StatusBadRequest)
		return
	}

	settlement, err := h.service.GetSettlement(uint(settlementID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settlement)
}

// ListSettlements lists settlements with filters
func (h *SettlementHandler) ListSettlements(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	var status *models.SettlementStatus
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		s := models.SettlementStatus(statusStr)
		status = &s
	}

	var organizerID *uint
	if orgIDStr := r.URL.Query().Get("organizer_id"); orgIDStr != "" {
		if oid, err := strconv.ParseUint(orgIDStr, 10, 32); err == nil {
			oidUint := uint(oid)
			organizerID = &oidUint
		}
	}

	var startDate, endDate *time.Time
	if sd := r.URL.Query().Get("start_date"); sd != "" {
		if t, err := time.Parse(time.RFC3339, sd); err == nil {
			startDate = &t
		}
	}
	if ed := r.URL.Query().Get("end_date"); ed != "" {
		if t, err := time.Parse(time.RFC3339, ed); err == nil {
			endDate = &t
		}
	}

	settlements, total, err := h.service.ListSettlements(status, organizerID, startDate, endDate, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	response := map[string]interface{}{
		"settlements": settlements,
		"total_count": total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ApproveSettlement approves a pending settlement
func (h *SettlementHandler) ApproveSettlement(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	settlementID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid settlement ID", http.StatusBadRequest)
		return
	}

	var req struct {
		ApprovedByUserID uint    `json:"approved_by_user_id"`
		Notes            *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.ApproveSettlement(uint(settlementID), req.ApprovedByUserID, req.Notes); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Settlement approved successfully"})
}

// ProcessSettlement processes an approved settlement
func (h *SettlementHandler) ProcessSettlement(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	settlementID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid settlement ID", http.StatusBadRequest)
		return
	}

	var req struct {
		PaymentGatewayID  uint    `json:"payment_gateway_id"`
		ProcessedByUserID uint    `json:"processed_by_user_id"`
		Notes             *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	processReq := ProcessSettlementRequest{
		SettlementRecordID: uint(settlementID),
		PaymentGatewayID:   req.PaymentGatewayID,
		ProcessedByUserID:  req.ProcessedByUserID,
		Notes:              req.Notes,
	}

	if err := h.service.ProcessSettlement(processReq); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Settlement processing initiated"})
}

// CancelSettlement cancels a pending settlement
func (h *SettlementHandler) CancelSettlement(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	settlementID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid settlement ID", http.StatusBadRequest)
		return
	}

	var req struct {
		CancelledByUserID uint   `json:"cancelled_by_user_id"`
		Reason            string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.CancelSettlement(uint(settlementID), req.CancelledByUserID, req.Reason); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Settlement cancelled successfully"})
}

// WithholdSettlement withholds a settlement
func (h *SettlementHandler) WithholdSettlement(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	settlementID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid settlement ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.WithholdSettlement(uint(settlementID), req.Reason); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Settlement withheld successfully"})
}

// GenerateSettlementReport generates a detailed settlement report
func (h *SettlementHandler) GenerateSettlementReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	settlementID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid settlement ID", http.StatusBadRequest)
		return
	}

	report, err := h.service.GenerateSettlementReport(uint(settlementID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// GetOrganizerSettlementSummary gets settlement summary for an organizer
func (h *SettlementHandler) GetOrganizerSettlementSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizerID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid organizer ID", http.StatusBadRequest)
		return
	}

	// Parse date range
	startDate := time.Now().AddDate(0, -3, 0) // Default: last 3 months
	endDate := time.Now()

	if sd := r.URL.Query().Get("start_date"); sd != "" {
		if t, err := time.Parse(time.RFC3339, sd); err == nil {
			startDate = t
		}
	}
	if ed := r.URL.Query().Get("end_date"); ed != "" {
		if t, err := time.Parse(time.RFC3339, ed); err == nil {
			endDate = t
		}
	}

	summary, err := h.service.GetOrganizerSettlementSummary(uint(organizerID), startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// GetPlatformSettlementSummary gets platform-wide settlement metrics
func (h *SettlementHandler) GetPlatformSettlementSummary(w http.ResponseWriter, r *http.Request) {
	// Parse date range
	startDate := time.Now().AddDate(0, -1, 0) // Default: last month
	endDate := time.Now()

	if sd := r.URL.Query().Get("start_date"); sd != "" {
		if t, err := time.Parse(time.RFC3339, sd); err == nil {
			startDate = t
		}
	}
	if ed := r.URL.Query().Get("end_date"); ed != "" {
		if t, err := time.Parse(time.RFC3339, ed); err == nil {
			endDate = t
		}
	}

	summary, err := h.service.GetPlatformSettlementSummary(startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// ExportSettlements exports settlement data
func (h *SettlementHandler) ExportSettlements(w http.ResponseWriter, r *http.Request) {
	// Parse date range
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()

	if sd := r.URL.Query().Get("start_date"); sd != "" {
		if t, err := time.Parse(time.RFC3339, sd); err == nil {
			startDate = t
		}
	}
	if ed := r.URL.Query().Get("end_date"); ed != "" {
		if t, err := time.Parse(time.RFC3339, ed); err == nil {
			endDate = t
		}
	}

	exportData, err := h.service.ExportSettlementsForPeriod(startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exportData)
}

// GetSettlementHistory gets settlement history for an organizer
func (h *SettlementHandler) GetSettlementHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizerID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid organizer ID", http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	items, total, err := h.service.GetSettlementHistory(uint(organizerID), limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	response := map[string]interface{}{
		"settlements": items,
		"total_count": total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPendingSettlements lists all pending settlements
func (h *SettlementHandler) GetPendingSettlements(w http.ResponseWriter, r *http.Request) {
	settlements, err := h.service.GetPendingSettlements()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settlements)
}

// GetFailedSettlements lists all failed settlements
func (h *SettlementHandler) GetFailedSettlements(w http.ResponseWriter, r *http.Request) {
	settlements, err := h.service.GetFailedSettlements()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settlements)
}

// RetryFailedSettlement retries a failed settlement
func (h *SettlementHandler) RetryFailedSettlement(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	settlementID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid settlement ID", http.StatusBadRequest)
		return
	}

	var req struct {
		PaymentGatewayID uint `json:"payment_gateway_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.RetryFailedSettlement(uint(settlementID), req.PaymentGatewayID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Settlement retry initiated"})
}

// CompleteSettlementItem marks a settlement item as completed (webhook handler)
func (h *SettlementHandler) CompleteSettlementItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	var req struct {
		ExternalTransactionID string `json:"external_transaction_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.CompleteSettlementItem(uint(itemID), req.ExternalTransactionID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Settlement item completed"})
}

// FailSettlementItem marks a settlement item as failed
func (h *SettlementHandler) FailSettlementItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	var req struct {
		FailureReason string `json:"failure_reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.FailSettlementItem(uint(itemID), req.FailureReason); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Settlement item marked as failed"})
}

// HandleSettlementWebhook handles settlement completion webhooks from payment gateway
func (h *SettlementHandler) HandleSettlementWebhook(w http.ResponseWriter, r *http.Request) {
	var webhook struct {
		SettlementItemID      uint   `json:"settlement_item_id"`
		ExternalTransactionID string `json:"external_transaction_id"`
		Status                string `json:"status"` // "completed" or "failed"
		FailureReason         string `json:"failure_reason,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		http.Error(w, "Invalid webhook payload", http.StatusBadRequest)
		return
	}

	if webhook.Status == "completed" {
		if err := h.service.CompleteSettlementItem(webhook.SettlementItemID, webhook.ExternalTransactionID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if webhook.Status == "failed" {
		if err := h.service.FailSettlementItem(webhook.SettlementItemID, webhook.FailureReason); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Webhook processed"})
}
