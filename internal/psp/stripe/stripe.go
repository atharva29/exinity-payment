package stripe

import (
	"fmt"
	"os"
	"payment-gateway/db"
	"payment-gateway/internal/kafka"
	"time"

	"github.com/sony/gobreaker"
	stripe "github.com/stripe/stripe-go/v81"
)

// StripeClient represents a Stripe payment processor
type StripeClient struct {
	secretKey string
	accountID string
	kafka     *kafka.Kafka
	cb        *gobreaker.CircuitBreaker
}

// Init initializes the Stripe client with API key from environment
func Init(k *kafka.Kafka, db *db.DB) *StripeClient {
	secretKey := os.Getenv("STRIPE_SECRET_KEY")
	accountID := os.Getenv("STRIPE_ACCOUNT_ID")

	if secretKey == "" || accountID == "" {
		panic(fmt.Sprintf("missing Stripe secret key: %s or account ID: %s", secretKey, accountID))
	}

	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "StripePaymentIntent", // Name of the circuit breaker
		MaxRequests: 2,                     // Number of requests allowed in half-open state
		Interval:    60 * time.Second,      // Reset failure count every 60 seconds
		Timeout:     30 * time.Second,      // Time to switch from open to half-open
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 5 // Open circuit after 5 consecutive failures
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			fmt.Printf("Circuit Breaker '%s' changed from %s to %s\n", name, from, to)
		},
	})

	stripe.Key = secretKey
	client := &StripeClient{
		secretKey: secretKey,
		accountID: accountID,
		kafka:     k,
		cb:        cb,
	}

	go client.kafka.ConsumeStripeWebhook(client.GetTopic(), db, client.HandleWebhook)
	return client
}

func (s *StripeClient) GetTopic() string {
	return "gateway.stripe"
}

func (s *StripeClient) GetName() string {
	return "STRIPE"
}
