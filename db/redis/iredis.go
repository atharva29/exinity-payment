package redis

import (
	"context"
	"payment-gateway/internal/models"
)

type IRedis interface {
	SaveGatewaysToRedisHashSet(ctx context.Context, countryID string, gateways []models.Gateway) error
	GetGatewaysByCountry(ctx context.Context, countryID string) ([]models.Gateway, error)
	IncrementGatewayScore(ctx context.Context, countryID string, gatewayID string) error
	DecrementGatewayScore(ctx context.Context, countryID string, gatewayID string) error
	HSet(ctx context.Context, key string, values map[string]interface{}) error
}
