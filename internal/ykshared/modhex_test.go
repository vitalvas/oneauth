package ykshared

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
)

func TestIsValidModhex(t *testing.T) {
	tests := map[string]bool{
		"cccccccccccc":        true,  // valid modhex
		"cbdefghijklnrtuv":    true,  // all valid modhex chars
		"":                    false, // empty string
		"ccccccccccccz":       false, // contains invalid char 'z'
		"abcdef":              false, // contains invalid chars
		"CCCCCCCCCCCC":        false, // uppercase not allowed
		"cbdefghijklnrtuvxyz": false, // contains invalid chars at end
	}

	for input, expected := range tests {
		result := IsValidModhex(input)
		if result != expected {
			t.Errorf("IsValidModhex(%q) = %v, expected %v", input, result, expected)
		}
	}
}

func TestValidateModhexString(t *testing.T) {
	tests := map[string]error{
		"cccccccccccc":     nil,                    // valid modhex
		"cbdefghijklnrtuv": nil,                    // all valid modhex chars
		"":                 ErrInvalidModhexLength, // empty string
		"ccccccccccccz":    ErrInvalidModhexChar,   // contains invalid char
		"abcdef":           ErrInvalidModhexChar,   // contains invalid chars
		"CCCCCCCCCCCC":     ErrInvalidModhexChar,   // uppercase not allowed
	}

	for input, expected := range tests {
		err := ValidateModhexString(input)
		if err != expected {
			t.Errorf("ValidateModhexString(%q) = %v, expected %v", input, err, expected)
		}
	}
}

func TestValidateModhexWithLength(t *testing.T) {
	tests := []struct {
		input    string
		length   int
		expected error
	}{
		{"cccccccccccc", 12, nil},                     // valid 12-char modhex
		{"cccccccccccc", 10, ErrInvalidModhexLength},  // wrong length
		{"ccccccccccc", 12, ErrInvalidModhexLength},   // too short
		{"ccccccccccccc", 12, ErrInvalidModhexLength}, // too long
		{"cccccccccccc", 12, nil},                     // valid
		{"ccccccccccgz", 12, ErrInvalidModhexChar},    // invalid chars
		{"", 5, ErrInvalidModhexLength},               // empty string wrong length
	}

	for _, test := range tests {
		err := ValidateModhexWithLength(test.input, test.length)
		if err != test.expected {
			t.Errorf("ValidateModhexWithLength(%q, %d) = %v, expected %v",
				test.input, test.length, err, test.expected)
		}
	}
}

func TestValidateKeyIDFormat(t *testing.T) {
	tests := map[string]error{
		"cccccccccccc":  nil,                    // valid key ID
		"cbdefghijkln":  nil,                    // valid key ID with different chars
		"ccccccccccc":   ErrInvalidModhexLength, // too short
		"ccccccccccccc": ErrInvalidModhexLength, // too long
		"ccccccccccgz":  ErrInvalidModhexChar,   // invalid chars
		"CCCCCCCCCCCC":  ErrInvalidModhexChar,   // uppercase not allowed
		"":              ErrInvalidModhexLength, // empty string
		"abcdefghijkl":  ErrInvalidModhexChar,   // invalid chars
	}

	for input, expected := range tests {
		err := ValidateKeyIDFormat(input)
		if err != expected {
			t.Errorf("ValidateKeyIDFormat(%q) = %v, expected %v", input, err, expected)
		}
	}
}

func TestModhexToHex(t *testing.T) {
	tests := map[string]struct {
		expected string
		hasError bool
	}{
		"cccccccccccc": {"000000000000", false}, // all 'c' = all '0'
		"cbdefghijkln": {"0123456789ab", false}, // various chars
		"vvvvvvvvvvvv": {"ffffffffffff", false}, // all 'v' = all 'f'
		"":             {"", true},              // empty string
		"ccccccccccgz": {"", true},              // invalid chars
		"CCCCCCCCCCCC": {"", true},              // uppercase
	}

	for input, expected := range tests {
		result, err := ModhexToHex(input)
		if expected.hasError {
			if err == nil {
				t.Errorf("ModhexToHex(%q) should have returned an error", input)
			}
		} else {
			if err != nil {
				t.Errorf("ModhexToHex(%q) returned unexpected error: %v", input, err)
			}
			if result != expected.expected {
				t.Errorf("ModhexToHex(%q) = %q, expected %q", input, result, expected.expected)
			}
		}
	}
}

