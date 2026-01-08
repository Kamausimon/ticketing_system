package support

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"ticketing_system/internal/ai"
	"ticketing_system/internal/analytics"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"ticketing_system/internal/notifications"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type SupportHandler struct {
	db            *gorm.DB
	metrics       *analytics.PrometheusMetrics
	notifications *notifications.NotificationService
	aiClassifier  *ai.ClassifierService
}

func NewSupportHandler(db *gorm.DB, metrics *analytics.PrometheusMetrics, notificationService *notifications.NotificationService) *SupportHandler {
	return &SupportHandler{
		db:            db,
		metrics:       metrics,
		notifications: notificationService,
		aiClassifier:  ai.NewClassifierService(),
	}
}

// Request/Response structures
type CreateTicketRequest struct {
	Subject     string `json:"subject"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number,omitempty"`
	OrderID     *uint  `json:"order_id,omitempty"`
	EventID     *uint  `json:"event_id,omitempty"`
}

type UpdateTicketRequest struct {
	Status          *string `json:"status,omitempty"`
	Priority        *string `json:"priority,omitempty"`
	AssignedToID    *uint   `json:"assigned_to_id,omitempty"`
	ResolutionNotes *string `json:"resolution_notes,omitempty"`
}

type AddCommentRequest struct {
	Comment    string `json:"comment"`
	IsInternal bool   `json:"is_internal,omitempty"`
}

