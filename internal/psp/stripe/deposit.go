package stripe

import (
	"fmt"
	"payment-gateway/internal/models"
	"strconv"

	"github.com/sony/gobreaker"
	stripe "github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/paymentintent"
)

// Deposit creates a payment intent in Stripe for accepting money with circuit breaking
func (s *StripeClient) Deposit(req models.DepositRequest) (string, string, error) {
	amount, err := strconv.ParseInt(req.Amount, 10, 64)
	if err != nil {
		return "", "", fmt.Errorf("invalid amount: %s", req.Amount)
	}

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(req.Currency),
	}
	stripe.Key = s.secretKey
	params.Metadata = map[string]string{
		"user_id":      req.UserID,
		"gateway_id":   req.GatewayID,
		"gateway_name": "STRIPE",
		"country_id":   req.CountryID,
	}

	// Wrap the Stripe API call in the circuit breaker
	result, err := s.cb.Execute(func() (interface{}, error) {
		intent, err := paymentintent.New(params)
		if err != nil {
			return nil, fmt.Errorf("failed to create payment intent: %w", err)
		}
		return intent, nil
	})
	if err != nil {
		// Handle circuit breaker open or execution error
		if err == gobreaker.ErrOpenState {
			// Circuit is open; implement fallback logic here
			return "", "", fmt.Errorf("stripe service unavailable, circuit breaker open")
		}
		return "", "", err // Propagate other errors (e.g., Stripe API errors)
	}

	// Cast the result back to *stripe.PaymentIntent
	intent := result.(*stripe.PaymentIntent)

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
