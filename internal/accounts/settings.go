package accounts

import (
	"encoding/json"
	"net/http"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"

	"gorm.io/gorm"
)

// GetAccountPreferences handles getting user's account preferences
func (h *AccountHandler) GetAccountPreferences(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get user to access AccountID
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get account
	var account models.Account
	if err := h.db.Where("id = ?", user.AccountID).First(&account).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "account not found")
		return
	}

	// Convert to preferences format
	prefs := AccountPreferences{
		TimezoneID:         account.TimezoneID,
		DateFormatID:       account.DateFormatID,
		DateTimeFormatID:   account.DateTimeFormatID,
		CurrencyID:         account.CurrencyID,
		EmailNotifications: true,  // Default - could be stored in separate preferences table
		SmsNotifications:   false, // Default - could be stored in separate preferences table
	}

	json.NewEncoder(w).Encode(prefs)
}

// UpdateAccountPreferences handles updating user's account preferences
func (h *AccountHandler) UpdateAccountPreferences(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request
	var req AccountPreferences
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get user to access AccountID
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get account
	var account models.Account
	if err := h.db.Where("id = ?", user.AccountID).First(&account).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "account not found")
		return
	}

	// Update preference fields
	account.TimezoneID = req.TimezoneID
	account.DateFormatID = req.DateFormatID
	account.DateTimeFormatID = req.DateTimeFormatID
	account.CurrencyID = req.CurrencyID

	// Save account
	if err := h.db.Save(&account).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update preferences")
		return
	}

	// Log activity
	h.logAccountActivity(account.ID, "preferences_updated", "Account preferences updated", getClientIP(r))

	response := map[string]interface{}{
		"message":     "Preferences updated successfully",
		"preferences": req,
	}

	json.NewEncoder(w).Encode(response)
}

// GetAvailableTimezones handles getting available timezone options
func (h *AccountHandler) GetAvailableTimezones(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Hardcoded timezone list - in production this might come from database
	timezones := []map[string]interface{}{
		{"id": 1, "name": "UTC", "offset": "+00:00"},
		{"id": 2, "name": "EAT (East Africa Time)", "offset": "+03:00"},
		{"id": 3, "name": "EST (Eastern Standard Time)", "offset": "-05:00"},
		{"id": 4, "name": "PST (Pacific Standard Time)", "offset": "-08:00"},
		{"id": 5, "name": "GMT (Greenwich Mean Time)", "offset": "+00:00"},
		{"id": 6, "name": "IST (India Standard Time)", "offset": "+05:30"},
		{"id": 7, "name": "JST (Japan Standard Time)", "offset": "+09:00"},
		{"id": 8, "name": "AEST (Australian Eastern Time)", "offset": "+10:00"},
	}

	response := map[string]interface{}{
		"timezones": timezones,
	}

	json.NewEncoder(w).Encode(response)
}

// GetAvailableCurrencies handles getting available currency options
func (h *AccountHandler) GetAvailableCurrencies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Hardcoded currency list - in production this might come from database
	currencies := []map[string]interface{}{
		{"id": 1, "code": "USD", "name": "US Dollar", "symbol": "$"},
		{"id": 2, "code": "KSH", "name": "Kenyan Shilling", "symbol": "KSh"},
		{"id": 3, "code": "EUR", "name": "Euro", "symbol": "€"},
		{"id": 4, "code": "GBP", "name": "British Pound", "symbol": "£"},
		{"id": 5, "code": "NGN", "name": "Nigerian Naira", "symbol": "₦"},
		{"id": 6, "code": "ZAR", "name": "South African Rand", "symbol": "R"},
		{"id": 7, "code": "INR", "name": "Indian Rupee", "symbol": "₹"},
		{"id": 8, "code": "JPY", "name": "Japanese Yen", "symbol": "¥"},
	}

	response := map[string]interface{}{
		"currencies": currencies,
	}

	json.NewEncoder(w).Encode(response)
}

// GetDateFormats handles getting available date format options
func (h *AccountHandler) GetDateFormats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Hardcoded date format list - in production this might come from database
	dateFormats := []map[string]interface{}{
		{"id": 1, "format": "YYYY-MM-DD", "example": "2024-12-25"},
		{"id": 2, "format": "DD/MM/YYYY", "example": "25/12/2024"},
		{"id": 3, "format": "MM/DD/YYYY", "example": "12/25/2024"},
		{"id": 4, "format": "DD-MM-YYYY", "example": "25-12-2024"},
		{"id": 5, "format": "MMM DD, YYYY", "example": "Dec 25, 2024"},
		{"id": 6, "format": "DD MMM YYYY", "example": "25 Dec 2024"},
	}

	response := map[string]interface{}{
		"date_formats": dateFormats,
	}

	json.NewEncoder(w).Encode(response)
}

