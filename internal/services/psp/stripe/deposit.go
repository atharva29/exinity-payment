package stripe

import (
	"fmt"
	"payment-gateway/internal/models"
	"strconv"

	stripe "github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/paymentintent"
)

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

	params.Metadata = map[string]string{
		"user_id":      req.UserID.String(),
		"gateway_id":   req.GatewayID,
		"gateway_name": "STRIPE",
		"country_id":   req.CountryID,
	}

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
