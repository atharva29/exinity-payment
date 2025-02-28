package redis

import (
	"context"
	"fmt"
	"payment-gateway/internal/models"
	"time"
)

// GetGatewaysByCountryFromRedisSet fetches gateways from Redis Hash Set
func (r *RedisClient) GetGatewaysByCountryFromRedisSet(ctx context.Context, countryID string) ([]models.Gateway, error) {
	key := "gateway-by-country:" + countryID
	gatewayData, err := r.Client.HGetAll(ctx, key).Result()
	if err != nil || len(gatewayData) == 0 {
		return nil, fmt.Errorf("no gateways found in Redis for country %s", countryID)
	}

	var gateways []models.Gateway
	for id, name := range gatewayData {
		gateways = append(gateways, models.Gateway{ID: id, Name: name})
	}

	return gateways, nil
}

// SaveGatewaysToRedisHashSet stores gateways as a Redis Hash Set
func (r *RedisClient) SaveGatewaysToRedisHashSet(ctx context.Context, countryID string, gateways []models.Gateway) error {
	key := "gateway-by-country:" + countryID
	pipeline := r.Client.Pipeline()

	for _, gateway := range gateways {
		pipeline.HSet(ctx, key, gateway.ID, gateway.Name)
	}

	_, err := pipeline.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to save gateways to Redis: %v", err)
	}

	_ = r.Client.Expire(ctx, key, 24*time.Hour).Err() // Set expiry
	return nil
}

// GetGatewaysSortedByScore retrieves gateways sorted by score
func (r *RedisClient) GetGatewaysSortedByScore(ctx context.Context, gateways []models.Gateway) ([]models.Gateway, error) {
	sortedGateways := make([]models.Gateway, len(gateways))
	for i, gateway := range gateways {
		sortedGateways[i] = models.Gateway{ID: gateway.ID, Name: gateway.Name}
	}

	return sortedGateways, nil
}
