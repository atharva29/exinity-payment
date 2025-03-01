package stripe

import (
	"fmt"
	"os"
	"payment-gateway/db"
	"payment-gateway/internal/kafka"

	stripe "github.com/stripe/stripe-go/v81"
)

// StripeClient represents a Stripe payment processor
type StripeClient struct {
	secretKey   string
	accountID   string
	kafkaWriter *kafka.WriterPool
	db          *db.DB
}

// Init initializes the Stripe client with API key from environment
func Init(kafkaWriter *kafka.WriterPool, db *db.DB) *StripeClient {
	secretKey := os.Getenv("STRIPE_SECRET_KEY")
	accountID := os.Getenv("STRIPE_ACCOUNT_ID")

	if secretKey == "" || accountID == "" {
		panic(fmt.Sprintf("missing Stripe secret key: %s or account ID: %s", secretKey, accountID))
	}

	stripe.Key = secretKey

	client := &StripeClient{
		secretKey:   secretKey,
		accountID:   accountID,
		kafkaWriter: kafkaWriter,
		db:          db,
	}
	client.startKafkaConsumer()
	return client
}

func (s *StripeClient) GetName() string {
	return "STRIPE"
}
