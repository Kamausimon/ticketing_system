package attendees

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"ticketing_system/internal/analytics"
	"ticketing_system/internal/models"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB initializes an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.Attendee{},
		&models.Ticket{},
		&models.Event{},
		&models.Order{},
		&models.Account{},
		&models.TicketClass{},
		&models.OrderItem{},
		&models.Organizer{},
	)
	require.NoError(t, err)

	return db
}

// setupTestData creates sample data for testing
func setupTestData(t *testing.T, db *gorm.DB) (models.Account, models.Event, models.Order, models.Ticket, models.Attendee, models.TicketClass) {
	// Create account
	account := models.Account{
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}
	require.NoError(t, db.Create(&account).Error)

	// Create organizer
	organizer := models.Organizer{
		Name:      "Test Organizer",
		AccountID: account.ID,
	}
	require.NoError(t, db.Create(&organizer).Error)

	// Create event
	startDate := time.Now().Add(24 * time.Hour)
	endDate := startDate.Add(3 * time.Hour)
	event := models.Event{
		Title:       "Test Event",
		Description: "Test Description",
		StartDate:   startDate,
		EndDate:     endDate,
		OrganizerID: organizer.ID,
		AccountID:   account.ID,
		Location:    "Test Location",
		Currency:    "USD",
		Category:    models.CategoryConference,
		BarcodeType: "QR",
	}
	require.NoError(t, db.Create(&event).Error)

	// Create ticket class
	ticketClass := models.TicketClass{
		EventID:           event.ID,
		Name:              "General Admission",
		Price:             models.Money(5000), // 50.00 in cents
		Currency:          "USD",
		QuantityAvailable: intPtr(100),
		QuantitySold:      10,
	}
	require.NoError(t, db.Create(&ticketClass).Error)

	// Create order
	order := models.Order{
		AccountID:         account.ID,
		EventID:           event.ID,
		TotalAmount:       models.Money(5000),
		Amount:            50.00,
		Status:            models.OrderPaid,
		PaymentStatus:     models.PaymentCompleted,
		Currency:          "USD",
		IsPaymentReceived: true,
	}
	require.NoError(t, db.Create(&order).Error)

	// Create order item
	orderItem := models.OrderItem{
		OrderID:       order.ID,
		TicketClassID: ticketClass.ID,
		Quantity:      1,
		UnitPrice:     models.Money(5000),
		TotalPrice:    models.Money(5000),
	}
	require.NoError(t, db.Create(&orderItem).Error)

	// Create ticket
	ticket := models.Ticket{
		TicketNumber: "TKT-12345",
		OrderItemID:  orderItem.ID,
		Status:       models.TicketActive,
		HolderName:   "John Doe",
		HolderEmail:  "john.doe@example.com",
	}
	require.NoError(t, db.Create(&ticket).Error)

	// Create attendee
	attendee := models.Attendee{
		OrderID:                order.ID,
		EventID:                event.ID,
		TicketID:               ticket.ID,
		FirstName:              "John",
		LastName:               "Doe",
		Email:                  "john.doe@example.com",
		HasArrived:             false,
		AccountID:              account.ID,
		IsRefunded:             false,
		PrivateReferenceNumber: 12345,
	}
	require.NoError(t, db.Create(&attendee).Error)

	return account, event, order, ticket, attendee, ticketClass
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}

// TestNewAttendeeHandler tests the handler constructor
func TestNewAttendeeHandler(t *testing.T) {
	db := setupTestDB(t)
	metrics := &analytics.PrometheusMetrics{}

	handler := NewAttendeeHandler(db, metrics)

	assert.NotNil(t, handler)
	assert.Equal(t, db, handler.db)
}

// TestListAttendees_Success tests successful listing of attendees
func TestListAttendees_Success(t *testing.T) {
	db := setupTestDB(t)
	_, _, _, _, _, _ = setupTestData(t, db)
	handler := NewAttendeeHandler(db, nil)

	req := httptest.NewRequest(http.MethodGet, "/attendees?page=1&limit=10", nil)
	w := httptest.NewRecorder()

	handler.ListAttendees(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response AttendeeListResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, int64(1), response.TotalCount)
	assert.Equal(t, 1, len(response.Attendees))
	assert.Equal(t, "John", response.Attendees[0].FirstName)
	assert.Equal(t, "Doe", response.Attendees[0].LastName)
}

