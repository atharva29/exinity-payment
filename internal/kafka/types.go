package kafka

import "time"

// ConsumerConfig holds configuration for Kafka consumers
type ConsumerConfig struct {
	BrokerURL      string
	Topic          string
	GroupID        string
	MinBytes       int
	MaxBytes       int
	MaxWait        time.Duration
	CommitInterval time.Duration
}
