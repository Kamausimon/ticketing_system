package main

import (
	"fmt"
	"log"
	"time"

	"ticketing_system/pkg/pdf"
	"ticketing_system/pkg/qrcode"
)

func main() {
	fmt.Println("🎫 PDF Ticket Generator Example")
	fmt.Println("=================================")

	// Example ticket data
	ticketData := pdf.TicketData{
		TicketNumber:  "TKT-2024-VIP-001234",
		EventName:     "Summer Music Festival 2024",
		EventDate:     time.Date(2024, 7, 15, 19, 0, 0, 0, time.UTC),
		EventTime:     "7:00 PM - 11:00 PM",
		VenueName:     "Central Park Arena",
		VenueAddress:  "123 Park Avenue, New York, NY 10001",
		AttendeeName:  "John Doe",
		AttendeeEmail: "john.doe@example.com",
		TicketType:    "VIP",
		SeatNumber:    "A-15",
		Price:         150.00,
		Currency:      "USD",
		OrderNumber:   "ORD-2024-001234",
		PurchaseDate:  time.Now(),
		SpecialNotes:  "⚠️ Please arrive 30 minutes early. Gates open at 6:30 PM.",
	}

	// Step 1: Generate QR Code
	fmt.Println("📱 Step 1: Generating QR code...")
	qrContent := fmt.Sprintf("TICKET:%s|EVENT:%s|ATTENDEE:%s",
		ticketData.TicketNumber,
		ticketData.EventName,
		ticketData.AttendeeName,
	)

	qrGenerator := qrcode.NewGenerator().WithSize(512)
	qrBytes, err := qrGenerator.GenerateBytes(qrContent)
	if err != nil {
		log.Fatalf("Failed to generate QR code: %v", err)
	}
	fmt.Printf("✅ QR code generated (%d bytes)\n\n", len(qrBytes))

	// Add QR code to ticket data
	ticketData.QRCode = qrBytes

	// Step 2: Generate PDF Ticket
	fmt.Println("📄 Step 2: Generating PDF ticket...")
	pdfGenerator := pdf.NewTicketGenerator()

	pdfBytes, err := pdfGenerator.Generate(ticketData)
	if err != nil {
		log.Fatalf("Failed to generate PDF: %v", err)
	}
	fmt.Printf("✅ PDF generated (%d bytes)\n\n", len(pdfBytes))

	// Step 3: Save to file
	fmt.Println("💾 Step 3: Saving ticket to file...")
	filename := fmt.Sprintf("ticket_%s.pdf", ticketData.TicketNumber)
	err = pdfGenerator.GenerateToFile(ticketData, filename)
	if err != nil {
		log.Fatalf("Failed to save PDF: %v", err)
	}
	fmt.Printf("✅ Ticket saved to: %s\n\n", filename)

	fmt.Println("🎉 Done! Open the PDF file to see your ticket.")
	fmt.Println("\nThe ticket includes:")
	fmt.Println("  ✅ Event details (name, date, time, venue)")
	fmt.Println("  ✅ Attendee information")
	fmt.Println("  ✅ Scannable QR code for check-in")
	fmt.Println("  ✅ Ticket and order numbers")
	fmt.Println("  ✅ Professional design")
}
