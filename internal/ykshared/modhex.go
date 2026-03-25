package ykshared

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	matchModhexChar = regexp.MustCompile("^[cbdefghijklnrtuv]+$")

	modhexToHex = strings.NewReplacer(
		"l", "a",
		"n", "b",
		"r", "c",
		"t", "d",
		"u", "e",
		"v", "f",
		"c", "0",
		"b", "1",
		"d", "2",
		"e", "3",
		"f", "4",
		"g", "5",
		"h", "6",
		"i", "7",
		"j", "8",
		"k", "9",
	)
)

// IsValidModhex checks if a string contains only valid modhex characters
func IsValidModhex(s string) bool {
	if s == "" {
		return false
	}
	return matchModhexChar.MatchString(s)
}

// ValidateModhexString validates a modhex string of any length
func ValidateModhexString(s string) error {
	if s == "" {
		return ErrInvalidModhexLength
	}

	if !matchModhexChar.MatchString(s) {
		return ErrInvalidModhexChar
	}

	return nil
}

// ValidateModhexWithLength validates a modhex string with a specific length
func ValidateModhexWithLength(s string, expectedLength int) error {
	if len(s) != expectedLength {
		return ErrInvalidModhexLength
	}

	if !matchModhexChar.MatchString(s) {
		return ErrInvalidModhexChar
	}

	return nil
}

// ValidateKeyIDFormat validates a YubiKey key ID format (12 modhex characters)
func ValidateKeyIDFormat(keyID string) error {
	return ValidateModhexWithLength(keyID, 12)
}

// ModhexToHex converts a modhex string to hexadecimal
func ModhexToHex(modhex string) (string, error) {
	if err := ValidateModhexString(modhex); err != nil {
		return "", err
	}

	return modhexToHex.Replace(modhex), nil
}

// ModhexToInt converts a modhex string to integer
func ModhexToInt(modhex string) (int64, error) {
	hex, err := ModhexToHex(modhex)
	if err != nil {
		return 0, err
	}

	dec, err := strconv.ParseInt(hex, 16, 64)
	if err != nil {
		return 0, ErrConvertModhexToHex
	}

	return dec, nil
}

// HexToModhex converts a hexadecimal string to modhex
// This is the reverse operation of ModhexToHex
func HexToModhex(hex string) (string, error) {
	// Validate hex string
	if len(hex)%2 != 0 {
		return "", errors.New("hex string must have even length")
	}

	for _, c := range hex {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return "", errors.New("invalid hex character")
		}
	}

	// Create reverse mapping of modhexToHex
	hexToModhex := strings.NewReplacer(
		"0", "c", "1", "b", "2", "d", "3", "e",
		"4", "f", "5", "g", "6", "h", "7", "i",
		"8", "j", "9", "k", "a", "l", "b", "n",
		"c", "r", "d", "t", "e", "u", "f", "v",
	)

	return hexToModhex.Replace(strings.ToLower(hex)), nil
}

// BytesToModhex converts bytes to modhex string
func BytesToModhex(data []byte) string {
	hexStr := hex.EncodeToString(data)
	modhex, err := HexToModhex(hexStr)
	if err != nil {
		// This should not happen with valid byte data, but handle it gracefully
		panic(fmt.Sprintf("BytesToModhex: unexpected error converting hex to modhex: %v", err))
	}
	return modhex
}

// ModhexToBytes converts modhex string to bytes
func ModhexToBytes(modhex string) ([]byte, error) {
	hexStr, err := ModhexToHex(modhex)
	if err != nil {
		return nil, err
	}
	return hex.DecodeString(hexStr)
}

// GenerateRandomModhex generates a random modhex string of specified length
func GenerateRandomModhex(length int) (string, error) {
	if length%2 != 0 {
		return "", fmt.Errorf("length must be even")
	}

	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return BytesToModhex(bytes), nil
}

// GenerateKeyID generates a random 12-character modhex key ID
func GenerateKeyID() (string, error) {
	return GenerateRandomModhex(12)
}

// ValidateOTP validates a YubiKey OTP and returns the key ID as integer
// docs: https://developers.yubico.com/OTP/OTPs_Explained.html
func ValidateOTP(otp string) (int64, error) {
	if len(otp) != 44 {
		return 0, ErrOTPHasInvalidLength
	}

	if !IsValidModhex(otp) {
		return 0, ErrWrongOTPFormat
	}

	keyID := otp[:12]

	// Use the ModhexToInt function
	dec, err := ModhexToInt(keyID)
	if err != nil {
		return 0, err
	}

	return dec, nil
}

// ExtractKeyID extracts and validates the key ID from a YubiKey OTP
func ExtractKeyID(otp string) (string, error) {
	// Validate the OTP format first
	_, err := ValidateOTP(otp)
	if err != nil {
		return "", err
	}

	// Extract key ID (first 12 characters)
	return otp[:12], nil
}
