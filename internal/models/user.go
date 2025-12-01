package models

import (
	"database/sql/driver"
	"time"

	"gorm.io/gorm"
)

type Role string

const (
	RoleCustomer  Role = "customer"
	RoleOrganizer Role = "organizer"
	RoleAdmin     Role = "admin"
)

func (P *Role) Scan(value interface{}) error {
	*P = Role(value.([]byte))
	return nil
}

func (P Role) Value() (driver.Value, error) {
	return string(P), nil
}

type User struct {
	gorm.Model
	AccountID        uint    `gorm:"not null;index:idx_user_account"`
	Account          Account `gorm:"foreignKey:AccountID"`
	FirstName        string  `gorm:"not null"`
	LastName         string  `gorm:"not null"`
	Username         string  `gorm:"uniqueIndex;not null"`
	Phone            string  `gorm:"uniqueIndex;not null"`
	Email            string  `gorm:"uniqueIndex;not null"`
	Password         string  `gorm:"not null"`
	ConfirmationCode string
	Isconfirmed      bool `gorm:"default:false"`
	Role             Role `gorm:"type:varchar(20);default:'customer';not null"`
	IsActive         bool `gorm:"default:true"`
	ProfilePicture   *string

	// Email verification fields
	EmailVerified        bool       `gorm:"default:false;index"`
	EmailVerifiedAt      *time.Time `gorm:"index"`
	VerificationTokenExp *time.Time

	// JWT Token fields
	RefreshToken    *string `gorm:"type:text"` // Store refresh token
	RefreshTokenExp *int64  // Refresh token expiration timestamp
	LastLoginAt     *int64  // Track last login time
	TokenVersion    int     `gorm:"default:1"` // For token invalidation
}
