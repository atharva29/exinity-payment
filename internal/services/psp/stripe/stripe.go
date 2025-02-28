package stripe

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"payment-gateway/internal/models"
	"strconv"

	event "github.com/stripe/stripe-go"
	stripe "github.com/stripe/stripe-go/v81"

	"github.com/stripe/stripe-go/v81/paymentintent"
	"github.com/stripe/stripe-go/v81/payout"
)

// StripeClient represents a Stripe payment processor
type StripeClient struct {
	secretKey string
}

// Init initializes the Stripe client with API key from environment
func Init() *StripeClient {
	secretKey := os.Getenv("STRIPE_SECRET_KEY")

	if secretKey == "" {
		panic("STRIPE_SECRET_KEY environment variable is not set")
	}

	stripe.Key = secretKey
	return &StripeClient{
		secretKey: secretKey,
	}
}

func (s *StripeClient) GetName() string {
	return "STRIPE"
}

// Deposit creates a payment intent in Stripe for accepting money
func (s *StripeClient) Deposit(req models.DepositRequest) (string, string, error) {
	amount, err := strconv.ParseInt(req.Amount, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("invalid amount: %s", req.Amount))
	}
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(req.Currency),
	}

	// if req.Description != "" {
	// 	params.Description = stripe.String(req.Description)
	// }

	// if req.CustomerID != "" {
	// 	params.Customer = stripe.String(req.CustomerID)
	// }

	// if req.PaymentMethodID != "" {
	// 	params.PaymentMethod = stripe.String(req.PaymentMethodID)
	// 	params.ConfirmationMethod = stripe.String(string(stripe.PaymentIntentConfirmationMethodManual))
	// 	params.Confirm = stripe.Bool(true)
	// }

	// if req.ReceiptEmail != "" {
	// 	params.ReceiptEmail = stripe.String(req.ReceiptEmail)
	// }

	// if req.Metadata != nil {
	// 	params.Metadata = req.Metadata
	// }

	intent, err := paymentintent.New(params)
	if err != nil {
		return "", "", fmt.Errorf("failed to create payment intent: %w", err)
	}

	// Return the payment intent ID and client secret
	return intent.ID, intent.ClientSecret, nil
}

// GetDepositStatus retrieves the status of a deposit (payment intent)
func (s *StripeClient) GetDepositStatus(depositID string) (string, error) {
	intent, err := paymentintent.Get(depositID, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get payment intent: %w", err)
	}

	return string(intent.Status), nil
}

// Withdraw sends a payout via Stripe to a connected bank account
func (s *StripeClient) Withdrawal(req models.WithdrawalRequest) (string, error) {
	params := &stripe.PayoutParams{
		Amount:      stripe.Int64(req.Amount),
		Currency:    stripe.String(req.Currency),
		Destination: stripe.String(req.Destination),
	}

	if req.Description != "" {
		params.Description = stripe.String(req.Description)
	}

	if req.Method != "" {
		params.Method = stripe.String(req.Method)
	}

	if req.StatementDescriptor != "" {
		params.StatementDescriptor = stripe.String(req.StatementDescriptor)
	}

	if req.Metadata != nil {
		params.Metadata = req.Metadata
	}

	p, err := payout.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create payout: %w", err)
	}

	return p.ID, nil
}

func (s *StripeClient) HandleWebhook(event event.Event) error {
	switch event.Type {
	case "payment_intent.succeeded":
		var paymentIntent stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
			log.Printf("❌ Error parsing event data: %v", err)
			return err
		}

		// Handle successful payment (update DB, notify user, etc.)
		log.Printf("✅ Payment successful: Amount: %d", paymentIntent.Amount)

	case "payment_intent.payment_failed":
		var paymentIntent stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
			log.Printf("❌ Error parsing event data: %v", err)
			return err
		}

		// Handle failed payment logic
		log.Printf("❌ Payment failed: Amount: %d", paymentIntent.Amount)

	default:
		log.Printf("Unhandled event type: %s", event.Type)
	}

	return nil
}

// GetWithdrawalStatus retrieves the status of a withdrawal (payout)
func (s *StripeClient) GetWithdrawalStatus(withdrawalID string) (string, error) {
	p, err := payout.Get(withdrawalID, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get payout status: %w", err)
	}

	return string(p.Status), nil
}

// CancelWithdrawal cancels a pending withdrawal (payout)
func (s *StripeClient) CancelWithdrawal(withdrawalID string) error {
	_, err := payout.Cancel(withdrawalID, nil)
	if err != nil {
		return fmt.Errorf("failed to cancel payout: %w", err)
	}

	return nil
}
