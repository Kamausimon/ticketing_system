package models

import "time"

type OutboxEvent struct {
	ID          uint   `gorm:"primaryKey"`
	Topic       string `gorm:"index;not null"`
	Payload     string `gorm:"type:json;not null"`
	Status      string `gorm:"default:'pending';not null;index"`
	CreatedAt   time.Time
	PublishedAt *time.Time
	RetryCount  int `gorm:"default:0;not null"`
}