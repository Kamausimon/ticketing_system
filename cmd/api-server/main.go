package main

import (
	"fmt"
	"net/http"
	"os"
	"ticketing_system/internal/accounts"
	"ticketing_system/internal/analytics"
	"ticketing_system/internal/attendees"
	"ticketing_system/internal/auth"
	"ticketing_system/internal/config"
	"ticketing_system/internal/database"
	"ticketing_system/internal/events"
	"ticketing_system/internal/inventory"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
	"ticketing_system/internal/notifications"
	"ticketing_system/internal/orders"
	"ticketing_system/internal/organizers"
	"ticketing_system/internal/payments"
	"ticketing_system/internal/promotions"
	"ticketing_system/internal/refunds"
	"ticketing_system/internal/security"
	"ticketing_system/internal/settlement"
	"ticketing_system/internal/tickets"
	"ticketing_system/internal/venues"
	"ticketing_system/pkg/ratelimit"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	DB := database.Init()

	err := DB.AutoMigrate(&models.User{}, &models.EmailVerification{}, &models.WaitlistEntry{}, &models.TicketTransferHistory{})
	if err != nil {
		fmt.Printf("Migration failed: %v\n", err)
	} else {
		fmt.Println("✅ Database migration completed successfully")
	}

	// Initialize Prometheus metrics
	fmt.Println("🔧 Initializing Prometheus metrics...")
	metrics := analytics.NewPrometheusMetrics()

	// Start system metrics collector
	analytics.StartSystemMetricsCollector(metrics, DB)
	fmt.Println("✅ System metrics collector started")

	// Load configuration for notifications and security
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to load config: %v\n", err)
		fmt.Println("⚠️  Notification and encryption services will not be available")
	}
	var notificationHandler *notifications.Handler
	var notificationService *notifications.NotificationService
	if cfg != nil {
		notificationService = notifications.NewNotificationService(cfg)
		notificationHandler = notifications.NewHandler(notificationService)
		fmt.Println("✅ Notification service initialized")
	}

	// Initialize auth handler with notification service
	var authHandler *auth.AuthHandler
	if notificationService != nil {
		authHandler = auth.NewAuthHandlerWithNotifications(DB, metrics, notificationService)
	} else {
		authHandler = auth.NewAuthHandler(DB, metrics)
	}

	// Initialize encryption service for sensitive data
	var encryptionService *security.EncryptionService
	if cfg != nil {
		encryptionService, err = security.NewEncryptionService(cfg.Security.EncryptionKey)
		if err != nil {
			fmt.Printf("⚠️  Warning: Failed to initialize encryption service: %v\n", err)
			fmt.Println("⚠️  Bank details encryption will not be available")
		} else {
			fmt.Println("✅ Encryption service initialized")
		}
	}

	// Initialize 2FA handler
	twoFactorHandler := auth.NewTwoFactorHandler(DB, "Ticketing System")
	if notificationService != nil {
		twoFactorHandler.SetEmailService(notificationService.GetEmailService())
	}

	var organizerHandler *organizers.OrganizerHandler
	if notificationService != nil {
		organizerHandler = organizers.NewOrganizerHandler(DB, metrics, notificationService, encryptionService)
	} else {
		organizerHandler = organizers.NewOrganizerHandler(DB, metrics, nil, encryptionService)
	}
	eventHandler := events.NewEventHandler(DB, metrics)
	accountHandler := accounts.NewAccountHandler(DB, metrics)
	orderHandler := orders.NewOrderHandler(DB, metrics)
	var ticketHandler *tickets.TicketHandler
	if cfg != nil && cfg.Email.Host != "" {
		notificationService := notifications.NewNotificationService(cfg)
		ticketHandler = tickets.NewTicketHandlerWithNotifications(DB, metrics, notificationService)
	} else {
		ticketHandler = tickets.NewTicketHandler(DB, metrics)
	}
	promotionHandler := promotions.NewPromotionHandler(DB, metrics)
	inventoryHandler := inventory.NewInventoryHandler(DB, metrics)
	paymentHandler := payments.NewPaymentHandler(DB, metrics)
	refundHandler := refunds.NewRefundHandler(DB, metrics, notificationService, paymentHandler.IntasendSecretKey, paymentHandler.IntasendWebhookSecret, paymentHandler.IntasendTestMode)
	settlementService := settlement.NewService(DB)
	settlementHandler := settlement.NewSettlementHandler(settlementService)
	attendeeHandler := attendees.NewAttendeeHandler(DB, metrics)
	venueHandler := venues.NewVenueHandler(DB, metrics)
	router := mux.NewRouter()

	// Initialize rate limiters for different endpoint categories
	gov := ratelimit.NewTokenBucketGovernor()

	// Configure rate limiters
	gov.GetOrCreate("auth", ratelimit.Presets.Auth)         // 10 req/min per IP
	gov.GetOrCreate("login", ratelimit.Presets.Login)       // 5 attempts/min per IP (strict login protection)
	gov.GetOrCreate("payment", ratelimit.Presets.Payment)   // 5 req/min per IP
	gov.GetOrCreate("api", ratelimit.Presets.API)           // 100 req/s per IP
	gov.GetOrCreate("download", ratelimit.Presets.Download) // 3 req/s per user
	gov.GetOrCreate("inventory", ratelimit.Config{
		RequestsPerSecond: 50,
		BurstSize:         100,
		CleanupInterval:   5 * 60 * 60,
	})

	// Create rate limiting middleware wrappers
	authLimiter := ratelimit.NewMiddleware(gov.Get("auth"), ratelimit.KeyFuncs.ByIP)
	loginLimiter := ratelimit.NewMiddleware(gov.Get("login"), ratelimit.KeyFuncs.ByIP)
	paymentLimiter := ratelimit.NewMiddleware(gov.Get("payment"), ratelimit.KeyFuncs.ByIP)
	apiLimiter := ratelimit.NewMiddleware(gov.Get("api"), ratelimit.KeyFuncs.ByIP)
	downloadLimiter := ratelimit.NewMiddleware(gov.Get("download"), ratelimit.KeyFuncs.ByIP)
	inventoryLimiter := ratelimit.NewMiddleware(gov.Get("inventory"), ratelimit.KeyFuncs.ByIP)

	// Create email verification middleware
	emailVerificationMiddleware := middleware.RequireEmailVerification(DB)

	// Add Prometheus middleware
	router.Use(analytics.PrometheusMiddleware(metrics))

	// Expose Prometheus metrics endpoint
	router.Handle("/metrics", promhttp.Handler())

	//auth routes - with rate limiting
	router.HandleFunc("/register", authLimiter.HandlerFunc(authHandler.RegisterUser)).Methods(http.MethodPost)
	router.HandleFunc("/login", loginLimiter.HandlerFunc(authHandler.LoginUser)).Methods(http.MethodPost)
	router.HandleFunc("/logout", authLimiter.HandlerFunc(authHandler.LogoutUser)).Methods(http.MethodPost)
	router.HandleFunc("/forgot-password", authLimiter.HandlerFunc(authHandler.ForgotPassword)).Methods(http.MethodPost)
	router.HandleFunc("/resetPassword", authLimiter.HandlerFunc(authHandler.ResetPassword)).Methods(http.MethodPost)

	// Email verification routes
	router.HandleFunc("/verify-email", authLimiter.HandlerFunc(authHandler.VerifyEmail)).Methods(http.MethodPost)
	router.HandleFunc("/resend-verification", authLimiter.HandlerFunc(authHandler.ResendVerification)).Methods(http.MethodPost)
	router.HandleFunc("/verify-email/status", authHandler.CheckEmailVerificationStatus).Methods(http.MethodGet)

	// Two-Factor Authentication routes - with rate limiting
	router.HandleFunc("/2fa/setup", authLimiter.HandlerFunc(twoFactorHandler.Setup2FA)).Methods(http.MethodPost)
	router.HandleFunc("/2fa/verify-setup", authLimiter.HandlerFunc(twoFactorHandler.VerifySetup)).Methods(http.MethodPost)
	router.HandleFunc("/2fa/verify-login", loginLimiter.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := godotenv.Load(".env")
		if err != nil {
			middleware.WriteJSONError(w, http.StatusInternalServerError, "error loading env variables")
			return
		}
		tokenSecret := os.Getenv("JWTSECRET")
		twoFactorHandler.VerifyLogin(w, r, tokenSecret)
	})).Methods(http.MethodPost)
	router.HandleFunc("/2fa/disable", authLimiter.HandlerFunc(twoFactorHandler.Disable2FA)).Methods(http.MethodPost)
	router.HandleFunc("/2fa/status", twoFactorHandler.GetStatus).Methods(http.MethodGet)
	router.HandleFunc("/2fa/recovery-codes", authLimiter.HandlerFunc(twoFactorHandler.RegenerateRecoveryCodes)).Methods(http.MethodPost)
	router.HandleFunc("/2fa/attempts", twoFactorHandler.GetRecentAttempts).Methods(http.MethodGet)

	//organizer routes
	router.HandleFunc("/organizers/apply", organizerHandler.OrganizerApply).Methods(http.MethodPost)
	router.HandleFunc("/organizers/profile", organizerHandler.GetOrganizerProfile).Methods(http.MethodGet)
	router.HandleFunc("/organizers/profile", organizerHandler.UpdateOrganizerProfile).Methods(http.MethodPut)
	router.HandleFunc("/organizers/onboarding/status", organizerHandler.GetOnboardingStatus).Methods(http.MethodGet)
	router.HandleFunc("/organizers/dashboard", organizerHandler.GetOrganizerDashboard).Methods(http.MethodGet)
	router.HandleFunc("/organizers/dashboard/stats", organizerHandler.GetQuickStats).Methods(http.MethodGet)
	router.HandleFunc("/organizers/logo", organizerHandler.UploadOrganizerLogo).Methods(http.MethodPost)
	router.HandleFunc("/organizers/verification/email", organizerHandler.SendVerificationEmail).Methods(http.MethodPost)

	// Organizer routes - Bank Details (encrypted)
	router.HandleFunc("/organizers/bank-details", organizerHandler.UpdateBankDetails).Methods(http.MethodPut)
	router.HandleFunc("/organizers/bank-details", organizerHandler.GetBankDetails).Methods(http.MethodGet)

	// Organizer routes - Payment Gateway Configuration
	router.HandleFunc("/organizers/payment-gateway", organizerHandler.ConfigurePaymentGateway).Methods(http.MethodPost)
	router.HandleFunc("/organizers/payment-gateway", organizerHandler.GetPaymentGatewayConfig).Methods(http.MethodGet)

	// Admin organizer routes
	router.HandleFunc("/admin/organizers/pending", organizerHandler.GetPendingOrganizers).Methods(http.MethodGet)
	router.HandleFunc("/admin/organizers/{id}/verify", organizerHandler.VerifyOrganizer).Methods(http.MethodPost)

	// Event routes - Public
	router.HandleFunc("/events", eventHandler.ListEvents).Methods(http.MethodGet)
	router.HandleFunc("/events/search", eventHandler.SearchEvents).Methods(http.MethodGet)
	router.HandleFunc("/events/{id}", eventHandler.GetEventDetails).Methods(http.MethodGet)
	router.HandleFunc("/events/{id}/images", eventHandler.GetEventImages).Methods(http.MethodGet)

	// Event routes - Organizer only
	router.HandleFunc("/organizers/events", eventHandler.ListOrganizerEvents).Methods(http.MethodGet)
	router.HandleFunc("/organizers/events/search", eventHandler.SearchOrganizerEvents).Methods(http.MethodGet)
	router.HandleFunc("/organizers/events", eventHandler.CreateEvent).Methods(http.MethodPost)
	router.HandleFunc("/organizers/events/{id}", eventHandler.UpdateEvent).Methods(http.MethodPut)
	router.HandleFunc("/organizers/events/{id}", eventHandler.DeleteEvent).Methods(http.MethodDelete)
	router.HandleFunc("/organizers/events/{id}/publish", eventHandler.PublishEvent).Methods(http.MethodPost)
	router.HandleFunc("/organizers/events/{id}/images", eventHandler.UploadEventImage).Methods(http.MethodPost)
	router.HandleFunc("/organizers/events/{id}/images/{imageId}", eventHandler.DeleteEventImage).Methods(http.MethodDelete)

	// Account routes - Profile
	router.HandleFunc("/account/profile", accountHandler.GetAccountProfile).Methods(http.MethodGet)
	router.HandleFunc("/account/profile", accountHandler.UpdateAccountProfile).Methods(http.MethodPut)
	router.HandleFunc("/account", accountHandler.DeleteAccount).Methods(http.MethodDelete)

	// Account routes - Address
	router.HandleFunc("/account/address", accountHandler.GetAccountAddress).Methods(http.MethodGet)
	router.HandleFunc("/account/address", accountHandler.UpdateAccountAddress).Methods(http.MethodPut)
	router.HandleFunc("/account/address", accountHandler.ClearAccountAddress).Methods(http.MethodDelete)
	router.HandleFunc("/account/countries", accountHandler.GetSupportedCountries).Methods(http.MethodGet)

	// Account routes - Settings & Preferences
	router.HandleFunc("/account/preferences", accountHandler.GetAccountPreferences).Methods(http.MethodGet)
	router.HandleFunc("/account/preferences", accountHandler.UpdateAccountPreferences).Methods(http.MethodPut)
	router.HandleFunc("/account/settings", accountHandler.GetAccountSettings).Methods(http.MethodGet)
	router.HandleFunc("/account/settings", accountHandler.UpdateAccountSettings).Methods(http.MethodPut)
	router.HandleFunc("/account/timezones", accountHandler.GetAvailableTimezones).Methods(http.MethodGet)
	router.HandleFunc("/account/currencies", accountHandler.GetAvailableCurrencies).Methods(http.MethodGet)
	router.HandleFunc("/account/date-formats", accountHandler.GetDateFormats).Methods(http.MethodGet)

	// Account routes - Security
	router.HandleFunc("/account/security", accountHandler.GetSecuritySettings).Methods(http.MethodGet)
	router.HandleFunc("/account/security/password", accountHandler.ChangePassword).Methods(http.MethodPost)
	router.HandleFunc("/account/security/login-history", accountHandler.GetLoginHistory).Methods(http.MethodGet)
	router.HandleFunc("/account/security/lock", accountHandler.LockAccount).Methods(http.MethodPost)
	router.HandleFunc("/account/security/unlock", accountHandler.UnlockAccount).Methods(http.MethodPost)

	// Account routes - Activity
	router.HandleFunc("/account/activity", accountHandler.GetAccountActivity).Methods(http.MethodGet)
	router.HandleFunc("/account/activity/types", accountHandler.GetActivityTypes).Methods(http.MethodGet)
	router.HandleFunc("/account/activity/log", accountHandler.LogActivity).Methods(http.MethodPost)
	router.HandleFunc("/account/activity/clear", accountHandler.ClearActivityLog).Methods(http.MethodDelete)
	router.HandleFunc("/account/stats", accountHandler.GetAccountStats).Methods(http.MethodGet)

	// Account routes - Payment Methods
	router.HandleFunc("/account/payment-methods", accountHandler.GetPaymentMethods).Methods(http.MethodGet)
	router.HandleFunc("/account/payment-gateway", accountHandler.GetPaymentGatewaySettings).Methods(http.MethodGet)
	router.HandleFunc("/account/payment-gateway/info", accountHandler.GetPaymentGatewayInfo).Methods(http.MethodGet)

	// Account routes - Stripe Integration (Organizers only)
	router.HandleFunc("/account/stripe/setup", accountHandler.SetupStripeIntegration).Methods(http.MethodPost)
	router.HandleFunc("/account/stripe/connect", accountHandler.SetupStripeConnect).Methods(http.MethodPost)
	router.HandleFunc("/account/stripe/complete", accountHandler.CompleteStripeSetup).Methods(http.MethodPost)
	router.HandleFunc("/account/stripe/disconnect", accountHandler.DisconnectStripe).Methods(http.MethodDelete)

	// Order routes - Creation & Calculation - with rate limiting
	router.HandleFunc("/orders", apiLimiter.HandlerFunc(orderHandler.CreateOrder)).Methods(http.MethodPost)
	router.HandleFunc("/orders/calculate", apiLimiter.HandlerFunc(orderHandler.CalculateOrder)).Methods(http.MethodPost)

	// Order routes - Viewing
	router.HandleFunc("/orders", orderHandler.ListOrders).Methods(http.MethodGet)
	router.HandleFunc("/orders/search", orderHandler.SearchOrders).Methods(http.MethodGet)
	router.HandleFunc("/orders/{id}", orderHandler.GetOrderDetails).Methods(http.MethodGet)
	router.HandleFunc("/orders/{id}/summary", orderHandler.GetOrderSummary).Methods(http.MethodGet)
	router.HandleFunc("/orders/stats", orderHandler.GetOrderStats).Methods(http.MethodGet)

	// Order routes - Management - with rate limiting
	router.HandleFunc("/orders/{id}/status", paymentLimiter.HandlerFunc(orderHandler.UpdateOrderStatus)).Methods(http.MethodPut)
	router.HandleFunc("/orders/{id}/cancel", paymentLimiter.HandlerFunc(orderHandler.CancelOrder)).Methods(http.MethodPost)
	router.HandleFunc("/orders/{id}/refund", paymentLimiter.HandlerFunc(orderHandler.RefundOrder)).Methods(http.MethodPost)

	// Order routes - Payment - with rate limiting
	router.HandleFunc("/orders/{id}/payment", paymentLimiter.HandlerFunc(orderHandler.ProcessPayment)).Methods(http.MethodPost)
	router.HandleFunc("/orders/{id}/payment/verify", paymentLimiter.HandlerFunc(orderHandler.VerifyPayment)).Methods(http.MethodPost)

	// Order routes - Organizer view
	router.HandleFunc("/organizers/orders", orderHandler.ListOrganizerOrders).Methods(http.MethodGet)
	router.HandleFunc("/organizers/orders/search", orderHandler.SearchOrganizerOrders).Methods(http.MethodGet)

	// Ticket routes - Generation
	router.Handle("/tickets/generate", emailVerificationMiddleware(http.HandlerFunc(ticketHandler.GenerateTickets))).Methods(http.MethodPost)
	router.Handle("/tickets/regenerate-qr", emailVerificationMiddleware(http.HandlerFunc(ticketHandler.RegenerateTicketQR))).Methods(http.MethodPost)

	// Ticket routes - Viewing
	router.HandleFunc("/tickets", ticketHandler.ListUserTickets).Methods(http.MethodGet)
	router.HandleFunc("/tickets/{id}", ticketHandler.GetTicketDetails).Methods(http.MethodGet)
	router.HandleFunc("/tickets/number", ticketHandler.GetTicketByNumber).Methods(http.MethodGet)
	router.HandleFunc("/tickets/stats", ticketHandler.GetTicketStats).Methods(http.MethodGet)

	// Ticket routes - PDF Download - with rate limiting and email verification
	router.Handle("/tickets/{id}/pdf", emailVerificationMiddleware(downloadLimiter.HandlerFunc(ticketHandler.DownloadTicketPDF))).Methods(http.MethodGet)

	// Ticket routes - Transfer - with rate limiting and email verification
	router.Handle("/tickets/{id}/transfer", emailVerificationMiddleware(paymentLimiter.HandlerFunc(ticketHandler.TransferTicket))).Methods(http.MethodPost)
	router.HandleFunc("/tickets/{id}/transfer-history", apiLimiter.HandlerFunc(ticketHandler.GetTransferHistory)).Methods(http.MethodGet)

	// Ticket routes - Validation (Organizer only)
	router.HandleFunc("/tickets/validate", ticketHandler.ValidateTicket).Methods(http.MethodPost)
	router.HandleFunc("/tickets/validate/qr", ticketHandler.ValidateTicketByQR).Methods(http.MethodPost)

	// Ticket routes - Check-in (Organizer only)
	router.HandleFunc("/tickets/checkin", ticketHandler.CheckInTicket).Methods(http.MethodPost)
	router.HandleFunc("/tickets/checkin/bulk", ticketHandler.BulkCheckIn).Methods(http.MethodPost)
	router.HandleFunc("/tickets/checkin/undo", ticketHandler.UndoCheckIn).Methods(http.MethodPost)
	router.HandleFunc("/tickets/checkin/stats", ticketHandler.GetCheckInStats).Methods(http.MethodGet)

	// Ticket routes - Bulk Operations
	router.HandleFunc("/tickets/bulk/export", ticketHandler.BulkExportTickets).Methods(http.MethodPost)
	router.HandleFunc("/tickets/bulk/stats", ticketHandler.GetBulkTicketStats).Methods(http.MethodGet)
	router.HandleFunc("/tickets/bulk/status", ticketHandler.BulkUpdateTicketStatus).Methods(http.MethodPost)

	// Ticket routes - Event tickets (Organizer only)
	router.HandleFunc("/organizers/tickets", ticketHandler.ListEventTickets).Methods(http.MethodGet)
	router.HandleFunc("/organizers/tickets/filter", ticketHandler.FilterEventTicketsAdvanced).Methods(http.MethodGet)
	router.HandleFunc("/organizers/tickets/search", ticketHandler.SearchEventTickets).Methods(http.MethodGet)

	// Promotion routes - Creation & Management
	router.HandleFunc("/promotions", promotionHandler.CreatePromotion).Methods(http.MethodPost)
	router.HandleFunc("/promotions/{id}", promotionHandler.GetPromotionDetails).Methods(http.MethodGet)
	router.HandleFunc("/promotions/code/{code}", promotionHandler.GetPromotionByCode).Methods(http.MethodGet)
	router.HandleFunc("/promotions/{id}", promotionHandler.UpdatePromotion).Methods(http.MethodPut)
	router.HandleFunc("/promotions/{id}", promotionHandler.DeletePromotion).Methods(http.MethodDelete)
	router.HandleFunc("/promotions/{id}/clone", promotionHandler.ClonePromotion).Methods(http.MethodPost)

	// Promotion routes - Status Management
	router.HandleFunc("/promotions/{id}/activate", promotionHandler.ActivatePromotion).Methods(http.MethodPost)
	router.HandleFunc("/promotions/{id}/pause", promotionHandler.PausePromotion).Methods(http.MethodPost)
	router.HandleFunc("/promotions/{id}/deactivate", promotionHandler.DeactivatePromotion).Methods(http.MethodPost)
	router.HandleFunc("/promotions/{id}/extend", promotionHandler.ExtendPromotionDate).Methods(http.MethodPost)

	// Promotion routes - Listing & Search
	router.HandleFunc("/promotions", promotionHandler.ListPromotions).Methods(http.MethodGet)
	router.HandleFunc("/promotions/active", promotionHandler.ListActivePromotions).Methods(http.MethodGet)
	router.HandleFunc("/promotions/search", promotionHandler.SearchPromotions).Methods(http.MethodGet)

	// Promotion routes - Validation & Usage
	router.HandleFunc("/promotions/validate", promotionHandler.ValidatePromotionCode).Methods(http.MethodPost)
	router.HandleFunc("/promotions/eligibility", promotionHandler.CheckPromotionEligibility).Methods(http.MethodPost)
	router.HandleFunc("/promotions/{id}/usage", promotionHandler.GetPromotionUsage).Methods(http.MethodGet)
	router.HandleFunc("/promotions/{id}/usage", promotionHandler.RecordPromotionUsage).Methods(http.MethodPost)
	router.HandleFunc("/promotions/usage/revoke", promotionHandler.RevokePromotionUsage).Methods(http.MethodPost)
	router.HandleFunc("/promotions/{id}/usage/details", promotionHandler.GetPromotionUsageDetails).Methods(http.MethodGet)

	// Promotion routes - Analytics
	router.HandleFunc("/promotions/{id}/stats", promotionHandler.GetPromotionStats).Methods(http.MethodGet)
	router.HandleFunc("/promotions/{id}/analytics", promotionHandler.GetPromotionAnalytics).Methods(http.MethodGet)
	router.HandleFunc("/promotions/{id}/roi", promotionHandler.GetPromotionROI).Methods(http.MethodGet)
	router.HandleFunc("/promotions/{id}/conversion", promotionHandler.GetConversionMetrics).Methods(http.MethodGet)
	router.HandleFunc("/promotions/{id}/revenue", promotionHandler.GetRevenueImpact).Methods(http.MethodGet)

	// Promotion routes - Organizer
	router.HandleFunc("/organizers/promotions", promotionHandler.ListOrganizerPromotions).Methods(http.MethodGet)
	router.HandleFunc("/organizers/promotions/stats", promotionHandler.GetOrganizerPromotionStats).Methods(http.MethodGet)

	// Inventory routes - Availability - with rate limiting
	router.HandleFunc("/inventory/tickets/{id}", inventoryLimiter.HandlerFunc(inventoryHandler.GetTicketAvailability)).Methods(http.MethodGet)
	router.HandleFunc("/inventory/events/{id}", inventoryLimiter.HandlerFunc(inventoryHandler.GetEventInventory)).Methods(http.MethodGet)
	router.HandleFunc("/inventory/status/{id}", inventoryLimiter.HandlerFunc(inventoryHandler.GetInventoryStatus)).Methods(http.MethodGet)
	router.HandleFunc("/inventory/bulk-check", inventoryLimiter.HandlerFunc(inventoryHandler.BulkCheckAvailability)).Methods(http.MethodPost)

	// Inventory routes - Capacity Management - with rate limiting
	router.HandleFunc("/inventory/capacity/tickets/{id}", inventoryLimiter.HandlerFunc(inventoryHandler.GetTicketClassCapacity)).Methods(http.MethodGet)
	router.HandleFunc("/inventory/capacity/events/{id}", inventoryLimiter.HandlerFunc(inventoryHandler.GetEventCapacity)).Methods(http.MethodGet)
	router.HandleFunc("/inventory/capacity/events/{id}/monitor", apiLimiter.HandlerFunc(inventoryHandler.MonitorCapacity)).Methods(http.MethodGet)

	// Inventory routes - Waitlist - with rate limiting
	router.HandleFunc("/inventory/waitlist", inventoryLimiter.HandlerFunc(inventoryHandler.JoinWaitlist)).Methods(http.MethodPost)
	router.HandleFunc("/inventory/waitlist/{id}", apiLimiter.HandlerFunc(inventoryHandler.GetWaitlistPosition)).Methods(http.MethodGet)
	router.HandleFunc("/inventory/waitlist", apiLimiter.HandlerFunc(inventoryHandler.ListUserWaitlist)).Methods(http.MethodGet)
	router.HandleFunc("/inventory/waitlist/{id}/leave", inventoryLimiter.HandlerFunc(inventoryHandler.LeaveWaitlist)).Methods(http.MethodDelete)
	router.HandleFunc("/inventory/waitlist/events/{id}/stats", apiLimiter.HandlerFunc(inventoryHandler.GetWaitlistStats)).Methods(http.MethodGet)
	router.HandleFunc("/inventory/waitlist/notify", paymentLimiter.HandlerFunc(inventoryHandler.NotifyNextInWaitlist)).Methods(http.MethodPost)

	// Inventory routes - Reservations - with rate limiting
	router.HandleFunc("/inventory/reservations", inventoryLimiter.HandlerFunc(inventoryHandler.CreateReservation)).Methods(http.MethodPost)
	router.HandleFunc("/inventory/reservations/{id}", apiLimiter.HandlerFunc(inventoryHandler.GetReservation)).Methods(http.MethodGet)
	router.HandleFunc("/inventory/reservations", apiLimiter.HandlerFunc(inventoryHandler.ListUserReservations)).Methods(http.MethodGet)
	router.HandleFunc("/inventory/reservations/{id}/validate", inventoryLimiter.HandlerFunc(inventoryHandler.ValidateReservation)).Methods(http.MethodGet)
	router.HandleFunc("/inventory/reservations/{id}/extend", inventoryLimiter.HandlerFunc(inventoryHandler.ExtendReservation)).Methods(http.MethodPost)

	// Inventory routes - Release - with rate limiting
	router.HandleFunc("/inventory/reservations/{id}/release", inventoryLimiter.HandlerFunc(inventoryHandler.ReleaseReservation)).Methods(http.MethodDelete)
	router.HandleFunc("/inventory/reservations/expired", paymentLimiter.HandlerFunc(inventoryHandler.ReleaseExpiredReservations)).Methods(http.MethodPost)
	router.HandleFunc("/inventory/reservations/convert", paymentLimiter.HandlerFunc(inventoryHandler.ConvertReservationToOrder)).Methods(http.MethodPost)
	router.HandleFunc("/inventory/reservations/session", inventoryLimiter.HandlerFunc(inventoryHandler.ReleaseSessionReservations)).Methods(http.MethodDelete)
	router.HandleFunc("/inventory/events/{id}/reservations", apiLimiter.HandlerFunc(inventoryHandler.GetReservationsByEvent)).Methods(http.MethodGet)

	// Payment routes - Processing - with rate limiting
	router.HandleFunc("/payments/initiate", paymentLimiter.HandlerFunc(paymentHandler.InitiatePayment)).Methods(http.MethodPost)
	router.HandleFunc("/payments/verify/{id}", paymentLimiter.HandlerFunc(paymentHandler.VerifyPayment)).Methods(http.MethodGet)
	router.HandleFunc("/payments/orders/{id}/status", apiLimiter.HandlerFunc(paymentHandler.GetPaymentStatus)).Methods(http.MethodGet)
	router.HandleFunc("/payments/history", apiLimiter.HandlerFunc(paymentHandler.GetPaymentHistory)).Methods(http.MethodGet)

	// Payment routes - Methods (Saved payment methods) - with rate limiting
	router.HandleFunc("/payments/methods", paymentLimiter.HandlerFunc(paymentHandler.SavePaymentMethod)).Methods(http.MethodPost)
	router.HandleFunc("/payments/methods", apiLimiter.HandlerFunc(paymentHandler.GetPaymentMethods)).Methods(http.MethodGet)
	router.HandleFunc("/payments/methods/{id}", paymentLimiter.HandlerFunc(paymentHandler.DeletePaymentMethod)).Methods(http.MethodDelete)
	router.HandleFunc("/payments/methods/{id}/default", paymentLimiter.HandlerFunc(paymentHandler.SetDefaultPaymentMethod)).Methods(http.MethodPost)
	router.HandleFunc("/payments/methods/{id}/expiry", paymentLimiter.HandlerFunc(paymentHandler.UpdatePaymentMethodExpiry)).Methods(http.MethodPut)

	// Payment routes - Refunds - with rate limiting
	router.HandleFunc("/payments/refunds", paymentLimiter.HandlerFunc(paymentHandler.InitiateRefund)).Methods(http.MethodPost)
	router.HandleFunc("/payments/refunds/{id}/status", apiLimiter.HandlerFunc(paymentHandler.GetRefundStatus)).Methods(http.MethodGet)
	router.HandleFunc("/payments/refunds", apiLimiter.HandlerFunc(paymentHandler.ListRefunds)).Methods(http.MethodGet)
	router.HandleFunc("/payments/refunds/{id}/approve", paymentLimiter.HandlerFunc(paymentHandler.ApproveRefund)).Methods(http.MethodPost)

	// Payment routes - Webhooks
	router.HandleFunc("/webhooks/intasend", paymentHandler.HandleIntasendWebhook).Methods(http.MethodPost)
	router.HandleFunc("/webhooks/logs", paymentHandler.GetWebhookLogs).Methods(http.MethodGet)
	router.HandleFunc("/webhooks/logs/{id}/retry", paymentHandler.RetryFailedWebhook).Methods(http.MethodPost)

	// Payment routes - Gateways
	router.HandleFunc("/payments/gateways", paymentHandler.GetAvailableGateways).Methods(http.MethodGet)

	// Refund routes - Customer - with rate limiting
	router.HandleFunc("/refunds", paymentLimiter.HandlerFunc(refundHandler.RequestRefund)).Methods(http.MethodPost)
	router.HandleFunc("/refunds", apiLimiter.HandlerFunc(refundHandler.ListRefunds)).Methods(http.MethodGet)
	router.HandleFunc("/refunds/{id}", apiLimiter.HandlerFunc(refundHandler.GetRefundStatus)).Methods(http.MethodGet)
	router.HandleFunc("/refunds/{id}/cancel", paymentLimiter.HandlerFunc(refundHandler.CancelRefundRequest)).Methods(http.MethodPost)

	// Refund routes - Admin/Organizer
	router.HandleFunc("/admin/refunds/pending", refundHandler.ListPendingRefunds).Methods(http.MethodGet)
	router.HandleFunc("/admin/refunds/{id}", refundHandler.GetRefundDetails).Methods(http.MethodGet)
	router.HandleFunc("/admin/refunds/{id}/approve", refundHandler.ApproveRefund).Methods(http.MethodPost)
	router.HandleFunc("/admin/refunds/{id}/process", refundHandler.ProcessRefund).Methods(http.MethodPost)
	router.HandleFunc("/admin/refunds/{id}/retry", refundHandler.RetryFailedRefund).Methods(http.MethodPost)
	router.HandleFunc("/admin/refunds/statistics", refundHandler.GetRefundStatistics).Methods(http.MethodGet)

	// Refund routes - Organizer
	router.HandleFunc("/organizers/refunds", refundHandler.ListRefundsByOrganizer).Methods(http.MethodGet)

	// Refund routes - Bulk Operations
	router.HandleFunc("/refunds/bulk/process", refundHandler.ProcessBulkRefunds).Methods(http.MethodPost)
	router.HandleFunc("/refunds/bulk/auto-approve", refundHandler.AutoApproveBulkRefunds).Methods(http.MethodPost)
	router.HandleFunc("/refunds/bulk/stats", refundHandler.GetBulkRefundStats).Methods(http.MethodGet)

	// Settlement routes - Calculation & Preview
	router.HandleFunc("/settlements/calculate/event/{id}", settlementHandler.CalculateEventSettlement).Methods(http.MethodGet)
	router.HandleFunc("/settlements/preview", settlementHandler.GetSettlementPreview).Methods(http.MethodGet)
	router.HandleFunc("/settlements/eligibility/event/{id}", settlementHandler.ValidateSettlementEligibility).Methods(http.MethodGet)

	// Settlement routes - Batch Creation & Processing
	router.HandleFunc("/settlements/batch", settlementHandler.CreateSettlementBatch).Methods(http.MethodPost)
	router.HandleFunc("/settlements/{id}", settlementHandler.GetSettlement).Methods(http.MethodGet)
	router.HandleFunc("/settlements", settlementHandler.ListSettlements).Methods(http.MethodGet)
	router.HandleFunc("/settlements/{id}/approve", settlementHandler.ApproveSettlement).Methods(http.MethodPost)
	router.HandleFunc("/settlements/{id}/process", settlementHandler.ProcessSettlement).Methods(http.MethodPost)
	router.HandleFunc("/settlements/{id}/cancel", settlementHandler.CancelSettlement).Methods(http.MethodPost)
	router.HandleFunc("/settlements/{id}/withhold", settlementHandler.WithholdSettlement).Methods(http.MethodPost)

	// Settlement routes - Reports & Analytics
	router.HandleFunc("/settlements/{id}/report", settlementHandler.GenerateSettlementReport).Methods(http.MethodGet)
	router.HandleFunc("/settlements/summary/organizer/{id}", settlementHandler.GetOrganizerSettlementSummary).Methods(http.MethodGet)
	router.HandleFunc("/settlements/summary/platform", settlementHandler.GetPlatformSettlementSummary).Methods(http.MethodGet)
	router.HandleFunc("/settlements/export", settlementHandler.ExportSettlements).Methods(http.MethodGet)
	router.HandleFunc("/settlements/history/organizer/{id}", settlementHandler.GetSettlementHistory).Methods(http.MethodGet)

	// Settlement routes - Status & Management
	router.HandleFunc("/settlements/pending", settlementHandler.GetPendingSettlements).Methods(http.MethodGet)
	router.HandleFunc("/settlements/failed", settlementHandler.GetFailedSettlements).Methods(http.MethodGet)
	router.HandleFunc("/settlements/{id}/retry", settlementHandler.RetryFailedSettlement).Methods(http.MethodPost)
	router.HandleFunc("/settlements/items/{id}/complete", settlementHandler.CompleteSettlementItem).Methods(http.MethodPost)
	router.HandleFunc("/settlements/items/{id}/fail", settlementHandler.FailSettlementItem).Methods(http.MethodPost)

	// Settlement routes - Organizer View
	router.HandleFunc("/organizers/settlements", settlementHandler.ListSettlements).Methods(http.MethodGet)
	router.HandleFunc("/organizers/settlements/summary", settlementHandler.GetOrganizerSettlementSummary).Methods(http.MethodGet)

	// Settlement routes - Webhooks (for payment gateway callbacks)
	router.HandleFunc("/webhooks/settlements/complete", settlementHandler.HandleSettlementWebhook).Methods(http.MethodPost)

	// Attendee routes - Listing & Search
	router.HandleFunc("/attendees", attendeeHandler.ListAttendees).Methods(http.MethodGet)
	router.HandleFunc("/attendees/filter", attendeeHandler.FilterAttendees).Methods(http.MethodGet)
	router.HandleFunc("/attendees/search", attendeeHandler.SearchAttendees).Methods(http.MethodGet)
	router.HandleFunc("/attendees/search/event", attendeeHandler.SearchAttendeesByEvent).Methods(http.MethodGet)
	router.HandleFunc("/attendees/count", attendeeHandler.GetAttendeeCount).Methods(http.MethodGet)
	router.HandleFunc("/attendees/{id}", attendeeHandler.GetAttendeeDetails).Methods(http.MethodGet)
	router.HandleFunc("/attendees/ticket", attendeeHandler.GetAttendeeByTicket).Methods(http.MethodGet)
	router.HandleFunc("/attendees/order/{id}", attendeeHandler.GetAttendeesByOrder).Methods(http.MethodGet)

	// Attendee routes - Check-in Management
	router.HandleFunc("/attendees/checkin", attendeeHandler.CheckInAttendee).Methods(http.MethodPost)
	router.HandleFunc("/attendees/checkin/bulk", attendeeHandler.BulkCheckIn).Methods(http.MethodPost)
	router.HandleFunc("/attendees/checkin/undo", attendeeHandler.UndoCheckIn).Methods(http.MethodPost)

	// Attendee routes - Update & Management
	router.HandleFunc("/attendees/{id}", attendeeHandler.UpdateAttendeeInfo).Methods(http.MethodPut)
	router.HandleFunc("/attendees/{id}/no-show", attendeeHandler.MarkAttendeeAsNoShow).Methods(http.MethodPost)
	router.HandleFunc("/attendees/{id}/transfer", attendeeHandler.TransferAttendee).Methods(http.MethodPost)

	// Attendee routes - Bulk Operations
	router.HandleFunc("/attendees/bulk/email", func(w http.ResponseWriter, r *http.Request) {
		attendeeHandler.SendBulkEmail(w, r, notificationService)
	}).Methods(http.MethodPost)
	router.HandleFunc("/attendees/bulk/export", attendeeHandler.ExportAttendeesData).Methods(http.MethodPost)
	router.HandleFunc("/attendees/event/update-email", func(w http.ResponseWriter, r *http.Request) {
		attendeeHandler.SendEventUpdateEmail(w, r, notificationService)
	}).Methods(http.MethodPost)

	// Attendee routes - Export & Reports
	router.HandleFunc("/attendees/export", attendeeHandler.ExportAttendeeList).Methods(http.MethodGet)
	router.HandleFunc("/attendees/badges", attendeeHandler.ExportBadgeData).Methods(http.MethodGet)

	// Attendee routes - Analytics
	router.HandleFunc("/attendees/stats", attendeeHandler.GetAttendanceStats).Methods(http.MethodGet)
	router.HandleFunc("/attendees/report/checkin", attendeeHandler.GetCheckInReport).Methods(http.MethodGet)
	router.HandleFunc("/attendees/timeline", attendeeHandler.GetAttendanceTimeline).Methods(http.MethodGet)
	router.HandleFunc("/attendees/no-shows", attendeeHandler.GetNoShowList).Methods(http.MethodGet)

	// Attendee routes - Organizer View
	router.HandleFunc("/organizers/attendees", attendeeHandler.ListEventAttendees).Methods(http.MethodGet)

	// Venue routes - CRUD Operations
	router.HandleFunc("/venues", venueHandler.CreateVenue).Methods(http.MethodPost)
	router.HandleFunc("/venues", venueHandler.ListVenues).Methods(http.MethodGet)
	router.HandleFunc("/venues/{id}", venueHandler.GetVenueDetails).Methods(http.MethodGet)
	router.HandleFunc("/venues/{id}", venueHandler.UpdateVenue).Methods(http.MethodPut)
	router.HandleFunc("/venues/{id}", venueHandler.DeleteVenue).Methods(http.MethodDelete)

	// Venue routes - Search & Discovery
	router.HandleFunc("/venues/search/location", venueHandler.SearchVenuesByLocation).Methods(http.MethodGet)
	router.HandleFunc("/venues/type", venueHandler.GetVenuesByType).Methods(http.MethodGet)

	// Venue routes - Statistics & Information
	router.HandleFunc("/venues/{id}/stats", venueHandler.GetVenueStats).Methods(http.MethodGet)
	router.HandleFunc("/venues/{id}/events", venueHandler.GetVenueEvents).Methods(http.MethodGet)

	// Venue routes - Availability Management
	router.HandleFunc("/venues/{id}/availability", venueHandler.CheckVenueAvailability).Methods(http.MethodGet)
	router.HandleFunc("/venues/{id}/calendar", venueHandler.GetVenueCalendar).Methods(http.MethodGet)
	router.HandleFunc("/venues/available", venueHandler.FindAvailableVenues).Methods(http.MethodGet)

	// Venue routes - Advanced Operations
	router.HandleFunc("/venues/{id}/restore", venueHandler.RestoreVenue).Methods(http.MethodPost)
	router.HandleFunc("/venues/{id}/permanent", venueHandler.PermanentlyDeleteVenue).Methods(http.MethodDelete)

	// Notification routes (if service is available)
	if notificationHandler != nil {
		router.HandleFunc("/notifications/test", notificationHandler.TestEmail).Methods(http.MethodPost)
		router.HandleFunc("/notifications/welcome", notificationHandler.SendWelcomeEmail).Methods(http.MethodPost)
		router.HandleFunc("/notifications/verification", notificationHandler.SendVerificationEmail).Methods(http.MethodPost)
		router.HandleFunc("/notifications/password-reset", notificationHandler.SendPasswordReset).Methods(http.MethodPost)
		fmt.Println("✅ Notification routes registered")
	}

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	fmt.Println("\n🚀 Server starting on port 8080")
	fmt.Println("📊 Prometheus metrics available at http://localhost:8080/metrics")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Printf("❌ Server failed to start: %v\n", err)
		os.Exit(1)
	}
}
