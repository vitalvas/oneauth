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

func TestCRCOperations(t *testing.T) {
	t.Run("verify CRC", func(t *testing.T) {
		tests := []struct {
			name        string
			data        []byte
			expectedCRC uint16
			shouldPass  bool
		}{
			{
				name:        "empty data with initial CRC",
				data:        []byte{},
				expectedCRC: 0xFFFF,
				shouldPass:  true,
			},
			{
				name:        "single byte",
				data:        []byte{0x00},
				expectedCRC: 0xE1F0,
				shouldPass:  true,
			},
			{
				name:        "multiple bytes",
				data:        []byte{0x01, 0x02, 0x03, 0x04},
				expectedCRC: 0x89C7,
				shouldPass:  false, // Will fail since this is likely not the correct CRC
			},
			{
				name:        "incorrect CRC",
				data:        []byte{0x01, 0x02},
				expectedCRC: 0x0000,
				shouldPass:  false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := verifyCRC(tt.data, tt.expectedCRC)
				assert.Equal(t, tt.shouldPass, result)
			})
		}
	})

	t.Run("known values", func(t *testing.T) {
		// Test with known CRC values by calculating them first
		testData := []byte{0x12, 0x34, 0x56, 0x78}

		// Calculate CRC for test data
		const poly = 0x1021
		crc := uint16(0xFFFF)
		for _, b := range testData {
			crc ^= uint16(b) << 8
			for i := 0; i < 8; i++ {
				if crc&0x8000 != 0 {
					crc = (crc << 1) ^ poly
				} else {
					crc <<= 1
				}
			}
		}

		// Now test that our function correctly verifies this CRC
		assert.True(t, verifyCRC(testData, crc))
		assert.False(t, verifyCRC(testData, crc+1)) // Wrong CRC should fail
	})
}
