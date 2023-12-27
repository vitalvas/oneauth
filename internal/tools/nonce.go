package tools

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateNonce(length int) (string, error) {
	byteSize := (length*3 + 3) / 4

	randomBytes := make([]byte, byteSize)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	nonce := base64.RawURLEncoding.EncodeToString(randomBytes)[:length]

	return nonce, nil
}
