package models

import (
	"time"

	"gorm.io/gorm"
)

// MetricType defines the type of metric being tracked
type MetricType string

const (
	MetricSales       MetricType = "sales"       // Revenue and sales data
	MetricEngagement  MetricType = "engagement"  // User interaction metrics
	MetricPerformance MetricType = "performance" // System performance
	MetricSecurity    MetricType = "security"    // Security-related events
	MetricInventory   MetricType = "inventory"   // Ticket availability
	MetricConversion  MetricType = "conversion"  // Conversion funnel
	MetricRetention   MetricType = "retention"   // User retention
)

// MetricGranularity defines the time granularity of metrics
type MetricGranularity string

const (
	GranularityMinute MetricGranularity = "minute"
	GranularityHour   MetricGranularity = "hour"
	GranularityDay    MetricGranularity = "day"
	GranularityWeek   MetricGranularity = "week"
	GranularityMonth  MetricGranularity = "month"
	GranularityYear   MetricGranularity = "year"
)

// SystemMetric represents system-wide metrics
// Use this for high-level KPIs and business metrics
type SystemMetric struct {
	gorm.Model

	// Metric identification
	MetricName  string            `gorm:"not null;index:idx_metric_lookup"` // "total_revenue", "active_users"
	MetricType  MetricType        `gorm:"not null;index:idx_metric_lookup"` // Type of metric
	Granularity MetricGranularity `gorm:"not null;index:idx_metric_lookup"` // Time granularity

	// Time dimension
	Timestamp time.Time `gorm:"not null;index:idx_metric_lookup"`            // When metric was recorded
	Date      time.Time `gorm:"type:date;index"`                             // Date for daily rollups
	Hour      int       `gorm:"check:hour >= 0 AND hour <= 23"`              // Hour (0-23)
	DayOfWeek int       `gorm:"check:day_of_week >= 0 AND day_of_week <= 6"` // Day of week (0=Sunday)
	Week      int       `gorm:"check:week >= 1 AND week <= 53"`              // Week of year
	Month     int       `gorm:"check:month >= 1 AND month <= 12"`            // Month
	Year      int       `gorm:"not null"`                                    // Year

	// Metric values
	Value float64  `gorm:"not null"`  // Main metric value
	Count int64    `gorm:"default:0"` // Count (for averages)
	Sum   float64  `gorm:"default:0"` // Sum (for totals)
	Min   *float64 // Minimum value
	Max   *float64 // Maximum value

	// Dimensional breakdowns
	EventID     *uint      `gorm:"index"` // Event-specific metrics
	Event       *Event     `gorm:"foreignKey:EventID"`
	OrganizerID *uint      `gorm:"index"` // Organizer-specific metrics
	Organizer   *Organizer `gorm:"foreignKey:OrganizerID"`
	AccountID   *uint      `gorm:"index"` // Account-specific metrics
	Account     *Account   `gorm:"foreignKey:AccountID"`

	// Geographic dimensions
	Country *string `gorm:"index"` // Country code
	Region  *string `gorm:"index"` // State/region
	City    *string `gorm:"index"` // City

	// Additional dimensions (stored as JSON for flexibility)
	Dimensions string `gorm:"type:text"` // JSON object with custom dimensions
	Tags       string `gorm:"type:text"` // JSON array of tags

	// Metadata
	Source     string `gorm:"not null;default:'system'"` // Where metric came from
	Version    int    `gorm:"default:1"`                 // Metric schema version
	IsEstimate bool   `gorm:"default:false"`             // Is this an estimated value

	// Indexes for efficient queries
	// Composite index: (metric_name, metric_type, granularity, timestamp)
}

