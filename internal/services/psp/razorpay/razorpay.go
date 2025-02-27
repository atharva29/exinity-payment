package razorpay

import (
	"fmt"
	"log"
	"os"
	"payment-gateway/internal/models"
	"text/template"

	"github.com/razorpay/razorpay-go"
)

type RazoryPay struct {
	client       *razorpay.Client
	templatePath string
}

func Init() *RazoryPay {
	return &RazoryPay{
		client:       razorpay.NewClient(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_KEY_SECRET")),
		templatePath: "../templates/razorpay.html",
	}
}

func (p *RazoryPay) Deposit(req models.DepositRequest) (string, error) {
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
	razorId, _ := body["id"].(string)
	return razorId, nil

}

func (r *RazoryPay) GetPayment(paymentID string) error {
	// TODO: get payment details
	fmt.Println("get payment called", paymentID)
	return nil
}

func (r *RazoryPay) GetTemplate() (*template.Template, error) {
	tmpl, err := template.ParseFiles(r.templatePath) // Use Razorpay template path
	if err != nil {
		log.Println("Error parsing template:", err.Error())
		return nil, err
	}
	return tmpl, nil
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
