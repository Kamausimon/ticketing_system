package main

import (
	"fmt"
	"log"
	"os"

	"ticketing_system/internal/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Try to load .env file, but don't fail if it doesn't exist
	err := godotenv.Load("../.env") // Look in parent directory
	if err != nil {
		log.Println("⚠️  No .env file found, using environment variables or defaults", err)
	}

	dsn := os.Getenv("DSN")
	fmt.Printf("dsn: %v\n", dsn)
	if dsn == "" {
		dsn = "host=localhost port=5432 dbname=postgres user=postgres password=xxxxx connect_timeout=10 sslmode=prefer TimeZone=Africa/Nairobi"
		log.Println("📝 Using default database connection")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	err = runMigrations(db)
	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	log.Println("✅ All migrations completed successfully!")
}

func runMigrations(db *gorm.DB) error {
	log.Println("🚀 Starting database migrations...")

	// Create custom types first
	log.Println("🔧 Creating custom types...")
	err := createCustomTypes(db)
	if err != nil {
		return err
	}

	// Core models first (no dependencies)
	log.Println("📋 Migrating core models...")
	err = db.AutoMigrate(
		&models.Account{},
		&models.User{},
		&models.PaymentGateway{},
		&models.Venue{},
	)
	if err != nil {
		return err
	}

	// Models with basic dependencies
	log.Println("🏢 Migrating business models...")
	err = db.AutoMigrate(
		&models.Organizer{},
		&models.AccountPaymentGateway{},
		&models.Event{},
		&models.EventVenues{},
		&models.EventImages{},
	)
	if err != nil {
		return err
	}

	// Ticket-related models
	log.Println("🎫 Migrating ticket models...")
	err = db.AutoMigrate(
		&models.TicketClass{},
		&models.Order{},
		&models.OrderItem{},
		&models.Ticket{},
		&models.TicketOrder{},
		&models.ReservedTicket{},
		&models.Attendee{},
	)
	if err != nil {
		return err
	}

	// Payment and financial models
	log.Println("💰 Migrating payment models...")
	err = db.AutoMigrate(
		&models.PaymentTransaction{},
		&models.PaymentRecord{},
		&models.SettlementRecord{},
		&models.SettlementItem{},
		&models.RefundRecord{},
	)
	if err != nil {
		return err
	}

	// Promotion and marketing models
	log.Println("🎉 Migrating promotion models...")
	err = db.AutoMigrate(
		&models.Promotion{},
		&models.PromotionUsage{},
		&models.PromotionRule{},
	)
	if err != nil {
		return err
	}

	// Security and administrative models
	log.Println("🔒 Migrating security models...")
	err = db.AutoMigrate(
		&models.PasswordReset{},
		&models.PasswordResetAttempt{},
		&models.ResetConfiguration{},
	)
	if err != nil {
		return err
	}

	// Analytics and metrics models
	log.Println("📊 Migrating analytics models...")
	err = db.AutoMigrate(
		&models.EventStats{},
		&models.SystemMetric{},
		&models.EventMetric{},
		&models.UserEngagementMetric{},
		&models.SecurityMetric{},
	)
	if err != nil {
		return err
	}

	log.Println("✨ Creating custom indexes...")
	err = createCustomIndexes(db)
	if err != nil {
		return err
	}

	return nil
}

func createCustomTypes(db *gorm.DB) error {
	// Create ENUM types if you want to use them (optional)
	// Currently using varchar instead, so this is not needed
	// But keeping for future reference

	// Example of creating PostgreSQL ENUM:
	// err := db.Exec("CREATE TYPE role_type AS ENUM ('customer', 'organizer', 'admin');").Error
	// if err != nil {
	//     log.Printf("Warning: Could not create role_type enum (may already exist): %v", err)
	// }

	return nil
}

func createCustomIndexes(db *gorm.DB) error {
	// Custom composite indexes for better performance
	indexes := []string{
		// Payment performance indexes
		"CREATE INDEX IF NOT EXISTS idx_payment_records_lookup ON payment_records(type, status, initiated_at);",
		"CREATE INDEX IF NOT EXISTS idx_settlement_lookup ON settlement_records(status, earliest_payout_date);",

		// Promotion performance indexes
		"CREATE INDEX IF NOT EXISTS idx_promotions_active ON promotions(status, start_date, end_date) WHERE status = 'active';",
		"CREATE INDEX IF NOT EXISTS idx_promotions_usage ON promotions(usage_count, usage_limit) WHERE usage_limit IS NOT NULL;",

		// Security indexes
		"CREATE INDEX IF NOT EXISTS idx_password_resets_cleanup ON password_resets(cleanup_after, should_cleanup) WHERE should_cleanup = true;",
		"CREATE INDEX IF NOT EXISTS idx_security_events ON security_metrics(event_type, timestamp, severity);",

		// Analytics indexes
		"CREATE INDEX IF NOT EXISTS idx_metrics_time_series ON system_metrics(metric_name, granularity, timestamp);",
		"CREATE INDEX IF NOT EXISTS idx_event_analytics ON event_metrics(event_id, date);",
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			log.Printf("Warning: Failed to create index: %v", err)
			// Don't fail migration for index creation errors
		}
	}

	return nil
}
