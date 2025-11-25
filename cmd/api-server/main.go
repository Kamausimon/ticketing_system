package main

import (
	"fmt"
	"net/http"
	"ticketing_system/internal/accounts"
	"ticketing_system/internal/analytics"
	"ticketing_system/internal/attendees"
	"ticketing_system/internal/auth"
	"ticketing_system/internal/database"
	"ticketing_system/internal/events"
	"ticketing_system/internal/inventory"
	"ticketing_system/internal/models"
	"ticketing_system/internal/orders"
	"ticketing_system/internal/organizers"
	"ticketing_system/internal/payments"
	"ticketing_system/internal/promotions"
	"ticketing_system/internal/refunds"
	"ticketing_system/internal/settlement"
	"ticketing_system/internal/tickets"
	"ticketing_system/internal/venues"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	DB := database.Init()

	err := DB.AutoMigrate(&models.User{})
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

	authHandler := auth.NewAuthHandler(DB)
	organizerHandler := organizers.NewOrganizerHandler(DB)
	eventHandler := events.NewEventHandler(DB)
	accountHandler := accounts.NewAccountHandler(DB)
	orderHandler := orders.NewOrderHandler(DB)
	ticketHandler := tickets.NewTicketHandler(DB)
	promotionHandler := promotions.NewPromotionHandler(DB)
	inventoryHandler := inventory.NewInventoryHandler(DB)
	paymentHandler := payments.NewPaymentHandler(DB)
	refundHandler := refunds.NewRefundHandler(DB, paymentHandler.IntasendSecretKey, paymentHandler.IntasendWebhookSecret, paymentHandler.IntasendTestMode)
	settlementService := settlement.NewService(DB)
	settlementHandler := settlement.NewSettlementHandler(settlementService)
	attendeeHandler := attendees.NewAttendeeHandler(DB)
	venueHandler := venues.NewVenueHandler(DB)
	router := mux.NewRouter()

	// Add Prometheus middleware
	router.Use(analytics.PrometheusMiddleware(metrics))

	// Expose Prometheus metrics endpoint
	router.Handle("/metrics", promhttp.Handler())

	//auth routes
	router.HandleFunc("/register", authHandler.RegisterUser).Methods(http.MethodPost)
	router.HandleFunc("/login", authHandler.LoginUser).Methods(http.MethodPost)
	router.HandleFunc("/logout", authHandler.LogoutUser).Methods(http.MethodPost)
	router.HandleFunc("/forgot-passoword", authHandler.ForgotPassword).Methods(http.MethodPost)
	router.HandleFunc("/resetPassword", authHandler.ResetPassword).Methods(http.MethodPost)

	//organizer routes
	router.HandleFunc("/organizers/apply", organizerHandler.OrganizerApply).Methods(http.MethodPost)
	router.HandleFunc("/organizers/profile", organizerHandler.GetOrganizerProfile).Methods(http.MethodGet)
	router.HandleFunc("/organizers/profile", organizerHandler.UpdateOrganizerProfile).Methods(http.MethodPut)
	router.HandleFunc("/organizers/onboarding/status", organizerHandler.GetOnboardingStatus).Methods(http.MethodGet)
	router.HandleFunc("/organizers/dashboard", organizerHandler.GetOrganizerDashboard).Methods(http.MethodGet)
	router.HandleFunc("/organizers/dashboard/stats", organizerHandler.GetQuickStats).Methods(http.MethodGet)
	router.HandleFunc("/organizers/logo", organizerHandler.UploadOrganizerLogo).Methods(http.MethodPost)
	router.HandleFunc("/organizers/verification/email", organizerHandler.SendVerificationEmail).Methods(http.MethodPost)

	// Admin organizer routes
	router.HandleFunc("/admin/organizers/pending", organizerHandler.GetPendingOrganizers).Methods(http.MethodGet)
	router.HandleFunc("/admin/organizers/{id}/verify", organizerHandler.VerifyOrganizer).Methods(http.MethodPost)

	// Event routes - Public
	router.HandleFunc("/events", eventHandler.ListEvents).Methods(http.MethodGet)
	router.HandleFunc("/events/{id}", eventHandler.GetEventDetails).Methods(http.MethodGet)
	router.HandleFunc("/events/{id}/images", eventHandler.GetEventImages).Methods(http.MethodGet)

	// Event routes - Organizer only
	router.HandleFunc("/organizers/events", eventHandler.ListOrganizerEvents).Methods(http.MethodGet)
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

	// Order routes - Creation & Calculation
	router.HandleFunc("/orders", orderHandler.CreateOrder).Methods(http.MethodPost)
	router.HandleFunc("/orders/calculate", orderHandler.CalculateOrder).Methods(http.MethodPost)

	// Order routes - Viewing
	router.HandleFunc("/orders", orderHandler.ListOrders).Methods(http.MethodGet)
	router.HandleFunc("/orders/{id}", orderHandler.GetOrderDetails).Methods(http.MethodGet)
	router.HandleFunc("/orders/{id}/summary", orderHandler.GetOrderSummary).Methods(http.MethodGet)
	router.HandleFunc("/orders/stats", orderHandler.GetOrderStats).Methods(http.MethodGet)

	// Order routes - Management
	router.HandleFunc("/orders/{id}/status", orderHandler.UpdateOrderStatus).Methods(http.MethodPut)
	router.HandleFunc("/orders/{id}/cancel", orderHandler.CancelOrder).Methods(http.MethodPost)
	router.HandleFunc("/orders/{id}/refund", orderHandler.RefundOrder).Methods(http.MethodPost)

	// Order routes - Payment
	router.HandleFunc("/orders/{id}/payment", orderHandler.ProcessPayment).Methods(http.MethodPost)
	router.HandleFunc("/orders/{id}/payment/verify", orderHandler.VerifyPayment).Methods(http.MethodPost)

	// Order routes - Organizer view
	router.HandleFunc("/organizers/orders", orderHandler.ListOrganizerOrders).Methods(http.MethodGet)

	// Ticket routes - Generation
	router.HandleFunc("/tickets/generate", ticketHandler.GenerateTickets).Methods(http.MethodPost)
	router.HandleFunc("/tickets/regenerate-qr", ticketHandler.RegenerateTicketQR).Methods(http.MethodPost)

	// Ticket routes - Viewing
	router.HandleFunc("/tickets", ticketHandler.ListUserTickets).Methods(http.MethodGet)
	router.HandleFunc("/tickets/{id}", ticketHandler.GetTicketDetails).Methods(http.MethodGet)
	router.HandleFunc("/tickets/number", ticketHandler.GetTicketByNumber).Methods(http.MethodGet)
	router.HandleFunc("/tickets/stats", ticketHandler.GetTicketStats).Methods(http.MethodGet)

	// Ticket routes - PDF Download
	router.HandleFunc("/tickets/{id}/pdf", ticketHandler.DownloadTicketPDF).Methods(http.MethodGet)

	// Ticket routes - Transfer
	router.HandleFunc("/tickets/{id}/transfer", ticketHandler.TransferTicket).Methods(http.MethodPost)
	router.HandleFunc("/tickets/{id}/transfer-history", ticketHandler.GetTransferHistory).Methods(http.MethodGet)

	// Ticket routes - Validation (Organizer only)
	router.HandleFunc("/tickets/validate", ticketHandler.ValidateTicket).Methods(http.MethodPost)
	router.HandleFunc("/tickets/validate/qr", ticketHandler.ValidateTicketByQR).Methods(http.MethodPost)

	// Ticket routes - Check-in (Organizer only)
	router.HandleFunc("/tickets/checkin", ticketHandler.CheckInTicket).Methods(http.MethodPost)
	router.HandleFunc("/tickets/checkin/bulk", ticketHandler.BulkCheckIn).Methods(http.MethodPost)
	router.HandleFunc("/tickets/checkin/undo", ticketHandler.UndoCheckIn).Methods(http.MethodPost)
	router.HandleFunc("/tickets/checkin/stats", ticketHandler.GetCheckInStats).Methods(http.MethodGet)

	// Ticket routes - Event tickets (Organizer only)
	router.HandleFunc("/organizers/tickets", ticketHandler.ListEventTickets).Methods(http.MethodGet)

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

	// Inventory routes - Availability
	router.HandleFunc("/inventory/tickets/{id}", inventoryHandler.GetTicketAvailability).Methods(http.MethodGet)
	router.HandleFunc("/inventory/events/{id}", inventoryHandler.GetEventInventory).Methods(http.MethodGet)
	router.HandleFunc("/inventory/status/{id}", inventoryHandler.GetInventoryStatus).Methods(http.MethodGet)
	router.HandleFunc("/inventory/bulk-check", inventoryHandler.BulkCheckAvailability).Methods(http.MethodPost)

	// Inventory routes - Reservations
	router.HandleFunc("/inventory/reservations", inventoryHandler.CreateReservation).Methods(http.MethodPost)
	router.HandleFunc("/inventory/reservations/{id}", inventoryHandler.GetReservation).Methods(http.MethodGet)
	router.HandleFunc("/inventory/reservations", inventoryHandler.ListUserReservations).Methods(http.MethodGet)
	router.HandleFunc("/inventory/reservations/{id}/validate", inventoryHandler.ValidateReservation).Methods(http.MethodGet)
	router.HandleFunc("/inventory/reservations/{id}/extend", inventoryHandler.ExtendReservation).Methods(http.MethodPost)

	// Inventory routes - Release
	router.HandleFunc("/inventory/reservations/{id}/release", inventoryHandler.ReleaseReservation).Methods(http.MethodDelete)
	router.HandleFunc("/inventory/reservations/expired", inventoryHandler.ReleaseExpiredReservations).Methods(http.MethodPost)
	router.HandleFunc("/inventory/reservations/convert", inventoryHandler.ConvertReservationToOrder).Methods(http.MethodPost)
	router.HandleFunc("/inventory/reservations/session", inventoryHandler.ReleaseSessionReservations).Methods(http.MethodDelete)
	router.HandleFunc("/inventory/events/{id}/reservations", inventoryHandler.GetReservationsByEvent).Methods(http.MethodGet)

	// Payment routes - Processing
	router.HandleFunc("/payments/initiate", paymentHandler.InitiatePayment).Methods(http.MethodPost)
	router.HandleFunc("/payments/verify/{id}", paymentHandler.VerifyPayment).Methods(http.MethodGet)
	router.HandleFunc("/payments/orders/{id}/status", paymentHandler.GetPaymentStatus).Methods(http.MethodGet)
	router.HandleFunc("/payments/history", paymentHandler.GetPaymentHistory).Methods(http.MethodGet)

	// Payment routes - Methods (Saved payment methods)
	router.HandleFunc("/payments/methods", paymentHandler.SavePaymentMethod).Methods(http.MethodPost)
	router.HandleFunc("/payments/methods", paymentHandler.GetPaymentMethods).Methods(http.MethodGet)
	router.HandleFunc("/payments/methods/{id}", paymentHandler.DeletePaymentMethod).Methods(http.MethodDelete)
	router.HandleFunc("/payments/methods/{id}/default", paymentHandler.SetDefaultPaymentMethod).Methods(http.MethodPost)
	router.HandleFunc("/payments/methods/{id}/expiry", paymentHandler.UpdatePaymentMethodExpiry).Methods(http.MethodPut)

	// Payment routes - Refunds
	router.HandleFunc("/payments/refunds", paymentHandler.InitiateRefund).Methods(http.MethodPost)
	router.HandleFunc("/payments/refunds/{id}/status", paymentHandler.GetRefundStatus).Methods(http.MethodGet)
	router.HandleFunc("/payments/refunds", paymentHandler.ListRefunds).Methods(http.MethodGet)
	router.HandleFunc("/payments/refunds/{id}/approve", paymentHandler.ApproveRefund).Methods(http.MethodPost)

	// Payment routes - Webhooks
	router.HandleFunc("/webhooks/intasend", paymentHandler.HandleIntasendWebhook).Methods(http.MethodPost)
	router.HandleFunc("/webhooks/logs", paymentHandler.GetWebhookLogs).Methods(http.MethodGet)
	router.HandleFunc("/webhooks/logs/{id}/retry", paymentHandler.RetryFailedWebhook).Methods(http.MethodPost)

	// Payment routes - Gateways
	router.HandleFunc("/payments/gateways", paymentHandler.GetAvailableGateways).Methods(http.MethodGet)

	// Refund routes - Customer
	router.HandleFunc("/refunds", refundHandler.RequestRefund).Methods(http.MethodPost)
	router.HandleFunc("/refunds", refundHandler.ListRefunds).Methods(http.MethodGet)
	router.HandleFunc("/refunds/{id}", refundHandler.GetRefundStatus).Methods(http.MethodGet)
	router.HandleFunc("/refunds/{id}/cancel", refundHandler.CancelRefundRequest).Methods(http.MethodPost)

	// Refund routes - Admin/Organizer
	router.HandleFunc("/admin/refunds/pending", refundHandler.ListPendingRefunds).Methods(http.MethodGet)
	router.HandleFunc("/admin/refunds/{id}", refundHandler.GetRefundDetails).Methods(http.MethodGet)
	router.HandleFunc("/admin/refunds/{id}/approve", refundHandler.ApproveRefund).Methods(http.MethodPost)
	router.HandleFunc("/admin/refunds/{id}/process", refundHandler.ProcessRefund).Methods(http.MethodPost)
	router.HandleFunc("/admin/refunds/{id}/retry", refundHandler.RetryFailedRefund).Methods(http.MethodPost)
	router.HandleFunc("/admin/refunds/statistics", refundHandler.GetRefundStatistics).Methods(http.MethodGet)

	// Refund routes - Organizer
	router.HandleFunc("/organizers/refunds", refundHandler.ListRefundsByOrganizer).Methods(http.MethodGet)

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
	router.HandleFunc("/attendees/search", attendeeHandler.SearchAttendees).Methods(http.MethodGet)
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

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	fmt.Println("\n🚀 Server starting on port 8080")
	fmt.Println("📊 Prometheus metrics available at http://localhost:8080/metrics")
	server.ListenAndServe()

}
