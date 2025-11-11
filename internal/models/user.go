package models

import (
	"database/sql/driver"

	"gorm.io/gorm"
)

type Role string

const (
	Customer  Role = "customer"
	Organizer Role = "organizer"
	Admin     Role = "admin"
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
	FirstName        string `gorm:"not null"`
	LastName         string `gorm:"not null"`
	Username         string `gorm:"unique;not null"`
	Phone            string `gorm:"unique;not null"`
	Email            string `gorm:"unique;not null"`
	Password         string `gorm:"not null"`
	ConfirmationCode string
	Isconfirmed      bool `gorm:"default:false"`
	Role             Role `gorm:"type:Role;default:'customer';not null"`
	IsActive         bool `gorm:"default:true"`
	ProfilePicture   *string
}
