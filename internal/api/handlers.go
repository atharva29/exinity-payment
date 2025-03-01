package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"payment-gateway/db"
	"payment-gateway/db/redis"
	"payment-gateway/internal/models"
	"payment-gateway/internal/services/psp"

	"github.com/gorilla/mux"
)

// DepositHandler handles deposit requests via POST request
func DepositHandler(w http.ResponseWriter, r *http.Request, psp *psp.PSP, redisClient *redis.RedisClient) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode request body
	var reqBody models.DepositRequest
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		log.Println("Error decoding request body:", err.Error())
		http.Error(w, "Bad Request: Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate request body
	if err := validateDepositRequest(reqBody, psp); err != nil {
		log.Println("Error validating request body:", err.Error())
		http.Error(w, fmt.Sprintf("Bad Request: %s", err.Error()), http.StatusBadRequest)
		return
	}

	// Generate Order ID
	p, _ := psp.Get(reqBody.GatewayName)
	orderID, client_secret, err := p.Deposit(reqBody)
	if err != nil {
		log.Println("Error during deposit:", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	log.Println("Generated OrderID:", orderID)

	// Store Data in Redis
	data := map[string]interface{}{
		"amount":        reqBody.Amount,
		"user_id":       reqBody.UserID,
		"currency":      reqBody.Currency,
		"gateway_id":    reqBody.GatewayID,
		"country_id":    reqBody.CountryID,
		"order_id":      orderID,
		"client_secret": client_secret,
		"status":        "created",
	}

	key := fmt.Sprintf("deposit:userid:%s:orderid:%s", reqBody.UserID, orderID)
	err = redisClient.HSet(key, data)
	if err != nil {
		log.Println("Error storing data in redis:", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
}

// WithdrawalHandler handles withdrawal requests.
func WithdrawalHandler(w http.ResponseWriter, r *http.Request, psp *psp.PSP, db *db.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode request body
	var reqBody models.CustomWithdrawalRequest
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		log.Println("Error decoding request body:", err.Error())
		http.Error(w, "Bad Request: Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate request body
	if err := validateWithdrawalRequest(reqBody, psp); err != nil {
		log.Println("Error validating request body:", err.Error())
		http.Error(w, fmt.Sprintf("Bad Request: %s", err.Error()), http.StatusBadRequest)
		return
	}

	// Generate Order ID
	p, err := psp.Get(reqBody.GatewayName)
	if err != nil {
		log.Println("Error invalid Gateway Name", reqBody.GatewayName)
		http.Error(w, fmt.Sprintf("Bad Request: %s", err.Error()), http.StatusNotFound)
		return
	}
	payoutID, err := p.Withdrawal(reqBody)
	if err != nil {
		log.Println("Error during deposit:", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	log.Println("Generated payoutID:", payoutID)

	// Store Data in Redis
	data := map[string]interface{}{
		"amount":     reqBody.Amount,
		"user_id":    reqBody.UserID,
		"currency":   reqBody.Currency,
		"gateway_id": reqBody.GatewayID,
		"country_id": reqBody.CountryID,
		"orderid":    payoutID,
		"status":     "created",
	}

	key := fmt.Sprintf("withdrawal:userid:%s:orderid:%s", reqBody.UserID, payoutID)
	err = db.Redis.HSet(key, data)
	if err != nil {
		log.Println("Error storing data in redis:", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"data": data})

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

// GetGatewayByCountryHandler
func GetGatewayByCountryHandler(w http.ResponseWriter, r *http.Request, db *db.DB) {
	vars := mux.Vars(r)
	countryID := vars["countryID"]

	if countryID == "" {
		http.Error(w, "countryID is required", http.StatusBadRequest)
		return
	}

	// Get gateways for the country from Redis Set
	gatewayIDs, err := db.Redis.GetGatewaysByCountry(r.Context(), countryID)
	if err != nil || len(gatewayIDs) == 0 {
		log.Println("Cache miss, fetching from DB")

		// Fetch from DB
		gatewayIDs, err = db.DB.GetSupportedGatewaysByCountries(countryID)
		if err != nil {
			http.Error(w, "Error fetching gateways", http.StatusInternalServerError)
			return
		}
		if len(gatewayIDs) == 0 {
			http.Error(w, "No gateways ID present for the country", http.StatusNotFound)
			return
		}

		// Store the gateways in Redis Set
		err = db.Redis.SaveGatewaysToRedisHashSet(r.Context(), countryID, gatewayIDs)
		if err != nil {
			http.Error(w, "Error redis insert gateways", http.StatusInternalServerError)
			return
		}
	}

	// Get scores for each gateway and sort by score
	sortedGateways, err := db.Redis.GetGatewaysByCountry(r.Context(), countryID)
	if err != nil {
		http.Error(w, "Error fetching gateway scores", http.StatusInternalServerError)
		return
	}

	// Respond with sorted gateways
	json.NewEncoder(w).Encode(GatewayResponse{CountryID: countryID, Gateways: sortedGateways})

}
