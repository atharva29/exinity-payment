package main

import (
	"log"
	"net/http"
	"payment-gateway/db/db"
	"payment-gateway/db/redis"
	"payment-gateway/internal/api"
	"payment-gateway/internal/services/psp"
	"payment-gateway/internal/services/psp/razorpay"

	"github.com/joho/godotenv"
)

func main() {

	// Load environment variables from .env file
	err := godotenv.Load("config")
	if err != nil {
		log.Println("Error loading .env file, using defaults or system environment variables")
	}

	psp := psp.Init([]psp.IPSP{razorpay.Init()})

	redisClient, err := redis.Init()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Initialize the database connection
	db := db.InitializeDB()

	// // Set up the HTTP server and routes
	router := api.SetupRouter(psp, redisClient, db)

	// // Start the server on port 8080
	log.Println("Starting server on port 8080...")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}

}
