package apievents

import "time"

// NotificationEmailEvent carries a pre-built email so the email-worker only
// needs to call Send() — no DB queries required in the worker.
// EmailType is for logging and future metrics (e.g. "refund_completed").
type NotificationEmailEvent struct {
	To        []string  `json:"to"`
	Subject   string    `json:"subject"`
	HTMLBody  string    `json:"html_body,omitempty"`
	TextBody  string    `json:"text_body,omitempty"`
	EmailType string    `json:"email_type"`
	Timestamp time.Time `json:"timestamp"`
}
