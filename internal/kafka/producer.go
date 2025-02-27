package kafka

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

var writer *kafka.Writer

// Initialize the Kafka writer
func init() {
	kafkaURL := os.Getenv("KAFKA_BROKER_URL")
	if kafkaURL == "" {
		kafkaURL = "kafka:9092"
	}

	writer = &kafka.Writer{
		Addr:                   kafka.TCP(kafkaURL),
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
		BatchTimeout:           10 * time.Millisecond,
	}

	log.Println("Kafka writer initialized successfully.")
}

// returns the appropriate Kafka topic based on the data format.
func GetTopic(dataFormat string) (string, error) {
	switch dataFormat {
	case "application/json":
		return "transactions.json", nil
	case "text/xml":
		return "transactions.soap", nil
	case "application/xml":
		return "transactions.soap", nil
	default:
		return "", fmt.Errorf("unsupported data format: %s", dataFormat)
	}
}

// publishes a message to the Kafka topic
func PublishTransaction(ctx context.Context, transactionID string, message []byte, dataFormat string) error {
	if writer == nil {
		log.Println("Kafka writer is nil, cannot publish to Kafka.")
		return fmt.Errorf("Kafka writer is not initialized")
	}

	topic, err := GetTopic(dataFormat)
	if err != nil {
		return err
	}

	log.Printf("Publishing message to Kafka topic: %s...", topic)

	kafkaMessage := kafka.Message{
		Key:   []byte(transactionID),
		Value: message,
		Topic: topic,
	}

	err = writer.WriteMessages(ctx, kafkaMessage)
	if err != nil {
		log.Printf("Error publishing to Kafka: %v", err)
		return err
	}

	log.Println("Message successfully published to Kafka on topic " + string(topic))
	return nil
}

// Close the writer when the system shut down
func Close() error {
	return writer.Close()
}
