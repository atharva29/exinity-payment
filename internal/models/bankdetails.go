package models

import "github.com/go-playground/validator/v10"

// BankAccountDetails contains the information needed to create a bank account
type BankAccountDetails struct {
	AccountHolderName string `json:"account_holder_name" validate:"required" example:"John Doe"`                            // Name of the account holder
	AccountNumber     string `json:"account_number" validate:"required" example:"1234567890"`                               // Bank account number
	RoutingNumber     string `json:"routing_number" validate:"required,len=9" example:"110000614"`                          // Bank routing number (ACH), typically 9 digits in the US
	Country           string `json:"country" validate:"required" example:"US"`                                        // Two-letter country code (e.g., "US")
	Currency          string `json:"currency" validate:"required,len=3" example:"usd"`                                      // Three-letter currency code (e.g., "usd")
	AccountHolderType string `json:"account_holder_type" validate:"required,oneof=individual company" example:"individual"` // "individual" or "company"
}

// CustomWithdrawalRequest represents a withdrawal request payload
type CustomWithdrawalRequest struct {
	Amount              int64              `json:"amount" validate:"required,gt=0" example:"5000"`                       // Amount in cents
	Currency            string             `json:"currency" validate:"required,len=3" example:"usd"`                     // 3-letter ISO code (e.g., "usd")
	Description         string             `json:"description" validate:"required" example:"Monthly payout"`             // Description of the payout
	BankDetails         BankAccountDetails `json:"bank_details" validate:"required"`                                     // Custom bank details
	Method              string             `json:"method" validate:"required,oneof=standard instant" example:"standard"` // "standard" or "instant" (default: "standard")
	StatementDescriptor string             `json:"statement_descriptor" validate:"max=22" example:"EXINITY PAYOUT"`      // Text on recipient's statement (max 22 chars)
	Metadata            map[string]string  `json:"metadata" example:"country_id:3,currency:USD,gateway_id:7"`            // Optional additional data
	UserID              string             `json:"user_id" validate:"required" example:"1"`                              // User ID making the withdrawal
	GatewayName         string             `json:"gateway_name" validate:"required" example:"DEFAULT_GATEWAY"`           // Name of the gateway
	GatewayID           string             `json:"gateway_id" validate:"required" example:"7"`                           // ID of the gateway
	CountryID           string             `json:"country_id" validate:"required" example:"US"`                    // 2-letter ISO country code
}

// ValidateCustomWithdrawalRequest validates the CustomWithdrawalRequest struct
func ValidateCustomWithdrawalRequest(req CustomWithdrawalRequest) error {
	validate := validator.New()
	err := validate.Struct(req)
	if err != nil {
		return err
	}
	return nil
}
