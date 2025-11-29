package models

import (
	"time"

	"gorm.io/gorm"
)

// TwoFactorAuth stores 2FA configuration for users
type TwoFactorAuth struct {
	gorm.Model
	UserID          uint       `gorm:"not null;uniqueIndex;index"` // One 2FA config per user
	User            User       `gorm:"foreignKey:UserID"`
	Enabled         bool       `gorm:"default:false;index"`
	Secret          string     `gorm:"type:varchar(255);not null"` // Encrypted TOTP secret
	BackupCodesHash string     `gorm:"type:text"`                  // Hashed backup codes (JSON array)
	VerifiedAt      *time.Time `gorm:"index"`                      // When 2FA was first enabled
	LastUsedAt      *time.Time // Last time 2FA was used for login
	Method          string     `gorm:"type:varchar(20);default:'totp'"` // totp, sms, email (for future)
}

// RecoveryCode stores single-use recovery codes for 2FA
type RecoveryCode struct {
	gorm.Model
	TwoFactorAuthID uint          `gorm:"not null;index"`
	TwoFactorAuth   TwoFactorAuth `gorm:"foreignKey:TwoFactorAuthID"`
	CodeHash        string        `gorm:"type:varchar(255);not null;uniqueIndex"` // Hashed recovery code
	Used            bool          `gorm:"default:false;index"`
	UsedAt          *time.Time
	UsedFromIP      *string `gorm:"type:varchar(45)"` // IPv4 or IPv6
}

// TwoFactorAttempt logs 2FA verification attempts
type TwoFactorAttempt struct {
	gorm.Model
	UserID      uint      `gorm:"not null;index"`
	User        User      `gorm:"foreignKey:UserID"`
	Success     bool      `gorm:"not null;index"`
	IPAddress   string    `gorm:"type:varchar(45);not null"` // IPv4 or IPv6
	UserAgent   string    `gorm:"type:text"`
	FailureType string    `gorm:"type:varchar(50)"` // invalid_code, expired_code, rate_limited, etc.
	AttemptedAt time.Time `gorm:"not null;index"`
}

// TwoFactorSession represents a temporary session during 2FA setup
type TwoFactorSession struct {
	gorm.Model
	UserID    uint      `gorm:"not null;index"`
	User      User      `gorm:"foreignKey:UserID"`
	Secret    string    `gorm:"type:varchar(255);not null"` // Temporary secret until verified
	Verified  bool      `gorm:"default:false"`
	ExpiresAt time.Time `gorm:"not null;index"`
	IPAddress string    `gorm:"type:varchar(45)"`
	UserAgent string    `gorm:"type:text"`
}

// TableName overrides the table name
func (TwoFactorAuth) TableName() string {
	return "two_factor_auths"
}

func (RecoveryCode) TableName() string {
	return "recovery_codes"
}

func (TwoFactorAttempt) TableName() string {
	return "two_factor_attempts"
}

func (TwoFactorSession) TableName() string {
	return "two_factor_sessions"
}
