package seed

import (
	"log"
	"ticketing_system/internal/models"

	"gorm.io/gorm"
)

// SeedPreferencesData seeds timezones, currencies, date formats, and datetime formats
func SeedPreferencesData(db *gorm.DB) error {
	log.Println("Seeding preferences data...")

	// Seed Timezones
	if err := seedTimezones(db); err != nil {
		return err
	}

	// Seed Currencies
	if err := seedCurrencies(db); err != nil {
		return err
	}

	// Seed Date Formats
	if err := seedDateFormats(db); err != nil {
		return err
	}

	// Seed DateTime Formats
	if err := seedDateTimeFormats(db); err != nil {
		return err
	}

	log.Println("Preferences data seeded successfully")
	return nil
}

func seedTimezones(db *gorm.DB) error {
	timezones := []models.Timezone{
		{Name: "UTC", DisplayName: "UTC - Coordinated Universal Time", Offset: "+00:00", IanaName: "UTC", IsActive: true},
		{Name: "EAT", DisplayName: "East Africa Time (Nairobi, Kampala, Dar es Salaam)", Offset: "+03:00", IanaName: "Africa/Nairobi", IsActive: true},
		{Name: "WAT", DisplayName: "West Africa Time (Lagos, Accra)", Offset: "+01:00", IanaName: "Africa/Lagos", IsActive: true},
		{Name: "CAT", DisplayName: "Central Africa Time (Johannesburg)", Offset: "+02:00", IanaName: "Africa/Johannesburg", IsActive: true},
		{Name: "EST", DisplayName: "Eastern Standard Time (New York)", Offset: "-05:00", IanaName: "America/New_York", IsActive: true},
		{Name: "CST", DisplayName: "Central Standard Time (Chicago)", Offset: "-06:00", IanaName: "America/Chicago", IsActive: true},
		{Name: "MST", DisplayName: "Mountain Standard Time (Denver)", Offset: "-07:00", IanaName: "America/Denver", IsActive: true},
		{Name: "PST", DisplayName: "Pacific Standard Time (Los Angeles)", Offset: "-08:00", IanaName: "America/Los_Angeles", IsActive: true},
		{Name: "GMT", DisplayName: "Greenwich Mean Time (London)", Offset: "+00:00", IanaName: "Europe/London", IsActive: true},
		{Name: "CET", DisplayName: "Central European Time (Paris, Berlin)", Offset: "+01:00", IanaName: "Europe/Paris", IsActive: true},
		{Name: "EET", DisplayName: "Eastern European Time (Athens, Cairo)", Offset: "+02:00", IanaName: "Europe/Athens", IsActive: true},
		{Name: "IST", DisplayName: "India Standard Time (Mumbai, Delhi)", Offset: "+05:30", IanaName: "Asia/Kolkata", IsActive: true},
		{Name: "CST_CHINA", DisplayName: "China Standard Time (Beijing, Shanghai)", Offset: "+08:00", IanaName: "Asia/Shanghai", IsActive: true},
		{Name: "JST", DisplayName: "Japan Standard Time (Tokyo)", Offset: "+09:00", IanaName: "Asia/Tokyo", IsActive: true},
		{Name: "AEST", DisplayName: "Australian Eastern Standard Time (Sydney)", Offset: "+10:00", IanaName: "Australia/Sydney", IsActive: true},
		{Name: "NZST", DisplayName: "New Zealand Standard Time (Auckland)", Offset: "+12:00", IanaName: "Pacific/Auckland", IsActive: true},
	}

	for _, tz := range timezones {
		var existing models.Timezone
		if err := db.Where("name = ?", tz.Name).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&tz).Error; err != nil {
					log.Printf("Error creating timezone %s: %v", tz.Name, err)
					return err
				}
				log.Printf("Created timezone: %s", tz.Name)
			}
		}
	}

	return nil
}

