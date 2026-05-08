package settlement

import (
	"errors"
	"fmt"
	"time"

	"ticketing_system/internal/models"

	"gorm.io/gorm"
)

// CreateSettlementBatchRequest holds parameters for creating a settlement batch
type CreateSettlementBatchRequest struct {
	Description       string
	Frequency         models.SettlementFrequency
	Trigger           models.SettlementTrigger
	PeriodStartDate   time.Time
	PeriodEndDate     time.Time
	HoldingPeriodDays int
	InitiatedByUserID uint
	EventID           *uint // Optional: for single event settlements
}

// ProcessSettlementRequest holds parameters for processing a settlement
type ProcessSettlementRequest struct {
	SettlementRecordID uint
	PaymentGatewayID   uint
	ProcessedByUserID  uint
	Notes              *string
}

// CreateSettlementBatch creates a new settlement batch with calculated items
func (s *Service) CreateSettlementBatch(req CreateSettlementBatchRequest) (*models.SettlementRecord, error) {
	// Validate holding period
	if req.HoldingPeriodDays < 0 {
		req.HoldingPeriodDays = 7 // Default 7 days
	}

	// Generate unique batch ID
	batchID := fmt.Sprintf("SETTLE-%s-%d", time.Now().Format("20060102"), time.Now().Unix())

	// Calculate settlements
	var calculations map[uint][]*SettlementCalculation
	var err error

	if req.EventID != nil {
		// Single event settlement
		calc, err := s.CalculateEventSettlement(*req.EventID)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate settlement: %w", err)
		}
		calculations = map[uint][]*SettlementCalculation{
			calc.OrganizerID: {calc},
		}
	} else {
		// Batch settlement for multiple organizers
		calculations, err = s.CalculateBatchSettlement(
			req.PeriodStartDate,
			req.PeriodEndDate,
			req.HoldingPeriodDays,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate batch settlement: %w", err)
		}
	}

	// Calculate totals
	var totalOrganizers, totalPaymentRecords int
	var totalAmount models.Money
	var currency string

	for _, calcs := range calculations {
		totalOrganizers++
		for _, calc := range calcs {
			totalAmount += calc.NetAmount
			totalPaymentRecords += len(calc.PaymentRecordIDs)
			if currency == "" {
				currency = calc.Currency
			}
		}
	}

	// Create settlement record
	holdingPeriodStart := req.PeriodEndDate
	holdingPeriodEnd := holdingPeriodStart.AddDate(0, 0, req.HoldingPeriodDays)
	earliestPayout := holdingPeriodEnd

	settlementRecord := models.SettlementRecord{
		SettlementBatchID:      batchID,
		Description:            req.Description,
		Status:                 models.SettlementPending,
		Frequency:              req.Frequency,
		Trigger:                req.Trigger,
		EventID:                req.EventID,
		HoldingPeriodDays:      req.HoldingPeriodDays,
		HoldingPeriodStartDate: &holdingPeriodStart,
		HoldingPeriodEndDate:   &holdingPeriodEnd,
		EarliestPayoutDate:     &earliestPayout,
		PeriodStartDate:        req.PeriodStartDate,
		PeriodEndDate:          req.PeriodEndDate,
		TotalOrganizers:        totalOrganizers,
		TotalAmount:            totalAmount,
		TotalPaymentRecords:    totalPaymentRecords,
		Currency:               currency,
		InitiatedBy:            &req.InitiatedByUserID,
	}

	// Start transaction
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Create settlement record
		if err := tx.Create(&settlementRecord).Error; err != nil {
			return fmt.Errorf("failed to create settlement record: %w", err)
		}

		// Create settlement items for each organizer
		for organizerID, calcs := range calculations {
			for _, calc := range calcs {
				// Get organizer payout account
				var payoutAccount models.PayoutAccount
				if err := tx.Where("organizer_id = ? AND is_default = ? AND status = ?",
					organizerID,
					true,
					models.PayoutVerified,
				).First(&payoutAccount).Error; err != nil {
					return fmt.Errorf("no verified payout account found for organizer %d: %w", organizerID, err)
				}

				// Get event details
				var event models.Event
				if err := tx.First(&event, calc.EventID).Error; err != nil {
					return fmt.Errorf("event not found: %w", err)
				}

				// Check for disputes
				var disputeCount int64
				tx.Model(&models.PaymentRecord{}).
					Where("event_id = ? AND status = ?", calc.EventID, models.RecordDisputed).
					Count(&disputeCount)

				hasDisputes := disputeCount > 0

				// Create settlement item
				description := fmt.Sprintf("Settlement for event: %s", event.Title)

				// Safely get bank account details
				bankAccountNumber := ""
				if payoutAccount.AccountNumber != nil {
					bankAccountNumber = *payoutAccount.AccountNumber
				}
				bankName := ""
				if payoutAccount.BankName != nil {
					bankName = *payoutAccount.BankName
				}
				accountHolderName := ""
				if payoutAccount.AccountHolderName != nil {
					accountHolderName = *payoutAccount.AccountHolderName
				}

				settlementItem := models.SettlementItem{
					SettlementRecordID: settlementRecord.ID,
					OrganizerID:        organizerID,
					EventID:            calc.EventID,
					EventStatus:        event.Status,
					EventEndDate:       event.EndDate,
					HasDisputes:        hasDisputes,
					RefundAmountIssued: calc.RefundDeduction,
					ChargebackAmount:   calc.ChargebackAmount,
					GrossAmount:        calc.GrossAmount,
					PlatformFeeAmount:  calc.PlatformFeeAmount,
					RefundDeduction:    calc.RefundDeduction,
					AdjustmentAmount:   calc.AdjustmentAmount,
					NetAmount:          calc.NetAmount,
					Currency:           calc.Currency,
					Status:             models.SettlementPending,
					BankAccountNumber:  bankAccountNumber,
					BankName:           bankName,
					BankCode:           payoutAccount.BankCode,
					AccountHolderName:  accountHolderName,
					Description:        description,
				}

				if err := tx.Create(&settlementItem).Error; err != nil {
					return fmt.Errorf("failed to create settlement item: %w", err)
				}

				// Link payment records to settlement item
				var paymentRecords []models.PaymentRecord
				if err := tx.Where("id IN ?", calc.PaymentRecordIDs).Find(&paymentRecords).Error; err != nil {
					return fmt.Errorf("failed to find payment records: %w", err)
				}

				if err := tx.Model(&settlementItem).Association("PaymentRecords").Append(paymentRecords); err != nil {
					return fmt.Errorf("failed to link payment records: %w", err)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &settlementRecord, nil
}

// ApproveSettlement approves a pending settlement for processing
func (s *Service) ApproveSettlement(settlementID uint, approvedByUserID uint, notes *string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var settlement models.SettlementRecord
		if err := tx.First(&settlement, settlementID).Error; err != nil {
			return fmt.Errorf("settlement not found: %w", err)
		}

		if settlement.Status != models.SettlementPending {
			return errors.New("only pending settlements can be approved")
		}

		// Check for issues
		if settlement.HasActiveDisputes {
			return errors.New("cannot approve settlement with active disputes")
		}

		// Verify holding period has passed
		if settlement.EarliestPayoutDate != nil && time.Now().Before(*settlement.EarliestPayoutDate) {
			return errors.New("holding period has not yet ended")
		}

		now := time.Now()
		settlement.Status = models.SettlementReadyToProcess
		settlement.ApprovedBy = &approvedByUserID
		settlement.ApprovedAt = &now
		if notes != nil {
			settlement.Notes = notes
		}

		return tx.Save(&settlement).Error
	})
}

