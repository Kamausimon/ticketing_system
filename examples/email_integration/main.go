package main

import (
	"fmt"
	"log"

	"ticketing_system/internal/config"
	"ticketing_system/internal/notifications"
)

// This is an example showing how to integrate the email system into your application

func main() {
	fmt.Println("🚀 Email System Integration Example")
	fmt.Println("=====================================")

	// Step 1: Load configuration
	fmt.Println("📝 Step 1: Loading configuration...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v\n", err)
	}
	fmt.Printf("✅ Configuration loaded (Provider: %s)\n\n", cfg.Email.Provider)

	// Step 2: Create notification service
	fmt.Println("📝 Step 2: Creating notification service...")
	notifService := notifications.NewNotificationService(cfg)
	fmt.Println("✅ Notification service created")

	// Step 3: Test the email configuration
	fmt.Println("📝 Step 3: Testing email configuration...")
	testEmail := "topstonewriters@gmail.com" // Change this to your email

	err = notifService.TestEmailConfiguration(testEmail)
	if err != nil {
		log.Printf("❌ Email test failed: %v\n", err)
		log.Println("\n⚠️  Make sure you have set up your .env file with valid credentials!")
		log.Println("    See .env.example for reference")
		return
	}
	fmt.Printf("✅ Test email sent to %s\n\n", testEmail)

	// Step 4: Send a welcome email
	fmt.Println("📝 Step 4: Sending welcome email...")
	err = notifService.SendWelcomeEmail("newuser@example.com", "John Doe")
	if err != nil {
		log.Printf("❌ Failed to send welcome email: %v\n", err)
	} else {
		fmt.Println("✅ Welcome email sent")
	}

	// Step 5: Send a verification email
	fmt.Println("📝 Step 5: Sending verification email...")
	err = notifService.SendVerificationEmail("user@example.com", "Jane Smith", "ABC123XYZ")
	if err != nil {
		log.Printf("❌ Failed to send verification email: %v\n", err)
	} else {
		fmt.Println("✅ Verification email sent")
	}

	// Step 6: Send a password reset email
	fmt.Println("📝 Step 6: Sending password reset email...")
	err = notifService.SendPasswordResetEmail("user@example.com", "Jane Smith", "reset_token_123")
	if err != nil {
		log.Printf("❌ Failed to send password reset email: %v\n", err)
	} else {
		fmt.Println("✅ Password reset email sent")
	}

	// Step 7: Send an order confirmation
	fmt.Println("📝 Step 7: Sending order confirmation...")
	orderData := notifications.OrderConfirmationData{
		CustomerName: "John Doe",
		OrderNumber:  "ORD-2024-001",
		EventName:    "Summer Music Festival 2024",
		EventDate:    "July 15, 2024",
		VenueName:    "Central Park Arena",
		Items: []notifications.OrderItem{
			{
				Name:     "VIP Ticket",
				Quantity: 2,
				Price:    150.00,
				Currency: "USD",
			},
			{
				Name:     "General Admission",
				Quantity: 3,
				Price:    50.00,
				Currency: "USD",
			},
		},
		Currency: "USD",
		Total:    450.00,
	}

	err = notifService.SendOrderConfirmationEmail("customer@example.com", orderData)
	if err != nil {
		log.Printf("❌ Failed to send order confirmation: %v\n", err)
	} else {
		fmt.Println("✅ Order confirmation sent")
	}

	// Step 8: Send a ticket generated email
	fmt.Println("📝 Step 8: Sending ticket email...")
	ticketData := notifications.TicketData{
		AttendeeName: "John Doe",
		EventName:    "Summer Music Festival 2024",
		EventDate:    "July 15, 2024",
		VenueName:    "Central Park Arena",
		TicketType:   "VIP",
		TicketNumber: "TKT-2024-VIP-001",
		QRCodeURL:    "https://example.com/qr/TKT-2024-VIP-001",
	}

	err = notifService.SendTicketGeneratedEmail("attendee@example.com", ticketData)
	if err != nil {
		log.Printf("❌ Failed to send ticket email: %v\n", err)
	} else {
		fmt.Println("✅ Ticket email sent")
	}

	fmt.Println("\n🎉 All done! Check your Mailtrap inbox to see the emails.")
	fmt.Println("📧 If you're using test mode, check the application logs instead.")
}

/*
INTEGRATION NOTES:
==================

1. In your auth handler (internal/auth/auth.go):

   import "ticketing_system/internal/notifications"

   type AuthHandler struct {
       db              *gorm.DB
       metrics         *analytics.PrometheusMetrics
       notifService    *notifications.NotificationService  // Add this
   }

   func NewAuthHandler(db *gorm.DB, metrics *analytics.PrometheusMetrics, notifService *notifications.NotificationService) *AuthHandler {
       return &AuthHandler{
           db:           db,
           metrics:      metrics,
           notifService: notifService,  // Initialize
       }
   }

   // Then in RegisterUser:
   func (h *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
       // ... existing registration logic ...

       // Send welcome email asynchronously
       go h.notifService.SendWelcomeEmail(user.Email, user.FirstName)

       // ... rest of handler ...
   }

2. In your main.go, update initialization:

   import (
       "ticketing_system/internal/config"
       "ticketing_system/internal/notifications"
   )

   func main() {
       DB := database.Init()

       // Load configuration
       cfg := config.LoadOrPanic()

       // Create notification service
       notifService := notifications.NewNotificationService(cfg)

       // Initialize metrics
       metrics := analytics.NewPrometheusMetrics()

       // Pass notification service to handlers
       authHandler := auth.NewAuthHandler(DB, metrics, notifService)
       orderHandler := orders.NewOrderHandler(DB, metrics, notifService)
       // ... etc

       // ... rest of main ...
   }

3. For password reset (internal/auth/auth.go):

   func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
       // ... generate reset token ...

       // Send reset email asynchronously
       go h.notifService.SendPasswordResetEmail(user.Email, user.FirstName, resetToken)

       // ... rest of handler ...
   }

4. For order confirmation (internal/orders/main.go):

   func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
       // ... create order logic ...

       // Prepare order data
       orderData := notifications.OrderConfirmationData{
           CustomerName: customer.Name,
           OrderNumber:  order.OrderNumber,
           EventName:    event.Name,
           EventDate:    event.StartDate.Format("January 2, 2006"),
           VenueName:    venue.Name,
           Items:        items,
           Currency:     order.Currency,
           Total:        order.TotalAmount,
       }

       // Send confirmation email asynchronously
       go h.notifService.SendOrderConfirmationEmail(customer.Email, orderData)

       // ... rest of handler ...
   }

IMPORTANT NOTES:
================

- Always send emails asynchronously using goroutines (go keyword)
- Don't block HTTP responses waiting for email delivery
- Log email errors but don't fail the main operation
- Use Mailtrap for development/testing
- Switch to Zoho or another provider for production
- Set EMAIL_TEST_MODE=true for unit tests

*/
