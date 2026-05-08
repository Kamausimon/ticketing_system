package orders

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	apievents "ticketing_system/internal/api_events"
	kafkatopics "ticketing_system/internal/kafka"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CreateOrder handles creating a new order
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
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

	// Parse request
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if err := validateCreateOrderRequest(req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get user to access AccountID
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Verify event exists and is live
	var event models.Event
	if err := h.db.Where("id = ? AND is_live = ?", req.EventID, true).First(&event).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "event not found or not available for booking")
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

	// Validate ticket availability and calculate order
	calculation, err := h.calculateOrderTotal(tx, req)
	if err != nil {
		tx.Rollback()
		middleware.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create order
	now := time.Now()
	order := models.Order{
		AccountID:         user.AccountID,
		EventID:           req.EventID,
		FirstName:         strings.TrimSpace(req.FirstName),
		LastName:          strings.TrimSpace(req.LastName),
		Email:             strings.ToLower(strings.TrimSpace(req.Email)),
		IsBusiness:        req.IsBusiness,
		Amount:            float32(calculation.Subtotal),
		TaxAmount:         float32(calculation.TaxAmount),
		TotalAmount:       models.Money(calculation.TotalAmount * 100), // Convert to cents
		Currency:          calculation.Currency,
		Status:            models.OrderPending,
		PaymentStatus:     models.PaymentPending,
		OrderDate:         &now,
		IsPaymentReceived: false,
	}

	// Set optional business fields
	if req.IsBusiness && req.BusinessName != "" {
		order.BusinessName = &req.BusinessName
		if req.BusinessTaxID != "" {
			order.BusinessTaxNumber = &req.BusinessTaxID
		}
		if req.BusinessAddress != "" {
			order.BusinessAddressLine = &req.BusinessAddress
		}
	}

	// Set fees and discounts
	if calculation.BookingFee > 0 {
		fee := float32(calculation.BookingFee)
		order.BookingFee = &fee
	}
	if calculation.OrganizerBookingFee > 0 {
		orgFee := float32(calculation.OrganizerBookingFee)
		order.OrganizerBookingFee = &orgFee
	}
	if calculation.Discount > 0 {
		discount := float32(calculation.Discount)
		order.Discount = &discount
	}

	// Create order
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to create order")
		return
	}

	// Create order items with pessimistic locking to prevent race conditions
	for _, itemReq := range req.Items {
		var ticketClass models.TicketClass

		// Use FOR UPDATE to lock the row and prevent concurrent modifications
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", itemReq.TicketClassID).First(&ticketClass).Error; err != nil {
			tx.Rollback()
			middleware.WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("ticket class %d not found", itemReq.TicketClassID))
			return
		}

		// Check if ticket class is paused or hidden
		if ticketClass.IsPaused {
			tx.Rollback()
			middleware.WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("ticket class '%s' is currently unavailable", ticketClass.Name))
			return
		}

		// Check sale dates
		if ticketClass.StartSaleDate != nil && now.Before(*ticketClass.StartSaleDate) {
			tx.Rollback()
			middleware.WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("ticket class '%s' is not yet on sale", ticketClass.Name))
			return
		}
		if ticketClass.EndSaleDate != nil && now.After(*ticketClass.EndSaleDate) {
			tx.Rollback()
			middleware.WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("ticket class '%s' sale has ended", ticketClass.Name))
			return
		}

		// Re-check availability after acquiring lock (another transaction may have purchased)
		if ticketClass.QuantityAvailable != nil {
			available := *ticketClass.QuantityAvailable - ticketClass.QuantitySold
			if available < itemReq.Quantity {
				tx.Rollback()
				middleware.WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("only %d tickets available for '%s'", available, ticketClass.Name))
				return
			}
		}

		unitPrice := float64(ticketClass.Price)
		totalPrice := unitPrice * float64(itemReq.Quantity)

		orderItem := models.OrderItem{
			OrderID:       order.ID,
			TicketClassID: itemReq.TicketClassID,
			Quantity:      itemReq.Quantity,
			UnitPrice:     models.Money(unitPrice),
			TotalPrice:    models.Money(totalPrice),
		}

		if req.PromoCode != "" {
			orderItem.PromoCodeUsed = &req.PromoCode
		}

		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to create order item")
			return
		}

		// Atomically update ticket class sold quantity using database-level increment
		// with optimistic locking to detect concurrent modifications
		// This prevents race conditions even if lock is somehow bypassed
		result := tx.Model(&models.TicketClass{}).
			Where("id = ? AND version = ?", itemReq.TicketClassID, ticketClass.Version).
			Updates(map[string]interface{}{
				"quantity_sold": gorm.Expr("quantity_sold + ?", itemReq.Quantity),
				"version":       gorm.Expr("version + 1"),
			})

		if result.Error != nil {
			tx.Rollback()
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update ticket inventory")
			return
		}

		if result.RowsAffected == 0 {
			tx.Rollback()
			middleware.WriteJSONError(w, http.StatusConflict, "ticket inventory changed during checkout, please try again")
			return
		}
	}

	// Delete any reservations for this user and these tickets (reservation converted to order)
	sessionID := fmt.Sprintf("user_%d", userID)
	tx.Where("session_id = ? AND event_id = ?", sessionID, req.EventID).Delete(&models.ReservedTicket{})

	// Build the outbox event — must happen inside the transaction so the event and the order either both commit or both roll back.
	if err := h.publishOrderCreatedEvent(tx, order, req, calculation); err != nil {
		tx.Rollback()
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to queue order event")
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to complete order")
		return
	}
	committed = true

	// Track metrics - order created
	if h.metrics != nil {
		h.metrics.TrackOrderCreated(string(models.OrderPending))
		h.metrics.OrderValue.WithLabelValues(calculation.Currency).Observe(calculation.TotalAmount)
	}

	// Load order with relations
	var createdOrder models.Order
	h.db.Preload("Event").Preload("OrderItems.TicketClass").First(&createdOrder, order.ID)

	response := map[string]interface{}{
		"message":     "Order created successfully",
		"order":       convertToOrderResponse(createdOrder),
		"calculation": calculation,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// CalculateOrder handles calculating order total without creating it
func (h *OrderHandler) CalculateOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate basic request
	if req.EventID == 0 || len(req.Items) == 0 {
		middleware.WriteJSONError(w, http.StatusBadRequest, "event_id and items are required")
		return
	}

	// Calculate order total
	calculation, err := h.calculateOrderTotal(h.db, req)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	json.NewEncoder(w).Encode(calculation)
}

