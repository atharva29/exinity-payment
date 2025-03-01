package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"payment-gateway/internal/services"
	"time"

	_ "github.com/lib/pq"
)

type DB struct {
	db *sql.DB
}

// InitializeDB initializes the database connection
func InitializeDB() (*DB, error) {
	var db *sql.DB
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	dbURL := "postgres://" + dbUser + ":" + dbPassword + "@" + dbHost + ":" + dbPort + "/" + dbName + "?sslmode=disable"

	var err error

	err = services.RetryOperation(func() error {
		db, err = sql.Open("postgres", dbURL)
		if err != nil {
			return err
		}

		return db.Ping()
	}, 5)

	if err != nil {
		return nil, fmt.Errorf("could not connect to the database: %v", err.Error())
	}

	log.Println("Successfully connected to the database.")
	return &DB{db}, nil
}

func (d *DB) CreateUser(user User) error {
	query := `INSERT INTO users (username, email, country_id, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5) RETURNING id`

	err := d.db.QueryRow(query, user.Username, user.Email, user.CountryID, time.Now(), time.Now()).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("failed to insert user: %v", err)
	}
	return nil
}

func (d *DB) GetUsers() ([]User, error) {
	rows, err := d.db.Query(`SELECT id, username, email, country_id, created_at, updated_at FROM users`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %v", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.CountryID, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %v", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (d *DB) CreateGateway(gateway Gateway) error {
	query := `INSERT INTO gateways (name, data_format_supported, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4) RETURNING id`

	err := d.db.QueryRow(query, gateway.Name, gateway.DataFormatSupported, time.Now(), time.Now()).Scan(&gateway.ID)
	if err != nil {
		return fmt.Errorf("failed to insert gateway: %v", err)
	}
	return nil
}

func (d *DB) GetGateways() ([]Gateway, error) {
	rows, err := d.db.Query(`SELECT id, name, data_format_supported, created_at, updated_at FROM gateways`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch gateways: %v", err)
	}
	defer rows.Close()

	var gateways []Gateway
	for rows.Next() {
		var gateway Gateway
		if err := rows.Scan(&gateway.ID, &gateway.Name, &gateway.DataFormatSupported, &gateway.CreatedAt, &gateway.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan gateway: %v", err)
		}
		gateways = append(gateways, gateway)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return gateways, nil
}

func (d *DB) CreateCountry(country Country) error {
	query := `INSERT INTO countries (name, code, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4) RETURNING id`

	err := d.db.QueryRow(query, country.Name, country.Code, time.Now(), time.Now()).Scan(&country.ID)
	if err != nil {
		return fmt.Errorf("failed to insert country: %v", err)
	}
	return nil
}

func (d *DB) GetCountries() ([]Country, error) {
	rows, err := d.db.Query(`SELECT id, name, code, created_at, updated_at FROM countries`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch countries: %v", err)
	}
	defer rows.Close()

	var countries []Country
	for rows.Next() {
		var country Country
		if err := rows.Scan(&country.ID, &country.Name, &country.Code, &country.CreatedAt, &country.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan country: %v", err)
		}
		countries = append(countries, country)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return countries, nil
}

func (d *DB) CreateTransaction(transaction Transaction) error {
	// Start a transaction
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	// Defer a rollback in case anything fails
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	// Insert into transactions table
	queryTransaction := `
		INSERT INTO transactions (
			order_id, amount, type, status, gateway_id, 
			country_id, user_id, created_at, currency
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
		RETURNING id`

	err = tx.QueryRow(queryTransaction,
		transaction.OrderID,
		transaction.Amount,
		transaction.Type,
		transaction.Status,
		transaction.GatewayID,
		transaction.CountryID,
		transaction.UserID,
		time.Now(),
		transaction.Currency,
	).Scan(&transaction.ID)
	if err != nil {
		return fmt.Errorf("failed to insert transaction: %v", err)
	}

	// Update or insert into ledger table
	// Using UPSERT (INSERT ... ON CONFLICT) to handle both new and existing ledger entries
	queryLedger := `
		INSERT INTO ledger (user_id, currency, amount, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, currency)
		DO UPDATE SET 
			amount = ledger.amount + EXCLUDED.amount,
			updated_at = EXCLUDED.updated_at
		RETURNING id`

	var ledgerID int
	err = tx.QueryRow(queryLedger,
		transaction.UserID,
		transaction.Currency,
		transaction.Amount,
		time.Now(),
	).Scan(&ledgerID)
	if err != nil {
		return fmt.Errorf("failed to update ledger: %v", err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func (d *DB) GetTransactions() ([]Transaction, error) {
	rows, err := d.db.Query(`SELECT id, amount, type, status, user_id, gateway_id, country_id, created_at FROM transactions`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %v", err)
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var transaction Transaction
		if err := rows.Scan(&transaction.ID, &transaction.Amount, &transaction.Type, &transaction.Status, &transaction.UserID, &transaction.GatewayID, &transaction.CountryID, &transaction.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %v", err)
		}
		transactions = append(transactions, transaction)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return transactions, nil
}

func (d *DB) GetSupportedCountriesByGateway(gatewayID int) ([]Country, error) {
	query := `
		SELECT c.id AS country_id, c.name AS country_name
		FROM countries c
		JOIN gateway_countries gc ON c.id = gc.country_id
		WHERE gc.gateway_id = $1
		ORDER BY c.name
	`

	rows, err := d.db.Query(query, gatewayID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch countries for gateway %d: %v", gatewayID, err)
	}
	defer rows.Close()

	var countries []Country
	for rows.Next() {
		var country Country
		if err := rows.Scan(&country.ID, &country.Name); err != nil {
			return nil, fmt.Errorf("failed to scan country: %v", err)
		}
		countries = append(countries, country)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %v", err)
	}

	return countries, nil
}

// CheckUserBalance checks if the user has sufficient balance for a withdrawal
func (d *DB) CheckUserBalance(userID int, currency string, amount float64) (bool, float64, error) {
	var currentBalance float64

	query := `
		SELECT amount 
		FROM ledger 
		WHERE user_id = $1 AND currency = $2`

	err := d.db.QueryRow(query, userID, currency).Scan(&currentBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			// No ledger entry exists for this user/currency
			return false, 0, fmt.Errorf("no balance found for user %d in currency %s", userID, currency)
		}
		return false, 0, fmt.Errorf("failed to check balance: %v", err)
	}

	// Check if balance is sufficient
	hasEnough := currentBalance >= amount
	return hasEnough, currentBalance, nil
}