// TestListAttendees_WithEventFilter tests listing attendees filtered by event
func TestListAttendees_WithEventFilter(t *testing.T) {
	db := setupTestDB(t)
	_, event, _, _, _, _ := setupTestData(t, db)
	handler := NewAttendeeHandler(db, nil)

	req := httptest.NewRequest(http.MethodGet, "/attendees?event_id="+string(rune(event.ID)), nil)
	w := httptest.NewRecorder()

	handler.ListAttendees(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestListAttendees_WithSearchTerm tests searching attendees
func TestListAttendees_WithSearchTerm(t *testing.T) {
	db := setupTestDB(t)
	_, _, _, _, _, _ = setupTestData(t, db)
	handler := NewAttendeeHandler(db, nil)

	req := httptest.NewRequest(http.MethodGet, "/attendees?search=John", nil)
	w := httptest.NewRecorder()

	handler.ListAttendees(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response AttendeeListResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, int64(1), response.TotalCount)
}

// TestListAttendees_Pagination tests pagination
func TestListAttendees_Pagination(t *testing.T) {
	db := setupTestDB(t)
	account, event, order, _, _, ticketClass := setupTestData(t, db)
	handler := NewAttendeeHandler(db, nil)

	// Create additional attendees for pagination testing
	for i := 2; i <= 15; i++ {
		orderItem2 := models.OrderItem{
			OrderID:       order.ID,
			TicketClassID: ticketClass.ID,
			Quantity:      1,
			UnitPrice:     models.Money(5000),
			TotalPrice:    models.Money(5000),
		}
		db.Create(&orderItem2)

		ticket := models.Ticket{
			TicketNumber: "TKT-" + string(rune(i+48)),
			OrderItemID:  orderItem2.ID,
			Status:       models.TicketActive,
		}
		db.Create(&ticket)

		attendee := models.Attendee{
			OrderID:    order.ID,
			EventID:    event.ID,
			TicketID:   ticket.ID,
			FirstName:  "Attendee",
			LastName:   string(rune(i)),
			Email:      "attendee" + string(rune(i)) + "@example.com",
			AccountID:  account.ID,
			IsRefunded: false,
		}
		db.Create(&attendee)
	}

	// Test first page
	req := httptest.NewRequest(http.MethodGet, "/attendees?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	handler.ListAttendees(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response AttendeeListResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, 10, len(response.Attendees))
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 2, response.TotalPages)
}

// TestCheckInAttendee_Success tests successful check-in
func TestCheckInAttendee_Success(t *testing.T) {
	db := setupTestDB(t)
	account, _, _, ticket, attendee, _ := setupTestData(t, db)
	handler := NewAttendeeHandler(db, nil)

	reqBody := CheckInRequest{
		TicketNumber: ticket.TicketNumber,
		CheckedInBy:  account.ID,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/attendees/checkin", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.CheckInAttendee(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify attendee is checked in
	var updatedAttendee models.Attendee
	db.First(&updatedAttendee, attendee.ID)
	assert.True(t, updatedAttendee.HasArrived)
	assert.NotNil(t, updatedAttendee.ArrivalTime)
}

// TestCheckInAttendee_InvalidTicket tests check-in with invalid ticket
func TestCheckInAttendee_InvalidTicket(t *testing.T) {
	db := setupTestDB(t)
	account, _, _, _, _, _ := setupTestData(t, db)
	handler := NewAttendeeHandler(db, nil)

	reqBody := CheckInRequest{
		TicketNumber: "INVALID-TICKET",
		CheckedInBy:  account.ID,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/attendees/checkin", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.CheckInAttendee(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestCheckInAttendee_AlreadyCheckedIn tests double check-in prevention
func TestCheckInAttendee_AlreadyCheckedIn(t *testing.T) {
	db := setupTestDB(t)
	account, _, _, ticket, attendee, _ := setupTestData(t, db)
	handler := NewAttendeeHandler(db, nil)

	// First check-in
	now := time.Now()
	attendee.HasArrived = true
	attendee.ArrivalTime = &now
	db.Save(&attendee)

	// Try to check-in again
	reqBody := CheckInRequest{
		TicketNumber: ticket.TicketNumber,
		CheckedInBy:  account.ID,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/attendees/checkin", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.CheckInAttendee(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestCheckInAttendee_RefundedTicket tests check-in with refunded ticket
func TestCheckInAttendee_RefundedTicket(t *testing.T) {
	db := setupTestDB(t)
	account, _, _, ticket, attendee, _ := setupTestData(t, db)
	handler := NewAttendeeHandler(db, nil)

	// Mark as refunded
	attendee.IsRefunded = true
	db.Save(&attendee)

	reqBody := CheckInRequest{
		TicketNumber: ticket.TicketNumber,
		CheckedInBy:  account.ID,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/attendees/checkin", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.CheckInAttendee(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestCheckInAttendee_EmptyTicketNumber tests check-in with empty ticket number
func TestCheckInAttendee_EmptyTicketNumber(t *testing.T) {
	db := setupTestDB(t)
	account, _, _, _, _, _ := setupTestData(t, db)
	handler := NewAttendeeHandler(db, nil)

	reqBody := CheckInRequest{
		TicketNumber: "",
		CheckedInBy:  account.ID,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/attendees/checkin", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.CheckInAttendee(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestBulkCheckIn_Success tests successful bulk check-in
func TestBulkCheckIn_Success(t *testing.T) {
	db := setupTestDB(t)
	account, event, order, ticket, _, ticketClass := setupTestData(t, db)
	handler := NewAttendeeHandler(db, nil)

	// Create additional tickets and attendees
	orderItem2 := models.OrderItem{
		OrderID:       order.ID,
		TicketClassID: ticketClass.ID,
		Quantity:      1,
		UnitPrice:     models.Money(5000),
		TotalPrice:    models.Money(5000),
	}
	db.Create(&orderItem2)

	ticket2 := models.Ticket{
		TicketNumber: "TKT-67890",
		OrderItemID:  orderItem2.ID,
		Status:       models.TicketActive,
	}
	db.Create(&ticket2)

	attendee2 := models.Attendee{
		OrderID:    order.ID,
		EventID:    event.ID,
		TicketID:   ticket2.ID,
		FirstName:  "Jane",
		LastName:   "Smith",
		Email:      "jane.smith@example.com",
		AccountID:  account.ID,
		IsRefunded: false,
	}
	db.Create(&attendee2)

	reqBody := BulkCheckInRequest{
		TicketNumbers: []string{ticket.TicketNumber, ticket2.TicketNumber},
		CheckedInBy:   account.ID,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/attendees/bulk-checkin", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.BulkCheckIn(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestUpdateAttendeeInfo_Success tests successful attendee update
func TestUpdateAttendeeInfo_Success(t *testing.T) {
	db := setupTestDB(t)
	_, _, _, _, attendee, _ := setupTestData(t, db)
	handler := NewAttendeeHandler(db, nil)

	reqBody := UpdateAttendeeRequest{
		FirstName: "UpdatedJohn",
		LastName:  "UpdatedDoe",
		Email:     "updated@example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/attendees/"+string(rune(attendee.ID)), bytes.NewBuffer(body))
	req = mux.SetURLVars(req, map[string]string{"id": string(rune(attendee.ID))})
	w := httptest.NewRecorder()

	handler.UpdateAttendeeInfo(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify update
	var updatedAttendee models.Attendee
	db.First(&updatedAttendee, attendee.ID)
	assert.Equal(t, "UpdatedJohn", updatedAttendee.FirstName)
	assert.Equal(t, "UpdatedDoe", updatedAttendee.LastName)
	assert.Equal(t, "updated@example.com", updatedAttendee.Email)
}

// TestUpdateAttendeeInfo_NotFound tests updating non-existent attendee
func TestUpdateAttendeeInfo_NotFound(t *testing.T) {
	db := setupTestDB(t)
	handler := NewAttendeeHandler(db, nil)

	reqBody := UpdateAttendeeRequest{
		FirstName: "Updated",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/attendees/99999", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, map[string]string{"id": "99999"})
	w := httptest.NewRecorder()

	handler.UpdateAttendeeInfo(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestUpdateAttendeeInfo_InvalidID tests update with invalid ID
func TestUpdateAttendeeInfo_InvalidID(t *testing.T) {
	db := setupTestDB(t)
	handler := NewAttendeeHandler(db, nil)

	reqBody := UpdateAttendeeRequest{
		FirstName: "Updated",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/attendees/invalid", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, map[string]string{"id": "invalid"})
	w := httptest.NewRecorder()

	handler.UpdateAttendeeInfo(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGetAttendeeDetails_Success tests getting attendee details
func TestGetAttendeeDetails_Success(t *testing.T) {
	db := setupTestDB(t)
	_, _, _, _, attendee, _ := setupTestData(t, db)
	handler := NewAttendeeHandler(db, nil)

	req := httptest.NewRequest(http.MethodGet, "/attendees/"+string(rune(attendee.ID)), nil)
	req = mux.SetURLVars(req, map[string]string{"id": string(rune(attendee.ID))})
	w := httptest.NewRecorder()

	handler.GetAttendeeDetails(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response AttendeeResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "John", response.FirstName)
	assert.Equal(t, "Doe", response.LastName)
}

// TestGetAttendeeDetails_NotFound tests getting non-existent attendee
func TestGetAttendeeDetails_NotFound(t *testing.T) {
	db := setupTestDB(t)
	handler := NewAttendeeHandler(db, nil)

	req := httptest.NewRequest(http.MethodGet, "/attendees/99999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "99999"})
	w := httptest.NewRecorder()

	handler.GetAttendeeDetails(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestFilterAttendees_ByArrivalStatus tests filtering by arrival status
func TestFilterAttendees_ByArrivalStatus(t *testing.T) {
	db := setupTestDB(t)
	account, event, order, _, _, ticketClass := setupTestData(t, db)
	handler := NewAttendeeHandler(db, nil)

	// Create checked-in attendee
	orderItem2 := models.OrderItem{
		OrderID:       order.ID,
		TicketClassID: ticketClass.ID,
		Quantity:      1,
		UnitPrice:     models.Money(5000),
		TotalPrice:    models.Money(5000),
	}
	db.Create(&orderItem2)

	ticket2 := models.Ticket{
		TicketNumber: "TKT-ARRIVED",
		OrderItemID:  orderItem2.ID,
		Status:       models.TicketUsed,
	}
	db.Create(&ticket2)

	now := time.Now()
	attendee2 := models.Attendee{
		OrderID:     order.ID,
		EventID:     event.ID,
		TicketID:    ticket2.ID,
		FirstName:   "Arrived",
		LastName:    "User",
		Email:       "arrived@example.com",
		HasArrived:  true,
		ArrivalTime: &now,
		AccountID:   account.ID,
		IsRefunded:  false,
	}
	db.Create(&attendee2)

	req := httptest.NewRequest(http.MethodGet, "/attendees?has_arrived=true", nil)
	w := httptest.NewRecorder()

	handler.ListAttendees(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response AttendeeListResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, int64(1), response.TotalCount)
	assert.True(t, response.Attendees[0].HasArrived)
}

// TestFilterAttendees_ByRefundStatus tests filtering by refund status
func TestFilterAttendees_ByRefundStatus(t *testing.T) {
	db := setupTestDB(t)
	account, event, order, _, attendee, ticketClass := setupTestData(t, db)
	handler := NewAttendeeHandler(db, nil)

	// Mark first attendee as refunded
	attendee.IsRefunded = true
	db.Save(&attendee)

	// Create non-refunded attendee
	orderItem2 := models.OrderItem{
		OrderID:       order.ID,
		TicketClassID: ticketClass.ID,
		Quantity:      1,
		UnitPrice:     models.Money(5000),
		TotalPrice:    models.Money(5000),
	}
	db.Create(&orderItem2)

	ticket2 := models.Ticket{
		TicketNumber: "TKT-ACTIVE",
		OrderItemID:  orderItem2.ID,
		Status:       models.TicketActive,
	}
	db.Create(&ticket2)

	attendee2 := models.Attendee{
		OrderID:    order.ID,
		EventID:    event.ID,
		TicketID:   ticket2.ID,
		FirstName:  "Active",
		LastName:   "User",
		Email:      "active@example.com",
		AccountID:  account.ID,
		IsRefunded: false,
	}
	db.Create(&attendee2)

	req := httptest.NewRequest(http.MethodGet, "/attendees?is_refunded=false", nil)
	w := httptest.NewRecorder()

	handler.ListAttendees(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response AttendeeListResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, int64(1), response.TotalCount)
	assert.False(t, response.Attendees[0].IsRefunded)
}

// TestGetAttendeeCount_Success tests getting attendee count
func TestGetAttendeeCount_Success(t *testing.T) {
	db := setupTestDB(t)
	_, event, _, _, _, _ := setupTestData(t, db)
	handler := NewAttendeeHandler(db, nil)

	req := httptest.NewRequest(http.MethodGet, "/attendees/count?event_id="+string(rune(event.ID)), nil)
	w := httptest.NewRecorder()

	handler.GetAttendeeCount(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestConvertToAttendeeResponse tests the conversion function
func TestConvertToAttendeeResponse(t *testing.T) {
	now := time.Now()
	attendee := models.Attendee{
		Model:                  gorm.Model{ID: 1, CreatedAt: now},
		OrderID:                1,
		EventID:                1,
		TicketID:               1,
		FirstName:              "Test",
		LastName:               "User",
		Email:                  "test@example.com",
		HasArrived:             true,
		ArrivalTime:            &now,
		AccountID:              1,
		IsRefunded:             false,
		PrivateReferenceNumber: 12345,
		Event: models.Event{
			Model: gorm.Model{ID: 1},
			Title: "Test Event",
		},
		Ticket: models.Ticket{
			Model:        gorm.Model{ID: 1},
			TicketNumber: "TKT-TEST",
			OrderItem: models.OrderItem{
				TicketClass: models.TicketClass{
					Name: "VIP",
				},
			},
		},
	}

	response := convertToAttendeeResponse(attendee)

	assert.Equal(t, uint(1), response.ID)
	assert.Equal(t, "Test", response.FirstName)
	assert.Equal(t, "User", response.LastName)
	assert.Equal(t, "test@example.com", response.Email)
	assert.True(t, response.HasArrived)
	assert.NotNil(t, response.ArrivalTime)
	assert.Equal(t, "Test Event", response.EventTitle)
	assert.Equal(t, "TKT-TEST", response.TicketNumber)
	assert.Equal(t, "VIP", response.TicketClassName)
}

// TestCheckInAttendee_InactiveTicket tests check-in with inactive ticket
func TestCheckInAttendee_InactiveTicket(t *testing.T) {
	db := setupTestDB(t)
	account, _, _, ticket, _, _ := setupTestData(t, db)
	handler := NewAttendeeHandler(db, nil)

	// Mark ticket as cancelled
	ticket.Status = models.TicketCancelled
	db.Save(&ticket)

	reqBody := CheckInRequest{
		TicketNumber: ticket.TicketNumber,
		CheckedInBy:  account.ID,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/attendees/checkin", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.CheckInAttendee(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestListAttendees_EmptyDatabase tests listing with no attendees
func TestListAttendees_EmptyDatabase(t *testing.T) {
	db := setupTestDB(t)
	handler := NewAttendeeHandler(db, nil)

	req := httptest.NewRequest(http.MethodGet, "/attendees?page=1&limit=10", nil)
	w := httptest.NewRecorder()

	handler.ListAttendees(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response AttendeeListResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, int64(0), response.TotalCount)
	assert.Equal(t, 0, len(response.Attendees))
}

// TestBulkCheckIn_PartialSuccess tests bulk check-in with some failures
func TestBulkCheckIn_PartialSuccess(t *testing.T) {
	db := setupTestDB(t)
	account, _, _, ticket, _, _ := setupTestData(t, db)
	handler := NewAttendeeHandler(db, nil)

	reqBody := BulkCheckInRequest{
		TicketNumbers: []string{ticket.TicketNumber, "INVALID-TICKET"},
		CheckedInBy:   account.ID,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/attendees/bulk-checkin", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.BulkCheckIn(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestUpdateAttendeeInfo_PartialUpdate tests partial field update
func TestUpdateAttendeeInfo_PartialUpdate(t *testing.T) {
	db := setupTestDB(t)
	_, _, _, _, attendee, _ := setupTestData(t, db)
	handler := NewAttendeeHandler(db, nil)

	originalEmail := attendee.Email

	reqBody := UpdateAttendeeRequest{
		FirstName: "NewFirstName",
		// LastName and Email not provided
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/attendees/"+string(rune(attendee.ID)), bytes.NewBuffer(body))
	req = mux.SetURLVars(req, map[string]string{"id": string(rune(attendee.ID))})
	w := httptest.NewRecorder()

	handler.UpdateAttendeeInfo(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify only first name was updated
	var updatedAttendee models.Attendee
	db.First(&updatedAttendee, attendee.ID)
	assert.Equal(t, "NewFirstName", updatedAttendee.FirstName)
	assert.Equal(t, "Doe", updatedAttendee.LastName) // Original value
	assert.Equal(t, originalEmail, updatedAttendee.Email)
}
