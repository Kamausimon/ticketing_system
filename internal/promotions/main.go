package promotions

import (
	"crypto/rand"
	"encoding/base32"
	"strings"
	"ticketing_system/internal/analytics"
	"ticketing_system/internal/models"
	"time"

	"gorm.io/gorm"
)

// PromotionHandler handles all promotion-related operations
type PromotionHandler struct {
	db      *gorm.DB
	metrics *analytics.PrometheusMetrics
}

// NewPromotionHandler creates a new promotion handler
func NewPromotionHandler(db *gorm.DB, metrics *analytics.PrometheusMetrics) *PromotionHandler {
	return &PromotionHandler{
		db:      db,
		metrics: metrics,
	}
}

// PromotionResponse represents the promotion response structure
type PromotionResponse struct {
	ID                 uint                   `json:"id"`
	Code               string                 `json:"code"`
	Name               string                 `json:"name"`
	Description        string                 `json:"description"`
	Type               models.PromotionType   `json:"type"`
	Status             models.PromotionStatus `json:"status"`
	Target             models.PromotionTarget `json:"target"`
	DiscountPercentage *int32                 `json:"discount_percentage,omitempty"`
	DiscountAmount     *models.Money          `json:"discount_amount,omitempty"`
	FreeQuantity       *int32                 `json:"free_quantity,omitempty"`
	MinimumPurchase    *models.Money          `json:"minimum_purchase,omitempty"`
	MaximumDiscount    *models.Money          `json:"maximum_discount,omitempty"`
	EventID            *uint                  `json:"event_id,omitempty"`
	EventTitle         string                 `json:"event_title,omitempty"`
	OrganizerID        *uint                  `json:"organizer_id,omitempty"`
	OrganizerName      string                 `json:"organizer_name,omitempty"`
	StartDate          time.Time              `json:"start_date"`
	EndDate            time.Time              `json:"end_date"`
	EarlyBirdCutoff    *time.Time             `json:"early_bird_cutoff,omitempty"`
	UsageLimit         *int32                 `json:"usage_limit,omitempty"`
	UsageCount         int32                  `json:"usage_count"`
	PerUserLimit       *int32                 `json:"per_user_limit,omitempty"`
	PerOrderLimit      *int32                 `json:"per_order_limit,omitempty"`
	IsUnlimited        bool                   `json:"is_unlimited"`
	IsPublic           bool                   `json:"is_public"`
	FirstTimeCustomers bool                   `json:"first_time_customers"`
	TotalRevenue       models.Money           `json:"total_revenue"`
	TotalDiscount      models.Money           `json:"total_discount"`
	ConversionRate     *float64               `json:"conversion_rate,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

// PromotionListResponse represents a paginated list of promotions
type PromotionListResponse struct {
	Promotions []PromotionResponse `json:"promotions"`
	TotalCount int64               `json:"total_count"`
	Page       int                 `json:"page"`
	Limit      int                 `json:"limit"`
	TotalPages int                 `json:"total_pages"`
}

// CreatePromotionRequest represents the request to create a promotion
type CreatePromotionRequest struct {
	Code               string                 `json:"code"`
	Name               string                 `json:"name"`
	Description        string                 `json:"description"`
	Type               models.PromotionType   `json:"type"`
	Target             models.PromotionTarget `json:"target"`
	DiscountPercentage *int32                 `json:"discount_percentage,omitempty"`
	DiscountAmount     *int64                 `json:"discount_amount,omitempty"` // in cents
	FreeQuantity       *int32                 `json:"free_quantity,omitempty"`
	MinimumPurchase    *int64                 `json:"minimum_purchase,omitempty"` // in cents
	MaximumDiscount    *int64                 `json:"maximum_discount,omitempty"` // in cents
	EventID            *uint                  `json:"event_id,omitempty"`
	OrganizerID        *uint                  `json:"organizer_id,omitempty"`
	StartDate          time.Time              `json:"start_date"`
	EndDate            time.Time              `json:"end_date"`
	EarlyBirdCutoff    *time.Time             `json:"early_bird_cutoff,omitempty"`
	UsageLimit         *int32                 `json:"usage_limit,omitempty"`
	PerUserLimit       *int32                 `json:"per_user_limit,omitempty"`
	PerOrderLimit      *int32                 `json:"per_order_limit,omitempty"`
	IsPublic           bool                   `json:"is_public"`
	FirstTimeCustomers bool                   `json:"first_time_customers"`
	GenerateCode       bool                   `json:"generate_code"` // Auto-generate code
}

// UpdatePromotionRequest represents the request to update a promotion
type UpdatePromotionRequest struct {
	Name               *string    `json:"name,omitempty"`
	Description        *string    `json:"description,omitempty"`
	DiscountPercentage *int32     `json:"discount_percentage,omitempty"`
	DiscountAmount     *int64     `json:"discount_amount,omitempty"`
	MinimumPurchase    *int64     `json:"minimum_purchase,omitempty"`
	MaximumDiscount    *int64     `json:"maximum_discount,omitempty"`
	EndDate            *time.Time `json:"end_date,omitempty"`
	UsageLimit         *int32     `json:"usage_limit,omitempty"`
	PerUserLimit       *int32     `json:"per_user_limit,omitempty"`
	IsPublic           *bool      `json:"is_public,omitempty"`
}

// ValidatePromotionRequest represents the request to validate a promotion
type ValidatePromotionRequest struct {
	Code           string `json:"code"`
	EventID        *uint  `json:"event_id,omitempty"`
	TicketClassID  *uint  `json:"ticket_class_id,omitempty"`
	OrderAmount    int64  `json:"order_amount"` // in cents
	TicketQuantity int    `json:"ticket_quantity"`
	AccountID      uint   `json:"account_id"`
}

// ValidatePromotionResponse represents the validation result
type ValidatePromotionResponse struct {
	Valid          bool   `json:"valid"`
	PromotionID    uint   `json:"promotion_id,omitempty"`
	Code           string `json:"code"`
	DiscountAmount int64  `json:"discount_amount"` // in cents
	FinalAmount    int64  `json:"final_amount"`    // in cents
	Message        string `json:"message"`
	ErrorReason    string `json:"error_reason,omitempty"`
}

// PromotionStatsResponse represents promotion statistics
type PromotionStatsResponse struct {
	TotalPromotions   int64        `json:"total_promotions"`
	ActivePromotions  int64        `json:"active_promotions"`
	ExpiredPromotions int64        `json:"expired_promotions"`
	TotalUsage        int64        `json:"total_usage"`
	TotalRevenue      models.Money `json:"total_revenue"`
	TotalDiscount     models.Money `json:"total_discount"`
	AverageDiscount   float64      `json:"average_discount"`
	ConversionRate    float64      `json:"conversion_rate"`
}

// PromotionUsageResponse represents a promotion usage record
type PromotionUsageResponse struct {
	ID             uint         `json:"id"`
	PromotionCode  string       `json:"promotion_code"`
	OrderID        uint         `json:"order_id"`
	AccountEmail   string       `json:"account_email"`
	DiscountAmount models.Money `json:"discount_amount"`
	OriginalAmount models.Money `json:"original_amount"`
	FinalAmount    models.Money `json:"final_amount"`
	UsedAt         time.Time    `json:"used_at"`
}

// PromotionAnalyticsResponse represents detailed analytics
type PromotionAnalyticsResponse struct {
	PromotionID      uint         `json:"promotion_id"`
	Code             string       `json:"code"`
	TotalUsage       int64        `json:"total_usage"`
	UniqueUsers      int64        `json:"unique_users"`
	TotalRevenue     models.Money `json:"total_revenue"`
	TotalDiscount    models.Money `json:"total_discount"`
	AverageOrderSize models.Money `json:"average_order_size"`
	ConversionRate   float64      `json:"conversion_rate"`
	ROI              float64      `json:"roi"` // Return on investment
	UsageByDay       []DailyUsage `json:"usage_by_day"`
}

// DailyUsage represents usage statistics for a single day
type DailyUsage struct {
	Date     string `json:"date"`
	Count    int64  `json:"count"`
	Revenue  int64  `json:"revenue"`
	Discount int64  `json:"discount"`
}

// PromotionFilter represents filtering options
type PromotionFilter struct {
	Page        int
	Limit       int
	Status      *models.PromotionStatus
	Type        *models.PromotionType
	EventID     *uint
	OrganizerID *uint
	SearchTerm  string
	StartDate   *time.Time
	EndDate     *time.Time
	IsPublic    *bool
}

// Helper function to convert models.Promotion to PromotionResponse
func convertToPromotionResponse(promo models.Promotion) PromotionResponse {
	response := PromotionResponse{
		ID:                 promo.ID,
		Code:               promo.Code,
		Name:               promo.Name,
		Description:        promo.Description,
		Type:               promo.Type,
		Status:             promo.Status,
		Target:             promo.Target,
		DiscountPercentage: promo.DiscountPercentage,
		DiscountAmount:     promo.DiscountAmount,
		FreeQuantity:       promo.FreeQuantity,
		MinimumPurchase:    promo.MinimumPurchase,
		MaximumDiscount:    promo.MaximumDiscount,
		EventID:            promo.EventID,
		OrganizerID:        promo.OrganizerID,
		StartDate:          promo.StartDate,
		EndDate:            promo.EndDate,
		EarlyBirdCutoff:    promo.EarlyBirdCutoff,
		UsageLimit:         promo.UsageLimit,
		UsageCount:         promo.UsageCount,
		PerUserLimit:       promo.PerUserLimit,
		PerOrderLimit:      promo.PerOrderLimit,
		IsUnlimited:        promo.IsUnlimited,
		IsPublic:           promo.IsPublic,
		FirstTimeCustomers: promo.FirstTimeCustomers,
		TotalRevenue:       promo.TotalRevenue,
		TotalDiscount:      promo.TotalDiscount,
		ConversionRate:     promo.ConversionRate,
		CreatedAt:          promo.CreatedAt,
		UpdatedAt:          promo.UpdatedAt,
	}

	// Add event details if loaded
	if promo.Event != nil {
		response.EventTitle = promo.Event.Title
	}

	// Add organizer details if loaded
	if promo.Organizer != nil {
		response.OrganizerName = promo.Organizer.Name
	}

	return response
}

// Helper function to generate a random promotion code
func generatePromotionCode() string {
	// Generate 8 random bytes
	b := make([]byte, 8)
	rand.Read(b)

	// Encode to base32 and trim padding
	code := base32.StdEncoding.EncodeToString(b)
	code = strings.TrimRight(code, "=")

	// Take first 8 characters and make uppercase
	if len(code) > 8 {
		code = code[:8]
	}

	return strings.ToUpper(code)
}

// Helper function to calculate discount amount
func calculatePromotionDiscount(promo *models.Promotion, orderAmount models.Money) models.Money {
	var discount models.Money

	switch promo.Type {
	case models.PromotionPercentage:
		if promo.DiscountPercentage != nil {
			discount = models.Money(int64(orderAmount) * int64(*promo.DiscountPercentage) / 100)
		}
	case models.PromotionFixedAmount:
		if promo.DiscountAmount != nil {
			discount = *promo.DiscountAmount
		}
	case models.PromotionEarlyBird:
		if promo.DiscountPercentage != nil {
			discount = models.Money(int64(orderAmount) * int64(*promo.DiscountPercentage) / 100)
		}
	case models.PromotionBulk:
		if promo.DiscountPercentage != nil {
			discount = models.Money(int64(orderAmount) * int64(*promo.DiscountPercentage) / 100)
		}
	}

	// Apply maximum discount cap if set
	if promo.MaximumDiscount != nil && discount > *promo.MaximumDiscount {
		discount = *promo.MaximumDiscount
	}

	// Ensure discount doesn't exceed order amount
	if discount > orderAmount {
		discount = orderAmount
	}

	return discount
}

// Helper function to check if promotion is currently valid
func isPromotionCurrentlyValid(promo *models.Promotion) bool {
	now := time.Now()

	// Check status
	if promo.Status != models.PromotionActive {
		return false
	}

	// Check dates
	if now.Before(promo.StartDate) || now.After(promo.EndDate) {
		return false
	}

	// Check usage limit
	if !promo.IsUnlimited && promo.UsageLimit != nil {
		if promo.UsageCount >= *promo.UsageLimit {
			return false
		}
	}

	return true
}
