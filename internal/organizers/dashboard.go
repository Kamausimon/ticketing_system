package organizers

import (
	"encoding/json"
	"net/http"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"time"
)

// Dashboard-related structures
type OrganizerDashboardResponse struct {
	TotalEvents      int            `json:"total_events"`
	ActiveEvents     int            `json:"active_events"`
	TotalTicketsSold int            `json:"total_tickets_sold"`
	TotalRevenue     float64        `json:"total_revenue"`
	PendingPayouts   float64        `json:"pending_payouts"`
	RecentEvents     []EventSummary `json:"recent_events"`
}

type EventSummary struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	StartDate   time.Time `json:"start_date"`
	Status      string    `json:"status"`
	TicketsSold int       `json:"tickets_sold"`
	Revenue     float64   `json:"revenue"`
}

type QuickStats struct {
	ThisMonth struct {
		Events  int     `json:"events"`
		Revenue float64 `json:"revenue"`
		Tickets int     `json:"tickets"`
	} `json:"this_month"`
	LastMonth struct {
		Events  int     `json:"events"`
		Revenue float64 `json:"revenue"`
		Tickets int     `json:"tickets"`
	} `json:"last_month"`
}

// GetOrganizerDashboard returns dashboard statistics for organizer
func (h *OrganizerHandler) GetOrganizerDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)

	// Get organizer
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	var organizer models.Organizer
	if err := h.db.Where("account_id = ?", user.AccountID).First(&organizer).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "organizer profile not found")
		return
	}

	// Get dashboard statistics
	var totalEvents, activeEvents int64
	var totalTicketsSold int64
	var totalRevenue int64

	h.db.Model(&models.Event{}).Where("organizer_id = ?", organizer.ID).Count(&totalEvents)
	h.db.Model(&models.Event{}).Where("organizer_id = ? AND status IN ?", organizer.ID, []string{"live", "pending_approval"}).Count(&activeEvents)

	// Calculate total tickets sold and revenue from completed orders
	// Join through: Event -> Order -> OrderItem -> Ticket
	h.db.Model(&models.Ticket{}).
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Joins("JOIN events ON events.id = orders.event_id").
		Where("events.organizer_id = ? AND orders.status IN ?", organizer.ID, []string{"paid", "fulfilled"}).
		Count(&totalTicketsSold)

	// Calculate total revenue from completed orders
	h.db.Model(&models.Order{}).
		Joins("JOIN events ON events.id = orders.event_id").
		Where("events.organizer_id = ? AND orders.status IN ?", organizer.ID, []string{"paid", "fulfilled"}).
		Select("COALESCE(SUM(total_amount), 0)").
		Row().Scan(&totalRevenue)

	// Calculate pending payouts from settlement records
	var pendingPayoutsAmount int64
	h.db.Model(&models.SettlementItem{}).
		Where("organizer_id = ? AND status IN ?", organizer.ID, []string{"pending", "awaiting_event", "holding_period", "ready_to_process", "processing"}).
		Select("COALESCE(SUM(net_amount), 0)").
		Row().Scan(&pendingPayoutsAmount)
	pendingPayouts := float64(pendingPayoutsAmount) / 100.0

	// Get recent events (last 5)
	var events []models.Event
	h.db.Where("organizer_id = ?", organizer.ID).Order("created_at DESC").Limit(5).Find(&events)

	var recentEvents []EventSummary
	for _, event := range events {
		// Calculate event-specific tickets sold and revenue
		var eventTickets int64
		var eventRevenue int64

		h.db.Model(&models.Ticket{}).
			Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
			Joins("JOIN orders ON orders.id = order_items.order_id").
			Where("orders.event_id = ? AND orders.status IN ?", event.ID, []string{"paid", "fulfilled"}).
			Count(&eventTickets)

		h.db.Model(&models.Order{}).
			Where("event_id = ? AND status IN ?", event.ID, []string{"paid", "fulfilled"}).
			Select("COALESCE(SUM(total_amount), 0)").
			Row().Scan(&eventRevenue)

		recentEvents = append(recentEvents, EventSummary{
			ID:          event.ID,
			Title:       event.Title,
			StartDate:   event.StartDate,
			Status:      string(event.Status),
			TicketsSold: int(eventTickets),
			Revenue:     float64(eventRevenue) / 100.0, // Convert from cents to currency
		})
	}

	response := OrganizerDashboardResponse{
		TotalEvents:      int(totalEvents),
		ActiveEvents:     int(activeEvents),
		TotalTicketsSold: int(totalTicketsSold),
		TotalRevenue:     float64(totalRevenue) / 100.0, // Convert from cents to currency
		PendingPayouts:   pendingPayouts,
		RecentEvents:     recentEvents,
	}

	json.NewEncoder(w).Encode(response)
}