func TestModhexToInt(t *testing.T) {
	tests := map[string]struct {
		expected int64
		hasError bool
	}{
		"cccccccccccc": {0, false},             // all 'c' = 0
		"cccccccccccb": {1, false},             // 'b' = '1' = 1
		"cccccccccccn": {11, false},            // 'n' = 'b' = 11
		"cbdefghijkln": {0x123456789ab, false}, // hex conversion
		"":             {0, true},              // empty string
		"ccccccccccgz": {0, true},              // invalid chars
	}

	for input, expected := range tests {
		result, err := ModhexToInt(input)
		if expected.hasError {
			if err == nil {
				t.Errorf("ModhexToInt(%q) should have returned an error", input)
			}
		} else {
			if err != nil {
				t.Errorf("ModhexToInt(%q) returned unexpected error: %v", input, err)
			}
			if result != expected.expected {
				t.Errorf("ModhexToInt(%q) = %d, expected %d", input, result, expected.expected)
			}
		}
	}
}

func TestHexToModhex(t *testing.T) {
	tests := map[string]struct {
		expected string
		hasError bool
	}{
		"000000000000": {"cccccccccccc", false}, // all '0' = all 'c'
		"0123456789ab": {"cbdefghijkln", false}, // various chars
		"ffffffffffff": {"vvvvvvvvvvvv", false}, // all 'f' = all 'v'
		"ABCDEF":       {"lnrtuv", false},       // uppercase hex
		"":             {"", false},             // empty string
		"012":          {"", true},              // odd length
		"xyz":          {"", true},              // invalid hex chars
		"0g":           {"", true},              // invalid hex character
	}

	for input, expected := range tests {
		result, err := HexToModhex(input)
		if expected.hasError {
			if err == nil {
				t.Errorf("HexToModhex(%q) should have returned an error", input)
			}
		} else {
			if err != nil {
				t.Errorf("HexToModhex(%q) returned unexpected error: %v", input, err)
			}
			if result != expected.expected {
				t.Errorf("HexToModhex(%q) = %q, expected %q", input, result, expected.expected)
			}
		}
	}
}

