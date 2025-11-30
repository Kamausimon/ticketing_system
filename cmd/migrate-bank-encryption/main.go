package main

import (
	"fmt"
	"os"
	"ticketing_system/internal/config"
	"ticketing_system/internal/database"
	"ticketing_system/internal/models"
	"ticketing_system/internal/security"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("🔐 Bank Details Encryption Migration Tool")
	fmt.Println("=========================================")
	fmt.Println()

	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("⚠️  No .env file found, using environment variables")
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize encryption service
	encryptionService, err := security.NewEncryptionService(cfg.Security.EncryptionKey)
	if err != nil {
		fmt.Printf("❌ Failed to initialize encryption service: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Encryption service initialized (key length: %d bytes)\n", len(cfg.Security.EncryptionKey))

	// Connect to database
	db := database.Init()
	fmt.Println("✅ Database connection established")
	fmt.Println()

	// Get all organizers with bank details
	var organizers []models.Organizer
	if err := db.Where("bank_account_number IS NOT NULL AND bank_account_number != ?", "").Find(&organizers).Error; err != nil {
		fmt.Printf("❌ Failed to fetch organizers: %v\n", err)
		os.Exit(1)
	}

	if len(organizers) == 0 {
		fmt.Println("ℹ️  No organizers with bank details found")
		fmt.Println("✅ Nothing to migrate")
		return
	}

	fmt.Printf("📊 Found %d organizer(s) with bank details\n", len(organizers))
	fmt.Println()

	// Ask for confirmation
	fmt.Print("⚠️  This will encrypt all bank account numbers and bank codes. Continue? (yes/no): ")
	var response string
	fmt.Scanln(&response)
	if response != "yes" {
		fmt.Println("❌ Migration cancelled")
		os.Exit(0)
	}

	fmt.Println()
	fmt.Println("🔄 Starting migration...")
	fmt.Println()

	// Migrate each organizer
	migrated := 0
	skipped := 0
	failed := 0

	for i, organizer := range organizers {
		fmt.Printf("[%d/%d] Processing organizer ID %d (%s)...\n",
			i+1, len(organizers), organizer.ID, organizer.Name)

		// Check if already encrypted
		// Encrypted data is longer (due to nonce + auth tag + base64) and contains base64 characters
		if len(organizer.BankAccountNumber) > 50 {
			fmt.Printf("  ⏭️  Skipped - appears already encrypted (length: %d)\n", len(organizer.BankAccountNumber))
			skipped++
			continue
		}

		// Store original values for verification
		originalNumber := organizer.BankAccountNumber
		originalCode := organizer.BankCode

		// Encrypt the data
		encryptedNumber, encryptedCode, err := encryptionService.EncryptBankDetails(
			organizer.BankAccountNumber,
			organizer.BankCode,
		)
		if err != nil {
			fmt.Printf("  ❌ Failed to encrypt: %v\n", err)
			failed++
			continue
		}

		// Update the database
		err = db.Model(&organizer).Updates(map[string]interface{}{
			"bank_account_number": encryptedNumber,
			"bank_code":           encryptedCode,
		}).Error
		if err != nil {
			fmt.Printf("  ❌ Failed to update database: %v\n", err)
			failed++
			continue
		}

		// Verify by decrypting
		decryptedNumber, decryptedCode, err := encryptionService.DecryptBankDetails(
			encryptedNumber,
			encryptedCode,
		)
		if err != nil {
			fmt.Printf("  ⚠️  WARNING: Verification failed - could not decrypt: %v\n", err)
			fmt.Printf("  ⚠️  Data may be corrupted for organizer ID %d\n", organizer.ID)
			failed++
			continue
		}

		if decryptedNumber != originalNumber || decryptedCode != originalCode {
			fmt.Printf("  ⚠️  WARNING: Decrypted data doesn't match original!\n")
			fmt.Printf("  ⚠️  Original: %s / %s\n", originalNumber, originalCode)
			fmt.Printf("  ⚠️  Decrypted: %s / %s\n", decryptedNumber, decryptedCode)
			failed++
			continue
		}

		maskedNumber := security.MaskBankAccountNumber(originalNumber)
		fmt.Printf("  ✅ Encrypted and verified: %s → %s (encrypted length: %d)\n",
			maskedNumber, maskedNumber, len(encryptedNumber))
		migrated++
	}

	// Summary
	fmt.Println()
	fmt.Println("=========================================")
	fmt.Println("📊 Migration Summary")
	fmt.Println("=========================================")
	fmt.Printf("Total organizers:    %d\n", len(organizers))
	fmt.Printf("✅ Successfully migrated: %d\n", migrated)
	fmt.Printf("⏭️  Skipped (already encrypted): %d\n", skipped)
	fmt.Printf("❌ Failed: %d\n", failed)
	fmt.Println()

	if failed > 0 {
		fmt.Println("⚠️  Some migrations failed. Please review the errors above.")
		os.Exit(1)
	} else if migrated > 0 {
		fmt.Println("✅ All bank details successfully encrypted!")
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("1. Test bank details retrieval via API")
		fmt.Println("2. Verify encrypted data in database")
		fmt.Println("3. Backup encryption key securely")
	} else {
		fmt.Println("ℹ️  All bank details were already encrypted")
	}
}
