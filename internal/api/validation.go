package api

import (
	"fmt"
	"payment-gateway/internal/models"
	"payment-gateway/internal/psp"
)

// validateDepositRequest validates the deposit request body.
func validateDepositRequest(req models.DepositRequest, psp *psp.PSP) error {
	if req.Amount == "" {
		return fmt.Errorf("amount is required")
	}
	if req.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if req.Currency == "" {
		return fmt.Errorf("currency is required")
	}
	if req.GatewayID == "" {
		return fmt.Errorf("gateway_id is required")
	}
	if req.GatewayName == "" {
		return fmt.Errorf("gateway_name is required")
	}
	if req.CountryID == "" {
		return fmt.Errorf("country_id is required")
	}
	if _, err := psp.Get(req.GatewayName); err != nil {
		return err
	}
	// Add more validation rules here as needed (e.g., currency format, amount range)
	return nil
}

// validateDepositRequest validates the deposit request body.
func validateWithdrawalRequest(req models.CustomWithdrawalRequest, psp *psp.PSP) error {
	// Add more validation rules here as needed (e.g., currency format, amount range)
	return nil
}
