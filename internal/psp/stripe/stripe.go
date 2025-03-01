package stripe

import (
	"fmt"
	"os"

	stripe "github.com/stripe/stripe-go/v81"
)

// StripeClient represents a Stripe payment processor
type StripeClient struct {
	secretKey string
	accountID string
}

// Init initializes the Stripe client with API key from environment
func Init() *StripeClient {
	secretKey := os.Getenv("STRIPE_SECRET_KEY")
	accountID := os.Getenv("STRIPE_ACCOUNT_ID")

	if secretKey == "" || accountID == "" {
		panic(fmt.Sprintf("missing Stripe secret key: %s or account ID: %s", secretKey, accountID))
	}

	stripe.Key = secretKey
	return &StripeClient{
		secretKey: secretKey,
		accountID: accountID,
	}
}

func (s *StripeClient) GetName() string {
	return "STRIPE"
}
