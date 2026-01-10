package settlement

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"ticketing_system/internal/models"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate all required tables
	err = db.AutoMigrate(
		&models.Account{},
		&models.User{},
		&models.Organizer{},
		&models.PayoutAccount{},
		&models.Event{},
		&models.TicketClass{},
		&models.Order{},
		&models.OrderItem{},
		&models.Ticket{},
		&models.PaymentRecord{},
		&models.RefundRecord{},
		&models.SettlementRecord{},
		&models.SettlementItem{},
	)
	assert.NoError(t, err)

	return db
}

// createTestEvent creates a test event with completed payments
func createTestEvent(t *testing.T, db *gorm.DB, status models.EventStatus) (uint, uint, uint) {
	// Create account
	account := models.Account{
		Email:    "test@example.com",
		IsActive: true,
	}
	err := db.Create(&account).Error
	assert.NoError(t, err)

	// Create user
	user := models.User{
		AccountID: account.ID,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}
	err = db.Create(&user).Error
	assert.NoError(t, err)

	// Create organizer
	organizer := models.Organizer{
		Name:              "Test Organizer",
		Email:             "organizer@example.com",
		AccountID:         account.ID,
		Phone:             "1234567890",
		BankAccountName:   "Test Organizer",
		BankAccountNumber: "1234567890",
		BankCode:          "001",
		BankCountry:       "US",
	}
	err = db.Create(&organizer).Error
	assert.NoError(t, err)

	// Create verified payout account for organizer
	payoutAccount := models.PayoutAccount{
		OrganizerID:       organizer.ID,
		AccountType:       models.PayoutBank,
		Status:            models.PayoutVerified,
		DisplayName:       "Test Bank Account",
		IsDefault:         true,
		BankName:          stringPtr("Test Bank"),
		BankCode:          stringPtr("001"),
		BankCountry:       stringPtr("US"),
		AccountNumber:     stringPtr("1234567890"),
		AccountHolderName: stringPtr("Test Organizer"),
		Currency:          "USD",
		IsVerified:        true,
	}
	err = db.Create(&payoutAccount).Error
	assert.NoError(t, err)

	// Create event
	pastDate := time.Now().Add(-7 * 24 * time.Hour) // 7 days ago
	event := models.Event{
		Title:       "Test Event",
		Description: "Test Description",
		StartDate:   pastDate,
		EndDate:     pastDate.Add(2 * time.Hour),
		Location:    "Test Location",
		AccountID:   account.ID,
		OrganizerID: organizer.ID,
		Currency:    "USD",
		BarcodeType: "qr",
		Status:      status,
		Category:    models.CategoryConference,
	}
	err = db.Create(&event).Error
	assert.NoError(t, err)

	// Create ticket class
	quantity := 100
	now := time.Now()
	ticketClass := models.TicketClass{
		EventID:           event.ID,
		Name:              "General Admission",
		Description:       "General admission ticket",
		Price:             models.Money(5000), // $50.00
		Currency:          "USD",
		QuantityAvailable: &quantity,
		StartSaleDate:     &now,
		EndSaleDate:       &pastDate,
	}
	err = db.Create(&ticketClass).Error
	assert.NoError(t, err)

	// Create completed order
	order := models.Order{
		AccountID:         account.ID,
		EventID:           event.ID,
		FirstName:         "John",
		LastName:          "Doe",
		Email:             "john@example.com",
		Amount:            100.00,
		Status:            models.OrderFulfilled,
		PaymentStatus:     models.PaymentCompleted,
		TaxAmount:         0,
		IsPaymentReceived: true,
	}
	err = db.Create(&order).Error
	assert.NoError(t, err)

	// Create payment records
	paymentRecord := models.PaymentRecord{
		OrderID:               &order.ID,
		EventID:               &event.ID,
		OrganizerID:           &organizer.ID,
		Type:                  models.RecordCustomerPayment,
		Status:                models.RecordCompleted,
		Amount:                models.Money(10000), // $100.00
		Currency:              "USD",
		Description:           "Ticket payment",
		InitiatedAt:           now,
		ProcessedAt:           &now,
		ExternalTransactionID: stringPtr("txn_123456"),
		NetAmount:             models.Money(10000),
	}
	err = db.Create(&paymentRecord).Error
	assert.NoError(t, err)

	// Create platform fee record
	platformFee := models.PaymentRecord{
		OrderID:     &order.ID,
		EventID:     &event.ID,
		OrganizerID: &organizer.ID,
		Type:        models.RecordPlatformFee,
		Status:      models.RecordCompleted,
		Amount:      models.Money(500), // $5.00
		Currency:    "USD",
		Description: "Platform commission (5%)",
		InitiatedAt: now,
		ProcessedAt: &now,
		NetAmount:   models.Money(500),
	}
	err = db.Create(&platformFee).Error
	assert.NoError(t, err)

	// Create gateway fee record
	gatewayFee := models.PaymentRecord{
		OrderID:     &order.ID,
		EventID:     &event.ID,
		OrganizerID: &organizer.ID,
		Type:        models.RecordGatewayFee,
		Status:      models.RecordCompleted,
		Amount:      models.Money(300), // $3.00
		Currency:    "USD",
		Description: "Payment gateway fee",
		InitiatedAt: now,
		ProcessedAt: &now,
		NetAmount:   models.Money(300),
	}
	err = db.Create(&gatewayFee).Error
	assert.NoError(t, err)

	return event.ID, organizer.ID, user.ID
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// TestCalculateEventSettlement tests the calculate event settlement endpoint
func TestCalculateEventSettlement(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	handler := NewSettlementHandler(service)

	eventID, _, _ := createTestEvent(t, db, models.EventCompleted)

	// Create request
	req, err := http.NewRequest("GET", fmt.Sprintf("/settlements/calculate/event/%d", eventID), nil)
	assert.NoError(t, err)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Create router with mux
	router := mux.NewRouter()
	router.HandleFunc("/settlements/calculate/event/{id}", handler.CalculateEventSettlement)

	// Serve request
	router.ServeHTTP(rr, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse response
	var calculation SettlementCalculation
	err = json.Unmarshal(rr.Body.Bytes(), &calculation)
	assert.NoError(t, err)

	// Verify calculation
	assert.Equal(t, eventID, calculation.EventID)
	assert.Equal(t, models.Money(10000), calculation.GrossAmount)     // $100.00
	assert.Equal(t, models.Money(500), calculation.PlatformFeeAmount) // $5.00
	assert.Equal(t, models.Money(300), calculation.GatewayFeeAmount)  // $3.00
	assert.Equal(t, models.Money(9200), calculation.NetAmount)        // $92.00 (100 - 5 - 3)
}

// TestCalculateEventSettlement_NotCompleted tests calculating settlement for non-completed event
func TestCalculateEventSettlement_NotCompleted(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	handler := NewSettlementHandler(service)

	eventID, _, _ := createTestEvent(t, db, models.EventLive) // Not completed

	req, err := http.NewRequest("GET", fmt.Sprintf("/settlements/calculate/event/%d", eventID), nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/settlements/calculate/event/{id}", handler.CalculateEventSettlement)
	router.ServeHTTP(rr, req)

	// Should return error
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// TestGetSettlementPreview tests the settlement preview endpoint
func TestGetSettlementPreview(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	handler := NewSettlementHandler(service)

	eventID, organizerID, _ := createTestEvent(t, db, models.EventCompleted)

	// Create request with query params
	req, err := http.NewRequest("GET", fmt.Sprintf("/settlements/preview?organizer_id=%d&event_id=%d", organizerID, eventID), nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/settlements/preview", handler.GetSettlementPreview)
	router.ServeHTTP(rr, req)

	// Should return preview
	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestValidateSettlementEligibility tests the eligibility validation endpoint
func TestValidateSettlementEligibility(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	handler := NewSettlementHandler(service)

	eventID, _, _ := createTestEvent(t, db, models.EventCompleted)

	// Create request
	req, err := http.NewRequest("GET", fmt.Sprintf("/settlements/eligibility/event/%d", eventID), nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/settlements/eligibility/event/{id}", handler.ValidateSettlementEligibility)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse response
	var result map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Contains(t, result, "eligible")
}

// TestCreateSettlementBatch tests creating a settlement batch
func TestCreateSettlementBatch(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	handler := NewSettlementHandler(service)

	eventID, _, userID := createTestEvent(t, db, models.EventCompleted)

	// Create request body
	reqBody := CreateSettlementBatchRequest{
		Description:       "Test Settlement Batch",
		Frequency:         models.SettlementPostEvent,
		Trigger:           models.TriggerEventCompletion,
		PeriodStartDate:   time.Now().Add(-30 * 24 * time.Hour),
		PeriodEndDate:     time.Now(),
		HoldingPeriodDays: 7,
		InitiatedByUserID: userID,
		EventID:           &eventID,
	}

	body, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/settlements/batch", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/settlements/batch", handler.CreateSettlementBatch)
	router.ServeHTTP(rr, req)

	// Check status
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Parse response
	var settlement models.SettlementRecord
	err = json.Unmarshal(rr.Body.Bytes(), &settlement)
	assert.NoError(t, err)
	assert.Equal(t, models.SettlementPending, settlement.Status)
	assert.NotEmpty(t, settlement.SettlementBatchID)
}

// TestGetSettlement tests retrieving a specific settlement
func TestGetSettlement(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	handler := NewSettlementHandler(service)

	eventID, _, userID := createTestEvent(t, db, models.EventCompleted)

	// Create a settlement first
	reqBody := CreateSettlementBatchRequest{
		Description:       "Test Settlement",
		Frequency:         models.SettlementPostEvent,
		Trigger:           models.TriggerEventCompletion,
		PeriodStartDate:   time.Now().Add(-30 * 24 * time.Hour),
		PeriodEndDate:     time.Now(),
		HoldingPeriodDays: 7,
		InitiatedByUserID: userID,
		EventID:           &eventID,
	}

	settlement, err := service.CreateSettlementBatch(reqBody)
	assert.NoError(t, err)

	// Get the settlement
	req, err := http.NewRequest("GET", fmt.Sprintf("/settlements/%d", settlement.ID), nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/settlements/{id}", handler.GetSettlement)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse response
	var result models.SettlementRecord
	err = json.Unmarshal(rr.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, settlement.ID, result.ID)
}

// TestListSettlements tests listing settlements with filters
func TestListSettlements(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	handler := NewSettlementHandler(service)

	eventID, _, userID := createTestEvent(t, db, models.EventCompleted)

	// Create a settlement
	reqBody := CreateSettlementBatchRequest{
		Description:       "Test Settlement",
		Frequency:         models.SettlementPostEvent,
		Trigger:           models.TriggerEventCompletion,
		PeriodStartDate:   time.Now().Add(-30 * 24 * time.Hour),
		PeriodEndDate:     time.Now(),
		HoldingPeriodDays: 7,
		InitiatedByUserID: userID,
		EventID:           &eventID,
	}
	_, err := service.CreateSettlementBatch(reqBody)
	assert.NoError(t, err)

	// List settlements
	req, err := http.NewRequest("GET", "/settlements?page=1&limit=10", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/settlements", handler.ListSettlements)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse response
	var result map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Contains(t, result, "settlements")
	assert.Contains(t, result, "total_count")

	settlements := result["settlements"].([]interface{})
	assert.GreaterOrEqual(t, len(settlements), 1)
}

// TestListSettlementsWithFilters tests listing settlements with status filter
func TestListSettlementsWithFilters(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	handler := NewSettlementHandler(service)

	eventID, _, userID := createTestEvent(t, db, models.EventCompleted)

	// Create a settlement
	reqBody := CreateSettlementBatchRequest{
		Description:       "Test Settlement",
		Frequency:         models.SettlementPostEvent,
		Trigger:           models.TriggerEventCompletion,
		PeriodStartDate:   time.Now().Add(-30 * 24 * time.Hour),
		PeriodEndDate:     time.Now(),
		HoldingPeriodDays: 7,
		InitiatedByUserID: userID,
		EventID:           &eventID,
	}
	_, err := service.CreateSettlementBatch(reqBody)
	assert.NoError(t, err)

	// List with filter
	req, err := http.NewRequest("GET", "/settlements?status=pending&page=1&limit=10", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/settlements", handler.ListSettlements)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestApproveSettlement tests approving a settlement
func TestApproveSettlement(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	handler := NewSettlementHandler(service)

	eventID, _, userID := createTestEvent(t, db, models.EventCompleted)

	// Create a settlement
	reqBody := CreateSettlementBatchRequest{
		Description:       "Test Settlement",
		Frequency:         models.SettlementPostEvent,
		Trigger:           models.TriggerEventCompletion,
		PeriodStartDate:   time.Now().Add(-30 * 24 * time.Hour),
		PeriodEndDate:     time.Now(),
		HoldingPeriodDays: 7,
		InitiatedByUserID: userID,
		EventID:           &eventID,
	}
	settlement, err := service.CreateSettlementBatch(reqBody)
	assert.NoError(t, err)

	// Update the settlement to bypass holding period for testing
	pastDate := time.Now().Add(-1 * time.Hour)
	db.Model(&settlement).Update("earliest_payout_date", &pastDate)

	// Approve the settlement
	approveReq := map[string]interface{}{
		"approved_by_user_id": userID,
		"notes":               "Approved for testing",
	}
	body, err := json.Marshal(approveReq)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", fmt.Sprintf("/settlements/%d/approve", settlement.ID), bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/settlements/{id}/approve", handler.ApproveSettlement)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify in database
	var updatedSettlement models.SettlementRecord
	err = db.First(&updatedSettlement, settlement.ID).Error
	assert.NoError(t, err)
	assert.NotNil(t, updatedSettlement.ApprovedAt)
	assert.Equal(t, userID, *updatedSettlement.ApprovedBy)
}

// TestCancelSettlement tests canceling a settlement
func TestCancelSettlement(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	handler := NewSettlementHandler(service)

	eventID, _, userID := createTestEvent(t, db, models.EventCompleted)

	// Create a settlement
	reqBody := CreateSettlementBatchRequest{
		Description:       "Test Settlement",
		Frequency:         models.SettlementPostEvent,
		Trigger:           models.TriggerEventCompletion,
		PeriodStartDate:   time.Now().Add(-30 * 24 * time.Hour),
		PeriodEndDate:     time.Now(),
		HoldingPeriodDays: 7,
		InitiatedByUserID: userID,
		EventID:           &eventID,
	}
	settlement, err := service.CreateSettlementBatch(reqBody)
	assert.NoError(t, err)

	// Cancel the settlement
	cancelReq := map[string]interface{}{
		"cancelled_by_user_id": userID,
		"reason":               "Testing cancellation",
	}
	body, err := json.Marshal(cancelReq)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", fmt.Sprintf("/settlements/%d/cancel", settlement.ID), bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/settlements/{id}/cancel", handler.CancelSettlement)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify in database
	var updatedSettlement models.SettlementRecord
	err = db.First(&updatedSettlement, settlement.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, models.SettlementCancelled, updatedSettlement.Status)
}

// TestWithholdSettlement tests withholding a settlement
func TestWithholdSettlement(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	handler := NewSettlementHandler(service)

	eventID, _, userID := createTestEvent(t, db, models.EventCompleted)

	// Create a settlement
	reqBody := CreateSettlementBatchRequest{
		Description:       "Test Settlement",
		Frequency:         models.SettlementPostEvent,
		Trigger:           models.TriggerEventCompletion,
		PeriodStartDate:   time.Now().Add(-30 * 24 * time.Hour),
		PeriodEndDate:     time.Now(),
		HoldingPeriodDays: 7,
		InitiatedByUserID: userID,
		EventID:           &eventID,
	}
	settlement, err := service.CreateSettlementBatch(reqBody)
	assert.NoError(t, err)

	// Withhold the settlement
	withholdReq := map[string]interface{}{
		"withheld_by_user_id": userID,
		"reason":              "Fraud investigation",
	}
	body, err := json.Marshal(withholdReq)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", fmt.Sprintf("/settlements/%d/withhold", settlement.ID), bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/settlements/{id}/withhold", handler.WithholdSettlement)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify in database
	var updatedSettlement models.SettlementRecord
	err = db.First(&updatedSettlement, settlement.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, models.SettlementWithheld, updatedSettlement.Status)
	assert.NotNil(t, updatedSettlement.WithholdingReason)
}

// TestGetPendingSettlements tests retrieving pending settlements
func TestGetPendingSettlements(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	handler := NewSettlementHandler(service)

	eventID, _, userID := createTestEvent(t, db, models.EventCompleted)

	// Create a pending settlement
	reqBody := CreateSettlementBatchRequest{
		Description:       "Pending Settlement",
		Frequency:         models.SettlementPostEvent,
		Trigger:           models.TriggerEventCompletion,
		PeriodStartDate:   time.Now().Add(-30 * 24 * time.Hour),
		PeriodEndDate:     time.Now(),
		HoldingPeriodDays: 7,
		InitiatedByUserID: userID,
		EventID:           &eventID,
	}
	_, err := service.CreateSettlementBatch(reqBody)
	assert.NoError(t, err)

	// Get pending settlements
	req, err := http.NewRequest("GET", "/settlements/pending", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/settlements/pending", handler.GetPendingSettlements)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse response
	var settlements []models.SettlementRecord
	err = json.Unmarshal(rr.Body.Bytes(), &settlements)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(settlements), 1)
}

// TestInvalidEventID tests handling of invalid event IDs
func TestInvalidEventID(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	handler := NewSettlementHandler(service)

	req, err := http.NewRequest("GET", "/settlements/calculate/event/invalid", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/settlements/calculate/event/{id}", handler.CalculateEventSettlement)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestInvalidSettlementID tests handling of invalid settlement IDs
func TestInvalidSettlementID(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	handler := NewSettlementHandler(service)

	req, err := http.NewRequest("GET", "/settlements/invalid", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/settlements/{id}", handler.GetSettlement)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestNonExistentSettlement tests retrieving a non-existent settlement
func TestNonExistentSettlement(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	handler := NewSettlementHandler(service)

	req, err := http.NewRequest("GET", "/settlements/99999", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/settlements/{id}", handler.GetSettlement)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// TestCreateSettlementBatch_InvalidRequest tests creating batch with invalid request
func TestCreateSettlementBatch_InvalidRequest(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	handler := NewSettlementHandler(service)

	// Send invalid JSON
	req, err := http.NewRequest("POST", "/settlements/batch", bytes.NewBuffer([]byte("invalid json")))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/settlements/batch", handler.CreateSettlementBatch)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
