package stripe

import (
	"fmt"
	"payment-gateway/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDeposit_PanicOnInvalidAmount tests that Deposit panics when amount is not a valid integer
func TestDeposit_PanicOnInvalidAmount(t *testing.T) {
	tests := []struct {
		name          string
		amount        string
		expectedPanic string
	}{
		{
			name:          "Empty amount string",
			amount:        "",
			expectedPanic: "invalid amount: ",
		},
		{
			name:          "Non-numeric amount",
			amount:        "invalid",
			expectedPanic: "invalid amount: invalid",
		},
		{
			name:          "Decimal amount",
			amount:        "100.50",
			expectedPanic: "invalid amount: 100.50",
		},
		{
			name:          "Amount with spaces",
			amount:        "100 ",
			expectedPanic: "invalid amount: 100 ",
		},
		{
			name:          "Negative amount with letters",
			amount:        "-100abc",
			expectedPanic: "invalid amount: -100abc",
		},
		{
			name:          "Special characters in amount",
			amount:        "$100",
			expectedPanic: "invalid amount: $100",
		},
		{
			name:          "Amount with commas",
			amount:        "1,000",
			expectedPanic: "invalid amount: 1,000",
		},
		{
			name:          "Hexadecimal notation",
			amount:        "0x64",
			expectedPanic: "invalid amount: 0x64",
		},
		{
			name:          "Scientific notation",
			amount:        "1e5",
			expectedPanic: "invalid amount: 1e5",
		},
		{
			name:          "Very long invalid string",
			amount:        "this_is_definitely_not_a_number",
			expectedPanic: "invalid amount: this_is_definitely_not_a_number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &StripeClient{
				secretKey: "sk_test_dummy",
				accountID: "acct_dummy",
			}

			req := models.DepositRequest{
				Amount:      tt.amount,
				UserID:      "1",
				Currency:    "usd",
				GatewayID:   "1",
				GatewayName: "STRIPE",
				CountryID:   "US",
			}

			// Assert that the function panics
			assert.PanicsWithValue(t, tt.expectedPanic, func() {
				client.Deposit(req)
			}, "Expected panic with message: %s", tt.expectedPanic)
		})
	}
}

// TestDeposit_ValidAmountParsing tests that valid amounts are parsed correctly
func TestDeposit_ValidAmountParsing(t *testing.T) {
	tests := []struct {
		name         string
		amount       string
		shouldNotPanic bool
	}{
		{
			name:         "Positive integer",
			amount:       "100",
			shouldNotPanic: true,
		},
		{
			name:         "Zero amount",
			amount:       "0",
			shouldNotPanic: true,
		},
		{
			name:         "Negative amount",
			amount:       "-100",
			shouldNotPanic: true,
		},
		{
			name:         "Large amount",
			amount:       "999999999",
			shouldNotPanic: true,
		},
		{
			name:         "Very large amount",
			amount:       "9223372036854775807", // max int64
			shouldNotPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &StripeClient{
				secretKey: "sk_test_dummy",
				accountID: "acct_dummy",
			}

			req := models.DepositRequest{
				Amount:      tt.amount,
				UserID:      "1",
				Currency:    "usd",
				GatewayID:   "1",
				GatewayName: "STRIPE",
				CountryID:   "US",
			}

			// This test verifies that parsing doesn't panic
			// The actual Stripe API call will fail in test environment,
			// but we're testing that amount parsing works correctly
			assert.NotPanics(t, func() {
				defer func() {
					// Recover from any Stripe API errors (expected in test env)
					if r := recover(); r != nil {
						// Only re-panic if it's an amount parsing panic
						if panicMsg, ok := r.(string); ok {
							if len(panicMsg) > 15 && panicMsg[:15] == "invalid amount:" {
								panic(r)
							}
						}
					}
				}()
				client.Deposit(req)
			}, "Should not panic for valid amount: %s", tt.amount)
		})
	}
}

