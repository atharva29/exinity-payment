package redis

import (
	"context"
	"fmt"
	"payment-gateway/internal/models"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedisClient_GetGatewaysByCountry(t *testing.T) {
	// Create mock Redis client
	db, mock := redismock.NewClientMock()
	client := &RedisClient{Client: db}
	ctx := context.Background()

	tests := []struct {
		name           string
		countryID      string
		prepareMock    func(countryID string)
		expectedResult []models.Gateway
		expectError    bool
	}{
		{
			name:      "successful retrieval with multiple gateways",
			countryID: "1",
			prepareMock: func(countryID string) {
				pattern := fmt.Sprintf("gateway-by-country:%s:*", countryID)
				mock.ExpectKeys(pattern).SetVal([]string{
					fmt.Sprintf("gateway-by-country:%s:100", countryID),
					fmt.Sprintf("gateway-by-country:%s:101", countryID),
				})

				// Mock first gateway HGetAll
				key1 := fmt.Sprintf("gateway-by-country:%s:100", countryID)
				mock.ExpectHGetAll(key1).SetVal(map[string]string{
					"gateway_name": "Gateway 1",
					"score":        "90",
				})
				mock.ExpectHGet(key1, "score").SetVal("90")

				// Mock second gateway HGetAll
				key2 := fmt.Sprintf("gateway-by-country:%s:101", countryID)
				mock.ExpectHGetAll(key2).SetVal(map[string]string{
					"gateway_name": "Gateway 2",
					"score":        "85",
				})
				mock.ExpectHGet(key2, "score").SetVal("85")
			},
			expectedResult: []models.Gateway{
				{ID: "100", Name: "Gateway 1", Score: 90},
				{ID: "101", Name: "Gateway 2", Score: 85},
			},
			expectError: false,
		},
		{
			name:      "empty result",
			countryID: "2",
			prepareMock: func(countryID string) {
				pattern := fmt.Sprintf("gateway-by-country:%s:*", countryID)
				mock.ExpectKeys(pattern).SetVal([]string{})
			},
			expectedResult: nil,
			expectError:    false,
		},
		{
			name:      "keys error",
			countryID: "3",
			prepareMock: func(countryID string) {
				pattern := fmt.Sprintf("gateway-by-country:%s:*", countryID)
				mock.ExpectKeys(pattern).SetErr(fmt.Errorf("redis error"))
			},
			expectedResult: nil,
			expectError:    true,
		},
		{
			name:      "hgetall error",
			countryID: "4",
			prepareMock: func(countryID string) {
				pattern := fmt.Sprintf("gateway-by-country:%s:*", countryID)
				key := fmt.Sprintf("gateway-by-country:%s:200", countryID)
				mock.ExpectKeys(pattern).SetVal([]string{key})
				mock.ExpectHGetAll(key).SetErr(fmt.Errorf("hgetall error"))
			},
			expectedResult: nil,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations and prepare new ones
			mock.ClearExpect()
			tt.prepareMock(tt.countryID)

			// Execute the function
			result, err := client.GetGatewaysByCountry(ctx, tt.countryID)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			// Check if all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
