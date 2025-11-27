# PDF Ticket Generator with QR Code

Complete PDF ticket generation system with scannable QR codes for event ticketing.

## ✨ Features

- ✅ **Professional PDF Tickets** - Beautiful, branded ticket design
- ✅ **QR Code Integration** - Scannable codes for quick check-in
- ✅ **Customizable** - Colors, logos, and layouts
- ✅ **Complete Information** - Event details, attendee info, seat numbers
- ✅ **Secure** - Unique ticket numbers and QR codes
- ✅ **Easy to Use** - Simple API for ticket generation

## 📦 Components

### 1. QR Code Generator (`pkg/qrcode/`)
Generates QR codes with various customization options:
- Custom sizes
- Error correction levels
- Custom colors
- Logo overlay support

### 2. PDF Ticket Generator (`pkg/pdf/`)
Creates professional PDF tickets with:
- Event details
- Attendee information
- Scannable QR code
- Ticket and order numbers
- Custom branding

## 🚀 Quick Start

### Generate a Simple Ticket

```go
package main

import (
    "fmt"
    "time"
    "ticketing_system/pkg/pdf"
    "ticketing_system/pkg/qrcode"
)

func main() {
    // Prepare ticket data
    ticketData := pdf.TicketData{
        TicketNumber:  "TKT-2024-001",
        EventName:     "Summer Concert 2024",
        EventDate:     time.Date(2024, 7, 15, 19, 0, 0, 0, time.UTC),
        EventTime:     "7:00 PM",
        VenueName:     "City Arena",
        VenueAddress:  "123 Main St, City, State",
        AttendeeName:  "John Doe",
        AttendeeEmail: "john@example.com",
        TicketType:    "VIP",
        SeatNumber:    "A-15",
        OrderNumber:   "ORD-2024-001",
        PurchaseDate:  time.Now(),
    }

    // Generate QR code
    qrContent := fmt.Sprintf("TICKET:%s", ticketData.TicketNumber)
    qrBytes, _ := qrcode.Generate(qrContent)
    ticketData.QRCode = qrBytes

    // Generate PDF
    generator := pdf.NewTicketGenerator()
    generator.GenerateToFile(ticketData, "ticket.pdf")
}
```

### Generate QR Code Only

```go
// Simple QR code
qrBytes, err := qrcode.Generate("TICKET-12345")

// Custom size QR code
qrBytes, err := qrcode.GenerateCustom("TICKET-12345", 512)

// Advanced QR code with customization
generator := qrcode.NewGenerator().
    WithSize(512).
    WithRecoveryLevel(qrcode.High).
    WithColors(color.Black, color.White)

qrBytes, err := generator.GenerateBytes("TICKET-12345")
```

## 📖 API Reference

### QR Code Generator

#### `NewGenerator() *Generator`
Creates a new QR code generator with default settings.

#### `WithSize(size int) *Generator`
Sets the QR code size in pixels (default: 256).

#### `WithRecoveryLevel(level qrcode.RecoveryLevel) *Generator`
Sets error correction level:
- `qrcode.Low` - 7% recovery
- `qrcode.Medium` - 15% recovery (default)
- `qrcode.High` - 25% recovery
- `qrcode.Highest` - 30% recovery

#### `WithColors(fg, bg color.Color) *Generator`
Sets custom foreground and background colors.

#### `GenerateBytes(content string) ([]byte, error)`
Generates QR code and returns PNG bytes.

#### `GenerateImage(content string) (image.Image, error)`
Generates QR code and returns as image.Image.

#### `GenerateFile(content, filename string) error`
Generates QR code and saves to file.

### PDF Ticket Generator

#### `NewTicketGenerator() *TicketGenerator`
Creates a new PDF ticket generator.

#### `WithLogo(logoPath string) *TicketGenerator`
Sets a logo to display on the ticket.

#### `WithColors(primary, secondary string) *TicketGenerator`
Sets custom brand colors (hex format).

#### `Generate(data TicketData) ([]byte, error)`
Generates PDF ticket and returns bytes.

#### `GenerateToFile(data TicketData, filename string) error`
Generates PDF ticket and saves to file.

### TicketData Structure

```go
type TicketData struct {
    TicketNumber  string    // Unique ticket identifier
    EventName     string    // Name of the event
    EventDate     time.Time // Event date
    EventTime     string    // Event time (formatted)
    VenueName     string    // Venue name
    VenueAddress  string    // Full venue address
    AttendeeName  string    // Attendee's full name
    AttendeeEmail string    // Attendee's email
    TicketType    string    // e.g., "VIP", "General", "Early Bird"
    SeatNumber    string    // Seat assignment (optional)
    Price         float64   // Ticket price
    Currency      string    // Currency code (e.g., "USD")
    QRCode        []byte    // QR code image bytes
    OrderNumber   string    // Associated order number
    PurchaseDate  time.Time // When ticket was purchased
    SpecialNotes  string    // Special instructions (optional)
}
```