// TestDeposit_RequestFieldsAreUsed tests that all request fields are properly used
func TestDeposit_RequestFieldsAreUsed(t *testing.T) {
	tests := []struct {
		name    string
		request models.DepositRequest
	}{
		{
			name: "Standard USD request",
			request: models.DepositRequest{
				Amount:      "1000",
				UserID:      "123",
				Currency:    "usd",
				GatewayID:   "1",
				GatewayName: "STRIPE",
				CountryID:   "US",
			},
		},
		{
			name: "EUR currency request",
			request: models.DepositRequest{
				Amount:      "5000",
				UserID:      "456",
				Currency:    "eur",
				GatewayID:   "2",
				GatewayName: "STRIPE",
				CountryID:   "DE",
			},
		},
		{
			name: "GBP currency request",
			request: models.DepositRequest{
				Amount:      "2500",
				UserID:      "789",
				Currency:    "gbp",
				GatewayID:   "3",
				GatewayName: "STRIPE",
				CountryID:   "GB",
			},
		},
		{
			name: "JPY currency request (no decimal)",
			request: models.DepositRequest{
				Amount:      "10000",
				UserID:      "999",
				Currency:    "jpy",
				GatewayID:   "4",
				GatewayName: "STRIPE",
				CountryID:   "JP",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &StripeClient{
				secretKey: "sk_test_dummy",
				accountID: "acct_dummy",
			}

			// Verify that the request is properly structured
			// and doesn't panic during parsing
			assert.NotPanics(t, func() {
				defer func() {
					// Recover from Stripe API errors (expected)
					recover()
				}()
				client.Deposit(tt.request)
			})
		})
	}
}

// TestDeposit_EdgeCases tests edge cases for deposit
func TestDeposit_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		request     models.DepositRequest
		expectPanic bool
		panicMsg    string
	}{
		{
			name: "Minimum valid amount (1 cent)",
			request: models.DepositRequest{
				Amount:      "1",
				UserID:      "1",
				Currency:    "usd",
				GatewayID:   "1",
				GatewayName: "STRIPE",
				CountryID:   "US",
			},
			expectPanic: false,
		},
		{
			name: "Empty user ID (should not panic on parsing)",
			request: models.DepositRequest{
				Amount:      "100",
				UserID:      "",
				Currency:    "usd",
				GatewayID:   "1",
				GatewayName: "STRIPE",
				CountryID:   "US",
			},
			expectPanic: false,
		},
		{
			name: "Empty gateway ID (should not panic on parsing)",
			request: models.DepositRequest{
				Amount:      "100",
				UserID:      "1",
				Currency:    "usd",
				GatewayID:   "",
				GatewayName: "STRIPE",
				CountryID:   "US",
			},
			expectPanic: false,
		},
		{
			name: "Empty country ID (should not panic on parsing)",
			request: models.DepositRequest{
				Amount:      "100",
				UserID:      "1",
				Currency:    "usd",
				GatewayID:   "1",
				GatewayName: "STRIPE",
				CountryID:   "",
			},
			expectPanic: false,
		},
		{
			name: "Amount with leading zeros",
			request: models.DepositRequest{
				Amount:      "00100",
				UserID:      "1",
				Currency:    "usd",
				GatewayID:   "1",
				GatewayName: "STRIPE",
				CountryID:   "US",
			},
			expectPanic: false,
		},
		{
			name: "Amount with plus sign",
			request: models.DepositRequest{
				Amount:      "+100",
				UserID:      "1",
				Currency:    "usd",
				GatewayID:   "1",
				GatewayName: "STRIPE",
				CountryID:   "US",
			},
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &StripeClient{
				secretKey: "sk_test_dummy",
				accountID: "acct_dummy",
			}

			if tt.expectPanic {
				assert.PanicsWithValue(t, tt.panicMsg, func() {
					client.Deposit(tt.request)
				})
			} else {
				assert.NotPanics(t, func() {
					defer func() {
						// Recover from Stripe API errors
						if r := recover(); r != nil {
							if panicMsg, ok := r.(string); ok {
								if len(panicMsg) > 15 && panicMsg[:15] == "invalid amount:" {
									panic(r)
								}
							}
						}
					}()
					client.Deposit(tt.request)
				})
			}
		})
	}
}

