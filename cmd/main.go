package main

import (
	"log"
	"net/http"
	"payment-gateway/db"
	"payment-gateway/internal/api"
	"payment-gateway/internal/kafka"
	"payment-gateway/internal/psp"
	"payment-gateway/internal/psp/defaultgateway"
	"payment-gateway/internal/psp/razorpay"
	"payment-gateway/internal/psp/stripe"

	_ "payment-gateway/docs" // Import generated docs

	"github.com/joho/godotenv"
)

// @title Deposit API
// @version 1.0
// @description This is a deposit processing API
// @host localhost:8080
// @BasePath /
func main() {

	// Load environment variables from .env file
	err := godotenv.Load("config")
	if err != nil {
		log.Println("Error loading .env file, using defaults or system environment variables")
	}

	kafka := kafka.NewWriterPool()

	db, err := db.NewDB()
	if err != nil {
		log.Fatalf("error loading DB : %v", err.Error())
	}

	psp := psp.Init([]psp.IPSP{razorpay.Init(), stripe.Init(kafka, db), defaultgateway.Init()})

	// // Set up the HTTP server and routes
	router := api.SetupRouter(psp, db)

	// // Start the server on port 8080
	log.Println("Starting server on port 8080...")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}

}
