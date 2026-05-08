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

	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
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

	var timezones []models.Timezone
	if err := h.db.Where("is_active = ?", true).Order(`"offset", display_name`).Find(&timezones).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch timezones")
		return
	}

	// Format response
	var response []map[string]interface{}
	for _, tz := range timezones {
		response = append(response, map[string]interface{}{
			"id":           tz.ID,
			"name":         tz.Name,
			"display_name": tz.DisplayName,
			"offset":       tz.Offset,
			"iana_name":    tz.IanaName,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"timezones": response,
	})
}

// GetAvailableCurrencies handles getting available currency options
func (h *AccountHandler) GetAvailableCurrencies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var currencies []models.Currency
	if err := h.db.Where("is_active = ?", true).Order("code").Find(&currencies).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch currencies")
		return
	}

	// Format response
	var response []map[string]interface{}
	for _, curr := range currencies {
		response = append(response, map[string]interface{}{
			"id":     curr.ID,
			"code":   curr.Code,
			"name":   curr.Name,
			"symbol": curr.Symbol,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"currencies": response,
	})
}

// GetDateFormats handles getting available date format options
func (h *AccountHandler) GetDateFormats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var dateFormats []models.DateFormat
	if err := h.db.Where("is_active = ?", true).Order("id").Find(&dateFormats).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch date formats")
		return
	}

	// Format response
	var response []map[string]interface{}
	for _, df := range dateFormats {
		response = append(response, map[string]interface{}{
			"id":      df.ID,
			"format":  df.Format,
			"example": df.Example,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"date_formats": response,
	})
}

// GetDateTimeFormats handles getting available datetime format options
func (h *AccountHandler) GetDateTimeFormats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var dateTimeFormats []models.DateTimeFormat
	if err := h.db.Where("is_active = ?", true).Order("id").Find(&dateTimeFormats).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "failed to fetch datetime formats")
		return
	}

	// Format response
	var response []map[string]interface{}
	for _, dtf := range dateTimeFormats {
		response = append(response, map[string]interface{}{
			"id":      dtf.ID,
			"format":  dtf.Format,
			"example": dtf.Example,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"datetime_formats": response,
	})
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

	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
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
		Notifications:  getNotificationSettings(h.db, user.AccountID),
	}

	json.NewEncoder(w).Encode(settings)
}

// UpdateAccountSettings handles updating account settings
func (h *AccountHandler) UpdateAccountSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := middleware.GetUserIDFromTokenWithError(r)
	if err != nil || userID == 0 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
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

func getNotificationSettings(db *gorm.DB, accountID uint) NotificationSettings {
	var prefs models.NotificationPreferences

	// Try to fetch existing preferences
	if err := db.Where("account_id = ?", accountID).First(&prefs).Error; err != nil {
		// If not found, create default preferences
		prefs = models.NotificationPreferences{
			AccountID:            accountID,
			EmailNotifications:   true,
			SMSNotifications:     false,
			PushNotifications:    true,
			EventUpdates:         true,
			PaymentNotifications: true,
			SecurityAlerts:       true,
			MarketingEmails:      false,
		}
		db.Create(&prefs)
	}

	return NotificationSettings{
		EmailNotifications:   prefs.EmailNotifications,
		SMSNotifications:     prefs.SMSNotifications,
		PushNotifications:    prefs.PushNotifications,
		EventUpdates:         prefs.EventUpdates,
		PaymentNotifications: prefs.PaymentNotifications,
		SecurityAlerts:       prefs.SecurityAlerts,
		MarketingEmails:      prefs.MarketingEmails,
	}
}

func updateNotificationSettings(db *gorm.DB, accountID uint, settings NotificationSettings) error {
	var prefs models.NotificationPreferences

	// Check if preferences exist
	if err := db.Where("account_id = ?", accountID).First(&prefs).Error; err != nil {
		// Create new preferences if not found
		prefs = models.NotificationPreferences{
			AccountID:            accountID,
			EmailNotifications:   settings.EmailNotifications,
			SMSNotifications:     settings.SMSNotifications,
			PushNotifications:    settings.PushNotifications,
			EventUpdates:         settings.EventUpdates,
			PaymentNotifications: settings.PaymentNotifications,
			SecurityAlerts:       settings.SecurityAlerts,
			MarketingEmails:      settings.MarketingEmails,
		}
		return db.Create(&prefs).Error
	}

	// Update existing preferences
	prefs.EmailNotifications = settings.EmailNotifications
	prefs.SMSNotifications = settings.SMSNotifications
	prefs.PushNotifications = settings.PushNotifications
	prefs.EventUpdates = settings.EventUpdates
	prefs.PaymentNotifications = settings.PaymentNotifications
	prefs.SecurityAlerts = settings.SecurityAlerts
	prefs.MarketingEmails = settings.MarketingEmails

	return db.Save(&prefs).Error
}
