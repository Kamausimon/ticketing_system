package models

import (
	"time"

	"gorm.io/gorm"
)

type EventStatus string

const (
	EventDraft     EventStatus = "draft"
	EventPending   EventStatus = "pending_approval"
	EventLive      EventStatus = "live"
	EventCancelled EventStatus = "cancelled"
	EventCompleted EventStatus = "completed"
)

type EventCategory string

const (
	CategoryMusic            EventCategory = "music"
	CategoryConference       EventCategory = "conference"
	CategorySeminar          EventCategory = "seminar"
	CategoryTradeShow        EventCategory = "TradeShow"
	CategoryProductLaunch    EventCategory = "Product Launch"
	CategoryTeamBuilding     EventCategory = "Team Building"
	CategoryCorporateMeeting EventCategory = "Corporate Meeting"
	CategoryCorporateRetreat EventCategory = "Corporate Retreat"
	CategorySports           EventCategory = "sports"
	CategoryEducational      EventCategory = "educational"
	CategoryFestival         EventCategory = "festival"
	CategoryArt              EventCategory = "art"
)

type Event struct {
	gorm.Model
	Title                   string `gorm:"not null"`
	Location                string `gorm:"not null"`
	BgType                  string
	BgColor                 string
	EventImages             []EventImages
	Description             string    `gorm:"not null"`
	StartDate               time.Time `gorm:"not null"`
	EndDate                 time.Time `gorm:"not null"`
	OnSaleDate              *time.Time
	OrganizerID             uint      `gorm:"not null;index"`
	Organizer               Organizer `gorm:"foreignKey:OrganizerID"`
	AccountID               uint      `gorm:"not null;index"`
	Account                 Account   `gorm:"foreignKey:AccountID"`
	SalesVolume             float32
	OrganizerFeesVolume     float32
	OrganizerFeeFixed       float32
	OrganizerFeePercentage  float32
	Currency                string  `gorm:"not null"`
	Venue                   []Venue `gorm:"many2many:event_venues"`
	LocationAddress         *string
	LocationAddressLine     *string
	LocationCountry         *string
	PreOrderMessageDisplay  *string
	PostOrderMessageDisplay *string
	IsLive                  bool   `gorm:"not null;default:false"`
	BarcodeType             string `gorm:"not null"`
	IsBarcodeEnabled        bool   `gorm:"not null;default:false"`
	TicketBorderColor       string
	TicketBgColor           string
	TicketTextColor         string
	TicketSubTextColor      string
	EnableOfflinePayment    bool `gorm:"not null;default:false"`
	MaxCapacity             *int
	Status                  EventStatus   `gorm:"default:'draft';not null"`
	Category                EventCategory `gorm:"index;not null"`
	Tags                    string
	MinAge                  *int
	IsPrivate               bool `gorm:"default:false"`
}
