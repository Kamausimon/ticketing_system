package tickets

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"time"
)

// BulkExportRequest represents a bulk export request for tickets
type BulkExportRequest struct {
	EventID   uint                 `json:"event_id"`
	TicketIDs []uint               `json:"ticket_ids,omitempty"`
	Format    string               `json:"format"` // "csv", "json"
	IncludeQR bool                 `json:"include_qr,omitempty"`
	Filters   *TicketExportFilters `json:"filters,omitempty"`
}

// TicketExportFilters represents filters for ticket export
type TicketExportFilters struct {
	Status         string `json:"status,omitempty"` // "active", "used", "transferred", "refunded"
	TicketClassIDs []uint `json:"ticket_class_ids,omitempty"`
	IsCheckedIn    *bool  `json:"is_checked_in,omitempty"`
	IsTransferred  *bool  `json:"is_transferred,omitempty"`
	IsRefunded     *bool  `json:"is_refunded,omitempty"`
	DateFrom       string `json:"date_from,omitempty"` // YYYY-MM-DD
	DateTo         string `json:"date_to,omitempty"`   // YYYY-MM-DD
}

// BulkExportTickets exports tickets in bulk
func (h *TicketHandler) BulkExportTickets(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request
	var req BulkExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate input
	if req.EventID == 0 {
		middleware.WriteJSONError(w, http.StatusBadRequest, "event_id is required")
		return
	}

	if req.Format == "" {
		req.Format = "csv" // Default to CSV
	}

	// Get user and verify organizer status
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Verify user owns the event
	var event models.Event
	if err := h.db.Where("id = ? AND account_id = ?", req.EventID, user.AccountID).First(&event).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied or event not found")
		return
	}

	// Build query for tickets
	query := h.db.Model(&models.Ticket{}).
		Preload("OrderItem.TicketClass").
		Preload("OrderItem.Order").
		Preload("TransferHistory").
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.event_id = ?", req.EventID)

	// Apply specific ticket IDs if provided
	if len(req.TicketIDs) > 0 {
		query = query.Where("tickets.id IN ?", req.TicketIDs)
	}

	// Apply filters if provided
	if req.Filters != nil {
		if req.Filters.Status != "" {
			query = query.Where("tickets.status = ?", req.Filters.Status)
		}
		if len(req.Filters.TicketClassIDs) > 0 {
			query = query.Where("order_items.ticket_class_id IN ?", req.Filters.TicketClassIDs)
		}
		if req.Filters.IsCheckedIn != nil {
			if *req.Filters.IsCheckedIn {
				query = query.Where("tickets.checked_in_at IS NOT NULL")
			} else {
				query = query.Where("tickets.checked_in_at IS NULL")
			}
		}
		if req.Filters.IsTransferred != nil {
			if *req.Filters.IsTransferred {
				query = query.Where("EXISTS (SELECT 1 FROM ticket_transfer_histories WHERE ticket_transfer_histories.ticket_id = tickets.id)")
			} else {
				query = query.Where("NOT EXISTS (SELECT 1 FROM ticket_transfer_histories WHERE ticket_transfer_histories.ticket_id = tickets.id)")
			}
		}
		if req.Filters.IsRefunded != nil {
			if *req.Filters.IsRefunded {
				query = query.Where("tickets.status = ?", models.TicketRefunded)
			} else {
				query = query.Where("tickets.status != ?", models.TicketRefunded)
			}
		}
		if req.Filters.DateFrom != "" {
			dateFrom, err := time.Parse("2006-01-02", req.Filters.DateFrom)
			if err == nil {
				query = query.Where("tickets.created_at >= ?", dateFrom)
			}
		}
		if req.Filters.DateTo != "" {
			dateTo, err := time.Parse("2006-01-02", req.Filters.DateTo)
			if err == nil {
				query = query.Where("tickets.created_at <= ?", dateTo.Add(24*time.Hour))
			}
		}
	}

	// Get tickets
	var tickets []models.Ticket
	if err := query.Find(&tickets).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch tickets")
		return
	}

	if len(tickets) == 0 {
		middleware.WriteJSONError(w, http.StatusNotFound, "no tickets found matching criteria")
		return
	}

	// Export based on format
	switch strings.ToLower(req.Format) {
	case "csv":
		h.exportTicketsCSV(w, tickets, event.Title)
	case "json":
		h.exportTicketsJSON(w, tickets)
	default:
		middleware.WriteJSONError(w, http.StatusBadRequest, "unsupported format. Use 'csv' or 'json'")
	}
}

