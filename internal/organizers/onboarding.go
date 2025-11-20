package organizers

import (
	"encoding/json"
	"net/http"
	"ticketing_system/internal/middleware"
	"ticketing_system/internal/models"
)

// Onboarding-related structures
type OnboardingStep struct {
	Step        string `json:"step"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
	Required    bool   `json:"required"`
}

type OnboardingStatusResponse struct {
	CurrentStep    string           `json:"current_step"`
	Progress       float64          `json:"progress"`
	Steps          []OnboardingStep `json:"steps"`
	CanCreateEvents bool            `json:"can_create_events"`
}

// GetOnboardingStatus returns the current onboarding status
func (h *OrganizerHandler) GetOnboardingStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	userID := middleware.GetUserIDFromToken(r)
	
	// Get user and organizer info
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}
	
	var organizer models.Organizer
	if err := h.db.Where("account_id = ?", user.AccountID).First(&organizer).Error; err != nil {
		middleware.WriteJSONError(w, http.StatusNotFound, "organizer profile not found")
		return
	}
	
	// Define onboarding steps
	steps := []OnboardingStep{
		{
			Step:        "profile_complete",
			Title:       "Complete Profile",
			Description: "Fill out your business information",
			Completed:   isProfileComplete(organizer),
			Required:    true,
		},
		{
			Step:        "email_verified",
			Title:       "Verify Email",
			Description: "Confirm your email address",
			Completed:   organizer.IsEmailConfirmed,
			Required:    true,
		},
		{
			Step:        "tax_info",
			Title:       "Tax Information",
			Description: "Provide tax details for payouts",
			Completed:   organizer.TaxName != "" && organizer.TaxPin != "",
			Required:    true,
		},
		{
			Step:        "payment_setup",
			Title:       "Payment Gateway",
			Description: "Set up payment processing",
			Completed:   false, // TODO: Check if payment gateway is configured
			Required:    true,
		},
		{
			Step:        "branding",
			Title:       "Branding Setup",
			Description: "Upload logo and customize page",
			Completed:   organizer.LogoPath != nil && *organizer.LogoPath != "",
			Required:    false,
		},
	}
	
	// Calculate progress
	completedSteps := 0
	requiredSteps := 0
	var currentStep string
	
	for _, step := range steps {
		if step.Required {
			requiredSteps++
		}
		if step.Completed {
			completedSteps++
		} else if currentStep == "" && step.Required {
			currentStep = step.Step
		}
	}
	
	progress := float64(completedSteps) / float64(len(steps)) * 100
	allRequiredComplete := completedSteps >= requiredSteps
	
	if currentStep == "" {
		currentStep = "completed"
	}
	
	response := OnboardingStatusResponse{
		CurrentStep:     currentStep,
		Progress:        progress,
		Steps:           steps,
		CanCreateEvents: allRequiredComplete,
	}
	
	json.NewEncoder(w).Encode(response)
}

// Helper function to check if profile is complete
func isProfileComplete(organizer models.Organizer) bool {
	return organizer.Name != "" &&
		organizer.About != "" &&
		organizer.Email != "" &&
		organizer.Phone != ""
}
