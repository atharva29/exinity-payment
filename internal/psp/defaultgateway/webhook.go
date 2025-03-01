package defaultgateway

import (
	"context"
	"fmt"
	"log"
	"payment-gateway/db"
	database "payment-gateway/db/db"
	"payment-gateway/db/redis"
	"payment-gateway/internal/models"
	"time"
)

func (s *DefaultGatewayClient) PublishWebhookToKafka(ev any) error {
	e := ev.(*models.DefaultGatewayEvent)
	err := s.kafka.PublishData(s.GetTopic(), e.ID, e)
	if err != nil {
		log.Printf("Failed to publish data: %v", err)
	}
	return nil
}

// HandleWebhook processes incoming Stripe webhook events
func (s *DefaultGatewayClient) HandleWebhook(ev any, db *db.DB) error {
	e := ev.(*models.DefaultGatewayEvent)

	switch e.Type {
	// Deposit handlers
	case "payment_intent.succeeded":
		return s.handlePaymentIntentSucceeded(e, db)
	case "payment_intent.payment_failed":
		return s.handlePaymentIntentFailed(e, db)
	case "payment_intent.created":
		return s.handlePaymentIntentCreated(e, db)

	// Withdrawal handlers
	case "payout.created":
		return s.handlePayoutCreated(e, db.Redis)
	case "payout.paid":
		return s.handlePayoutCompleted(e, db)
	case "payout.failed":
		return s.handlePayoutFailed(e, db)
	case "payout.canceled":
		return s.handlePayoutCancelled(e, db.Redis)

	default:
		log.Printf("Unhandled event type: %s", e.Type)
	}

	return nil
}

// handlePaymentIntentSucceeded handles successful payments
func (s *DefaultGatewayClient) handlePaymentIntentSucceeded(e *models.DefaultGatewayEvent, db *db.DB) error {
	metadata, err := validateMetadata(e.Data.Metadata)
	if err != nil {
		log.Printf("❌ Error converting gateway_id to int: %v", err)
		return fmt.Errorf("invalid gateway_id format: %v", err)
	}

	data := map[string]interface{}{
		"status": "success",
	}

	key := fmt.Sprintf("deposit:userid:%s:orderid:%s", e.Data.Metadata["user_id"], e.ID)
	if err := db.Redis.HSet(context.TODO(), key, data); err != nil {
		log.Println("Error storing data in redis:", err.Error())
		return fmt.Errorf("failed to store data in redis: %v", err.Error())
	}

	db.Redis.IncrementGatewayScore(context.Background(),
		e.Data.Metadata["country_id"],
		e.Data.Metadata["gateway_id"])

	err = db.DB.CreateTransaction(database.Transaction{
		OrderID:   e.ID,
		Amount:    float64(e.Amount) / 100,
		Status:    "success",
		Type:      "credit",
		GatewayID: metadata["gateway_id"],
		CountryID: metadata["country_id"],
		UserID:    metadata["user_id"],
		Currency:  e.Currency,
	})
	if err != nil {
		log.Println("Error storing deposit transaction data in db:", err.Error())
		return fmt.Errorf("failed to store deposit transaction data in db: %v", err.Error())
	}
	log.Printf("✅ Payment successful: Amount: %d", e.Amount)
	return nil
}

// handlePaymentIntentSucceeded handles successful payments
func (s *DefaultGatewayClient) handlePaymentIntentFailed(e *models.DefaultGatewayEvent, db *db.DB) error {

	data := map[string]interface{}{
		"status": "failed",
	}

	key := fmt.Sprintf("deposit:userid:%s:orderid:%s", e.Data.Metadata["user_id"], e.ID)
	if err := db.Redis.HSet(context.TODO(), key, data); err != nil {
		log.Println("Error storing data in redis:", err.Error())
		return fmt.Errorf("failed to store data in redis: %v", err.Error())
	}

	db.Redis.DecrementGatewayScore(context.Background(),
		e.Data.Metadata["country_id"],
		e.Data.Metadata["gateway_id"])

	log.Printf("❌ Payment failed: Amount: %d", e.Amount)
	return nil
}

// handlePaymentIntentSucceeded handles successful payments
func (s *DefaultGatewayClient) handlePaymentIntentCreated(e *models.DefaultGatewayEvent, db *db.DB) error {
	data := map[string]interface{}{
		"status": "pending",
	}

	key := fmt.Sprintf("deposit:userid:%s:orderid:%s", e.Data.Metadata["user_id"], e.ID)
	if err := db.Redis.HSet(context.TODO(), key, data); err != nil {
		log.Println("Error storing data in redis:", err.Error())
		return fmt.Errorf("failed to store data in redis: %v", err.Error())
	}

	db.Redis.DecrementGatewayScore(context.Background(),
		e.Data.Metadata["country_id"],
		e.Data.Metadata["gateway_id"])

	log.Printf("Payment Intent Pending ID: %s", e.ID)
	return nil
}

