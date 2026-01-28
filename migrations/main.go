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
	err := godotenv.Load(".env") // Look in parent directory
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
		&models.PaymentMethod{},
		&models.PaymentTransaction{},
		&models.PaymentRecord{},
		&models.SettlementRecord{},
		&models.SettlementItem{},
		&models.RefundRecord{},
		&models.RefundLineItem{},
		&models.PayoutAccount{},
		&models.WebhookLog{},
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
		&models.TwoFactorAuth{},
		&models.RecoveryCode{},
		&models.TwoFactorAttempt{},
		&models.TwoFactorSession{},
		&models.EmailVerification{},
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

	// Activity logging models
	log.Println("📝 Migrating activity logging models...")
	err = db.AutoMigrate(
		&models.AccountActivity{},
		&models.LoginHistory{},
		&models.NotificationPreferences{},
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
		"CREATE INDEX IF NOT EXISTS idx_payment_methods_active ON payment_methods(account_id, status, is_default) WHERE status = 'active';",
		"CREATE INDEX IF NOT EXISTS idx_payout_accounts_verified ON payout_accounts(organizer_id, status, is_verified) WHERE status = 'verified';",
		"CREATE INDEX IF NOT EXISTS idx_webhook_logs_processing ON webhook_logs(provider, status, created_at);",
		"CREATE INDEX IF NOT EXISTS idx_webhook_logs_events ON webhook_logs(event_id, event_type, created_at);",

		// Promotion performance indexes
		"CREATE INDEX IF NOT EXISTS idx_promotions_active ON promotions(status, start_date, end_date) WHERE status = 'active';",
		"CREATE INDEX IF NOT EXISTS idx_promotions_usage ON promotions(usage_count, usage_limit) WHERE usage_limit IS NOT NULL;",

		// Security indexes
		"CREATE INDEX IF NOT EXISTS idx_password_resets_cleanup ON password_resets(cleanup_after, should_cleanup) WHERE should_cleanup = true;",
		"CREATE INDEX IF NOT EXISTS idx_security_events ON security_metrics(event_type, timestamp, severity);",
		"CREATE INDEX IF NOT EXISTS idx_2fa_enabled_users ON two_factor_auths(user_id, enabled) WHERE enabled = true;",
		"CREATE INDEX IF NOT EXISTS idx_recovery_codes_unused ON recovery_codes(two_factor_auth_id, used) WHERE used = false;",
		"CREATE INDEX IF NOT EXISTS idx_2fa_attempts_recent ON two_factor_attempts(user_id, attempted_at, success);",
		"CREATE INDEX IF NOT EXISTS idx_2fa_sessions_active ON two_factor_sessions(user_id, expires_at, verified) WHERE verified = false;",

		// Analytics indexes
		"CREATE INDEX IF NOT EXISTS idx_metrics_time_series ON system_metrics(metric_name, granularity, timestamp);",
		"CREATE INDEX IF NOT EXISTS idx_event_analytics ON event_metrics(event_id, date);",

		// Refund line items indexes
		"CREATE INDEX IF NOT EXISTS idx_refund_line_items_lookup ON refund_line_items(refund_record_id, order_item_id);",
		"CREATE INDEX IF NOT EXISTS idx_refund_line_items_ticket ON refund_line_items(ticket_id) WHERE ticket_id IS NOT NULL;",

		// Activity logging indexes
		"CREATE INDEX IF NOT EXISTS idx_activities_recent ON account_activities(account_id, timestamp DESC);",
		"CREATE INDEX IF NOT EXISTS idx_activities_by_category ON account_activities(category, timestamp DESC);",
		"CREATE INDEX IF NOT EXISTS idx_activities_by_action ON account_activities(action, success, timestamp DESC);",
		"CREATE INDEX IF NOT EXISTS idx_activities_failed ON account_activities(success, severity, timestamp DESC) WHERE success = false;",
		"CREATE INDEX IF NOT EXISTS idx_login_history_recent ON login_history(account_id, login_at DESC);",
		"CREATE INDEX IF NOT EXISTS idx_login_history_failed ON login_history(ip_address, success, login_at DESC) WHERE success = false;",

		//searching index
		"CREATE INDEX IF NOT EXISTS idx_reserved_tickets_expires_deletedat ON reserved_tickets(expires_at, deleted_at) WHERE deleted_at IS NULL;",
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			log.Printf("Warning: Failed to create index: %v", err)
			// Don't fail migration for index creation errors
		}
	}

	return nil
}
