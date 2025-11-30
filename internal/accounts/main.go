package accounts

import (
	"ticketing_system/internal/analytics"
	"ticketing_system/internal/models"
	"time"

	"gorm.io/gorm"
)

// AccountHandler handles all account-related operations
type AccountHandler struct {
	db       *gorm.DB
	_metrics *analytics.PrometheusMetrics // Reserved for future instrumentation
}

// NewAccountHandler creates a new account handler
func NewAccountHandler(db *gorm.DB, metrics *analytics.PrometheusMetrics) *AccountHandler {
	return &AccountHandler{
		db:       db,
		_metrics: metrics,
	}
}

// Account response structures
type AccountResponse struct {
	ID               uint                `json:"id"`
	FirstName        string              `json:"first_name"`
	LastName         string              `json:"last_name"`
	Email            string              `json:"email"`
	TimezoneID       *int                `json:"timezone_id"`
	DateFormatID     *int                `json:"date_format_id"`
	DateTimeFormatID *int                `json:"datetime_format_id"`
	CurrencyID       *int                `json:"currency_id"`
	LastIP           *string             `json:"last_ip"`
	LastLoginDate    *time.Time          `json:"last_login_date"`
	Address1         *string             `json:"address1"`
	Address2         *string             `json:"address2"`
	City             *string             `json:"city"`
	County           *string             `json:"county"`
	PostalCode       *string             `json:"postal_code"`
	IsActive         bool                `json:"is_active"`
	IsBanned         bool                `json:"is_banned"`
	PaymentGateway   *PaymentGatewayInfo `json:"payment_gateway,omitempty"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

type PaymentGatewayInfo struct {
	ID                   uint   `json:"id"`
	Name                 string `json:"name"`
	IsActive             bool   `json:"is_active"`
	HasStripeIntegration bool   `json:"has_stripe_integration"`
}

// Profile update request structure
type UpdateProfileRequest struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Email        string `json:"email"`
	TimezoneID   *int   `json:"timezone_id"`
	DateFormatID *int   `json:"date_format_id"`
	CurrencyID   *int   `json:"currency_id"`
}

// Address update request structure
type UpdateAddressRequest struct {
	Address1   *string `json:"address1"`
	Address2   *string `json:"address2"`
	City       *string `json:"city"`
	County     *string `json:"county"`
	PostalCode *string `json:"postal_code"`
}

// Security settings request structure
type SecuritySettingsRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// Account preferences structure
type AccountPreferences struct {
	TimezoneID         *int `json:"timezone_id"`
	DateFormatID       *int `json:"date_format_id"`
	DateTimeFormatID   *int `json:"datetime_format_id"`
	CurrencyID         *int `json:"currency_id"`
	EmailNotifications bool `json:"email_notifications"`
	SmsNotifications   bool `json:"sms_notifications"`
}

// Activity log structure
type AccountActivity struct {
	ID          uint      `json:"id"`
	AccountID   uint      `json:"account_id"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   *string   `json:"user_agent"`
	Timestamp   time.Time `json:"timestamp"`
}

// Payment method structure
type PaymentMethod struct {
	ID           uint   `json:"id"`
	Type         string `json:"type"`  // "stripe", "paypal", etc.
	Last4        string `json:"last4"` // Last 4 digits of card
	Brand        string `json:"brand"` // "visa", "mastercard", etc.
	ExpiryMonth  int    `json:"expiry_month"`
	ExpiryYear   int    `json:"expiry_year"`
	IsDefault    bool   `json:"is_default"`
	StripeCardID string `json:"stripe_card_id,omitempty"`
}

// Stripe integration request structure
type StripeIntegrationRequest struct {
	StripeAccessToken    string `json:"stripe_access_token"`
	StripeRefreshToken   string `json:"stripe_refresh_token"`
	StripeSecretKey      string `json:"stripe_secret_key"`
	StripePublishableKey string `json:"stripe_publishable_key"`
}

