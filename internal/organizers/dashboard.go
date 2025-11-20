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

	h.db.Model(&models.Event{}).Where("organizer_id = ?", organizer.ID).Count(&totalEvents)
	h.db.Model(&models.Event{}).Where("organizer_id = ? AND status IN ?", organizer.ID, []string{"live", "pending_approval"}).Count(&activeEvents)

	// TODO: Calculate total tickets sold and revenue from orders/tickets
	// This would require joins with orders, order_items, and tickets tables
	totalTicketsSold = 0  // Placeholder
	totalRevenue := 0.0   // Placeholder
	pendingPayouts := 0.0 // Placeholder

	// Get recent events (last 5)
	var events []models.Event
	h.db.Where("organizer_id = ?", organizer.ID).Order("created_at DESC").Limit(5).Find(&events)

	var recentEvents []EventSummary
	for _, event := range events {
		recentEvents = append(recentEvents, EventSummary{
			ID:          event.ID,
			Title:       event.Title,
			StartDate:   event.StartDate,
			Status:      string(event.Status),
			TicketsSold: 0, // TODO: Calculate from actual sales
			Revenue:     0, // TODO: Calculate from actual sales
		})
	}

	response := OrganizerDashboardResponse{
		TotalEvents:      int(totalEvents),
		ActiveEvents:     int(activeEvents),
		TotalTicketsSold: int(totalTicketsSold),
		TotalRevenue:     totalRevenue,
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

	// TODO: Calculate revenue and ticket stats from actual sales data
	stats := QuickStats{}
	stats.ThisMonth.Events = int(thisMonthEvents)
	stats.ThisMonth.Revenue = 0.0 // Placeholder
	stats.ThisMonth.Tickets = 0   // Placeholder

	stats.LastMonth.Events = int(lastMonthEvents)
	stats.LastMonth.Revenue = 0.0 // Placeholder
	stats.LastMonth.Tickets = 0   // Placeholder

	json.NewEncoder(w).Encode(stats)
}