## 🎨 Customization

### Custom Branding

```go
generator := pdf.NewTicketGenerator().
    WithLogo("/path/to/logo.png").
    WithColors("#4F46E5", "#10B981")

pdfBytes, err := generator.Generate(ticketData)
```

### QR Code Content Format

Recommended format for QR code content:

```go
qrContent := fmt.Sprintf(
    "TICKET:%s|EVENT:%s|ATTENDEE:%s|DATE:%s",
    ticketNumber,
    eventName,
    attendeeName,
    eventDate.Format("2006-01-02"),
)
```

## 💡 Integration Examples

### With Email System

```go
// Generate ticket
ticketBytes, err := generator.Generate(ticketData)
if err != nil {
    log.Fatal(err)
}

// Send via email
emailData := notifications.EmailData{
    To:       []string{attendee.Email},
    Subject:  "Your Ticket for " + event.Name,
    HTMLBody: "Please find your ticket attached",
    Attachments: []notifications.Attachment{
        {
            Filename: "ticket.pdf",
            Content:  ticketBytes,
            MimeType: "application/pdf",
        },
    },
}

notificationService.Send(emailData)
```

### With HTTP Handler

```go
func (h *TicketHandler) DownloadTicket(w http.ResponseWriter, r *http.Request) {
    // Get ticket data from database
    ticket := getTicketFromDB(ticketID)
    
    // Generate PDF
    pdfBytes, err := generateTicketPDF(ticket)
    if err != nil {
        http.Error(w, "Failed to generate ticket", 500)
        return
    }
    
    // Send as download
    w.Header().Set("Content-Type", "application/pdf")
    w.Header().Set("Content-Disposition", 
        fmt.Sprintf("attachment; filename=ticket_%s.pdf", ticket.Number))
    w.Write(pdfBytes)
}
```

### Batch Generation

```go
func GenerateTicketsForOrder(orderID string) error {
    tickets := getTicketsForOrder(orderID)
    
    for _, ticket := range tickets {
        // Generate QR code
        qrBytes, _ := qrcode.Generate(ticket.Number)
        
        // Prepare ticket data
        ticketData := prepareTicketData(ticket)
        ticketData.QRCode = qrBytes
        
        // Generate PDF
        filename := fmt.Sprintf("ticket_%s.pdf", ticket.Number)
        err := generator.GenerateToFile(ticketData, filename)
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

## 🔒 Security Best Practices

1. **Unique Ticket Numbers**
   ```go
   ticketNumber := fmt.Sprintf("TKT-%s-%d", 
       time.Now().Format("20060102"), 
       rand.Int63())
   ```

2. **Secure QR Content**
   ```go
   // Include verification data in QR code
   qrContent := fmt.Sprintf("VERIFY:%s|HASH:%s", 
       ticketNumber, 
       generateHash(ticketNumber, secret))
   ```

3. **Expiration Checks**
   ```go
   if ticket.EventDate.Before(time.Now()) {
       return errors.New("ticket expired")
   }
   ```

## 📊 Performance Tips

1. **Generate QR codes once and cache**
   ```go
   // Store QR bytes in database with ticket
   ticket.QRCodeData = qrBytes
   ```

2. **Use goroutines for batch generation**
   ```go
   for _, ticket := range tickets {
       go func(t Ticket) {
           generateTicketPDF(t)
       }(ticket)
   }
   ```

3. **Optimize QR code size**
   ```go
   // Use 256x256 for most cases (good balance)
   qrGenerator.WithSize(256)
   ```

## 🧪 Testing

Run the example:

```bash
go run examples/ticket_pdf/main.go
```

This will generate a sample ticket PDF in the current directory.

## 📝 Dependencies

```bash
go get github.com/jung-kurt/gofpdf
go get github.com/skip2/go-qrcode
```

Both packages are lightweight and well-maintained.

## 🎯 Next Steps

1. ✅ PDF generation with QR codes - **DONE**
2. ⏳ Integrate with ticket generation handler
3. ⏳ Add ticket delivery via email
4. ⏳ Implement QR code scanning for check-in
5. ⏳ Add batch ticket generation
6. ⏳ Create ticket templates system

## 📚 Additional Resources

- [gofpdf Documentation](https://github.com/jung-kurt/gofpdf)
- [go-qrcode Documentation](https://github.com/skip2/go-qrcode)
- [QR Code Spec](https://www.qrcode.com/en/about/standards.html)

## 🤝 Usage in Ticketing System

The PDF and QR code generators are now ready to be integrated into the main ticketing system. They can be used in:

- `/tickets/generate` endpoint
- `/tickets/{id}/pdf` endpoint  
- Email ticket delivery
- Batch ticket processing
- Check-in validation system

See `cmd/ticket-generator/` for the command-line tool implementation.
