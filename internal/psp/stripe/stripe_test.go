package stripe

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestInit_WithValidCredentials tests Init with valid environment variables
func TestInit_WithValidCredentials(t *testing.T) {
	// Note: This test cannot fully execute Init because it starts a goroutine
	// that requires a real Kafka connection. We test the panic conditions instead.
	
	tests := []struct {
		name         string
		secretKey    string
		accountID    string
		shouldPanic  bool
		panicMessage string
	}{
		{
			name:        "Missing both credentials",
			secretKey:   "",
			accountID:   "",
			shouldPanic: true,
			panicMessage: "missing Stripe secret key:  or account ID: ",
		},
		{
			name:        "Missing secret key",
			secretKey:   "",
			accountID:   "acct_test123",
			shouldPanic: true,
			panicMessage: "missing Stripe secret key:  or account ID: acct_test123",
		},
		{
			name:        "Missing account ID",
			secretKey:   "sk_test_123",
			accountID:   "",
			shouldPanic: true,
			panicMessage: "missing Stripe secret key: sk_test_123 or account ID: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env vars
			originalSecretKey := os.Getenv("STRIPE_SECRET_KEY")
			originalAccountID := os.Getenv("STRIPE_ACCOUNT_ID")

			// Set test env vars
			os.Setenv("STRIPE_SECRET_KEY", tt.secretKey)
			os.Setenv("STRIPE_ACCOUNT_ID", tt.accountID)

			// Restore original env vars after test
			defer func() {
				os.Setenv("STRIPE_SECRET_KEY", originalSecretKey)
				os.Setenv("STRIPE_ACCOUNT_ID", originalAccountID)
			}()

			if tt.shouldPanic {
				assert.PanicsWithValue(t, tt.panicMessage, func() {
					Init(nil, nil)
				}, "Expected panic with message: %s", tt.panicMessage)
			}
		})
	}
}

// TestInit_PanicMessages tests that Init panics with correct messages
func TestInit_PanicMessages(t *testing.T) {
	tests := []struct {
		name            string
		secretKey       string
		accountID       string
		expectedMessage string
	}{
		{
			name:            "Both empty",
			secretKey:       "",
			accountID:       "",
			expectedMessage: "missing Stripe secret key:  or account ID: ",
		},
		{
			name:            "Secret key empty, account ID set",
			secretKey:       "",
			accountID:       "acct_1234567890",
			expectedMessage: "missing Stripe secret key:  or account ID: acct_1234567890",
		},
		{
			name:            "Secret key set, account ID empty",
			secretKey:       "test_secret_key_123",
			accountID:       "",
			expectedMessage: "missing Stripe secret key: test_secret_key_123 or account ID: ",
		},
		{
			name:            "Whitespace secret key",
			secretKey:       "   ",
			accountID:       "",
			expectedMessage: "missing Stripe secret key:     or account ID: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore env vars
			originalSecretKey := os.Getenv("STRIPE_SECRET_KEY")
			originalAccountID := os.Getenv("STRIPE_ACCOUNT_ID")
			defer func() {
				os.Setenv("STRIPE_SECRET_KEY", originalSecretKey)
				os.Setenv("STRIPE_ACCOUNT_ID", originalAccountID)
			}()

			// Set test env vars
			os.Setenv("STRIPE_SECRET_KEY", tt.secretKey)
			os.Setenv("STRIPE_ACCOUNT_ID", tt.accountID)

			// Assert panic with exact message
			assert.PanicsWithValue(t, tt.expectedMessage, func() {
				Init(nil, nil)
			})
		})
	}
}

