package orders

import (
	"fmt"
	"ticketing_system/internal/models"
	"time"

	"gorm.io/gorm"
)

// calculateOrderTotal calculates the total cost of an order
func (h *OrderHandler) calculateOrderTotal(db *gorm.DB, req CreateOrderRequest) (*OrderCalculation, error) {
	calculation := &OrderCalculation{
		Currency: "KSH", // Default currency
	}

	var subtotal float64
	now := time.Now()

	// Calculate subtotal from items
	for _, item := range req.Items {
		var ticketClass models.TicketClass
		if err := db.Where("id = ?", item.TicketClassID).First(&ticketClass).Error; err != nil {
			return nil, fmt.Errorf("ticket class %d not found", item.TicketClassID)
		}

		// Check if ticket class is paused or hidden
		if ticketClass.IsPaused {
			return nil, fmt.Errorf("ticket class '%s' is currently unavailable", ticketClass.Name)
		}

		// Check sale dates
		if ticketClass.StartSaleDate != nil && now.Before(*ticketClass.StartSaleDate) {
			return nil, fmt.Errorf("ticket class '%s' is not yet on sale", ticketClass.Name)
		}
		if ticketClass.EndSaleDate != nil && now.After(*ticketClass.EndSaleDate) {
			return nil, fmt.Errorf("ticket class '%s' sale has ended", ticketClass.Name)
		}

		// Check ticket availability
		if ticketClass.QuantityAvailable != nil {
			available := *ticketClass.QuantityAvailable - ticketClass.QuantitySold
			if available < item.Quantity {
				return nil, fmt.Errorf("only %d tickets available for '%s'", available, ticketClass.Name)
			}
		}

		// Check min/max per order
		if ticketClass.MinPerOrder != nil && item.Quantity < *ticketClass.MinPerOrder {
			return nil, fmt.Errorf("minimum %d tickets required for '%s'", *ticketClass.MinPerOrder, ticketClass.Name)
		}
		if ticketClass.MaxPerOrder != nil && item.Quantity > *ticketClass.MaxPerOrder {
			return nil, fmt.Errorf("maximum %d tickets allowed for '%s'", *ticketClass.MaxPerOrder, ticketClass.Name)
		}

		// Set currency from ticket class
		if calculation.Currency == "KSH" {
			calculation.Currency = ticketClass.Currency
		}

		itemTotal := float64(ticketClass.Price) * float64(item.Quantity)
		subtotal += itemTotal
	}

	calculation.Subtotal = subtotal

	// Apply promo code if provided
	if req.PromoCode != "" {
		discount, err := h.calculatePromoDiscount(db, req.PromoCode, req.EventID, subtotal)
		if err != nil {
			// Don't fail the order if promo code is invalid, just ignore it
			// In production, you might want to return an error
			calculation.Discount = 0
		} else {
			calculation.Discount = discount
			calculation.PromoCodeApplied = req.PromoCode
		}
	}

	// Calculate booking fees (typically 2-5% of subtotal)
	calculation.BookingFee = calculateBookingFee(subtotal)

	// Calculate organizer booking fee (organizer pays a portion)
	calculation.OrganizerBookingFee = calculateOrganizerFee(subtotal)

	// Calculate tax (e.g., VAT at 16% in Kenya)
	calculation.TaxAmount = calculateTax(subtotal - calculation.Discount)

	// Calculate total
	calculation.TotalAmount = subtotal + calculation.BookingFee + calculation.TaxAmount - calculation.Discount

	return calculation, nil
}

// calculatePromoDiscount calculates discount from promo code
func (h *OrderHandler) calculatePromoDiscount(db *gorm.DB, promoCode string, eventID uint, subtotal float64) (float64, error) {
	var promo models.Promotion

	// Find active promotion
	if err := db.Where("code = ? AND status = ?", promoCode, models.PromotionActive).First(&promo).Error; err != nil {
		return 0, fmt.Errorf("invalid or expired promo code")
	}

	// Check if promotion applies to this event
	if promo.EventID != nil && *promo.EventID != eventID {
		return 0, fmt.Errorf("promo code not valid for this event")
	}

	// Check if promotion is within date range
	// now := time.Now()
	// if now.Before(promo.StartDate) || now.After(promo.EndDate) {
	// 	return 0, fmt.Errorf("promo code not valid at this time")
	// }

	// Check usage limits
	if promo.UsageLimit != nil && promo.UsageCount >= *promo.UsageLimit {
		return 0, fmt.Errorf("promo code usage limit reached")
	}

	// Calculate discount based on type
	var discount float64

	switch promo.Type {
	case models.PromotionPercentage:
		if promo.DiscountPercentage != nil {
			discount = subtotal * (float64(*promo.DiscountPercentage) / 100)
		}
	case models.PromotionFixedAmount:
		if promo.DiscountAmount != nil {
			discount = float64(*promo.DiscountAmount)
		}
	}

	// Apply minimum purchase requirement
	if promo.MinimumPurchase != nil && subtotal < float64(*promo.MinimumPurchase) {
		return 0, fmt.Errorf("minimum purchase of %s required", promo.MinimumPurchase)
	}

	// Apply maximum discount cap
	if promo.MaximumDiscount != nil && discount > float64(*promo.MaximumDiscount) {
		discount = float64(*promo.MaximumDiscount)
	}

	// Ensure discount doesn't exceed subtotal
	if discount > subtotal {
		discount = subtotal
	}

	return discount, nil
}

// calculateBookingFee calculates the booking fee for customers
func calculateBookingFee(subtotal float64) float64 {
	// Typically 2-5% of subtotal
	const bookingFeePercentage = 0.03 // 3%
	return subtotal * bookingFeePercentage
}

// calculateOrganizerFee calculates the fee charged to organizers
func calculateOrganizerFee(subtotal float64) float64 {
	// Typically 5-10% of subtotal paid by organizer
	const organizerFeePercentage = 0.05 // 5%
	return subtotal * organizerFeePercentage
}

// calculateTax calculates tax on the order
func calculateTax(amount float64) float64 {
	// Kenya VAT rate is 16%
	const taxRate = 0.16
	return amount * taxRate
}
