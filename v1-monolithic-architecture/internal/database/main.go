package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DbConfig struct {
	dsn string
}

func Init() *gorm.DB {
	// Try to load .env file (for local development)
	// On Railway/production, environment variables are already set
	err := godotenv.Load(".env")
	if err != nil {
		// Not fatal - Railway injects variables as system environment variables
		log.Printf("⚠️  .env file not found (using system environment variables): %v\n", err)
	}
	
	// Try DATABASE_URL first (Railway, Heroku standard), then fall back to DSN
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DSN")
	}
	
	// Validate DSN is set
	if dsn == "" {
		log.Fatal("❌ Database connection string not found. Set DATABASE_URL or DSN environment variable")
	}
	
	log.Printf("📊 Connecting to database...")

	cfg := &DbConfig{
		dsn: dsn,
	}

	// Configure GORM with performance optimizations
	db, err := gorm.Open(postgres.Open(cfg.dsn), &gorm.Config{
		PrepareStmt: false, // Disable prepared statements to avoid caching issues
		Logger:      logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v\nConnection string format: postgresql://user:pass@host:port/dbname", err)
	}

	// Get underlying SQL DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("failed to get database instance", err)
	}
	if err := sqlDB.Ping(); err != nil {
		fmt.Println("🔴 DB PING FAILED:", err)
	}

	// Configure connection pool - aggressive cleanup to prevent hanging connections
	sqlDB.SetMaxIdleConns(0)                   // No idle connections - close immediately after use
	sqlDB.SetMaxOpenConns(25)                  // Increase pool size for concurrent requests
	sqlDB.SetConnMaxLifetime(2 * time.Minute)  // Recycle connections every 2 minutes
	sqlDB.SetConnMaxIdleTime(10 * time.Second) // Close idle connections after 10 seconds

	fmt.Println("successfully connected to the database with optimized connection pool")
	return db
}
