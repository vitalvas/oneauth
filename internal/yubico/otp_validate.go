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

	matchModhex = regexp.MustCompile("^[cbdefghijklnrtuv]{44}$")

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

	keyIDHex := modhexToHex.Replace(keyID)

	dec, err := strconv.ParseInt(keyIDHex, 16, 64)
	if err != nil {
		return 0, ErrConvertModhexToHex
	}

	return dec, nil
}
