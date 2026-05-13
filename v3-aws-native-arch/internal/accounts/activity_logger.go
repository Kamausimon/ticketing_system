package accounts

import (
	"ticketing_system/internal/models"
	"time"

	"gorm.io/gorm"
)

// ActivityLogger provides methods for logging activities throughout the system
type ActivityLogger struct {
	db *gorm.DB
}

// NewActivityLogger creates a new activity logger
func NewActivityLogger(db *gorm.DB) *ActivityLogger {
	return &ActivityLogger{db: db}
}

// LogLogin logs a successful login
func (l *ActivityLogger) LogLogin(accountID uint, userID *uint, ipAddress, userAgent string) {
	l.logActivity(accountID, userID, models.ActionLogin, models.ActivityCategoryAuth,
		"User logged in successfully", ipAddress, userAgent, true, models.SeverityInfo, nil, nil)

	// Also create login history record
	l.createLoginHistory(accountID, userID, ipAddress, userAgent, "", "", "", "", true)
}

// LogLoginFailed logs a failed login attempt
func (l *ActivityLogger) LogLoginFailed(accountID uint, ipAddress, userAgent, reason string) {
	l.logActivity(accountID, nil, models.ActionLoginFailed, models.ActivityCategoryAuth,
		"Failed login attempt: "+reason, ipAddress, userAgent, false, models.SeverityWarning, nil, nil)

	// Also create login history record
	l.createLoginHistory(accountID, nil, ipAddress, userAgent, "", "", "", reason, false)
}

// LogLogout logs a logout
func (l *ActivityLogger) LogLogout(accountID uint, userID *uint, ipAddress, userAgent string) {
	l.logActivity(accountID, userID, models.ActionLogout, models.ActivityCategoryAuth,
		"User logged out", ipAddress, userAgent, true, models.SeverityInfo, nil, nil)
}

// Log2FAEnabled logs when 2FA is enabled
func (l *ActivityLogger) Log2FAEnabled(accountID uint, userID *uint, ipAddress, userAgent string) {
	l.logActivity(accountID, userID, models.Action2FAEnabled, models.ActivityCategorySecurity,
		"Two-factor authentication enabled", ipAddress, userAgent, true, models.SeverityInfo, nil, nil)
}

// Log2FADisabled logs when 2FA is disabled
func (l *ActivityLogger) Log2FADisabled(accountID uint, userID *uint, ipAddress, userAgent string) {
	l.logActivity(accountID, userID, models.Action2FADisabled, models.ActivityCategorySecurity,
		"Two-factor authentication disabled", ipAddress, userAgent, true, models.SeverityWarning, nil, nil)
}

// Log2FAVerified logs successful 2FA verification
func (l *ActivityLogger) Log2FAVerified(accountID uint, userID *uint, ipAddress, userAgent string) {
	l.logActivity(accountID, userID, models.Action2FAVerified, models.ActivityCategorySecurity,
		"Two-factor authentication verified", ipAddress, userAgent, true, models.SeverityInfo, nil, nil)
}

// Log2FAFailed logs failed 2FA attempt
func (l *ActivityLogger) Log2FAFailed(accountID uint, userID *uint, ipAddress, userAgent string) {
	l.logActivity(accountID, userID, models.Action2FAFailed, models.ActivityCategorySecurity,
		"Two-factor authentication failed", ipAddress, userAgent, false, models.SeverityWarning, nil, nil)
}

// LogPasswordChanged logs password change
func (l *ActivityLogger) LogPasswordChanged(accountID uint, userID *uint, ipAddress, userAgent string) {
	l.logActivity(accountID, userID, models.ActionPasswordChanged, models.ActivityCategorySecurity,
		"Password changed successfully", ipAddress, userAgent, true, models.SeverityInfo, nil, nil)
}

// LogPasswordResetRequest logs password reset request
func (l *ActivityLogger) LogPasswordResetRequest(accountID uint, userID *uint, ipAddress, userAgent string) {
	l.logActivity(accountID, userID, models.ActionPasswordResetRequest, models.ActivityCategorySecurity,
		"Password reset requested", ipAddress, userAgent, true, models.SeverityInfo, nil, nil)
}

// LogPasswordReset logs successful password reset
func (l *ActivityLogger) LogPasswordReset(accountID uint, userID *uint, ipAddress, userAgent string) {
	l.logActivity(accountID, userID, models.ActionPasswordReset, models.ActivityCategorySecurity,
		"Password reset completed", ipAddress, userAgent, true, models.SeverityInfo, nil, nil)
}

// LogRegistration logs user registration
func (l *ActivityLogger) LogRegistration(accountID uint, userID *uint, ipAddress, userAgent string) {
	l.logActivity(accountID, userID, models.ActionAccountCreated, models.ActivityCategoryAuth,
		"User account registered", ipAddress, userAgent, true, models.SeverityInfo, nil, nil)
}

// LogRecoveryCodesRegenerated logs when recovery codes are regenerated
func (l *ActivityLogger) LogRecoveryCodesRegenerated(accountID uint, userID *uint, ipAddress, userAgent string) {
	l.logActivity(accountID, userID, models.ActionRecoveryCodesRegenerated, models.ActivityCategorySecurity,
		"Two-factor recovery codes regenerated", ipAddress, userAgent, true, models.SeverityInfo, nil, nil)
}

