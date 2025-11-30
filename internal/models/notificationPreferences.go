package models

import "gorm.io/gorm"

// NotificationPreferences stores user notification preferences
type NotificationPreferences struct {
	gorm.Model
	AccountID            uint    `gorm:"uniqueIndex;not null"`
	Account              Account `gorm:"foreignKey:AccountID"`
	EmailNotifications   bool    `gorm:"default:true"`
	SMSNotifications     bool    `gorm:"default:false"`
	PushNotifications    bool    `gorm:"default:true"`
	EventUpdates         bool    `gorm:"default:true"`
	PaymentNotifications bool    `gorm:"default:true"`
	SecurityAlerts       bool    `gorm:"default:true"`
	MarketingEmails      bool    `gorm:"default:false"`
}

// TableName overrides the table name
func (NotificationPreferences) TableName() string {
	return "notification_preferences"
}
