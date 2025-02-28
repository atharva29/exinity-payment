package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"payment-gateway/db/redis"
	"payment-gateway/internal/services/psp"
)

// DepositHandler handles deposit requests.
func DepositHandler(w http.ResponseWriter, r *http.Request, psp psp.IPSP, redisClient *redis.RedisClient) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode request body
	reqBody, err := getDepositRequestFromQueryParams(r)
	if err != nil {
		log.Println("Error decoding request body:", err.Error())
		http.Error(w, "Bad Request: Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate request body
	if err := validateDepositRequest(reqBody); err != nil {
		log.Println("Error validating request body:", err.Error())
		http.Error(w, fmt.Sprintf("Bad Request: %s", err.Error()), http.StatusBadRequest)
		return
	}

	// Generate Order ID
	orderID, err := psp.Deposit(reqBody)
	if err != nil {
		log.Println("Error during deposit:", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	log.Println("generated OrderID ", orderID)

	// Store Data in Redis
	redisData := map[string]interface{}{
		"amount":     reqBody.Amount,
		"user_id":    reqBody.UserID.String(),
		"currency":   reqBody.Currency,
		"gateway_id": reqBody.GatewayID,
		"country_id": reqBody.CountryID,
		"order_id":   orderID,
	}

	err = redisClient.HSet("deposit:"+orderID, redisData)
	if err != nil {
		log.Println("Error storing data in redis:", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// get amount in paisa
	amountInPaisa, err := convertToPaisa(reqBody.Amount)
	if err != nil {
		log.Println("Error during amount conversion :", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Load HTML template from file
	tmpl, err := psp.GetTemplate() // Use Razorpay template path
	if err != nil {
		log.Println("Error parsing template:", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := psp.GetPaymentInfo(orderID, amountInPaisa, reqBody.Currency)
	// Set the Content-Type header
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	// Execute the template with the data
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println("Error executing template:", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// WithdrawalHandler handles withdrawal requests.
func WithdrawalHandler(w http.ResponseWriter, r *http.Request) {
	// withdrawal request logic
}

// WebhookHandler handles webhook events from Razorpay.
func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("webhook initiated")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading webhook body:", err.Error())
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Print the payload
	log.Println("Webhook Payload:")
	// log.Println(string(body))

	// Optionally, you can unmarshal the JSON payload to a struct
	var payload map[string]interface{}
	err = json.Unmarshal(body, &payload)
	if err != nil {
		log.Println("Error unmarshalling webhook payload:", err.Error())
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	// print the payload nicely
	formattedPayload, _ := json.MarshalIndent(payload, "", "  ")
	fmt.Println(string(formattedPayload))

	// TODO: Process the payload and update the order status accordingly.
	// Consider validating the signature for security.

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Webhook received successfully")
}
