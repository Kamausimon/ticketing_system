package attendees

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"ticketing_system/internal/notifications"
	"time"
)

// BulkEmailRequest represents a bulk email request to attendees
type BulkEmailRequest struct {
	EventID     uint              `json:"event_id"`
	AttendeeIDs []uint            `json:"attendee_ids,omitempty"` // Specific attendees, or nil for all
	Subject     string            `json:"subject"`
	Message     string            `json:"message"`
	HTMLMessage string            `json:"html_message,omitempty"`
	Filters     *BulkEmailFilters `json:"filters,omitempty"`
}

// BulkEmailFilters represents filters for selecting attendees
type BulkEmailFilters struct {
	HasArrived     *bool  `json:"has_arrived,omitempty"`
	IsRefunded     *bool  `json:"is_refunded,omitempty"`
	TicketClassIDs []uint `json:"ticket_class_ids,omitempty"`
}

// BulkEmailResponse represents the result of a bulk email operation
type BulkEmailResponse struct {
	TotalSent    int      `json:"total_sent"`
	TotalFailed  int      `json:"total_failed"`
	FailedEmails []string `json:"failed_emails,omitempty"`
	Message      string   `json:"message"`
}

// SendBulkEmail sends emails to multiple attendees
func (h *AttendeeHandler) SendBulkEmail(w http.ResponseWriter, r *http.Request, notificationService *notifications.NotificationService) {
	w.Header().Set("Content-Type", "application/json")

	if notificationService == nil {
		middleware.WriteJSONError(w, http.StatusServiceUnavailable, "notification service not available")
		return
	}

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
	var req BulkEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate input
	if req.EventID == 0 {
		middleware.WriteJSONError(w, http.StatusBadRequest, "event_id is required")
		return
	}

	if req.Subject == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "subject is required")
		return
	}

	if req.Message == "" && req.HTMLMessage == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "message or html_message is required")
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
	if err := h.db.Where("id = ? AND account_id = ?", req.EventID, user.AccountID).First(&event).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied or event not found")
		return
	}

	// Build query for attendees
	query := h.db.Model(&models.Attendee{}).
		Preload("Event").
		Preload("Ticket").
		Where("event_id = ?", req.EventID)

	// Apply specific attendee IDs if provided
	if len(req.AttendeeIDs) > 0 {
		query = query.Where("id IN ?", req.AttendeeIDs)
	}

	// Apply filters if provided
	if req.Filters != nil {
		if req.Filters.HasArrived != nil {
			query = query.Where("has_arrived = ?", *req.Filters.HasArrived)
		}
		if req.Filters.IsRefunded != nil {
			query = query.Where("is_refunded = ?", *req.Filters.IsRefunded)
		}
		if len(req.Filters.TicketClassIDs) > 0 {
			query = query.Joins("JOIN tickets ON tickets.id = attendees.ticket_id").
				Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
				Where("order_items.ticket_class_id IN ?", req.Filters.TicketClassIDs)
		}
	}

	// Get attendees
	var attendees []models.Attendee
	if err := query.Find(&attendees).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch attendees")
		return
	}

	if len(attendees) == 0 {
		middleware.WriteJSONError(w, http.StatusBadRequest, "no attendees found matching criteria")
		return
	}

	// Send emails
	totalSent := 0
	totalFailed := 0
	var failedEmails []string

	for _, attendee := range attendees {
		var err error
		if req.HTMLMessage != "" {
			err = notificationService.SendHTMLEmail(
				[]string{attendee.Email},
				req.Subject,
				req.HTMLMessage,
			)
		} else {
			err = notificationService.SendPlainEmail(
				[]string{attendee.Email},
				req.Subject,
				req.Message,
			)
		}

		if err != nil {
			totalFailed++
			failedEmails = append(failedEmails, attendee.Email)
		} else {
			totalSent++
		}
	}

	response := BulkEmailResponse{
		TotalSent:    totalSent,
		TotalFailed:  totalFailed,
		FailedEmails: failedEmails,
		Message:      fmt.Sprintf("Successfully sent %d emails, %d failed", totalSent, totalFailed),
	}

	json.NewEncoder(w).Encode(response)
}

// BulkExportRequest represents a bulk export request
type BulkExportRequest struct {
	EventID     uint              `json:"event_id"`
	AttendeeIDs []uint            `json:"attendee_ids,omitempty"`
	Format      string            `json:"format"` // "csv", "json", "pdf"
	Filters     *BulkEmailFilters `json:"filters,omitempty"`
}

