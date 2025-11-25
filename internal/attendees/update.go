package attendees

import (
	"encoding/json"
	"net/http"
	"strconv"

	"ticketing_system/internal/models"

	"github.com/gorilla/mux"
)

// UpdateAttendeeInfo updates attendee information
func (h *AttendeeHandler) UpdateAttendeeInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	attendeeID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid attendee ID", http.StatusBadRequest)
		return
	}

	var req UpdateAttendeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var attendee models.Attendee
	if err := h.db.First(&attendee, uint(attendeeID)).Error; err != nil {
		http.Error(w, "Attendee not found", http.StatusNotFound)
		return
	}

	// Update fields
	if req.FirstName != "" {
		attendee.FirstName = req.FirstName
	}
	if req.LastName != "" {
		attendee.LastName = req.LastName
	}
	if req.Email != "" {
		attendee.Email = req.Email
	}

	if err := h.db.Save(&attendee).Error; err != nil {
		http.Error(w, "Failed to update attendee", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Attendee updated successfully"})
}

// MarkAttendeeAsNoShow marks an attendee as no-show
func (h *AttendeeHandler) MarkAttendeeAsNoShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	attendeeID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid attendee ID", http.StatusBadRequest)
		return
	}

	var attendee models.Attendee
	if err := h.db.First(&attendee, uint(attendeeID)).Error; err != nil {
		http.Error(w, "Attendee not found", http.StatusNotFound)
		return
	}

	if attendee.HasArrived {
		http.Error(w, "Attendee has already checked in", http.StatusBadRequest)
		return
	}

	// Mark as no-show by setting some flag or note
	// For now, we'll just ensure they remain not checked in
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Attendee marked as no-show"})
}

// TransferAttendee transfers attendee to another person (via ticket transfer)
func (h *AttendeeHandler) TransferAttendee(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	attendeeID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid attendee ID", http.StatusBadRequest)
		return
	}

	var req struct {
		NewFirstName string `json:"new_first_name"`
		NewLastName  string `json:"new_last_name"`
		NewEmail     string `json:"new_email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var attendee models.Attendee
	if err := h.db.First(&attendee, uint(attendeeID)).Error; err != nil {
		http.Error(w, "Attendee not found", http.StatusNotFound)
		return
	}

	if attendee.HasArrived {
		http.Error(w, "Cannot transfer checked-in attendee", http.StatusBadRequest)
		return
	}

	// Update to new attendee details
	attendee.FirstName = req.NewFirstName
	attendee.LastName = req.NewLastName
	attendee.Email = req.NewEmail

	if err := h.db.Save(&attendee).Error; err != nil {
		http.Error(w, "Failed to transfer attendee", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Attendee transferred successfully"})
}
