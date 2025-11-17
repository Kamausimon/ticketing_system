package database

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DbConfig struct {
	dsn string
}

func Init() *gorm.DB {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("There was an error reading the env variables", err)
	}
	dsn := os.Getenv("DSN")

	cfg := &DbConfig{
		dsn: dsn,
	}
	db, err := gorm.Open(postgres.Open(cfg.dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("err connecting to the db", err)
	}
	fmt.Print("successfully connected to the database")
	return db
}