// ExportAttendeesData exports attendee data in bulk
func (h *AttendeeHandler) ExportAttendeesData(w http.ResponseWriter, r *http.Request) {
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

	// Build query for attendees
	query := h.db.Model(&models.Attendee{}).
		Preload("Event").
		Preload("Ticket").
		Preload("Ticket.OrderItem.TicketClass").
		Where("event_id = ?", req.EventID)

	// Apply specific attendee IDs if provided
	if len(req.AttendeeIDs) > 0 {
		query = query.Where("id IN ?", req.AttendeeIDs)
	}

	// Apply filters if provided
	if req.Filters != nil {
		if req.Filters.HasArrived != nil {
			query = query.Where("has_arrived = ?", *req.Filters.HasArrived)
		}
		if req.Filters.IsRefunded != nil {
			query = query.Where("is_refunded = ?", *req.Filters.IsRefunded)
		}
		if len(req.Filters.TicketClassIDs) > 0 {
			query = query.Joins("JOIN tickets ON tickets.id = attendees.ticket_id").
				Joins("JOIN order_items ON order_items.id = tickets.order_item_id").
				Where("order_items.ticket_class_id IN ?", req.Filters.TicketClassIDs)
		}
	}

	// Get attendees
	var attendees []models.Attendee
	if err := query.Find(&attendees).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch attendees")
		return
	}

	// Export based on format
	switch strings.ToLower(req.Format) {
	case "csv":
		h.exportAttendeesCSV(w, attendees, event.Title)
	case "json":
		h.exportAttendeesJSON(w, attendees)
	default:
		middleware.WriteJSONError(w, http.StatusBadRequest, "unsupported format. Use 'csv' or 'json'")
	}
}

// exportAttendeesCSV exports attendees as CSV
func (h *AttendeeHandler) exportAttendeesCSV(w http.ResponseWriter, attendees []models.Attendee, eventTitle string) {
	filename := fmt.Sprintf("attendees_%s_%s.csv",
		strings.ReplaceAll(eventTitle, " ", "_"),
		time.Now().Format("20060102"))

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{
		"ID",
		"First Name",
		"Last Name",
		"Email",
		"Ticket Number",
		"Ticket Class",
		"Has Arrived",
		"Arrival Time",
		"Is Refunded",
		"Reference Number",
		"Created At",
	}
	writer.Write(header)

	// Write data
	for _, attendee := range attendees {
		arrivalTime := ""
		if attendee.ArrivalTime != nil {
			arrivalTime = attendee.ArrivalTime.Format("2006-01-02 15:04:05")
		}

		ticketNumber := ""
		ticketClass := ""
		if attendee.Ticket.ID > 0 {
			ticketNumber = attendee.Ticket.TicketNumber
			if attendee.Ticket.OrderItem.ID > 0 && attendee.Ticket.OrderItem.TicketClass.ID > 0 {
				ticketClass = attendee.Ticket.OrderItem.TicketClass.Name
			}
		}

		row := []string{
			strconv.Itoa(int(attendee.ID)),
			attendee.FirstName,
			attendee.LastName,
			attendee.Email,
			ticketNumber,
			ticketClass,
			strconv.FormatBool(attendee.HasArrived),
			arrivalTime,
			strconv.FormatBool(attendee.IsRefunded),
			strconv.Itoa(attendee.PrivateReferenceNumber),
			attendee.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		writer.Write(row)
	}
}

// exportAttendeesJSON exports attendees as JSON
func (h *AttendeeHandler) exportAttendeesJSON(w http.ResponseWriter, attendees []models.Attendee) {
	w.Header().Set("Content-Type", "application/json")

	responses := make([]AttendeeResponse, len(attendees))
	for i, attendee := range attendees {
		responses[i] = convertToAttendeeResponse(attendee)
	}

	json.NewEncoder(w).Encode(responses)
}

// SendEventUpdateEmail sends an update email to all event attendees
func (h *AttendeeHandler) SendEventUpdateEmail(w http.ResponseWriter, r *http.Request, notificationService *notifications.NotificationService) {
	w.Header().Set("Content-Type", "application/json")

	if notificationService == nil {
		middleware.WriteJSONError(w, http.StatusServiceUnavailable, "notification service not available")
		return
	}

	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
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

	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid event_id")
		return
	}

	// Parse request body
	var req struct {
		Subject        string `json:"subject"`
		Message        string `json:"message"`
		HTMLMessage    string `json:"html_message,omitempty"`
		OnlyNonArrived bool   `json:"only_non_arrived,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Subject == "" || req.Message == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "subject and message are required")
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
	if err := h.db.Where("id = ? AND account_id = ?", uint(eventID), user.AccountID).First(&event).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied or event not found")
		return
	}

	// Get attendees
	query := h.db.Model(&models.Attendee{}).Where("event_id = ? AND is_refunded = ?", uint(eventID), false)

	if req.OnlyNonArrived {
		query = query.Where("has_arrived = ?", false)
	}

	var attendees []models.Attendee
	if err := query.Find(&attendees).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch attendees")
		return
	}

	// Send emails
	totalSent := 0
	totalFailed := 0
	var failedEmails []string

	for _, attendee := range attendees {
		var err error
		if req.HTMLMessage != "" {
			err = notificationService.SendHTMLEmail(
				[]string{attendee.Email},
				req.Subject,
				req.HTMLMessage,
			)
		} else {
			err = notificationService.SendPlainEmail(
				[]string{attendee.Email},
				req.Subject,
				req.Message,
			)
		}

		if err != nil {
			totalFailed++
			failedEmails = append(failedEmails, attendee.Email)
		} else {
			totalSent++
		}
	}

	response := BulkEmailResponse{
		TotalSent:    totalSent,
		TotalFailed:  totalFailed,
		FailedEmails: failedEmails,
		Message:      fmt.Sprintf("Event update sent to %d attendees, %d failed", totalSent, totalFailed),
	}

	json.NewEncoder(w).Encode(response)
}
