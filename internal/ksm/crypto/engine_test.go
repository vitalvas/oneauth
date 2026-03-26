package crypto

import (
	"crypto/aes"
	"encoding/base64"
	"encoding/binary"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vitalvas/oneauth/internal/ykshared"
)

func TestNewEngine(t *testing.T) {
	tests := []struct {
		name      string
		masterKey string
		wantErr   bool
	}{
		{
			name:      "valid master key",
			masterKey: "test-master-key-1234567890",
			wantErr:   false,
		},
		{
			name:      "empty master key",
			masterKey: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, err := NewEngine(tt.masterKey)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, engine)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, engine)
				assert.Equal(t, 32, len(engine.masterKey))
			}
		})
	}
}

func TestEncryptDecryptAESKey(t *testing.T) {
	engine, err := NewEngine("test-master-key-1234567890")
	assert.NoError(t, err)

	keyID := "cccccccccccc"
	aesKey := []byte("1234567890123456") // 16 bytes

	// Test encryption
	encrypted, err := engine.EncryptAESKey(keyID, aesKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	// Test decryption
	decrypted, err := engine.DecryptAESKey(keyID, encrypted)
	assert.NoError(t, err)
	assert.Equal(t, aesKey, decrypted)
}

func TestEncryptAESKey_InvalidLength(t *testing.T) {
	engine, err := NewEngine("test-master-key-1234567890")
	assert.NoError(t, err)

	keyID := "cccccccccccc"

	tests := []struct {
		name   string
		aesKey []byte
	}{
		{
			name:   "too short",
			aesKey: []byte("123456789012345"), // 15 bytes
		},
		{
			name:   "too long",
			aesKey: []byte("12345678901234567"), // 17 bytes
		},
		{
			name:   "empty",
			aesKey: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := engine.EncryptAESKey(keyID, tt.aesKey)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "AES key must be 16 bytes")
		})
	}
}

func TestDecryptAESKey_InvalidData(t *testing.T) {
	engine, err := NewEngine("test-master-key-1234567890")
	assert.NoError(t, err)

	keyID := "cccccccccccc"

	tests := []struct {
		name          string
		encryptedData string
		expectError   string
	}{
		{
			name:          "invalid base64",
			encryptedData: "invalid-base64!",
			expectError:   "failed to decode base64",
		},
		{
			name:          "too short data",
			encryptedData: base64.RawURLEncoding.EncodeToString([]byte("short")),
			expectError:   "encrypted data too short",
		},
		{
			name:          "corrupted data",
			encryptedData: base64.RawURLEncoding.EncodeToString(make([]byte, 32)), // valid length but invalid data
			expectError:   "failed to decrypt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := engine.DecryptAESKey(keyID, tt.encryptedData)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectError)
		})
	}
}

func TestDifferentKeyIDsProduceDifferentKeys(t *testing.T) {
	engine, err := NewEngine("test-master-key-1234567890")
	assert.NoError(t, err)

	aesKey := []byte("1234567890123456")

	// Encrypt same AES key with different key IDs
	encrypted1, err := engine.EncryptAESKey("cccccccccccc", aesKey)
	assert.NoError(t, err)

	encrypted2, err := engine.EncryptAESKey("dddddddddddd", aesKey)
	assert.NoError(t, err)

	// Different key IDs should produce different encrypted results
	assert.NotEqual(t, encrypted1, encrypted2)

	// But both should decrypt to the same original key when using correct key ID
	decrypted1, err := engine.DecryptAESKey("cccccccccccc", encrypted1)
	assert.NoError(t, err)
	assert.Equal(t, aesKey, decrypted1)

	decrypted2, err := engine.DecryptAESKey("dddddddddddd", encrypted2)
	assert.NoError(t, err)
	assert.Equal(t, aesKey, decrypted2)

	// Using wrong key ID should fail
	_, err = engine.DecryptAESKey("eeeeeeeeeeee", encrypted1)
	assert.Error(t, err)
}

func TestDeriveRowKey(t *testing.T) {
	engine, err := NewEngine("test-master-key-1234567890")
	assert.NoError(t, err)

	keyID1 := "cccccccccccc"
	keyID2 := "dddddddddddd"

	// Derive keys for different key IDs
	rowKey1, err := engine.deriveRowKey(keyID1)
	assert.NoError(t, err)
	assert.Equal(t, 32, len(rowKey1))

	rowKey2, err := engine.deriveRowKey(keyID2)
	assert.NoError(t, err)
	assert.Equal(t, 32, len(rowKey2))

	// Different key IDs should produce different row keys
	assert.NotEqual(t, rowKey1, rowKey2)

	// Same key ID should produce same row key
	rowKey1Again, err := engine.deriveRowKey(keyID1)
	assert.NoError(t, err)
	assert.Equal(t, rowKey1, rowKey1Again)
}

