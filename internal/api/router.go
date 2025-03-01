package api

import (
	"net/http"
	"payment-gateway/db"
	"payment-gateway/internal/psp"

	"github.com/gorilla/mux"
)

func SetupRouter(psp *psp.PSP, db *db.DB) *mux.Router {
	router := mux.NewRouter()
	router.Use(CORS)

	// get-gateway-by-country
	router.HandleFunc("/get-gateway-by-country/{countryID}", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			GetGatewayByCountryHandler(w, r, db) // Pass the psp instance here
		},
	)).Methods("GET", "OPTIONS")

	// deposit
	router.Handle("/deposit", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			DepositHandler(w, r, psp, db.Redis) // Pass the psp instance here
		},
	)).Methods("POST", "OPTIONS")

	// withdrawal
	router.Handle("/withdrawal", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			WithdrawalHandler(w, r, psp, db) // Pass the psp instance here
		},
	)).Methods("POST", "OPTIONS")

	// webhook
	router.Handle("/webhook", http.HandlerFunc(WebhookHandler)).Methods("POST", "OPTIONS")

	// stripe webhook
	router.Handle("/webhook/stripe", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			StripeWebhookHandler(w, r, psp, db) // Pass the psp instance here
		},
	)).Methods("POST", "OPTIONS")

	// stripe webhook
	router.Handle("/webhook/defaulte-gateway", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			StripeWebhookHandler(w, r, psp, db) // Pass the psp instance here
		},
	)).Methods("POST", "OPTIONS")

	return router
}
