package main

import (
	"fmt"
	"net/http"
	"ticketing_system/internal/accounts"
	"ticketing_system/internal/auth"
	"ticketing_system/internal/database"
	"ticketing_system/internal/events"
	"ticketing_system/internal/models"
	"ticketing_system/internal/organizers"

	"github.com/gorilla/mux"
)

func main() {
	DB := database.Init()

	err := DB.AutoMigrate(&models.User{})
	if err != nil {
		fmt.Printf("Migration failed: %v\n", err)
	} else {
		fmt.Println("✅ Database migration completed successfully")
	}

	authHandler := auth.NewAuthHandler(DB)
	organizerHandler := organizers.NewOrganizerHandler(DB)
	eventHandler := events.NewEventHandler(DB)
	accountHandler := accounts.NewAccountHandler(DB)
	router := mux.NewRouter()

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

	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	fmt.Println("\nserver starting on port 8080")
	server.ListenAndServe()

}
