package pkg

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// GenerateRandomString returns a random alphanumeric string of given length
func GenerateRandomString(length int) (string, error) {
	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}
	return string(result), nil
}

// GeneratePumBookID returns an ID
func GeneratePumBookID(eventID string) (string, error) {
	randomStr, err := GenerateRandomString(5)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("PB-%s%s", eventID, randomStr), nil
}
