package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"payment-gateway/db"
	"payment-gateway/internal/models"

	"github.com/IBM/sarama"
	event "github.com/stripe/stripe-go"
)

// ConsumeStripeWebhook consumes messages from the specified topic and processes Stripe events
func (k *Kafka) ConsumeStripeWebhook(topic string, db *db.DB, processFunc func(ev any, db *db.DB) error) error {
	partitionConsumer, err := k.consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		return fmt.Errorf("failed to create partition consumer: %v", err)
	}
	defer partitionConsumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			fmt.Println("Received message")
			var stripeEvent event.Event
			err := json.Unmarshal(msg.Value, &stripeEvent)
			if err != nil {
				log.Printf("Failed to unmarshal Stripe event: %v", err)
				continue
			}

			// Process the event using the provided function
			processFunc(&stripeEvent, db)

		case err := <-partitionConsumer.Errors():
			return fmt.Errorf("consumer error: %v", err)

		case <-ctx.Done():
			return nil
		}
	}
}

// ConsumeDefaultGatewayWebhook consumes messages from the specified topic and processes Stripe events
func (k *Kafka) ConsumeDefaultGatewayWebhook(topic string, db *db.DB, processFunc func(ev any, db *db.DB) error) error {
	partitionConsumer, err := k.consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		return fmt.Errorf("failed to create partition consumer: %v", err)
	}
	defer partitionConsumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			fmt.Println("Received message")
			var event *models.DefaultGatewayEvent
			err := json.Unmarshal(msg.Value, &event)
			if err != nil {
				log.Printf("Failed to unmarshal Default Gateway event: %v", err)
				continue
			}

			// Process the event using the provided function
			processFunc(event, db)

		case err := <-partitionConsumer.Errors():
			return fmt.Errorf("consumer error: %v", err)

		case <-ctx.Done():
			return nil
		}
	}
}
