package settlement

import (
	"fmt"
	"time"

	"ticketing_system/internal/models"
)

// SettlementReport represents a settlement summary report
type SettlementReport struct {
	SettlementID    uint                    `json:"settlement_id"`
	BatchID         string                  `json:"batch_id"`
	Status          models.SettlementStatus `json:"status"`
	Description     string                  `json:"description"`
	PeriodStart     time.Time               `json:"period_start"`
	PeriodEnd       time.Time               `json:"period_end"`
	TotalOrganizers int                     `json:"total_organizers"`
	TotalAmount     models.Money            `json:"total_amount"`
	Currency        string                  `json:"currency"`
	ProcessedAt     *time.Time              `json:"processed_at"`
	CompletedAt     *time.Time              `json:"completed_at"`
	Items           []SettlementItemReport  `json:"items"`
}

// SettlementItemReport represents individual settlement item in report
type SettlementItemReport struct {
	ItemID            uint                    `json:"item_id"`
	OrganizerID       uint                    `json:"organizer_id"`
	OrganizerName     string                  `json:"organizer_name"`
	EventID           uint                    `json:"event_id"`
	EventTitle        string                  `json:"event_title"`
	EventEndDate      time.Time               `json:"event_end_date"`
	GrossAmount       models.Money            `json:"gross_amount"`
	PlatformFeeAmount models.Money            `json:"platform_fee_amount"`
	RefundDeduction   models.Money            `json:"refund_deduction"`
	NetAmount         models.Money            `json:"net_amount"`
	Status            models.SettlementStatus `json:"status"`
	CompletedAt       *time.Time              `json:"completed_at"`
}

// OrganizerSettlementSummary provides settlement summary for an organizer
type OrganizerSettlementSummary struct {
	OrganizerID        uint         `json:"organizer_id"`
	OrganizerName      string       `json:"organizer_name"`
	TotalSettlements   int          `json:"total_settlements"`
	TotalGrossAmount   models.Money `json:"total_gross_amount"`
	TotalPlatformFees  models.Money `json:"total_platform_fees"`
	TotalRefunds       models.Money `json:"total_refunds"`
	TotalNetAmount     models.Money `json:"total_net_amount"`
	PendingAmount      models.Money `json:"pending_amount"`
	CompletedAmount    models.Money `json:"completed_amount"`
	Currency           string       `json:"currency"`
	LastSettlementDate *time.Time   `json:"last_settlement_date"`
	NextSettlementDate *time.Time   `json:"next_settlement_date"`
}

// PlatformSettlementSummary provides overall platform settlement metrics
type PlatformSettlementSummary struct {
	TotalSettlementBatches  int          `json:"total_settlement_batches"`
	TotalOrganizers         int          `json:"total_organizers"`
	TotalGrossAmount        models.Money `json:"total_gross_amount"`
	TotalPlatformFees       models.Money `json:"total_platform_fees"`
	TotalNetPayouts         models.Money `json:"total_net_payouts"`
	PendingSettlements      int          `json:"pending_settlements"`
	PendingAmount           models.Money `json:"pending_amount"`
	CompletedSettlements    int          `json:"completed_settlements"`
	CompletedAmount         models.Money `json:"completed_amount"`
	FailedSettlements       int          `json:"failed_settlements"`
	AverageSettlementAmount models.Money `json:"average_settlement_amount"`
	Currency                string       `json:"currency"`
}

// GenerateSettlementReport generates a detailed report for a settlement batch
func (s *Service) GenerateSettlementReport(settlementID uint) (*SettlementReport, error) {
	var settlement models.SettlementRecord
	if err := s.db.
		Preload("SettlementItems").
		Preload("SettlementItems.Organizer").
		Preload("SettlementItems.Event").
		First(&settlement, settlementID).Error; err != nil {
		return nil, fmt.Errorf("settlement not found: %w", err)
	}

	items := make([]SettlementItemReport, len(settlement.SettlementItems))
	for i, item := range settlement.SettlementItems {
		organizerName := item.Organizer.Name

		items[i] = SettlementItemReport{
			ItemID:            item.ID,
			OrganizerID:       item.OrganizerID,
			OrganizerName:     organizerName,
			EventID:           item.EventID,
			EventTitle:        item.Event.Title,
			EventEndDate:      item.EventEndDate,
			GrossAmount:       item.GrossAmount,
			PlatformFeeAmount: item.PlatformFeeAmount,
			RefundDeduction:   item.RefundDeduction,
			NetAmount:         item.NetAmount,
			Status:            item.Status,
			CompletedAt:       item.CompletedAt,
		}
	}

	return &SettlementReport{
		SettlementID:    settlement.ID,
		BatchID:         settlement.SettlementBatchID,
		Status:          settlement.Status,
		Description:     settlement.Description,
		PeriodStart:     settlement.PeriodStartDate,
		PeriodEnd:       settlement.PeriodEndDate,
		TotalOrganizers: settlement.TotalOrganizers,
		TotalAmount:     settlement.TotalAmount,
		Currency:        settlement.Currency,
		ProcessedAt:     settlement.ProcessedAt,
		CompletedAt:     settlement.CompletedAt,
		Items:           items,
	}, nil
}

