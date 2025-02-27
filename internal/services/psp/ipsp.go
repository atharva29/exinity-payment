package psp

import (
	"payment-gateway/internal/models"
	"text/template"
)

type IPSP interface {
	Deposit(reqBody models.DepositRequest) (string, error)
	GetPaymentInfo(orderID, amountInPaisa, currency string) interface{}
	GetTemplate() (*template.Template, error)
}
