package tickets

import (
	"fmt"
	"log"
	"os"

	"ticketing_system/internal/models"
	"ticketing_system/internal/notifications"
)

// sendTicketEmailWithPDF sends a ticket email with PDF attachment
func (h *TicketHandler) sendTicketEmailWithPDF(ticket *models.Ticket, pdfPath string) {
	if h.notificationService == nil {
		log.Printf("⚠️ Notification service not available, skipping email for ticket %s", ticket.TicketNumber)
		return
	}

	// Load full ticket data with relations
	var fullTicket models.Ticket
	if err := h.db.Preload("OrderItem.TicketClass.Event.Venue").
		Preload("OrderItem.Order").
		Where("id = ?", ticket.ID).
		First(&fullTicket).Error; err != nil {
		log.Printf("⚠️ Failed to load ticket details for email: %v", err)
		return
	}

	event := fullTicket.OrderItem.TicketClass.Event

	// Read PDF file
	pdfData, err := os.ReadFile(pdfPath)
	if err != nil {
		log.Printf("⚠️ Failed to read PDF file for email: %v", err)
		return
	}

	// Prepare email with attachment
	emailData := notifications.EmailData{
		To:      []string{fullTicket.HolderEmail},
		Subject: fmt.Sprintf("Your Ticket for %s", event.Title),
		HTMLBody: fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #8B5CF6; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 5px 5px; }
        .ticket-box { background: white; padding: 20px; border-radius: 5px; margin: 20px 0; border-left: 4px solid #8B5CF6; }
        .button { display: inline-block; background: #8B5CF6; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎫 Your Ticket is Ready!</h1>
        </div>
        <div class="content">
            <p>Hi %s,</p>
            <p>Your ticket for <strong>%s</strong> is now ready! Your PDF ticket is attached to this email.</p>
            
            <div class="ticket-box">
                <h3>Ticket Details</h3>
                <p><strong>Event:</strong> %s</p>
                <p><strong>Date:</strong> %s</p>
                <p><strong>Location:</strong> %s</p>
                <p><strong>Ticket Type:</strong> %s</p>
                <p><strong>Ticket Number:</strong> %s</p>
            </div>

            <p><strong>What's Next?</strong></p>
            <ul>
                <li>Download and save your ticket PDF</li>
                <li>Show the QR code on your ticket at entry</li>
                <li>Arrive early on the event day</li>
            </ul>

            <p><strong>Important:</strong> Keep your ticket safe. You'll need to show the QR code to check in at the event.</p>

            <div class="footer">
                <p>Questions? Contact us at support@ticketing.local</p>
                <p>&copy; 2024 Ticketing System. All rights reserved.</p>
            </div>
        </div>
    </div>
</body>
</html>
`, fullTicket.HolderName, event.Title, event.Title, event.StartDate.Format("Monday, January 2, 2006"), event.Location, fullTicket.OrderItem.TicketClass.Name, fullTicket.TicketNumber),
		Attachments: []notifications.Attachment{
			{
				Filename: fmt.Sprintf("ticket_%s.pdf", fullTicket.TicketNumber),
				Content:  pdfData,
				MimeType: "application/pdf",
			},
		},
	}

	// Send email
	err = h.notificationService.GetEmailService().Send(emailData)
	if err != nil {
		log.Printf("❌ Failed to send ticket email to %s for ticket %s: %v", fullTicket.HolderEmail, fullTicket.TicketNumber, err)
		return
	}

	log.Printf("✅ Ticket email sent to %s for ticket %s", fullTicket.HolderEmail, fullTicket.TicketNumber)
}
