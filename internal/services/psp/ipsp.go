package psp

import (
	"payment-gateway/db/redis"
	"payment-gateway/internal/models"
)

type IPSP interface {
	Deposit(reqBody models.DepositRequest) (string, string, error)
	// GetPaymentInfo(orderID, amountInPaisa, currency string) interface{}
	Withdrawal(req models.WithdrawalRequest) (string, error)
	GetName() string
	HandleWebhook(event any, redisClient *redis.RedisClient) error
}
