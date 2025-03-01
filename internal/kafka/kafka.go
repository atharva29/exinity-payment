package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
)

// Kafka represents the Kafka client configuration
type Kafka struct {
	producer sarama.SyncProducer
	consumer sarama.Consumer
	brokers  []string
	config   *sarama.Config
}

// Init initializes a new Kafka client
func Init(brokers []string) (*Kafka, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Consumer.Return.Errors = true

	// Initialize producer
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %v", err)
	}

	// Initialize consumer
	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %v", err)
	}

	return &Kafka{
		producer: producer,
		consumer: consumer,
		brokers:  brokers,
		config:   config,
	}, nil
}

// PublishData publishes a message to the specified topic
func (k *Kafka) PublishData(topic string, key string, value interface{}) error {
	// Check if topic exists, create if it doesn't
	err := k.ensureTopicExists(topic)
	if err != nil {
		return fmt.Errorf("failed to ensure topic exists: %v", err)
	}

	// Convert value to JSON
	msgValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %v", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(msgValue),
	}

	partition, offset, err := k.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	log.Printf("Message sent to partition %d at offset %d\n", partition, offset)
	return nil
}

// ensureTopicExists checks if a topic exists and creates it if it doesn't
func (k *Kafka) ensureTopicExists(topic string) error {
	admin, err := sarama.NewClusterAdmin(k.brokers, k.config)
	if err != nil {
		return fmt.Errorf("failed to create cluster admin: %v", err)
	}
	defer admin.Close()

	topics, err := admin.ListTopics()
	if err != nil {
		return fmt.Errorf("failed to list topics: %v", err)
	}

	if _, exists := topics[topic]; !exists {
		err = admin.CreateTopic(topic, &sarama.TopicDetail{
			NumPartitions:     1,
			ReplicationFactor: 1,
		}, false)
		if err != nil {
			return fmt.Errorf("failed to create topic: %v", err)
		}
		time.Sleep(time.Second) // Give some time for topic creation
	}

	return nil
}

// Close closes the Kafka producer and consumer
func (k *Kafka) Close() error {
	if err := k.producer.Close(); err != nil {
		return fmt.Errorf("failed to close producer: %v", err)
	}
	if err := k.consumer.Close(); err != nil {
		return fmt.Errorf("failed to close consumer: %v", err)
	}
	return nil
}
