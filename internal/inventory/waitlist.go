package inventory

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"ticketing_system/internal/notifications"
	"time"

	"github.com/gorilla/mux"
)

// WaitlistRequest represents a request to join the waitlist
type WaitlistRequest struct {
	EventID       uint    `json:"event_id"`
	TicketClassID *uint   `json:"ticket_class_id,omitempty"` // Optional: specific ticket class
	Email         string  `json:"email"`
	Name          string  `json:"name"`
	Phone         *string `json:"phone,omitempty"`
	Quantity      int     `json:"quantity"`
	SessionID     string  `json:"session_id,omitempty"`
}

// WaitlistResponse represents a waitlist entry response
type WaitlistResponse struct {
	ID              uint       `json:"id"`
	EventID         uint       `json:"event_id"`
	EventName       string     `json:"event_name"`
	TicketClassID   *uint      `json:"ticket_class_id,omitempty"`
	TicketClassName *string    `json:"ticket_class_name,omitempty"`
	Email           string     `json:"email"`
	Name            string     `json:"name"`
	Quantity        int        `json:"quantity"`
	Status          string     `json:"status"`
	Position        int        `json:"position"`
	EstimatedWait   string     `json:"estimated_wait"`
	NotifiedAt      *time.Time `json:"notified_at,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// WaitlistStats represents statistics for a waitlist
type WaitlistStats struct {
	EventID            uint       `json:"event_id"`
	TicketClassID      *uint      `json:"ticket_class_id,omitempty"`
	TotalWaiting       int        `json:"total_waiting"`
	TotalNotified      int        `json:"total_notified"`
	TotalConverted     int        `json:"total_converted"`
	ConversionRate     float64    `json:"conversion_rate"`
	AverageWaitTime    string     `json:"average_wait_time"`
	OldestWaitingEntry *time.Time `json:"oldest_waiting_entry,omitempty"`
}

// JoinWaitlist adds a user to the waitlist for sold-out tickets
func (h *InventoryHandler) JoinWaitlist(w http.ResponseWriter, r *http.Request) {
	var req WaitlistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.EventID == 0 {
		writeError(w, http.StatusBadRequest, "Event ID is required")
		return
	}
	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "Email is required")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required")
		return
	}
	if req.Quantity <= 0 || req.Quantity > 10 {
		writeError(w, http.StatusBadRequest, "Quantity must be between 1 and 10")
		return
	}

	// Verify event exists
	var event models.Event
	if err := h.db.First(&event, req.EventID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Event not found")
		return
	}

	// If ticket class specified, verify it exists and is sold out
	if req.TicketClassID != nil {
		var ticketClass models.TicketClass
		if err := h.db.First(&ticketClass, *req.TicketClassID).Error; err != nil {
			writeError(w, http.StatusNotFound, "Ticket class not found")
			return
		}

		// Check if really sold out
		available := h.calculateAvailableQuantity(&ticketClass)
		if available >= req.Quantity {
			writeError(w, http.StatusBadRequest, "Tickets are still available - no need to join waitlist")
			return
		}
	}

	// Check if already on waitlist (case-insensitive email comparison)
	var existingEntry models.WaitlistEntry
	query := h.db.Where("event_id = ? AND LOWER(email) = LOWER(?) AND status = 'waiting'", req.EventID, req.Email)
	if req.TicketClassID != nil {
		query = query.Where("ticket_class_id = ?", *req.TicketClassID)
	}
	if err := query.First(&existingEntry).Error; err == nil {
		// Update existing entry
		existingEntry.Quantity = req.Quantity
		existingEntry.Name = req.Name
		existingEntry.Phone = req.Phone
		if err := h.db.Save(&existingEntry).Error; err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to update waitlist entry")
			return
		}

		response := h.convertToWaitlistResponse(&existingEntry, event.Title, nil)
		writeJSON(w, http.StatusOK, response)
		return
	}

	// Create new waitlist entry
	entry := models.WaitlistEntry{
		EventID:       req.EventID,
		TicketClassID: req.TicketClassID,
		Email:         req.Email,
		Name:          req.Name,
		Phone:         req.Phone,
		Quantity:      req.Quantity,
		Status:        "waiting",
		SessionID:     req.SessionID,
	}

	if err := h.db.Create(&entry).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to join waitlist")
		return
	}

	// Get ticket class name if applicable
	var ticketClassName *string
	if req.TicketClassID != nil {
		var ticketClass models.TicketClass
		if err := h.db.First(&ticketClass, *req.TicketClassID).Error; err == nil {
			ticketClassName = &ticketClass.Name
		}
	}

	response := h.convertToWaitlistResponse(&entry, event.Title, ticketClassName)
	writeJSON(w, http.StatusCreated, response)
}

// GetWaitlistPosition returns the position of a user in the waitlist
func (h *InventoryHandler) GetWaitlistPosition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entryID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid waitlist entry ID")
		return
	}

	var entry models.WaitlistEntry
	if err := h.db.First(&entry, entryID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Waitlist entry not found")
		return
	}

	// Load event and ticket class info
	var event models.Event
	h.db.First(&event, entry.EventID)

	var ticketClassName *string
	if entry.TicketClassID != nil {
		var ticketClass models.TicketClass
		if err := h.db.First(&ticketClass, *entry.TicketClassID).Error; err == nil {
			ticketClassName = &ticketClass.Name
		}
	}

	response := h.convertToWaitlistResponse(&entry, event.Title, ticketClassName)
	writeJSON(w, http.StatusOK, response)
}

// ListUserWaitlist returns all waitlist entries for a user (by email or session)
func (h *InventoryHandler) ListUserWaitlist(w http.ResponseWriter, r *http.Request) {
	// Try to get email from query parameter first
	email := r.URL.Query().Get("email")
	sessionID := r.URL.Query().Get("session_id")

	// If not provided in query params, try to get from authenticated user
	if email == "" && sessionID == "" {
		// Try to get user ID from JWT token (if authenticated)
		userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
		if userID > 0 {
			// Get user's email from database
			var user struct {
				Email string
			}
			if err := h.db.Table("users").Select("email").Where("id = ?", userID).First(&user).Error; err == nil {
				email = user.Email
			}
		}
	}

	// If still no email or session_id, return error
	if email == "" && sessionID == "" {
		writeError(w, http.StatusBadRequest, "Email or session_id is required")
		return
	}

	query := h.db.Where("status IN ?", []string{"waiting", "notified"})
	if email != "" {
		// Case-insensitive email comparison
		query = query.Where("LOWER(email) = LOWER(?)", email)
	} else {
		query = query.Where("session_id = ?", sessionID)
	}

	var entries []models.WaitlistEntry
	if err := query.Order("created_at DESC").Find(&entries).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch waitlist entries")
		return
	}

	var responses []WaitlistResponse
	for _, entry := range entries {
		var event models.Event
		h.db.First(&event, entry.EventID)

		var ticketClassName *string
		if entry.TicketClassID != nil {
			var ticketClass models.TicketClass
			if err := h.db.First(&ticketClass, *entry.TicketClassID).Error; err == nil {
				ticketClassName = &ticketClass.Name
			}
		}

		responses = append(responses, h.convertToWaitlistResponse(&entry, event.Title, ticketClassName))
	}

	result := map[string]interface{}{
		"waitlist_entries": responses,
		"total":            len(responses),
	}

	// Include search criteria in response
	if email != "" {
		result["email"] = email
	}
	if sessionID != "" {
		result["session_id"] = sessionID
	}

	writeJSON(w, http.StatusOK, result)
}

// LeaveWaitlist removes a user from the waitlist
func (h *InventoryHandler) LeaveWaitlist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entryID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid waitlist entry ID")
		return
	}

	var entry models.WaitlistEntry
	if err := h.db.First(&entry, entryID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Waitlist entry not found")
		return
	}

	if entry.Status == "converted" {
		writeError(w, http.StatusBadRequest, "Cannot leave waitlist - already converted to purchase")
		return
	}

	// Update status instead of deleting
	entry.Status = "expired"
	if err := h.db.Save(&entry).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to leave waitlist")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Successfully removed from waitlist",
		"id":      entry.ID,
	})
}

// GetWaitlistStats returns statistics for an event's waitlist
func (h *InventoryHandler) GetWaitlistStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}

	var event models.Event
	if err := h.db.First(&event, eventID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Event not found")
		return
	}

	// Get ticket class ID if specified
	ticketClassIDStr := r.URL.Query().Get("ticket_class_id")
	var ticketClassID *uint
	if ticketClassIDStr != "" {
		tcID, err := strconv.ParseUint(ticketClassIDStr, 10, 64)
		if err == nil {
			tcIDUint := uint(tcID)
			ticketClassID = &tcIDUint
		}
	}

	stats := h.calculateWaitlistStats(uint(eventID), ticketClassID)
	writeJSON(w, http.StatusOK, stats)
}

// NotifyNextInWaitlist notifies the next person in the waitlist when tickets become available
func (h *InventoryHandler) NotifyNextInWaitlist(w http.ResponseWriter, r *http.Request) {
	var req struct {
		EventID       uint  `json:"event_id"`
		TicketClassID *uint `json:"ticket_class_id,omitempty"`
		AvailableQty  int   `json:"available_quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.EventID == 0 || req.AvailableQty <= 0 {
		writeError(w, http.StatusBadRequest, "Invalid event ID or available quantity")
		return
	}

	// Find waiting entries that can be notified
	query := h.db.Where("event_id = ? AND status = 'waiting' AND quantity <= ?", req.EventID, req.AvailableQty).
		Order("priority DESC, created_at ASC")

	if req.TicketClassID != nil {
		query = query.Where("ticket_class_id = ? OR ticket_class_id IS NULL", *req.TicketClassID)
	}

	var entries []models.WaitlistEntry
	if err := query.Limit(10).Find(&entries).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch waitlist entries")
		return
	}

	notifiedCount := 0
	notifiedIDs := []uint{}

	for _, entry := range entries {
		// Update entry status
		now := time.Now()
		expires := now.Add(24 * time.Hour) // 24 hours to complete purchase
		entry.Status = "notified"
		entry.NotifiedAt = &now
		entry.ExpiresAt = &expires

		if err := h.db.Save(&entry).Error; err != nil {
			continue
		}

		// Load event details for email
		var event models.Event
		if err := h.db.First(&event, entry.EventID).Error; err != nil {
			continue
		}

		// Load ticket class details if applicable
		var ticketClassName string
		var price float64
		if entry.TicketClassID != nil {
			var ticketClass models.TicketClass
			if err := h.db.First(&ticketClass, *entry.TicketClassID).Error; err == nil {
				ticketClassName = ticketClass.Name
				price = float64(ticketClass.Price) / 100.0
			}
		}

		// Get venue name from event's venues
		venueName := ""
		if len(event.Venue) > 0 {
			// Load venues if not already loaded
			if event.Venue[0].VenueName == "" {
				h.db.Model(&event).Association("Venue").Find(&event.Venue)
			}
			if len(event.Venue) > 0 {
				venueName = event.Venue[0].VenueName
			}
		}

		// Send notification email
		if h.notificationService != nil {
			emailData := notifications.WaitlistNotificationData{
				Name:            entry.Name,
				EventName:       event.Title,
				EventDate:       event.StartDate.Format("Monday, January 2, 2006 at 3:04 PM"),
				VenueName:       venueName,
				TicketClassName: ticketClassName,
				Quantity:        entry.Quantity,
				Price:           price,
				Currency:        "KES",
				ExpiresAt:       expires.Format("Monday, January 2, 2006 at 3:04 PM"),
				PurchaseURL:     fmt.Sprintf("%s/events/%d", h.baseURL, event.ID),
			}

			if err := h.notificationService.SendWaitlistNotificationEmail(entry.Email, emailData); err != nil {
				// Log error but don't fail the operation
				fmt.Printf("Warning: Failed to send waitlist notification to %s: %v\n", entry.Email, err)
			}
		}

		notifiedCount++
		notifiedIDs = append(notifiedIDs, entry.ID)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"notified_count": notifiedCount,
		"notified_ids":   notifiedIDs,
		"message":        fmt.Sprintf("Notified %d users from the waitlist", notifiedCount),
	})
}

