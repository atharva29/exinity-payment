package defaultgateway

import (
	"fmt"
	"payment-gateway/db"
	"payment-gateway/internal/models"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go"
)

// Withdrawal handles the full process of creating a bank account and making a payout
func (s *DefaultGatewayClient) Withdrawal(req models.CustomWithdrawalRequest, db *db.DB) (string, error) {
	userID, err := s.parseUserID(req.UserID)
	if err != nil {
		return "", err
	}

	amount, err := s.convertAmount(req.Amount)
	if err != nil {
		return "", err
	}

	if err := s.checkBalance(db, userID, req.Currency, amount); err != nil {
		return "", err
	}

	// payoutParams := s.buildPayoutParams(req)
	payoutID := uuid.New().String() // Placeholder; replace with actual payout creation if needed

	return payoutID, nil
}

// parseUserID converts the user ID string to an integer
func (s *DefaultGatewayClient) parseUserID(userID string) (int, error) {
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return 0, fmt.Errorf("invalid user_id format: %v", err)
	}
	return userIDInt, nil
}

// convertAmount converts Stripe amount (in cents) to float64
func (s *DefaultGatewayClient) convertAmount(amount int64) (float64, error) {
	// Assuming amount is in cents; adjust if different
	if amount < 0 {
		return 0, fmt.Errorf("amount cannot be negative: %d", amount)
	}
	return float64(amount) / 100, nil
}

// checkBalance verifies if the user has sufficient funds
func (s *DefaultGatewayClient) checkBalance(db *db.DB, userID int, currency string, amount float64) error {
	hasEnough, currentBalance, err := db.DB.CheckUserBalance(userID, strings.ToLower(currency), amount)
	if err != nil {
		return err
	}
	if !hasEnough {
		return fmt.Errorf("insufficient balance: current balance %.2f %s, requested %.2f %s",
			currentBalance, currency, amount, currency)
	}
	return nil
}

// buildPayoutParams constructs the Stripe payout parameters
func (s *DefaultGatewayClient) buildPayoutParams(req models.CustomWithdrawalRequest) *stripe.PayoutParams {
	params := &stripe.PayoutParams{
		Amount:   stripe.Int64(req.Amount),
		Currency: stripe.String(strings.ToLower(req.Currency)),
	}

	// Set optional parameters
	s.setOptionalPayoutParams(params, req)
	s.addPayoutMetadata(params, req)

	return params
}

// setOptionalPayoutParams adds optional fields to payout parameters
func (s *DefaultGatewayClient) setOptionalPayoutParams(params *stripe.PayoutParams, req models.CustomWithdrawalRequest) {
	if req.Description != "" {
		params.Description = stripe.String(req.Description)
	}

	if req.Method != "" {
		params.Method = stripe.String(req.Method)
	}

	if req.StatementDescriptor != "" {
		statementDescriptor := req.StatementDescriptor
		if len(statementDescriptor) > 22 {
			statementDescriptor = statementDescriptor[:22]
		}
		params.StatementDescriptor = stripe.String(statementDescriptor)
	}
}

// addPayoutMetadata adds metadata to payout parameters
func (s *DefaultGatewayClient) addPayoutMetadata(params *stripe.PayoutParams, req models.CustomWithdrawalRequest) {
	metadata := make(map[string]string)

	// Copy existing metadata
	if req.Metadata != nil {
		for k, v := range req.Metadata {
			metadata[k] = v
		}
	}

	// Add required metadata
	metadata["user_id"] = req.UserID
	metadata["bank_account_holder"] = req.BankDetails.AccountHolderName
	metadata["gateway_id"] = req.GatewayID
	metadata["country_id"] = req.CountryID

	params.Metadata = metadata
}
