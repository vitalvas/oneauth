package yubikey

import "testing"

func TestValidatePin(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1234567", true},
		{"12345678", true},
		{"12345", false},
		{"123456789", false},
		{"abcdefg", false},
		{"", false},
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
		{"12345678", true},
		{"1234567", false},
		{"123456789", false},
		{"abcdefg", false},
		{"", false},
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
