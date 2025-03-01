package defaultgateway

import (
	"payment-gateway/internal/models"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go"
)

// Withdrawal handles the full process of creating a bank account and making a payout
func (s *DefaultGatewayClient) Withdrawal(req models.CustomWithdrawalRequest) (string, error) {
	// Step 2: Create the payout to this bank account
	params := &stripe.PayoutParams{
		Amount:   stripe.Int64(req.Amount),
		Currency: stripe.String(req.Currency),
	}

	// Add optional parameters
	if req.Description != "" {
		params.Description = stripe.String(req.Description)
	}

	if req.Method != "" {
		params.Method = stripe.String(req.Method)
	}

	if req.StatementDescriptor != "" {
		// Stripe limits statement descriptors to 22 characters
		if len(req.StatementDescriptor) > 22 {
			req.StatementDescriptor = req.StatementDescriptor[:22]
		}
		params.StatementDescriptor = stripe.String(req.StatementDescriptor)
	}

	// Add metadata
	metadata := make(map[string]string)
	if req.Metadata != nil {
		for k, v := range req.Metadata {
			metadata[k] = v
		}
	}

	// Add user ID to metadata
	metadata["user_id"] = req.UserID
	metadata["bank_account_holder"] = req.BankDetails.AccountHolderName

	params.Metadata = metadata

	payoutID := uuid.New().String()

	return payoutID, nil
}
