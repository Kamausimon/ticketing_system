package events

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"time"

	"github.com/gorilla/mux"
)

type UploadResponse struct {
	Message   string `json:"message"`
	ImagePath string `json:"image_path"`
	ImageID   uint   `json:"image_id"`
}

// UploadEventImage handles event image uploads
func (h *EventHandler) UploadEventImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	fmt.Printf("user id %d", userID)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	vars := mux.Vars(r)
	eventIDStr := vars["id"]

	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid event ID")
		return
	}

	// Get user and verify organizer status
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Ensure user has an account ID
	fmt.Printf("account id %d", user.AccountID)
	if user.AccountID == 0 {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "user account not properly configured")
		return
	}

	if user.Role != models.RoleOrganizer {
		middleware.WriteJSONError(w, http.StatusForbidden, "only organizers can upload event images")
		return
	}

	// Get organizer
	var organizer models.Organizer
	if err := h.db.Where("account_id = ?", user.AccountID).First(&organizer).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "organizer profile not found")
		return
	}

	// Verify event ownership
	var event models.Event
	if err := h.db.Where("id = ? AND organizer_id = ?", eventID, organizer.ID).First(&event).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "event not found or access denied")
		return
	}

	// Parse multipart form
	err = r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "failed to parse form data")
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "image file is required")
		return
	}
	defer file.Close()

	// Validate file type
	if !isValidImageType(handler.Filename) {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid image type. Only PNG, JPEG, JPG, and GIF are allowed")
		return
	}

	// Validate file size
	if handler.Size > 5*1024*1024 { // 5MB limit
		middleware.WriteJSONError(w, http.StatusBadRequest, "image size must be less than 5MB")
		return
	}

	// Create uploads directory if it doesn't exist
	uploadsDir := "uploads/events"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to create upload directory")
		return
	}

	// Generate unique filename
	ext := filepath.Ext(handler.Filename)
	filename := fmt.Sprintf("event_%d_%d%s", eventID, time.Now().Unix(), ext)
	filepath := filepath.Join(uploadsDir, filename)

	// Create the file on disk
	dst, err := os.Create(filepath)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to create file")
		return
	}
	defer dst.Close()

	// Copy file content
	_, err = io.Copy(dst, file)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to save file")
		return
	}

	// Save image record to database
	eventImage := models.EventImages{
		EventID:   uint(eventID),
		ImagePath: filepath,
		AccountID: user.AccountID,
		UserID:    userID,
	}

	if err := h.db.Create(&eventImage).Error; err != nil {
		// If database save fails, remove the uploaded file
		os.Remove(filepath)
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to save image record")
		return
	}

	response := UploadResponse{
		Message:   "Image uploaded successfully",
		ImagePath: filepath,
		ImageID:   eventImage.ID,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// DeleteEventImage handles event image deletion
func (h *EventHandler) DeleteEventImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	vars := mux.Vars(r)
	eventIDStr := vars["id"]
	imageIDStr := vars["imageId"]

	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid event ID")
		return
	}

	imageID, err := strconv.ParseUint(imageIDStr, 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid image ID")
		return
	}

	// Get user and verify organizer status
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	if user.Role != models.RoleOrganizer {
		middleware.WriteJSONError(w, http.StatusForbidden, "only organizers can delete event images")
		return
	}

	// Get organizer
	var organizer models.Organizer
	if err := h.db.Where("account_id = ?", user.AccountID).First(&organizer).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "organizer profile not found")
		return
	}

	// Verify event ownership
	var event models.Event
	if err := h.db.Where("id = ? AND organizer_id = ?", eventID, organizer.ID).First(&event).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "event not found or access denied")
		return
	}

	// Get image record
	var eventImage models.EventImages
	if err := h.db.Where("id = ? AND event_id = ?", imageID, eventID).First(&eventImage).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "image not found")
		return
	}

	// Delete file from disk
	if err := os.Remove(eventImage.ImagePath); err != nil {
		// Log the error but continue with database deletion
		fmt.Printf("Failed to delete image file: %v\n", err)
	}

	// Delete database record
	if err := h.db.Delete(&eventImage).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to delete image record")
		return
	}

	response := map[string]interface{}{
		"message":  "Image deleted successfully",
		"image_id": imageID,
	}

	json.NewEncoder(w).Encode(response)
}

// GetEventImages handles getting all images for an event
func (h *EventHandler) GetEventImages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	eventIDStr := vars["id"]

	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid event ID")
		return
	}

	// Verify event exists
	var event models.Event
	if err := h.db.Where("id = ?", eventID).First(&event).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "event not found")
		return
	}

	// Get all images for the event
	var images []models.EventImages
	if err := h.db.Where("event_id = ?", eventID).Find(&images).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch images")
		return
	}

	// Convert to response format
	imageResponses := make([]EventImageResponse, len(images))
	for i, img := range images {
		imageResponses[i] = EventImageResponse{
			ID:        img.ID,
			ImagePath: img.ImagePath,
		}
	}

	response := map[string]interface{}{
		"images": imageResponses,
		"count":  len(imageResponses),
	}

	json.NewEncoder(w).Encode(response)
}

// Helper function to validate image file types
func isValidImageType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validTypes := []string{".png", ".jpg", ".jpeg", ".gif"}

	for _, validType := range validTypes {
		if ext == validType {
			return true
		}
	}
	return false
}
