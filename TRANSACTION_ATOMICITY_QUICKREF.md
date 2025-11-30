# Transaction Atomicity - Quick Reference

## Problem Solved
✅ Payment and ticket generation now happen atomically - either both succeed or both fail

## Key Files

### 1. `/internal/orders/transactions.go`
**Core Functions:**
- `ProcessPaymentWithTickets()` - Atomic payment + ticket generation
- `VerifyPaymentAndGenerateTickets()` - Wrapper for external calls
- `RollbackPayment()` - Manual rollback for exceptional cases

### 2. `/internal/orders/payment.go` (Updated)
- `ProcessPayment()` - Now uses atomic transaction
- `VerifyPayment()` - Now uses atomic transaction

### 3. `/internal/payments/webhooks.go` (Updated)
- `handleIntasendComplete()` - Webhook now generates tickets atomically

## How It Works

```go
// Start transaction
tx := h.db.Begin()

// Step 1: Update payment status
order.PaymentStatus = models.PaymentCompleted
tx.Save(&order)

// Step 2: Generate tickets
for each order item {
    for i := 0; i < quantity; i++ {
        ticket := models.Ticket{...}
        tx.Create(&ticket)
    }
}

// Step 3: Mark as fulfilled
order.Status = models.OrderFulfilled
tx.Save(&order)

// Commit - both succeed or both fail
tx.Commit()
```

## API Changes

### Before (2 separate calls):
```bash
# Step 1: Process payment
POST /orders/123/payment
# Step 2: Generate tickets
POST /tickets/generate
```

### After (1 atomic call):
```bash
# Single call does both
POST /orders/123/payment
Response: {
  "message": "Payment processed and tickets generated successfully",
  "tickets_created": 2
}
```

## Webhook Integration

Payment webhooks automatically generate tickets:

```bash
POST /webhooks/intasend
# Webhook now:
# 1. Verifies payment
# 2. Generates tickets
# 3. Both in ONE transaction
```

## Testing

Run tests:
```bash
cd internal/orders
go test -v -run TestProcessPaymentWithTickets
```

Test coverage:
- ✅ Successful atomic transaction
- ✅ Order state validation
- ✅ Transaction rollback on failure
- ✅ Idempotency (no duplicates)
- ✅ Manual rollback function
- ✅ Concurrent payment protection

## Benefits

1. **No Inconsistent States** - Payment + tickets always in sync
2. **Automatic Rollback** - Any error reverts all changes
3. **Idempotent** - Safe to retry failed operations
4. **Concurrent-Safe** - Multiple webhooks handled correctly

## Rollback Scenarios

Manual rollback for exceptional cases:
```go
err := orderHandler.RollbackPayment(orderID, "Fraud detected")
```

Automatic rollback on:
- Database errors
- Invalid order state
- Ticket creation failures
- Transaction commit errors
- Panic during execution

## Monitoring

Check for inconsistent states:
```sql
-- Should always return 0
SELECT COUNT(*) FROM orders
WHERE payment_status = 'completed'
  AND status != 'fulfilled';
```

## Key Guarantees

✅ If payment succeeds, tickets are ALWAYS created
✅ If ticket creation fails, payment is NEVER marked complete
✅ No partial states - atomicity guaranteed
✅ Safe concurrent webhook processing
✅ Idempotent - duplicate calls handled safely

## Error Messages

```json
{
  "error": "payment transaction failed: failed to create ticket"
}
```

If you see this, the entire transaction was rolled back - safe to retry.

## Documentation

Full docs: `/TRANSACTION_ATOMICITY_COMPLETE.md`