// ProcessSettlement processes a settlement batch by initiating payouts
func (s *Service) ProcessSettlement(req ProcessSettlementRequest) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var settlement models.SettlementRecord
		if err := tx.Preload("SettlementItems").First(&settlement, req.SettlementRecordID).Error; err != nil {
			return fmt.Errorf("settlement not found: %w", err)
		}

		if settlement.Status != models.SettlementReadyToProcess {
			return errors.New("settlement must be in ready_to_process status")
		}

		// Update settlement status
		processedTime := time.Now()
		settlement.Status = models.SettlementProcessing
		settlement.ProcessedAt = &processedTime
		settlement.PaymentGatewayID = &req.PaymentGatewayID
		if req.Notes != nil {
			settlement.Notes = req.Notes
		}

		if err := tx.Save(&settlement).Error; err != nil {
			return fmt.Errorf("failed to update settlement: %w", err)
		}

		// Process each settlement item
		successCount := 0
		failCount := 0

		for _, item := range settlement.SettlementItems {
			item.Status = models.SettlementProcessing
			item.ProcessedAt = &processedTime

			// Here you would integrate with payment gateway to initiate payout
			// For now, we'll mark as processing
			// In production, call payment gateway API here

			if err := tx.Save(&item).Error; err != nil {
				failCount++
				continue
			}
			successCount++
		}

		// If all items processed successfully, update settlement
		if failCount == 0 {
			// In real implementation, this would be updated by webhook/callback
			// when payment gateway confirms payout
			settlement.Status = models.SettlementCompleted
			completedAt := time.Now()
			settlement.CompletedAt = &completedAt
		} else if successCount == 0 {
			settlement.Status = models.SettlementFailed
			failedAt := time.Now()
			settlement.FailedAt = &failedAt
		} else {
			settlement.Status = models.SettlementPartial
		}

		return tx.Save(&settlement).Error
	})
}

