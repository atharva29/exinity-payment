package db

import (
	"fmt"
	"payment-gateway/internal/models"
)

// GetSupportedGatewaysByCountries fetches all gateways for a given country
func (db *DB) GetSupportedGatewaysByCountries(countryID string) ([]models.Gateway, error) {
	query := `
		SELECT g.id AS gateway_id, g.name AS gateway_name
		FROM gateways g
		JOIN gateway_countries gc ON g.id = gc.gateway_id
		WHERE gc.country_id = $1
		ORDER BY g.name
	`

	rows, err := db.db.Query(query, countryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch gateways for country %s: %v", countryID, err)
	}
	defer rows.Close()

	var gateways []models.Gateway
	for rows.Next() {
		var gateway models.Gateway
		if err := rows.Scan(&gateway.ID, &gateway.Name); err != nil {
			return nil, fmt.Errorf("failed to scan gateway: %v", err)
		}
		gateways = append(gateways, gateway)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %v", err)
	}

	return gateways, nil
}