// exportTicketsCSV exports tickets as CSV
func (h *TicketHandler) exportTicketsCSV(w http.ResponseWriter, tickets []models.Ticket, eventTitle string) {
	filename := fmt.Sprintf("tickets_%s_%s.csv",
		strings.ReplaceAll(eventTitle, " ", "_"),
		time.Now().Format("20060102"))

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{
		"Ticket ID",
		"Ticket Number",
		"Ticket Class",
		"Price",
		"Status",
		"Owner Email",
		"Owner Name",
		"Attendee Name",
		"Attendee Email",
		"Is Checked In",
		"Check-in Time",
		"Is Transferred",
		"Is Refunded",
		"QR Code",
		"Created At",
		"Updated At",
	}
	writer.Write(header)

	// Write data
	for _, ticket := range tickets {
		checkinTime := ""
		if ticket.CheckedInAt != nil {
			checkinTime = ticket.CheckedInAt.Format("2006-01-02 15:04:05")
		}

		ticketClass := ""
		price := ""
		if ticket.OrderItem.ID > 0 && ticket.OrderItem.TicketClass.ID > 0 {
			ticketClass = ticket.OrderItem.TicketClass.Name
			price = fmt.Sprintf("%.2f", float64(ticket.OrderItem.TicketClass.Price)/100) // Convert from cents
		}

		ownerEmail := ticket.OrderItem.Order.Email
		ownerName := fmt.Sprintf("%s %s", ticket.OrderItem.Order.FirstName, ticket.OrderItem.Order.LastName)

		attendeeName := ""
		attendeeEmail := ""
		// Get attendee info if exists
		var attendee models.Attendee
		if err := h.db.Where("ticket_id = ?", ticket.ID).First(&attendee).Error; err == nil {
			attendeeName = fmt.Sprintf("%s %s", attendee.FirstName, attendee.LastName)
			attendeeEmail = attendee.Email
		}

		// Calculate boolean fields
		isCheckedIn := ticket.CheckedInAt != nil
		isRefunded := ticket.Status == models.TicketRefunded
		// Check if ticket has transfer history (using preloaded data)
		isTransferred := len(ticket.TransferHistory) > 0

		row := []string{
			strconv.Itoa(int(ticket.ID)),
			ticket.TicketNumber,
			ticketClass,
			price,
			string(ticket.Status),
			ownerEmail,
			ownerName,
			attendeeName,
			attendeeEmail,
			strconv.FormatBool(isCheckedIn),
			checkinTime,
			strconv.FormatBool(isTransferred),
			strconv.FormatBool(isRefunded),
			ticket.QRCode,
			ticket.CreatedAt.Format("2006-01-02 15:04:05"),
			ticket.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		writer.Write(row)
	}
}

// exportTicketsJSON exports tickets as JSON
func (h *TicketHandler) exportTicketsJSON(w http.ResponseWriter, tickets []models.Ticket) {
	w.Header().Set("Content-Type", "application/json")

	type TicketExport struct {
		ID            uint       `json:"id"`
		TicketNumber  string     `json:"ticket_number"`
		TicketClass   string     `json:"ticket_class"`
		Price         float64    `json:"price"`
		Status        string     `json:"status"`
		OwnerEmail    string     `json:"owner_email"`
		OwnerName     string     `json:"owner_name"`
		AttendeeName  string     `json:"attendee_name,omitempty"`
		AttendeeEmail string     `json:"attendee_email,omitempty"`
		IsCheckedIn   bool       `json:"is_checked_in"`
		CheckInTime   *time.Time `json:"check_in_time,omitempty"`
		IsTransferred bool       `json:"is_transferred"`
		IsRefunded    bool       `json:"is_refunded"`
		QRCodeURL     string     `json:"qr_code_url"`
		CreatedAt     time.Time  `json:"created_at"`
		UpdatedAt     time.Time  `json:"updated_at"`
	}

	exports := make([]TicketExport, len(tickets))
	for i, ticket := range tickets {
		isCheckedIn := ticket.CheckedInAt != nil
		isRefunded := ticket.Status == models.TicketRefunded
		// Check if ticket has transfer history (using preloaded data)
		isTransferred := len(ticket.TransferHistory) > 0

		export := TicketExport{
			ID:            ticket.ID,
			TicketNumber:  ticket.TicketNumber,
			Status:        string(ticket.Status),
			IsCheckedIn:   isCheckedIn,
			CheckInTime:   ticket.CheckedInAt,
			IsTransferred: isTransferred,
			IsRefunded:    isRefunded,
			QRCodeURL:     ticket.QRCode,
			CreatedAt:     ticket.CreatedAt,
			UpdatedAt:     ticket.UpdatedAt,
		}

		if ticket.OrderItem.ID > 0 && ticket.OrderItem.TicketClass.ID > 0 {
			export.TicketClass = ticket.OrderItem.TicketClass.Name
			export.Price = float64(ticket.OrderItem.TicketClass.Price) / 100 // Convert from cents
		}

		// Get user info from order
		export.OwnerEmail = ticket.OrderItem.Order.Email
		export.OwnerName = fmt.Sprintf("%s %s",
			ticket.OrderItem.Order.FirstName,
			ticket.OrderItem.Order.LastName)

		// Get attendee info
		var attendee models.Attendee
		if err := h.db.Where("ticket_id = ?", ticket.ID).First(&attendee).Error; err == nil {
			export.AttendeeName = fmt.Sprintf("%s %s", attendee.FirstName, attendee.LastName)
			export.AttendeeEmail = attendee.Email
		}

		exports[i] = export
	}

	json.NewEncoder(w).Encode(exports)
}

// BulkTicketStats provides statistics about tickets for an event
type BulkTicketStats struct {
	EventID            uint               `json:"event_id"`
	TotalTickets       int                `json:"total_tickets"`
	ActiveTickets      int                `json:"active_tickets"`
	UsedTickets        int                `json:"used_tickets"`
	TransferredTickets int                `json:"transferred_tickets"`
	RefundedTickets    int                `json:"refunded_tickets"`
	CheckedInTickets   int                `json:"checked_in_tickets"`
	TotalRevenue       float64            `json:"total_revenue"`
	TicketsByClass     []TicketClassStats `json:"tickets_by_class"`
}

// TicketClassStats provides statistics per ticket class
type TicketClassStats struct {
	TicketClassID   uint    `json:"ticket_class_id"`
	TicketClassName string  `json:"ticket_class_name"`
	TotalSold       int     `json:"total_sold"`
	CheckedIn       int     `json:"checked_in"`
	Revenue         float64 `json:"revenue"`
}

// GetBulkTicketStats returns statistics about tickets for an event
func (h *TicketHandler) GetBulkTicketStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get event ID from query
	eventIDStr := r.URL.Query().Get("event_id")
	if eventIDStr == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "event_id is required")
		return
	}

	var eventID uint
	if _, err := fmt.Sscanf(eventIDStr, "%d", &eventID); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid event_id")
		return
	}

	// Get user and verify organizer status
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Verify user owns the event
	var event models.Event
	if err := h.db.Where("id = ? AND account_id = ?", eventID, user.AccountID).First(&event).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied or event not found")
		return
	}

	// Get ticket statistics
	var stats BulkTicketStats
	stats.EventID = eventID

	baseQuery := h.db.Model(&models.Ticket{}).
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.event_id = ?", eventID)

	// Total tickets
	var total int64
	baseQuery.Count(&total)
	stats.TotalTickets = int(total)

	// Count by status
	var active, used int64
	h.db.Model(&models.Ticket{}).
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.event_id = ? AND tickets.status = ?", eventID, models.TicketActive).
		Count(&active)
	stats.ActiveTickets = int(active)

	h.db.Model(&models.Ticket{}).
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.event_id = ? AND tickets.status = ?", eventID, models.TicketUsed).
		Count(&used)
	stats.UsedTickets = int(used)

	// Count transferred tickets (tickets with transfer history)
	var transferred int64
	h.db.Table("tickets").
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.event_id = ? AND EXISTS (SELECT 1 FROM ticket_transfer_histories WHERE ticket_transfer_histories.ticket_id = tickets.id)", eventID).
		Count(&transferred)
	stats.TransferredTickets = int(transferred)

	var refunded, checkedIn int64
	h.db.Model(&models.Ticket{}).
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.event_id = ? AND tickets.status = ?", eventID, models.TicketRefunded).
		Count(&refunded)
	stats.RefundedTickets = int(refunded)

	h.db.Model(&models.Ticket{}).
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.event_id = ? AND tickets.checked_in_at IS NOT NULL", eventID).
		Count(&checkedIn)
	stats.CheckedInTickets = int(checkedIn)

	// Calculate total revenue (sum of ticket prices for non-refunded tickets)
	var revenueCents int64
	h.db.Table("tickets").
		Select("COALESCE(SUM(ticket_classes.price), 0)").
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Joins("JOIN ticket_classes ON ticket_classes.id = order_items.ticket_class_id").
		Where("orders.event_id = ? AND tickets.status != ?", eventID, models.TicketRefunded).
		Scan(&revenueCents)
	stats.TotalRevenue = float64(revenueCents) / 100

	// Get ticket class breakdown
	type ClassResult struct {
		TicketClassID   uint
		TicketClassName string
		TotalSold       int
		CheckedIn       int
		Revenue         float64
	}

	var classResults []ClassResult
	h.db.Table("tickets").
		Select(`
			ticket_classes.id as ticket_class_id,
			ticket_classes.name as ticket_class_name,
			COUNT(tickets.id) as total_sold,
			SUM(CASE WHEN tickets.checked_in_at IS NOT NULL THEN 1 ELSE 0 END) as checked_in,
			SUM(CASE WHEN tickets.status != 'refunded' THEN ticket_classes.price ELSE 0 END) as revenue
		`).
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Joins("JOIN ticket_classes ON ticket_classes.id = order_items.ticket_class_id").
		Where("orders.event_id = ?", eventID).
		Group("ticket_classes.id, ticket_classes.name").
		Scan(&classResults)

	stats.TicketsByClass = make([]TicketClassStats, len(classResults))
	for i, result := range classResults {
		stats.TicketsByClass[i] = TicketClassStats{
			TicketClassID:   result.TicketClassID,
			TicketClassName: result.TicketClassName,
			TotalSold:       result.TotalSold,
			CheckedIn:       result.CheckedIn,
			Revenue:         result.Revenue / 100, // Convert from cents
		}
	}

	json.NewEncoder(w).Encode(stats)
}

