package models

import "github.com/google/uuid"

// a standard request structure for the transactions
type TransactionRequest struct {
}

// a standard response structure for the APIs
type APIResponse struct {
	StatusCode int         `json:"status_code" xml:"status_code"`
	Message    string      `json:"message" xml:"message"`
	Data       interface{} `json:"data,omitempty" xml:"data,omitempty"`
}

// DepositRequest struct for decoding the deposit request body.
type DepositRequest struct {
	Amount    string    `json:"amount"`
	UserID    uuid.UUID `json:"user_id"`
	Currency  string    `json:"currency"`
	GatewayID int       `json:"gateway_id"`
	CountryID int       `json:"country_id"`
}

// Gateway struct for response
type Gateway struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
