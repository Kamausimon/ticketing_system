package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"ticketing_system/internal/config"
	"ticketing_system/internal/database"
	"ticketing_system/internal/messaging"
	"ticketing_system/internal/models"
	"ticketing_system/internal/outbox"
)

func main() {
	cfg := config.LoadOrPanic()
	db := database.Init()

	if err := db.AutoMigrate(&models.OutboxEvent{}); err != nil {
		log.Fatalf("outbox: failed to migrate outbox table: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	snsClient, _ := messaging.NewAWSClients(ctx)
	publisher := messaging.NewSNSPublisher(snsClient, cfg.Messaging.TopicARNs)

	repo := outbox.NewRepository(db)
	outbox.NewPublisher(repo, publisher).Run(ctx)
}
