package main

import (
	"fmt"
	"log"
	"ticketing_system/internal/database"
	"ticketing_system/internal/models"
)

func main() {
	fmt.Println("Checking database migration status...")

	// Initialize database connection
	db := database.Init()

	// Check if the users table exists and has the new fields
	var user models.User

	// Try to find the first user (this will also validate the schema)
	result := db.First(&user)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			fmt.Println("✅ Users table exists and schema is valid (no users found yet)")
		} else {
			fmt.Printf("❌ Error: %v\n", result.Error)
			log.Fatal("Migration might have failed")
		}
	} else {
		fmt.Println("✅ Users table exists and schema is valid")
		fmt.Printf("Found user: %s %s (ID: %d)\n", user.FirstName, user.LastName, user.ID)
	}

	// Check if we can describe the table structure
	var columns []struct {
		ColumnName string `gorm:"column:column_name"`
		DataType   string `gorm:"column:data_type"`
	}

	err := db.Raw(`
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'users' 
		ORDER BY ordinal_position
	`).Scan(&columns)

	if err.Error == nil {
		fmt.Println("\n📋 Users table structure:")
		for _, col := range columns {
			fmt.Printf("  - %s (%s)\n", col.ColumnName, col.DataType)
		}

		// Check for specific JWT fields
		jwtFields := []string{"refresh_token", "refresh_token_exp", "last_login_at", "token_version"}
		fmt.Println("\n🔑 JWT Token fields status:")

		for _, jwtField := range jwtFields {
			found := false
			for _, col := range columns {
				if col.ColumnName == jwtField {
					found = true
					break
				}
			}
			if found {
				fmt.Printf("  ✅ %s - Present\n", jwtField)
			} else {
				fmt.Printf("  ❌ %s - Missing\n", jwtField)
			}
		}
	} else {
		fmt.Printf("Error checking table structure: %v\n", err.Error)
	}

	fmt.Println("\n🎉 Migration check complete!")
}
