package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func RandomHexStringKey(length int) (string, error) {
	key := make([]byte, int(length / 2))
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(key), nil
}
