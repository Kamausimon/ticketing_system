# Support Ticket System - AI Integration Complete

## ✅ Implementation Summary

### Email Notifications
**Implemented email notifications for support tickets:**

1. **Ticket Created** - Notifies support team
   - Includes ticket details, priority, category
   - Shows AI classification if available
   - Links directly to ticket in dashboard
   
2. **Ticket Status Updated** - Notifies customer
   - Shows old vs new status
   - Includes resolution notes if resolved
   - Provides link to view ticket details

**Email Templates:**
- `support_ticket_created` - Beautiful HTML email for support team
- `support_ticket_status_update` - Customer-friendly status update email

### AI Priority Classification

**Automatic Priority Assignment using OpenAI GPT-3.5-turbo:**

#### How it Works:
1. When a ticket is created, AI classifier runs asynchronously (non-blocking)
2. Analyzes subject, description, category, and context (order/event)
3. Returns priority (critical/high/medium/low) with confidence score
4. If confidence ≥ 85%, priority is auto-applied to ticket
5. All AI analysis is logged in the ticket for human review

#### Priority Logic:
- **Critical**: Payment failures, security issues, event cancellations (15min SLA)
- **High**: Login problems, booking errors, upcoming events (2hr SLA)
- **Medium**: Feature questions, minor bugs, account changes (24hr SLA)
- **Low**: Suggestions, feedback, documentation requests (72hr SLA)

#### AI Fields in Database:
```go
AIClassified      bool    // Whether AI has analyzed this ticket
AIPriority        string  // AI's suggested priority
AIConfidenceScore float64 // Confidence level (0-1)
AIReasoning       string  // Explanation for the priority
```

## 🔧 Configuration

### Environment Variables

Add to your `.env` file:
```bash
# OpenAI API Key for AI Classification
OPENAI_API_KEY=sk-your-api-key-here

# Optional: Email configuration (already exists)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM_EMAIL=support@ticketingsystem.com
SMTP_FROM_NAME=Ticketing Support
```

### API Costs
- Model: GPT-3.5-turbo
- Average cost: ~$0.001-0.002 per ticket classification
- Response time: 1-3 seconds
- Runs asynchronously (doesn't block ticket creation)

## 📋 Usage Examples

### Create a Ticket (AI will auto-classify)
```bash
curl -X POST http://localhost:8080/support/tickets \
  -H "Content-Type: application/json" \
  -d '{
    "subject": "Payment failed but money was deducted",
    "description": "I tried to buy tickets for event #123 but payment failed. However, I see the charge on my bank statement.",
    "category": "payment",
    "email": "customer@example.com",
    "name": "John Doe",
    "order_id": 456,
    "event_id": 123
  }'
```

**AI Response Example:**
```json
{
  "ticket": {
    "id": 1,
    "ticket_number": "TKT-20260108-0001",
    "priority": "critical",
    "ai_classified": true,
    "ai_priority": "critical",
    "ai_confidence_score": 0.95,
    "ai_reasoning": "Payment issue with confirmed charge requires immediate investigation to prevent customer loss and potential disputes"
  }
}
```

### View Ticket with AI Analysis
```bash
curl http://localhost:8080/support/tickets/TKT-20260108-0001 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Email Flow

**1. Support Team Receives:**
```
Subject: New Support Ticket #TKT-20260108-0001 - critical
Body: 
  🎫 New Support Ticket Created
  
  Ticket #TKT-20260108-0001
  Priority: critical
  Category: payment
  Submitted by: John Doe (customer@example.com)
  
  🤖 AI Analysis:
  Suggested Priority: critical (Confidence: 95%)
  Reasoning: Payment issue with confirmed charge requires...
  
  [View Ticket Button]
```

**2. Customer Receives (on status update):**
```
Subject: Ticket #TKT-20260108-0001 Updated - resolved

Hi John Doe,

Your support ticket has been updated:

Ticket #TKT-20260108-0001
Status: resolved
Priority: critical

📝 Resolution Notes:
We've identified and reversed the duplicate charge...

[View Ticket Details Button]
```

## 🎯 AI Classification Features

### Contextual Analysis
The AI considers:
- Ticket subject and description
- Category (payment, booking, account, etc.)
- Related order ID (payment/refund context)
- Related event ID (time-sensitivity)
- Keywords indicating urgency

### Confidence Thresholds
- **≥ 85%**: Auto-apply priority
- **70-84%**: Suggest to support staff
- **< 70%**: Show as reference only

### Fallback Handling
- If OpenAI API is unavailable: Defaults to "medium" priority
- If API key not set: AI features disabled gracefully
- All errors logged without affecting ticket creation

## 📊 Monitoring AI Performance

### View Classification Stats
```bash
curl http://localhost:8080/support/tickets/stats \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN"
```

### Metrics to Track:
1. **Accuracy**: Compare AI priority vs staff-adjusted priority
2. **Override Rate**: How often staff change AI's suggestion
3. **Response Time**: Average time to first response by priority
4. **SLA Compliance**: % of tickets resolved within SLA

## 🔄 Feedback Loop (Future Enhancement)

Track when support staff override AI priority:
```go
// Log for model improvement
if manualPriority != aiPriority {
    logPriorityOverride(ticketID, aiPriority, manualPriority, reason)
}
```

Use this data to:
- Fine-tune the classification prompt
- Identify patterns AI misses
- Generate accuracy reports
- Improve model over time

## 🚀 Testing

### Test AI Classification
```bash
# Create ticket and watch logs for AI classification
curl -X POST http://localhost:8080/support/tickets \
  -H "Content-Type: application/json" \
  -d '{
    "subject": "Cannot login to my account",
    "description": "Getting error message when trying to log in",
    "category": "account",
    "email": "test@example.com",
    "name": "Test User"
  }'

# Check server logs for:
# "Auto-applied AI priority 'high' to ticket #TKT-... (confidence: 0.89)"
```

### Test Email Notifications
1. Set up SMTP credentials in `.env`
2. Create a ticket
3. Check support email inbox for creation notification
4. Update ticket status as admin/support
5. Check customer email for status update

## 📁 Files Modified/Created

**New Files:**
- `internal/ai/classifier.go` - AI classification service
- `internal/notifications/support_templates.go` - Email templates

**Modified Files:**
- `internal/support/handler.go` - Integrated AI & emails
- `internal/notifications/notifications.go` - Added email functions
- `internal/notifications/email.go` - Registered templates
- `internal/models/support_tickets.go` - AI fields (already present)
- `internal/models/user.go` - Added RoleSupport

## 🎉 Benefits

### For Support Teams:
- ✅ Automatic priority assignment
- ✅ Instant email notifications
- ✅ AI reasoning helps understand urgency
- ✅ Faster triage and routing

### For Customers:
- ✅ Automated status update emails
- ✅ Transparent communication
- ✅ Faster resolution for critical issues
- ✅ Professional email templates

### For the Business:
- ✅ Improved SLA compliance
- ✅ Reduced manual triage time
- ✅ Data-driven priority decisions
- ✅ Scalable support operations

---

**Next Steps:**
1. Add `OPENAI_API_KEY` to your `.env` file
2. Configure SMTP settings for emails
3. Start the server and test ticket creation
4. Monitor AI accuracy and adjust prompts if needed
5. Consider adding more context (user tier, ticket history, etc.)

**Cost Optimization:**
- Use caching for similar tickets
- Batch classify during off-peak hours
- Switch to GPT-4 for higher accuracy (higher cost)
- Self-host smaller models (Mistral, Llama) for zero API cost
