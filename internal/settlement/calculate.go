package settlement

import (
	"errors"
	"fmt"
	"time"

	"ticketing_system/internal/models"
)

// SettlementCalculation holds the calculated settlement amounts
type SettlementCalculation struct {
	OrganizerID       uint
	EventID           uint
	GrossAmount       models.Money // Total ticket sales
	PlatformFeeAmount models.Money // Platform commission
	GatewayFeeAmount  models.Money // Payment gateway fees
	RefundDeduction   models.Money // Refunds to be deducted
	ChargebackAmount  models.Money // Chargebacks to be deducted
	AdjustmentAmount  models.Money // Manual adjustments
	NetAmount         models.Money // Final payout amount
	Currency          string
	PaymentRecordIDs  []uint // IDs of payment records included
}

// CalculateEventSettlement calculates settlement for a specific event
// This is called AFTER event completion + holding period
func (s *Service) CalculateEventSettlement(eventID uint) (*SettlementCalculation, error) {
	// 1. Verify event is completed and past holding period
	var event models.Event
	if err := s.db.First(&event, eventID).Error; err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}

	if event.Status != models.EventCompleted {
		return nil, errors.New("event must be completed before settlement")
	}

	// 2. Get all completed payment records for this event
	var paymentRecords []models.PaymentRecord
	if err := s.db.Where("event_id = ? AND type = ? AND status = ?",
		eventID,
		models.RecordCustomerPayment,
		models.RecordCompleted,
	).Find(&paymentRecords).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch payment records: %w", err)
	}

	if len(paymentRecords) == 0 {
		return nil, errors.New("no completed payments found for this event")
	}

	// 3. Calculate gross amount (total ticket sales)
	var grossAmount models.Money
	var paymentRecordIDs []uint
	for _, record := range paymentRecords {
		grossAmount += record.Amount
		paymentRecordIDs = append(paymentRecordIDs, record.ID)
	}

	// 4. Calculate platform fees
	var platformFeeAmount models.Money
	if err := s.db.Model(&models.PaymentRecord{}).
		Where("event_id = ? AND type = ? AND status = ?",
			eventID,
			models.RecordPlatformFee,
			models.RecordCompleted,
		).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&platformFeeAmount).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate platform fees: %w", err)
	}

	// 5. Calculate gateway fees
	var gatewayFeeAmount models.Money
	if err := s.db.Model(&models.PaymentRecord{}).
		Where("event_id = ? AND type = ? AND status = ?",
			eventID,
			models.RecordGatewayFee,
			models.RecordCompleted,
		).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&gatewayFeeAmount).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate gateway fees: %w", err)
	}

	// 6. Calculate refunds for this event
	var refundAmount models.Money
	if err := s.db.Model(&models.RefundRecord{}).
		Where("event_id = ? AND status = ?",
			eventID,
			models.RefundCompleted,
		).
		Select("COALESCE(SUM(refund_amount), 0)").
		Scan(&refundAmount).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate refunds: %w", err)
	}

	// 7. Calculate chargebacks
	var chargebackAmount models.Money
	if err := s.db.Model(&models.PaymentRecord{}).
		Where("event_id = ? AND type = ? AND status = ?",
			eventID,
			models.RecordChargeback,
			models.RecordCompleted,
		).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&chargebackAmount).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate chargebacks: %w", err)
	}

	// 8. Check for any manual adjustments
	var adjustmentAmount models.Money
	if err := s.db.Model(&models.PaymentRecord{}).
		Where("event_id = ? AND type = ?",
			eventID,
			models.RecordAdjustment,
		).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&adjustmentAmount).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate adjustments: %w", err)
	}

	// 9. Calculate net amount
	netAmount := grossAmount - platformFeeAmount - gatewayFeeAmount - refundAmount - chargebackAmount + adjustmentAmount

	// Ensure net amount is not negative
	if netAmount < 0 {
		netAmount = 0
	}

	return &SettlementCalculation{
		OrganizerID:       event.OrganizerID,
		EventID:           eventID,
		GrossAmount:       grossAmount,
		PlatformFeeAmount: platformFeeAmount,
		GatewayFeeAmount:  gatewayFeeAmount,
		RefundDeduction:   refundAmount,
		ChargebackAmount:  chargebackAmount,
		AdjustmentAmount:  adjustmentAmount,
		NetAmount:         netAmount,
		Currency:          event.Currency,
		PaymentRecordIDs:  paymentRecordIDs,
	}, nil
}

// CalculateOrganizerSettlement calculates settlement for all completed events by an organizer
// within a specific time period
func (s *Service) CalculateOrganizerSettlement(organizerID uint, startDate, endDate time.Time) ([]*SettlementCalculation, error) {
	// Get all completed events in the period
	var events []models.Event
	if err := s.db.Where(
		"organizer_id = ? AND status = ? AND end_date >= ? AND end_date <= ?",
		organizerID,
		models.EventCompleted,
		startDate,
		endDate,
	).Find(&events).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch events: %w", err)
	}

	if len(events) == 0 {
		return nil, errors.New("no completed events found for organizer in this period")
	}

	// Calculate settlement for each event
	calculations := make([]*SettlementCalculation, 0, len(events))
	for _, event := range events {
		calc, err := s.CalculateEventSettlement(event.ID)
		if err != nil {
			// Log error but continue with other events
			continue
		}
		calculations = append(calculations, calc)
	}

	if len(calculations) == 0 {
		return nil, errors.New("no valid settlements calculated")
	}

	return calculations, nil
}