// LogProfileUpdated logs profile update
func (l *ActivityLogger) LogProfileUpdated(accountID uint, userID *uint, ipAddress, userAgent string) {
	l.logActivity(accountID, userID, models.ActionProfileUpdated, models.ActivityCategoryProfile,
		"Profile information updated", ipAddress, userAgent, true, models.SeverityInfo, nil, nil)
}

// LogEmailVerified logs email verification
func (l *ActivityLogger) LogEmailVerified(accountID uint, userID *uint, ipAddress, userAgent string) {
	l.logActivity(accountID, userID, models.ActionEmailVerified, models.ActivityCategorySecurity,
		"Email address verified", ipAddress, userAgent, true, models.SeverityInfo, nil, nil)
}

// LogEventCreated logs event creation
func (l *ActivityLogger) LogEventCreated(accountID uint, userID *uint, eventID uint, eventName, ipAddress, userAgent string) {
	l.logActivity(accountID, userID, models.ActionEventCreated, models.ActivityCategoryEvent,
		"Event created: "+eventName, ipAddress, userAgent, true, models.SeverityInfo, stringPtr("event"), &eventID)
}

// LogEventPublished logs event publication
func (l *ActivityLogger) LogEventPublished(accountID uint, userID *uint, eventID uint, eventName, ipAddress, userAgent string) {
	l.logActivity(accountID, userID, models.ActionEventPublished, models.ActivityCategoryEvent,
		"Event published: "+eventName, ipAddress, userAgent, true, models.SeverityInfo, stringPtr("event"), &eventID)
}

// LogOrderPlaced logs order placement
func (l *ActivityLogger) LogOrderPlaced(accountID uint, userID *uint, orderID uint, amount float64, ipAddress, userAgent string) {
	l.logActivity(accountID, userID, models.ActionOrderPlaced, models.ActivityCategoryOrder,
		"Order placed", ipAddress, userAgent, true, models.SeverityInfo, stringPtr("order"), &orderID)
}

// LogPaymentProcessed logs successful payment
func (l *ActivityLogger) LogPaymentProcessed(accountID uint, userID *uint, orderID uint, amount float64, ipAddress, userAgent string) {
	l.logActivity(accountID, userID, models.ActionPaymentProcessed, models.ActivityCategoryPayment,
		"Payment processed successfully", ipAddress, userAgent, true, models.SeverityInfo, stringPtr("order"), &orderID)
}

// LogPaymentFailed logs failed payment
func (l *ActivityLogger) LogPaymentFailed(accountID uint, userID *uint, orderID uint, reason, ipAddress, userAgent string) {
	l.logActivity(accountID, userID, models.ActionPaymentFailed, models.ActivityCategoryPayment,
		"Payment failed: "+reason, ipAddress, userAgent, false, models.SeverityWarning, stringPtr("order"), &orderID)
}

// LogRefundRequested logs refund request
func (l *ActivityLogger) LogRefundRequested(accountID uint, userID *uint, orderID uint, amount float64, ipAddress, userAgent string) {
	l.logActivity(accountID, userID, models.ActionRefundRequested, models.ActivityCategoryRefund,
		"Refund requested", ipAddress, userAgent, true, models.SeverityInfo, stringPtr("order"), &orderID)
}

// LogSecurityAlert logs a security alert
func (l *ActivityLogger) LogSecurityAlert(accountID uint, userID *uint, description, ipAddress, userAgent string) {
	l.logActivity(accountID, userID, models.ActionSecurityAlert, models.ActivityCategorySecurity,
		description, ipAddress, userAgent, false, models.SeverityCritical, nil, nil)
}

// logActivity is the internal method that creates the activity record
func (l *ActivityLogger) logActivity(accountID uint, userID *uint, action, category, description, ipAddress, userAgent string,
	success bool, severity string, resource *string, resourceID *uint) {

	activity := models.AccountActivity{
		AccountID:   accountID,
		UserID:      userID,
		Action:      action,
		Category:    category,
		Description: description,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Success:     success,
		Metadata:    nil, // Explicitly set to nil for NULL in database
		Severity:    severity,
		Timestamp:   time.Now(),
	}

	if resource != nil {
		activity.Resource = *resource
	}
	if resourceID != nil {
		activity.ResourceID = resourceID
	}

	l.db.Create(&activity)
}

// createLoginHistory creates a login history record
func (l *ActivityLogger) createLoginHistory(accountID uint, userID *uint, ipAddress, userAgent, device,
	browser, location, failReason string, success bool) {

	history := models.LoginHistory{
		AccountID: accountID,
		UserID:    userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   success,
		LoginAt:   time.Now(),
	}

	if device != "" {
		history.Device = &device
	}
	if browser != "" {
		history.Browser = &browser
	}
	if location != "" {
		history.Location = &location
	}
	if failReason != "" {
		history.FailReason = &failReason
	}

	l.db.Create(&history)
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
