package models

type DefaultGatewayEvent struct {
	ID       string `json:"id"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Type     string `json:"type"`
	Data     Data   `json:"data"`
}

type Data struct {
	Metadata map[string]string `json:"metadata"`
}
