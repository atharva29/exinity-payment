package api

import (
	"fmt"
	"net/http"
	"payment-gateway/internal/models"
	"strconv"

	"github.com/google/uuid"
)

// getDepositRequestFromQueryParams extracts and validates the deposit request data from query parameters.
func getDepositRequestFromQueryParams(r *http.Request) (models.DepositRequest, error) {
	queryParams := r.URL.Query()

	amount := queryParams.Get("amount")
	if amount == "" {
		return models.DepositRequest{}, fmt.Errorf("amount is required")
	}

	userIDStr := queryParams.Get("user_id")
	if userIDStr == "" {
		return models.DepositRequest{}, fmt.Errorf("user_id is required")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return models.DepositRequest{}, fmt.Errorf("invalid user_id format")
	}

	currency := queryParams.Get("currency")
	if currency == "" {
		return models.DepositRequest{}, fmt.Errorf("currency is required")
	}

	gatewayIDStr := queryParams.Get("gateway_id")
	if gatewayIDStr == "" {
		return models.DepositRequest{}, fmt.Errorf("gateway_id is required")
	}
	gatewayID, err := strconv.Atoi(gatewayIDStr)
	if err != nil {
		return models.DepositRequest{}, fmt.Errorf("invalid gateway_id format")
	}

	countryIDStr := queryParams.Get("country_id")
	if countryIDStr == "" {
		return models.DepositRequest{}, fmt.Errorf("country_id is required")
	}
	countryID, err := strconv.Atoi(countryIDStr)
	if err != nil {
		return models.DepositRequest{}, fmt.Errorf("invalid country_id format")
	}

	return models.DepositRequest{
		Amount:    amount,
		UserID:    userID,
		Currency:  currency,
		GatewayID: gatewayID,
		CountryID: countryID,
	}, nil
}
