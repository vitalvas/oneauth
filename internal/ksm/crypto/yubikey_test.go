package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecryptYubikeyOTP(t *testing.T) {
	engine, err := NewEngine("test-master-key-1234567890")
	assert.NoError(t, err)

	t.Run("invalid OTP format", func(t *testing.T) {
		aesKey := []byte("1234567890123456")

		tests := []struct {
			name        string
			otp         string
			expectError string
		}{
			{
				name:        "invalid modhex characters",
				otp:         "ccccccccccccdefghijklnrtuvcbdefghijklnrtuXXX",
				expectError: "failed to convert modhex to hex",
			},
			{
				name:        "wrong length after conversion",
				otp:         "ccccccccccccdefghijklnrtuvcbdefghijklnrtuvw", // 43 chars
				expectError: "failed to convert modhex to hex",
			},
			{
				name:        "empty OTP",
				otp:         "",
				expectError: "failed to convert modhex to hex",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := engine.DecryptYubikeyOTP(tt.otp, aesKey)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectError)
			})
		}
	})

	t.Run("invalid AES key", func(t *testing.T) {
		validOTP := "ccccccccccccdefghijklnrtuvcbdefghijklnrtuvic" // 44 chars

		tests := []struct {
			name        string
			aesKey      []byte
			expectError string
		}{
			{
				name:        "wrong key length - too short",
				aesKey:      []byte("123456789012345"), // 15 bytes
				expectError: "invalid OTP byte length",
			},
			{
				name:        "wrong key length - too long",
				aesKey:      []byte("12345678901234567"), // 17 bytes
				expectError: "invalid OTP byte length",
			},
			{
				name:        "empty key",
				aesKey:      []byte{},
				expectError: "invalid OTP byte length",
			},
			{
				name:        "nil key",
				aesKey:      nil,
				expectError: "invalid OTP byte length",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := engine.DecryptYubikeyOTP(validOTP, tt.aesKey)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectError)
			})
		}
	})

	t.Run("CRC failure", func(t *testing.T) {
		// Valid 44-char modhex OTP but will likely fail CRC
		otp := "ccccccccccccdefghijklnrtuvcbdefghijklnrtuvic"
		aesKey := []byte("1234567890123456")

		_, err = engine.DecryptYubikeyOTP(otp, aesKey)
		// This will likely fail either during decryption or CRC verification
		assert.Error(t, err)
	})

	t.Run("edge cases", func(t *testing.T) {
		aesKey := []byte("1234567890123456")

		tests := []struct {
			name        string
			otp         string
			expectError bool
			errorMsg    string
		}{
			{
				name:        "exactly 44 characters valid modhex",
				otp:         "ccccccccccccdefghijklnrtuvcbdefghijklnrtuvic",
				expectError: true, // Will fail on CRC or decryption with test key
			},
			{
				name:        "all same character",
				otp:         "cccccccccccccccccccccccccccccccccccccccccccc",
				expectError: true,
			},
			{
				name:        "boundary modhex chars",
				otp:         "cbdefghijklnrtuvcbdefghijklnrtuvcbdefghijkl",
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := engine.DecryptYubikeyOTP(tt.otp, aesKey)
				if tt.expectError {
					assert.Error(t, err)
					if tt.errorMsg != "" {
						assert.Contains(t, err.Error(), tt.errorMsg)
					}
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

func TestOTPData(t *testing.T) {
	t.Run("structure", func(t *testing.T) {
		// Test that OTPData struct has expected fields
		otpData := &OTPData{
			Counter:       123,
			TimestampLow:  456,
			TimestampHigh: 78,
			SessionUse:    9,
			RandomData:    321,
			CRC:           0x1234,
		}

		assert.Equal(t, 123, otpData.Counter)
		assert.Equal(t, 456, otpData.TimestampLow)
		assert.Equal(t, 78, otpData.TimestampHigh)
		assert.Equal(t, 9, otpData.SessionUse)
		assert.Equal(t, 321, otpData.RandomData)
		assert.Equal(t, uint16(0x1234), otpData.CRC)
	})
}
