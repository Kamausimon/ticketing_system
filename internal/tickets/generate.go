package tickets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
)

// GenerateTickets handles generating tickets for a paid order
func (h *TicketHandler) GenerateTickets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request
	var req GenerateTicketsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get order with items and verify ownership
	var order models.Order
	if err := h.db.Preload("OrderItems.TicketClass.Event").
		Where("id = ?", req.OrderID).First(&order).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "order not found")
		return
	}

	// Verify order belongs to user
	if order.AccountID != user.AccountID {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Check if order is paid
	if order.Status != models.OrderPaid && order.Status != models.OrderFulfilled {
		middleware.WriteJSONError(w, http.StatusBadRequest, "order must be paid before generating tickets")
		return
	}

	// Check if tickets already generated
	var existingTicketsCount int64
	for _, item := range order.OrderItems {
		h.db.Model(&models.Ticket{}).Where("order_item_id = ?", item.ID).Count(&existingTicketsCount)
		if existingTicketsCount > 0 {
			middleware.WriteJSONError(w, http.StatusBadRequest, "tickets already generated for this order")
			return
		}
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var generatedTickets []models.Ticket

	// Generate tickets for each order item
	for _, item := range order.OrderItems {
		for i := 0; i < item.Quantity; i++ {
			// Create ticket
			ticket := models.Ticket{
				OrderItemID: item.ID,
				HolderName:  fmt.Sprintf("%s %s", order.FirstName, order.LastName),
				HolderEmail: order.Email,
				Status:      models.TicketActive,
			}

			// Save to get ID
			if err := tx.Create(&ticket).Error; err != nil {
				tx.Rollback()
				middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to create ticket")
				return
			}

			// Generate ticket number using IDs
			ticket.TicketNumber = generateTicketNumber(
				item.TicketClass.EventID,
				order.ID,
				ticket.ID,
			)

			// Generate QR code
			ticket.QRCode = generateQRCodeData(ticket.TicketNumber)
			ticket.BarcodeData = ticket.TicketNumber

			// Update ticket with generated data
			if err := tx.Save(&ticket).Error; err != nil {
				tx.Rollback()
				middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update ticket")
				return
			}

			generatedTickets = append(generatedTickets, ticket)
		}
	}

	// Update order status to fulfilled
	order.Status = models.OrderFulfilled
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update order status")
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to generate tickets")
		return
	}

	// Track metrics for ticket generation
	if h.metrics != nil {
		for _, item := range order.OrderItems {
			h.metrics.TicketsGenerated.WithLabelValues(
				fmt.Sprintf("%d", item.TicketClass.EventID),
				fmt.Sprintf("%d", order.ID),
			).Add(float64(item.Quantity))
		}
	}

	// Load full ticket details
	var tickets []models.Ticket
	h.db.Preload("OrderItem.TicketClass.Event").
		Where("order_item_id IN (?)",
			tx.Model(&models.OrderItem{}).Select("id").Where("order_id = ?", order.ID)).
		Find(&tickets)

	// Generate PDFs for all tickets asynchronously
	go func(ticketList []models.Ticket) {
		for i := range ticketList {
			if pdfPath, err := h.generateTicketPDF(&ticketList[i]); err == nil {
				// Update ticket with PDF path
				h.db.Model(&ticketList[i]).Update("pdf_path", pdfPath)
			} else {
				fmt.Printf("⚠️ Failed to generate PDF for ticket %s: %v\n", ticketList[i].TicketNumber, err)
			}
		}
	}(tickets)

	// Convert to response
	ticketResponses := make([]TicketResponse, len(tickets))
	for i, ticket := range tickets {
		ticketResponses[i] = convertToTicketResponse(ticket)
	}

	response := map[string]interface{}{
		"message":       "Tickets generated successfully",
		"tickets":       ticketResponses,
		"total_tickets": len(ticketResponses),
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// RegenerateTicketQR handles regenerating QR code for a ticket
func (h *TicketHandler) RegenerateTicketQR(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get ticket number from URL
	ticketNumber := r.URL.Query().Get("ticket_number")
	if ticketNumber == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "ticket_number is required")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get ticket with order
	var ticket models.Ticket
	if err := h.db.Preload("OrderItem.Order").
		Where("ticket_number = ?", ticketNumber).First(&ticket).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "ticket not found")
		return
	}

	// Verify ownership
	if ticket.OrderItem.Order.AccountID != user.AccountID {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Regenerate QR code
	ticket.QRCode = generateQRCodeData(ticket.TicketNumber)

	if err := h.db.Save(&ticket).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to regenerate QR code")
		return
	}

	response := map[string]interface{}{
		"message":       "QR code regenerated successfully",
		"ticket_number": ticket.TicketNumber,
		"qr_code":       ticket.QRCode,
	}

	json.NewEncoder(w).Encode(response)
}