// EventMetric represents event-specific detailed metrics
// Use this for detailed event analytics
type EventMetric struct {
	gorm.Model

	// Event association
	EventID uint  `gorm:"not null;index:idx_event_metrics"`
	Event   Event `gorm:"foreignKey:EventID"`

	// Time tracking
	Date time.Time `gorm:"type:date;not null;index:idx_event_metrics"`
	Hour int       `gorm:"check:hour >= 0 AND hour <= 23"`

	// Traffic metrics
	PageViews      int64   `gorm:"default:0"` // Event page views
	UniqueVisitors int64   `gorm:"default:0"` // Unique visitors
	BounceRate     float64 `gorm:"default:0"` // Bounce rate percentage
	AvgTimeOnPage  int     `gorm:"default:0"` // Average time in seconds

	// Sales funnel metrics
	AddToCart        int64   `gorm:"default:0"` // Add to cart events
	CheckoutStart    int64   `gorm:"default:0"` // Checkout initiations
	CheckoutComplete int64   `gorm:"default:0"` // Completed purchases
	ConversionRate   float64 `gorm:"default:0"` // Conversion percentage

	// Financial metrics
	GrossRevenue Money `gorm:"default:0"` // Total revenue
	NetRevenue   Money `gorm:"default:0"` // After fees
	PlatformFees Money `gorm:"default:0"` // Platform fees collected
	PaymentFees  Money `gorm:"default:0"` // Payment gateway fees
	RefundAmount Money `gorm:"default:0"` // Refunds issued

	// Ticket metrics
	TicketsSold        int64 `gorm:"default:0"` // Total tickets sold
	TicketsRefunded    int64 `gorm:"default:0"` // Tickets refunded
	TicketsCheckedIn   int64 `gorm:"default:0"` // Tickets used at event
	InventoryRemaining int64 `gorm:"default:0"` // Tickets still available

	// Promotional metrics
	PromoCodeUses int64 `gorm:"default:0"` // Promo codes used
	PromoDiscount Money `gorm:"default:0"` // Total discount given

	// Geographic breakdown
	TopCountries string `gorm:"type:text"` // JSON array of top countries
	TopCities    string `gorm:"type:text"` // JSON array of top cities

	// Device/platform breakdown
	MobilePercent  float64 `gorm:"default:0"` // Mobile traffic percentage
	DesktopPercent float64 `gorm:"default:0"` // Desktop traffic percentage
	AppPercent     float64 `gorm:"default:0"` // Mobile app percentage
}

// UserEngagementMetric tracks user behavior patterns
// Use this for understanding user journeys
type UserEngagementMetric struct {
	gorm.Model

	// User identification
	AccountID uint    `gorm:"not null;index:idx_user_engagement"`
	Account   Account `gorm:"foreignKey:AccountID"`

	// Time dimension
	Date         time.Time `gorm:"type:date;not null;index:idx_user_engagement"`
	SessionStart time.Time `gorm:"not null"`
	SessionEnd   *time.Time

	// Session metrics
	SessionDuration int `gorm:"default:0"` // Duration in seconds
	PageViews       int `gorm:"default:0"` // Pages viewed in session
	EventsViewed    int `gorm:"default:0"` // Events viewed
	SearchQueries   int `gorm:"default:0"` // Searches performed

	// Engagement actions
	TicketsPurchased int `gorm:"default:0"` // Tickets bought in session
	EventsBookmarked int `gorm:"default:0"` // Events saved/favorited
	SocialShares     int `gorm:"default:0"` // Social media shares
	EmailSignups     int `gorm:"default:0"` // Newsletter signups

	// Revenue impact
	RevenueGenerated Money `gorm:"default:0"` // Revenue from this session

	// Technical details
	UserAgent string  `gorm:"not null"` // Browser info
	IPAddress string  `gorm:"not null"` // IP address
	Country   *string `gorm:"index"`    // Geographic location
	City      *string

	// Attribution
	ReferrerSource *string `gorm:"index"` // How they found us
	CampaignID     *string `gorm:"index"` // Marketing campaign
	UTMSource      *string // UTM tracking
	UTMCampaign    *string
	UTMMedium      *string
}

// SecurityMetric tracks security-related events
// Use this for monitoring and alerting
type SecurityMetric struct {
	gorm.Model

	// Event identification
	EventType string `gorm:"not null;index:idx_security_events"` // "failed_login", "suspicious_activity"
	Severity  string `gorm:"not null;index"`                     // "low", "medium", "high", "critical"

	// Time tracking
	Timestamp time.Time `gorm:"not null;index:idx_security_events"` // When event occurred
	Date      time.Time `gorm:"type:date;index"`                    // Date for daily rollups
	Hour      int       `gorm:"check:hour >= 0 AND hour <= 23"`     // Hour of day

	// Source information
	IPAddress string  `gorm:"not null;index"` // Source IP
	UserAgent string  `gorm:"not null"`       // User agent
	Country   *string `gorm:"index"`          // IP geolocation

	// User context (if applicable)
	AccountID *uint    `gorm:"index"` // Target account
	Account   *Account `gorm:"foreignKey:AccountID"`
	UserID    *uint    `gorm:"index"` // Target user
	User      *User    `gorm:"foreignKey:UserID"`

	// Event details
	Description string `gorm:"not null"`  // Human-readable description
	RawData     string `gorm:"type:text"` // JSON with full event data

	// Risk assessment
	RiskScore   int     `gorm:"default:0;check:risk_score >= 0 AND risk_score <= 100"` // 0-100
	IsBlocked   bool    `gorm:"default:false"`                                         // Was action blocked
	ActionTaken *string // What action was taken

	// Resolution
	IsResolved     bool       `gorm:"default:false;index"` // Has been reviewed
	ResolvedAt     *time.Time // When resolved
	ResolvedBy     *uint      `gorm:"index"` // Who resolved
	ResolvedByUser *User      `gorm:"foreignKey:ResolvedBy"`
	Resolution     *string    // Resolution details
}
