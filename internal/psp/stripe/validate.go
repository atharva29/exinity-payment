package stripe

import (
	"fmt"
	"log"
	"strconv"
)

// validateMetadata validates and converts Stripe payment intent metadata to integers
func validateMetadata(metadata map[string]string) (map[string]int, error) {
	tm := make(map[string]int)
	// Validate and convert gateway_id
	gatewayID, err := strconv.Atoi(metadata["gateway_id"])
	if err != nil {
		log.Printf("❌ Error converting gateway_id to int: %v", err)
		return tm, fmt.Errorf("invalid gateway_id format: %v", err)
	}
	tm["gateway_id"] = gatewayID

	// Validate and convert country_id
	countryID, err := strconv.Atoi(metadata["country_id"])
	if err != nil {
		log.Printf("❌ Error converting country_id to int: %v", err)
		return tm, fmt.Errorf("invalid country_id format: %v", err)
	}
	tm["country_id"] = countryID

	// Validate and convert user_id
	userID, err := strconv.Atoi(metadata["user_id"])
	if err != nil {
		log.Printf("❌ Error converting user_id to int: %v", err)
		return tm, fmt.Errorf("invalid user_id format: %v", err)
	}
	tm["user_id"] = userID

	return tm, nil
}
