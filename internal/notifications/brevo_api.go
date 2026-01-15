package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// BrevoAPIClient handles email sending via Brevo API (for Railway deployment)
type BrevoAPIClient struct {
	apiKey    string
	fromEmail string
	fromName  string
	timeout   time.Duration
}

// NewBrevoAPIClient creates a new Brevo API client
func NewBrevoAPIClient(apiKey, fromEmail, fromName string, timeout int) *BrevoAPIClient {
	return &BrevoAPIClient{
		apiKey:    apiKey,
		fromEmail: fromEmail,
		fromName:  fromName,
		timeout:   time.Duration(timeout) * time.Second,
	}
}

type brevoEmail struct {
	Sender      brevoContact   `json:"sender"`
	To          []brevoContact `json:"to"`
	Subject     string         `json:"subject"`
	HTMLContent string         `json:"htmlContent,omitempty"`
	TextContent string         `json:"textContent,omitempty"`
}

type brevoContact struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// SendEmail sends an email via Brevo API
func (c *BrevoAPIClient) SendEmail(to []string, subject, textBody, htmlBody string) error {
	toContacts := make([]brevoContact, len(to))
	for i, email := range to {
		toContacts[i] = brevoContact{Email: email}
	}

	payload := brevoEmail{
		Sender: brevoContact{
			Email: c.fromEmail,
			Name:  c.fromName,
		},
		To:          toContacts,
		Subject:     subject,
		TextContent: textBody,
		HTMLContent: htmlBody,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email data: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.brevo.com/v3/smtp/email", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", c.apiKey)

	client := &http.Client{Timeout: c.timeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("brevo API error (status %d): %v", resp.StatusCode, errResp)
	}

	return nil
}
