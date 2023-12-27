package tools

import (
	"crypto/rand"
	"math/big"
)

const nonceAllowedChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateNonce(length int) (string, error) {
	nonce := make([]byte, length)

	size := len(nonceAllowedChars)

	for i := range nonce {
		pos, err := rand.Int(rand.Reader, big.NewInt(int64(size)))
		if err != nil {
			return "", err
		}

		nonce[i] = nonceAllowedChars[pos.Int64()]
	}

	return string(nonce), nil
}
