package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient struct holds the Redis client and context.
type RedisClient struct {
	Client *redis.Client
	Ctx    context.Context
}

// Init initializes the Redis client.
func Init(addr, password string, db int) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()

	// Test the connection
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{Client: client, Ctx: ctx}, nil
}

// HSet sets multiple hash fields to multiple values.
func (r *RedisClient) HSet(key string, values map[string]interface{}) error {
	if len(values) == 0 {
		return fmt.Errorf("no values provided for HSet")
	}
	return r.Client.HSet(r.Ctx, key, values).Err()
}

// GetExpiry set expiry to key
func (r *RedisClient) SetExpiry(key string, expiry time.Duration) error {
	err := r.Client.Expire(r.Ctx, key, expiry).Err()
	if err != nil {
		return fmt.Errorf("failed to set expiry to key:%s , err:%w", key, err)
	}
	return nil

}
