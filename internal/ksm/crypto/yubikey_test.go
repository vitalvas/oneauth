package crypto

import (
	"crypto/aes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vitalvas/oneauth/internal/ykshared"
)

func TestOTPDataStructFields(t *testing.T) {
	t.Run("zero value OTPData", func(t *testing.T) {
		data := OTPData{}
		assert.Equal(t, 0, data.Counter)
		assert.Equal(t, 0, data.TimestampLow)
		assert.Equal(t, 0, data.TimestampHigh)
		assert.Equal(t, 0, data.SessionUse)
		assert.Equal(t, 0, data.RandomData)
		assert.Equal(t, uint16(0), data.CRC)
	})

	t.Run("populated OTPData", func(t *testing.T) {
		data := OTPData{
			Counter:       42,
			TimestampLow:  0x1234,
			TimestampHigh: 0x56,
			SessionUse:    7,
			RandomData:    0xABCD,
			CRC:           0xBEEF,
		}
		assert.Equal(t, 42, data.Counter)
		assert.Equal(t, 0x1234, data.TimestampLow)
		assert.Equal(t, 0x56, data.TimestampHigh)
		assert.Equal(t, 7, data.SessionUse)
		assert.Equal(t, 0xABCD, data.RandomData)
		assert.Equal(t, uint16(0xBEEF), data.CRC)
	})

	t.Run("max value OTPData fields", func(t *testing.T) {
		data := OTPData{
			Counter:       65535,
			TimestampLow:  65535,
			TimestampHigh: 255,
			SessionUse:    255,
			RandomData:    65535,
			CRC:           0xFFFF,
		}
		assert.Equal(t, 65535, data.Counter)
		assert.Equal(t, 65535, data.TimestampLow)
		assert.Equal(t, 255, data.TimestampHigh)
		assert.Equal(t, 255, data.SessionUse)
		assert.Equal(t, 65535, data.RandomData)
		assert.Equal(t, uint16(0xFFFF), data.CRC)
	})
}

func TestDecryptYubikeyOTPParsedFields(t *testing.T) {
	engine, err := NewEngine("test-master-key-for-yubikey")
	assert.NoError(t, err)

	aesKey := []byte("abcdef0123456789") // 16 bytes

	t.Run("verify parsed counter value", func(t *testing.T) {
		plaintext := buildValidOTPPlaintext(500, 0x0001, 0x02, 3)
		otp := encryptAndEncodeOTP(t, plaintext, aesKey)

		result, err := engine.DecryptYubikeyOTP(otp, aesKey)
		assert.NoError(t, err)
		assert.Equal(t, 500, result.Counter)
	})

	t.Run("verify parsed timestamp fields", func(t *testing.T) {
		plaintext := buildValidOTPPlaintext(1, 0xABCD, 0xEF, 10)
		otp := encryptAndEncodeOTP(t, plaintext, aesKey)

		result, err := engine.DecryptYubikeyOTP(otp, aesKey)
		assert.NoError(t, err)
		assert.Equal(t, 0xABCD, result.TimestampLow)
		assert.Equal(t, 0xEF, result.TimestampHigh)
	})

	t.Run("verify parsed session use", func(t *testing.T) {
		plaintext := buildValidOTPPlaintext(1, 0x0000, 0x00, 200)
		otp := encryptAndEncodeOTP(t, plaintext, aesKey)

		result, err := engine.DecryptYubikeyOTP(otp, aesKey)
		assert.NoError(t, err)
		assert.Equal(t, 200, result.SessionUse)
	})

	t.Run("verify CRC field matches random data bytes", func(t *testing.T) {
		plaintext := buildValidOTPPlaintext(1, 0x0001, 0x02, 3)
		otp := encryptAndEncodeOTP(t, plaintext, aesKey)

		result, err := engine.DecryptYubikeyOTP(otp, aesKey)
		assert.NoError(t, err)
		// CRC and RandomData are both read from bytes 14-15
		assert.Equal(t, uint16(result.RandomData), result.CRC)
	})

	t.Run("wrong AES key fails decryption with CRC error", func(t *testing.T) {
		plaintext := buildValidOTPPlaintext(1, 0x0001, 0x02, 3)
		otp := encryptAndEncodeOTP(t, plaintext, aesKey)

		wrongKey := []byte("9876543210fedcba")
		_, err := engine.DecryptYubikeyOTP(otp, wrongKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CRC verification failed")
	})

	t.Run("nil AES key", func(t *testing.T) {
		validOTP := "cccccccccccccccccccccccccccccccccccccccccccc"
		_, err := engine.DecryptYubikeyOTP(validOTP, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create AES cipher")
	})

	t.Run("32 byte AES key", func(t *testing.T) {
		validOTP := "cccccccccccccccccccccccccccccccccccccccccccc"
		key32 := []byte("12345678901234561234567890123456")
		_, err := engine.DecryptYubikeyOTP(validOTP, key32)
		assert.Error(t, err)
		// 32-byte key is valid for AES-256, so it will decrypt but CRC will fail
		assert.Contains(t, err.Error(), "CRC verification failed")
	})
}

// buildValidOTPPlaintext creates a 16-byte OTP plaintext with valid CRC.
func buildValidOTPPlaintext(counter int, timestampLow int, timestampHigh int, sessionUse int) []byte {
	plaintext := make([]byte, 16)
	// Private ID (6 bytes)
	plaintext[0] = 0x01
	plaintext[1] = 0x02
	plaintext[2] = 0x03
	plaintext[3] = 0x04
	plaintext[4] = 0x05
	plaintext[5] = 0x06
	// Counter (2 bytes LE) at offset 8-9
	binary.LittleEndian.PutUint16(plaintext[8:10], uint16(counter))
	// Timestamp low (2 bytes LE) at offset 10-11
	binary.LittleEndian.PutUint16(plaintext[10:12], uint16(timestampLow))
	// Timestamp high (1 byte) at offset 12
	plaintext[12] = byte(timestampHigh)
	// Session use (1 byte) at offset 13
	plaintext[13] = byte(sessionUse)
	// Calculate CRC over first 14 bytes and place at offset 14-15
	crc := ykshared.CalculateCRC16(plaintext[:14])
	binary.LittleEndian.PutUint16(plaintext[14:16], crc)
	return plaintext
}

// encryptAndEncodeOTP encrypts plaintext with AES-128 ECB and returns modhex OTP string.
func encryptAndEncodeOTP(t *testing.T, plaintext []byte, aesKey []byte) string {
	t.Helper()
	block, err := aes.NewCipher(aesKey)
	assert.NoError(t, err)
	encrypted := make([]byte, 16)
	block.Encrypt(encrypted, plaintext)

	otpBytes := make([]byte, 0, 22)
	otpBytes = append(otpBytes, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00) // key ID
	otpBytes = append(otpBytes, encrypted...)

	return ykshared.BytesToModhex(otpBytes)
}
