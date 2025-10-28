package db

import "payment-gateway/internal/models"

type IDB interface {
	CheckUserBalance(userID int, currency string, amount float64) (bool, float64, error)
	GetSupportedGatewaysByCountries(countryID string) ([]models.Gateway, error)
	CreateTransaction(transaction Transaction) error
}
