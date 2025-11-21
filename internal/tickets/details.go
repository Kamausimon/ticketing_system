package tickets

import (
	"encoding/json"
	"net/http"
	"strconv"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"

	"github.com/gorilla/mux"
)

// GetTicketDetails handles getting detailed information about a specific ticket
func (h *TicketHandler) GetTicketDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get ticket ID from URL
	vars := mux.Vars(r)
	ticketID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid ticket ID")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get ticket with all related data
	var ticket models.Ticket
	if err := h.db.Preload("OrderItem.Order").
		Preload("OrderItem.TicketClass.Event").
		Where("id = ?", ticketID).First(&ticket).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "ticket not found")
		return
	}

	// Verify ownership
	if ticket.OrderItem.Order.AccountID != user.AccountID {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	response := convertToTicketResponse(ticket)
	json.NewEncoder(w).Encode(response)
}

// GetTicketByNumber handles getting a ticket by its ticket number
func (h *TicketHandler) GetTicketByNumber(w http.ResponseWriter, r *http.Request) {
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

	// Get ticket with all related data
	var ticket models.Ticket
	if err := h.db.Preload("OrderItem.Order").
		Preload("OrderItem.TicketClass.Event").
		Where("ticket_number = ?", ticketNumber).First(&ticket).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "ticket not found")
		return
	}

	// Verify ownership
	if ticket.OrderItem.Order.AccountID != user.AccountID {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	response := convertToTicketResponse(ticket)
	json.NewEncoder(w).Encode(response)
}

// DownloadTicketPDF handles downloading the ticket as a PDF
func (h *TicketHandler) DownloadTicketPDF(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get ticket ID from URL
	vars := mux.Vars(r)
	ticketID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid ticket ID")
		return
	}

	// Get user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get ticket
	var ticket models.Ticket
	if err := h.db.Preload("OrderItem.Order").
		Preload("OrderItem.TicketClass.Event").
		Where("id = ?", ticketID).First(&ticket).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "ticket not found")
		return
	}

	// Verify ownership
	if ticket.OrderItem.Order.AccountID != user.AccountID {
		middleware.WriteJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Generate PDF if not already generated
	if ticket.PdfPath == nil || *ticket.PdfPath == "" {
		pdfPath, err := h.generateTicketPDF(&ticket)
		if err != nil {
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to generate PDF")
			return
		}
		ticket.PdfPath = &pdfPath
		h.db.Save(&ticket)
	}

	// In production, this would stream the actual PDF file
	// For now, return the PDF path
	response := map[string]interface{}{
		"message":  "PDF ready for download",
		"pdf_path": *ticket.PdfPath,
		"pdf_url":  "/api/files/tickets/" + *ticket.PdfPath,
	}

	json.NewEncoder(w).Encode(response)
}
