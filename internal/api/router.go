package api

import (
	"net/http"
	"payment-gateway/db/db"
	"payment-gateway/db/redis"
	"payment-gateway/internal/services/psp"

	"github.com/gorilla/mux"
)

func SetupRouter(psp *psp.PSP, redisClient *redis.RedisClient, db *db.DB) *mux.Router {
	router := mux.NewRouter()
	router.Use(CORS)

	// deposit
	router.Handle("/deposit", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			DepositHandler(w, r, psp, redisClient) // Pass the psp instance here
		},
	)).Methods("POST", "OPTIONS")

	// withdrawal
	router.Handle("/withdrawal", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			WithdrawalHandler(w, r, psp, redisClient) // Pass the psp instance here
		},
	)).Methods("POST", "OPTIONS")

	// webhook
	router.Handle("/webhook", http.HandlerFunc(WebhookHandler)).Methods("POST", "OPTIONS")

	// get-gateway-by-country
	router.HandleFunc("/get-gateway-by-country/{countryID}", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			GetGatewayByCountryHandler(w, r, redisClient, db) // Pass the psp instance here
		},
	)).Methods("GET", "OPTIONS")

	// stripe webhook
	router.Handle("/webhook/stripe", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			PaymentCompleteHandler(w, r, psp, redisClient, db) // Pass the psp instance here
		},
	)).Methods("POST", "OPTIONS")

	return router
}
