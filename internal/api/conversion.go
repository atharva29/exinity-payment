package api

import (
	"fmt"
	"strconv"
)

func convertToPaisa(amount string) (string, error) {
	amountFloat, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return "", fmt.Errorf("invalid amount format: %w", err)
	}
	amountInPaisa := int64(amountFloat * 100) // Assuming 1 unit = 100 paise
	return strconv.FormatInt(amountInPaisa, 10), nil
}
