package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"payment-gateway/db"
	"payment-gateway/internal/models"
	"payment-gateway/internal/psp"

	"github.com/stripe/stripe-go/webhook"
)

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
		log.Printf("⚠️  Invalid module: %v\n", err)
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
		log.Printf("⚠️  Invalid module: %v\n", err)
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
