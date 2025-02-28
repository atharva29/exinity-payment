package api

import "payment-gateway/internal/models"

// GatewayResponse struct for API response
type GatewayResponse struct {
	CountryID string           `json:"country_id"`
	Gateways  []models.Gateway `json:"gateways"`
}
