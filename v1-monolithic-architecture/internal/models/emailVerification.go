package models

import (
	"time"

	"gorm.io/gorm"
)

// EmailVerificationStatus represents the status of email verification
type EmailVerificationStatus string

const (
	VerificationPending  EmailVerificationStatus = "pending"
	VerificationVerified EmailVerificationStatus = "verified"
	VerificationExpired  EmailVerificationStatus = "expired"
	VerificationInvalid  EmailVerificationStatus = "invalid"
	VerificationResent   EmailVerificationStatus = "resent"
)

// EmailVerification stores email verification tokens
type EmailVerification struct {
	gorm.Model
	UserID      uint                    `gorm:"index:idx_email_verification_user"`
	User        User                    `gorm:"foreignKey:UserID"`
	Token       string                  `gorm:"uniqueIndex"`                        // Verification token
	Email       string                  `gorm:"index:idx_email_verification_email"` // Email to be verified
	Status      EmailVerificationStatus `gorm:"default:'pending';index:idx_email_verification_status"`
	VerifiedAt  *time.Time              // When email was verified
	ExpiresAt   time.Time               `gorm:"index"` // Token expiration time
	LastSentAt  time.Time               // Last time email was sent
	ResendCount int                     `gorm:"default:0"` // Number of times resent
	MaxResends  int                     `gorm:"default:3"` // Maximum number of resends allowed
	IPAddress   string                  // IP address of verification request
	UserAgent   string                  // User agent for security tracking
	IssuedAt    time.Time               // When token was issued
}

// TableName specifies the table name for EmailVerification
func (EmailVerification) TableName() string {
	return "email_verifications"
}
