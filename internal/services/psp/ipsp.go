package psp

import (
	"payment-gateway/db"
	"payment-gateway/internal/models"
)

type IPSP interface {
	Deposit(reqBody models.DepositRequest) (string, string, error)
	// GetPaymentInfo(orderID, amountInPaisa, currency string) interface{}
	Withdrawal(req models.CustomWithdrawalRequest) (string, error)
	GetName() string
	HandleWebhook(event any, db *db.DB) error
}