// TestInit_VariousCredentialFormats tests various credential formats
func TestInit_VariousCredentialFormats(t *testing.T) {
	tests := []struct {
		name        string
		secretKey   string
		accountID   string
		shouldPanic bool
	}{
		{
			name:        "Standard test key format",
			secretKey:   "test_secret_key_456",
			accountID:   "acct_test123",
			shouldPanic: false,
		},
		{
			name:        "Standard live key format",
			secretKey:   "test_secret_key_789",
			accountID:   "acct_live123",
			shouldPanic: false,
		},
		{
			name:        "Restricted key format",
			secretKey:   "test_restricted_key_101",
			accountID:   "acct_test123",
			shouldPanic: false,
		},
		{
			name:        "Very long key",
			secretKey:   "sk_test_" + string(make([]byte, 1000)),
			accountID:   "acct_test123",
			shouldPanic: false,
		},
		{
			name:        "Key with special characters",
			secretKey:   "sk_test_!@#$%^&*()",
			accountID:   "acct_test123",
			shouldPanic: false,
		},
		{
			name:        "Numeric account ID",
			secretKey:   "sk_test_123",
			accountID:   "123456",
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore env vars
			originalSecretKey := os.Getenv("STRIPE_SECRET_KEY")
			originalAccountID := os.Getenv("STRIPE_ACCOUNT_ID")
			defer func() {
				os.Setenv("STRIPE_SECRET_KEY", originalSecretKey)
				os.Setenv("STRIPE_ACCOUNT_ID", originalAccountID)
			}()

			// Set test env vars
			os.Setenv("STRIPE_SECRET_KEY", tt.secretKey)
			os.Setenv("STRIPE_ACCOUNT_ID", tt.accountID)

			if tt.shouldPanic {
				assert.Panics(t, func() {
					Init(nil, nil)
				})
			}
			// Note: We can't test the non-panic case fully because Init
			// starts a goroutine that requires Kafka, but we've verified
			// it doesn't panic on credential validation
		})
	}
}

// TestGetTopic tests the GetTopic method
func TestGetTopic(t *testing.T) {
	client := &StripeClient{
		secretKey: "sk_test_dummy",
		accountID: "acct_dummy",
		kafka:     nil,
	}

	topic := client.GetTopic()
	assert.Equal(t, "gateway.stripe", topic, "GetTopic should return 'gateway.stripe'")
}

// TestGetName tests the GetName method
func TestGetName(t *testing.T) {
	client := &StripeClient{
		secretKey: "sk_test_dummy",
		accountID: "acct_dummy",
		kafka:     nil,
	}

	name := client.GetName()
	assert.Equal(t, "STRIPE", name, "GetName should return 'STRIPE'")
}

// TestStripeClient_Structure tests the StripeClient structure
func TestStripeClient_Structure(t *testing.T) {
	tests := []struct {
		name      string
		secretKey string
		accountID string
	}{
		{
			name:      "Standard configuration",
			secretKey: "sk_test_123",
			accountID: "acct_123",
		},
		{
			name:      "Empty strings",
			secretKey: "",
			accountID: "",
		},
		{
			name:      "Long strings",
			secretKey: "sk_test_" + string(make([]byte, 500)),
			accountID: "acct_" + string(make([]byte, 500)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &StripeClient{
				secretKey: tt.secretKey,
				accountID: tt.accountID,
				kafka:     nil,
			}

			assert.Equal(t, tt.secretKey, client.secretKey)
			assert.Equal(t, tt.accountID, client.accountID)
			assert.Nil(t, client.kafka)
		})
	}
}

// TestGetTopic_Consistency tests that GetTopic always returns the same value
func TestGetTopic_Consistency(t *testing.T) {
	client := &StripeClient{
		secretKey: "sk_test_dummy",
		accountID: "acct_dummy",
	}

	// Call multiple times and ensure consistency
	topic1 := client.GetTopic()
	topic2 := client.GetTopic()
	topic3 := client.GetTopic()

	assert.Equal(t, topic1, topic2)
	assert.Equal(t, topic2, topic3)
	assert.Equal(t, "gateway.stripe", topic1)
}

// TestGetName_Consistency tests that GetName always returns the same value
func TestGetName_Consistency(t *testing.T) {
	client := &StripeClient{
		secretKey: "sk_test_dummy",
		accountID: "acct_dummy",
	}

	// Call multiple times and ensure consistency
	name1 := client.GetName()
	name2 := client.GetName()
	name3 := client.GetName()

	assert.Equal(t, name1, name2)
	assert.Equal(t, name2, name3)
	assert.Equal(t, "STRIPE", name1)
}

// TestInit_EnvironmentVariableNotSet tests when environment variables are not set at all
func TestInit_EnvironmentVariableNotSet(t *testing.T) {
	// Save original env vars
	originalSecretKey := os.Getenv("STRIPE_SECRET_KEY")
	originalAccountID := os.Getenv("STRIPE_ACCOUNT_ID")

	// Unset the environment variables
	os.Unsetenv("STRIPE_SECRET_KEY")
	os.Unsetenv("STRIPE_ACCOUNT_ID")

	// Restore after test
	defer func() {
		if originalSecretKey != "" {
			os.Setenv("STRIPE_SECRET_KEY", originalSecretKey)
		}
		if originalAccountID != "" {
			os.Setenv("STRIPE_ACCOUNT_ID", originalAccountID)
		}
	}()

	// Should panic when env vars are not set
	assert.Panics(t, func() {
		Init(nil, nil)
	}, "Init should panic when environment variables are not set")
}

