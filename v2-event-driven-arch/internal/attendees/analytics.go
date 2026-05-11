package attendees

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"ticketing_system/internal/models"
)

// GetAttendanceStats returns detailed attendance statistics
func (h *AttendeeHandler) GetAttendanceStats(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.URL.Query().Get("event_id")
	if eventIDStr == "" {
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	// Get total attendees
	var total int64
	h.db.Model(&models.Attendee{}).
		Where("event_id = ?", uint(eventID)).
		Count(&total)

	// Get checked in count
	var checkedIn int64
	h.db.Model(&models.Attendee{}).
		Where("event_id = ? AND has_arrived = ?", uint(eventID), true).
		Count(&checkedIn)

	// Get not checked in count
	notCheckedIn := total - checkedIn

	// Get refunded tickets
	var refunded int64
	h.db.Model(&models.Attendee{}).
		Where("event_id = ? AND is_refunded = ?", uint(eventID), true).
		Count(&refunded)

	// Calculate check-in rate
	checkInRate := 0.0
	if total > 0 {
		checkInRate = float64(checkedIn) / float64(total) * 100
	}

	// Get hourly arrival trend
	arrivalTrend := h.getHourlyArrivalTrend(uint(eventID))

	stats := AttendanceStats{
		TotalAttendees:  total,
		CheckedIn:       checkedIn,
		NotCheckedIn:    notCheckedIn,
		RefundedTickets: refunded,
		CheckInRate:     checkInRate,
		ArrivalTrend:    arrivalTrend,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// getHourlyArrivalTrend calculates arrivals per hour
func (h *AttendeeHandler) getHourlyArrivalTrend(eventID uint) []HourlyArrival {
	var attendees []models.Attendee
	h.db.Where("event_id = ? AND has_arrived = ? AND arrival_time IS NOT NULL", eventID, true).
		Find(&attendees)

	hourlyMap := make(map[int]int64)
	for _, attendee := range attendees {
		if attendee.ArrivalTime != nil {
			hour := attendee.ArrivalTime.Hour()
			hourlyMap[hour]++
		}
	}

	trend := []HourlyArrival{}
	for hour := 0; hour < 24; hour++ {
		if count, exists := hourlyMap[hour]; exists {
			trend = append(trend, HourlyArrival{
				Hour:  hour,
				Count: count,
			})
		}
	}

	return trend
}

// GetCheckInReport generates a check-in report
func (h *AttendeeHandler) GetCheckInReport(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.URL.Query().Get("event_id")
	if eventIDStr == "" {
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	// Get check-in statistics by ticket type
	type TicketTypeStats struct {
		TicketType  string  `json:"ticket_type"`
		Total       int64   `json:"total"`
		CheckedIn   int64   `json:"checked_in"`
		CheckInRate float64 `json:"check_in_rate"`
	}

	var ticketStats []TicketTypeStats
	h.db.Raw(`
		SELECT 
			tc.name as ticket_type,
			COUNT(a.id) as total,
			COUNT(CASE WHEN a.has_arrived = true THEN 1 END) as checked_in,
			(COUNT(CASE WHEN a.has_arrived = true THEN 1 END)::float / COUNT(a.id)::float * 100) as check_in_rate
		FROM attendees a
		JOIN tickets t ON t.id = a.ticket_id
		JOIN order_items oi ON oi.id = t.order_item_id
		JOIN tickets_classes tc ON tc.id = oi.ticket_class_id
		WHERE a.event_id = ? AND a.deleted_at IS NULL
		GROUP BY tc.name
		ORDER BY tc.name
	`, uint(eventID)).Scan(&ticketStats)

	// Get peak check-in times
	type PeakTime struct {
		Hour  int   `json:"hour"`
		Count int64 `json:"count"`
	}

	var peakTimes []PeakTime
	h.db.Raw(`
		SELECT 
			EXTRACT(HOUR FROM arrival_time)::int as hour,
			COUNT(*) as count
		FROM attendees
		WHERE event_id = ? AND has_arrived = true AND arrival_time IS NOT NULL
		GROUP BY EXTRACT(HOUR FROM arrival_time)
		ORDER BY count DESC
		LIMIT 5
	`, uint(eventID)).Scan(&peakTimes)

	report := map[string]interface{}{
		"ticket_type_stats": ticketStats,
		"peak_times":        peakTimes,
		"generated_at":      time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// GetAttendanceTimeline returns check-in timeline
func (h *AttendeeHandler) GetAttendanceTimeline(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.URL.Query().Get("event_id")
	if eventIDStr == "" {
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	type TimelineEntry struct {
		Time  time.Time `json:"time"`
		Count int64     `json:"count"`
	}

	var timeline []TimelineEntry
	h.db.Raw(`
		SELECT 
			DATE_TRUNC('minute', arrival_time) as time,
			COUNT(*) as count
		FROM attendees
		WHERE event_id = ? AND has_arrived = true AND arrival_time IS NOT NULL
		GROUP BY DATE_TRUNC('minute', arrival_time)
		ORDER BY time ASC
	`, uint(eventID)).Scan(&timeline)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(timeline)
}

// GetNoShowList returns list of attendees who didn't check in
func (h *AttendeeHandler) GetNoShowList(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.URL.Query().Get("event_id")
	if eventIDStr == "" {
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	var attendees []models.Attendee
	if err := h.db.Preload("Ticket").
		Preload("Ticket.OrderItem.TicketClass").
		Where("event_id = ? AND has_arrived = ? AND is_refunded = ?", uint(eventID), false, false).
		Order("last_name, first_name").
		Find(&attendees).Error; err != nil {
		http.Error(w, "Failed to fetch no-show list", http.StatusInternalServerError)
		return
	}

	responses := make([]AttendeeResponse, len(attendees))
	for i, attendee := range attendees {
		responses[i] = convertToAttendeeResponse(attendee)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}