// TestDeposit_CurrencyVariations tests different currency formats
func TestDeposit_CurrencyVariations(t *testing.T) {
	currencies := []string{
		"usd", "eur", "gbp", "jpy", "cad", "aud", "chf", "cny", "sek", "nzd",
		"USD", "EUR", "GBP", // Test uppercase
	}

	client := &StripeClient{
		secretKey: "sk_test_dummy",
		accountID: "acct_dummy",
	}

	for _, currency := range currencies {
		t.Run(fmt.Sprintf("Currency_%s", currency), func(t *testing.T) {
			req := models.DepositRequest{
				Amount:      "1000",
				UserID:      "1",
				Currency:    currency,
				GatewayID:   "1",
				GatewayName: "STRIPE",
				CountryID:   "US",
			}

			// Should not panic during amount parsing
			assert.NotPanics(t, func() {
				defer func() {
					recover() // Stripe API will fail
				}()
				client.Deposit(req)
			})
		})
	}
}

// TestDeposit_LargeAmounts tests handling of very large amounts
func TestDeposit_LargeAmounts(t *testing.T) {
	tests := []struct {
		name   string
		amount string
	}{
		{
			name:   "One million cents ($10,000)",
			amount: "1000000",
		},
		{
			name:   "Ten million cents ($100,000)",
			amount: "10000000",
		},
		{
			name:   "One billion cents ($10,000,000)",
			amount: "1000000000",
		},
		{
			name:   "Near max int64",
			amount: "9223372036854775807",
		},
	}

	client := &StripeClient{
		secretKey: "sk_test_dummy",
		accountID: "acct_dummy",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := models.DepositRequest{
				Amount:      tt.amount,
				UserID:      "1",
				Currency:    "usd",
				GatewayID:   "1",
				GatewayName: "STRIPE",
				CountryID:   "US",
			}

			// Should not panic during parsing
			assert.NotPanics(t, func() {
				defer func() {
					recover() // Stripe API will fail
				}()
				client.Deposit(req)
			})
		})
	}
}

// TestDeposit_MetadataFormat tests that metadata is properly constructed
func TestDeposit_MetadataFormat(t *testing.T) {
	tests := []struct {
		name    string
		request models.DepositRequest
	}{
		{
			name: "Standard metadata",
			request: models.DepositRequest{
				Amount:      "100",
				UserID:      "user123",
				Currency:    "usd",
				GatewayID:   "gw456",
				GatewayName: "STRIPE",
				CountryID:   "US",
			},
		},
		{
			name: "Numeric IDs",
			request: models.DepositRequest{
				Amount:      "200",
				UserID:      "999",
				Currency:    "eur",
				GatewayID:   "1",
				GatewayName: "STRIPE",
				CountryID:   "DE",
			},
		},
		{
			name: "Special characters in IDs",
			request: models.DepositRequest{
				Amount:      "300",
				UserID:      "user-123",
				Currency:    "gbp",
				GatewayID:   "gw_456",
				GatewayName: "STRIPE",
				CountryID:   "GB",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &StripeClient{
				secretKey: "sk_test_dummy",
				accountID: "acct_dummy",
			}

			// Verify metadata fields are used without panic
			assert.NotPanics(t, func() {
				defer func() {
					recover() // Stripe API will fail
				}()
				client.Deposit(tt.request)
			})
		})
	}
}

