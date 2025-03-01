package api

import (
	"payment-gateway/internal/models"

	"github.com/go-playground/validator/v10"
)

// GatewayResponse struct for API response
type GatewayResponse struct {
	CountryID string           `json:"country_id" validate:"required" example:"US"` // 2-letter ISO country code
	Gateways  []models.Gateway `json:"gateways" validate:"dive,required"`                 // List of gateways sorted by score
}

// ValidateGatewayResponse validates the GatewayResponse struct
func ValidateGatewayResponse(resp GatewayResponse) error {
	validate := validator.New()
	err := validate.Struct(resp)
	if err != nil {
		return err
	}
	return nil
}
