# Payment History Endpoints

## Overview
Two separate endpoints for viewing payment history with proper access control.

## User Payment History Endpoint

**Endpoint:** `GET /payments/history`

**Authentication:** Required (JWT Bearer token)

**Description:** Returns payment history for the authenticated user only.

**How it works:**
1. Extracts user ID from JWT token in Authorization header
2. Loads the User record to get their AccountID
3. Queries all payments for that account
4. Returns up to 50 most recent payments

**Request Example:**
```bash
curl -X GET http://localhost:8080/payments/history \
  -H "Authorization: Bearer <jwt_token>"
```

**Response:**
```json
{
  "payments": [
    {
      "id": 1,
      "account_id": 123,
      "order_id": 456,
      "transaction_id": "txn_abc123",
      "invoice_id": "R0Z8KGR",
      "amount": 5000.00,
      "currency": "KES",
      "payment_method": "MPESA",
      "status": "SUCCESS",
      "initiated_at": "2024-01-15T10:30:00Z",
      "completed_at": "2024-01-15T10:31:00Z",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 1
}
```

## Admin Payment History Endpoint

**Endpoint:** `GET /admin/payments`

**Authentication:** Required (Admin role - to be enforced)

**Description:** Returns all payment history or filtered by account.

**Query Parameters:**
- `account_id` (optional): Filter by specific account ID
- `limit` (optional): Number of records to return (default: 100)

**How it works:**
1. Accepts optional query parameters for filtering
2. If account_id provided, filters payments for that account
3. Returns payments ordered by creation date (most recent first)
4. Limits results based on limit parameter

**Request Examples:**

Get all payments (up to 100):
```bash
curl -X GET http://localhost:8080/admin/payments \
  -H "Authorization: Bearer <admin_jwt_token>"
```

Filter by account:
```bash
curl -X GET "http://localhost:8080/admin/payments?account_id=123" \
  -H "Authorization: Bearer <admin_jwt_token>"
```

With custom limit:
```bash
curl -X GET "http://localhost:8080/admin/payments?limit=50" \
  -H "Authorization: Bearer <admin_jwt_token>"
```

**Response:**
```json
{
  "payments": [
    {
      "id": 1,
      "account_id": 123,
      "order_id": 456,
      "transaction_id": "txn_abc123",
      "invoice_id": "R0Z8KGR",
      "amount": 5000.00,
      "currency": "KES",
      "payment_method": "MPESA",
      "status": "SUCCESS",
      "initiated_at": "2024-01-15T10:30:00Z",
      "completed_at": "2024-01-15T10:31:00Z",
      "created_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": 2,
      "account_id": 456,
      "order_id": 789,
      "transaction_id": "txn_def456",
      "invoice_id": "ABC123X",
      "amount": 3000.00,
      "currency": "KES",
      "payment_method": "CARD",
      "status": "SUCCESS",
      "initiated_at": "2024-01-14T15:20:00Z",
      "completed_at": "2024-01-14T15:21:00Z",
      "created_at": "2024-01-14T15:20:00Z"
    }
  ],
  "total": 2
}
```

## Implementation Details

### Authentication Flow

**User Endpoint:**
1. JWT token parsed from `Authorization: Bearer <token>` header
2. User ID extracted from `claims.Subject`
3. User's AccountID retrieved from database
4. Only payments for that account are returned

**Admin Endpoint:**
- Currently uses same JWT validation
- TODO: Add admin role validation middleware
- Should verify user has admin/staff role before allowing access

### Code Location

**Handler Functions:**
- File: `internal/payments/process.go`
- User endpoint: `GetPaymentHistory()` (line ~205)
- Admin endpoint: `GetAllPayments()` (line ~243)

**Routes:**
- File: `cmd/api-server/main.go`
- User route: line ~432
- Admin route: line ~437

### Security Notes

1. **User Endpoint:** 
   - Fully secured - users can only see their own payments
   - JWT token required and validated
   - Account access restricted to token owner

2. **Admin Endpoint:**
   - ⚠️ **TODO:** Add admin role validation middleware
   - Currently accepts any valid JWT token
   - Should verify `user.Role == "admin"` or similar

### Next Steps

To properly secure the admin endpoint:

1. Create or use existing admin middleware:
```go
func AdminOnly(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := getUserIDFromToken(r)
        if userID == 0 {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        var user models.User
        if err := db.First(&user, userID).Error; err != nil {
            http.Error(w, "User not found", http.StatusNotFound)
            return
        }
        
        if user.Role != models.RoleAdmin {
            http.Error(w, "Forbidden - Admin access required", http.StatusForbidden)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

2. Apply to admin routes:
```go
router.HandleFunc("/admin/payments", 
    AdminOnly(apiLimiter.HandlerFunc(paymentHandler.GetAllPayments))).Methods(http.MethodGet)
```

## Testing

**User Endpoint Test:**
```bash
# First login to get JWT token
TOKEN=$(curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}' \
  | jq -r '.token')

# Get payment history
curl -X GET http://localhost:8080/payments/history \
  -H "Authorization: Bearer $TOKEN"
```

**Admin Endpoint Test:**
```bash
# Login as admin
ADMIN_TOKEN=$(curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"adminpass"}' \
  | jq -r '.token')

# Get all payments
curl -X GET http://localhost:8080/admin/payments \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Filter by account
curl -X GET "http://localhost:8080/admin/payments?account_id=1&limit=20" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```