func TestBytesToModhex(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{"zero bytes", []byte{0x00, 0x00, 0x00}, "cccccc"},
		{"known bytes", []byte{0x01, 0x23, 0x45}, "cbdefg"}, // 01->cb, 23->de, 45->fg
		{"max bytes", []byte{0xFF, 0xFF}, "vvvv"},
		{"empty", []byte{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BytesToModhex(tt.input)
			if result != tt.expected {
				t.Errorf("BytesToModhex(%v) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestModhexToBytes(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    []byte
		expectError bool
	}{
		{"valid modhex", "cccccc", []byte{0x00, 0x00, 0x00}, false},
		{"known conversion", "cbdefg", []byte{0x01, 0x23, 0x45}, false},
		{"max values", "vvvv", []byte{0xFF, 0xFF}, false},
		{"empty", "", nil, true}, // Empty string should fail validation
		{"invalid character", "ccccccz", nil, true},
		{"odd length from failing validation", "X", nil, true}, // Invalid character should fail yubico validation
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ModhexToBytes(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("ModhexToBytes(%q) should have returned an error", tt.input)
				}
				if result != nil {
					t.Errorf("ModhexToBytes(%q) should have returned nil result on error", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ModhexToBytes(%q) returned unexpected error: %v", tt.input, err)
				}
				if len(result) != len(tt.expected) {
					t.Errorf("ModhexToBytes(%q) result length = %d, expected %d", tt.input, len(result), len(tt.expected))
				}
				for i, b := range tt.expected {
					if result[i] != b {
						t.Errorf("ModhexToBytes(%q)[%d] = %02x, expected %02x", tt.input, i, result[i], b)
					}
				}
			}
		})
	}
}

func TestGenerateRandomModhex(t *testing.T) {
	t.Run("valid lengths", func(t *testing.T) {
		lengths := []int{2, 4, 6, 12, 32, 44}
		for _, length := range lengths {
			t.Run(fmt.Sprintf("length_%d", length), func(t *testing.T) {
				result, err := GenerateRandomModhex(length)
				if err != nil {
					t.Errorf("GenerateRandomModhex(%d) returned error: %v", length, err)
				}
				if len(result) != length {
					t.Errorf("GenerateRandomModhex(%d) length = %d, expected %d", length, len(result), length)
				}
				if !IsValidModhex(result) {
					t.Errorf("GenerateRandomModhex(%d) produced invalid modhex: %q", length, result)
				}
			})
		}
	})

	t.Run("odd length should fail", func(t *testing.T) {
		_, err := GenerateRandomModhex(3)
		if err == nil {
			t.Error("GenerateRandomModhex(3) should have returned an error")
		}
		if !strings.Contains(err.Error(), "length must be even") {
			t.Errorf("GenerateRandomModhex(3) error should contain 'length must be even', got: %v", err)
		}
	})

	t.Run("randomness", func(t *testing.T) {
		// Generate multiple random strings and ensure they're different
		results := make([]string, 10)
		for i := 0; i < 10; i++ {
			result, err := GenerateRandomModhex(12)
			if err != nil {
				t.Errorf("GenerateRandomModhex(12) returned error: %v", err)
			}
			results[i] = result
		}

		// Check that we have some variation (not all the same)
		unique := make(map[string]bool)
		for _, result := range results {
			unique[result] = true
		}
		if len(unique) <= 1 {
			t.Error("GenerateRandomModhex should produce different results")
		}
	})
}

func TestGenerateKeyID(t *testing.T) {
	t.Run("valid key ID generation", func(t *testing.T) {
		keyID, err := GenerateKeyID()
		if err != nil {
			t.Errorf("GenerateKeyID() returned error: %v", err)
		}
		if len(keyID) != 12 {
			t.Errorf("GenerateKeyID() length = %d, expected 12", len(keyID))
		}
		if !IsValidModhex(keyID) {
			t.Errorf("GenerateKeyID() produced invalid modhex: %q", keyID)
		}
	})

	t.Run("multiple generations are different", func(t *testing.T) {
		keyIDs := make([]string, 5)
		for i := 0; i < 5; i++ {
			keyID, err := GenerateKeyID()
			if err != nil {
				t.Errorf("GenerateKeyID() returned error: %v", err)
			}
			keyIDs[i] = keyID
		}

		// Check uniqueness
		unique := make(map[string]bool)
		for _, keyID := range keyIDs {
			unique[keyID] = true
		}
		if len(unique) <= 1 {
			t.Error("GenerateKeyID should produce different results")
		}
	})
}

func TestRoundTripConversion(t *testing.T) {
	t.Run("hex to modhex to bytes", func(t *testing.T) {
		original := "0123456789abcdef"

		// Convert to modhex
		modhex, err := HexToModhex(original)
		if err != nil {
			t.Errorf("HexToModhex(%q) returned error: %v", original, err)
		}
		if !IsValidModhex(modhex) {
			t.Errorf("HexToModhex(%q) produced invalid modhex: %q", original, modhex)
		}

		// Convert back to bytes
		bytes, err := ModhexToBytes(modhex)
		if err != nil {
			t.Errorf("ModhexToBytes(%q) returned error: %v", modhex, err)
		}

		// Convert bytes back to hex
		result := hex.EncodeToString(bytes)
		if result != original {
			t.Errorf("Round trip conversion failed: original=%q, result=%q", original, result)
		}
	})

	t.Run("bytes to modhex to bytes", func(t *testing.T) {
		original := []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF}

		// Convert to modhex
		modhex := BytesToModhex(original)
		if !IsValidModhex(modhex) {
			t.Errorf("BytesToModhex(%v) produced invalid modhex: %q", original, modhex)
		}

		// Convert back to bytes
		result, err := ModhexToBytes(modhex)
		if err != nil {
			t.Errorf("ModhexToBytes(%q) returned error: %v", modhex, err)
		}
		if len(result) != len(original) {
			t.Errorf("Round trip conversion failed: original length=%d, result length=%d", len(original), len(result))
		}
		for i, b := range original {
			if result[i] != b {
				t.Errorf("Round trip conversion failed at index %d: original=%02x, result=%02x", i, b, result[i])
			}
		}
	})
}

