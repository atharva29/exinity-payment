package redis

import (
	"context"
	"fmt"
	"payment-gateway/internal/models"
	"sort"
)

func (s *RedisClient) SaveGatewaysToRedisHashSet(ctx context.Context, countryID string, gateways []models.Gateway) error {
	pipeline := s.client.Pipeline()
	for _, gateway := range gateways {
		key := fmt.Sprintf("gateway-by-country:%s:%s", countryID, gateway.ID)
		pipeline.HSet(ctx, key, "gateway_name", gateway.Name, "score", 1)
	}
	_, err := pipeline.Exec(ctx)
	return err
}

func (s *RedisClient) GetGatewaysByCountry(ctx context.Context, countryID string) ([]models.Gateway, error) {
	pattern := fmt.Sprintf("gateway-by-country:%s:*", countryID)
	keys, err := s.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	var gateways []models.Gateway
	for _, key := range keys {
		data, err := s.client.HGetAll(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		if len(data) == 0 {
			continue
		}

		score, _ := s.client.HGet(ctx, key, "score").Int()
		gateway := models.Gateway{
			ID:    key[len(pattern)-1:], // Extracting gateway ID from key
			Name:  data["gateway_name"],
			Score: score,
		}
		gateways = append(gateways, gateway)
	}

	sort.Slice(gateways, func(i, j int) bool {
		return gateways[i].Score > gateways[j].Score
	})

	return gateways, nil
}

func (s *RedisClient) IncrementGatewayScore(ctx context.Context, countryID string, gatewayID string) error {
	key := fmt.Sprintf("gateway-by-country:%s:%s", countryID, gatewayID)
	return s.client.HIncrBy(ctx, key, "score", 1).Err()
}

func (s *RedisClient) DecrementGatewayScore(ctx context.Context, countryID string, gatewayID string) error {
	key := fmt.Sprintf("gateway-by-country:%s:%s", countryID, gatewayID)
	return s.client.HIncrBy(ctx, key, "score", -1).Err()
}
