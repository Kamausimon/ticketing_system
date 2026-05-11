package outbox

import (
	"ticketing_system/internal/models"
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Save(tx *gorm.DB, event models.OutboxEvent) error {
	return tx.Create(&event).Error
}

// FetchPending returns up to limit unpublished outbox events, oldest first.
func (r *Repository) FetchPending(limit int) ([]models.OutboxEvent, error) {
	var events []models.OutboxEvent
	err := r.db.
		Where("status = ?", "pending").
		Order("created_at ASC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

// IncrementRetry bumps the retry counter. If exhausted is true the status is
// set to "failed" so the publisher stops retrying and it can be investigated.
func (r *Repository) IncrementRetry(id uint, exhausted bool) {
	status := "pending"
	if exhausted {
		status = "failed"
	}
	r.db.Model(&models.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"retry_count": gorm.Expr("retry_count + 1"),
			"status":      status,
		})
}

// MarkPublished sets status to published and records the publish timestamp.
func (r *Repository) MarkPublished(id uint) error {
	now := time.Now()
	return r.db.Model(&models.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       "published",
			"published_at": now,
		}).Error
}