// BulkUpdateTicketStatus updates the status of multiple tickets
type BulkUpdateTicketStatusRequest struct {
	TicketIDs     []uint   `json:"ticket_ids,omitempty"`
	TicketNumbers []string `json:"ticket_numbers,omitempty"`
	Status        string   `json:"status"` // "active", "used", "cancelled"
}

// BulkUpdateTicketStatus updates the status of multiple tickets
func (h *TicketHandler) BulkUpdateTicketStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request
	var req BulkUpdateTicketStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate input
	if len(req.TicketIDs) == 0 && len(req.TicketNumbers) == 0 {
		middleware.WriteJSONError(w, http.StatusBadRequest, "ticket_ids or ticket_numbers is required")
		return
	}

	validStatuses := map[string]bool{"active": true, "used": true, "cancelled": true}
	if !validStatuses[req.Status] {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid status. Use 'active', 'used', or 'cancelled'")
		return
	}

	// Get user and verify organizer status
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Build query based on whether we have IDs or ticket numbers
	var tickets []models.Ticket
	query := h.db.Preload("OrderItem.Order.Event")

	if len(req.TicketIDs) > 0 {
		query = query.Where("id IN ?", req.TicketIDs)
	} else {
		query = query.Where("ticket_number IN ?", req.TicketNumbers)
	}

	if err := query.Find(&tickets).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch tickets")
		return
	}

	if len(tickets) == 0 {
		middleware.WriteJSONError(w, http.StatusNotFound, "no tickets found")
		return
	}

	// Verify all tickets belong to events owned by the user
	ticketIDs := make([]uint, 0, len(tickets))
	for _, ticket := range tickets {
		if ticket.OrderItem.Order.Event.AccountID != user.AccountID {
			middleware.WriteJSONError(w, http.StatusForbidden, "access denied to one or more tickets")
			return
		}
		ticketIDs = append(ticketIDs, ticket.ID)
	}

	// Update ticket statuses
	result := h.db.Model(&models.Ticket{}).
		Where("id IN ?", ticketIDs).
		Update("status", req.Status)

	if result.Error != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update tickets")
		return
	}

	response := map[string]interface{}{
		"updated_count": result.RowsAffected,
		"status":        req.Status,
		"message":       fmt.Sprintf("Successfully updated %d tickets to '%s'", result.RowsAffected, req.Status),
	}

	json.NewEncoder(w).Encode(response)
}
