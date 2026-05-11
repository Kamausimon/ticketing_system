package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	apievents "ticketing_system/internal/api_events"
	"ticketing_system/internal/config"
	"ticketing_system/internal/database"
	"ticketing_system/internal/kafka"
	kafkatopics "ticketing_system/internal/kafka"
	"ticketing_system/internal/models"
	"ticketing_system/internal/notifications"
	"ticketing_system/pkg/pdf"
	"ticketing_system/pkg/qrcode"

	gokafka "github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

func main() {
	db := database.Init()
	cfg := config.LoadOrPanic()

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	if len(brokers) == 0 || brokers[0] == "" {
		brokers = []string{"localhost:9092"}
	}

	consumer := kafka.NewConsumer(brokers, kafkatopics.TicketsGeneratedTopic, "ticket-email-worker")
	defer consumer.Close()

	notifService := notifications.NewNotificationService(cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Println("ticket-email-worker started, listening on", kafkatopics.TicketsGeneratedTopic)
	for {
		msg, err := consumer.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("ticket-email-worker: read error: %v", err)
			continue
		}

		if err := handleTicketsGenerated(db, notifService, msg); err != nil {
			log.Printf("ticket-email-worker: failed for %s: %v", msg.Key, err)
			// Do not commit — Kafka redelivers so we can retry the whole order.
			continue
		}

		if err := consumer.CommitMessage(ctx, msg); err != nil {
			log.Printf("ticket-email-worker: failed to commit offset: %v", err)
		}
	}
}

func handleTicketsGenerated(db *gorm.DB, notifService *notifications.NotificationService, msg gokafka.Message) error {
	var evt apievents.TicketsGeneratedEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	// Load all tickets for this order with the relations needed for PDF generation.
	var tickets []models.Ticket
	if err := db.
		Preload("OrderItem.TicketClass.Event").
		Preload("OrderItem.Order").
		Joins("JOIN order_items ON tickets.order_item_id = order_items.id").
		Where("order_items.order_id = ?", evt.OrderID).
		Find(&tickets).Error; err != nil {
		return fmt.Errorf("load tickets for order %d: %w", evt.OrderID, err)
	}

	if len(tickets) == 0 {
		log.Printf("ticket-email-worker: no tickets found for order %d, skipping", evt.OrderID)
		return nil
	}

	pdfGen := pdf.NewTicketGenerator()
	qrGen := qrcode.NewGenerator().WithSize(512)

	for i := range tickets {
		t := &tickets[i]
		if err := processTicket(db, notifService, t, pdfGen, qrGen); err != nil {
			// Log per-ticket failures but continue — one bad PDF shouldn't
			// block other attendees on the same order from getting their emails.
			log.Printf("ticket-email-worker: failed to process ticket %s: %v", t.TicketNumber, err)
		}
	}

	log.Printf("ticket-email-worker: sent ticket emails for order %d (%d tickets)", evt.OrderID, len(tickets))
	return nil
}

func processTicket(db *gorm.DB, notifService *notifications.NotificationService, t *models.Ticket, pdfGen *pdf.TicketGenerator, qrGen *qrcode.Generator) error {
	event := &t.OrderItem.TicketClass.Event
	order := &t.OrderItem.Order

	// Generate QR code bytes.
	qrContent := fmt.Sprintf("TICKET:%s|EVENT:%d|ATTENDEE:%s", t.TicketNumber, event.ID, t.HolderName)
	qrBytes, err := qrGen.GenerateBytes(qrContent)
	if err != nil {
		return fmt.Errorf("generate QR: %w", err)
	}

	ticketData := pdf.TicketData{
		TicketNumber:  t.TicketNumber,
		EventName:     event.Title,
		EventDate:     event.StartDate,
		EventTime:     event.StartDate.Format("3:04 PM"),
		VenueName:     event.Location,
		VenueAddress:  event.Location,
		AttendeeName:  t.HolderName,
		AttendeeEmail: t.HolderEmail,
		TicketType:    t.OrderItem.TicketClass.Name,
		Price:         float64(t.OrderItem.UnitPrice) / 100.0,
		Currency:      string(order.Currency),
		QRCode:        qrBytes,
		OrderNumber:   fmt.Sprintf("ORD-%d", order.ID),
		PurchaseDate:  order.CreatedAt,
	}

	storageDir := filepath.Join("storage", "tickets", fmt.Sprintf("%d", order.ID))
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return fmt.Errorf("create storage dir: %w", err)
	}

	pdfPath := filepath.Join(storageDir, fmt.Sprintf("ticket_%s.pdf", t.TicketNumber))
	if err := pdfGen.GenerateToFile(ticketData, pdfPath); err != nil {
		return fmt.Errorf("generate PDF: %w", err)
	}

	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		return fmt.Errorf("read PDF: %w", err)
	}

	// Persist the PDF path so the ticket record knows where its PDF lives.
	db.Model(t).Update("pdf_path", pdfPath)

	emailData := notifications.EmailData{
		To:      []string{t.HolderEmail},
		Subject: fmt.Sprintf("Your Ticket for %s", event.Title),
		HTMLBody: fmt.Sprintf(`<p>Hi <strong>%s</strong>,</p>
<p>Your ticket for <strong>%s</strong> is attached to this email.</p>
<p><strong>Ticket Number:</strong> <code>%s</code><br>
<strong>Date:</strong> %s<br>
<strong>Location:</strong> %s</p>
<p>Please present your ticket (printed or on your phone) at the event entrance.</p>`,
			t.HolderName,
			event.Title,
			t.TicketNumber,
			event.StartDate.Format("Monday, January 2, 2006 at 3:04 PM"),
			event.Location,
		),
		Attachments: []notifications.Attachment{
			{
				Filename: fmt.Sprintf("ticket_%s.pdf", t.TicketNumber),
				Content:  pdfBytes,
				MimeType: "application/pdf",
			},
		},
	}

	if err := notifService.GetEmailService().Send(emailData); err != nil {
		return fmt.Errorf("send email to %s: %w", t.HolderEmail, err)
	}

	log.Printf("ticket-email-worker: sent PDF ticket email to %s for ticket %s", t.HolderEmail, t.TicketNumber)
	return nil
}