func seedCurrencies(db *gorm.DB) error {
	currencies := []models.Currency{
		{Code: "USD", Name: "US Dollar", Symbol: "$", IsActive: true},
		{Code: "KSH", Name: "Kenyan Shilling", Symbol: "KSh", IsActive: true},
		{Code: "EUR", Name: "Euro", Symbol: "€", IsActive: true},
		{Code: "GBP", Name: "British Pound Sterling", Symbol: "£", IsActive: true},
		{Code: "NGN", Name: "Nigerian Naira", Symbol: "₦", IsActive: true},
		{Code: "ZAR", Name: "South African Rand", Symbol: "R", IsActive: true},
		{Code: "GHS", Name: "Ghanaian Cedi", Symbol: "GH₵", IsActive: true},
		{Code: "UGX", Name: "Ugandan Shilling", Symbol: "USh", IsActive: true},
		{Code: "TZS", Name: "Tanzanian Shilling", Symbol: "TSh", IsActive: true},
		{Code: "CAD", Name: "Canadian Dollar", Symbol: "C$", IsActive: true},
		{Code: "AUD", Name: "Australian Dollar", Symbol: "A$", IsActive: true},
		{Code: "INR", Name: "Indian Rupee", Symbol: "₹", IsActive: true},
		{Code: "JPY", Name: "Japanese Yen", Symbol: "¥", IsActive: true},
		{Code: "CNY", Name: "Chinese Yuan", Symbol: "¥", IsActive: true},
	}

	for _, currency := range currencies {
		var existing models.Currency
		if err := db.Where("code = ?", currency.Code).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&currency).Error; err != nil {
					log.Printf("Error creating currency %s: %v", currency.Code, err)
					return err
				}
				log.Printf("Created currency: %s", currency.Code)
			}
		}
	}

	return nil
}

func seedDateFormats(db *gorm.DB) error {
	dateFormats := []models.DateFormat{
		{Format: "YYYY-MM-DD", Example: "2024-12-25", IsActive: true},
		{Format: "DD/MM/YYYY", Example: "25/12/2024", IsActive: true},
		{Format: "MM/DD/YYYY", Example: "12/25/2024", IsActive: true},
		{Format: "DD-MM-YYYY", Example: "25-12-2024", IsActive: true},
		{Format: "MMM DD, YYYY", Example: "Dec 25, 2024", IsActive: true},
		{Format: "DD MMM YYYY", Example: "25 Dec 2024", IsActive: true},
		{Format: "YYYY/MM/DD", Example: "2024/12/25", IsActive: true},
		{Format: "DD.MM.YYYY", Example: "25.12.2024", IsActive: true},
	}

	for _, format := range dateFormats {
		var existing models.DateFormat
		if err := db.Where("format = ?", format.Format).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&format).Error; err != nil {
					log.Printf("Error creating date format %s: %v", format.Format, err)
					return err
				}
				log.Printf("Created date format: %s", format.Format)
			}
		}
	}

	return nil
}

func seedDateTimeFormats(db *gorm.DB) error {
	dateTimeFormats := []models.DateTimeFormat{
		{Format: "YYYY-MM-DD HH:mm", Example: "2024-12-25 14:30", IsActive: true},
		{Format: "DD/MM/YYYY HH:mm", Example: "25/12/2024 14:30", IsActive: true},
		{Format: "MM/DD/YYYY hh:mm A", Example: "12/25/2024 02:30 PM", IsActive: true},
		{Format: "DD-MM-YYYY HH:mm", Example: "25-12-2024 14:30", IsActive: true},
		{Format: "MMM DD, YYYY hh:mm A", Example: "Dec 25, 2024 02:30 PM", IsActive: true},
		{Format: "DD MMM YYYY HH:mm", Example: "25 Dec 2024 14:30", IsActive: true},
		{Format: "YYYY/MM/DD HH:mm", Example: "2024/12/25 14:30", IsActive: true},
		{Format: "DD.MM.YYYY HH:mm", Example: "25.12.2024 14:30", IsActive: true},
	}

	for _, format := range dateTimeFormats {
		var existing models.DateTimeFormat
		if err := db.Where("format = ?", format.Format).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&format).Error; err != nil {
					log.Printf("Error creating datetime format %s: %v", format.Format, err)
					return err
				}
				log.Printf("Created datetime format: %s", format.Format)
			}
		}
	}

	return nil
}
