package stripe

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"payment-gateway/db/redis"
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

func (s *StripeClient) HandleWebhook(ev any, redisClient *redis.RedisClient) error {
	e := ev.(*event.Event)
	switch e.Type {
	case "payment_intent.succeeded":
		var paymentIntent stripe.PaymentIntent
		if err := json.Unmarshal(e.Data.Raw, &paymentIntent); err != nil {
			log.Printf("❌ Error parsing event data: %v", err)
			return err
		}

		// Handle successful payment (update DB, notify user, etc.)
		// Store Data in Redis
		data := map[string]interface{}{
			"status": "success",
		}

		key := fmt.Sprintf("deposit:userid:%s:orderid:%s", paymentIntent.Metadata["user_id"], paymentIntent.ID)
		err := redisClient.HSet(key, data)
		if err != nil {
			log.Println("Error storing data in redis:", err.Error())
			return fmt.Errorf("failed to store data in redis: %v", err.Error())
		}

		redisClient.IncrementGatewayScore(context.Background(), paymentIntent.Metadata["country_id"], paymentIntent.Metadata["gateway_id"])
		log.Printf("✅ Payment successful: Amount: %d", paymentIntent.Amount)

	case "payment_intent.payment_failed":
		var paymentIntent stripe.PaymentIntent
		if err := json.Unmarshal(e.Data.Raw, &paymentIntent); err != nil {
			log.Printf("❌ Error parsing event data: %v", err)
			return err
		}

		data := map[string]interface{}{
			"status": "failed",
		}

		key := fmt.Sprintf("deposit:userid:%s:orderid:%s", paymentIntent.Metadata["user_id"], paymentIntent.ID)
		err := redisClient.HSet(key, data)
		if err != nil {
			log.Println("Error storing data in redis:", err.Error())
			return fmt.Errorf("failed to store data in redis: %v", err.Error())
		}

		redisClient.DecrementGatewayScore(context.Background(), paymentIntent.Metadata["country_id"], paymentIntent.Metadata["gateway_id"])
		// Handle failed payment logic
		log.Printf("❌ Payment failed: Amount: %d", paymentIntent.Amount)

	case "payment_intent.created":
		var paymentIntent stripe.PaymentIntent
		if err := json.Unmarshal(e.Data.Raw, &paymentIntent); err != nil {
			log.Printf("❌ Error parsing event data: %v", err)
			return err
		}

		data := map[string]interface{}{
			"status": "pending",
		}

		key := fmt.Sprintf("deposit:userid:%s:orderid:%s", paymentIntent.Metadata["user_id"], paymentIntent.ID)
		err := redisClient.HSet(key, data)
		if err != nil {
			log.Println("Error storing data in redis:", err.Error())
			return fmt.Errorf("failed to store data in redis: %v", err.Error())
		}

		// Handle failed payment logic
		log.Printf("Payment Intent Pending ID: %s", paymentIntent.ID)

	default:
		log.Printf("Unhandled event type: %s", e.Type)
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
