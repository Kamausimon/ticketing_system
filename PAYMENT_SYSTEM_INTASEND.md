# Payment System - Intasend API Only

## Overview

This ticketing system uses **Intasend API exclusively** for all payment processing. The system supports:
- **M-Pesa** payments (STK Push)
- **Card** payments (Visa, Mastercard)
- **Offline** payments (manual verification)

Stripe and other payment gateways have been deprecated.

## Configuration

### Environment Variables

Set these environment variables in your `.env` file or deployment configuration:

```bash
# Intasend API Credentials
INTASEND_PUBLISHABLE_KEY=your_publishable_key_here
INTASEND_SECRET_KEY=your_secret_key_here
INTASEND_WEBHOOK_SECRET=your_webhook_secret_here
INTASEND_TEST_MODE=true  # Set to false for production
```

### Getting Intasend Credentials

1. Sign up at [Intasend](https://intasend.com/)
2. Navigate to API Keys section
3. Copy your Publishable Key and Secret Key
4. Set up webhook endpoint (see Webhooks section below)

## Payment Flow

### 1. Create Order
```bash
POST /api/orders
```

### 2. Initiate Payment
```bash
POST /api/payments/initiate
```

**Request Body:**
```json
{
  "order_id": 123,
  "amount": 10000,
  "currency": "KES",
  "payment_method": "mpesa",  // or "card"
  "phone_number": "254712345678",  // Required for M-Pesa
  "email": "customer@example.com"
}
```

**Response for M-Pesa:**
```json
{
  "success": true,
  "transaction_id": "intasend_transaction_id",
  "status": "PENDING",
  "checkout_request_id": "ws_co_request_id",
  "message": "M-Pesa STK Push sent. Enter your PIN on your phone."
}
```

**Response for Card:**
```json
{
  "success": true,
  "transaction_id": "intasend_transaction_id",
  "status": "PENDING",
  "checkout_url": "https://checkout.intasend.com/...",
  "message": "Redirect to checkout URL to complete payment"
}
```

### 3. Payment Verification

Payment verification happens automatically via webhooks. Once Intasend processes the payment:
- Webhook is received at `/api/payments/webhook/intasend`
- Payment status is updated in database
- Tickets are generated automatically
- Customer receives confirmation email

### 4. Check Payment Status
```bash
GET /api/payments/verify/{transaction_id}
```

## API Endpoints

### Payment Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/payments/initiate` | POST | Initiate M-Pesa or Card payment |
| `/api/payments/verify/{id}` | GET | Check payment status |
| `/api/payments/webhook/intasend` | POST | Intasend webhook (automatic) |
| `/api/payments/methods` | GET | Get saved payment methods |

### Legacy Endpoints (Deprecated)

| Endpoint | Method | Status |
|----------|--------|--------|
| `/api/orders/{id}/payment` | POST | ⚠️ Deprecated - Use `/api/payments/initiate` |
| `/api/orders/{id}/verify` | POST | ⚠️ Deprecated - Automatic via webhooks |
| `/api/accounts/stripe/setup` | POST | ❌ Disabled - Stripe not supported |

## Payment Methods

### M-Pesa STK Push

**Payment Method:** `mpesa`

**Requirements:**
- Phone number in format: `254XXXXXXXXX` (Kenya country code + number)
- Amount in cents (e.g., 1000 = KES 10.00)
- Valid email address

**Flow:**
1. Customer initiates payment
2. System sends STK Push to customer's phone
3. Customer enters M-Pesa PIN on their phone
4. Payment processed by Safaricom
5. Webhook confirms payment
6. Tickets generated automatically

**Example:**
```json
{
  "order_id": 123,
  "amount": 50000,  // KES 500.00
  "currency": "KES",
  "payment_method": "mpesa",
  "phone_number": "254712345678",
  "email": "customer@example.com"
}
```

### Card Payments

**Payment Method:** `card`

**Requirements:**
- Valid email address
- Customer first name and last name
- Callback URL (optional)

**Flow:**
1. Customer initiates payment
2. System generates Intasend checkout URL
3. Customer redirected to secure payment page
4. Customer enters card details
5. 3D Secure verification if required
6. Webhook confirms payment
7. Tickets generated automatically

**Example:**
```json
{
  "order_id": 123,
  "amount": 50000,  // KES 500.00
  "currency": "KES",
  "payment_method": "card",
  "email": "customer@example.com",
  "callback_url": "https://yourdomain.com/payment/callback"
}
```

### Offline Payments

**Payment Method:** `offline`

For cash payments, bank transfers, or other manual payment methods.

**Flow:**
1. Order created with offline payment method
2. Order status: `pending_payment`
3. Admin/organizer manually verifies payment
4. Admin marks payment as completed
5. Tickets generated

## Webhooks

### Configuration

1. Log in to Intasend Dashboard
2. Go to Developers → Webhooks
3. Add webhook URL: `https://yourdomain.com/api/payments/webhook/intasend`
4. Copy webhook secret to `INTASEND_WEBHOOK_SECRET` environment variable
5. Enable webhook events: `COMPLETE`, `FAILED`, `PROCESSING`

### Webhook Events

The system handles these Intasend webhook events:

| Event | Description | Action |
|-------|-------------|--------|
| `COMPLETE` | Payment successful | Update order, generate tickets, send email |
| `FAILED` | Payment failed | Update status, notify customer |
| `PROCESSING` | Payment processing | Update status to processing |

### Webhook Security

- All webhooks are verified using HMAC SHA256 signature
- Invalid signatures are rejected
- Duplicate webhook events are detected and ignored
- All webhook attempts are logged in `webhook_logs` table

## Testing

### Test Mode

Set `INTASEND_TEST_MODE=true` to use Intasend Sandbox:
- Test M-Pesa: Use test phone numbers from Intasend docs
- Test Cards: Use Intasend test card numbers
- No real money charged

### Test Credentials

Get test credentials from [Intasend Sandbox](https://sandbox.intasend.com/)

### Test M-Pesa Numbers

According to Intasend documentation:
- `254700000000` - Success
- `254700000001` - Insufficient funds
- `254700000002` - User cancelled

### Test Card Numbers

Use Intasend provided test cards (check their documentation).

## Refunds

Refunds are processed through Intasend API:

```bash
POST /api/orders/{id}/refund
```

**Request:**
```json
{
  "reason": "Customer request",
  "amount": 50000,  // Optional - defaults to full refund
  "notify_customer": true
}
```

**Refund Flow:**
1. Refund initiated via API
2. Request sent to Intasend
3. Intasend processes refund
4. M-Pesa: Refund to phone number (instant)
5. Card: Refund to card (3-5 business days)
6. Customer notified via email

## Transaction Fees

Intasend charges transaction fees:
- **M-Pesa:** ~3.5% + fixed fee
- **Card:** ~3.8% + fixed fee

Fees are automatically calculated and stored in the `payment_records` table:
- `amount`: Original payment amount
- `gateway_fee_amount`: Intasend fee
- `net_amount`: Amount after fees

## Error Handling

### Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `Invalid phone number` | Wrong format | Use `254XXXXXXXXX` format |
| `Insufficient funds` | Customer has low balance | Customer should add funds |
| `User cancelled` | Customer cancelled STK | Retry payment |
| `Transaction timeout` | No response from customer | Retry payment |
| `Invalid card` | Card declined | Try different card |
| `3D Secure failed` | Authentication failed | Customer should contact bank |

### Error Response Format

```json
{
  "error": "Payment failed",
  "message": "Detailed error message",
  "code": "PAYMENT_ERROR"
}
```

## Monitoring

### Payment Records

All payments are logged in the `payment_records` table:
- `external_transaction_id`: Intasend transaction ID
- `external_reference`: API reference
- `status`: Payment status
- `gateway_fee_amount`: Transaction fee
- `net_amount`: Net amount after fees

### Webhook Logs

All webhook events are logged in `webhook_logs` table:
- `provider`: Always "intasend"
- `event_id`: Intasend event ID
- `status`: Success/failure
- `error_message`: If failed
- `ip_address`: Webhook source IP

### Metrics

The system tracks these metrics:
- Payment attempts (by method)
- Payment successes (by method)
- Payment failures (by method)
- Average transaction value
- Total revenue

## Migration from Old System

If you were using the old payment system with Stripe or mock implementations:

### Steps

1. **Update Environment Variables**
   ```bash
   # Remove old Stripe variables
   # STRIPE_SECRET_KEY=...
   # STRIPE_PUBLISHABLE_KEY=...
   
   # Add Intasend variables
   INTASEND_PUBLISHABLE_KEY=your_key
   INTASEND_SECRET_KEY=your_secret
   INTASEND_WEBHOOK_SECRET=your_webhook_secret
   INTASEND_TEST_MODE=false
   ```

2. **Update Client Code**
   - Change payment endpoint from `/api/orders/{id}/payment` to `/api/payments/initiate`
   - Update request format to match Intasend requirements
   - Handle card payment redirects to checkout URL

3. **Configure Webhooks**
   - Add webhook URL in Intasend Dashboard
   - Set webhook secret in environment variables

4. **Test in Sandbox**
   - Use test mode first
   - Verify M-Pesa and Card payments work
   - Check webhook delivery

5. **Go Live**
   - Switch to production credentials
   - Set `INTASEND_TEST_MODE=false`
   - Monitor first transactions carefully

### Deprecated Features

These features are no longer supported:

❌ Stripe payment processing  
❌ Direct Stripe integration via account settings  
❌ Generic M-Pesa implementation (use Intasend M-Pesa)  
❌ Mock payment implementations  

## Support

### Intasend Support

- Documentation: https://developers.intasend.com/
- Support: support@intasend.com
- Dashboard: https://intasend.com/

### System Issues

For issues with this ticketing system:
1. Check logs in `webhook_logs` and `payment_records` tables
2. Verify environment variables are set correctly
3. Ensure webhook URL is accessible from internet
4. Check Intasend Dashboard for transaction status

## Security Best Practices

1. **Environment Variables**
   - Never commit API keys to version control
   - Use different keys for test and production
   - Rotate keys periodically

2. **Webhook Security**
   - Always verify webhook signatures
   - Use HTTPS for webhook endpoint
   - Whitelist Intasend IPs if possible

3. **PCI Compliance**
   - Never store card details
   - Always use Intasend checkout page for cards
   - Log only masked/hashed sensitive data

4. **Transaction Security**
   - Validate amounts before processing
   - Check for duplicate transactions
   - Implement rate limiting on payment endpoints

## FAQ

**Q: Can I use multiple payment providers?**  
A: Currently, the system is configured for Intasend only. Adding multiple providers would require code modifications.

**Q: How do I switch from test to production?**  
A: Update environment variables with production keys and set `INTASEND_TEST_MODE=false`.

**Q: What currencies are supported?**  
A: Currently KES (Kenyan Shilling). Intasend supports other currencies - update code to add support.

**Q: Can customers save payment methods?**  
A: Card tokenization is not yet implemented. Each payment requires entering card details.

**Q: How long do refunds take?**  
A: M-Pesa refunds are instant. Card refunds take 3-5 business days.

**Q: What if a webhook fails?**  
A: Intasend retries failed webhooks automatically. You can also manually check transaction status via API.

## Code Examples

### Frontend Integration

```javascript
// Initiate M-Pesa Payment
async function initiatePayment(orderId, phoneNumber) {
  const response = await fetch('/api/payments/initiate', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${accessToken}`
    },
    body: JSON.stringify({
      order_id: orderId,
      amount: 50000, // KES 500.00
      currency: 'KES',
      payment_method: 'mpesa',
      phone_number: phoneNumber,
      email: 'customer@example.com'
    })
  });
  
  const data = await response.json();
  
  if (data.success) {
    // Show success message, wait for webhook confirmation
    alert('STK Push sent! Check your phone and enter your M-Pesa PIN.');
    
    // Poll payment status
    pollPaymentStatus(data.transaction_id);
  } else {
    alert('Payment initiation failed: ' + data.message);
  }
}

