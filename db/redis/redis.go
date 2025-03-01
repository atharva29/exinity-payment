package redis

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient struct holds the Redis client and context.
type RedisClient struct {
	client *redis.Client
}

// Init initializes the Redis client.
func Init() (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	ctx := context.Background()

	// Test the connection
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{client: client}, nil
}

// HSet sets multiple hash fields to multiple values.
func (r *RedisClient) HSet(ctx context.Context, key string, values map[string]interface{}) error {
	if len(values) == 0 {
		return fmt.Errorf("no values provided for HSet")
	}
	return r.client.HSet(ctx, key, values).Err()
}

// GetExpiry set expiry to key
func (r *RedisClient) SetExpiry(key string, expiry time.Duration) error {
	err := r.client.Expire(context.Background(), key, expiry).Err()
	if err != nil {
		return fmt.Errorf("failed to set expiry to key:%s , err:%w", key, err)
	}
	return nil

}
