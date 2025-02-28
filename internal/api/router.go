package api

import (
	"net/http"
	"os"
	"payment-gateway/db/db"
	"payment-gateway/db/redis"
	"payment-gateway/internal/services/psp"

	"github.com/gorilla/mux"
)

func SetupRouter(psp psp.IPSP, redisClient *redis.RedisClient, db *db.DB) *mux.Router {
	router := mux.NewRouter()

	// Create templates directory if it doesn't exist
	if _, err := os.Stat("templates"); os.IsNotExist(err) {
		os.Mkdir("templates", os.ModePerm)
	}

	// Ensure razorpay.html exists in templates
	if _, err := os.Stat("../templates/razorpay.html"); os.IsNotExist(err) {
		panic("razorpay.html not found in templates directory")
	}

	router.Handle("/deposit", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			DepositHandler(w, r, psp, redisClient) // Pass the psp instance here
		},
	)).Methods("GET")
	router.Handle("/withdrawal", http.HandlerFunc(WithdrawalHandler)).Methods("POST")
	router.Handle("/webhook", http.HandlerFunc(WebhookHandler)).Methods("POST")
	router.HandleFunc("/get-gateway-by-country/{countryID}", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			GetGatewayByCountryHandler(w, r, redisClient, db) // Pass the psp instance here
		},
	)).Methods("GET")

	return router
}
