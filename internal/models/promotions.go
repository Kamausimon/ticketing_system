package models

import (
	"time"

	"gorm.io/gorm"
)

// PromotionType defines the type of promotion
type PromotionType string

const (
	PromotionPercentage  PromotionType = "percentage"    // 20% off
	PromotionFixedAmount PromotionType = "fixed_amount"  // $10 off
	PromotionFreeTickets PromotionType = "free_tickets"  // Buy 2 get 1 free
	PromotionEarlyBird   PromotionType = "early_bird"    // Time-based discount
	PromotionBulk        PromotionType = "bulk_discount" // Volume discount
)

// PromotionStatus defines the current state of the promotion
type PromotionStatus string

const (
	PromotionDraft     PromotionStatus = "draft"     // Being created
	PromotionActive    PromotionStatus = "active"    // Live and usable
	PromotionPaused    PromotionStatus = "paused"    // Temporarily disabled
	PromotionExpired   PromotionStatus = "expired"   // Past end date
	PromotionExhausted PromotionStatus = "exhausted" // Usage limit reached
	PromotionCancelled PromotionStatus = "cancelled" // Manually cancelled
)

// PromotionTarget defines what the promotion applies to
type PromotionTarget string

const (
	TargetEntireOrder    PromotionTarget = "entire_order"    // Apply to whole order
	TargetSpecificTicket PromotionTarget = "specific_ticket" // Specific ticket class
	TargetEvent          PromotionTarget = "event"           // Entire event
	TargetCategory       PromotionTarget = "category"        // Event category
)

