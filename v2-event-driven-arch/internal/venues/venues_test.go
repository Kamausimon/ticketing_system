package venues

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
		&models.Venue{},
		&models.Event{},
		&models.EventVenues{},
		&models.Account{},
		&models.Organizer{},
	)
	require.NoError(t, err)

	return db
}

// setupTestData creates sample venue data for testing
func setupTestVenue(t *testing.T, db *gorm.DB) models.Venue {
	parkingCapacity := 200
	contactEmail := "venue@example.com"
	contactPhone := "+1234567890"
	website := "https://venue.example.com"

	venue := models.Venue{
		VenueName:        "Test Arena",
		VenueType:        models.VenueSportsArena,
		VenueCapacity:    5000,
		VenueSection:     "Main Hall",
		VenueLocation:    "Downtown",
		Address:          "123 Test Street",
		City:             "Test City",
		State:            "Test State",
		Country:          "Test Country",
		ZipCode:          "12345",
		ParkingAvailable: true,
		ParkingCapacity:  &parkingCapacity,
		IsAccessible:     true,
		HasWifi:          true,
		HasCatering:      true,
		ContactEmail:     &contactEmail,
		ContactPhone:     &contactPhone,
		Website:          &website,
	}
	require.NoError(t, db.Create(&venue).Error)

	return venue
}

// TestNewVenueHandler tests the handler constructor
func TestNewVenueHandler(t *testing.T) {
	db := setupTestDB(t)
	metrics := &analytics.PrometheusMetrics{}

	handler := NewVenueHandler(db, metrics)

	assert.NotNil(t, handler)
	assert.Equal(t, db, handler.db)
}