// GetOrganizerSettlementSummary generates settlement summary for an organizer
func (s *Service) GetOrganizerSettlementSummary(organizerID uint, startDate, endDate time.Time) (*OrganizerSettlementSummary, error) {
	var organizer models.Organizer
	if err := s.db.First(&organizer, organizerID).Error; err != nil {
		return nil, fmt.Errorf("organizer not found: %w", err)
	}

	// Get all settlement items for this organizer in the period
	var items []models.SettlementItem
	query := s.db.Where("organizer_id = ?", organizerID)

	if !startDate.IsZero() {
		query = query.Where("created_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("created_at <= ?", endDate)
	}

	if err := query.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch settlement items: %w", err)
	}

	// Calculate totals
	var totalGross, totalPlatformFees, totalRefunds, totalNet models.Money
	var pendingAmount, completedAmount models.Money
	var lastSettlementDate *time.Time

	for _, item := range items {
		totalGross += item.GrossAmount
		totalPlatformFees += item.PlatformFeeAmount
		totalRefunds += item.RefundDeduction
		totalNet += item.NetAmount

		if item.Status == models.SettlementCompleted {
			completedAmount += item.NetAmount
			if lastSettlementDate == nil || (item.CompletedAt != nil && item.CompletedAt.After(*lastSettlementDate)) {
				lastSettlementDate = item.CompletedAt
			}
		} else if item.Status == models.SettlementPending || item.Status == models.SettlementReadyToProcess {
			pendingAmount += item.NetAmount
		}
	}

	// Calculate next settlement date (estimate based on frequency)
	var nextSettlementDate *time.Time
	if lastSettlementDate != nil {
		// Assume monthly settlements by default
		estimatedNext := lastSettlementDate.AddDate(0, 1, 0)
		nextSettlementDate = &estimatedNext
	}

	return &OrganizerSettlementSummary{
		OrganizerID:        organizerID,
		OrganizerName:      organizer.Name,
		TotalSettlements:   len(items),
		TotalGrossAmount:   totalGross,
		TotalPlatformFees:  totalPlatformFees,
		TotalRefunds:       totalRefunds,
		TotalNetAmount:     totalNet,
		PendingAmount:      pendingAmount,
		CompletedAmount:    completedAmount,
		Currency:           "KSH", // Default
		LastSettlementDate: lastSettlementDate,
		NextSettlementDate: nextSettlementDate,
	}, nil
}

// GetPlatformSettlementSummary generates platform-wide settlement metrics
func (s *Service) GetPlatformSettlementSummary(startDate, endDate time.Time) (*PlatformSettlementSummary, error) {
	query := s.db.Model(&models.SettlementRecord{})

	if !startDate.IsZero() {
		query = query.Where("created_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("created_at <= ?", endDate)
	}

	// Count total settlement batches
	var totalBatches int64
	if err := query.Count(&totalBatches).Error; err != nil {
		return nil, fmt.Errorf("failed to count settlements: %w", err)
	}

	// Get all settlements in period
	var settlements []models.SettlementRecord
	if err := query.Preload("SettlementItems").Find(&settlements).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch settlements: %w", err)
	}

	// Calculate metrics
	organizerSet := make(map[uint]bool)
	var totalGross, totalPlatformFees, totalNetPayouts models.Money
	var pendingAmount, completedAmount models.Money
	var pendingCount, completedCount, failedCount int

	for _, settlement := range settlements {
		for _, item := range settlement.SettlementItems {
			organizerSet[item.OrganizerID] = true
			totalGross += item.GrossAmount
			totalPlatformFees += item.PlatformFeeAmount
			totalNetPayouts += item.NetAmount
		}

		switch settlement.Status {
		case models.SettlementPending, models.SettlementReadyToProcess:
			pendingCount++
			pendingAmount += settlement.TotalAmount
		case models.SettlementCompleted:
			completedCount++
			completedAmount += settlement.TotalAmount
		case models.SettlementFailed:
			failedCount++
		}
	}

	totalOrganizers := len(organizerSet)

	// Calculate average
	averageAmount := models.Money(0)
	if totalBatches > 0 {
		averageAmount = totalNetPayouts / models.Money(totalBatches)
	}

	return &PlatformSettlementSummary{
		TotalSettlementBatches:  int(totalBatches),
		TotalOrganizers:         totalOrganizers,
		TotalGrossAmount:        totalGross,
		TotalPlatformFees:       totalPlatformFees,
		TotalNetPayouts:         totalNetPayouts,
		PendingSettlements:      pendingCount,
		PendingAmount:           pendingAmount,
		CompletedSettlements:    completedCount,
		CompletedAmount:         completedAmount,
		FailedSettlements:       failedCount,
		AverageSettlementAmount: averageAmount,
		Currency:                "KSH",
	}, nil
}

