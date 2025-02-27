package api

import (
	"fmt"
	"payment-gateway/internal/models"

	"github.com/google/uuid"
)

// validateDepositRequest validates the deposit request body.
func validateDepositRequest(req models.DepositRequest) error {
	if req.Amount == "" {
		return fmt.Errorf("amount is required")
	}
	if req.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required")
	}
	if req.Currency == "" {
		return fmt.Errorf("currency is required")
	}
	if req.GatewayID == 0 {
		return fmt.Errorf("gateway_id is required")
	}
	if req.CountryID == 0 {
		return fmt.Errorf("country_id is required")
	}
	// Add more validation rules here as needed (e.g., currency format, amount range)
	return nil
}
