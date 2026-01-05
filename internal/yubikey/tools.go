package yubikey

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GeneratePinCode() (string, error) {
	for i := 0; i < 10; i++ {
		newPINInt, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
		if err != nil {
			return "", err
		}

		pin := fmt.Sprintf("%06d", newPINInt)

		if ValidatePin(pin) {
			return pin, nil
		}
	}

	return "", fmt.Errorf("could not generate a valid PIN code")
}

func GeneratePukCode() (string, error) {
	for i := 0; i < 10; i++ {
		newPUKInt, err := rand.Int(rand.Reader, big.NewInt(100_000_000))
		if err != nil {
			return "", err
		}

		puk := fmt.Sprintf("%08d", newPUKInt)
		if ValidatePuk(puk) {
			return puk, nil
		}
	}

	return "", fmt.Errorf("could not generate a valid PUK code")
}

func GenerateManagementKey() ([]byte, error) {
	newKey := make([]byte, 24)
	if _, err := rand.Read(newKey); err != nil {
		return nil, err
	}

	return newKey, nil
}
