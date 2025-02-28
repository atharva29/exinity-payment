package razorpay

import (
	"errors"
	"fmt"
	"os"
	"payment-gateway/db/redis"
	"payment-gateway/internal/models"

	"github.com/razorpay/razorpay-go"
)

type RazoryPay struct {
	client *razorpay.Client
}

func Init() *RazoryPay {
	return &RazoryPay{
		client: razorpay.NewClient(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_KEY_SECRET")),
	}
}

func (p *RazoryPay) GetName() string {
	return "RAZORPAY"
}

func (p *RazoryPay) Deposit(req models.DepositRequest) (string, string, error) {
	data := map[string]interface{}{
		"amount":          req.Amount,
		"currency":        req.Currency,
		"receipt":         "some_receipt_id",
		"partial_payment": false,
		"notes":           map[string]interface{}{},
	}
	body, err := p.client.Order.Create(data, nil)
	if err != nil {
		return "", "", err
	}
	fmt.Println(body)
	razorId, _ := body["id"].(string)
	return razorId, "", nil

}

func (p *RazoryPay) Withdrawal(req models.WithdrawalRequest) (string, error) {
	data := map[string]interface{}{
		"amount":          req.Amount,
		"currency":        req.Currency,
		"receipt":         "some_receipt_id",
		"partial_payment": false,
		"notes":           map[string]interface{}{},
	}
	body, err := p.client.Order.Create(data, nil)
	if err != nil {
		return "", err
	}
	fmt.Println(body)
	razorId, _ := body["id"].(string)
	return razorId, nil

}

func (r *RazoryPay) GetPayment(paymentID string) error {
	// TODO: get payment details
	fmt.Println("get payment called", paymentID)
	return nil
}

func (r *RazoryPay) GetPaymentInfo(orderID, amountInPaisa, currency string) interface{} {
	// Prepare data for the template
	return PaymentPageData{ //Use Razorpay struct here.
		OrderID:      orderID,
		RazorpayKey:  os.Getenv("RAZORPAY_KEY_ID"), // Load from environment variable
		Amount:       amountInPaisa,                // convert to paise
		Currency:     currency,
		MerchantName: "Hiring At Exinity",                                                                                                                                                                                 // Replace with your merchant name
		ImageURL:     "https://media.licdn.com/dms/image/v2/C4D0BAQHKsboC6kuHfA/company-logo_200_200/company-logo_200_200/0/1630580504879/exinity_logo?e=2147483647&v=beta&t=Zc72r_d8x3O6u_ywVVT--7aE_K0wh0prA9cS2gAJCRs", // Replace with your logo URL
	}
}

func (r *RazoryPay) HandleWebhook(event any, redisClient *redis.RedisClient) error {
	return nil
}

func (r *RazoryPay) ExtractWebhookData(payload map[string]interface{}) (string, int64, string, error) {
	// Extract the necessary fields
	if payload["payload"] != nil {
		if paymentDowntime, ok := payload["payload"].(map[string]interface{})["payment.downtime"]; ok {
			if entity, ok := paymentDowntime.(map[string]interface{})["entity"]; ok {
				if entityMap, ok := entity.(map[string]interface{}); ok {
					status, ok1 := entityMap["status"].(string)
					updatedAt, ok2 := entityMap["updated_at"].(float64)
					id, ok3 := entityMap["id"].(string)

					if ok1 && ok2 && ok3 {
						fmt.Println("Extracted Data:")
						fmt.Println("Status:", status)
						fmt.Println("Updated At:", int64(updatedAt))
						fmt.Println("ID:", id)
						return status, int64(updatedAt), id, nil
					} else {
						return "", 0, "", errors.New("error extracting data from payload: Type assertion failed")
					}
				} else {
					return "", 0, "", errors.New("error extracting data from payload: entity is not map[string]interface{} ")
				}
			} else {
				return "", 0, "", errors.New("error extracting data from payload: payment.downtime does not contain entity")
			}
		} else {
			return "", 0, "", errors.New("error extracting data from payload: payload does not contain payment.downtime")
		}
	} else {
		return "", 0, "", errors.New("error extracting data from payload: payload not found")
	}
}
