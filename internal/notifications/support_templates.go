package notifications

// Support ticket email templates

const supportTicketCreatedTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #DC2626; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 5px 5px; }
        .ticket-box { background: white; border: 2px solid #DC2626; border-radius: 5px; padding: 20px; margin: 20px 0; }
        .priority-critical { background: #FEE2E2; border-left: 4px solid #DC2626; padding: 10px; margin: 10px 0; }
        .priority-high { background: #FEF3C7; border-left: 4px solid #F59E0B; padding: 10px; margin: 10px 0; }
        .priority-medium { background: #DBEAFE; border-left: 4px solid #3B82F6; padding: 10px; margin: 10px 0; }
        .priority-low { background: #E5E7EB; border-left: 4px solid #6B7280; padding: 10px; margin: 10px 0; }
        .button { display: inline-block; padding: 12px 30px; background: #DC2626; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .info-row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #E5E7EB; }
        .label { font-weight: bold; color: #6B7280; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎫 New Support Ticket Created</h1>
        </div>
        <div class="content">
            <div class="ticket-box">
                <h2>Ticket #{{.TicketNumber}}</h2>
                
                <div class="priority-{{.Priority}}">
                    <strong>Priority: {{.Priority}}</strong>
                </div>

                <div class="info-row">
                    <span class="label">Category:</span>
                    <span>{{.Category}}</span>
                </div>
                
                <div class="info-row">
                    <span class="label">Submitted by:</span>
                    <span>{{.CustomerName}} ({{.CustomerEmail}})</span>
                </div>

                {{if .OrderID}}
                <div class="info-row">
                    <span class="label">Order ID:</span>
                    <span>#{{.OrderID}}</span>
                </div>
                {{end}}

                {{if .EventID}}
                <div class="info-row">
                    <span class="label">Event ID:</span>
                    <span>#{{.EventID}}</span>
                </div>
                {{end}}

                <div class="info-row">
                    <span class="label">Created:</span>
                    <span>{{.CreatedAt}}</span>
                </div>
            </div>

            <h3>Subject:</h3>
            <p><strong>{{.Subject}}</strong></p>

            <h3>Description:</h3>
            <p>{{.Description}}</p>

            {{if .AIClassified}}
            <div style="background: #EDE9FE; border-left: 4px solid #8B5CF6; padding: 15px; margin: 20px 0;">
                <strong>🤖 AI Analysis:</strong>
                <p><strong>Suggested Priority:</strong> {{.AIPriority}} (Confidence: {{.AIConfidence}}%)</p>
                <p><strong>Reasoning:</strong> {{.AIReasoning}}</p>
            </div>
            {{end}}

            <a href="{{.DashboardURL}}/support/tickets/{{.TicketID}}" class="button">View Ticket</a>
            
            <p style="margin-top: 30px; font-size: 14px; color: #6B7280;">
                Please respond to this ticket as soon as possible.
            </p>
        </div>
        <div class="footer">
            <p>Support Dashboard: {{.DashboardURL}}</p>
            <p>&copy; 2026 Ticketing System. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

const supportTicketStatusUpdateTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #10B981; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 5px 5px; }
        .ticket-box { background: white; border: 2px solid #10B981; border-radius: 5px; padding: 20px; margin: 20px 0; }
        .status-open { background: #DBEAFE; border-left: 4px solid #3B82F6; padding: 10px; margin: 10px 0; }
        .status-in_progress { background: #FEF3C7; border-left: 4px solid #F59E0B; padding: 10px; margin: 10px 0; }
        .status-resolved { background: #D1FAE5; border-left: 4px solid #10B981; padding: 10px; margin: 10px 0; }
        .status-closed { background: #E5E7EB; border-left: 4px solid #6B7280; padding: 10px; margin: 10px 0; }
        .button { display: inline-block; padding: 12px 30px; background: #10B981; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .info-row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #E5E7EB; }
        .label { font-weight: bold; color: #6B7280; }
        .resolution-box { background: #ECFDF5; border: 1px solid #10B981; border-radius: 5px; padding: 15px; margin: 15px 0; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📩 Support Ticket Update</h1>
        </div>
        <div class="content">
            <h2>Hi {{.CustomerName}},</h2>
            <p>Your support ticket has been updated:</p>

            <div class="ticket-box">
                <h3>Ticket #{{.TicketNumber}}</h3>
                
                <div class="status-{{.NewStatus}}">
                    <strong>Status: {{.NewStatus}}</strong>
                </div>

                <div class="info-row">
                    <span class="label">Previous Status:</span>
                    <span>{{.OldStatus}}</span>
                </div>

                {{if .Priority}}
                <div class="info-row">
                    <span class="label">Priority:</span>
                    <span>{{.Priority}}</span>
                </div>
                {{end}}

                {{if .AssignedTo}}
                <div class="info-row">
                    <span class="label">Assigned To:</span>
                    <span>{{.AssignedTo}}</span>
                </div>
                {{end}}

                <div class="info-row">
                    <span class="label">Updated:</span>
                    <span>{{.UpdatedAt}}</span>
                </div>
            </div>

            <h3>Subject:</h3>
            <p><strong>{{.Subject}}</strong></p>

            {{if .ResolutionNotes}}
            <div class="resolution-box">
                <h3>📝 Resolution Notes:</h3>
                <p>{{.ResolutionNotes}}</p>
            </div>
            {{end}}

            {{if .ResolvedAt}}
            <div style="background: #D1FAE5; border-left: 4px solid #10B981; padding: 15px; margin: 20px 0;">
                <strong>✅ Ticket Resolved!</strong>
                <p>Your issue has been resolved. If you need further assistance, please reply to this ticket or create a new one.</p>
                <p><strong>Resolved on:</strong> {{.ResolvedAt}}</p>
            </div>
            {{end}}

            <a href="{{.TicketURL}}" class="button">View Ticket Details</a>
            
            <p style="margin-top: 30px; font-size: 14px; color: #6B7280;">
                If you have any questions or need further assistance, feel free to add a comment to your ticket.
            </p>
        </div>
        <div class="footer">
            <p>Need help? Contact us at {{.SupportEmail}}</p>
            <p>&copy; 2026 Ticketing System. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

const supportTicketCommentAddedTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #3B82F6; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 5px 5px; }
        .ticket-box { background: white; border: 2px solid #3B82F6; border-radius: 5px; padding: 20px; margin: 20px 0; }
        .comment-box { background: #EFF6FF; border-left: 4px solid #3B82F6; padding: 15px; margin: 15px 0; }
        .button { display: inline-block; padding: 12px 30px; background: #3B82F6; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .info-row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #E5E7EB; }
        .label { font-weight: bold; color: #6B7280; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>💬 New Comment on Your Ticket</h1>
        </div>
        <div class="content">
            <h2>Hi {{.CustomerName}},</h2>
            <p>A new comment has been added to your support ticket:</p>

            <div class="ticket-box">
                <h3>Ticket #{{.TicketNumber}}</h3>
                
                <div class="info-row">
                    <span class="label">Subject:</span>
                    <span>{{.Subject}}</span>
                </div>

                <div class="info-row">
                    <span class="label">Status:</span>
                    <span>{{.Status}}</span>
                </div>
            </div>

            <div class="comment-box">
                <p><strong>{{.CommentAuthor}}</strong> wrote:</p>
                <p>{{.Comment}}</p>
                <p style="font-size: 12px; color: #6B7280; margin-top: 10px;">{{.CommentTime}}</p>
            </div>

            <a href="{{.TicketURL}}" class="button">View Full Conversation</a>
            
            <p style="margin-top: 30px; font-size: 14px; color: #6B7280;">
                You can reply by adding a comment to your ticket or by replying to this email.
            </p>
        </div>
        <div class="footer">
            <p>Need help? Contact us at {{.SupportEmail}}</p>
            <p>&copy; 2026 Ticketing System. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`