// autoNotifyWaitlist automatically notifies waitlist when tickets become available
// This is called internally after reservations are released or tickets become available
func (h *InventoryHandler) autoNotifyWaitlist(eventID uint, ticketClassID *uint, availableQty int) {
	if availableQty <= 0 {
		return
	}

	// Find waiting entries that can be notified
	query := h.db.Where("event_id = ? AND status = 'waiting' AND quantity <= ?", eventID, availableQty).
		Order("priority DESC, created_at ASC")

	if ticketClassID != nil {
		query = query.Where("ticket_class_id = ? OR ticket_class_id IS NULL", *ticketClassID)
	}

	var entries []models.WaitlistEntry
	if err := query.Limit(10).Find(&entries).Error; err != nil {
		log.Printf("⚠️  Error fetching waitlist entries: %v", err)
		return
	}

	if len(entries) == 0 {
		return
	}

	notifiedCount := 0

	for _, entry := range entries {
		// Update entry status
		now := time.Now()
		expires := now.Add(24 * time.Hour)
		entry.Status = "notified"
		entry.NotifiedAt = &now
		entry.ExpiresAt = &expires

		if err := h.db.Save(&entry).Error; err != nil {
			continue
		}

		// Load event details
		var event models.Event
		if err := h.db.First(&event, entry.EventID).Error; err != nil {
			continue
		}

		// Load ticket class details if applicable
		var ticketClassName string
		var price float64
		if entry.TicketClassID != nil {
			var ticketClass models.TicketClass
			if err := h.db.First(&ticketClass, *entry.TicketClassID).Error; err == nil {
				ticketClassName = ticketClass.Name
				price = float64(ticketClass.Price) / 100.0
			}
		}

		// Get venue name
		venueName := ""
		if len(event.Venue) > 0 {
			if event.Venue[0].VenueName == "" {
				h.db.Model(&event).Association("Venue").Find(&event.Venue)
			}
			if len(event.Venue) > 0 {
				venueName = event.Venue[0].VenueName
			}
		}

		// Send notification
		if h.notificationService != nil {
			emailData := notifications.WaitlistNotificationData{
				Name:            entry.Name,
				EventName:       event.Title,
				EventDate:       event.StartDate.Format("Monday, January 2, 2006 at 3:04 PM"),
				VenueName:       venueName,
				TicketClassName: ticketClassName,
				Quantity:        entry.Quantity,
				Price:           price,
				Currency:        "KES",
				ExpiresAt:       expires.Format("Monday, January 2, 2006 at 3:04 PM"),
				PurchaseURL:     fmt.Sprintf("%s/events/%d", h.baseURL, event.ID),
			}

			if err := h.notificationService.SendWaitlistNotificationEmail(entry.Email, emailData); err == nil {
				notifiedCount++
			}
		}
	}

	if notifiedCount > 0 {
		log.Printf("📧 Automatically notified %d users from waitlist for event %d", notifiedCount, eventID)
	}
}