// Poll payment status
async function pollPaymentStatus(transactionId) {
  const maxAttempts = 30; // 30 attempts
  let attempts = 0;
  
  const interval = setInterval(async () => {
    attempts++;
    
    const response = await fetch(`/api/payments/verify/${transactionId}`, {
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    });
    
    const data = await response.json();
    
    if (data.status === 'COMPLETE') {
      clearInterval(interval);
      alert('Payment successful! Check your email for tickets.');
      window.location.href = '/orders/' + data.order_id;
    } else if (data.status === 'FAILED') {
      clearInterval(interval);
      alert('Payment failed. Please try again.');
    } else if (attempts >= maxAttempts) {
      clearInterval(interval);
      alert('Payment verification timeout. Please refresh to check status.');
    }
  }, 2000); // Check every 2 seconds
}

// Initiate Card Payment
async function initiateCardPayment(orderId) {
  const response = await fetch('/api/payments/initiate', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${accessToken}`
    },
    body: JSON.stringify({
      order_id: orderId,
      amount: 50000,
      currency: 'KES',
      payment_method: 'card',
      email: 'customer@example.com',
      callback_url: window.location.origin + '/payment/callback'
    })
  });
  
  const data = await response.json();
  
  if (data.success && data.checkout_url) {
    // Redirect to Intasend checkout page
    window.location.href = data.checkout_url;
  } else {
    alert('Payment initiation failed: ' + data.message);
  }
}
```

## Changelog

### Version 2.0 (Current)
- ✅ Migrated to Intasend API exclusively
- ✅ Removed Stripe integration
- ✅ Removed generic M-Pesa mock implementation
- ✅ Implemented proper webhook handling
- ✅ Added comprehensive logging
- ✅ Added automatic ticket generation
- ✅ Added refund support

### Version 1.0 (Legacy)
- ❌ Stripe payment integration (deprecated)
- ❌ Mock M-Pesa implementation (deprecated)
- ❌ Manual payment verification (automated now)

---

**Last Updated:** December 2025  
**Payment Provider:** Intasend  
**Supported Countries:** Kenya (expandable)