// handlePayoutCreated handles newly created payouts
func (s *DefaultGatewayClient) handlePayoutCreated(e *models.DefaultGatewayEvent, redisClient redis.IRedis) error {

	data := map[string]interface{}{
		"status":     "created",
		"created_at": time.Now().Unix(),
		"amount":     e.Amount,
		"currency":   string(e.Currency),
	}

	key := fmt.Sprintf("withdrawal:userid:%s:payoutid:%s", e.Data.Metadata["user_id"], e.ID)
	if err := redisClient.HSet(context.TODO(), key, data); err != nil {
		log.Println("Error storing withdrawal data in redis:", err.Error())
		return fmt.Errorf("failed to store withdrawal data in redis: %v", err.Error())
	}

	log.Printf("✅ Payout created: ID: %s, Amount: %d %s", e.ID, e.Amount, e.Currency)
	return nil
}

// handlePayoutCompleted handles newly created payouts
func (s *DefaultGatewayClient) handlePayoutCompleted(e *models.DefaultGatewayEvent, db *db.DB) error {
	// Convert string metadata to integers
	metadata, err := validateMetadata(e.Data.Metadata)
	if err != nil {
		log.Printf("❌ Error converting gateway_id to int: %v", err)
		return fmt.Errorf("invalid gateway_id format: %v", err)
	}

	data := map[string]interface{}{
		"status":       "completed",
		"completed_at": time.Now().Unix(),
	}

	key := fmt.Sprintf("withdrawal:userid:%s:payoutid:%s", e.Data.Metadata["user_id"], e.ID)
	if err := db.Redis.HSet(context.TODO(), key, data); err != nil {
		log.Println("Error storing withdrawal data in redis:", err.Error())
		return fmt.Errorf("failed to store withdrawal data in redis: %v", err.Error())
	}

	err = db.DB.CreateTransaction(database.Transaction{
		OrderID:   e.ID,
		Amount:    -float64(e.Amount) / 100,
		Status:    "success",
		Type:      "debit",
		GatewayID: metadata["gateway_id"],
		CountryID: metadata["country_id"],
		UserID:    metadata["user_id"],
		Currency:  e.Currency,
	})
	if err != nil {
		log.Println("Error storing withdrawal transaction data in db:", err.Error())
		return fmt.Errorf("failed to store withdrawal transaction data in db: %v", err.Error())
	}

	log.Printf("✅ Payout succcess: ID: %s, Amount: %d %s", e.ID, e.Amount, e.Currency)
	return nil
}

// handlePayoutFailed handles newly created payouts
func (s *DefaultGatewayClient) handlePayoutFailed(e *models.DefaultGatewayEvent, db *db.DB) error {

	_, err := validateMetadata(e.Data.Metadata)
	if err != nil {
		log.Printf("❌ Error converting gateway_id to int: %v", err)
		return fmt.Errorf("invalid gateway_id format: %v", err)
	}

	data := map[string]interface{}{
		"status":    "failed",
		"failed_at": time.Now().Unix(),
	}

	key := fmt.Sprintf("withdrawal:userid:%s:payoutid:%s", e.Data.Metadata["user_id"], e.ID)
	if err := db.Redis.HSet(context.TODO(), key, data); err != nil {
		log.Println("Error storing withdrawal data in redis:", err.Error())
		return fmt.Errorf("failed to store withdrawal data in redis: %v", err.Error())
	}

	log.Printf("✅ Payout Failed: ID: %s, Amount: %d %s", e.ID, e.Amount, e.Currency)
	return nil
}

// handlePayoutCancelled handles newly created payouts
func (s *DefaultGatewayClient) handlePayoutCancelled(e *models.DefaultGatewayEvent, redisClient redis.IRedis) error {

	data := map[string]interface{}{
		"status":       "cancelled",
		"cancelled_at": time.Now().Unix(),
	}

	key := fmt.Sprintf("withdrawal:userid:%s:payoutid:%s", e.Data.Metadata["user_id"], e.ID)
	if err := redisClient.HSet(context.TODO(), key, data); err != nil {
		log.Println("Error storing withdrawal data in redis:", err.Error())
		return fmt.Errorf("failed to store withdrawal data in redis: %v", err.Error())
	}

	log.Printf("✅ Payout Cancelled: ID: %s, Amount: %d %s", e.ID, e.Amount, e.Currency)
	return nil
}