// Helper: Convert waitlist entry to response
func (h *InventoryHandler) convertToWaitlistResponse(entry *models.WaitlistEntry, eventName string, ticketClassName *string) WaitlistResponse {
	// Calculate position in queue
	var position int64
	query := h.db.Model(&models.WaitlistEntry{}).
		Where("event_id = ? AND status = 'waiting' AND created_at <= ?", entry.EventID, entry.CreatedAt)

	if entry.TicketClassID != nil {
		query = query.Where("ticket_class_id = ? OR ticket_class_id IS NULL", *entry.TicketClassID)
	}

	query.Count(&position)

	// Estimate wait time
	estimatedWait := "Unknown"
	if position > 0 {
		if position <= 10 {
			estimatedWait = "Within 24 hours"
		} else if position <= 50 {
			estimatedWait = "1-3 days"
		} else {
			estimatedWait = "3-7 days"
		}
	}

	return WaitlistResponse{
		ID:              entry.ID,
		EventID:         entry.EventID,
		EventName:       eventName,
		TicketClassID:   entry.TicketClassID,
		TicketClassName: ticketClassName,
		Email:           entry.Email,
		Name:            entry.Name,
		Quantity:        entry.Quantity,
		Status:          entry.Status,
		Position:        int(position),
		EstimatedWait:   estimatedWait,
		NotifiedAt:      entry.NotifiedAt,
		ExpiresAt:       entry.ExpiresAt,
		CreatedAt:       entry.CreatedAt,
	}
}

