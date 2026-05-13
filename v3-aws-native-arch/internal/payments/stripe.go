package payments

// STRIPE INTEGRATION - COMMENTED OUT FOR FUTURE INTERNATIONAL EXPANSION
// Uncomment and implement when expanding beyond Kenya
// Install: go get github.com/stripe/stripe-go/v76

/*
import (
	"fmt"
	"ticketing_system/internal/models"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/refund"
)

// InitializeStripe sets up Stripe with API key
func (h *PaymentHandler) InitializeStripe() {
	stripe.Key = h.StripeSecretKey
}

// CreateStripePaymentIntent creates a payment intent for card payments
func (h *PaymentHandler) CreateStripePaymentIntent(orderID uint, amount int64, currency, customerEmail string) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(currency),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
		Metadata: map[string]string{
			"order_id": fmt.Sprintf("%d", orderID),
		},
		ReceiptEmail: stripe.String(customerEmail),
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	return pi, nil
}

// ConfirmStripePaymentIntent confirms a payment intent
func (h *PaymentHandler) ConfirmStripePaymentIntent(paymentIntentID string) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentConfirmParams{}
	pi, err := paymentintent.Confirm(paymentIntentID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to confirm payment intent: %w", err)
	}

	return pi, nil
}

// CreateStripeRefund creates a refund for a payment
func (h *PaymentHandler) CreateStripeRefund(chargeID string, amount int64, reason string) (*stripe.Refund, error) {
	params := &stripe.RefundParams{
		Charge: stripe.String(chargeID),
		Amount: stripe.Int64(amount),
		Reason: stripe.String(reason),
	}

	r, err := refund.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	return r, nil
}

// GetStripePaymentIntent retrieves a payment intent
func (h *PaymentHandler) GetStripePaymentIntent(paymentIntentID string) (*stripe.PaymentIntent, error) {
	pi, err := paymentintent.Get(paymentIntentID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment intent: %w", err)
	}

	return pi, nil
}

// mapStripeStatusToInternal maps Stripe status to internal status
func mapStripeStatusToInternal(status stripe.PaymentIntentStatus) models.TransactionStatus {
	switch status {
	case stripe.PaymentIntentStatusSucceeded:
		return models.TransactionCompleted
	case stripe.PaymentIntentStatusProcessing:
		return models.TransactionPending
	case stripe.PaymentIntentStatusRequiresPaymentMethod,
		stripe.PaymentIntentStatusRequiresConfirmation,
		stripe.PaymentIntentStatusRequiresAction:
		return models.TransactionPending
	case stripe.PaymentIntentStatusCanceled:
		return models.TransactionCancelled
	default:
		return models.TransactionFailed
	}
}
*/
