package stripe

import (
	"fmt"
	"payment-gateway/internal/models"

	str "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	stripe "github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/bankaccount"
	"github.com/stripe/stripe-go/v81/payout"
	"github.com/stripe/stripe-go/v81/token"
)

// Withdrawal handles the full process of creating a bank account and making a payout
func (s *StripeClient) Withdrawal(req models.CustomWithdrawalRequest) (string, error) {
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
	metadata["user_id"] = req.UserID.String()
	metadata["bank_account_holder"] = req.BankDetails.AccountHolderName

	params.Metadata = metadata

	// Create the payout
	p, err := payout.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create payout: %w", err)
	}

	return p.ID, nil
}

// createExternalBankAccount creates a bank account in Stripe from custom details
func (s *StripeClient) createExternalBankAccount(details models.BankAccountDetails) (string, error) {
	// Create token parameters
	tokenParams := &stripe.TokenParams{
		BankAccount: &stripe.BankAccountParams{
			Country:           stripe.String(details.Country),
			Currency:          stripe.String(details.Currency),
			AccountNumber:     stripe.String(details.AccountNumber),
			RoutingNumber:     stripe.String(details.RoutingNumber),
			AccountHolderName: stripe.String(details.AccountHolderName),
			AccountHolderType: stripe.String(details.AccountHolderType),
		},
	}

	// Create a token for the bank account
	token, err := token.New(tokenParams)
	if err != nil {
		return "", fmt.Errorf("failed to create token for bank account: %w", err)
	}

	customerParams := &str.CustomerParams{
		Email: stripe.String("user@example.com"),
		Name:  &details.AccountHolderName,
	}
	customer, err := customer.New(customerParams)
	if err != nil {
		return "", fmt.Errorf("failed to create customer: %w", err)
	}

	// Create bank account parameters
	bankParams := &stripe.BankAccountParams{
		Customer: stripe.String(customer.ID), // Your Stripe account ID
		Token:    stripe.String(token.ID),
	}

	// Create the bank account
	bankAccount, err := bankaccount.New(bankParams)
	if err != nil {
		return "", fmt.Errorf("failed to create bank account: %w", err)
	}

	return bankAccount.ID, nil
}
