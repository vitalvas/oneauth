package ykshared

import "errors"

// Shared error definitions for YubiKey operations
var (
	ErrInvalidModhexChar   = errors.New("invalid modhex character")
	ErrInvalidModhexLength = errors.New("invalid modhex length")
	ErrConvertModhexToHex  = errors.New("failed to convert modhex to hex")

	// OTP validation errors
	ErrOTPHasInvalidLength = errors.New("otp has invalid length")
	ErrWrongOTPFormat      = errors.New("wrong otp format")
)
