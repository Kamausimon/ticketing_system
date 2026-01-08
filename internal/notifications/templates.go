package notifications

// Email templates

const welcomeTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #4F46E5; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 5px 5px; }
        .button { display: inline-block; padding: 12px 30px; background: #4F46E5; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to Ticketing System!</h1>
        </div>
        <div class="content">
            <h2>Hi {{.Name}}!</h2>
            <p>Thank you for joining our platform. We're excited to have you on board!</p>
            <p>With your new account, you can:</p>
            <ul>
                <li>Browse and purchase tickets for amazing events</li>
                <li>Manage your ticket purchases</li>
                <li>Create events if you're an organizer</li>
                <li>Track your attendance history</li>
            </ul>
            <p>Get started by exploring events near you:</p>
            <a href="{{.BaseURL}}/events" class="button">Browse Events</a>
            <p>If you have any questions, feel free to reach out to our support team.</p>
        </div>
        <div class="footer">
            <p>&copy; 2024 Ticketing System. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

const verificationTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #4F46E5; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 5px 5px; }
        .button { display: inline-block; padding: 12px 30px; background: #4F46E5; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .code { font-size: 24px; font-weight: bold; letter-spacing: 5px; background: #e5e7eb; padding: 15px; text-align: center; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Email Verification</h1>
        </div>
        <div class="content">
            <h2>Hi {{.Name}}!</h2>
            <p>Please verify your email address to complete your registration.</p>
            <p>Click the button below or use the verification code:</p>
            <div class="code">{{.VerificationCode}}</div>
            <a href="{{.VerificationURL}}" class="button">Verify Email</a>
            <p>This link will expire in 24 hours.</p>
            <p>If you didn't create an account, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>&copy; 2024 Ticketing System. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

const passwordResetTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #DC2626; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 5px 5px; }
        .button { display: inline-block; padding: 12px 30px; background: #DC2626; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .warning { background: #FEF3C7; border-left: 4px solid #F59E0B; padding: 15px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Password Reset Request</h1>
        </div>
        <div class="content">
            <h2>Hi {{.Name}}!</h2>
            <p>We received a request to reset your password.</p>
            <p>Click the button below to reset your password:</p>
            <a href="{{.ResetURL}}" class="button">Reset Password</a>
            <div class="warning">
                <strong>Important:</strong> This link will expire in 1 hour for security reasons.
            </div>
            <p>If you didn't request a password reset, please ignore this email or contact support if you have concerns.</p>
        </div>
        <div class="footer">
            <p>&copy; 2024 Ticketing System. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

const orderConfirmationTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #10B981; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 5px 5px; }
        .order-details { background: white; padding: 20px; border-radius: 5px; margin: 20px 0; }
        .item { display: flex; justify-content: space-between; padding: 10px 0; border-bottom: 1px solid #e5e7eb; }
        .total { font-size: 18px; font-weight: bold; padding: 15px 0; }
        .button { display: inline-block; padding: 12px 30px; background: #10B981; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Order Confirmed! 🎉</h1>
        </div>
        <div class="content">
            <h2>Hi {{.CustomerName}}!</h2>
            <p>Thank you for your order. Your tickets have been confirmed!</p>
            
            <div class="order-details">
                <h3>Order Details</h3>
                <p><strong>Order Number:</strong> {{.OrderNumber}}</p>
                <p><strong>Event:</strong> {{.EventName}}</p>
                <p><strong>Date:</strong> {{.EventDate}}</p>
                <p><strong>Venue:</strong> {{.VenueName}}</p>
                
                <h4>Items:</h4>
                {{range .Items}}
                <div class="item">
                    <span>{{.Name}} x {{.Quantity}}</span>
                    <span>{{.Currency}} {{.Price}}</span>
                </div>
                {{end}}
                
                <div class="item total">
                    <span>Total</span>
                    <span>{{.Currency}} {{.Total}}</span>
                </div>
            </div>
            
            <a href="{{.TicketsURL}}" class="button">View Tickets</a>
            
            <p>Your tickets have been sent to you and are also available in your account.</p>
            <p>We look forward to seeing you at the event!</p>
        </div>
        <div class="footer">
            <p>&copy; 2024 Ticketing System. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

const ticketGeneratedTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #8B5CF6; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 5px 5px; }
        .ticket { background: white; padding: 20px; border-radius: 5px; margin: 20px 0; border: 2px solid #8B5CF6; }
        .button { display: inline-block; padding: 12px 30px; background: #8B5CF6; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .qr-code { text-align: center; padding: 20px; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Your Ticket is Ready! 🎫</h1>
        </div>
        <div class="content">
            <h2>Hi {{.AttendeeNa me}}!</h2>
            <p>Your ticket has been generated and is ready to use.</p>
            
            <div class="ticket">
                <h3>{{.EventName}}</h3>
                <p><strong>Date:</strong> {{.EventDate}}</p>
                <p><strong>Venue:</strong> {{.VenueName}}</p>
                <p><strong>Ticket Type:</strong> {{.TicketType}}</p>
                <p><strong>Ticket Number:</strong> {{.TicketNumber}}</p>
                
                {{if .QRCodeURL}}
                <div class="qr-code">
                    <img src="{{.QRCodeURL}}" alt="Ticket QR Code" style="max-width: 200px;" />
                </div>
                {{end}}
            </div>
            
            <a href="{{.DownloadURL}}" class="button">Download Ticket PDF</a>
            
            <p><strong>Important:</strong> Please bring this ticket (digital or printed) to the event for entry.</p>
        </div>
        <div class="footer">
            <p>&copy; 2024 Ticketing System. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

const eventReminderTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #F59E0B; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 5px 5px; }
        .event-info { background: white; padding: 20px; border-radius: 5px; margin: 20px 0; }
        .button { display: inline-block; padding: 12px 30px; background: #F59E0B; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .countdown { font-size: 24px; font-weight: bold; text-align: center; color: #F59E0B; padding: 20px; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Event Reminder ⏰</h1>
        </div>
        <div class="content">
            <h2>Hi {{.AttendeeName}}!</h2>
            <p>This is a friendly reminder about your upcoming event:</p>
            
            <div class="countdown">{{.TimeUntil}}</div>
            
            <div class="event-info">
                <h3>{{.EventName}}</h3>
                <p><strong>📅 Date:</strong> {{.EventDate}}</p>
                <p><strong>🕐 Time:</strong> {{.EventTime}}</p>
                <p><strong>📍 Venue:</strong> {{.VenueName}}</p>
                <p><strong>📌 Address:</strong> {{.VenueAddress}}</p>
            </div>
            
            <a href="{{.TicketsURL}}" class="button">View Your Tickets</a>
            
            <p><strong>Don't forget to:</strong></p>
            <ul>
                <li>Bring your ticket (digital or printed)</li>
                <li>Arrive early for smooth entry</li>
                <li>Check the event page for any updates</li>
            </ul>
            
            <p>We're excited to see you there!</p>
        </div>
        <div class="footer">
            <p>&copy; 2024 Ticketing System. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

const paymentConfirmationTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #10B981; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 5px 5px; }
        .payment-details { background: white; padding: 20px; border-radius: 5px; margin: 20px 0; }
        .success-icon { text-align: center; font-size: 48px; }
        .button { display: inline-block; padding: 12px 30px; background: #10B981; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="success-icon">✓</div>
            <h1>Payment Successful!</h1>
        </div>
        <div class="content">
            <h2>Hi {{.CustomerName}}!</h2>
            <p>Your payment has been successfully processed.</p>
            
            <div class="payment-details">
                <h3>Payment Details</h3>
                <p><strong>Transaction ID:</strong> {{.TransactionID}}</p>
                <p><strong>Amount:</strong> {{.Currency}} {{.Amount}}</p>
                <p><strong>Payment Method:</strong> {{.PaymentMethod}}</p>
                <p><strong>Date:</strong> {{.PaymentDate}}</p>
                <p><strong>Order Number:</strong> {{.OrderNumber}}</p>
            </div>
            
            <a href="{{.ReceiptURL}}" class="button">Download Receipt</a>
            
            <p>A receipt has been sent to your email for your records.</p>
            <p>If you have any questions about this payment, please contact our support team.</p>
        </div>
        <div class="footer">
            <p>&copy; 2024 Ticketing System. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

const refundProcessedTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #3B82F6; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 5px 5px; }
        .refund-details { background: white; padding: 20px; border-radius: 5px; margin: 20px 0; }
        .info-box { background: #DBEAFE; border-left: 4px solid #3B82F6; padding: 15px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Refund Processed</h1>
        </div>
        <div class="content">
            <h2>Hi {{.CustomerName}}!</h2>
            <p>Your refund request has been processed successfully.</p>
            
            <div class="refund-details">
                <h3>Refund Details</h3>
                <p><strong>Refund ID:</strong> {{.RefundID}}</p>
                <p><strong>Original Order:</strong> {{.OrderNumber}}</p>
                <p><strong>Refund Amount:</strong> {{.Currency}} {{.RefundAmount}}</p>
                <p><strong>Processing Date:</strong> {{.ProcessedDate}}</p>
                <p><strong>Refund Method:</strong> {{.RefundMethod}}</p>
            </div>
            
            <div class="info-box">
                <strong>Please note:</strong> The refund will be credited to your original payment method within {{.ProcessingDays}} business days.
            </div>
            
            <p>If you have any questions about your refund, please contact our support team.</p>
        </div>
        <div class="footer">
            <p>&copy; 2024 Ticketing System. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

const organizerApprovalTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #10B981; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 5px 5px; }
        .success-box { background: #D1FAE5; border-left: 4px solid #10B981; padding: 15px; margin: 20px 0; }
        .next-steps { background: white; padding: 20px; border-radius: 5px; margin: 20px 0; }
        .button { display: inline-block; padding: 12px 30px; background: #10B981; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .step-number { display: inline-block; width: 30px; height: 30px; background: #10B981; color: white; border-radius: 50%; text-align: center; line-height: 30px; margin-right: 10px; font-weight: bold; }
        .step { margin: 15px 0; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>✓ Account Approved!</h1>
        </div>
        <div class="content">
            <h2>Hi {{.OrganizerName}}!</h2>
            
            <div class="success-box">
                <strong>Great news!</strong> Your organizer account has been approved and verified by our admin team.
            </div>
            
            <p>Your account is now active and you can start creating events on the platform.</p>
            
            <div class="next-steps">
                <h3>Next Steps:</h3>
                <div class="step">
                    <span class="step-number">1</span>
                    <strong>Add Bank Details</strong> - Set up your bank account for payouts
                </div>
                <div class="step">
                    <span class="step-number">2</span>
                    <strong>Create Your First Event</strong> - Start selling tickets
                </div>
                <div class="step">
                    <span class="step-number">3</span>
                    <strong>Customize Your Page</strong> - Add branding and logos
                </div>
            </div>
            
            <a href="{{.DashboardURL}}" class="button">Go to Organizer Dashboard</a>
            
            <p><strong>Account Details:</strong></p>
            <ul>
                <li>Business Name: {{.OrganizerName}}</li>
                <li>Email: {{.OrganizerEmail}}</li>
                <li>Account Status: Verified & Active</li>
            </ul>
            
            <p>If you have any questions or need assistance, please contact our support team.</p>
            <p>Happy event hosting! 🎉</p>
        </div>
        <div class="footer">
            <p>&copy; 2024 Ticketing System. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

const waitlistNotificationTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #10B981; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 5px 5px; }
        .highlight { background: #FEF3C7; padding: 15px; border-radius: 5px; margin: 20px 0; border-left: 4px solid #F59E0B; }
        .button { display: inline-block; padding: 12px 30px; background: #10B981; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; font-weight: bold; }
        .details { background: white; padding: 15px; border-radius: 5px; margin: 15px 0; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
        .urgent { color: #DC2626; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎉 Tickets Are Now Available!</h1>
        </div>
        <div class="content">
            <h2>Hi {{.Name}}!</h2>
            <p>Great news! The tickets you were waiting for are now available for purchase.</p>
            
            <div class="highlight">
                <p class="urgent">⏰ Act fast! You have 24 hours to complete your purchase before this opportunity expires.</p>
            </div>
            
            <div class="details">
                <h3>Event Details:</h3>
                <p><strong>Event:</strong> {{.EventName}}</p>
                <p><strong>Date:</strong> {{.EventDate}}</p>
                {{if .VenueName}}<p><strong>Venue:</strong> {{.VenueName}}</p>{{end}}
                {{if .TicketClassName}}<p><strong>Ticket Type:</strong> {{.TicketClassName}}</p>{{end}}
                <p><strong>Quantity Available:</strong> {{.Quantity}}</p>
                {{if .Price}}<p><strong>Price:</strong> {{.Currency}} {{.Price}}</p>{{end}}
            </div>
            
            <p>Your purchase link expires at: <strong>{{.ExpiresAt}}</strong></p>
            
            <center>
                <a href="{{.PurchaseURL}}" class="button">Purchase Tickets Now</a>
            </center>
            
            <p><small>If you no longer wish to purchase these tickets, you can ignore this email. Your spot will be offered to the next person on the waitlist after 24 hours.</small></p>
        </div>
        <div class="footer">
            <p>&copy; 2024 Ticketing System. All rights reserved.</p>
            <p>You received this email because you joined the waitlist for {{.EventName}}</p>
        </div>
    </div>
</body>
</html>
`

const organizerApplicationConfirmationTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #3B82F6; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 5px 5px; }
        .info-box { background: #DBEAFE; border-left: 4px solid #3B82F6; padding: 15px; margin: 20px 0; }
        .steps { background: white; padding: 20px; border-radius: 5px; margin: 20px 0; }
        .step { margin: 15px 0; padding-left: 25px; position: relative; }
        .step::before { content: "✓"; position: absolute; left: 0; color: #10B981; font-weight: bold; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎉 Application Received!</h1>
        </div>
        <div class="content">
            <h2>Hi {{.Name}}!</h2>
            
            <p>Thank you for applying to become an event organizer on our platform. We've received your application and it's currently under review.</p>
            
            <div class="info-box">
                <strong>📧 Application Details</strong>
                <p><strong>Business Name:</strong> {{.Name}}</p>
                <p><strong>Email:</strong> {{.Email}}</p>
            </div>
            
            <div class="steps">
                <h3>What Happens Next?</h3>
                <div class="step">Your application has been submitted successfully</div>
                <div class="step">Our team will review your information (typically within 1-3 business days)</div>
                <div class="step">You'll receive an email notification once your application is approved</div>
                <div class="step">After approval, you can start creating and managing events</div>
            </div>
            
            <p><strong>While you wait, here's what we'll be reviewing:</strong></p>
            <ul>
                <li>Business information and contact details</li>
                <li>Tax identification documents</li>
                <li>Banking information for payouts</li>
                <li>Compliance with our platform policies</li>
            </ul>
            
            <p>If we need any additional information, we'll reach out to you at {{.Email}}.</p>
            
            <p><strong>Questions?</strong> Feel free to reply to this email or contact our support team.</p>
        </div>
        <div class="footer">
            <p>&copy; 2024 Ticketing System. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

const adminOrganizerNotificationTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #F59E0B; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 5px 5px; }
        .urgent { background: #FEF3C7; border-left: 4px solid #F59E0B; padding: 15px; margin: 20px 0; }
        .details { background: white; padding: 20px; border-radius: 5px; margin: 15px 0; }
        .detail-row { display: flex; padding: 8px 0; border-bottom: 1px solid #E5E7EB; }
        .detail-label { font-weight: bold; width: 150px; }
        .button { display: inline-block; padding: 12px 30px; background: #3B82F6; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; font-weight: bold; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔔 New Organizer Application</h1>
        </div>
        <div class="content">
            <h2>Hi {{.AdminName}}!</h2>
            
            <div class="urgent">
                <strong>⚠️ Action Required:</strong> A new organizer application requires your review and approval.
            </div>
            
            <p>A new event organizer has applied to join the platform. Please review their application details below.</p>
            
            <div class="details">
                <h3>Applicant Information</h3>
                <div class="detail-row">
                    <div class="detail-label">Business Name:</div>
                    <div>{{.OrganizerName}}</div>
                </div>
                <div class="detail-row">
                    <div class="detail-label">Email:</div>
                    <div>{{.OrganizerEmail}}</div>
                </div>
                <div class="detail-row">
                    <div class="detail-label">Phone:</div>
                    <div>{{.OrganizerPhone}}</div>
                </div>
                <div class="detail-row">
                    <div class="detail-label">Tax Name:</div>
                    <div>{{.TaxName}}</div>
                </div>
                <div class="detail-row">
                    <div class="detail-label">Tax PIN:</div>
                    <div>{{.TaxPin}}</div>
                </div>
                <div class="detail-row">
                    <div class="detail-label">Applied On:</div>
                    <div>{{.AppliedDate}}</div>
                </div>
            </div>
            
            <p><strong>Review Checklist:</strong></p>
            <ul>
                <li>Verify business name and contact information</li>
                <li>Validate tax identification details</li>
                <li>Check for any duplicate or fraudulent applications</li>
                <li>Review compliance with platform policies</li>
                <li>Confirm banking information is complete</li>
            </ul>
            
            <center>
                <a href="{{.ReviewURL}}" class="button">Review Application Now</a>
            </center>
            
            <p><small>Please review this application within 3 business days to maintain our service standards.</small></p>
        </div>
        <div class="footer">
            <p>&copy; 2024 Ticketing System. All rights reserved.</p>
            <p>Admin Notification - Organizer Management System</p>
        </div>
    </div>
</body>
</html>
`

const organizerRejectionTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #EF4444; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 5px 5px; }
        .reason-box { background: #FEE2E2; border-left: 4px solid #EF4444; padding: 15px; margin: 20px 0; }
        .action-box { background: #FEF3C7; border-left: 4px solid #F59E0B; padding: 15px; margin: 20px 0; }
        .button { display: inline-block; padding: 12px 30px; background: #3B82F6; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Account Application Status</h1>
        </div>
        <div class="content">
            <h2>Hi {{.OrganizerName}}!</h2>
            
            <p>Thank you for submitting your organizer account application. Unfortunately, we are unable to approve it at this time.</p>
            
            <div class="reason-box">
                <strong>Reason for Rejection:</strong>
                <p>{{.RejectionReason}}</p>
            </div>
            
            <div class="action-box">
                <strong>What happens next?</strong>
                <p>You can reapply after addressing the issues mentioned above. Please review our organizer requirements and ensure all information is accurate and complete.</p>
            </div>
            
            <p><strong>Our organizer requirements include:</strong></p>
            <ul>
                <li>Valid business information and contact details</li>
                <li>Complete tax identification</li>
                <li>Valid bank account for payouts</li>
                <li>Professional event organization history (preferred)</li>
                <li>Compliance with our terms of service</li>
            </ul>
            
            <a href="{{.ReapplyURL}}" class="button">Reapply for Approval</a>
            
            <p>If you believe this is a mistake or would like more information about why your application was rejected, please contact our support team.</p>
            <p>We're here to help! Support: {{.SupportEmail}}</p>
        </div>
        <div class="footer">
            <p>&copy; 2024 Ticketing System. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`


