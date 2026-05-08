package models

import (
	"time"

	"gorm.io/gorm"
)

// AccountActivity represents a logged activity for an account
type AccountActivity struct {
	gorm.Model
	AccountID   uint      `gorm:"not null;index:idx_activity_account_time"`
	Account     Account   `gorm:"foreignKey:AccountID"`
	UserID      *uint     `gorm:"index"` // Optional - which user performed the action
	User        *User     `gorm:"foreignKey:UserID"`
	Action      string    `gorm:"type:varchar(100);not null;index:idx_activity_action"`
	Category    string    `gorm:"type:varchar(50);index"` // auth, profile, event, order, payment, etc.
	Description string    `gorm:"type:text;not null"`
	IPAddress   string    `gorm:"type:varchar(45)"` // IPv4 or IPv6
	UserAgent   string    `gorm:"type:text"`
	Success     bool      `gorm:"default:true;index"`
	Metadata    *string   `gorm:"type:jsonb"`                      // Additional context as JSON
	Severity    string    `gorm:"type:varchar(20);default:'info'"` // info, warning, error, critical
	Resource    string    `gorm:"type:varchar(100)"`               // What was affected (event_id, order_id, etc.)
	ResourceID  *uint     // ID of the affected resource
	Timestamp   time.Time `gorm:"not null;index:idx_activity_account_time;index:idx_activity_timestamp"`
}

// LoginHistory represents a login attempt
type LoginHistory struct {
	gorm.Model
	AccountID       uint      `gorm:"not null;index"`
	Account         Account   `gorm:"foreignKey:AccountID"`
	UserID          *uint     `gorm:"index"`
	User            *User     `gorm:"foreignKey:UserID"`
	IPAddress       string    `gorm:"type:varchar(45);not null;index"`
	UserAgent       string    `gorm:"type:text"`
	Location        *string   `gorm:"type:varchar(255)"` // Geographic location (optional)
	Device          *string   `gorm:"type:varchar(100)"` // Device type
	Browser         *string   `gorm:"type:varchar(100)"` // Browser info
	Success         bool      `gorm:"not null;index"`
	FailReason      *string   `gorm:"type:varchar(255)"` // Reason for failure
	LoginAt         time.Time `gorm:"not null;index"`
	LogoutAt        *time.Time
	SessionDuration *int // Session duration in seconds
}

// TableName overrides the table name
func (AccountActivity) TableName() string {
	return "account_activities"
}

func (LoginHistory) TableName() string {
	return "login_history"
}

// Activity categories
const (
	ActivityCategoryAuth     = "auth"
	ActivityCategoryProfile  = "profile"
	ActivityCategoryEvent    = "event"
	ActivityCategoryOrder    = "order"
	ActivityCategoryPayment  = "payment"
	ActivityCategorySecurity = "security"
	ActivityCategoryTicket   = "ticket"
	ActivityCategoryRefund   = "refund"
	ActivityCategorySettings = "settings"
	ActivityCategoryAdmin    = "admin"
)

// Activity severity levels
const (
	SeverityInfo     = "info"
	SeverityWarning  = "warning"
	SeverityError    = "error"
	SeverityCritical = "critical"
)

// Common activity actions
const (
	ActionLogin                    = "login"
	ActionLoginFailed              = "login_failed"
	ActionLogout                   = "logout"
	ActionAccountCreated           = "account_created"
	ActionPasswordChanged          = "password_changed"
	ActionPasswordResetRequest     = "password_reset_request"
	ActionPasswordReset            = "password_reset"
	Action2FAEnabled               = "2fa_enabled"
	Action2FADisabled              = "2fa_disabled"
	Action2FAVerified              = "2fa_verified"
	Action2FAFailed                = "2fa_failed"
	ActionRecoveryCodesRegenerated = "recovery_codes_regenerated"
	ActionProfileUpdated           = "profile_updated"
	ActionAddressUpdated           = "address_updated"
	ActionPreferencesUpdated       = "preferences_updated"
	ActionEmailVerified            = "email_verified"
	ActionEventCreated             = "event_created"
	ActionEventPublished           = "event_published"
	ActionEventUpdated             = "event_updated"
	ActionEventDeleted             = "event_deleted"
	ActionOrderPlaced              = "order_placed"
	ActionOrderCancelled           = "order_cancelled"
	ActionOrderRefunded            = "order_refunded"
	ActionPaymentProcessed         = "payment_processed"
	ActionPaymentFailed            = "payment_failed"
	ActionPaymentMethodAdded       = "payment_method_added"
	ActionPaymentMethodRemoved     = "payment_method_removed"
	ActionTicketGenerated          = "ticket_generated"
	ActionTicketTransferred        = "ticket_transferred"
	ActionTicketCheckedIn          = "ticket_checked_in"
	ActionRefundRequested          = "refund_requested"
	ActionRefundApproved           = "refund_approved"
	ActionRefundProcessed          = "refund_processed"
	ActionSettlementProcessed      = "settlement_processed"
	ActionStripeConnected          = "stripe_connected"
	ActionStripeDisconnected       = "stripe_disconnected"
	ActionAccountDeleted           = "account_deleted"
	ActionSecurityAlert            = "security_alert"
)
