package tools

import "math/rand"

const nonceAllowedChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateNonce(length int) (string, error) {
	nonce := make([]byte, length)

	for i := range nonce {
		nonce[i] = nonceAllowedChars[rand.Intn(len(nonceAllowedChars))]
	}

	return string(nonce), nil
}
