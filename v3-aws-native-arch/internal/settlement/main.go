package settlement

import (
	"gorm.io/gorm"
)

// Service handles all settlement-related operations
type Service struct {
	db *gorm.DB
}

// NewService creates a new settlement service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}
