package orders

import (
	"fmt"
	"ticketing_system/internal/models"
	"time"
)

// ProcessPaymentWithTickets handles payment verification and ticket generation atomically
// This ensures that either both operations succeed or both fail (transaction atomicity)
func (h *OrderHandler) ProcessPaymentWithTickets(
	orderID uint,
	paymentMethod string,
	paymentResult map[string]interface{},
) error {
	// Start transaction
	tx := h.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	// Ensure rollback on panic or error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get order with items
	var order models.Order
	if err := tx.Preload("OrderItems.TicketClass.Event").
		First(&order, orderID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to find order: %w", err)
	}

	// Verify order is in correct state for payment
	if order.Status != models.OrderPending {
		tx.Rollback()
		return fmt.Errorf("order is not in pending state")
	}

	if order.PaymentStatus != models.PaymentPending {
		tx.Rollback()
		return fmt.Errorf("payment is not in pending state")
	}

	// Step 1: Update order payment status
	order.PaymentStatus = models.PaymentCompleted
	order.IsPaymentReceived = true
	order.Status = models.OrderPaid
	completedTime := time.Now()
	order.CompletedAt = &completedTime

	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update order status: %w", err)
	}

	// Step 2: Generate tickets for each order item
	for _, item := range order.OrderItems {
		// Check if tickets already exist for this item
		var existingCount int64
		if err := tx.Model(&models.Ticket{}).
			Where("order_item_id = ?", item.ID).
			Count(&existingCount).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to check existing tickets: %w", err)
		}

		if existingCount > 0 {
			// Tickets already generated, skip
			continue
		}

		// Generate tickets for this item
		for i := 0; i < item.Quantity; i++ {
			ticket := models.Ticket{
				OrderItemID:  item.ID,
				TicketNumber: generateTicketNumber(item.TicketClass.EventID, order.ID, item.ID, i),
				HolderName:   fmt.Sprintf("%s %s", order.FirstName, order.LastName),
				HolderEmail:  order.Email,
				QRCode:       generateQRCode(item.TicketClass.EventID, order.ID, i),
				Status:       models.TicketActive,
			}

			if err := tx.Create(&ticket).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create ticket: %w", err)
			}
		}
	}

	// Step 3: Update order to fulfilled status
	order.Status = models.OrderFulfilled
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to mark order as fulfilled: %w", err)
	}

	// Commit transaction - both payment and tickets succeed together
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Track metrics (after successful commit)
	if h.metrics != nil {
		h.metrics.OrdersCompleted.WithLabelValues(paymentMethod).Inc()

		for _, item := range order.OrderItems {
			h.metrics.TicketsGenerated.WithLabelValues(
				fmt.Sprintf("%d", item.TicketClass.EventID),
				fmt.Sprintf("%d", order.ID),
			).Add(float64(item.Quantity))
		}
	}

	return nil
}

// Helper function to generate ticket number
func generateTicketNumber(eventID, orderID, itemID uint, index int) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("TKT-%d-%d-%d-%d-%d", eventID, orderID, itemID, index, timestamp)
}

// Helper function to generate QR code data
func generateQRCode(eventID, orderID uint, index int) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("TICKET:EVENT%d:ORDER%d:IDX%d:TIME%d", eventID, orderID, index, timestamp)
}

// VerifyPaymentAndGenerateTickets is a wrapper for the atomic operation
// This is the main entry point for processing verified payments
func (h *OrderHandler) VerifyPaymentAndGenerateTickets(
	order models.Order,
	paymentMethod string,
	paymentResult map[string]interface{},
) error {
	return h.ProcessPaymentWithTickets(order.ID, paymentMethod, paymentResult)
}

// RollbackPayment rolls back a payment transaction if ticket generation fails
// This should only be called in exceptional circumstances where the transaction
// failed after payment was already processed externally
func (h *OrderHandler) RollbackPayment(orderID uint, reason string) error {
	tx := h.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start rollback transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get order
	var order models.Order
	if err := tx.First(&order, orderID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to find order: %w", err)
	}

	// Delete any tickets that may have been created
	if err := tx.Where("order_item_id IN (?)",
		tx.Model(&models.OrderItem{}).Select("id").Where("order_id = ?", orderID),
	).Delete(&models.Ticket{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete tickets: %w", err)
	}

	// Revert order status
	order.Status = models.OrderPending
	order.PaymentStatus = models.PaymentPending
	order.IsPaymentReceived = false
	order.CompletedAt = nil

	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to revert order status: %w", err)
	}

	// Log the rollback
	fmt.Printf("⚠️ Payment rolled back for order %d: %s\n", orderID, reason)

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit rollback: %w", err)
	}

	return nil
}