// CompleteSettlementItem marks a settlement item as completed
// This should be called by webhook handler when payment gateway confirms payout
func (s *Service) CompleteSettlementItem(itemID uint, externalTransactionID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var item models.SettlementItem
		if err := tx.First(&item, itemID).Error; err != nil {
			return fmt.Errorf("settlement item not found: %w", err)
		}

		if item.Status != models.SettlementProcessing {
			return errors.New("item must be in processing status")
		}

		now := time.Now()
		item.Status = models.SettlementCompleted
		item.CompletedAt = &now
		item.ExternalTransactionID = &externalTransactionID

		if err := tx.Save(&item).Error; err != nil {
			return err
		}

		// Check if all items in the settlement are completed
		var settlement models.SettlementRecord
		if err := tx.Preload("SettlementItems").First(&settlement, item.SettlementRecordID).Error; err != nil {
			return err
		}

		allCompleted := true
		for _, si := range settlement.SettlementItems {
			if si.Status != models.SettlementCompleted {
				allCompleted = false
				break
			}
		}

		if allCompleted {
			settlement.Status = models.SettlementCompleted
			settlement.CompletedAt = &now
			return tx.Save(&settlement).Error
		}

		return nil
	})
}

// FailSettlementItem marks a settlement item as failed
func (s *Service) FailSettlementItem(itemID uint, failureReason string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var item models.SettlementItem
		if err := tx.First(&item, itemID).Error; err != nil {
			return fmt.Errorf("settlement item not found: %w", err)
		}

		now := time.Now()
		item.Status = models.SettlementFailed
		item.FailedAt = &now
		item.FailureReason = &failureReason

		return tx.Save(&item).Error
	})
}

// CancelSettlement cancels a pending settlement
func (s *Service) CancelSettlement(settlementID uint, cancelledByUserID uint, reason string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var settlement models.SettlementRecord
		if err := tx.First(&settlement, settlementID).Error; err != nil {
			return fmt.Errorf("settlement not found: %w", err)
		}

		if settlement.Status != models.SettlementPending && settlement.Status != models.SettlementReadyToProcess {
			return errors.New("only pending or ready settlements can be cancelled")
		}

		settlement.Status = models.SettlementCancelled
		settlement.Notes = &reason

		if err := tx.Save(&settlement).Error; err != nil {
			return err
		}

		// Cancel all settlement items
		return tx.Model(&models.SettlementItem{}).
			Where("settlement_record_id = ?", settlementID).
			Update("status", models.SettlementCancelled).Error
	})
}

// WithholdSettlement withholds a settlement due to issues
func (s *Service) WithholdSettlement(settlementID uint, withholdingReason string) error {
	return s.db.Model(&models.SettlementRecord{}).
		Where("id = ?", settlementID).
		Updates(map[string]interface{}{
			"status":             models.SettlementWithheld,
			"withholding_reason": withholdingReason,
		}).Error
}

// RetryFailedSettlement retries a failed settlement
func (s *Service) RetryFailedSettlement(settlementID uint, paymentGatewayID uint) error {
	return s.ProcessSettlement(ProcessSettlementRequest{
		SettlementRecordID: settlementID,
		PaymentGatewayID:   paymentGatewayID,
	})
}

// GetSettlement retrieves a settlement record with all items
func (s *Service) GetSettlement(settlementID uint) (*models.SettlementRecord, error) {
	var settlement models.SettlementRecord
	if err := s.db.
		Preload("SettlementItems").
		Preload("SettlementItems.Organizer").
		Preload("SettlementItems.Event").
		Preload("InitiatedByUser").
		Preload("ApprovedByUser").
		First(&settlement, settlementID).Error; err != nil {
		return nil, fmt.Errorf("settlement not found: %w", err)
	}
	return &settlement, nil
}

// ListSettlements retrieves settlements with filters
func (s *Service) ListSettlements(status *models.SettlementStatus, organizerID *uint, startDate, endDate *time.Time, limit, offset int) ([]models.SettlementRecord, int64, error) {
	query := s.db.Model(&models.SettlementRecord{})

	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if startDate != nil {
		query = query.Where("period_start_date >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("period_end_date <= ?", *endDate)
	}
	if organizerID != nil {
		query = query.Joins("JOIN settlement_items ON settlement_items.settlement_record_id = settlement_records.id").
			Where("settlement_items.organizer_id = ?", *organizerID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var settlements []models.SettlementRecord
	if err := query.
		Preload("SettlementItems").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&settlements).Error; err != nil {
		return nil, 0, err
	}

	return settlements, total, nil
}