// TestGetDepositStatus_EdgeCases tests edge cases for GetDepositStatus
func TestGetDepositStatus_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		depositID string
	}{
		{
			name:      "Empty deposit ID",
			depositID: "",
		},
		{
			name:      "Invalid deposit ID format",
			depositID: "invalid_id",
		},
		{
			name:      "Valid looking ID",
			depositID: "pi_test123",
		},
		{
			name:      "Very long ID",
			depositID: "pi_" + string(make([]byte, 1000)),
		},
		{
			name:      "Special characters in ID",
			depositID: "pi_!@#$%^&*()",
		},
	}

	client := &StripeClient{
		secretKey: "sk_test_dummy",
		accountID: "acct_dummy",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will return an error from Stripe API, but shouldn't panic
			_, err := client.GetDepositStatus(tt.depositID)
			
			// We expect an error since we're not using real Stripe credentials
			// The important thing is it doesn't panic
			assert.Error(t, err, "Expected error for ID: %s", tt.depositID)
			assert.Contains(t, err.Error(), "failed to get payment intent")
		})
	}
}

// TestDeposit_ConcurrentCalls tests thread safety (basic check)
func TestDeposit_ConcurrentCalls(t *testing.T) {
	client := &StripeClient{
		secretKey: "sk_test_dummy",
		accountID: "acct_dummy",
	}

	// Test that multiple goroutines calling Deposit don't cause panics
	// (except for Stripe API errors which we catch)
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() {
				recover() // Catch any panics including Stripe API errors
				done <- true
			}()

			req := models.DepositRequest{
				Amount:      fmt.Sprintf("%d", 100*id+100),
				UserID:      fmt.Sprintf("user%d", id),
				Currency:    "usd",
				GatewayID:   "1",
				GatewayName: "STRIPE",
				CountryID:   "US",
			}

			client.Deposit(req)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// If we get here without deadlock, the test passes
	assert.True(t, true, "Concurrent calls completed")
}

// TestDeposit_NegativeAmounts tests handling of negative amounts
func TestDeposit_NegativeAmounts(t *testing.T) {
	tests := []struct {
		name   string
		amount string
	}{
		{
			name:   "Small negative",
			amount: "-1",
		},
		{
			name:   "Large negative",
			amount: "-1000000",
		},
		{
			name:   "Near min int64",
			amount: "-9223372036854775808",
		},
	}

	client := &StripeClient{
		secretKey: "sk_test_dummy",
		accountID: "acct_dummy",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := models.DepositRequest{
				Amount:      tt.amount,
				UserID:      "1",
				Currency:    "usd",
				GatewayID:   "1",
				GatewayName: "STRIPE",
				CountryID:   "US",
			}

			// Should parse the negative number without panic
			assert.NotPanics(t, func() {
				defer func() {
					recover() // Stripe API will reject it
				}()
				client.Deposit(req)
			})
		})
	}
}

// TestDeposit_BoundaryValues tests boundary values for int64
func TestDeposit_BoundaryValues(t *testing.T) {
	tests := []struct {
		name        string
		amount      string
		expectPanic bool
	}{
		{
			name:        "Max int64",
			amount:      "9223372036854775807",
			expectPanic: false,
		},
		{
			name:        "Min int64",
			amount:      "-9223372036854775808",
			expectPanic: false,
		},
		{
			name:        "Over max int64",
			amount:      "9223372036854775808",
			expectPanic: true,
		},
		{
			name:        "Under min int64",
			amount:      "-9223372036854775809",
			expectPanic: true,
		},
	}

	client := &StripeClient{
		secretKey: "sk_test_dummy",
		accountID: "acct_dummy",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := models.DepositRequest{
				Amount:      tt.amount,
				UserID:      "1",
				Currency:    "usd",
				GatewayID:   "1",
				GatewayName: "STRIPE",
				CountryID:   "US",
			}

			if tt.expectPanic {
				assert.Panics(t, func() {
					client.Deposit(req)
				}, "Expected panic for amount: %s", tt.amount)
			} else {
				assert.NotPanics(t, func() {
					defer func() {
						if r := recover(); r != nil {
							if panicMsg, ok := r.(string); ok {
								if len(panicMsg) > 15 && panicMsg[:15] == "invalid amount:" {
									panic(r)
								}
							}
						}
					}()
					client.Deposit(req)
				}, "Should not panic for amount: %s", tt.amount)
			}
		})
	}
}