// TestCreateVenue_Success tests successful venue creation
func TestCreateVenue_Success(t *testing.T) {
	db := setupTestDB(t)
	handler := NewVenueHandler(db, nil)

	parkingCapacity := 100
	contactEmail := "contact@venue.com"
	contactPhone := "+1-555-0100"
	website := "https://myvenue.com"

	reqBody := CreateVenueRequest{
		VenueName:        "New Venue",
		VenueType:        models.VenueIndoor,
		VenueCapacity:    1000,
		VenueSection:     "Section A",
		VenueLocation:    "City Center",
		Address:          "456 New St",
		City:             "New City",
		State:            "NC",
		Country:          "USA",
		ZipCode:          "54321",
		ParkingAvailable: true,
		ParkingCapacity:  &parkingCapacity,
		IsAccessible:     true,
		HasWifi:          true,
		HasCatering:      false,
		ContactEmail:     &contactEmail,
		ContactPhone:     &contactPhone,
		Website:          &website,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/venues", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.CreateVenue(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response VenueResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "New Venue", response.VenueName)
	assert.Equal(t, models.VenueIndoor, response.VenueType)
	assert.Equal(t, 1000, response.VenueCapacity)
	assert.True(t, response.ParkingAvailable)
}

// TestCreateVenue_InvalidPayload tests creation with invalid payload
func TestCreateVenue_InvalidPayload(t *testing.T) {
	db := setupTestDB(t)
	handler := NewVenueHandler(db, nil)

	req := httptest.NewRequest(http.MethodPost, "/venues", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()

	handler.CreateVenue(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestCreateVenue_MissingRequiredFields tests validation
func TestCreateVenue_MissingRequiredFields(t *testing.T) {
	db := setupTestDB(t)
	handler := NewVenueHandler(db, nil)

	testCases := []struct {
		name    string
		request CreateVenueRequest
	}{
		{
			name: "Missing venue name",
			request: CreateVenueRequest{
				VenueName:     "",
				VenueType:     models.VenueIndoor,
				VenueCapacity: 1000,
				VenueLocation: "Location",
				City:          "City",
				State:         "State",
				Country:       "Country",
			},
		},
		{
			name: "Zero capacity",
			request: CreateVenueRequest{
				VenueName:     "Test Venue",
				VenueType:     models.VenueIndoor,
				VenueCapacity: 0,
				VenueLocation: "Location",
				City:          "City",
				State:         "State",
				Country:       "Country",
			},
		},
		{
			name: "Negative capacity",
			request: CreateVenueRequest{
				VenueName:     "Test Venue",
				VenueType:     models.VenueIndoor,
				VenueCapacity: -100,
				VenueLocation: "Location",
				City:          "City",
				State:         "State",
				Country:       "Country",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.request)
			req := httptest.NewRequest(http.MethodPost, "/venues", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.CreateVenue(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

// TestListVenues_Success tests successful venue listing
func TestListVenues_Success(t *testing.T) {
	db := setupTestDB(t)
	_ = setupTestVenue(t, db)
	handler := NewVenueHandler(db, nil)

	req := httptest.NewRequest(http.MethodGet, "/venues?page=1&page_size=10", nil)
	w := httptest.NewRecorder()

	handler.ListVenues(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response VenueListResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, int64(1), response.Total)
	assert.Equal(t, 1, len(response.Venues))
	assert.Equal(t, "Test Arena", response.Venues[0].VenueName)
}

// TestListVenues_EmptyDatabase tests listing with no venues
func TestListVenues_EmptyDatabase(t *testing.T) {
	db := setupTestDB(t)
	handler := NewVenueHandler(db, nil)

	req := httptest.NewRequest(http.MethodGet, "/venues?page=1&page_size=10", nil)
	w := httptest.NewRecorder()

	handler.ListVenues(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response VenueListResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, int64(0), response.Total)
	assert.Equal(t, 0, len(response.Venues))
}

// TestListVenues_Pagination tests pagination
func TestListVenues_Pagination(t *testing.T) {
	db := setupTestDB(t)
	handler := NewVenueHandler(db, nil)

	// Create multiple venues
	for i := 1; i <= 15; i++ {
		venue := models.Venue{
			VenueName:     "Venue " + string(rune(i)),
			VenueType:     models.VenueIndoor,
			VenueCapacity: 1000 * i,
			VenueLocation: "Location " + string(rune(i)),
			City:          "City",
			State:         "State",
			Country:       "Country",
		}
		db.Create(&venue)
	}

	// Test first page
	req := httptest.NewRequest(http.MethodGet, "/venues?page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	handler.ListVenues(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response VenueListResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, int64(15), response.Total)
	assert.Equal(t, 10, len(response.Venues))
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 2, response.TotalPages)

	// Test second page
	req = httptest.NewRequest(http.MethodGet, "/venues?page=2&page_size=10", nil)
	w = httptest.NewRecorder()
	handler.ListVenues(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, 5, len(response.Venues))
	assert.Equal(t, 2, response.Page)
}

// TestGetVenueDetails_Success tests getting venue details
func TestGetVenueDetails_Success(t *testing.T) {
	db := setupTestDB(t)
	venue := setupTestVenue(t, db)
	handler := NewVenueHandler(db, nil)

	req := httptest.NewRequest(http.MethodGet, "/venues/"+string(rune(venue.ID)), nil)
	req = mux.SetURLVars(req, map[string]string{"id": string(rune(venue.ID))})
	w := httptest.NewRecorder()

	handler.GetVenueDetails(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response VenueResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Test Arena", response.VenueName)
	assert.Equal(t, models.VenueSportsArena, response.VenueType)
	assert.Equal(t, 5000, response.VenueCapacity)
}

// TestGetVenueDetails_NotFound tests getting non-existent venue
func TestGetVenueDetails_NotFound(t *testing.T) {
	db := setupTestDB(t)
	handler := NewVenueHandler(db, nil)

	req := httptest.NewRequest(http.MethodGet, "/venues/99999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "99999"})
	w := httptest.NewRecorder()

	handler.GetVenueDetails(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestGetVenueDetails_InvalidID tests getting venue with invalid ID
func TestGetVenueDetails_InvalidID(t *testing.T) {
	db := setupTestDB(t)
	handler := NewVenueHandler(db, nil)

	req := httptest.NewRequest(http.MethodGet, "/venues/invalid", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "invalid"})
	w := httptest.NewRecorder()

	handler.GetVenueDetails(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestUpdateVenue_Success tests successful venue update
func TestUpdateVenue_Success(t *testing.T) {
	db := setupTestDB(t)
	venue := setupTestVenue(t, db)
	handler := NewVenueHandler(db, nil)

	newName := "Updated Arena"
	newCapacity := 6000
	reqBody := UpdateVenueRequest{
		VenueName:     &newName,
		VenueCapacity: &newCapacity,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/venues/"+string(rune(venue.ID)), bytes.NewBuffer(body))
	req = mux.SetURLVars(req, map[string]string{"id": string(rune(venue.ID))})
	w := httptest.NewRecorder()

	handler.UpdateVenue(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify update
	var updatedVenue models.Venue
	db.First(&updatedVenue, venue.ID)
	assert.Equal(t, "Updated Arena", updatedVenue.VenueName)
	assert.Equal(t, 6000, updatedVenue.VenueCapacity)
}

// TestUpdateVenue_PartialUpdate tests partial field update
func TestUpdateVenue_PartialUpdate(t *testing.T) {
	db := setupTestDB(t)
	venue := setupTestVenue(t, db)
	handler := NewVenueHandler(db, nil)

	originalCapacity := venue.VenueCapacity
	newName := "Only Name Updated"

	reqBody := UpdateVenueRequest{
		VenueName: &newName,
		// Other fields not provided
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/venues/"+string(rune(venue.ID)), bytes.NewBuffer(body))
	req = mux.SetURLVars(req, map[string]string{"id": string(rune(venue.ID))})
	w := httptest.NewRecorder()

	handler.UpdateVenue(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify only name was updated
	var updatedVenue models.Venue
	db.First(&updatedVenue, venue.ID)
	assert.Equal(t, "Only Name Updated", updatedVenue.VenueName)
	assert.Equal(t, originalCapacity, updatedVenue.VenueCapacity)
}

// TestUpdateVenue_NotFound tests updating non-existent venue
func TestUpdateVenue_NotFound(t *testing.T) {
	db := setupTestDB(t)
	handler := NewVenueHandler(db, nil)

	newName := "Updated"
	reqBody := UpdateVenueRequest{
		VenueName: &newName,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/venues/99999", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, map[string]string{"id": "99999"})
	w := httptest.NewRecorder()

	handler.UpdateVenue(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestUpdateVenue_InvalidID tests update with invalid ID
func TestUpdateVenue_InvalidID(t *testing.T) {
	db := setupTestDB(t)
	handler := NewVenueHandler(db, nil)

	newName := "Updated"
	reqBody := UpdateVenueRequest{
		VenueName: &newName,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/venues/invalid", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, map[string]string{"id": "invalid"})
	w := httptest.NewRecorder()

	handler.UpdateVenue(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestUpdateVenue_InvalidPayload tests update with invalid payload
func TestUpdateVenue_InvalidPayload(t *testing.T) {
	db := setupTestDB(t)
	venue := setupTestVenue(t, db)
	handler := NewVenueHandler(db, nil)

	req := httptest.NewRequest(http.MethodPut, "/venues/"+string(rune(venue.ID)), bytes.NewBufferString("invalid json"))
	req = mux.SetURLVars(req, map[string]string{"id": string(rune(venue.ID))})
	w := httptest.NewRecorder()

	handler.UpdateVenue(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestDeleteVenue_Success tests successful venue deletion
func TestDeleteVenue_Success(t *testing.T) {
	db := setupTestDB(t)
	venue := setupTestVenue(t, db)
	handler := NewVenueHandler(db, nil)

	req := httptest.NewRequest(http.MethodDelete, "/venues/"+string(rune(venue.ID)), nil)
	req = mux.SetURLVars(req, map[string]string{"id": string(rune(venue.ID))})
	w := httptest.NewRecorder()

	handler.DeleteVenue(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify venue is soft deleted
	var deletedVenue models.Venue
	err := db.First(&deletedVenue, venue.ID).Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

// TestDeleteVenue_NotFound tests deleting non-existent venue
func TestDeleteVenue_NotFound(t *testing.T) {
	db := setupTestDB(t)
	handler := NewVenueHandler(db, nil)

	req := httptest.NewRequest(http.MethodDelete, "/venues/99999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "99999"})
	w := httptest.NewRecorder()

	handler.DeleteVenue(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestDeleteVenue_WithUpcomingEvents tests deletion prevention with upcoming events
func TestDeleteVenue_WithUpcomingEvents(t *testing.T) {
	db := setupTestDB(t)
	venue := setupTestVenue(t, db)
	handler := NewVenueHandler(db, nil)

	// Create account for organizer
	account := models.Account{
		Email:     "organizer@example.com",
		FirstName: "Test",
		LastName:  "Organizer",
	}
	db.Create(&account)

	// Create organizer
	organizer := models.Organizer{
		Name:      "Test Organizer",
		AccountID: account.ID,
	}
	db.Create(&organizer)

	// Create upcoming event
	futureDate := time.Now().Add(48 * time.Hour)
	endDate := futureDate.Add(2 * time.Hour)
	event := models.Event{
		Title:       "Upcoming Event",
		Description: "Description",
		StartDate:   futureDate,
		EndDate:     endDate,
		OrganizerID: organizer.ID,
		AccountID:   account.ID,
		Location:    "Test Location",
		Currency:    "USD",
		Category:    models.CategoryConference,
		BarcodeType: "QR",
	}
	db.Create(&event)

	// Link event to venue
	eventVenue := models.EventVenues{
		EventID: event.ID,
		VenueID: venue.ID,
	}
	db.Create(&eventVenue)

	req := httptest.NewRequest(http.MethodDelete, "/venues/"+string(rune(venue.ID)), nil)
	req = mux.SetURLVars(req, map[string]string{"id": string(rune(venue.ID))})
	w := httptest.NewRecorder()

	handler.DeleteVenue(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	// Verify venue still exists
	var existingVenue models.Venue
	err := db.First(&existingVenue, venue.ID).Error
	assert.NoError(t, err)
}

// TestDeleteVenue_InvalidID tests deletion with invalid ID
func TestDeleteVenue_InvalidID(t *testing.T) {
	db := setupTestDB(t)
	handler := NewVenueHandler(db, nil)

	req := httptest.NewRequest(http.MethodDelete, "/venues/invalid", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "invalid"})
	w := httptest.NewRecorder()

	handler.DeleteVenue(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestConvertToVenueResponse tests the conversion function
func TestConvertToVenueResponse(t *testing.T) {
	parkingCapacity := 150
	contactEmail := "test@venue.com"
	contactPhone := "+1234567890"
	website := "https://test.com"

	venue := models.Venue{
		Model:            gorm.Model{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		VenueName:        "Test Venue",
		VenueType:        models.VenueIndoor,
		VenueCapacity:    2000,
		VenueSection:     "Section B",
		VenueLocation:    "Downtown",
		Address:          "123 Test St",
		City:             "Test City",
		State:            "TS",
		Country:          "Test Country",
		ZipCode:          "12345",
		ParkingAvailable: true,
		ParkingCapacity:  &parkingCapacity,
		IsAccessible:     true,
		HasWifi:          true,
		HasCatering:      false,
		ContactEmail:     &contactEmail,
		ContactPhone:     &contactPhone,
		Website:          &website,
	}

	response := convertToVenueResponse(&venue)

	assert.Equal(t, uint(1), response.ID)
	assert.Equal(t, "Test Venue", response.VenueName)
	assert.Equal(t, models.VenueIndoor, response.VenueType)
	assert.Equal(t, 2000, response.VenueCapacity)
	assert.Equal(t, "Section B", response.VenueSection)
	assert.Equal(t, "Downtown", response.VenueLocation)
	assert.True(t, response.ParkingAvailable)
	assert.NotNil(t, response.ParkingCapacity)
	assert.Equal(t, 150, *response.ParkingCapacity)
	assert.True(t, response.IsAccessible)
	assert.True(t, response.HasWifi)
	assert.False(t, response.HasCatering)
}

// TestListVenues_WithSearchFilter tests search functionality
func TestListVenues_WithSearchFilter(t *testing.T) {
	db := setupTestDB(t)
	handler := NewVenueHandler(db, nil)

	// Create venues with different names
	venue1 := models.Venue{
		VenueName:     "Convention Center",
		VenueType:     models.VenueConference,
		VenueCapacity: 3000,
		VenueLocation: "Location 1",
		City:          "City A",
		State:         "State",
		Country:       "Country",
	}
	db.Create(&venue1)

	venue2 := models.Venue{
		VenueName:     "Sports Arena",
		VenueType:     models.VenueSportsArena,
		VenueCapacity: 8000,
		VenueLocation: "Location 2",
		City:          "City B",
		State:         "State",
		Country:       "Country",
	}
	db.Create(&venue2)

	req := httptest.NewRequest(http.MethodGet, "/venues?search=Arena", nil)
	w := httptest.NewRecorder()

	handler.ListVenues(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response VenueListResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Should only return the Sports Arena
	assert.Equal(t, int64(1), response.Total)
	if len(response.Venues) > 0 {
		assert.Equal(t, "Sports Arena", response.Venues[0].VenueName)
	}
}

// TestListVenues_FilterByCity tests filtering by city
func TestListVenues_FilterByCity(t *testing.T) {
	db := setupTestDB(t)
	handler := NewVenueHandler(db, nil)

	// Create venues in different cities
	venue1 := models.Venue{
		VenueName:     "Venue NYC",
		VenueType:     models.VenueIndoor,
		VenueCapacity: 1000,
		VenueLocation: "Manhattan",
		City:          "New York",
		State:         "NY",
		Country:       "USA",
	}
	db.Create(&venue1)

	venue2 := models.Venue{
		VenueName:     "Venue LA",
		VenueType:     models.VenueOutdoor,
		VenueCapacity: 2000,
		VenueLocation: "Hollywood",
		City:          "Los Angeles",
		State:         "CA",
		Country:       "USA",
	}
	db.Create(&venue2)

	req := httptest.NewRequest(http.MethodGet, "/venues?city=New York", nil)
	w := httptest.NewRecorder()

	handler.ListVenues(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response VenueListResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, int64(1), response.Total)
	if len(response.Venues) > 0 {
		assert.Equal(t, "New York", response.Venues[0].City)
	}
}

// TestListVenues_FilterByVenueType tests filtering by venue type
func TestListVenues_FilterByVenueType(t *testing.T) {
	db := setupTestDB(t)
	handler := NewVenueHandler(db, nil)

	// Create venues of different types
	venue1 := models.Venue{
		VenueName:     "Indoor Venue",
		VenueType:     models.VenueIndoor,
		VenueCapacity: 1000,
		VenueLocation: "Location",
		City:          "City",
		State:         "State",
		Country:       "Country",
	}
	db.Create(&venue1)

	venue2 := models.Venue{
		VenueName:     "Outdoor Venue",
		VenueType:     models.VenueOutdoor,
		VenueCapacity: 2000,
		VenueLocation: "Location",
		City:          "City",
		State:         "State",
		Country:       "Country",
	}
	db.Create(&venue2)

	req := httptest.NewRequest(http.MethodGet, "/venues?venue_type=indoor", nil)
	w := httptest.NewRecorder()

	handler.ListVenues(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response VenueListResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, int64(1), response.Total)
	if len(response.Venues) > 0 {
		assert.Equal(t, models.VenueIndoor, response.Venues[0].VenueType)
	}
}

// TestCreateVenue_AllVenueTypes tests creation with all venue types
func TestCreateVenue_AllVenueTypes(t *testing.T) {
	db := setupTestDB(t)
	handler := NewVenueHandler(db, nil)

	venueTypes := []models.VenueType{
		models.VenueIndoor,
		models.VenueOutdoor,
		models.VenueSportsArena,
		models.VenueConference,
		models.VenueConvectionCenter,
		models.VenueHotel,
		models.VenueResort,
		models.VenueBreweryWinery,
		models.VenueRestaurant,
		models.VenueTheatre,
	}

	for _, vt := range venueTypes {
		t.Run(string(vt), func(t *testing.T) {
			reqBody := CreateVenueRequest{
				VenueName:        "Test " + string(vt),
				VenueType:        vt,
				VenueCapacity:    500,
				VenueLocation:    "Test Location",
				City:             "Test City",
				State:            "TS",
				Country:          "Test Country",
				ParkingAvailable: true,
				IsAccessible:     true,
				HasWifi:          false,
				HasCatering:      false,
			}

			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/venues", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.CreateVenue(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			var response VenueResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, vt, response.VenueType)
		})
	}
}

// TestUpdateVenue_AllFields tests updating all venue fields
func TestUpdateVenue_AllFields(t *testing.T) {
	db := setupTestDB(t)
	venue := setupTestVenue(t, db)
	handler := NewVenueHandler(db, nil)

	newName := "Completely Updated"
	newType := models.VenueTheatre
	newCapacity := 3000
	newSection := "Updated Section"
	newLocation := "Updated Location"
	newAddress := "789 Updated St"
	newCity := "Updated City"
	newState := "UC"
	newCountry := "Updated Country"
	newZip := "99999"
	parking := false
	accessible := false
	wifi := false
	catering := true
	newEmail := "new@email.com"
	newPhone := "+9999999999"
	newWebsite := "https://new.com"

	reqBody := UpdateVenueRequest{
		VenueName:        &newName,
		VenueType:        &newType,
		VenueCapacity:    &newCapacity,
		VenueSection:     &newSection,
		VenueLocation:    &newLocation,
		Address:          &newAddress,
		City:             &newCity,
		State:            &newState,
		Country:          &newCountry,
		ZipCode:          &newZip,
		ParkingAvailable: &parking,
		IsAccessible:     &accessible,
		HasWifi:          &wifi,
		HasCatering:      &catering,
		ContactEmail:     &newEmail,
		ContactPhone:     &newPhone,
		Website:          &newWebsite,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/venues/"+string(rune(venue.ID)), bytes.NewBuffer(body))
	req = mux.SetURLVars(req, map[string]string{"id": string(rune(venue.ID))})
	w := httptest.NewRecorder()

	handler.UpdateVenue(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify all updates
	var updatedVenue models.Venue
	db.First(&updatedVenue, venue.ID)
	assert.Equal(t, newName, updatedVenue.VenueName)
	assert.Equal(t, newType, updatedVenue.VenueType)
	assert.Equal(t, newCapacity, updatedVenue.VenueCapacity)
	assert.Equal(t, newCity, updatedVenue.City)
	assert.False(t, updatedVenue.ParkingAvailable)
	assert.True(t, updatedVenue.HasCatering)
}

// TestListVenues_DefaultPagination tests default pagination values
func TestListVenues_DefaultPagination(t *testing.T) {
	db := setupTestDB(t)
	_ = setupTestVenue(t, db)
	handler := NewVenueHandler(db, nil)

	req := httptest.NewRequest(http.MethodGet, "/venues", nil)
	w := httptest.NewRecorder()

	handler.ListVenues(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response VenueListResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Default values should be applied
	assert.Equal(t, 1, response.Page)
}
