package yubikey

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GeneratePinCode() (string, error) {
	newPINInt, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", newPINInt), nil
}

func GeneratePukCode() (string, error) {
	newPUKInt, err := rand.Int(rand.Reader, big.NewInt(100_000_000))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%08d", newPUKInt), nil
}

func GenerateManagementKey() ([24]byte, error) {
	var newKey [24]byte
	if _, err := rand.Read(newKey[:]); err != nil {
		return newKey, err
	}

	return newKey, nil
}
