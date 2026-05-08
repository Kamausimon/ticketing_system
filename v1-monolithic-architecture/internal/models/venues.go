package models

import "gorm.io/gorm"

type VenueType string

const (
	VenueIndoor           VenueType = "indoor"
	VenueOutdoor          VenueType = "outdoor"
	VenueSportsArena      VenueType = "sports_arena"
	VenueConference       VenueType = "conference_center"
	VenueConvectionCenter VenueType = "convection_center"
	VenueHotel            VenueType = "hotel"
	VenueResort           VenueType = "resort"
	VenueBreweryWinery    VenueType = "brewery_winery"
	VenueRestaurant       VenueType = "restaurant"
	VenueTheatre          VenueType = "theater"
)

type Venue struct {
	gorm.Model
	VenueName        string `gorm:"not null;index"`
	VenueCapacity    int    `gorm:"not null"`
	VenueSection     string
	VenueType        VenueType `gorm:"not null;index"`
	VenueLocation    string    `gorm:"not null"`
	Address          string
	City             string `gorm:"index"`
	State            string
	Country          string `gorm:"index"`
	ZipCode          string
	ParkingAvailable bool `gorm:"default:true"`
	ParkingCapacity  *int
	IsAccessible     bool `gorm:"default:true"`
	HasWifi          bool `gorm:"default:false"`
	HasCatering      bool `gorm:"default:false"`

	ContactEmail *string
	ContactPhone *string
	Website      *string
}