// validateCreateOrderRequest validates the create order request
func validateCreateOrderRequest(req CreateOrderRequest) error {
	if req.EventID == 0 {
		return fmt.Errorf("event_id is required")
	}
	if strings.TrimSpace(req.FirstName) == "" {
		return fmt.Errorf("first_name is required")
	}
	if strings.TrimSpace(req.LastName) == "" {
		return fmt.Errorf("last_name is required")
	}
	if strings.TrimSpace(req.Email) == "" {
		return fmt.Errorf("email is required")
	}
	if !isValidEmail(req.Email) {
		return fmt.Errorf("invalid email format")
	}
	if len(req.Items) == 0 {
		return fmt.Errorf("at least one order item is required")
	}
	for i, item := range req.Items {
		if item.TicketClassID == 0 {
			return fmt.Errorf("item %d: ticket_class_id is required", i)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("item %d: quantity must be greater than 0", i)
		}
	}
	if req.IsBusiness && strings.TrimSpace(req.BusinessName) == "" {
		return fmt.Errorf("business_name is required for business orders")
	}
	if req.PaymentMethod == "" {
		return fmt.Errorf("payment_method is required")
	}

	return nil
}

// isValidEmail validates email format
func isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	return len(email) > 3 && strings.Contains(email, "@") && strings.Contains(email, ".")
}

// publishOrderCreatedEvent builds the OrderCreatedEvent payload and writes it
// to the outbox table using the provided transaction.
func (h *OrderHandler) publishOrderCreatedEvent(tx *gorm.DB, order models.Order, req CreateOrderRequest, calc *OrderCalculation) error {
	items := make([]apievents.OrderEventItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = apievents.OrderEventItem{
			TicketClassID: fmt.Sprintf("%d", item.TicketClassID),
			Quantity:      item.Quantity,
			UnitPrice:     calc.Subtotal / float64(item.Quantity),
			TotalPrice:    calc.Subtotal,
		}
	}

	evt := apievents.OrderCreatedEvent{
		OrderID:       fmt.Sprintf("%d", order.ID),
		OrderNumber:   generateOrderNumber(order.ID),
		AccountID:     fmt.Sprintf("%d", order.AccountID),
		EventID:       fmt.Sprintf("%d", order.EventID),
		FirstName:     order.FirstName,
		LastName:      order.LastName,
		Email:         order.Email,
		Amount:        float64(order.Amount),
		TaxAmount:     float64(order.TaxAmount),
		TotalAmount:   calc.TotalAmount,
		Currency:      order.Currency,
		Status:        string(order.Status),
		PaymentStatus: string(order.PaymentStatus),
		PaymentMethod: req.PaymentMethod,
		IsBusiness:    order.IsBusiness,
		BusinessName:  order.BusinessName,
		Items:         items,
		Timestamp:     time.Now(),
	}

	payload, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	return h.outboxRepo.Save(tx, models.OutboxEvent{
		Topic:   kafkatopics.OrderCreatedTopic,
		Payload: string(payload),
		Status:  "pending",
	})
}