type TicketListResponse struct {
	Tickets    []models.SupportTicket `json:"tickets"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	PerPage    int                    `json:"per_page"`
	TotalPages int                    `json:"total_pages"`
}

// CreateTicket handles creating a new support ticket
func (h *SupportHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields
	if err := validateCreateTicketRequest(req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get user ID if authenticated (optional)
	var userID *uint
	if id := middleware.GetUserIDFromToken(r); id != 0 {
		userID = &id
	}

	// Generate unique ticket number
	ticketNumber := generateTicketNumber()

	// Get organizer ID if this is about an event
	var organizerID *uint
	if req.EventID != nil {
		var event models.Event
		if err := h.db.First(&event, *req.EventID).Error; err == nil {
			organizerID = &event.OrganizerID
		}
	}

	// Create support ticket
	ticket := models.SupportTicket{
		TicketNumber: ticketNumber,
		Subject:      strings.TrimSpace(req.Subject),
		Description:  strings.TrimSpace(req.Description),
		Category:     models.SupportTicketCategory(req.Category),
		Priority:     models.TicketPriorityMedium, // Default priority
		Status:       models.TicketStatusOpen,
		UserID:       userID,
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		Name:         strings.TrimSpace(req.Name),
		PhoneNumber:  req.PhoneNumber,
		OrderID:      req.OrderID,
		EventID:      req.EventID,
		OrganizerID:  organizerID,
	}

	if err := h.db.Create(&ticket).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to create support ticket")
		return
	}

	// Run AI classification asynchronously if available
	if h.aiClassifier != nil {
		h.aiClassifier.ClassifyAsync(
			ticket.Subject,
			ticket.Description,
			string(ticket.Category),
			ticket.OrderID,
			ticket.EventID,
			func(classification *ai.TicketClassification, err error) {
				if err != nil {
					fmt.Printf("AI classification failed for ticket #%s: %v\n", ticket.TicketNumber, err)
					return
				}

				// Update ticket with AI classification results
				updates := map[string]interface{}{
					"ai_classified":       true,
					"ai_priority":         classification.Priority,
					"ai_confidence_score": classification.Confidence,
					"ai_reasoning":        classification.Reasoning,
				}

				// Auto-apply priority if confidence is high enough
				if classification.Confidence >= 0.85 {
					updates["priority"] = classification.Priority
					fmt.Printf("Auto-applied AI priority '%s' to ticket #%s (confidence: %.2f)\n",
						classification.Priority, ticket.TicketNumber, classification.Confidence)
				}

				if err := h.db.Model(&models.SupportTicket{}).Where("id = ?", ticket.ID).Updates(updates).Error; err != nil {
					fmt.Printf("Failed to update ticket with AI classification: %v\n", err)
				}
			},
		)
	}

	// Load relationships
	h.db.Preload("User").Preload("Order").Preload("Event").Preload("Organizer").First(&ticket, ticket.ID)

	// Send email notification to support team
	if h.notifications != nil {
		emailData := notifications.SupportTicketCreatedData{
			TicketID:      ticket.ID,
			TicketNumber:  ticket.TicketNumber,
			Subject:       ticket.Subject,
			Description:   ticket.Description,
			Category:      string(ticket.Category),
			Priority:      string(ticket.Priority),
			CustomerName:  ticket.Name,
			CustomerEmail: ticket.Email,
			OrderID:       ticket.OrderID,
			EventID:       ticket.EventID,
			CreatedAt:     ticket.CreatedAt.Format("January 2, 2006 at 3:04 PM"),
			DashboardURL:  "http://localhost:3000", // TODO: Get from config
			AIClassified:  ticket.AIClassified,
			AIPriority:    ticket.AIPriority,
			AIConfidence:  int(ticket.AIConfidenceScore * 100),
			AIReasoning:   ticket.AIReasoning,
		}

		if err := h.notifications.SendSupportTicketCreated(emailData); err != nil {
			fmt.Printf("Failed to send ticket creation email: %v\n", err)
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Support ticket created successfully",
		"ticket":  ticket,
	})
}

// GetTicket retrieves a single support ticket by ID or ticket number
func (h *SupportHandler) GetTicket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	identifier := vars["id"]

	var ticket models.SupportTicket
	query := h.db.Preload("User").Preload("Order").Preload("Event").Preload("Organizer").
		Preload("AssignedTo").Preload("ResolvedBy").Preload("Comments.User")

	// Try to find by ID first, then by ticket number
	var err error
	if id, parseErr := strconv.ParseUint(identifier, 10, 32); parseErr == nil {
		err = query.First(&ticket, uint(id)).Error
	} else {
		err = query.Where("ticket_number = ?", identifier).First(&ticket).Error
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.WriteJSONError(w, http.StatusNotFound, "support ticket not found")
		} else {
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to retrieve support ticket")
		}
		return
	}

	// Check access permissions
	userID := middleware.GetUserIDFromToken(r)
	if !h.canAccessTicket(userID, ticket) {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	json.NewEncoder(w).Encode(ticket)
}

// ListTickets retrieves support tickets with filtering
func (h *SupportHandler) ListTickets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)

	// Get user to check role
	var user models.User
	if userID != 0 {
		h.db.First(&user, userID)
	}

	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	status := r.URL.Query().Get("status")
	priority := r.URL.Query().Get("priority")
	category := r.URL.Query().Get("category")
	search := r.URL.Query().Get("search")

	// Build query
	query := h.db.Model(&models.SupportTicket{}).
		Preload("User").Preload("Order").Preload("Event").Preload("Organizer").
		Preload("AssignedTo")

	// Apply role-based filtering
	switch user.Role {
	case models.RoleAdmin, models.RoleSupport:
		// Admins and support staff can see all tickets
	case models.RoleOrganizer:
		// Organizers see tickets related to their events
		query = query.Joins("LEFT JOIN events ON events.id = support_tickets.event_id").
			Where("events.account_id = ? OR support_tickets.user_id = ?", user.AccountID, userID)
	default:
		// Regular users see only their own tickets
		if userID == 0 {
			middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		query = query.Where("user_id = ?", userID)
	}

	// Apply filters
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if priority != "" {
		query = query.Where("priority = ?", priority)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		query = query.Where(
			"LOWER(ticket_number) LIKE ? OR LOWER(subject) LIKE ? OR LOWER(description) LIKE ? OR LOWER(email) LIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get paginated results
	var tickets []models.SupportTicket
	offset := (page - 1) * perPage
	if err := query.Order("created_at DESC").Offset(offset).Limit(perPage).Find(&tickets).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to retrieve support tickets")
		return
	}

	totalPages := int((total + int64(perPage) - 1) / int64(perPage))

	response := TicketListResponse{
		Tickets:    tickets,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}

	json.NewEncoder(w).Encode(response)
}

// UpdateTicket handles updating a support ticket
func (h *SupportHandler) UpdateTicket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	ticketID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid ticket ID")
		return
	}

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get user to check permissions
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Only admin and support staff can update tickets
	switch user.Role {
	case models.RoleAdmin, models.RoleSupport:
		// Authorized to update tickets
	default:
		middleware.WriteJSONError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	var req UpdateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get existing ticket
	var ticket models.SupportTicket
	if err := h.db.First(&ticket, uint(ticketID)).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "support ticket not found")
		return
	}

	// Update fields
	updates := make(map[string]interface{})

	if req.Status != nil {
		if !isValidStatus(*req.Status) {
			middleware.WriteJSONError(w, http.StatusBadRequest, "invalid status")
			return
		}
		updates["status"] = *req.Status

		// If marking as resolved/closed, set resolved fields
		if *req.Status == string(models.TicketStatusResolved) || *req.Status == string(models.TicketStatusClosed) {
			now := time.Now()
			updates["resolved_at"] = now
			updates["resolved_by_id"] = userID
		}
	}

	if req.Priority != nil {
		if !isValidPriority(*req.Priority) {
			middleware.WriteJSONError(w, http.StatusBadRequest, "invalid priority")
			return
		}
		updates["priority"] = *req.Priority
	}

	if req.AssignedToID != nil {
		updates["assigned_to_id"] = *req.AssignedToID
	}

	if req.ResolutionNotes != nil {
		updates["resolution_notes"] = *req.ResolutionNotes
	}

	if err := h.db.Model(&ticket).Updates(updates).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update support ticket")
		return
	}

	// Reload ticket with relationships
	h.db.Preload("User").Preload("Order").Preload("Event").Preload("Organizer").
		Preload("AssignedTo").Preload("ResolvedBy").First(&ticket, ticket.ID)

	// Send email notification about status change
	if h.notifications != nil && req.Status != nil {
		var assignedToName string
		if ticket.AssignedTo != nil {
			assignedToName = ticket.AssignedTo.FirstName + " " + ticket.AssignedTo.LastName
		}

		var resolvedAt string
		if ticket.ResolvedAt != nil {
			resolvedAt = ticket.ResolvedAt.Format("January 2, 2006 at 3:04 PM")
		}

		emailData := notifications.SupportTicketStatusUpdateData{
			TicketNumber:    ticket.TicketNumber,
			Subject:         ticket.Subject,
			CustomerName:    ticket.Name,
			OldStatus:       string(ticket.Status), // Note: This shows new status due to update
			NewStatus:       *req.Status,
			Priority:        string(ticket.Priority),
			AssignedTo:      assignedToName,
			ResolutionNotes: ticket.ResolutionNotes,
			ResolvedAt:      resolvedAt,
			UpdatedAt:       ticket.UpdatedAt.Format("January 2, 2006 at 3:04 PM"),
			TicketURL:       fmt.Sprintf("http://localhost:3000/support/tickets/%s", ticket.TicketNumber),
		}

		if err := h.notifications.SendTicketStatusUpdate(ticket.Email, emailData); err != nil {
			fmt.Printf("Failed to send ticket status update email: %v\n", err)
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Support ticket updated successfully",
		"ticket":  ticket,
	})
}

// AddComment adds a comment to a support ticket
func (h *SupportHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	ticketID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid ticket ID")
		return
	}

	userID := middleware.GetUserIDFromToken(r)

	var req AddCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Comment) == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "comment cannot be empty")
		return
	}

	// Get ticket
	var ticket models.SupportTicket
	if err := h.db.First(&ticket, uint(ticketID)).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "support ticket not found")
		return
	}

	// Check access permissions
	if !h.canAccessTicket(userID, ticket) {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Get user info
	var authorName, authorEmail string
	var userIDPtr *uint
	if userID != 0 {
		var user models.User
		if err := h.db.First(&user, userID).Error; err == nil {
			authorName = user.FirstName + " " + user.LastName
			authorEmail = user.Email
			userIDPtr = &userID
		}
	} else {
		authorName = ticket.Name
		authorEmail = ticket.Email
	}

	// Create comment
	comment := models.SupportTicketComment{
		TicketID:    uint(ticketID),
		UserID:      userIDPtr,
		Comment:     strings.TrimSpace(req.Comment),
		IsInternal:  req.IsInternal,
		AuthorName:  authorName,
		AuthorEmail: authorEmail,
	}

	if err := h.db.Create(&comment).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to add comment")
		return
	}

	// Reload comment with relationships
	h.db.Preload("User").First(&comment, comment.ID)

	// Send email notification about new comment (only if not internal and ticket has an email)
	if h.notifications != nil && !comment.IsInternal && ticket.Email != "" {
		emailData := notifications.SupportTicketCommentData{
			TicketNumber:  ticket.TicketNumber,
			Subject:       ticket.Subject,
			CustomerName:  ticket.Name,
			Status:        string(ticket.Status),
			CommentAuthor: authorName,
			Comment:       comment.Comment,
			CommentTime:   comment.CreatedAt.Format("January 2, 2006 at 3:04 PM"),
			TicketURL:     fmt.Sprintf("http://localhost:3000/support/tickets/%s", ticket.TicketNumber),
		}

		if err := h.notifications.SendTicketCommentAdded(ticket.Email, emailData); err != nil {
			fmt.Printf("Failed to send comment notification email: %v\n", err)
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Comment added successfully",
		"comment": comment,
	})
}

// GetTicketStats returns statistics about support tickets
func (h *SupportHandler) GetTicketStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get user to check role
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Only admin and support staff can see stats
	switch user.Role {
	case models.RoleAdmin, models.RoleSupport:
		// Authorized to view stats
	default:
		middleware.WriteJSONError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	// Calculate stats
	stats := make(map[string]interface{})

	// Total tickets
	var total int64
	h.db.Model(&models.SupportTicket{}).Count(&total)
	stats["total"] = total

	// By status
	var statusCounts []struct {
		Status string
		Count  int64
	}
	h.db.Model(&models.SupportTicket{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusCounts)

	statusMap := make(map[string]int64)
	for _, sc := range statusCounts {
		statusMap[sc.Status] = sc.Count
	}
	stats["by_status"] = statusMap

	// By priority
	var priorityCounts []struct {
		Priority string
		Count    int64
	}
	h.db.Model(&models.SupportTicket{}).
		Select("priority, COUNT(*) as count").
		Group("priority").
		Scan(&priorityCounts)

	priorityMap := make(map[string]int64)
	for _, pc := range priorityCounts {
		priorityMap[pc.Priority] = pc.Count
	}
	stats["by_priority"] = priorityMap

	// By category
	var categoryCounts []struct {
		Category string
		Count    int64
	}
	h.db.Model(&models.SupportTicket{}).
		Select("category, COUNT(*) as count").
		Group("category").
		Scan(&categoryCounts)

	categoryMap := make(map[string]int64)
	for _, cc := range categoryCounts {
		categoryMap[cc.Category] = cc.Count
	}
	stats["by_category"] = categoryMap

	// Average resolution time (for resolved tickets)
	var avgResolutionHours float64
	h.db.Model(&models.SupportTicket{}).
		Where("resolved_at IS NOT NULL").
		Select("AVG(EXTRACT(EPOCH FROM (resolved_at - created_at)) / 3600) as avg_hours").
		Row().Scan(&avgResolutionHours)
	stats["avg_resolution_hours"] = avgResolutionHours

	json.NewEncoder(w).Encode(stats)
}

// Helper functions

func validateCreateTicketRequest(req CreateTicketRequest) error {
	if strings.TrimSpace(req.Subject) == "" {
		return fmt.Errorf("subject is required")
	}
	if strings.TrimSpace(req.Description) == "" {
		return fmt.Errorf("description is required")
	}
	if strings.TrimSpace(req.Email) == "" {
		return fmt.Errorf("email is required")
	}
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if !isValidCategory(req.Category) {
		return fmt.Errorf("invalid category")
	}
	return nil
}

func isValidCategory(category string) bool {
	validCategories := []string{
		string(models.TicketCategoryPayment),
		string(models.TicketCategoryBooking),
		string(models.TicketCategoryAccount),
		string(models.TicketCategoryEvent),
		string(models.TicketCategoryTechnical),
		string(models.TicketCategoryRefund),
		string(models.TicketCategoryGeneral),
		string(models.TicketCategoryFeatureRequest),
	}
	for _, v := range validCategories {
		if v == category {
			return true
		}
	}
	return false
}

func isValidStatus(status string) bool {
	validStatuses := []string{
		string(models.TicketStatusOpen),
		string(models.TicketStatusInProgress),
		string(models.TicketStatusResolved),
		string(models.TicketStatusClosed),
	}
	for _, v := range validStatuses {
		if v == status {
			return true
		}
	}
	return false
}

func isValidPriority(priority string) bool {
	validPriorities := []string{
		string(models.TicketPriorityCritical),
		string(models.TicketPriorityHigh),
		string(models.TicketPriorityMedium),
		string(models.TicketPriorityLow),
	}
	for _, v := range validPriorities {
		if v == priority {
			return true
		}
	}
	return false
}

func generateTicketNumber() string {
	// Generate ticket number in format: TKT-YYYYMMDD-XXXX
	now := time.Now()
	dateStr := now.Format("20060102")
	randomNum := rand.Intn(10000)
	return fmt.Sprintf("TKT-%s-%04d", dateStr, randomNum)
}

func (h *SupportHandler) canAccessTicket(userID uint, ticket models.SupportTicket) bool {
	if userID == 0 {
		return false
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		return false
	}

	// Check access based on role
	switch user.Role {
	case models.RoleAdmin, models.RoleSupport:
		// Admin and support staff can access all tickets
		return true
	case models.RoleOrganizer:
		// User can access their own tickets
		if ticket.UserID != nil && *ticket.UserID == userID {
			return true
		}
		// Organizers can access tickets related to their events
		if ticket.EventID != nil {
			var event models.Event
			if err := h.db.First(&event, *ticket.EventID).Error; err == nil {
				if event.AccountID == user.AccountID {
					return true
				}
			}
		}
	default:
		// Regular users can only access their own tickets
		if ticket.UserID != nil && *ticket.UserID == userID {
			return true
		}
	}

	return false
}