// GetSettlementsByStatus retrieves settlements grouped by status
func (s *Service) GetSettlementsByStatus(startDate, endDate time.Time) (map[models.SettlementStatus][]models.SettlementRecord, error) {
	query := s.db.Model(&models.SettlementRecord{})

	if !startDate.IsZero() {
		query = query.Where("created_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("created_at <= ?", endDate)
	}

	var settlements []models.SettlementRecord
	if err := query.Preload("SettlementItems").Order("created_at DESC").Find(&settlements).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch settlements: %w", err)
	}

	// Group by status
	result := make(map[models.SettlementStatus][]models.SettlementRecord)
	for _, settlement := range settlements {
		result[settlement.Status] = append(result[settlement.Status], settlement)
	}

	return result, nil
}

// GetPendingSettlements retrieves all settlements pending processing
func (s *Service) GetPendingSettlements() ([]models.SettlementRecord, error) {
	var settlements []models.SettlementRecord
	if err := s.db.
		Where("status IN ?", []models.SettlementStatus{
			models.SettlementPending,
			models.SettlementReadyToProcess,
		}).
		Preload("SettlementItems").
		Preload("SettlementItems.Organizer").
		Order("earliest_payout_date ASC").
		Find(&settlements).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch pending settlements: %w", err)
	}

	return settlements, nil
}

// GetFailedSettlements retrieves all failed settlements for retry
func (s *Service) GetFailedSettlements() ([]models.SettlementRecord, error) {
	var settlements []models.SettlementRecord
	if err := s.db.
		Where("status = ?", models.SettlementFailed).
		Preload("SettlementItems").
		Preload("SettlementItems.Organizer").
		Order("failed_at DESC").
		Find(&settlements).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch failed settlements: %w", err)
	}

	return settlements, nil
}

// ExportSettlementReport exports settlement data for accounting/reconciliation
type SettlementExportData struct {
	SettlementBatchID string       `json:"settlement_batch_id"`
	SettlementDate    time.Time    `json:"settlement_date"`
	OrganizerID       uint         `json:"organizer_id"`
	OrganizerName     string       `json:"organizer_name"`
	EventID           uint         `json:"event_id"`
	EventTitle        string       `json:"event_title"`
	EventDate         time.Time    `json:"event_date"`
	GrossAmount       models.Money `json:"gross_amount"`
	PlatformFee       models.Money `json:"platform_fee"`
	GatewayFee        models.Money `json:"gateway_fee"`
	Refunds           models.Money `json:"refunds"`
	Chargebacks       models.Money `json:"chargebacks"`
	NetAmount         models.Money `json:"net_amount"`
	Currency          string       `json:"currency"`
	Status            string       `json:"status"`
	BankAccountNumber string       `json:"bank_account_number"`
	BankName          string       `json:"bank_name"`
	TransactionID     string       `json:"transaction_id"`
}

// ExportSettlementsForPeriod exports settlement data for a specific period
func (s *Service) ExportSettlementsForPeriod(startDate, endDate time.Time) ([]SettlementExportData, error) {
	var settlements []models.SettlementRecord
	if err := s.db.
		Where("period_start_date >= ? AND period_end_date <= ?", startDate, endDate).
		Preload("SettlementItems").
		Preload("SettlementItems.Organizer").
		Preload("SettlementItems.Event").
		Find(&settlements).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch settlements: %w", err)
	}

	var exportData []SettlementExportData
	for _, settlement := range settlements {
		for _, item := range settlement.SettlementItems {
			transactionID := ""
			if item.ExternalTransactionID != nil {
				transactionID = *item.ExternalTransactionID
			}

			exportData = append(exportData, SettlementExportData{
				SettlementBatchID: settlement.SettlementBatchID,
				SettlementDate:    settlement.CreatedAt,
				OrganizerID:       item.OrganizerID,
				OrganizerName:     item.Organizer.Name,
				EventID:           item.EventID,
				EventTitle:        item.Event.Title,
				EventDate:         item.EventEndDate,
				GrossAmount:       item.GrossAmount,
				PlatformFee:       item.PlatformFeeAmount,
				GatewayFee:        0, // Not stored at item level
				Refunds:           item.RefundDeduction,
				Chargebacks:       item.ChargebackAmount,
				NetAmount:         item.NetAmount,
				Currency:          item.Currency,
				Status:            string(item.Status),
				BankAccountNumber: item.BankAccountNumber,
				BankName:          item.BankName,
				TransactionID:     transactionID,
			})
		}
	}

	return exportData, nil
}

// GetSettlementHistory retrieves settlement history for an organizer
func (s *Service) GetSettlementHistory(organizerID uint, limit, offset int) ([]models.SettlementItem, int64, error) {
	var total int64
	if err := s.db.Model(&models.SettlementItem{}).
		Where("organizer_id = ?", organizerID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.SettlementItem
	if err := s.db.
		Where("organizer_id = ?", organizerID).
		Preload("SettlementRecord").
		Preload("Event").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}
