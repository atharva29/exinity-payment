package db

import (
	"payment-gateway/db/db"
	"payment-gateway/db/redis"
)

type DB struct {
	DB    db.IDB
	Redis redis.IRedis
}

// NewDB creates a new DB instance
func NewDB() (*DB, error) {
	redisClient, err := redis.Init()
	if err != nil {
		return nil, err
	}
	// Initialize the database connection
	db, err := db.InitializeDB()
	if err != nil {
		return nil, err
	}
	return &DB{
		DB:    db,
		Redis: redisClient,
	}, nil
}
