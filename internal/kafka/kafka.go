package kafka

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaConfig holds configuration for Kafka writers and readers
type KafkaConfig struct {
	BrokerURL              string
	AllowAutoTopicCreation bool
	BatchTimeout           time.Duration
}

// WriterPool manages Kafka writers for different topics
type WriterPool struct {
	writers map[string]*kafka.Writer
	config  KafkaConfig
}

// NewWriterPool initializes a new WriterPool with the given config
func NewWriterPool() *WriterPool {
	config := KafkaConfig{}
	config.BrokerURL = GetBrokerURL()
	if config.BatchTimeout == 0 {
		config.BatchTimeout = 10 * time.Millisecond // Default batch timeout
	}
	config.AllowAutoTopicCreation = true
	return &WriterPool{
		writers: make(map[string]*kafka.Writer),
		config:  config,
	}
}

func GetBrokerURL() string {
	if os.Getenv("KAFKA_BROKER_URL") == "" {
		return "kafka:9092"
	}
	return os.Getenv("KAFKA_BROKER_URL")
}

// GetWriter returns a Kafka writer for the specified topic, creating it if it doesn't exist
func (wp *WriterPool) GetWriter(topic string) *kafka.Writer {
	if writer, exists := wp.writers[topic]; exists {
		return writer
	}

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(wp.config.BrokerURL),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: wp.config.AllowAutoTopicCreation,
		BatchTimeout:           wp.config.BatchTimeout,
	}

	wp.writers[topic] = writer
	log.Printf("Kafka writer initialized for topic '%s' at %s", topic, wp.config.BrokerURL)
	return writer
}

// PublishMessage publishes a message to the specified Kafka topic
func (wp *WriterPool) PublishMessage(ctx context.Context, topic string, key, value []byte) error {
	writer := wp.GetWriter(topic)
	msg := kafka.Message{
		Key:   key,
		Value: value,
		Time:  time.Now(),
	}

	err := writer.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to publish message to topic '%s': %v", topic, err)
	}
	return nil
}

// NewConsumer creates a new Kafka consumer with the given config
func NewConsumer(config ConsumerConfig) *kafka.Reader {
	if config.BrokerURL == "" {
		config.BrokerURL = GetBrokerURL()
	}
	if config.MinBytes == 0 {
		config.MinBytes = 10e3 // 10KB
	}
	if config.MaxBytes == 0 {
		config.MaxBytes = 10e6 // 10MB
	}
	if config.MaxWait == 0 {
		config.MaxWait = 1 * time.Second
	}
	if config.CommitInterval == 0 {
		config.CommitInterval = 1 * time.Second
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{config.BrokerURL},
		Topic:          config.Topic,
		GroupID:        config.GroupID,
		MinBytes:       config.MinBytes,
		MaxBytes:       config.MaxBytes,
		MaxWait:        config.MaxWait,
		CommitInterval: config.CommitInterval,
	})

	log.Printf("Kafka consumer initialized for topic '%s', group '%s'", config.Topic, config.GroupID)
	return reader
}
