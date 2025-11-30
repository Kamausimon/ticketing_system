# Bank Details Encryption - Migration Guide

## Overview
This document provides instructions for migrating existing plaintext bank details to encrypted format.

## Security Implementation

### What's Encrypted
- **Bank Account Number** - Fully encrypted using AES-256-GCM
- **Bank Code** (SWIFT/Routing Code) - Fully encrypted using AES-256-GCM

### What's NOT Encrypted
- **Bank Account Name** - Public-facing business name
- **Bank Country** - Not sensitive information

## Environment Configuration

### Required Environment Variable

Add the following to your `.env` file:

```env
# Encryption key for sensitive data (MUST be 16, 24, or 32 bytes)
# For AES-256 (recommended), use exactly 32 bytes
ENCRYPTION_KEY=your-32-byte-secret-key-here!!
```

### Generating a Secure Encryption Key

**Production (32-byte key for AES-256):**
```bash
# Using OpenSSL
openssl rand -base64 32 | head -c 32

# Using Python
python3 -c "import secrets; print(secrets.token_urlsafe(32)[:32])"

# Using Go
go run -c 'package main; import ("crypto/rand"; "encoding/base64"; "fmt"); func main() { b := make([]byte, 32); rand.Read(b); fmt.Println(base64.StdEncoding.EncodeToString(b)[:32]) }'
```

**Development (provided default):**
```env
ENCRYPTION_KEY=dev-key-32-bytes-length-aes!!
```

⚠️ **WARNING**: Never use the development key in production!

## Migration Strategy

### Option 1: Fresh Installation (Recommended)
If you have no existing bank details in production:
1. Set the `ENCRYPTION_KEY` environment variable
2. Deploy the updated code
3. All new bank details will be encrypted automatically

### Option 2: Migrate Existing Data

If you have existing plaintext bank details, run this migration script:

```sql
-- Backup existing data first!
CREATE TABLE organizers_bank_backup AS 
SELECT id, bank_account_name, bank_account_number, bank_code, bank_country 
FROM organizers 
WHERE bank_account_number IS NOT NULL AND bank_account_number != '';

-- The application will handle encryption on next update
-- Users will need to re-enter their bank details OR
-- Use the migration tool (see below)
```

### Option 3: Automated Migration Tool

Create a migration utility (`cmd/migrate-bank-encryption/main.go`):

```go
package main

import (
	"fmt"
	"ticketing_system/internal/config"
	"ticketing_system/internal/database"
	"ticketing_system/internal/models"
	"ticketing_system/internal/security"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Initialize encryption service
	encryptionService, err := security.NewEncryptionService(cfg.Security.EncryptionKey)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize encryption: %v", err))
	}

	// Connect to database
	db := database.Init()

	// Get all organizers with unencrypted bank details
	var organizers []models.Organizer
	db.Where("bank_account_number != ?", "").Find(&organizers)

	fmt.Printf("Found %d organizers to migrate\n", len(organizers))

	// Migrate each organizer
	for i, organizer := range organizers {
		// Check if already encrypted (basic check - encrypted data is longer and base64)
		if len(organizer.BankAccountNumber) > 50 {
			fmt.Printf("Skipping organizer %d - appears already encrypted\n", organizer.ID)
			continue
		}

		// Encrypt the data
		encryptedNumber, encryptedCode, err := encryptionService.EncryptBankDetails(
			organizer.BankAccountNumber,
			organizer.BankCode,
		)
		if err != nil {
			fmt.Printf("ERROR: Failed to encrypt organizer %d: %v\n", organizer.ID, err)
			continue
		}

		// Update the database
		err = db.Model(&organizer).Updates(map[string]interface{}{
			"bank_account_number": encryptedNumber,
			"bank_code":           encryptedCode,
		}).Error
		if err != nil {
			fmt.Printf("ERROR: Failed to update organizer %d: %v\n", organizer.ID, err)
			continue
		}

		fmt.Printf("✅ Migrated organizer %d (%d/%d)\n", organizer.ID, i+1, len(organizers))
	}

	fmt.Println("Migration complete!")
}
```

