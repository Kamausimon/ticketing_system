package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"ticketing_system/internal/database"
	"ticketing_system/internal/kafka"
	"ticketing_system/internal/models"
	"ticketing_system/internal/outbox"
)

func main() {
	db := database.Init()

	// Migrate the outbox table if it doesn't exist yet.
	if err := db.AutoMigrate(&models.OutboxEvent{}); err != nil {
		log.Fatalf("outbox: failed to migrate outbox table: %v", err)
	}

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	if len(brokers) == 0 || brokers[0] == "" {
		brokers = []string{"localhost:9092"}
	}

	producer := kafka.NewProducer(brokers)
	defer producer.Close()

	repo := outbox.NewRepository(db)
	publisher := outbox.NewPublisher(repo, producer)

	// Run until SIGINT or SIGTERM.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	publisher.Run(ctx)
}
