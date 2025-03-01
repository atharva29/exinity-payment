package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"payment-gateway/db"
	"payment-gateway/internal/models"
	"payment-gateway/internal/psp"

	"github.com/stripe/stripe-go/webhook"
)

// StripeWebhookHandler handles webhook events from Stripe.
// @Summary Handle Stripe webhook events
// @Description Processes incoming webhook events from Stripe, verifies the signature, and delegates to the PSP service layer
// @Tags webhooks
// @Accept json
// @Produce plain
// @Param payload body object true "Stripe webhook payload (dynamic JSON structure)"
// @Param Stripe-Signature header string true "Stripe signature for verification" example:"t=123456789,v1=abc123..."
// @Success 200 {string} string "Webhook processed successfully"
// @Failure 400 {object} map[string]string "Bad Request - Payload too large"
// @Failure 401 {object} map[string]string "Unauthorized - Invalid signature"
// @Failure 500 {object} map[string]string "Internal Server Error - Processing or module error"
// @Router /webhook/stripe [post]
func StripeWebhookHandler(w http.ResponseWriter, r *http.Request, psp *psp.PSP, db *db.DB) {
	const MaxBodyBytes = int64(65536) // Limit request size
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
		return
	}

	// Verify Stripe signature
	event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), os.Getenv("STRIPE_WEBHOOK_SECRET"))
	if err != nil {
		log.Printf("⚠️  Webhook signature verification failed: %v\n", err)
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	p, err := psp.Get("STRIPE")
	if err != nil {
		log.Printf("⚠️  Invalid gateway: %v\n", err)
		http.Error(w, "Invalid error", http.StatusInternalServerError)
		return
	}
	// Process event in the service layer
	err = p.HandleWebhook(&event, db)
	if err != nil {
		log.Printf("❌ Error handling event: %v", err)
		http.Error(w, "Error processing event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DefaultGatewayWebhookHandler handles webhook events from the default gateway.
// @Summary Handle default gateway webhook events
// @Description Processes incoming webhook events from the default gateway, parses the payload, and delegates to the PSP service layer
// @Tags webhooks
// @Accept json
// @Produce plain
// @Param payload body models.DefaultGatewayEvent true "Default gateway webhook payload"
// @Success 200 {string} string "Webhook processed successfully"
// @Failure 400 {object} map[string]string "Bad Request - Payload too large"
// @Failure 500 {object} map[string]string "Internal Server Error - Parsing or processing error"
// @Router /webhook/default-gateway  [post]
func DefaultGatewayWebhookHandler(w http.ResponseWriter, r *http.Request, psp *psp.PSP, db *db.DB) {
	const MaxBodyBytes = int64(65536) // Limit request size
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
		return
	}

	// Verify Stripe signature
	// TODO verify signature for default gateway

	p, err := psp.Get("DEFAULT_GATEWAY")
	if err != nil {
		log.Printf("⚠️  Invalid gateway: %v\n", err)
		http.Error(w, "Invalid error", http.StatusInternalServerError)
		return
	}

	event := &models.DefaultGatewayEvent{}
	err = json.Unmarshal(payload, event)
	if err != nil {
		log.Printf("❌ Error parsin event: %v", err)
		http.Error(w, "Error parsin event", http.StatusInternalServerError)
		return
	}

	// Process event in the service layer
	err = p.HandleWebhook(event, db)
	if err != nil {
		log.Printf("❌ Error handling event: %v", err)
		http.Error(w, "Error processing event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Webhook received successfully")
}
