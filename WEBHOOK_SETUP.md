# Intasend Webhook Configuration for Your System

## Your Webhook Endpoint

**Webhook URL:** `https://ticketingapp.ngrok.dev/api/payments/webhook/intasend`

---

## Setup Steps

### 1. Configure Intasend Webhook

1. Log in to your [Intasend Dashboard](https://intasend.com/)
2. Navigate to **Developers** → **Webhooks**
3. Click **Add Webhook** or **Configure Webhook**
4. Enter the following details:

   **Webhook URL:**
   ```
   https://ticketingapp.ngrok.dev/api/payments/webhook/intasend
   ```

   **Events to Subscribe:**
   - ✅ `COMPLETE` - Payment completed successfully
   - ✅ `FAILED` - Payment failed
   - ✅ `PROCESSING` - Payment is being processed

5. Click **Save** or **Create**
6. **Important:** Copy the **Webhook Secret** that Intasend generates

### 2. Update Environment Variables

Add the webhook secret to your environment:

```bash
# In your .env file or environment configuration
INTASEND_WEBHOOK_SECRET=the_secret_you_copied_from_intasend
```

### 3. Restart Your Application

```bash
# Stop the current server (Ctrl+C) and restart
cd /home/kamau/projects/ticketing_system
go run cmd/api-server/main.go
```

---

## Test Your Webhook

### Method 1: Test from Intasend Dashboard

Most Intasend dashboards have a "Test Webhook" button that sends a sample webhook event.

1. Go to Webhooks section in Intasend Dashboard
2. Find your webhook configuration
3. Click **Test** or **Send Test Event**
4. Check your application logs to verify receipt

### Method 2: Make a Test Payment

#### Test M-Pesa Payment

```bash
curl -X POST https://ticketingapp.ngrok.dev/api/payments/initiate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "order_id": 1,
    "amount": 100,
    "currency": "KES",
    "payment_method": "mpesa",
    "phone_number": "254700000000",
    "email": "test@example.com"
  }'
```

**Note:** Use Intasend test phone numbers when in test mode:
- `254700000000` - Success
- `254700000001` - Insufficient funds  
- `254700000002` - User cancelled

#### Test Card Payment

```bash
curl -X POST https://ticketingapp.ngrok.dev/api/payments/initiate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "order_id": 1,
    "amount": 100,
    "currency": "KES",
    "payment_method": "card",
    "email": "test@example.com",
    "callback_url": "https://ticketingapp.ngrok.dev/payment/callback"
  }'
```

---

## Verify Webhook Reception

### Check Application Logs

Your application logs webhook events. Look for entries like:

```
✅ Intasend webhook received: event_id_xxxxx
✅ Payment completed for order #123
✅ Tickets generated: 2 tickets
```

### Check Database

Query the webhook logs table:

```sql
SELECT 
  id,
  event_id,
  provider,
  status,
  error_message,
  created_at
FROM webhook_logs
WHERE provider = 'intasend'
ORDER BY created_at DESC
LIMIT 10;
```

Check payment records:

```sql
SELECT 
  id,
  order_id,
  external_transaction_id,
  status,
  amount,
  gateway_fee_amount,
  net_amount,
  created_at
FROM payment_records
WHERE status IN ('completed', 'failed', 'processing')
ORDER BY created_at DESC
LIMIT 10;
```

---

## Current Environment Configuration

Your system should have these environment variables:

```bash
# Intasend Configuration
INTASEND_PUBLISHABLE_KEY=ISPubKey_test_xxxxx     # From Intasend Dashboard
INTASEND_SECRET_KEY=ISSecretKey_test_xxxxx       # From Intasend Dashboard
INTASEND_WEBHOOK_SECRET=webhook_secret_here      # Set this after webhook setup
INTASEND_TEST_MODE=true                          # true for sandbox, false for production

# Database (already configured)
DB_HOST=localhost
DB_PORT=5432
DB_USER=kamau
DB_PASSWORD=your_password
DB_NAME=ticketing_system

# Server (ngrok handles the public URL)
PORT=8080
```

---

## API Endpoints Available

With your ngrok domain, these endpoints are publicly accessible:

| Endpoint | URL | Purpose |
|----------|-----|---------|
| Initiate Payment | `POST https://ticketingapp.ngrok.dev/api/payments/initiate` | Start M-Pesa or Card payment |
| Verify Payment | `GET https://ticketingapp.ngrok.dev/api/payments/verify/{id}` | Check payment status |
| Intasend Webhook | `POST https://ticketingapp.ngrok.dev/api/payments/webhook/intasend` | Receive payment notifications |
| Create Order | `POST https://ticketingapp.ngrok.dev/api/orders` | Create new order |
| Get Orders | `GET https://ticketingapp.ngrok.dev/api/orders` | List user orders |

---

## Testing Checklist

### Initial Setup
- [ ] Intasend webhook configured with URL: `https://ticketingapp.ngrok.dev/api/payments/webhook/intasend`
- [ ] Webhook secret copied to `INTASEND_WEBHOOK_SECRET` environment variable
- [ ] Application restarted with new webhook secret
- [ ] Ngrok running and forwarding to localhost:8080

### Test M-Pesa Flow
- [ ] Create test order
- [ ] Initiate M-Pesa payment with test phone number
- [ ] Verify STK push sent (check response)
- [ ] Wait for webhook callback
- [ ] Check webhook received in logs
- [ ] Verify payment status updated to `COMPLETE`
- [ ] Verify tickets generated
- [ ] Check email sent to customer

### Test Card Flow
- [ ] Create test order
- [ ] Initiate card payment
- [ ] Receive checkout URL
- [ ] Navigate to checkout URL
- [ ] Enter test card details
- [ ] Complete payment
- [ ] Verify redirect to callback URL
- [ ] Check webhook received
- [ ] Verify tickets generated

### Test Webhook Security
- [ ] Attempt webhook with invalid signature (should fail)
- [ ] Verify duplicate webhooks are detected
- [ ] Check webhook logs for all attempts

---

## Troubleshooting

### Webhook Not Received

**Problem:** Payment successful but webhook not received

**Solutions:**
1. Verify webhook URL is correct in Intasend Dashboard
2. Check ngrok is running: `curl https://ticketingapp.ngrok.dev/api/health`
3. Check firewall/security groups allow incoming HTTPS
4. Verify webhook events are enabled (COMPLETE, FAILED, PROCESSING)
5. Check Intasend Dashboard → Webhooks → Delivery Logs

### Invalid Webhook Signature

**Problem:** Webhook received but signature verification fails

**Solutions:**
1. Verify `INTASEND_WEBHOOK_SECRET` matches the secret in Intasend Dashboard
2. Check for extra spaces or newlines in environment variable
3. Restart application after updating webhook secret
4. Verify webhook secret hasn't been regenerated in Intasend

### Payment Stuck in PENDING

**Problem:** Payment initiated but never completes

**Solutions:**
1. Check if webhook was sent by Intasend (Dashboard → Webhooks → Logs)
2. Check webhook delivery failed (retry or trigger manually)
3. Manually verify payment status:
   ```bash
   curl https://ticketingapp.ngrok.dev/api/payments/verify/{transaction_id}
   ```
4. Check Intasend Dashboard for transaction status

### Ngrok Connection Issues

**Problem:** Ngrok domain not accessible

**Solutions:**
1. Verify ngrok is running: `ngrok status`
2. Check ngrok auth token is configured
3. Verify domain is correctly set to `ticketingapp.ngrok.dev`
4. Test locally first: `curl http://localhost:8080/api/health`

---

## Production Readiness

Before going live with real payments:

### 1. Switch to Production Mode

```bash
# Update environment variables
INTASEND_PUBLISHABLE_KEY=ISPubKey_live_xxxxx     # Production key
INTASEND_SECRET_KEY=ISSecretKey_live_xxxxx       # Production key
INTASEND_WEBHOOK_SECRET=production_webhook_secret
INTASEND_TEST_MODE=false                         # Important!
```

### 2. Update Webhook URL

If moving to a permanent domain:
1. Update webhook URL in Intasend Dashboard
2. Update callback URLs in your frontend
3. Test thoroughly before announcing

### 3. Monitor Initially

- Watch webhook logs closely
- Monitor payment success rate
- Check ticket generation is working
- Verify customer emails are sent
- Monitor for any errors

---

## Quick Commands

### Check Server Status
```bash
curl https://ticketingapp.ngrok.dev/api/health
```

### Test Payment Initiation
```bash
curl -X POST https://ticketingapp.ngrok.dev/api/payments/initiate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"order_id":1,"amount":100,"currency":"KES","payment_method":"mpesa","phone_number":"254700000000","email":"test@example.com"}'
```

### Check Webhook Logs (in database)
```sql
SELECT * FROM webhook_logs ORDER BY created_at DESC LIMIT 5;
```

### Check Payment Records
```sql
SELECT * FROM payment_records ORDER BY created_at DESC LIMIT 5;
```

---

## Support Resources

- **Intasend Documentation:** https://developers.intasend.com/
- **Intasend Dashboard:** https://intasend.com/
- **Intasend Support:** support@intasend.com
- **Your Webhook URL:** https://ticketingapp.ngrok.dev/api/payments/webhook/intasend
- **System Documentation:** See `PAYMENT_SYSTEM_INTASEND.md` for full details

---

**Next Steps:**
1. ✅ Set up webhook in Intasend Dashboard with the URL above
2. ✅ Copy webhook secret to environment variables
3. ✅ Restart your application
4. ✅ Test with a small M-Pesa payment
5. ✅ Verify webhook is received and payment completes

Good luck with your payment integration testing! 🚀
