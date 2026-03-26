package yubikey

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePin(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1234567", false},
		{"12345678", false},
		{"12345", false},
		{"123456789", false},
		{"abcdefg", false},
		{"", false},
		{"5413380", true},
		{"45165603", true},
		{"4739347", true},
		{"111111", false},
		{"11111111", false},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := ValidatePin(test.input)
			if result != test.expected {
				t.Errorf("For input '%s', expected %v, but got %v", test.input, test.expected, result)
			}
		})
	}
}

func TestValidatePuk(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"12345678", false},
		{"1234567", false},
		{"123456789", false},
		{"abcdefg", false},
		{"", false},
		{"5413380", false},
		{"45165603", true},
		{"4739347", false},
		{"21297607", true},
		{"43298881", true},
		{"111111", false},
		{"11111111", false},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := ValidatePuk(test.input)
			if result != test.expected {
				t.Errorf("For input '%s', expected %v, but got %v", test.input, test.expected, result)
			}
		})
	}
}

func TestValidatePin_AllBlockedPINs(t *testing.T) {
	blockedPins := []string{
		"123456", "123123", "654321", "123321",
		"112233", "121212", "123456789", "12345678",
		"1234567", "520520", "123654", "1234567890",
		"159753",
	}
	for _, pin := range blockedPins {
		t.Run("blocked_"+pin, func(t *testing.T) {
			assert.False(t, ValidatePin(pin), "PIN %s should be blocked", pin)
		})
	}
}

func TestValidatePuk_AllBlockedPUKs(t *testing.T) {
	blockedPuks := []string{
		"12345678", "11111111",
	}
	for _, puk := range blockedPuks {
		t.Run("blocked_"+puk, func(t *testing.T) {
			assert.False(t, ValidatePuk(puk), "PUK %s should be blocked", puk)
		})
	}
}

func TestValidatePin_SingleCharacterPIN(t *testing.T) {
	// All same digits should fail (unique chars <= 1)
	assert.False(t, ValidatePin("000000"))
	assert.False(t, ValidatePin("999999"))
	assert.False(t, ValidatePin("5555555"))
	assert.False(t, ValidatePin("88888888"))
}

func TestValidatePuk_SingleCharacterPUK(t *testing.T) {
	assert.False(t, ValidatePuk("00000000"))
	assert.False(t, ValidatePuk("99999999"))
}

func TestValidatePin_SpecialCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"letters only", "abcdef", false},
		{"mixed alphanumeric", "123abc", false},
		{"too short", "12345", false},
		{"too long 9 digits", "123456789", false},
		{"empty string", "", false},
		{"spaces", "      ", false},
		{"valid 7 digit", "4739347", true},
		{"valid 8 digit", "45165603", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ValidatePin(tt.input))
		})
	}
}

func TestValidatePuk_SpecialCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"letters only", "abcdefgh", false},
		{"7 digits", "1234567", false},
		{"9 digits", "123456789", false},
		{"empty", "", false},
		{"spaces", "        ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ValidatePuk(tt.input))
		})
	}
}
