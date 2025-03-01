package models

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
	Amount      string `json:"amount"`
	UserID      string `json:"user_id"`
	Currency    string `json:"currency"`
	GatewayID   string `json:"gateway_id"`
	GatewayName string `json:"gateway_name"`
	CountryID   string `json:"country_id"`
}

// WithdrawalRequest struct for decoding the deposit request body.
type WithdrawalRequest struct {
	Amount              int64             `json:"amount"`               // Amount in cents
	Currency            string            `json:"currency"`             // 3-letter ISO code (e.g., "usd")
	Description         string            `json:"description"`          // Description of the payout
	Destination         string            `json:"destination"`          // Destination bank account or card ID
	Method              string            `json:"method"`               // "standard" or "instant" (default: "standard")
	StatementDescriptor string            `json:"statement_descriptor"` // Text that appears on recipient's statement
	Metadata            map[string]string `json:"metadata"`             // Optional: additional data
	GatewayID           string            `json:"gateway_id"`
	UserID              string            `json:"user_id"`
}

// Gateway struct for response
type Gateway struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Score int    `json:"score"`
}