// SettingsUpdateRequest represents settings update request
type SettingsUpdateRequest struct {
	TimezoneID       *int                  `json:"timezone_id"`
	DateFormatID     *int                  `json:"date_format_id"`
	DateTimeFormatID *int                  `json:"date_time_format_id"`
	CurrencyID       *int                  `json:"currency_id"`
	Notifications    *NotificationSettings `json:"notifications"`
}

// GetAccountSettings handles getting account settings
func (h *AccountHandler) GetAccountSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get user to find account ID
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Get account
	var account models.Account
	if err := h.db.Where("id = ?", user.AccountID).First(&account).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "account not found")
		return
	}

	// Build settings response
	settings := AccountSettings{
		Timezone:       getTimezoneString(account.TimezoneID),
		DateFormat:     getDateFormatString(account.DateFormatID),
		DateTimeFormat: getDateTimeFormatString(account.DateTimeFormatID),
		Currency:       getCurrencyString(account.CurrencyID),
		Language:       "en", // Default for now
		Notifications:  getNotificationSettings(user.AccountID),
	}

	json.NewEncoder(w).Encode(settings)
}

// UpdateAccountSettings handles updating account settings
func (h *AccountHandler) UpdateAccountSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := middleware.GetUserIDFromToken(r)
	if userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get user to find account ID
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Parse request
	var req SettingsUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get account
	var account models.Account
	if err := h.db.Where("id = ?", user.AccountID).First(&account).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "account not found")
		return
	}

	// Update settings
	if req.TimezoneID != nil {
		account.TimezoneID = req.TimezoneID
	}
	if req.DateFormatID != nil {
		account.DateFormatID = req.DateFormatID
	}
	if req.DateTimeFormatID != nil {
		account.DateTimeFormatID = req.DateTimeFormatID
	}
	if req.CurrencyID != nil {
		account.CurrencyID = req.CurrencyID
	}

	// Save account
	if err := h.db.Save(&account).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update settings")
		return
	}

	// Update notifications if provided
	if req.Notifications != nil {
		if err := updateNotificationSettings(h.db, user.AccountID, *req.Notifications); err != nil {
			middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to update notification settings")
			return
		}
	}

	response := map[string]interface{}{
		"message": "Settings updated successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// Helper functions
func getTimezoneString(timezoneID *int) string {
	if timezoneID == nil {
		return "UTC"
	}
	timezoneMap := map[int]string{
		1: "UTC",
		2: "Africa/Nairobi",
		3: "America/New_York",
		4: "Europe/London",
		5: "Asia/Tokyo",
		6: "Australia/Sydney",
	}
	if tz, exists := timezoneMap[*timezoneID]; exists {
		return tz
	}
	return "UTC"
}

func getDateFormatString(dateFormatID *int) string {
	if dateFormatID == nil {
		return "YYYY-MM-DD"
	}
	formatMap := map[int]string{
		1: "YYYY-MM-DD",
		2: "DD/MM/YYYY",
		3: "MM/DD/YYYY",
		4: "DD-MM-YYYY",
	}
	if format, exists := formatMap[*dateFormatID]; exists {
		return format
	}
	return "YYYY-MM-DD"
}

func getDateTimeFormatString(dateTimeFormatID *int) string {
	if dateTimeFormatID == nil {
		return "YYYY-MM-DD HH:mm"
	}
	formatMap := map[int]string{
		1: "YYYY-MM-DD HH:mm",
		2: "DD/MM/YYYY HH:mm",
		3: "MM/DD/YYYY HH:mm AM/PM",
		4: "DD-MM-YYYY HH:mm",
	}
	if format, exists := formatMap[*dateTimeFormatID]; exists {
		return format
	}
	return "YYYY-MM-DD HH:mm"
}

func getCurrencyString(currencyID *int) string {
	if currencyID == nil {
		return "USD"
	}
	currencyMap := map[int]string{
		1: "USD",
		2: "KSH",
		3: "EUR",
		4: "GBP",
		5: "CAD",
		6: "AUD",
	}
	if currency, exists := currencyMap[*currencyID]; exists {
		return currency
	}
	return "USD"
}

func getNotificationSettings(accountID uint) NotificationSettings {
	// Default notification settings
	// In a real implementation, this would fetch from a notifications table
	return NotificationSettings{
		EmailNotifications:   true,
		SMSNotifications:     false,
		PushNotifications:    true,
		EventUpdates:         true,
		PaymentNotifications: true,
		SecurityAlerts:       true,
		MarketingEmails:      false,
	}
}

func updateNotificationSettings(db *gorm.DB, accountID uint, settings NotificationSettings) error {
	// In a real implementation, this would update a notifications table
	// For now, we'll just return nil (settings are stored in memory/defaults)
	return nil
}
