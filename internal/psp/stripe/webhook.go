package stripe

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"payment-gateway/db"
	database "payment-gateway/db/db"
	"payment-gateway/db/redis"
	"time"

	event "github.com/stripe/stripe-go"

	stripe "github.com/stripe/stripe-go/v81"
)

func (s *StripeClient) PublishWebhookToKafka(ev any) error {
	e := ev.(*event.Event)
	err := s.kafka.PublishData(s.GetTopic(), e.ID, e)
	if err != nil {
		log.Printf("Failed to publish data: %v", err)
	}
	return nil
}

// HandleWebhook processes incoming Stripe webhook events
func (s *StripeClient) HandleWebhook(ev any, db *db.DB) error {
	e := ev.(*event.Event)

	switch e.Type {
	// Deposit handlers
	case "payment_intent.succeeded":
		return s.handlePaymentIntentSucceeded(e, db)
	case "payment_intent.payment_failed":
		return s.handlePaymentIntentFailed(e, db)
	case "payment_intent.created":
		return s.handlePaymentIntentCreated(e, db.Redis)

	// Withdrawal handlers
	case "payout.created":
		return s.handlePayoutCreated(e, db.Redis)
	case "payout.paid":
		return s.handlePayoutPaid(e, db)
	case "payout.failed":
		return s.handlePayoutFailed(e, db.Redis)
	case "payout.canceled":
		return s.handlePayoutCanceled(e, db.Redis)

	default:
		log.Printf("Unhandled event type: %s", e.Type)
	}

	return nil
}

// ----- Deposit Event Handlers -----

// handlePaymentIntentSucceeded handles successful payments
func (s *StripeClient) handlePaymentIntentSucceeded(e *event.Event, db *db.DB) error {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(e.Data.Raw, &paymentIntent); err != nil {
		log.Printf("❌ Error parsing event data: %v", err)
		return err
	}

	metadata, err := validateMetadata(paymentIntent.Metadata)
	if err != nil {
		log.Printf("❌ Error converting gateway_id to int: %v", err)
		return fmt.Errorf("invalid gateway_id format: %v", err)
	}

	data := map[string]interface{}{
		"status": "success",
	}

	key := fmt.Sprintf("deposit:userid:%s:orderid:%s", paymentIntent.Metadata["user_id"], paymentIntent.ID)
	if err := db.Redis.HSet(context.TODO(), key, data); err != nil {
		log.Println("Error storing data in redis:", err.Error())
		return fmt.Errorf("failed to store data in redis: %v", err.Error())
	}

	db.Redis.IncrementGatewayScore(context.Background(),
		paymentIntent.Metadata["country_id"],
		paymentIntent.Metadata["gateway_id"])

	err = db.DB.CreateTransaction(database.Transaction{
		OrderID:   e.ID,
		Amount:    float64(paymentIntent.Amount) / 100, // stripe deals in paisa
		Status:    "success",
		Type:      "credit",
		GatewayID: metadata["gateway_id"],
		CountryID: metadata["country_id"],
		UserID:    metadata["user_id"],
		Currency:  string(paymentIntent.Currency),
	})
	if err != nil {
		log.Println("Error storing withdrawal transaction data in db:", err.Error())
		return fmt.Errorf("failed to store withdrawal transaction data in db: %v", err.Error())
	}

	log.Printf("✅ Payment successful: Amount: %d", paymentIntent.Amount)
	return nil
}

// handlePaymentIntentFailed handles failed payments
func (s *StripeClient) handlePaymentIntentFailed(e *event.Event, db *db.DB) error {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(e.Data.Raw, &paymentIntent); err != nil {
		log.Printf("❌ Error parsing event data: %v", err)
		return err
	}

	_, err := validateMetadata(paymentIntent.Metadata)
	if err != nil {
		log.Printf("❌ Error converting gateway_id to int: %v", err)
		return fmt.Errorf("invalid gateway_id format: %v", err)
	}

	data := map[string]interface{}{
		"status": "failed",
	}

	key := fmt.Sprintf("deposit:userid:%s:orderid:%s", paymentIntent.Metadata["user_id"], paymentIntent.ID)
	if err := db.Redis.HSet(context.TODO(), key, data); err != nil {
		log.Println("Error storing data in redis:", err.Error())
		return fmt.Errorf("failed to store data in redis: %v", err.Error())
	}

	db.Redis.DecrementGatewayScore(context.Background(),
		paymentIntent.Metadata["country_id"],
		paymentIntent.Metadata["gateway_id"])

	log.Printf("❌ Payment failed: Amount: %d", paymentIntent.Amount)
	return nil
}

// handlePaymentIntentCreated handles newly created payment intents
func (s *StripeClient) handlePaymentIntentCreated(e *event.Event, redisClient redis.IRedis) error {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(e.Data.Raw, &paymentIntent); err != nil {
		log.Printf("❌ Error parsing event data: %v", err)
		return err
	}

	data := map[string]interface{}{
		"status": "pending",
	}

	key := fmt.Sprintf("deposit:userid:%s:orderid:%s", paymentIntent.Metadata["user_id"], paymentIntent.ID)
	if err := redisClient.HSet(context.TODO(), key, data); err != nil {
		log.Println("Error storing data in redis:", err.Error())
		return fmt.Errorf("failed to store data in redis: %v", err.Error())
	}

	log.Printf("Payment Intent Pending ID: %s", paymentIntent.ID)
	return nil
}

// ----- Withdrawal Event Handlers -----

