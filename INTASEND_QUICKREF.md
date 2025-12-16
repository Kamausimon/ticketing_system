# Intasend Payment Integration - Quick Reference

## 🚀 Quick Start

### Required Environment Variables
```bash
INTASEND_PUBLISHABLE_KEY=ISPubKey_xxxxx
INTASEND_SECRET_KEY=ISSecretKey_xxxxx
INTASEND_WEBHOOK_SECRET=your_webhook_secret
INTASEND_TEST_MODE=false  # true for sandbox
```

---

## 📡 API Endpoints

### Initiate Payment
```
POST /api/payments/initiate
```

### Verify Payment
```
GET /api/payments/verify/{transaction_id}
```

### Webhook (Intasend calls this)
```
POST /api/payments/webhook/intasend
```

---

## 💳 Payment Methods

### M-Pesa STK Push
```json
{
  "order_id": 123,
  "amount": 50000,
  "currency": "KES",
  "payment_method": "mpesa",
  "phone_number": "254712345678",
  "email": "customer@example.com"
}
```

### Card Payment
```json
{
  "order_id": 123,
  "amount": 50000,
  "currency": "KES",
  "payment_method": "card",
  "email": "customer@example.com",
  "callback_url": "https://yourdomain.com/callback"
}
```

---

## 🔄 Payment Status

| Status | Description |
|--------|-------------|
| `PENDING` | Payment initiated, awaiting completion |
| `PROCESSING` | Payment being processed |
| `COMPLETE` | Payment successful |
| `FAILED` | Payment failed |

---

## 🔐 Webhook Setup

1. Go to Intasend Dashboard → Webhooks
2. Add URL: `https://yourdomain.com/api/payments/webhook/intasend`
3. Copy webhook secret to `INTASEND_WEBHOOK_SECRET`
4. Enable events: `COMPLETE`, `FAILED`, `PROCESSING`

---

## ✅ Testing

### Test Mode
```bash
INTASEND_TEST_MODE=true
```

### Test Phone Numbers (M-Pesa)
- Success: `254700000000`
- Insufficient funds: `254700000001`
- User cancelled: `254700000002`

### Test Cards
Check Intasend documentation for test card numbers.

---

## ⚠️ Deprecated

These are **NO LONGER SUPPORTED**:

- ❌ Stripe integration
- ❌ Direct M-Pesa (non-Intasend)
- ❌ `/api/orders/{id}/payment` for online payments
- ❌ `/api/accounts/stripe/setup`

Use `/api/payments/initiate` instead!

---

## 📊 Monitor

### Check Webhook Logs
```sql
SELECT * FROM webhook_logs 
WHERE provider = 'intasend' 
ORDER BY created_at DESC LIMIT 10;
```

### Check Payment Records
```sql
SELECT * FROM payment_records 
WHERE order_id = 123;
```

---

## 🆘 Troubleshooting

| Issue | Solution |
|-------|----------|
| Webhook not received | Check URL is publicly accessible via HTTPS |
| STK Push not on phone | Verify phone format: `254XXXXXXXXX` |
| Payment stuck | Check Intasend Dashboard for status |
| Card declined | Customer should try different card |

---

## 📚 Documentation

- Full Guide: `PAYMENT_SYSTEM_INTASEND.md`
- Migration Info: `PAYMENT_MIGRATION_SUMMARY.md`
- Intasend Docs: https://developers.intasend.com/

---

## 💰 Transaction Fees

- M-Pesa: ~3.5% + fixed fee
- Card: ~3.8% + fixed fee

Fees automatically calculated and stored in `gateway_fee_amount`.

---

**Quick Contact:**
- Intasend Support: support@intasend.com
- Dashboard: https://intasend.com/
