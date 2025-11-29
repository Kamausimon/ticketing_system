package payments

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	IntasendAPIBaseURL     = "https://api.intasend.com/api/v1"
	IntasendSandboxBaseURL = "https://sandbox.intasend.com/api/v1"
)

// Intasend API request/response structures

type IntasendSTKPushRequest struct {
	Amount      float64 `json:"amount"`
	PhoneNumber string  `json:"phone_number"` // Format: 254XXXXXXXXX
	Email       string  `json:"email"`
	APIRef      string  `json:"api_ref"` // Your unique reference
	Narrative   string  `json:"narrative,omitempty"`
}

type IntasendSTKPushResponse struct {
	ID              string  `json:"id"`
	InvoiceID       string  `json:"invoice_id"`
	State           string  `json:"state"` // "PENDING", "PROCESSING", "COMPLETE", "FAILED"
	Provider        string  `json:"provider"`
	Charges         float64 `json:"charges"`
	NetAmount       float64 `json:"net_amount"`
	Currency        string  `json:"currency"`
	Value           float64 `json:"value"`
	Account         string  `json:"account"` // Phone number
	APIRef          string  `json:"api_ref"`
	Host            string  `json:"host"`
	RetryCount      int     `json:"retry_count"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
	CheckoutID      string  `json:"checkout_id,omitempty"`
	CheckoutRequest string  `json:"checkout_request,omitempty"`
}

type IntasendCardPaymentRequest struct {
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Email       string  `json:"email"`
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	PhoneNumber string  `json:"phone_number,omitempty"`
	APIRef      string  `json:"api_ref"`
	RedirectURL string  `json:"redirect_url"` // Where to redirect after payment
}

type IntasendCardPaymentResponse struct {
	ID       string  `json:"id"`
	URL      string  `json:"url"` // Checkout page URL
	APIRef   string  `json:"api_ref"`
	Currency string  `json:"currency"`
	Value    float64 `json:"value"`
}

type IntasendRefundRequest struct {
	TransactionID string  `json:"transaction_id"`
	Amount        float64 `json:"amount,omitempty"` // Optional, defaults to full refund
	Reason        string  `json:"reason,omitempty"`
}

type IntasendRefundResponse struct {
	ID            string  `json:"id"`
	TransactionID string  `json:"transaction_id"`
	Amount        float64 `json:"amount"`
	State         string  `json:"state"` // "PENDING", "COMPLETE", "FAILED"
	CreatedAt     string  `json:"created_at"`
}

type IntasendTransactionStatusResponse struct {
	ID           string  `json:"id"`
	InvoiceID    string  `json:"invoice_id"`
	State        string  `json:"state"`
	Provider     string  `json:"provider"`
	Charges      float64 `json:"charges"`
	NetAmount    float64 `json:"net_amount"`
	Currency     string  `json:"currency"`
	Value        float64 `json:"value"`
	Account      string  `json:"account"`
	APIRef       string  `json:"api_ref"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
	FailedReason string  `json:"failed_reason,omitempty"`
}

// InitiateMpesaPayment initiates M-Pesa STK Push via Intasend
func (h *PaymentHandler) InitiateMpesaPayment(orderID uint, amount int64, phoneNumber, email, apiRef string) (*IntasendSTKPushResponse, error) {
	baseURL := IntasendAPIBaseURL
	if h.IntasendTestMode {
		baseURL = IntasendSandboxBaseURL
	}

	// Convert amount from cents to decimal
	amountDecimal := float64(amount) / 100.0

	reqBody := IntasendSTKPushRequest{
		Amount:      amountDecimal,
		PhoneNumber: phoneNumber,
		Email:       email,
		APIRef:      apiRef,
		Narrative:   fmt.Sprintf("Payment for Order #%d", orderID),
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", baseURL+"/payment/mpesa-stk-push/", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.IntasendSecretKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("intasend API error (status %d): %s", resp.StatusCode, string(body))
	}

	var stkResp IntasendSTKPushResponse
	if err := json.Unmarshal(body, &stkResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &stkResp, nil
}

// InitiateCardPayment initiates card payment via Intasend
func (h *PaymentHandler) InitiateCardPayment(orderID uint, amount int64, email, firstName, lastName, apiRef, redirectURL string) (*IntasendCardPaymentResponse, error) {
	baseURL := IntasendAPIBaseURL
	if h.IntasendTestMode {
		baseURL = IntasendSandboxBaseURL
	}

	amountDecimal := float64(amount) / 100.0

	reqBody := IntasendCardPaymentRequest{
		Amount:      amountDecimal,
		Currency:    "KES",
		Email:       email,
		FirstName:   firstName,
		LastName:    lastName,
		APIRef:      apiRef,
		RedirectURL: redirectURL,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", baseURL+"/checkout/", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.IntasendSecretKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("intasend API error (status %d): %s", resp.StatusCode, string(body))
	}

	var cardResp IntasendCardPaymentResponse
	if err := json.Unmarshal(body, &cardResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &cardResp, nil
}

// GetTransactionStatus retrieves transaction status from Intasend
func (h *PaymentHandler) GetIntasendTransactionStatus(transactionID string) (*IntasendTransactionStatusResponse, error) {
	baseURL := IntasendAPIBaseURL
	if h.IntasendTestMode {
		baseURL = IntasendSandboxBaseURL
	}

	req, err := http.NewRequest("GET", baseURL+"/payment/status/?invoice_id="+transactionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+h.IntasendSecretKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("intasend API error (status %d): %s", resp.StatusCode, string(body))
	}

	var statusResp IntasendTransactionStatusResponse
	if err := json.Unmarshal(body, &statusResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &statusResp, nil
}

// InitiateIntasendRefund initiates a refund via Intasend
func (h *PaymentHandler) InitiateIntasendRefund(transactionID string, amount int64, reason string) (*IntasendRefundResponse, error) {
	baseURL := IntasendAPIBaseURL
	if h.IntasendTestMode {
		baseURL = IntasendSandboxBaseURL
	}

	amountDecimal := float64(amount) / 100.0

	reqBody := IntasendRefundRequest{
		TransactionID: transactionID,
		Amount:        amountDecimal,
		Reason:        reason,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", baseURL+"/payment/refund/", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.IntasendSecretKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("intasend API error (status %d): %s", resp.StatusCode, string(body))
	}

	var refundResp IntasendRefundResponse
	if err := json.Unmarshal(body, &refundResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &refundResp, nil
}
