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

	router.Handle("/deposit", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			DepositHandler(w, r, psp, redisClient) // Pass the psp instance here
		},
	)).Methods("POST", "OPTIONS")

	router.Handle("/withdrawal", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			WithdrawalHandler(w, r, psp, redisClient) // Pass the psp instance here
		},
	)).Methods("POST", "OPTIONS")

	router.Handle("/webhook", http.HandlerFunc(WebhookHandler)).Methods("POST", "OPTIONS")
	router.HandleFunc("/get-gateway-by-country/{countryID}", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			GetGatewayByCountryHandler(w, r, redisClient, db) // Pass the psp instance here
		},
	)).Methods("GET", "OPTIONS")

	router.Handle("/payment-complete", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			PaymentCompleteHandler(w, r, psp, redisClient, db) // Pass the psp instance here
		},
	)).Methods("POST", "OPTIONS")

	return router
}
