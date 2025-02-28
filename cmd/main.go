package main

import (
	"log"
	"net/http"
	"os"
	"payment-gateway/db"
	"payment-gateway/db/redis"
	"payment-gateway/internal/api"
	"payment-gateway/internal/services/psp/razorpay"

	"github.com/joho/godotenv"
)

func main() {

	// Load environment variables from .env file
	err := godotenv.Load("config")
	if err != nil {
		log.Println("Error loading .env file, using defaults or system environment variables")
	}

	psp := razorpay.Init()
	redisAddr := os.Getenv("REDIS_ADDR")         // e.g., "localhost:6379"
	redisPassword := os.Getenv("REDIS_PASSWORD") // Leave blank if no password
	redisDB := 0
	redisClient, err := redis.Init(redisAddr, redisPassword, redisDB)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// // Set up the HTTP server and routes

	// Initialize the database connection
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	dbURL := "postgres://" + dbUser + ":" + dbPassword + "@" + dbHost + ":" + dbPort + "/" + dbName + "?sslmode=disable"

	db.InitializeDB(dbURL)

	// // Set up the HTTP server and routes
	router := api.SetupRouter(psp, redisClient)

	// // Start the server on port 8080
	log.Println("Starting server on port 8080...")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}

}