// GetQuickStats returns quick statistics for the organizer dashboard
func (h *OrganizerHandler) GetQuickStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)

	// Get organizer
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	var organizer models.Organizer
	if err := h.db.Where("account_id = ?", user.AccountID).First(&organizer).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "organizer profile not found")
		return
	}

	// Calculate date ranges
	now := time.Now()
	thisMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastMonthStart := thisMonthStart.AddDate(0, -1, 0)
	lastMonthEnd := thisMonthStart.Add(-time.Second)

	// Get this month's stats
	var thisMonthEvents, lastMonthEvents int64
	h.db.Model(&models.Event{}).Where("organizer_id = ? AND created_at >= ?", organizer.ID, thisMonthStart).Count(&thisMonthEvents)
	h.db.Model(&models.Event{}).Where("organizer_id = ? AND created_at >= ? AND created_at <= ?", organizer.ID, lastMonthStart, lastMonthEnd).Count(&lastMonthEvents)

	// Calculate this month's revenue and tickets
	var thisMonthRevenue, lastMonthRevenue int64
	var thisMonthTickets, lastMonthTickets int64

	h.db.Model(&models.Order{}).
		Joins("JOIN events ON events.id = orders.event_id").
		Where("events.organizer_id = ? AND orders.status IN ? AND orders.created_at >= ?", organizer.ID, []string{"paid", "fulfilled"}, thisMonthStart).
		Select("COALESCE(SUM(total_amount), 0)").
		Row().Scan(&thisMonthRevenue)

	h.db.Model(&models.Order{}).
		Joins("JOIN events ON events.id = orders.event_id").
		Where("events.organizer_id = ? AND orders.status IN ? AND orders.created_at >= ? AND orders.created_at <= ?", organizer.ID, []string{"paid", "fulfilled"}, lastMonthStart, lastMonthEnd).
		Select("COALESCE(SUM(total_amount), 0)").
		Row().Scan(&lastMonthRevenue)

	h.db.Model(&models.Ticket{}).
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Joins("JOIN events ON events.id = orders.event_id").
		Where("events.organizer_id = ? AND orders.status IN ? AND orders.created_at >= ?", organizer.ID, []string{"paid", "fulfilled"}, thisMonthStart).
		Count(&thisMonthTickets)

	h.db.Model(&models.Ticket{}).
		Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Joins("JOIN events ON events.id = orders.event_id").
		Where("events.organizer_id = ? AND orders.status IN ? AND orders.created_at >= ? AND orders.created_at <= ?", organizer.ID, []string{"paid", "fulfilled"}, lastMonthStart, lastMonthEnd).
		Count(&lastMonthTickets)

	stats := QuickStats{}
	stats.ThisMonth.Events = int(thisMonthEvents)
	stats.ThisMonth.Revenue = float64(thisMonthRevenue) / 100.0 // Convert from cents
	stats.ThisMonth.Tickets = int(thisMonthTickets)

	stats.LastMonth.Events = int(lastMonthEvents)
	stats.LastMonth.Revenue = float64(lastMonthRevenue) / 100.0 // Convert from cents
	stats.LastMonth.Tickets = int(lastMonthTickets)

	json.NewEncoder(w).Encode(stats)
}
