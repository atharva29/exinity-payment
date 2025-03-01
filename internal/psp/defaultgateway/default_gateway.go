package defaultgateway

import (
	"os"

	stripe "github.com/stripe/stripe-go/v81"
)

// DefaultGatewayClient represents a Stripe payment processor
type DefaultGatewayClient struct {
	secretKey string
	accountID string
}

// Init initializes the Stripe client with API key from environment
func Init() *DefaultGatewayClient {
	secretKey := os.Getenv("GATEWAY_SECRET_KEY")
	accountID := os.Getenv("GATEWAY_ACCOUNT_ID")

	// if secretKey == "" || accountID == "" {
	// 	panic(fmt.Sprintf("missing Stripe secret key: %s or account ID: %s", secretKey, accountID))
	// }

	stripe.Key = secretKey
	return &DefaultGatewayClient{
		secretKey: secretKey,
		accountID: accountID,
	}
}

func (s *DefaultGatewayClient) GetName() string {
	return "DEFAULT_GATEWAY"
}
