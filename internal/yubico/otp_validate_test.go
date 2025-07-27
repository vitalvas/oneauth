package yubico

import (
	"testing"
)

func TestValidateOTP(t *testing.T) {
	otp := map[string]int64{
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

	for otp, expected := range otp {
		result, err := ValidateOTP(otp)
		if err != nil {
			t.Fatal(err)
		}

		if result != expected {
			t.Fatalf("expected %d, got %d", expected, result)
		}
	}
}

func TestValidateOTPFail(t *testing.T) {
	otp := map[string]error{
		"qwerty": ErrOTPHasInvalidLength,
		"cccccbhuinjdhkhclghtejrntflfbevvrvfvtkffghzz": ErrWrongOTPFormat,
	}

	for otp, expected := range otp {
		_, err := ValidateOTP(otp)
		if err != expected {
			t.Fatalf("expected %s, got %s", expected, err)
		}
	}
}

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

func TestExtractKeyID(t *testing.T) {
	tests := []struct {
		otp      string
		expected string
		hasError bool
	}{
		{"ccccccccccccdefghijklnrtuvcbdefghijklnrtuvic", "cccccccccccc", false}, // valid OTP
		{"cbdefghijklndefghijklnrtuvcbdefghijklnrtuvic", "cbdefghijkln", false}, // different key ID
		{"invalid", "", true}, // invalid OTP
		{"", "", true},        // empty OTP
		{"ccccccccccccdefghijklnrtuvcbdefghijklnrtuviz", "", true}, // invalid char at end
	}

	for _, test := range tests {
		keyID, err := ExtractKeyID(test.otp)
		if test.hasError {
			if err == nil {
				t.Errorf("ExtractKeyID(%q) should have returned an error", test.otp)
			}
		} else {
			if err != nil {
				t.Errorf("ExtractKeyID(%q) returned unexpected error: %v", test.otp, err)
			}
			if keyID != test.expected {
				t.Errorf("ExtractKeyID(%q) = %q, expected %q", test.otp, keyID, test.expected)
			}
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
