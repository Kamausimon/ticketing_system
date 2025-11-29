# Organizer Onboarding Completeness - Quick Reference

## Summary
Fixed all issues preventing complete organizer onboarding verification flow:
- ✅ Payment gateway configuration tracking
- ✅ Bank details enforcement
- ✅ Approval/rejection email notifications
- ✅ Complete verification workflow

## Key Changes

### 1. Data Model (Organizer)
```go
PaymentGatewayID    *uint   // Link to payment gateway
IsPaymentConfigured bool    // Track if payment is set up
BankAccountName     string  // Required for payouts
BankAccountNumber   string  // Required for payouts
BankCode            string  // Required for payouts
BankCountry         string  // Required for payouts
IsVerified          bool    // Admin approval status
VerificationStatus  string  // "pending", "approved", "rejected"
RejectionReason     string  // Why application was rejected
```

### 2. New Handlers

#### Bank Details
- `UpdateBankDetails(w, r)` - Add/update bank account
- `GetBankDetails(w, r)` - Retrieve bank details

#### Payment Gateway
- `ConfigurePaymentGateway(w, r)` - Setup payment processor
- `GetPaymentGatewayConfig(w, r)` - Get current config

#### Verification (Updated)
- `VerifyOrganizer(w, r)` - Approve/reject with automatic emails

### 3. Email Templates
- **Approval:** Next steps, dashboard link, requirements
- **Rejection:** Reason, requirements, reapply link

### 4. Onboarding Flow (7 steps)
1. Profile Complete ✓
2. Email Verified ✓
3. Account Approved ✓ (NEW)
4. Tax Information ✓
5. Bank Details ✓ (NEW - Required)
6. Payment Setup ✓ (Now tracked)
7. Branding (Optional)

## API Usage

### Submit Bank Details
```bash
curl -X PUT http://localhost:8000/api/organizers/bank-details \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "bank_account_name": "My Business Inc",
    "bank_account_number": "1234567890",
    "bank_code": "SWIFTXXX",
    "bank_country": "US"
  }'
```

### Configure Payment Gateway
```bash
curl -X POST http://localhost:8000/api/organizers/payment-gateway \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "payment_gateway_id": 1,
    "config": "{\"api_key\": \"pk_xxx\"}"
  }'
```

### Check Onboarding Status
```bash
curl -X GET http://localhost:8000/api/organizers/onboarding-status \
  -H "Authorization: Bearer $TOKEN"
```

Response shows:
- `current_step` - Where they are in flow
- `progress` - Percentage complete
- `can_create_events` - true when all required steps done
- `steps` - Array of all steps with completion status

### Admin: Approve/Reject Organizer
```bash
# Approve
curl -X POST http://localhost:8000/api/organizers/123/verify \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"action": "approve"}'

# Reject
curl -X POST http://localhost:8000/api/organizers/123/verify \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"action": "reject", "reason": "Invalid tax information"}'
```

## Validation Rules

### Bank Details
- All 4 fields required (name, number, code, country)
- Stored as-is (consider encryption in production)
- Only account owner can view/edit

### Payment Gateway
- Must reference existing gateway from `payment_gateways` table
- Config stored as JSON string
- One primary gateway per organizer

### Verification
- "approve" action → sends success email, marks as verified
- "reject" action → sends rejection email with reason
- Email errors don't block the operation (logged)

## Database Changes Required

```sql
ALTER TABLE organizers ADD COLUMN (
    payment_gateway_id BIGINT UNSIGNED,
    bank_account_name VARCHAR(255),
    bank_account_number VARCHAR(255),
    bank_code VARCHAR(50),
    bank_country VARCHAR(100),
    is_payment_configured BOOLEAN DEFAULT false,
    is_verified BOOLEAN DEFAULT false,
    verification_status VARCHAR(50),
    rejection_reason TEXT
);

CREATE INDEX idx_organizers_payment_gateway 
ON organizers(payment_gateway_id);
```

## Handler Constructor Update

Must pass notification service:

```go
handler := organizers.NewOrganizerHandler(
    db,
    metrics,
    notificationService,  // NEW REQUIRED PARAMETER
)
```

## Email Notifications

### Approval Email
- Title: "Your Organizer Account Has Been Approved!"
- Includes: 4 next steps, dashboard link, account details
- Auto-sent when admin approves

### Rejection Email
- Title: "Organizer Account Application Status"
- Includes: Rejection reason, requirements, reapply link
- Auto-sent when admin rejects

## Testing Checklist

- [ ] Bank details validated (all fields required)
- [ ] Payment gateway saved to AccountPaymentGateway table
- [ ] Organizer.IsPaymentConfigured set to true
- [ ] Approval email sent with correct content
- [ ] Rejection email sent with reason
- [ ] Onboarding progress reflects new steps
- [ ] cannot_create_events = false only when all required steps done
- [ ] Bank details retrieved correctly
- [ ] Payment config linked to correct gateway

## Priority Issues Fixed

| Issue | Status | Details |
|-------|--------|---------|
| Payment gateway config not tracked | ✅ Fixed | Now stored with IsPaymentConfigured flag |
| Bank details not enforced | ✅ Fixed | Required step with validation |
| Approval email not sent | ✅ Fixed | Auto-sent on approval |
| Rejection email not sent | ✅ Fixed | Auto-sent on rejection with reason |
| Verification flow incomplete | ✅ Fixed | Complete workflow with notifications |

---

For complete documentation, see `ORGANIZER_ONBOARDING_COMPLETENESS.md`
