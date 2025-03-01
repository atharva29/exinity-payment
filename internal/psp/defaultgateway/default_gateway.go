package defaultgateway

import (
	"os"
	"payment-gateway/db"
	"payment-gateway/internal/kafka"

	stripe "github.com/stripe/stripe-go/v81"
)

// DefaultGatewayClient represents a Stripe payment processor
type DefaultGatewayClient struct {
	secretKey string
	accountID string
	kafka     *kafka.Kafka
}

// Init initializes the Stripe client with API key from environment
func Init(k *kafka.Kafka, db *db.DB) *DefaultGatewayClient {
	secretKey := os.Getenv("GATEWAY_SECRET_KEY")
	accountID := os.Getenv("GATEWAY_ACCOUNT_ID")

	// if secretKey == "" || accountID == "" {
	// 	panic(fmt.Sprintf("missing Stripe secret key: %s or account ID: %s", secretKey, accountID))
	// }

	stripe.Key = secretKey

	client := &DefaultGatewayClient{
		secretKey: secretKey,
		accountID: accountID,
		kafka:     k,
	}
	go client.kafka.ConsumeDefaultGatewayWebhook(client.GetTopic(), db, client.HandleWebhook)

	return client
}

func (s *DefaultGatewayClient) GetName() string {
	return "DEFAULT_GATEWAY"
}

func (s *DefaultGatewayClient) GetTopic() string {
	return "gateway.default-gateway"
}