func TestEncryptionConsistency(t *testing.T) {
	engine, err := NewEngine("test-master-key-1234567890")
	assert.NoError(t, err)

	keyID := "cccccccccccc"
	aesKey := []byte("1234567890123456")

	// Encrypt multiple times - should produce different results due to random nonce
	encrypted1, err := engine.EncryptAESKey(keyID, aesKey)
	assert.NoError(t, err)

	encrypted2, err := engine.EncryptAESKey(keyID, aesKey)
	assert.NoError(t, err)

	// Different encryptions due to random nonce
	assert.NotEqual(t, encrypted1, encrypted2)

	// But both should decrypt to same value
	decrypted1, err := engine.DecryptAESKey(keyID, encrypted1)
	assert.NoError(t, err)

	decrypted2, err := engine.DecryptAESKey(keyID, encrypted2)
	assert.NoError(t, err)

	assert.Equal(t, aesKey, decrypted1)
	assert.Equal(t, aesKey, decrypted2)
	assert.Equal(t, decrypted1, decrypted2)
}

func TestEngine_MasterKeyHashing(t *testing.T) {
	// Test that different master keys produce different engines
	engine1, err := NewEngine("master-key-1")
	assert.NoError(t, err)

	engine2, err := NewEngine("master-key-2")
	assert.NoError(t, err)

	// Same operation with different engines should produce different results
	keyID := "cccccccccccc"
	aesKey := []byte("1234567890123456")

	encrypted1, err := engine1.EncryptAESKey(keyID, aesKey)
	assert.NoError(t, err)

	encrypted2, err := engine2.EncryptAESKey(keyID, aesKey)
	assert.NoError(t, err)

	// Different master keys should produce different encrypted results
	assert.NotEqual(t, encrypted1, encrypted2)

	// Cross-decryption should fail
	_, err = engine1.DecryptAESKey(keyID, encrypted2)
	assert.Error(t, err)

	_, err = engine2.DecryptAESKey(keyID, encrypted1)
	assert.Error(t, err)
}

func TestMasterKeyVariations(t *testing.T) {
	t.Run("long master key", func(t *testing.T) {
		// Test with very long master key
		longKey := strings.Repeat("a", 1000)

		engine, err := NewEngine(longKey)
		assert.NoError(t, err)
		assert.NotNil(t, engine)
		assert.Equal(t, 32, len(engine.masterKey)) // Should be hashed to 32 bytes
	})

	t.Run("special characters master key", func(t *testing.T) {
		// Test with special characters in master key
		specialKey := "!@#$%^&*()_+{}|:<>?[]\\;'\",./"

		engine, err := NewEngine(specialKey)
		assert.NoError(t, err)
		assert.NotNil(t, engine)

		// Should work normally
		keyID := "cccccccccccc"
		aesKey := []byte("1234567890123456")

		encrypted, err := engine.EncryptAESKey(keyID, aesKey)
		assert.NoError(t, err)

		decrypted, err := engine.DecryptAESKey(keyID, encrypted)
		assert.NoError(t, err)
		assert.Equal(t, aesKey, decrypted)
	})

	t.Run("unicode master key", func(t *testing.T) {
		// Test with Unicode characters in master key
		unicodeKey := "测试密钥🔐🗝️"

		engine, err := NewEngine(unicodeKey)
		assert.NoError(t, err)
		assert.NotNil(t, engine)

		// Should work normally
		keyID := "cccccccccccc"
		aesKey := []byte("1234567890123456")

		encrypted, err := engine.EncryptAESKey(keyID, aesKey)
		assert.NoError(t, err)

		decrypted, err := engine.DecryptAESKey(keyID, encrypted)
		assert.NoError(t, err)
		assert.Equal(t, aesKey, decrypted)
	})
}

