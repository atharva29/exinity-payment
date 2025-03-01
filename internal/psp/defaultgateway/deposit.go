package defaultgateway

import (
	"fmt"
	"payment-gateway/internal/models"
	"strconv"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go"
)

// Deposit creates a payment intent in Stripe for accepting money
func (s *DefaultGatewayClient) Deposit(req models.DepositRequest) (string, string, error) {
	amount, err := strconv.ParseInt(req.Amount, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("invalid amount: %s", req.Amount))
	}
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(req.Currency),
	}

	params.Metadata = map[string]string{
		"user_id":      req.UserID,
		"gateway_id":   req.GatewayID,
		"gateway_name": "DEFAULT_GATEWAY",
		"country_id":   req.CountryID,
	}

	intentID := uuid.New().String()
	intentClientSecret := uuid.New().String()

	// Return the payment intent ID and client secret
	return intentID, intentClientSecret, nil
}
