package yubikey

import "testing"

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
