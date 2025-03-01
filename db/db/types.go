package db

import "time"

type User struct {
	ID        int
	Username  string
	Email     string
	CountryID int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Gateway struct {
	ID                  int
	Name                string
	DataFormatSupported string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type Country struct {
	ID        int
	Name      string
	Code      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Transaction struct {
	ID        int
	Amount    float64
	Type      string
	Status    string
	UserID    int
	GatewayID int
	CountryID int
	CreatedAt time.Time
}
