package pdf

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// TicketData
type TicketData struct {
	TicketNumber  string
	EventName     string
	EventDate     time.Time
	EventTime     string
	VenueName     string
	VenueAddress  string
	AttendeeName  string
	AttendeeEmail string
	TicketType    string
	SeatNumber    string
	Price         float64
	Currency      string
	QRCode        []byte
	OrderNumber   string
	PurchaseDate  time.Time
	SpecialNotes  string
}

// TicketGenerator
type TicketGenerator struct {
	pdf            *gofpdf.Fpdf
	primaryColor   string
	secondaryColor string
	logoPath       string
}

// NewTicketGenerator
func NewTicketGenerator() *TicketGenerator {
	pdf := gofpdf.New("P", "mm", "A4", "")
	return &TicketGenerator{
		pdf:            pdf,
		primaryColor:   "#4F46E5", // Indigo
		secondaryColor: "#10B981", // Green
	}
}

// WithLogo sets the logo path
func (g *TicketGenerator) WithLogo(logoPath string) *TicketGenerator {
	g.logoPath = logoPath
	return g
}

// WithColors sets custom colors
func (g *TicketGenerator) WithColors(primary, secondary string) *TicketGenerator {
	g.primaryColor = primary
	g.secondaryColor = secondary
	return g
}

// Generate creates a PDF ticket
func (g *TicketGenerator) Generate(data TicketData) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "", 12)

	if err := g.drawTicket(pdf, data); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to output PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// drawTicket draws the ticket content