func TestEncryptAESKey_EmptyKeyID(t *testing.T) {
	engine, err := NewEngine("test-master-key-1234567890")
	assert.NoError(t, err)

	aesKey := []byte("1234567890123456")

	// Empty key ID should still work (it's used in HKDF)
	encrypted, err := engine.EncryptAESKey("", aesKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := engine.DecryptAESKey("", encrypted)
	assert.NoError(t, err)
	assert.Equal(t, aesKey, decrypted)
}

func TestEncryptAESKey_LongKeyID(t *testing.T) {
	engine, err := NewEngine("test-master-key-1234567890")
	assert.NoError(t, err)

	aesKey := []byte("1234567890123456")
	longKeyID := strings.Repeat("c", 1000)

	// Long key ID should work
	encrypted, err := engine.EncryptAESKey(longKeyID, aesKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := engine.DecryptAESKey(longKeyID, encrypted)
	assert.NoError(t, err)
	assert.Equal(t, aesKey, decrypted)
}

func TestDeriveRowKey_Consistency(t *testing.T) {
	engine, err := NewEngine("test-master-key-1234567890")
	assert.NoError(t, err)

	keyID := "cccccccccccc"

	// Multiple calls should produce the same row key
	rowKey1, err := engine.deriveRowKey(keyID)
	assert.NoError(t, err)

	rowKey2, err := engine.deriveRowKey(keyID)
	assert.NoError(t, err)

	rowKey3, err := engine.deriveRowKey(keyID)
	assert.NoError(t, err)

	assert.Equal(t, rowKey1, rowKey2)
	assert.Equal(t, rowKey1, rowKey3)
	assert.Equal(t, 32, len(rowKey1))
}

func TestDecryptAESKey_WrongKeyID(t *testing.T) {
	engine, err := NewEngine("test-master-key-1234567890")
	assert.NoError(t, err)

	keyID1 := "cccccccccccc"
	keyID2 := "dddddddddddd"
	aesKey := []byte("1234567890123456")

	// Encrypt with keyID1
	encrypted, err := engine.EncryptAESKey(keyID1, aesKey)
	assert.NoError(t, err)

	// Try to decrypt with keyID2 - should fail
	_, err = engine.DecryptAESKey(keyID2, encrypted)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decrypt")
}

func TestEngine_MemoryClearing(t *testing.T) {
	engine, err := NewEngine("test-master-key-1234567890")
	assert.NoError(t, err)

	keyID := "cccccccccccc"

	// The derive row key should clear sensitive data
	// We can't directly test this, but we can ensure the function completes
	rowKey, err := engine.deriveRowKey(keyID)
	assert.NoError(t, err)
	assert.Equal(t, 32, len(rowKey))
	clear(rowKey) // Ensure we can clear it
}

func TestEngine_ConcurrentAccess(t *testing.T) {
	engine, err := NewEngine("test-master-key-1234567890")
	assert.NoError(t, err)

	keyID := "cccccccccccc"
	aesKey := []byte("1234567890123456")

	// Run multiple encryptions concurrently
	const numRoutines = 10
	results := make(chan string, numRoutines)
	errors := make(chan error, numRoutines)

	for i := 0; i < numRoutines; i++ {
		go func() {
			encrypted, err := engine.EncryptAESKey(keyID, aesKey)
			if err != nil {
				errors <- err
				return
			}
			results <- encrypted
		}()
	}

	// Collect results
	var encrypted []string
	for i := 0; i < numRoutines; i++ {
		select {
		case result := <-results:
			encrypted = append(encrypted, result)
		case err := <-errors:
			t.Fatalf("Encryption failed: %v", err)
		}
	}

	// All should succeed and be different (due to random nonces)
	assert.Equal(t, numRoutines, len(encrypted))
	for i := 0; i < numRoutines; i++ {
		for j := i + 1; j < numRoutines; j++ {
			assert.NotEqual(t, encrypted[i], encrypted[j])
		}
	}

	// All should decrypt to the same value
	for _, enc := range encrypted {
		decrypted, err := engine.DecryptAESKey(keyID, enc)
		assert.NoError(t, err)
		assert.Equal(t, aesKey, decrypted)
	}
}

func TestDecryptYubikeyOTP(t *testing.T) {
	engine, err := NewEngine("test-master-key-1234567890")
	assert.NoError(t, err)

	aesKey := []byte("1234567890123456")

	t.Run("invalid modhex characters", func(t *testing.T) {
		_, err := engine.DecryptYubikeyOTP("ccccccccccccdefghijklnrtuvcbdefghijklnrtuXXX", aesKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to convert modhex to hex")
	})

	t.Run("empty OTP", func(t *testing.T) {
		_, err := engine.DecryptYubikeyOTP("", aesKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to convert modhex to hex")
	})

	t.Run("wrong length after modhex conversion - too short", func(t *testing.T) {
		_, err := engine.DecryptYubikeyOTP("cccccccccccc", aesKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid OTP length after modhex conversion")
	})

	t.Run("wrong length after modhex conversion - too long", func(t *testing.T) {
		otp := strings.Repeat("c", 46)
		_, err := engine.DecryptYubikeyOTP(otp, aesKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid OTP length after modhex conversion")
	})

	t.Run("invalid AES key size", func(t *testing.T) {
		validOTP := "cccccccccccccccccccccccccccccccccccccccccccc"
		_, err := engine.DecryptYubikeyOTP(validOTP, []byte("short"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create AES cipher")
	})

	t.Run("CRC verification failure", func(t *testing.T) {
		// 44-char valid modhex OTP with valid AES key decrypts but CRC will not match
		validOTP := "ccccccccccccdefghijklnrtuvcbdefghijklnrtuvic"
		_, err := engine.DecryptYubikeyOTP(validOTP, aesKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CRC verification failed")
	})

	t.Run("successful decryption with valid CRC", func(t *testing.T) {
		// Build a valid OTP by encrypting known plaintext with proper CRC
		plaintext := make([]byte, 16)
		// Private ID (6 bytes)
		plaintext[0] = 0x01
		plaintext[1] = 0x02
		plaintext[2] = 0x03
		plaintext[3] = 0x04
		plaintext[4] = 0x05
		plaintext[5] = 0x06
		// Counter (2 bytes LE) at offset 8-9
		binary.LittleEndian.PutUint16(plaintext[8:10], 42)
		// Timestamp low (2 bytes LE) at offset 10-11
		binary.LittleEndian.PutUint16(plaintext[10:12], 0x1234)
		// Timestamp high (1 byte) at offset 12
		plaintext[12] = 0x56
		// Session use (1 byte) at offset 13
		plaintext[13] = 7

		// Calculate CRC over first 14 bytes and place at offset 14-15
		crc := ykshared.CalculateCRC16(plaintext[:14])
		binary.LittleEndian.PutUint16(plaintext[14:16], crc)

		// Encrypt with AES-128 ECB
		block, cipherErr := aes.NewCipher(aesKey)
		assert.NoError(t, cipherErr)
		encrypted := make([]byte, 16)
		block.Encrypt(encrypted, plaintext)

		// Key ID bytes (6 bytes) + encrypted (16 bytes) = 22 bytes
		otpBytes := make([]byte, 0, 22)
		otpBytes = append(otpBytes, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00)
		otpBytes = append(otpBytes, encrypted...)

		// Convert to modhex
		otpModhex := ykshared.BytesToModhex(otpBytes)
		assert.Equal(t, 44, len(otpModhex))

		result, err := engine.DecryptYubikeyOTP(otpModhex, aesKey)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 42, result.Counter)
		assert.Equal(t, 0x1234, result.TimestampLow)
		assert.Equal(t, 0x56, result.TimestampHigh)
		assert.Equal(t, 7, result.SessionUse)
	})

	t.Run("various modhex patterns with CRC failure", func(t *testing.T) {
		patterns := []string{
			"cbdefghijklnrtuvcbdefghijklnrtuvcbdefghijkl",
			"cdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcd",
			"dddddddddddddefghijklnrtuvcbdefghijklnrtuvic",
		}
		for _, otp := range patterns {
			_, err := engine.DecryptYubikeyOTP(otp, aesKey)
			assert.Error(t, err)
		}
	})
}

func BenchmarkNewEngine(b *testing.B) {
	masterKey := "test-master-key-1234567890"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewEngine(masterKey)
	}
}

func BenchmarkEncryptAESKey(b *testing.B) {
	engine, _ := NewEngine("test-master-key-1234567890")
	keyID := "cccccccccccc"
	aesKey := []byte("1234567890123456")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.EncryptAESKey(keyID, aesKey)
	}
}

func BenchmarkDecryptAESKey(b *testing.B) {
	engine, _ := NewEngine("test-master-key-1234567890")
	keyID := "cccccccccccc"
	aesKey := []byte("1234567890123456")

	// Pre-encrypt data for benchmarking
	encrypted, _ := engine.EncryptAESKey(keyID, aesKey)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.DecryptAESKey(keyID, encrypted)
	}
}

func BenchmarkDeriveRowKey(b *testing.B) {
	engine, _ := NewEngine("test-master-key-1234567890")
	keyID := "cccccccccccc"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rowKey, _ := engine.deriveRowKey(keyID)
		clear(rowKey) // Clean up
	}
}

func BenchmarkDecryptYubikeyOTP(b *testing.B) {
	engine, _ := NewEngine("test-master-key-1234567890")
	otp := "ccccccccccccdefghijklnrtuvcbdefghijklnrtuvic"
	aesKey := []byte("1234567890123456")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.DecryptYubikeyOTP(otp, aesKey)
	}
}

func BenchmarkEncryptDecryptRoundtrip(b *testing.B) {
	engine, _ := NewEngine("test-master-key-1234567890")
	keyID := "cccccccccccc"
	aesKey := []byte("1234567890123456")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encrypted, _ := engine.EncryptAESKey(keyID, aesKey)
		decrypted, _ := engine.DecryptAESKey(keyID, encrypted)
		clear(decrypted) // Clean up
	}
}