// Activity filter structure
type ActivityFilter struct {
	Page      int        `json:"page"`
	Limit     int        `json:"limit"`
	Action    *string    `json:"action"`
	Category  *string    `json:"category"`
	Success   *bool      `json:"success"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
}

// Account statistics structure
type AccountStats struct {
	TotalEvents     int        `json:"total_events"`
	TotalOrders     int        `json:"total_orders"`
	TotalRevenue    float64    `json:"total_revenue"`
	LastLoginDate   *time.Time `json:"last_login_date"`
	AccountAge      int        `json:"account_age_days"`
	IsEmailVerified bool       `json:"is_email_verified"`
	IsPhoneVerified bool       `json:"is_phone_verified"`
}

// Login history structure
type LoginHistory struct {
	ID        uint      `json:"id"`
	AccountID uint      `json:"account_id"`
	IPAddress string    `json:"ip_address"`
	UserAgent *string   `json:"user_agent"`
	Location  *string   `json:"location"`
	Success   bool      `json:"success"`
	Timestamp time.Time `json:"timestamp"`
}

// Pagination response structure
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	TotalCount int64       `json:"total_count"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}

// Helper function to convert models.Account to AccountResponse
func convertToAccountResponse(account models.Account) AccountResponse {
	response := AccountResponse{
		ID:               account.ID,
		FirstName:        account.FirstName,
		LastName:         account.LastName,
		Email:            account.Email,
		TimezoneID:       account.TimezoneID,
		DateFormatID:     account.DateFormatID,
		DateTimeFormatID: account.DateTimeFormatID,
		CurrencyID:       account.CurrencyID,
		LastIP:           account.LastIP,
		LastLoginDate:    account.LastLoginDate,
		Address1:         account.Address1,
		Address2:         account.Address2,
		City:             account.City,
		County:           account.County,
		PostalCode:       account.PostalCode,
		IsActive:         account.IsActive,
		IsBanned:         account.IsBanned,
		CreatedAt:        account.CreatedAt,
		UpdatedAt:        account.UpdatedAt,
	}

	// Add payment gateway info if available
	if account.PaymentGateway.ID > 0 {
		hasStripe := account.StripeAccessToken != nil &&
			account.StripeSecretKey != nil &&
			account.StripePublishableKey != nil

		response.PaymentGateway = &PaymentGatewayInfo{
			ID:                   account.PaymentGateway.ID,
			Name:                 account.PaymentGateway.Name,
			IsActive:             true, // Default to true for existing gateways
			HasStripeIntegration: hasStripe,
		}
	}

	return response
}

// NotificationSettings represents notification preferences
type NotificationSettings struct {
	EmailNotifications   bool `json:"email_notifications"`
	SMSNotifications     bool `json:"sms_notifications"`
	PushNotifications    bool `json:"push_notifications"`
	EventUpdates         bool `json:"event_updates"`
	PaymentNotifications bool `json:"payment_notifications"`
	SecurityAlerts       bool `json:"security_alerts"`
	MarketingEmails      bool `json:"marketing_emails"`
}

// AccountSettings represents account settings
type AccountSettings struct {
	Timezone       string               `json:"timezone"`
	DateFormat     string               `json:"date_format"`
	DateTimeFormat string               `json:"datetime_format"`
	Currency       string               `json:"currency"`
	Language       string               `json:"language"`
	Notifications  NotificationSettings `json:"notifications"`
}

// AddressInfo represents address information
type AddressInfo struct {
	Address1   string `json:"address_1"`
	Address2   string `json:"address_2"`
	City       string `json:"city"`
	County     string `json:"county"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
	IsDefault  bool   `json:"is_default"`
}

// PaymentMethodInfo represents payment method information
type PaymentMethodInfo struct {
	ID          uint      `json:"id"`
	Type        string    `json:"type"`
	Last4       string    `json:"last4"`
	Brand       string    `json:"brand"`
	ExpiryMonth int       `json:"expiry_month"`
	ExpiryYear  int       `json:"expiry_year"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
}

// ActivityLog represents an activity log entry
type ActivityLog struct {
	ID          uint      `json:"id"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	Timestamp   time.Time `json:"timestamp"`
	Success     bool      `json:"success"`
}
