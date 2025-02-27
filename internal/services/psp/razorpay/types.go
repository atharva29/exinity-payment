package razorpay

// PaymentPageData struct to hold the data passed to the HTML template.
type PaymentPageData struct {
	OrderID      string
	RazorpayKey  string
	Amount       string
	Currency     string
	MerchantName string
	ImageURL     string
}