// CalculateBatchSettlement calculates settlement for multiple organizers
// This is used for scheduled settlement runs (weekly, monthly, etc.)
func (s *Service) CalculateBatchSettlement(startDate, endDate time.Time, holdingPeriodDays int) (map[uint][]*SettlementCalculation, error) {
	// Calculate earliest event end date (must account for holding period)
	holdingPeriodEnd := time.Now().AddDate(0, 0, -holdingPeriodDays)

	// Get all completed events that are past the holding period
	var events []models.Event
	if err := s.db.Where(
		"status = ? AND end_date >= ? AND end_date <= ? AND end_date <= ?",
		models.EventCompleted,
		startDate,
		endDate,
		holdingPeriodEnd,
	).Find(&events).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch events: %w", err)
	}

	if len(events) == 0 {
		return nil, errors.New("no events eligible for settlement in this period")
	}

	// Group by organizer
	organizerEvents := make(map[uint][]uint)
	for _, event := range events {
		organizerEvents[event.OrganizerID] = append(organizerEvents[event.OrganizerID], event.ID)
	}

	// Calculate settlements per organizer
	result := make(map[uint][]*SettlementCalculation)
	for organizerID, eventIDs := range organizerEvents {
		calculations := make([]*SettlementCalculation, 0, len(eventIDs))
		for _, eventID := range eventIDs {
			calc, err := s.CalculateEventSettlement(eventID)
			if err != nil {
				// Log error but continue
				continue
			}
			calculations = append(calculations, calc)
		}
		if len(calculations) > 0 {
			result[organizerID] = calculations
		}
	}

	if len(result) == 0 {
		return nil, errors.New("no valid settlements calculated for batch")
	}

	return result, nil
}

// GetSettlementPreview provides a preview of what will be settled
// without creating actual settlement records
func (s *Service) GetSettlementPreview(organizerID uint, eventID *uint) (map[string]interface{}, error) {
	var calculations []*SettlementCalculation
	var err error

	if eventID != nil {
		// Single event preview
		calc, err := s.CalculateEventSettlement(*eventID)
		if err != nil {
			return nil, err
		}
		calculations = []*SettlementCalculation{calc}
	} else {
		// All eligible events for organizer
		endDate := time.Now().AddDate(0, 0, -7) // 7 day holding period
		startDate := endDate.AddDate(0, -1, 0)  // Last month
		calculations, err = s.CalculateOrganizerSettlement(organizerID, startDate, endDate)
		if err != nil {
			return nil, err
		}
	}

	// Aggregate totals
	var totalGross, totalPlatformFees, totalGatewayFees, totalRefunds, totalChargebacks, totalNet models.Money
	eventCount := len(calculations)

	for _, calc := range calculations {
		totalGross += calc.GrossAmount
		totalPlatformFees += calc.PlatformFeeAmount
		totalGatewayFees += calc.GatewayFeeAmount
		totalRefunds += calc.RefundDeduction
		totalChargebacks += calc.ChargebackAmount
		totalNet += calc.NetAmount
	}

	return map[string]interface{}{
		"organizer_id":        organizerID,
		"event_count":         eventCount,
		"total_gross":         totalGross,
		"total_platform_fees": totalPlatformFees,
		"total_gateway_fees":  totalGatewayFees,
		"total_refunds":       totalRefunds,
		"total_chargebacks":   totalChargebacks,
		"total_net_payout":    totalNet,
		"calculations":        calculations,
	}, nil
}

// ValidateSettlementEligibility checks if an event is eligible for settlement
func (s *Service) ValidateSettlementEligibility(eventID uint, holdingPeriodDays int) error {
	var event models.Event
	if err := s.db.First(&event, eventID).Error; err != nil {
		return fmt.Errorf("event not found: %w", err)
	}

	// Check event status
	if event.Status != models.EventCompleted {
		return errors.New("event must be completed")
	}

	// Check holding period
	holdingPeriodEnd := time.Now().AddDate(0, 0, -holdingPeriodDays)
	if event.EndDate.After(holdingPeriodEnd) {
		return fmt.Errorf("event is still in holding period (ends %s)", event.EndDate.AddDate(0, 0, holdingPeriodDays))
	}

	// Check for active disputes
	var disputeCount int64
	if err := s.db.Model(&models.PaymentRecord{}).
		Where("event_id = ? AND status = ?", eventID, models.RecordDisputed).
		Count(&disputeCount).Error; err != nil {
		return fmt.Errorf("failed to check for disputes: %w", err)
	}

	if disputeCount > 0 {
		return fmt.Errorf("event has %d active disputes", disputeCount)
	}

	// Check if already settled
	var existingSettlement models.SettlementItem
	if err := s.db.Where("event_id = ? AND status IN ?", eventID,
		[]models.SettlementStatus{
			models.SettlementCompleted,
			models.SettlementProcessing,
			models.SettlementReadyToProcess,
		}).First(&existingSettlement).Error; err == nil {
		return errors.New("event has already been settled or is being processed")
	}

	return nil
}
