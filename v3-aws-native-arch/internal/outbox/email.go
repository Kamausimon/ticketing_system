package outbox

import (
	"encoding/json"
	"fmt"
	"time"

	apievents "ticketing_system/internal/api_events"
	kafkatopics "ticketing_system/internal/messaging"
	"ticketing_system/internal/models"

	"gorm.io/gorm"
)

// QueueEmail saves a pre-built email to the outbox so the email-worker sends
// it asynchronously with retry. Call this in place of any fire-and-forget
// goroutine that calls notificationService.Send*().
//
// Pass the db handle from the handler — no separate outboxRepo wiring needed.
// If the call is inside an open transaction, pass the tx so the outbox write
// is atomic with the surrounding business operation.
func QueueEmail(db *gorm.DB, to []string, subject, body, emailType string) error {
	evt := apievents.NotificationEmailEvent{
		To:        to,
		Subject:   subject,
		TextBody:  body,
		EmailType: emailType,
		Timestamp: time.Now(),
	}
	payload, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("outbox.QueueEmail marshal: %w", err)
	}
	return db.Create(&models.OutboxEvent{
		Topic:   kafkatopics.NotificationEmailTopic,
		Payload: string(payload),
		Status:  "pending",
	}).Error
}

// QueueHTMLEmail is the same as QueueEmail but for HTML-body emails.
func QueueHTMLEmail(db *gorm.DB, to []string, subject, htmlBody, emailType string) error {
	evt := apievents.NotificationEmailEvent{
		To:        to,
		Subject:   subject,
		HTMLBody:  htmlBody,
		EmailType: emailType,
		Timestamp: time.Now(),
	}
	payload, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("outbox.QueueHTMLEmail marshal: %w", err)
	}
	return db.Create(&models.OutboxEvent{
		Topic:   kafkatopics.NotificationEmailTopic,
		Payload: string(payload),
		Status:  "pending",
	}).Error
}
