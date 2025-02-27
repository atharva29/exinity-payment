package services

import (
	"encoding/base64"
	"fmt"
)

// masks data using base64 (feel free to enhance it and use stronger encryption algorithm)
func MaskData(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// unmasks data using base64
func UnmaskData(maskedData string) ([]byte, error) {
	decodedData, err := base64.StdEncoding.DecodeString(maskedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode data: %v", err)
	}
	return decodedData, nil
}