// Promotion represents a discount/promotional offer
// Designed for high-performance during traffic spikes
type Promotion struct {
	gorm.Model

	// Basic promotion info
	Code        string          `gorm:"unique;not null;index"` // SUMMER20, EARLY50
	Name        string          `gorm:"not null"`              // "Summer Sale 20% Off"
	Description string          `gorm:"not null"`              // Marketing description
	Type        PromotionType   `gorm:"not null;index"`        // Type of discount
	Status      PromotionStatus `gorm:"not null;index"`        // Current status
	Target      PromotionTarget `gorm:"not null"`              // What it applies to

	// Discount configuration
	DiscountPercentage *int32 `gorm:"check:discount_percentage <= 100"` // 0-100 percentage
	DiscountAmount     *Money // Fixed amount discount
	FreeQuantity       *int32 // Free tickets quantity
	MinimumPurchase    *Money // Minimum order amount
	MaximumDiscount    *Money // Cap the discount amount

	// Scope and targeting
	EventID         *uint      `gorm:"index"` // Specific event (if applicable)
	Event           *Event     `gorm:"foreignKey:EventID"`
	TicketClassIDs  string     `gorm:"type:text"` // JSON array of ticket class IDs
	EventCategories string     `gorm:"type:text"` // JSON array of categories
	OrganizerID     *uint      `gorm:"index"`     // Specific organizer
	Organizer       *Organizer `gorm:"foreignKey:OrganizerID"`

	// Time constraints (indexed for fast queries)
	StartDate       time.Time  `gorm:"not null;index"` // When promotion starts
	EndDate         time.Time  `gorm:"not null;index"` // When promotion ends
	EarlyBirdCutoff *time.Time `gorm:"index"`          // Early bird deadline

	// Usage limits and tracking (critical for performance)
	UsageLimit    *int32 // Max total uses (NULL = unlimited)
	UsageCount    int32  `gorm:"default:0;index"` // Current usage count
	PerUserLimit  *int32 // Max uses per user
	PerOrderLimit *int32 `gorm:"default:1"` // Max uses per order

	// Performance optimization fields
	IsUnlimited       bool       `gorm:"default:false;index"` // Unlimited usage (fast path)
	PrecomputedActive bool       `gorm:"default:false;index"` // Is currently active (cached)
	LastUsageCheck    *time.Time // Last time usage was verified

	// Customer restrictions
	FirstTimeCustomers bool   `gorm:"default:false"` // First-time customers only
	MinimumAge         *int32 // Age restriction
	AllowedUserIDs     string `gorm:"type:text"` // JSON array of specific user IDs
	ExcludedUserIDs    string `gorm:"type:text"` // JSON array of excluded users

	// Administrative
	CreatedBy        uint `gorm:"not null;index"` // Admin who created
	CreatedByUser    User `gorm:"foreignKey:CreatedBy"`
	IsPublic         bool `gorm:"default:true"`  // Public promo code?
	RequiresApproval bool `gorm:"default:false"` // Needs approval to use?

	// Analytics and tracking
	TotalRevenue   Money    `gorm:"default:0"` // Revenue generated
	TotalDiscount  Money    `gorm:"default:0"` // Total discount given
	ConversionRate *float64 // Views to usage rate

	// Metadata
	InternalNotes *string // Admin notes
	MarketingTags string  `gorm:"type:text"` // JSON array of tags

	// Soft delete
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// PromotionUsage tracks individual uses of promotions
// Separate table for performance - don't join during checkout
type PromotionUsage struct {
	gorm.Model

	// Core tracking
	PromotionID uint      `gorm:"not null;index:idx_promo_usage"`
	Promotion   Promotion `gorm:"foreignKey:PromotionID"`
	OrderID     uint      `gorm:"not null;index:idx_order_usage"`
	Order       Order     `gorm:"foreignKey:OrderID"`
	AccountID   uint      `gorm:"not null;index:idx_user_usage"`
	Account     Account   `gorm:"foreignKey:AccountID"`

	// Usage details
	DiscountAmount Money `gorm:"not null"` // Actual discount applied
	OriginalAmount Money `gorm:"not null"` // Order amount before discount
	FinalAmount    Money `gorm:"not null"` // Order amount after discount

	// Performance tracking
	UsedAt    time.Time `gorm:"not null;index"` // When used
	IPAddress *string   // User IP
	UserAgent *string   // Browser info

	// Validation results (for debugging)
	ValidationTime *time.Duration // How long validation took
	CacheHit       bool           `gorm:"default:false"` // Was promotion cached?
}

// PromotionCache represents cached promotion data for high-speed lookups
// This would typically be stored in Redis, but we define the structure here
type PromotionCache struct {
	Code               string `json:"code"`
	IsActive           bool   `json:"is_active"`
	Type               string `json:"type"`
	DiscountPercentage *int32 `json:"discount_percentage,omitempty"`
	DiscountAmount     *int64 `json:"discount_amount,omitempty"`  // Money in cents
	MinimumPurchase    *int64 `json:"minimum_purchase,omitempty"` // Money in cents
	MaximumDiscount    *int64 `json:"maximum_discount,omitempty"` // Money in cents
	UsageLimit         *int32 `json:"usage_limit,omitempty"`
	UsageCount         int32  `json:"usage_count"`
	IsUnlimited        bool   `json:"is_unlimited"`
	StartDate          int64  `json:"start_date"` // Unix timestamp
	EndDate            int64  `json:"end_date"`   // Unix timestamp
	EventID            *uint  `json:"event_id,omitempty"`
	TicketClassIDs     []uint `json:"ticket_class_ids,omitempty"`
	PerUserLimit       *int32 `json:"per_user_limit,omitempty"`
	LastUpdated        int64  `json:"last_updated"` // Cache timestamp
}

// PromotionRule represents business rules for promotion validation
// Pre-computed for fast evaluation during checkout
type PromotionRule struct {
	ID          uint      `gorm:"primaryKey"`
	PromotionID uint      `gorm:"not null;index"`
	Promotion   Promotion `gorm:"foreignKey:PromotionID"`

	// Rule definition
	RuleType     string `gorm:"not null;index"` // "min_amount", "user_limit", "time_window"
	RuleOperator string `gorm:"not null"`       // ">=", "<=", "==", "in", "not_in"
	RuleValue    string `gorm:"not null"`       // JSON value to compare
	ErrorMessage string `gorm:"not null"`       // User-friendly error message

	// Performance
	IsActive       bool  `gorm:"default:true;index"`
	ExecutionOrder int32 `gorm:"default:0"` // Order of rule evaluation

	CreatedAt time.Time
	UpdatedAt time.Time
}
