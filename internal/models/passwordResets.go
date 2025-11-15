package models

import (
	"time"

	"gorm.io/gorm"
)

// ResetStatus defines the current state of a password reset request
type ResetStatus string

const (
	ResetPending ResetStatus = "pending" // Request created, waiting for user action
	ResetUsed    ResetStatus = "used"    // Token was used successfully
	ResetExpired ResetStatus = "expired" // Token has expired
	ResetRevoked ResetStatus = "revoked" // Token was manually revoked
	ResetInvalid ResetStatus = "invalid" // Token was marked invalid (security)
)

// ResetMethod defines how the reset was initiated
type ResetMethod string

const (
	ResetByEmail   ResetMethod = "email"   // Email-based reset
	ResetByPhone   ResetMethod = "phone"   // SMS-based reset
	ResetByAdmin   ResetMethod = "admin"   // Admin-initiated reset
	ResetBySupport ResetMethod = "support" // Support team reset
)

// PasswordReset represents a password reset request
// Designed with security best practices and efficient cleanup
type PasswordReset struct {
	gorm.Model

	// Core identification
	Token  string      `gorm:"unique;not null;index"`            // Secure random token
	Email  string      `gorm:"not null;index"`                   // Email for reset
	Status ResetStatus `gorm:"not null;default:'pending';index"` // Current status
	Method ResetMethod `gorm:"not null;default:'email'"`         // How reset was initiated

	// User association
	UserID    *uint    `gorm:"index"`                // Associated user (if found)
	User      *User    `gorm:"foreignKey:UserID"`    // User relationship
	AccountID *uint    `gorm:"index"`                // Associated account
	Account   *Account `gorm:"foreignKey:AccountID"` // Account relationship

	// Security tracking
	IPAddress    string `gorm:"not null"`           // IP that requested reset
	UserAgent    string `gorm:"not null"`           // Browser/device info
	AttemptCount int    `gorm:"default:0"`          // Number of use attempts
	MaxAttempts  int    `gorm:"not null;default:3"` // Maximum allowed attempts

	// Time constraints (critical for security)
	ExpiresAt     time.Time  `gorm:"not null;index"` // When token expires
	IssuedAt      time.Time  `gorm:"not null;index"` // When token was issued
	UsedAt        *time.Time // When token was used
	RevokedAt     *time.Time // When token was revoked
	LastAttemptAt *time.Time // Last usage attempt

	// Usage validation
	OriginalIP     string  `gorm:"not null"` // IP that created request
	UsedFromIP     *string // IP that used token
	SameIPRequired bool    `gorm:"default:false"` // Must use from same IP

	// Security features
	RequireCurrentPassword bool `gorm:"default:false"` // Require current password
	RequireTwoFactor       bool `gorm:"default:false"` // Require 2FA
	IsSecurityReset        bool `gorm:"default:false"` // High-security reset

	// Administrative tracking
	RequestedBy     *uint `gorm:"index"`                  // Who requested (admin/support)
	RequestedByUser *User `gorm:"foreignKey:RequestedBy"` // User who requested
	ApprovedBy      *uint `gorm:"index"`                  // Who approved (for admin resets)
	ApprovedByUser  *User `gorm:"foreignKey:ApprovedBy"`  // User who approved

	// Rate limiting data
	RateLimitKey    string     `gorm:"index"` // Key for rate limiting (IP+Email)
	PreviousResetAt *time.Time // Previous reset for this user
	CooldownUntil   *time.Time `gorm:"index"` // Cooldown period end

	// Metadata and notes
	ResetReason *string // Reason for reset request
	AdminNotes  *string // Admin notes
	UserMessage *string // Message to user

	// Cleanup optimization
	ShouldCleanup bool      `gorm:"default:true;index"` // Mark for cleanup
	CleanupAfter  time.Time `gorm:"not null;index"`     // When to clean up

	// Soft delete (security: keep audit trail)
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// PasswordResetAttempt tracks individual attempts to use reset tokens
// Separate table for security auditing without impacting main performance
type PasswordResetAttempt struct {
	gorm.Model

	// Reset association
	PasswordResetID uint          `gorm:"not null;index"`
	PasswordReset   PasswordReset `gorm:"foreignKey:PasswordResetID"`

	// Attempt details
	IPAddress     string    `gorm:"not null;index"`         // IP of attempt
	UserAgent     string    `gorm:"not null"`               // Browser info
	AttemptedAt   time.Time `gorm:"not null;index"`         // When attempted
	WasSuccessful bool      `gorm:"not null;default:false"` // Did attempt succeed

	// Security validation results
	TokenValid      bool `gorm:"default:false"` // Was token valid
	NotExpired      bool `gorm:"default:false"` // Was token not expired
	IPMatched       bool `gorm:"default:true"`  // IP validation result
	RateLimitPassed bool `gorm:"default:true"`  // Rate limit check

	// Failure details
	FailureReason *string // Why attempt failed
	ErrorCode     *string // System error code

	// Geographic tracking (for security analysis)
	Country *string // Country of IP
	City    *string // City of IP
	ISP     *string // Internet service provider

	// Response time for performance monitoring
	ResponseTimeMs *int // Response time in milliseconds
}

// ResetConfiguration holds system-wide password reset settings
// These can be stored in database or environment variables
type ResetConfiguration struct {
	gorm.Model

	// Token settings
	TokenLength        int    `gorm:"not null;default:32"`       // Length of reset token
	TokenExpiryMinutes int    `gorm:"not null;default:15"`       // Token validity period
	TokenAlgorithm     string `gorm:"not null;default:'random'"` // Token generation method

	// Security settings
	MaxAttemptsPerToken int `gorm:"not null;default:3"`  // Max attempts per token
	MaxRequestsPerHour  int `gorm:"not null;default:5"`  // Max requests per user/hour
	MaxRequestsPerIP    int `gorm:"not null;default:10"` // Max requests per IP/hour
	CooldownMinutes     int `gorm:"not null;default:30"` // Cooldown between requests

	// IP validation
	RequireSameIP     bool `gorm:"default:false"` // Must use from same IP
	AllowVPNs         bool `gorm:"default:true"`  // Allow VPN usage
	BlockKnownProxies bool `gorm:"default:false"` // Block known proxy IPs

	// Cleanup settings
	CleanupAfterDays   int  `gorm:"not null;default:7"`  // Clean up old records
	KeepAuditDays      int  `gorm:"not null;default:90"` // Keep audit records
	AutoCleanupEnabled bool `gorm:"default:true"`        // Enable auto cleanup

	// Notification settings
	SendConfirmationEmail bool `gorm:"default:true"` // Send confirmation after reset
	NotifyOnSuspicious    bool `gorm:"default:true"` // Notify on suspicious activity
	LogAllAttempts        bool `gorm:"default:true"` // Log all attempts

	// Feature flags
	EmailResetEnabled bool `gorm:"default:true"`  // Email-based resets
	SMSResetEnabled   bool `gorm:"default:false"` // SMS-based resets
	AdminResetEnabled bool `gorm:"default:true"`  // Admin-initiated resets

	// Configuration metadata
	ConfigName  string `gorm:"unique;not null"` // Configuration name
	Description string `gorm:"not null"`        // Human-readable description
	IsActive    bool   `gorm:"default:true"`    // Is this config active

	// Audit
	CreatedBy          uint  `gorm:"not null"` // Admin who created
	CreatedByUser      User  `gorm:"foreignKey:CreatedBy"`
	LastModifiedBy     *uint `gorm:"index"` // Last modifier
	LastModifiedByUser *User `gorm:"foreignKey:LastModifiedBy"`
}
