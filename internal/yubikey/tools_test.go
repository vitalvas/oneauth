package yubikey

import (
	"testing"
)

func TestGeneratePinCode(t *testing.T) {
	t.Run("GeneratedPINShouldBeSixDigits", func(t *testing.T) {
		pin, err := GeneratePinCode()
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
		if len(pin) != 6 {
			t.Errorf("Expected a 6-digit PIN, but got %s", pin)
		}
	})

	t.Run("GeneratedPINShouldBeNumeric", func(t *testing.T) {
		pin, err := GeneratePinCode()
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
		for _, digit := range pin {
			if digit < '0' || digit > '9' {
				t.Errorf("Expected a numeric PIN, but got %s", pin)
				break
			}
		}
	})
}

func TestGeneratePukCode(t *testing.T) {
	t.Run("GeneratedPUKShouldBeEightDigits", func(t *testing.T) {
		puk, err := GeneratePukCode()
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
		if len(puk) != 8 {
			t.Errorf("Expected an 8-digit PUK, but got %s", puk)
		}
	})

	t.Run("GeneratedPUKShouldBeNumeric", func(t *testing.T) {
		puk, err := GeneratePukCode()
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
		for _, digit := range puk {
			if digit < '0' || digit > '9' {
				t.Errorf("Expected a numeric PUK, but got %s", puk)
				break
			}
		}
	})
}

func TestGenerateManagementKey(t *testing.T) {
	key, err := GenerateManagementKey()
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	if len(key) != 24 {
		t.Errorf("Expected key length of 24 bytes, but got %d", len(key))
	}

	var zero int
	for _, b := range key {
		if b == 0 {
			zero++
		}
	}

	if zero == 24 {
		t.Errorf("Expected key to be non-zero, but got %v", key)
	}
}
