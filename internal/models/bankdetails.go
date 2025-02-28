package models

import "github.com/google/uuid"

// BankAccountDetails contains the information needed to create a bank account
type BankAccountDetails struct {
	AccountHolderName string `json:"account_holder_name"` // Name of the account holder
	AccountNumber     string `json:"account_number"`      // Bank account number
	RoutingNumber     string `json:"routing_number"`      // Bank routing number (ACH)
	Country           string `json:"country"`             // Two-letter country code (e.g., "US")
	Currency          string `json:"currency"`            // Three-letter currency code (e.g., "usd")
	AccountHolderType string `json:"account_holder_type"` // "individual" or "company"
}

// WithdrawalRequest struct for creating a custom bank withdrawal
type CustomWithdrawalRequest struct {
	Amount              int64              `json:"amount"`               // Amount in cents
	Currency            string             `json:"currency"`             // 3-letter ISO code (e.g., "usd")
	Description         string             `json:"description"`          // Description of the payout
	BankDetails         BankAccountDetails `json:"bank_details"`         // Custom bank details
	Method              string             `json:"method"`               // "standard" or "instant" (default: "standard")
	StatementDescriptor string             `json:"statement_descriptor"` // Text on recipient's statement
	Metadata            map[string]string  `json:"metadata"`             // Optional additional data
	UserID              uuid.UUID          `json:"user_id"`              // User ID making the withdrawal
	GatewayName         string             `json:"gateway_name"`         // Name of the gateway
	GatewayID           string             `json:"gateway_id"`           // ID of the gateway
	CountryID           string             `json:"country_id"`           // ID of the country
}
