package database

import (
	"ticketing_system/internal/models"

	"gorm.io/gorm"
)

// MigrateTicketTransferHistory creates the ticket_transfer_histories table
func MigrateTicketTransferHistory(db *gorm.DB) error {
	return db.AutoMigrate(&models.TicketTransferHistory{})
}