func (g *TicketGenerator) drawTicket(pdf *gofpdf.Fpdf, data TicketData) error {

	pageWidth, pageHeight := pdf.GetPageSize()
	marginX := 15.0
	marginY := 15.0

	pdf.SetDrawColor(79, 70, 229)
	pdf.SetLineWidth(0.5)
	pdf.Rect(marginX, marginY, pageWidth-(2*marginX), pageHeight-(2*marginY), "D")

	currentY := marginY + 10

	if g.logoPath != "" {
		pdf.Image(g.logoPath, marginX+5, currentY, 30, 0, false, "", 0, "")
	}

	pdf.SetFont("Arial", "", 20)
	pdf.SetTextColor(79, 70, 229)
	pdf.SetXY(marginX+40, currentY)
	pdf.CellFormat(pageWidth-(2*marginX)-45, 10, data.EventName, "", 0, "L", false, 0, "")
	currentY += 15

	pdf.SetFillColor(16, 185, 129)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(marginX+40, currentY)
	pdf.CellFormat(40, 7, data.TicketType, "0", 0, "C", true, 0, "")
	currentY += 15

	// Divider line
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(marginX+5, currentY, pageWidth-marginX-5, currentY)
	currentY += 10

	// Event details section
	pdf.SetFont("Arial", "", 12)
	pdf.SetTextColor(0, 0, 0)

	// Date and Time
	pdf.SetXY(marginX+5, currentY)
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(100, 100, 100)
	pdf.Cell(40, 6, "Date & Time:")
	pdf.SetFont("Arial", "", 12)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 6, data.EventDate.Format("Monday, January 2, 2006")+" at "+data.EventTime)
	currentY += 8

	// Venue
	pdf.SetXY(marginX+5, currentY)
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(100, 100, 100)
	pdf.Cell(40, 6, "Venue:")
	pdf.SetFont("Arial", "", 12)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 6, data.VenueName)
	currentY += 8

	// Venue Address
	pdf.SetXY(marginX+5, currentY)
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(100, 100, 100)
	pdf.Cell(40, 6, "Address:")
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(0, 0, 0)
	pdf.MultiCell(0, 6, data.VenueAddress, "", "L", false)
	currentY += 12

	// Divider line
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(marginX+5, currentY, pageWidth-marginX-5, currentY)
	currentY += 10

	// Attendee details section
	pdf.SetXY(marginX+5, currentY)
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(100, 100, 100)
	pdf.Cell(40, 6, "Attendee:")
	pdf.SetFont("Arial", "", 12)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 6, data.AttendeeName)
	currentY += 8

	// Email
	pdf.SetXY(marginX+5, currentY)
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(100, 100, 100)
	pdf.Cell(40, 6, "Email:")
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 6, data.AttendeeEmail)
	currentY += 8

	// Seat number (if provided)
	if data.SeatNumber != "" {
		pdf.SetXY(marginX+5, currentY)
		pdf.SetFont("Arial", "", 10)
		pdf.SetTextColor(100, 100, 100)
		pdf.Cell(40, 6, "Seat:")
		pdf.SetFont("Arial", "", 12)
		pdf.SetTextColor(0, 0, 0)
		pdf.Cell(0, 6, data.SeatNumber)
		currentY += 8
	}

	currentY += 5

	// Divider line
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(marginX+5, currentY, pageWidth-marginX-5, currentY)
	currentY += 10

	// QR Code section
	if len(data.QRCode) > 0 {

		qrOpt := gofpdf.ImageOptions{ImageType: "PNG"}
		pdf.RegisterImageOptionsReader("qrcode", qrOpt, bytes.NewReader(data.QRCode))

		qrSize := 60.0
		qrX := pageWidth - marginX - qrSize - 10
		qrY := currentY

		pdf.Image("qrcode", qrX, qrY, qrSize, qrSize, false, "", 0, "")

		pdf.SetXY(marginX+5, qrY+10)
		pdf.SetFont("Arial", "", 14)
		pdf.SetTextColor(79, 70, 229)
		pdf.Cell(0, 8, "Scan to Check In")

		pdf.SetXY(marginX+5, qrY+20)
		pdf.SetFont("Arial", "", 9)
		pdf.SetTextColor(100, 100, 100)
		pdf.MultiCell(qrX-marginX-15, 5, "Present this QR code at the venue entrance for quick check-in", "", "L", false)

		currentY = qrY + qrSize + 10
	}

	// Ticket number section
	currentY += 5
	pdf.SetXY(marginX+5, currentY)
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(100, 100, 100)
	pdf.Cell(40, 6, "Ticket Number:")
	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 6, data.TicketNumber)
	currentY += 8

	// Order number
	pdf.SetXY(marginX+5, currentY)
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(100, 100, 100)
	pdf.Cell(40, 6, "Order Number:")
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 6, data.OrderNumber)
	currentY += 8

	// Purchase date
	pdf.SetXY(marginX+5, currentY)
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(100, 100, 100)
	pdf.Cell(40, 6, "Purchased:")
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 6, data.PurchaseDate.Format("January 2, 2006"))
	currentY += 12

	// Special notes (if any)
	if data.SpecialNotes != "" {
		pdf.SetFillColor(255, 247, 237)
		pdf.SetDrawColor(251, 191, 36)
		pdf.Rect(marginX+5, currentY, pageWidth-(2*marginX)-10, 15, "FD")

		pdf.SetXY(marginX+10, currentY+3)
		pdf.SetFont("Arial", "", 9)
		pdf.SetTextColor(146, 64, 14)
		pdf.MultiCell(pageWidth-(2*marginX)-20, 4, data.SpecialNotes, "", "L", false)
		currentY += 18
	}

	// Footer
	currentY = pageHeight - marginY - 15
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(marginX+5, currentY, pageWidth-marginX-5, currentY)
	currentY += 5

	pdf.SetXY(marginX+5, currentY)
	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(150, 150, 150)
	pdf.Cell(0, 4, "This ticket is non-transferable and must be presented at the venue entrance.")
	currentY += 4

	pdf.SetXY(marginX+5, currentY)
	pdf.Cell(0, 4, "For support, contact: support@ticketing.com | Keep this ticket safe until after the event.")

	return nil
}

// GenerateToFile generates a PDF ticket and saves to file
func (g *TicketGenerator) GenerateToFile(data TicketData, filename string) error {
	bytes, err := g.Generate(data)
	if err != nil {
		return err
	}

	return saveBytesToFile(bytes, filename)
}

// Helper function to save bytes to file
func saveBytesToFile(data []byte, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}
