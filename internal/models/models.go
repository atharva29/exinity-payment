package models

import "github.com/go-playground/validator/v10"

// a standard request structure for the transactions
type TransactionRequest struct {
}

// a standard response structure for the APIs
type APIResponse struct {
	StatusCode int         `json:"status_code" xml:"status_code"`
	Message    string      `json:"message" xml:"message"`
	Data       interface{} `json:"data,omitempty" xml:"data,omitempty"`
}

// DepositRequest represents a deposit request payload
type DepositRequest struct {
	Amount      string `json:"amount" validate:"required" example:"100"`          // Deposit amount
	UserID      string `json:"user_id" validate:"required" example:"1"`           // Unique user identifier
	Currency    string `json:"currency" validate:"required" example:"USD"`        // Currency code
	GatewayID   string `json:"gateway_id" validate:"required" example:"1"`        // Payment gateway ID
	GatewayName string `json:"gateway_name" validate:"required" example:"STRIPE"` // Payment gateway name
	CountryID   string `json:"country_id" validate:"required" example:"3"`        // Country code
}

// ValidateDepositRequest validates the DepositRequest struct
func ValidateDepositRequest(req DepositRequest) error {
	validate := validator.New()
	err := validate.Struct(req)
	if err != nil {
		return err
	}
	return nil
}

// Gateway struct for response
type Gateway struct {
	ID    string `json:"id" validate:"required" example:"1"`        // Unique gateway identifier
	Name  string `json:"name" validate:"required" example:"STRIPE"` // Name of the gateway
	Score int
}

// ValidateGateway validates the Gateway struct
func ValidateGateway(gateway Gateway) error {
	validate := validator.New()
	err := validate.Struct(gateway)
	if err != nil {
		return err
	}
	return nil
}
