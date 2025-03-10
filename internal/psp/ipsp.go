package psp

import (
	"payment-gateway/db"
	"payment-gateway/internal/models"
)

type IPSP interface {
	Deposit(reqBody models.DepositRequest) (string, string, error)
	// GetPaymentInfo(orderID, amountInPaisa, currency string) interface{}
	Withdrawal(req models.CustomWithdrawalRequest, db *db.DB) (string, error)
	GetName() string
	HandleWebhook(event any, db *db.DB) error
	PublishWebhookToKafka(ev any) error
}