// Helper: Calculate waitlist statistics
func (h *InventoryHandler) calculateWaitlistStats(eventID uint, ticketClassID *uint) WaitlistStats {
	query := h.db.Model(&models.WaitlistEntry{}).Where("event_id = ?", eventID)
	if ticketClassID != nil {
		query = query.Where("ticket_class_id = ?", *ticketClassID)
	}

	var totalWaiting int64
	query.Where("status = 'waiting'").Count(&totalWaiting)

	var totalNotified int64
	query.Where("status = 'notified'").Count(&totalNotified)

	var totalConverted int64
	query.Where("status = 'converted'").Count(&totalConverted)

	// Calculate conversion rate
	conversionRate := 0.0
	if totalNotified > 0 {
		conversionRate = float64(totalConverted) / float64(totalNotified) * 100
	}

	// Get oldest waiting entry
	var oldestEntry models.WaitlistEntry
	var oldestTime *time.Time
	if err := query.Where("status = 'waiting'").Order("created_at ASC").First(&oldestEntry).Error; err == nil {
		oldestTime = &oldestEntry.CreatedAt
	}

	// Calculate average wait time
	var avgWaitSeconds float64
	h.db.Model(&models.WaitlistEntry{}).
		Select("AVG(EXTRACT(EPOCH FROM (converted_at - created_at)))").
		Where("event_id = ? AND status = 'converted' AND converted_at IS NOT NULL", eventID).
		Scan(&avgWaitSeconds)

	avgWaitTime := "N/A"
	if avgWaitSeconds > 0 {
		hours := int(avgWaitSeconds / 3600)
		avgWaitTime = fmt.Sprintf("%d hours", hours)
	}

	return WaitlistStats{
		EventID:            eventID,
		TicketClassID:      ticketClassID,
		TotalWaiting:       int(totalWaiting),
		TotalNotified:      int(totalNotified),
		TotalConverted:     int(totalConverted),
		ConversionRate:     conversionRate,
		AverageWaitTime:    avgWaitTime,
		OldestWaitingEntry: oldestTime,
	}
}
