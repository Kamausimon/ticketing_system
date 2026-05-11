package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	apievents "ticketing_system/internal/api_events"
	"ticketing_system/internal/config"
	"ticketing_system/internal/kafka"
	kafkatopics "ticketing_system/internal/kafka"
	"ticketing_system/internal/notifications"

	gokafka "github.com/segmentio/kafka-go"
)

func main() {
	cfg := config.LoadOrPanic()

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	if len(brokers) == 0 || brokers[0] == "" {
		brokers = []string{"localhost:9092"}
	}

	consumer := kafka.NewConsumer(brokers, kafkatopics.NotificationEmailTopic, "email-worker")
	defer consumer.Close()

	notifService := notifications.NewNotificationService(cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Println("email-worker started, listening on", kafkatopics.NotificationEmailTopic)
	for {
		msg, err := consumer.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("email-worker: read error: %v", err)
			continue
		}

		if err := handleEmail(notifService, msg); err != nil {
			log.Printf("email-worker: failed for %s: %v", msg.Key, err)
			// Do not commit — Kafka redelivers so the email will be retried.
			continue
		}

		if err := consumer.CommitMessage(ctx, msg); err != nil {
			log.Printf("email-worker: failed to commit offset: %v", err)
		}
	}
}

func handleEmail(notifService *notifications.NotificationService, msg gokafka.Message) error {
	var evt apievents.NotificationEmailEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	// Prefer HTML body when available; fall back to plain text.
	if evt.HTMLBody != "" {
		if err := notifService.SendHTMLEmail(evt.To, evt.Subject, evt.HTMLBody); err != nil {
			return fmt.Errorf("send HTML email (%s): %w", evt.EmailType, err)
		}
	} else {
		if err := notifService.SendPlainEmail(evt.To, evt.Subject, evt.TextBody); err != nil {
			return fmt.Errorf("send plain email (%s): %w", evt.EmailType, err)
		}
	}

	log.Printf("email-worker: sent %s email to %v", evt.EmailType, evt.To)
	return nil
}
