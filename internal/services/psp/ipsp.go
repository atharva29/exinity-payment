package psp

import (
	"payment-gateway/internal/models"

	"github.com/stripe/stripe-go"
)

type IPSP interface {
	Deposit(reqBody models.DepositRequest) (string, string, error)
	// GetPaymentInfo(orderID, amountInPaisa, currency string) interface{}
	Withdrawal(req models.WithdrawalRequest) (string, error)
	GetName() string
	HandleWebhook(event stripe.Event) error
}
