package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"payment-gateway/db"
	"payment-gateway/db/redis"
	"payment-gateway/internal/models"
	"payment-gateway/internal/psp"

	"github.com/gorilla/mux"
)

// DepositHandler handles deposit requests via POST request
// @Summary Process a new deposit request
// @Description Handles deposit creation with payment gateway integration and stores result in Redis
// @Tags deposits
// @Accept json
// @Produce json
// @Param deposit body models.DepositRequest true "Deposit request payload"
// @Success 200 {object} map[string]interface{} "Deposit created successfully"
// @Failure 400 {object} map[string]string "Invalid request payload"
// @Failure 405 {object} map[string]string "Method not allowed"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /deposit [post]
func DepositHandler(w http.ResponseWriter, r *http.Request, psp *psp.PSP, redisClient redis.IRedis) {
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
	if err := models.ValidateDepositRequest(reqBody); err != nil {
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
	err = redisClient.HSet(r.Context(), key, data)
	if err != nil {
		log.Println("Error storing data in redis:", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
}

// WithdrawalHandler handles withdrawal requests.
// @Summary Process a new withdrawal request
// @Description Handles withdrawal creation with payment gateway integration and stores result in Redis
// @Tags withdrawals
// @Accept json
// @Produce json
// @Param withdrawal body models.CustomWithdrawalRequest true "Withdrawal request payload"
// @Success 200 {object} map[string]interface{} "Withdrawal created successfully"
// @Failure 400 {object} map[string]string "Invalid request payload"
// @Failure 404 {object} map[string]string "Invalid gateway name"
// @Failure 405 {object} map[string]string "Method not allowed"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /withdrawal [post]
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
	if err := models.ValidateCustomWithdrawalRequest(reqBody); err != nil {
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
	payoutID, err := p.Withdrawal(reqBody, db)
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
	err = db.Redis.HSet(r.Context(), key, data)
	if err != nil {
		log.Println("Error storing data in redis:", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"data": data})

}

// GetGatewayByCountryHandler retrieves supported gateways for a given country.
// @Summary Get payment gateways by country
// @Description Fetches a list of supported payment gateway IDs for a specified country from Redis or DB, sorted by score
// @Tags gateways
// @Accept json
// @Produce json
// @Param countryID path string true "Country ID" example:"3"
// @Success 200 {object} GatewayResponse "List of gateway IDs for the country"
// @Failure 400 {object} map[string]string "Bad Request - Missing countryID"
// @Failure 404 {object} map[string]string "Not Found - No gateways for the country"
// @Failure 500 {object} map[string]string "Internal Server Error - Database or Redis failure"
// @Router /gateways/{countryID} [get]
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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GatewayResponse{CountryID: countryID, Gateways: sortedGateways})
}
