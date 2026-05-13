package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	apievents "ticketing_system/internal/api_events"
	"ticketing_system/internal/config"
	"ticketing_system/internal/messaging"
	"ticketing_system/internal/notifications"
)

func main() {
	cfg := config.LoadOrPanic()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	_, sqsClient := messaging.NewAWSClients(ctx)
	consumer := messaging.NewSQSConsumer(sqsClient, cfg.Messaging.EmailWorkerQueue)
	notifService := notifications.NewNotificationService(cfg)

	log.Println("email-worker started, listening on", cfg.Messaging.EmailWorkerQueue)
	for {
		msg, err := consumer.ReceiveMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("email-worker: receive error: %v", err)
			continue
		}
		if msg == nil {
			continue
		}

		if err := handleEmail(notifService, msg); err != nil {
			log.Printf("email-worker: failed for %s: %v", msg.MessageID, err)
			// Do not delete — SQS redelivers after visibility timeout so the email is retried.
			continue
		}

		if err := consumer.DeleteMessage(ctx, msg.ReceiptHandle); err != nil {
			log.Printf("email-worker: failed to delete message: %v", err)
		}
	}
}

func handleEmail(notifService *notifications.NotificationService, msg *messaging.Message) error {
	var evt apievents.NotificationEmailEvent
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

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
