package yubico

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

// docs: https://developers.yubico.com/OTP/OTPs_Explained.html

var (
	ErrOTPHasInvalidLength = errors.New("otp has invalid length")
	ErrWrongOTPFormat      = errors.New("wrong otp format")
	ErrConvertModhexToHex  = errors.New("failed to convert modhex to hex")
	ErrInvalidModhexChar   = errors.New("invalid modhex character")
	ErrInvalidModhexLength = errors.New("invalid modhex length")

	matchModhex     = regexp.MustCompile("^[cbdefghijklnrtuv]{44}$")
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

func ValidateOTP(otp string) (int64, error) {
	if len(otp) != 44 {
		return 0, ErrOTPHasInvalidLength
	}

	if !matchModhex.MatchString(otp) {
		return 0, ErrWrongOTPFormat
	}

	keyID := otp[:12]

	// Use the public ModhexToInt function
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