// handlePayoutCreated handles newly created payouts
func (s *StripeClient) handlePayoutCreated(e *event.Event, redisClient redis.IRedis) error {
	var payout stripe.Payout
	if err := json.Unmarshal(e.Data.Raw, &payout); err != nil {
		log.Printf("❌ Error parsing payout event data: %v", err)
		return err
	}

	data := map[string]interface{}{
		"status":       "pending",
		"created_at":   time.Now().Unix(),
		"amount":       payout.Amount,
		"currency":     string(payout.Currency),
		"arrival_date": payout.ArrivalDate,
	}

	key := fmt.Sprintf("withdrawal:userid:%s:payoutid:%s", payout.Metadata["user_id"], payout.ID)
	if err := redisClient.HSet(context.TODO(), key, data); err != nil {
		log.Println("Error storing withdrawal data in redis:", err.Error())
		return fmt.Errorf("failed to store withdrawal data in redis: %v", err.Error())
	}

	log.Printf("✅ Payout created: ID: %s, Amount: %d %s", payout.ID, payout.Amount, payout.Currency)
	return nil
}

// handlePayoutPaid handles successful payouts
func (s *StripeClient) handlePayoutPaid(e *event.Event, db *db.DB) error {
	// TODO validate metadata
	var payout stripe.Payout
	if err := json.Unmarshal(e.Data.Raw, &payout); err != nil {
		log.Printf("❌ Error parsing payout event data: %v", err)
		return err
	}

	data := map[string]interface{}{
		"status":       "completed",
		"completed_at": time.Now().Unix(),
	}

	key := fmt.Sprintf("withdrawal:userid:%s:payoutid:%s", payout.Metadata["user_id"], payout.ID)
	if err := db.Redis.HSet(context.TODO(), key, data); err != nil {
		log.Println("Error updating withdrawal data in redis:", err.Error())
		return fmt.Errorf("failed to update withdrawal data in redis: %v", err.Error())
	}

	// If using a gateway scoring system for withdrawals
	if gateway, ok := payout.Metadata["gateway_id"]; ok {
		if country, ok := payout.Metadata["country_id"]; ok {
			db.Redis.IncrementGatewayScore(context.Background(), country, gateway)
		}
	}

	// TODO add transaction insertion in database
	log.Printf("✅ Payout successful: ID: %s, Amount: %d %s", payout.ID, payout.Amount, payout.Currency)
	return nil
}

// handlePayoutFailed handles failed payouts
func (s *StripeClient) handlePayoutFailed(e *event.Event, redisClient redis.IRedis) error {
	var payout stripe.Payout
	if err := json.Unmarshal(e.Data.Raw, &payout); err != nil {
		log.Printf("❌ Error parsing payout event data: %v", err)
		return err
	}

	data := map[string]interface{}{
		"status":          "failed",
		"failed_at":       time.Now().Unix(),
		"failure_code":    string(payout.FailureCode),
		"failure_message": payout.FailureMessage,
	}

	key := fmt.Sprintf("withdrawal:userid:%s:payoutid:%s", payout.Metadata["user_id"], payout.ID)
	if err := redisClient.HSet(context.TODO(), key, data); err != nil {
		log.Println("Error updating withdrawal data in redis:", err.Error())
		return fmt.Errorf("failed to update withdrawal data in redis: %v", err.Error())
	}

	// If using a gateway scoring system for withdrawals
	if gateway, ok := payout.Metadata["gateway_id"]; ok {
		if country, ok := payout.Metadata["country_id"]; ok {
			redisClient.DecrementGatewayScore(context.Background(), country, gateway)
		}
	}

	log.Printf("❌ Payout failed: ID: %s, Amount: %d %s, Reason: %s",
		payout.ID, payout.Amount, payout.Currency, payout.FailureMessage)
	return nil
}

// handlePayoutCanceled handles canceled payouts
func (s *StripeClient) handlePayoutCanceled(e *event.Event, redisClient redis.IRedis) error {
	var payout stripe.Payout
	if err := json.Unmarshal(e.Data.Raw, &payout); err != nil {
		log.Printf("❌ Error parsing payout event data: %v", err)
		return err
	}

	data := map[string]interface{}{
		"status":      "canceled",
		"canceled_at": time.Now().Unix(),
	}

	key := fmt.Sprintf("withdrawal:userid:%s:payoutid:%s", payout.Metadata["user_id"], payout.ID)
	if err := redisClient.HSet(context.TODO(), key, data); err != nil {
		log.Println("Error updating withdrawal data in redis:", err.Error())
		return fmt.Errorf("failed to update withdrawal data in redis: %v", err.Error())
	}

	log.Printf("Payout canceled: ID: %s, Amount: %d %s", payout.ID, payout.Amount, payout.Currency)
	return nil
}

// handlePayoutUpdated handles payout status updates
func (s *StripeClient) handlePayoutUpdated(e *event.Event, redisClient redis.IRedis) error {
	var payout stripe.Payout
	if err := json.Unmarshal(e.Data.Raw, &payout); err != nil {
		log.Printf("❌ Error parsing payout event data: %v", err)
		return err
	}

	data := map[string]interface{}{
		"status":     string(payout.Status),
		"updated_at": time.Now().Unix(),
	}

	key := fmt.Sprintf("withdrawal:userid:%s:payoutid:%s", payout.Metadata["user_id"], payout.ID)
	if err := redisClient.HSet(context.TODO(), key, data); err != nil {
		log.Println("Error updating withdrawal data in redis:", err.Error())
		return fmt.Errorf("failed to update withdrawal data in redis: %v", err.Error())
	}

	log.Printf("Payout updated: ID: %s, Status: %s", payout.ID, payout.Status)
	return nil
}