func TestValidateOTP(t *testing.T) {
	tests := map[string]int64{
		// owned keys
		"cccccbhuinjdhkhclghtejrntflfbevvrvfvtkffghkj": 24017794, // 5c nano
		"cccccbhuinjdtgrgjfnelevvfjhteujcdigiicvujvcl": 24017794, // 5c nano

		"ccccccnlrctbvtjgucihjthunectghivervfrnnikvtr": 12239057, // 5 nfc
		"ccccccnlrctbfindlgneifuiteefgggtltlufeccrujt": 12239057, // 5 nfc

		"cccccbjudbfivcjjnrghdhetftgdrnkgeikhcfurrcdv": 26091847, // 5c
		"cccccbjudbfigtctgrheiiivvdgutieecvhtbunuvhhr": 26091847, // 5c

		// unowned keys
		"ccccccccltncdjjifceergtnukivgiujhgehgnkrfcef": 44464,
		"ccccccbchvthlivuitriujjifivbvtrjkjfirllluurj": 1077206,
	}

	for otp, expected := range tests {
		t.Run("otp_"+otp[:12], func(t *testing.T) {
			result, err := ValidateOTP(otp)
			if err != nil {
				t.Errorf("ValidateOTP(%q) returned error: %v", otp, err)
			}
			if result != expected {
				t.Errorf("ValidateOTP(%q) = %d, expected %d", otp, result, expected)
			}
		})
	}
}

func TestValidateOTPErrors(t *testing.T) {
	tests := map[string]error{
		"qwerty": ErrOTPHasInvalidLength,
		"cccccbhuinjdhkhclghtejrntflfbevvrvfvtkffghzz": ErrWrongOTPFormat,
	}

	for otp, expected := range tests {
		t.Run("invalid_otp", func(t *testing.T) {
			_, err := ValidateOTP(otp)
			if err != expected {
				t.Errorf("ValidateOTP(%q) error = %v, expected %v", otp, err, expected)
			}
		})
	}
}

func TestExtractKeyID(t *testing.T) {
	tests := []struct {
		name      string
		otp       string
		expected  string
		shouldErr bool
	}{
		{
			name:     "valid OTP",
			otp:      "ccccccccccccdefghijklnrtuvcbdefghijklnrtuvic",
			expected: "cccccccccccc",
		},
		{
			name:     "different key ID",
			otp:      "cbdefghijklndefghijklnrtuvcbdefghijklnrtuvic",
			expected: "cbdefghijkln",
		},
		{
			name:      "invalid OTP",
			otp:       "invalid",
			shouldErr: true,
		},
		{
			name:      "empty OTP",
			otp:       "",
			shouldErr: true,
		},
		{
			name:      "invalid char at end",
			otp:       "ccccccccccccdefghijklnrtuvcbdefghijklnrtuviz",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyID, err := ExtractKeyID(tt.otp)
			if tt.shouldErr {
				if err == nil {
					t.Errorf("ExtractKeyID(%q) should have returned an error", tt.otp)
				}
			} else {
				if err != nil {
					t.Errorf("ExtractKeyID(%q) returned unexpected error: %v", tt.otp, err)
				}
				if keyID != tt.expected {
					t.Errorf("ExtractKeyID(%q) = %q, expected %q", tt.otp, keyID, tt.expected)
				}
			}
		})
	}
}