**Run the migration:**
```bash
go run cmd/migrate-bank-encryption/main.go
```

## Verification

### Test Encryption/Decryption

```bash
# Create a test to verify encryption is working
curl -X PUT http://localhost:8080/api/organizers/bank-details \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "bank_account_name": "Test Business Inc",
    "bank_account_number": "1234567890",
    "bank_code": "TESTSWIFT123",
    "bank_country": "US"
  }'

# Retrieve and verify
curl -X GET http://localhost:8080/api/organizers/bank-details \
  -H "Authorization: Bearer $TOKEN"

# Expected response includes masked number:
# {
#   "bank_account_name": "Test Business Inc",
#   "bank_account_number": "1234567890",
#   "bank_account_number_mask": "****7890",
#   "bank_code": "TESTSWIFT123",
#   "bank_country": "US"
# }
```

### Check Database

```sql
-- Bank details should now be encrypted (base64 encoded)
SELECT 
    id, 
    bank_account_name,
    LEFT(bank_account_number, 20) as encrypted_number_preview,
    LENGTH(bank_account_number) as encrypted_length
FROM organizers 
WHERE bank_account_number IS NOT NULL;

-- Encrypted data will be ~100+ characters (base64 encoded ciphertext)
```

## Security Features

### Encryption Algorithm
- **Algorithm**: AES-GCM (Galois/Counter Mode)
- **Key Size**: 256-bit (32 bytes)
- **Authentication**: Built-in authentication tag prevents tampering
- **Nonce**: Unique random nonce for each encryption operation

### Additional Security
- **Masked Display**: Account numbers displayed with only last 4 digits visible
- **Access Control**: Only account owner can view/update their bank details
- **Backward Compatibility**: Gracefully handles unencrypted legacy data
- **Error Handling**: Encryption failures don't expose sensitive data

## Rollback Plan

If you need to rollback encryption:

1. **Restore from backup:**
```sql
-- Restore plaintext data from backup
UPDATE organizers o
SET 
    bank_account_number = b.bank_account_number,
    bank_code = b.bank_code
FROM organizers_bank_backup b
WHERE o.id = b.id;
```

2. **Revert code changes** (git checkout previous version)

3. **Remove encryption key** from environment

## Production Checklist

- [ ] Generate strong 32-byte encryption key
- [ ] Set `ENCRYPTION_KEY` in production environment
- [ ] Backup existing bank details data
- [ ] Run migration tool (if needed)
- [ ] Test encryption/decryption with test account
- [ ] Verify database shows encrypted data
- [ ] Monitor application logs for encryption errors
- [ ] Document encryption key storage location (secrets manager)
- [ ] Set up key rotation schedule (annually recommended)

## Key Management

### Key Storage
- **Development**: `.env` file (git-ignored)
- **Production**: Use a secrets manager:
  - AWS Secrets Manager
  - HashiCorp Vault
  - Azure Key Vault
  - Google Secret Manager

### Key Rotation
For maximum security, rotate encryption keys annually:

1. Generate new encryption key
2. Re-encrypt all data with new key
3. Update environment variable
4. Remove old key securely

## Compliance

This encryption implementation helps meet compliance requirements for:
- **PCI DSS**: Bank account data protection
- **GDPR**: Personal financial data security
- **SOC 2**: Data encryption at rest
- **ISO 27001**: Information security management

## Troubleshooting

### Error: "encryption key must be 16, 24, or 32 bytes"
- **Cause**: Invalid key length
- **Fix**: Generate a proper 32-byte key (see above)

### Error: "encryption service not available"
- **Cause**: Encryption service failed to initialize
- **Fix**: Check `ENCRYPTION_KEY` is set and valid length

### Error: "failed to decrypt bank details"
- **Cause**: Data encrypted with different key or corrupted
- **Fix**: Verify correct encryption key is being used

### Legacy unencrypted data still works
- **Expected**: System has backward compatibility
- **Action**: Run migration tool to encrypt legacy data

## Support

For questions or issues with encryption implementation, contact the development team.
