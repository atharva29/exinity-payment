package stripe

import (
	"context"
	"encoding/json"
	"errors"
	"payment-gateway/db"
	database "payment-gateway/db/db"
	"payment-gateway/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	event "github.com/stripe/stripe-go"
	stripe "github.com/stripe/stripe-go/v81"
)

// MockDB implements IDB interface for testing
type MockDB struct {
	mock.Mock
}

func (m *MockDB) CheckUserBalance(userID int, currency string, amount float64) (bool, float64, error) {
	args := m.Called(userID, currency, amount)
	return args.Bool(0), args.Get(1).(float64), args.Error(2)
}

func (m *MockDB) GetSupportedGatewaysByCountries(countryID string) ([]models.Gateway, error) {
	args := m.Called(countryID)
	return args.Get(0).([]models.Gateway), args.Error(1)
}

func (m *MockDB) CreateTransaction(transaction database.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

// MockRedis implements IRedis interface for testing
type MockRedis struct {
	mock.Mock
}

func (m *MockRedis) SaveGatewaysToRedisHashSet(ctx context.Context, countryID string, gateways []models.Gateway) error {
	args := m.Called(ctx, countryID, gateways)
	return args.Error(0)
}

func (m *MockRedis) GetGatewaysByCountry(ctx context.Context, countryID string) ([]models.Gateway, error) {
	args := m.Called(ctx, countryID)
	return args.Get(0).([]models.Gateway), args.Error(1)
}

func (m *MockRedis) IncrementGatewayScore(ctx context.Context, countryID string, gatewayID string) error {
	args := m.Called(ctx, countryID, gatewayID)
	return args.Error(0)
}

func (m *MockRedis) DecrementGatewayScore(ctx context.Context, countryID string, gatewayID string) error {
	args := m.Called(ctx, countryID, gatewayID)
	return args.Error(0)
}

func (m *MockRedis) HSet(ctx context.Context, key string, values map[string]interface{}) error {
	args := m.Called(ctx, key, values)
	return args.Error(0)
}

func TestHandlePaymentIntentSucceeded(t *testing.T) {
	tests := []struct {
		name          string
		event         *event.Event
		setupMocks    func(*MockDB, *MockRedis)
		expectedError error
	}{
		{
			name: "Successful payment processing",
			event: &event.Event{
				ID: "evt_123",
				Data: &event.EventData{ // Changed to non-pointer type as per error
					Raw: mustMarshal(t, stripe.PaymentIntent{
						ID:       "pi_123",
						Amount:   1000,
						Currency: "usd",
						Metadata: map[string]string{
							"user_id":    "1",
							"gateway_id": "7",
							"country_id": "1",
						},
					}),
				},
			},
			setupMocks: func(db *MockDB, redis *MockRedis) {
				redis.On("HSet", mock.Anything, "deposit:userid:1:orderid:pi_123", map[string]interface{}{
					"status": "success",
				}).Return(nil)
				redis.On("IncrementGatewayScore", mock.Anything, "1", "7").Return(nil)
				db.On("CreateTransaction", mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Invalid metadata",
			event: &event.Event{
				ID: "evt_123",
				Data: &event.EventData{ // Changed to non-pointer per stripe-go convention
					Raw: mustMarshal(t, stripe.PaymentIntent{
						ID:       "pi_123",
						Amount:   1000,
						Currency: "usd",
						Metadata: map[string]string{
							"user_id": "1",
							// Missing required fields
						},
					}),
				},
			},
			setupMocks:    func(db *MockDB, redis *MockRedis) {},
			expectedError: errors.New("invalid gateway_id format: strconv.Atoi: parsing \"\": invalid syntax"),
		},
		{
			name: "Redis HSet failure",
			event: &event.Event{
				ID: "evt_123",
				Data: &event.EventData{
					Raw: mustMarshal(t, stripe.PaymentIntent{
						ID:       "pi_123",
						Amount:   1000,
						Currency: "usd",
						Metadata: map[string]string{
							"user_id":    "1",
							"gateway_id": "7",
							"country_id": "1",
						},
					}),
				},
			},
			setupMocks: func(db *MockDB, redis *MockRedis) {
				redis.On("HSet", mock.Anything, "deposit:userid:1:orderid:pi_123", mock.Anything).
					Return(errors.New("redis error"))
			},
			expectedError: errors.New("failed to store data in redis: redis error"),
		},
		{
			name: "DB transaction failure",
			event: &event.Event{
				ID: "evt_123",
				Data: &event.EventData{
					Raw: mustMarshal(t, stripe.PaymentIntent{
						ID:       "pi_123",
						Amount:   1000,
						Currency: "usd",
						Metadata: map[string]string{
							"user_id":    "1",
							"gateway_id": "7",
							"country_id": "1",
						},
					}),
				},
			},
			setupMocks: func(db *MockDB, redis *MockRedis) {
				redis.On("HSet", mock.Anything, "deposit:userid:1:orderid:pi_123", mock.Anything).Return(nil)
				redis.On("IncrementGatewayScore", mock.Anything, "1", "7").Return(nil)
				db.On("CreateTransaction", mock.Anything).Return(errors.New("db error"))
			},
			expectedError: errors.New("failed to store withdrawal transaction data in db: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockDB)
			mockRedis := new(MockRedis)
			db := &db.DB{DB: mockDB, Redis: mockRedis}
			client := &StripeClient{}

			tt.setupMocks(mockDB, mockRedis)

			err := client.handlePaymentIntentSucceeded(tt.event, db)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
			mockRedis.AssertExpectations(t)
		})
	}
}

func TestHandlePaymentIntentFailed(t *testing.T) {
	tests := []struct {
		name          string
		event         *event.Event
		setupMocks    func(*MockDB, *MockRedis)
		expectedError error
	}{
		{
			name: "Successful payment failure handling",
			event: &event.Event{
				ID: "evt_123",
				Data: &event.EventData{
					Raw: mustMarshal(t, stripe.PaymentIntent{
						ID:       "pi_123",
						Amount:   1000,
						Currency: "usd",
						Metadata: map[string]string{
							"user_id":    "1",
							"gateway_id": "7",
							"country_id": "2",
						},
					}),
				},
			},
			setupMocks: func(db *MockDB, redis *MockRedis) {
				redis.On("HSet", mock.Anything, "deposit:userid:1:orderid:pi_123", map[string]interface{}{
					"status": "failed",
				}).Return(nil)
				redis.On("DecrementGatewayScore", mock.Anything, "2", "7").Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Invalid JSON data",
			event: &event.Event{
				ID: "evt_123",
				Data: &event.EventData{
					Raw: []byte("{invalid json}"),
				},
			},
			setupMocks:    func(db *MockDB, redis *MockRedis) {},
			expectedError: errors.New("invalid character 'i' looking for beginning of object key string"),
		},
		{
			name: "Missing metadata fields",
			event: &event.Event{
				ID: "evt_123",
				Data: &event.EventData{
					Raw: mustMarshal(t, stripe.PaymentIntent{
						ID:       "pi_123",
						Amount:   1000,
						Currency: "usd",
						Metadata: map[string]string{
							"user_id": "1",
							// Missing gateway_id and country_id
						},
					}),
				},
			},
			setupMocks:    func(db *MockDB, redis *MockRedis) {},
			expectedError: errors.New("invalid gateway_id format: strconv.Atoi: parsing \"\": invalid syntax"),
		},
		{
			name: "Invalid gateway_id format",
			event: &event.Event{
				ID: "evt_123",
				Data: &event.EventData{
					Raw: mustMarshal(t, stripe.PaymentIntent{
						ID:       "pi_123",
						Amount:   1000,
						Currency: "usd",
						Metadata: map[string]string{
							"user_id":    "1",
							"gateway_id": "not-a-number",
							"country_id": "2",
						},
					}),
				},
			},
			setupMocks:    func(db *MockDB, redis *MockRedis) {},
			expectedError: errors.New("invalid gateway_id format: strconv.Atoi: parsing \"not-a-number\": invalid syntax"),
		},
		{
			name: "Redis HSet failure",
			event: &event.Event{
				ID: "evt_123",
				Data: &event.EventData{
					Raw: mustMarshal(t, stripe.PaymentIntent{
						ID:       "pi_123",
						Amount:   1000,
						Currency: "usd",
						Metadata: map[string]string{
							"user_id":    "1",
							"gateway_id": "7",
							"country_id": "2",
						},
					}),
				},
			},
			setupMocks: func(db *MockDB, redis *MockRedis) {
				redis.On("HSet", mock.Anything, "deposit:userid:1:orderid:pi_123", map[string]interface{}{
					"status": "failed",
				}).Return(errors.New("redis error"))
			},
			expectedError: errors.New("failed to store data in redis: redis error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockDB)
			mockRedis := new(MockRedis)
			db := &db.DB{DB: mockDB, Redis: mockRedis}
			client := &StripeClient{}

			tt.setupMocks(mockDB, mockRedis)

			err := client.handlePaymentIntentFailed(tt.event, db)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
			mockRedis.AssertExpectations(t)
		})
	}
}

func TestHandlePaymentIntentCreated(t *testing.T) {
	tests := []struct {
		name          string
		event         *event.Event
		setupMocks    func(*MockRedis)
		expectedError error
	}{
		{
			name: "Successful payment intent creation",
			event: &event.Event{
				ID: "evt_123",
				Data: &event.EventData{
					Raw: mustMarshal(t, stripe.PaymentIntent{
						ID:       "pi_123",
						Amount:   1000,
						Currency: "usd",
						Metadata: map[string]string{
							"user_id": "1",
						},
					}),
				},
			},
			setupMocks: func(redis *MockRedis) {
				redis.On("HSet", mock.Anything, "deposit:userid:1:orderid:pi_123", map[string]interface{}{
					"status": "pending",
				}).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Invalid JSON data",
			event: &event.Event{
				ID: "evt_123",
				Data: &event.EventData{
					Raw: []byte("{invalid json}"),
				},
			},
			setupMocks:    func(redis *MockRedis) {},
			expectedError: errors.New("invalid character 'i' looking for beginning of object key string"),
		},
		{
			name: "Redis HSet failure",
			event: &event.Event{
				ID: "evt_123",
				Data: &event.EventData{
					Raw: mustMarshal(t, stripe.PaymentIntent{
						ID:       "pi_123",
						Amount:   1000,
						Currency: "usd",
						Metadata: map[string]string{
							"user_id": "1",
						},
					}),
				},
			},
			setupMocks: func(redis *MockRedis) {
				redis.On("HSet", mock.Anything, "deposit:userid:1:orderid:pi_123", map[string]interface{}{
					"status": "pending",
				}).Return(errors.New("redis error"))
			},
			expectedError: errors.New("failed to store data in redis: redis error"),
		},
		{
			name: "Missing user_id in metadata",
			event: &event.Event{
				ID: "evt_123",
				Data: &event.EventData{
					Raw: mustMarshal(t, stripe.PaymentIntent{
						ID:       "pi_123",
						Amount:   1000,
						Currency: "usd",
						Metadata: map[string]string{}, // No user_id
					}),
				},
			},
			setupMocks: func(redis *MockRedis) {
				redis.On("HSet", mock.Anything, "deposit:userid::orderid:pi_123", map[string]interface{}{
					"status": "pending",
				}).Return(nil)
			},
			expectedError: nil, // Function doesn't validate this, so it should still succeed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRedis := new(MockRedis)
			client := &StripeClient{}

			tt.setupMocks(mockRedis)

			err := client.handlePaymentIntentCreated(tt.event, mockRedis)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRedis.AssertExpectations(t)
		})
	}
}

func TestHandlePayoutPaid(t *testing.T) {
	tests := []struct {
		name          string
		event         *event.Event
		setupMocks    func(*MockDB, *MockRedis) // Changed to pointer types
		expectedError error
	}{
		{
			name: "Successful payout with all metadata",
			event: &event.Event{
				ID: "evt_123",
				Data: &event.EventData{
					Raw: mustMarshal(t, stripe.Payout{
						ID:       "po_123",
						Amount:   5000,
						Currency: "usd",
						Metadata: map[string]string{
							"user_id":    "1",
							"gateway_id": "7",
							"country_id": "2",
						},
					}),
				},
			},
			setupMocks: func(db *MockDB, redis *MockRedis) {
				redis.On("HSet", mock.Anything, "withdrawal:userid:1:payoutid:po_123", mock.MatchedBy(func(data map[string]interface{}) bool {
					return data["status"] == "completed" &&
						assert.IsType(t, int64(0), data["completed_at"]) &&
						data["completed_at"].(int64) <= time.Now().Unix()
				})).Return(nil)
				redis.On("IncrementGatewayScore", mock.Anything, "2", "7").Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Successful payout with only user_id",
			event: &event.Event{
				ID: "evt_123",
				Data: &event.EventData{
					Raw: mustMarshal(t, stripe.Payout{
						ID:       "po_123",
						Amount:   5000,
						Currency: "usd",
						Metadata: map[string]string{
							"user_id": "1",
						},
					}),
				},
			},
			setupMocks: func(db *MockDB, redis *MockRedis) {
				redis.On("HSet", mock.Anything, "withdrawal:userid:1:payoutid:po_123", mock.MatchedBy(func(data map[string]interface{}) bool {
					return data["status"] == "completed" &&
						assert.IsType(t, int64(0), data["completed_at"])
				})).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Invalid JSON data",
			event: &event.Event{
				ID: "evt_123",
				Data: &event.EventData{
					Raw: []byte("{invalid json}"),
				},
			},
			setupMocks:    func(db *MockDB, redis *MockRedis) {},
			expectedError: errors.New("invalid character 'i' looking for beginning of object key string"),
		},
		{
			name: "Redis HSet failure",
			event: &event.Event{
				ID: "evt_123",
				Data: &event.EventData{
					Raw: mustMarshal(t, stripe.Payout{
						ID:       "po_123",
						Amount:   5000,
						Currency: "usd",
						Metadata: map[string]string{
							"user_id":    "1",
							"gateway_id": "7",
							"country_id": "2",
						},
					}),
				},
			},
			setupMocks: func(db *MockDB, redis *MockRedis) {
				redis.On("HSet", mock.Anything, "withdrawal:userid:1:payoutid:po_123", mock.Anything).
					Return(errors.New("redis error"))
			},
			expectedError: errors.New("failed to update withdrawal data in redis: redis error"),
		},
		{
			name: "Missing user_id in metadata",
			event: &event.Event{
				ID: "evt_123",
				Data: &event.EventData{
					Raw: mustMarshal(t, stripe.Payout{
						ID:       "po_123",
						Amount:   5000,
						Currency: "usd",
						Metadata: map[string]string{},
					}),
				},
			},
			setupMocks: func(db *MockDB, redis *MockRedis) {
				redis.On("HSet", mock.Anything, "withdrawal:userid::payoutid:po_123", mock.MatchedBy(func(data map[string]interface{}) bool {
					return data["status"] == "completed" &&
						assert.IsType(t, int64(0), data["completed_at"])
				})).Return(nil)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockDB)
			mockRedis := new(MockRedis)
			db := &db.DB{DB: mockDB, Redis: mockRedis}
			client := &StripeClient{}

			tt.setupMocks(mockDB, mockRedis) // Pass pointers directly

			err := client.handlePayoutPaid(tt.event, db)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
			mockRedis.AssertExpectations(t)
		})
	}
}

// Helper function to marshal test data
func mustMarshal(t *testing.T, v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}
	return data
}
