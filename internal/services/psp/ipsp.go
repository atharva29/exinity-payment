package psp

import (
	"payment-gateway/internal/models"
)

type IPSP interface {
	Deposit(reqBody models.DepositRequest) (string, error)
	GetPaymentInfo(orderID, amountInPaisa, currency string) interface{}
	Withdrawal(req models.WithdrawalRequest) (string, error)
	GetName() string
}
