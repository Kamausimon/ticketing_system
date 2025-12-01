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
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("There was an error reading the env variables", err)
	}
	dsn := os.Getenv("DSN")

	cfg := &DbConfig{
		dsn: dsn,
	}

	// Configure GORM with performance optimizations
	db, err := gorm.Open(postgres.Open(cfg.dsn), &gorm.Config{
		PrepareStmt: true, // Cache prepared statements for faster queries
		Logger:      logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("err connecting to the db", err)
	}

	// Get underlying SQL DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("failed to get database instance", err)
	}

	// Configure connection pool for optimal performance
	sqlDB.SetMaxIdleConns(10)                  // Keep 10 idle connections ready
	sqlDB.SetMaxOpenConns(100)                 // Allow up to 100 concurrent connections
	sqlDB.SetConnMaxLifetime(time.Hour)        // Recycle connections every hour
	sqlDB.SetConnMaxIdleTime(10 * time.Minute) // Close idle connections after 10 minutes

	fmt.Println("successfully connected to the database with optimized connection pool")
	return db
}