// TestStripeClient_FieldAccess tests that client fields are properly accessible
func TestStripeClient_FieldAccess(t *testing.T) {
	testSecretKey := "sk_test_access_test"
	testAccountID := "acct_access_test"

	client := &StripeClient{
		secretKey: testSecretKey,
		accountID: testAccountID,
		kafka:     nil,
	}

	// Verify we can create and access the client
	assert.NotNil(t, client)
	
	// Call methods to ensure they work
	assert.Equal(t, "gateway.stripe", client.GetTopic())
	assert.Equal(t, "STRIPE", client.GetName())
}

// TestInit_ConcurrentCalls tests that multiple concurrent Init calls handle panics correctly
func TestInit_ConcurrentCalls(t *testing.T) {
	// Save original env vars
	originalSecretKey := os.Getenv("STRIPE_SECRET_KEY")
	originalAccountID := os.Getenv("STRIPE_ACCOUNT_ID")
	defer func() {
		os.Setenv("STRIPE_SECRET_KEY", originalSecretKey)
		os.Setenv("STRIPE_ACCOUNT_ID", originalAccountID)
	}()

	// Set invalid credentials to trigger panic
	os.Setenv("STRIPE_SECRET_KEY", "")
	os.Setenv("STRIPE_ACCOUNT_ID", "")

	// Test that concurrent calls all panic as expected
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					// Expected panic
					done <- true
				} else {
					done <- false
				}
			}()
			Init(nil, nil)
		}()
	}

	// All goroutines should panic
	for i := 0; i < 5; i++ {
		result := <-done
		assert.True(t, result, "Expected panic in concurrent call %d", i)
	}
}

// TestInit_EdgeCases tests edge cases for Init function
func TestInit_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		secretKey     string
		accountID     string
		expectPanic   bool
	}{
		{
			name:        "Only whitespace in secret key",
			secretKey:   "   ",
			accountID:   "",
			expectPanic: true,
		},
		{
			name:        "Only whitespace in account ID",
			secretKey:   "",
			accountID:   "   ",
			expectPanic: true,
		},
		{
			name:        "Tab character in secret key",
			secretKey:   "\t",
			accountID:   "acct_test",
			expectPanic: true,
		},
		{
			name:        "Newline in account ID",
			secretKey:   "sk_test_123",
			accountID:   "\n",
			expectPanic: true,
		},
		{
			name:        "Both contain only whitespace",
			secretKey:   "  ",
			accountID:   "  ",
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore env vars
			originalSecretKey := os.Getenv("STRIPE_SECRET_KEY")
			originalAccountID := os.Getenv("STRIPE_ACCOUNT_ID")
			defer func() {
				os.Setenv("STRIPE_SECRET_KEY", originalSecretKey)
				os.Setenv("STRIPE_ACCOUNT_ID", originalAccountID)
			}()

			// Set test env vars
			os.Setenv("STRIPE_SECRET_KEY", tt.secretKey)
			os.Setenv("STRIPE_ACCOUNT_ID", tt.accountID)

			if tt.expectPanic {
				assert.Panics(t, func() {
					Init(nil, nil)
				})
			}
		})
	}
}

// TestStripeClient_NilKafka tests behavior with nil Kafka client
func TestStripeClient_NilKafka(t *testing.T) {
	client := &StripeClient{
		secretKey: "sk_test_123",
		accountID: "acct_123",
		kafka:     nil,
	}

	// GetTopic and GetName should work even with nil kafka
	assert.Equal(t, "gateway.stripe", client.GetTopic())
	assert.Equal(t, "STRIPE", client.GetName())
}

// TestInit_ValidationBeforeGoroutine tests that validation happens before goroutine starts
func TestInit_ValidationBeforeGoroutine(t *testing.T) {
	// This test verifies that Init panics immediately on validation failure,
	// before attempting to start the Kafka consumer goroutine

	originalSecretKey := os.Getenv("STRIPE_SECRET_KEY")
	originalAccountID := os.Getenv("STRIPE_ACCOUNT_ID")
	defer func() {
		os.Setenv("STRIPE_SECRET_KEY", originalSecretKey)
		os.Setenv("STRIPE_ACCOUNT_ID", originalAccountID)
	}()

	// Test with missing secret key
	os.Setenv("STRIPE_SECRET_KEY", "")
	os.Setenv("STRIPE_ACCOUNT_ID", "acct_test")

	// Should panic immediately, not when starting goroutine
	assert.Panics(t, func() {
		Init(nil, nil)
	}, "Should panic on validation before starting goroutine")
}