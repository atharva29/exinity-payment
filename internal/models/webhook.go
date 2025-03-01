package models

// Data represents additional data in the default gateway event
type Data struct {
	Metadata map[string]string `json:"metadata" example:"country_id:3,currency:USD,gateway_id:7,user_id:1"` // Optional metadata
}

// DefaultGatewayEvent represents the event structure for default gateway webhooks
type DefaultGatewayEvent struct {
	ID       string `json:"id" example:"fb848efc-2ea4-4de9-bece-d0e640ceb1ad"` // Unique event identifier
	Amount   int64  `json:"amount" example:"5000"`                             // Amount in cents
	Currency string `json:"currency" example:"usd"`                            // 3-letter ISO currency code
	Type     string `json:"type" example:"payment_intent.created"`             // Event type
	Data     Data   `json:"data"`                                              // Additional event data
}
