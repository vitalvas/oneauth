package yubikey

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestGeneratePinCode_ValidPIN(t *testing.T) {
	t.Run("generates valid PIN", func(t *testing.T) {
		pin, err := GeneratePinCode()
		assert.NoError(t, err)
		assert.Len(t, pin, 6)
		assert.True(t, ValidatePin(pin), "generated PIN should pass validation")
	})

	t.Run("uniqueness across multiple calls", func(t *testing.T) {
		pins := make(map[string]bool)
		for i := 0; i < 10; i++ {
			pin, err := GeneratePinCode()
			assert.NoError(t, err)
			pins[pin] = true
		}
		// With 10 random 6-digit PINs, we should get at least 2 unique values
		assert.True(t, len(pins) >= 2, "should generate diverse PINs")
	})
}

func TestGeneratePukCode_ValidPUK(t *testing.T) {
	t.Run("generates valid PUK", func(t *testing.T) {
		puk, err := GeneratePukCode()
		assert.NoError(t, err)
		assert.Len(t, puk, 8)
		assert.True(t, ValidatePuk(puk), "generated PUK should pass validation")
	})

	t.Run("uniqueness across multiple calls", func(t *testing.T) {
		puks := make(map[string]bool)
		for i := 0; i < 10; i++ {
			puk, err := GeneratePukCode()
			assert.NoError(t, err)
			puks[puk] = true
		}
		assert.True(t, len(puks) >= 2, "should generate diverse PUKs")
	})
}

func TestGenerateManagementKey_Properties(t *testing.T) {
	t.Run("key length is 24 bytes", func(t *testing.T) {
		key, err := GenerateManagementKey()
		assert.NoError(t, err)
		assert.Len(t, key, 24)
	})

	t.Run("uniqueness across multiple calls", func(t *testing.T) {
		key1, err := GenerateManagementKey()
		assert.NoError(t, err)
		key2, err := GenerateManagementKey()
		assert.NoError(t, err)
		assert.NotEqual(t, key1, key2, "two keys should not be identical")
	})
}
