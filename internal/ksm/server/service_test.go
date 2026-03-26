package server

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vitalvas/oneauth/internal/ksm/crypto"
	"github.com/vitalvas/oneauth/internal/ksm/database"
	"github.com/vitalvas/oneauth/internal/yksoft"
)

func TestParseAESKey(t *testing.T) {
	server := setupTestServer(t)

	tests := []struct {
		name     string
		input    string
		expected []byte
		hasError bool
	}{
		// Hex format tests
		{
			name:     "Hex/Valid 32 chars lowercase",
			input:    "31323334353637383930313233343536",
			expected: []byte("1234567890123456"),
			hasError: false,
		},
		{
			name:     "Hex/Valid 32 chars uppercase",
			input:    "31323334353637383930313233343536",
			expected: []byte("1234567890123456"),
			hasError: false,
		},
		{
			name:     "Hex/Valid mixed case",
			input:    "31323334353637383930313233343536",
			expected: []byte("1234567890123456"),
			hasError: false,
		},
		{
			name:     "Hex/With whitespace",
			input:    " 31323334353637383930313233343536 ",
			expected: []byte("1234567890123456"),
			hasError: false,
		},
		{
			name:     "Hex/Invalid characters",
			input:    "3132333435363738393031323334353Z",
			expected: nil,
			hasError: true,
		},
		{
			name:     "Hex/Wrong length - too short",
			input:    "313233343536373839303132333435",
			expected: nil,
			hasError: true,
		},
		{
			name:     "Hex/Wrong length - too long",
			input:    "3132333435363738393031323334353637",
			expected: nil,
			hasError: true,
		},

		// Base64 format tests
		{
			name:     "Base64/Valid URL encoding",
			input:    "MTIzNDU2Nzg5MDEyMzQ1Ng",
			expected: []byte("1234567890123456"),
			hasError: false,
		},
		{
			name:     "Base64/Valid standard encoding",
			input:    "MTIzNDU2Nzg5MDEyMzQ1Ng==",
			expected: []byte("1234567890123456"),
			hasError: false,
		},
		{
			name:     "Base64/With whitespace",
			input:    " MTIzNDU2Nzg5MDEyMzQ1Ng== ",
			expected: []byte("1234567890123456"),
			hasError: false,
		},
		{
			name:     "Base64/Invalid characters",
			input:    "MTIzNDU2Nzg5MDEyMzQ1Ng==!",
			expected: nil,
			hasError: true,
		},
		{
			name:     "Base64/Wrong length",
			input:    "MTIzNDU2Nzg5MA==",
			expected: nil,
			hasError: true,
		},

		// Edge cases
		{
			name:     "EdgeCase/Empty string",
			input:    "",
			expected: nil,
			hasError: true,
		},
		{
			name:     "EdgeCase/Only whitespace",
			input:    "   ",
			expected: nil,
			hasError: true,
		},
		{
			name:     "EdgeCase/Random string",
			input:    "not_hex_or_base64",
			expected: nil,
			hasError: true,
		},
		{
			name:     "EdgeCase/Almost hex but wrong chars",
			input:    "31323334353637383930313233343536GH",
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := server.parseAESKey(tt.input)

			if tt.hasError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestValidateOTPFormat(t *testing.T) {
	mockDB, err := database.NewMockDB()
	assert.NoError(t, err)
	defer mockDB.Close()

	cryptoEngine, err := crypto.NewEngine("test-master-key-12345678901234567890")
	assert.NoError(t, err)

	server := createTestServer(mockDB, cryptoEngine)

	tests := []struct {
		name      string
		otp       string
		expectErr bool
	}{
		{
			name:      "valid OTP",
			otp:       "ccccccccccccdefghijklnrtuvcbdefghijklnrtuvic", // This is exactly 44 chars
			expectErr: false,
		},
		{
			name:      "too short",
			otp:       "ccccccccccccdefghijklnrtuvcbdefghijklnrtu",
			expectErr: true,
		},
		{
			name:      "too long",
			otp:       "ccccccccccccdefghijklnrtuvcbdefghijklnrtuvice",
			expectErr: true,
		},
		{
			name:      "invalid characters",
			otp:       "ccccccccccccdefghijklnrtuvcbdefghijklnrtuZi",
			expectErr: true,
		},
		{
			name:      "empty OTP",
			otp:       "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := server.ValidateOTPFormat(tt.otp)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStoreKey(t *testing.T) {
	mockDB, err := database.NewMockDB()
	assert.NoError(t, err)
	defer mockDB.Close()

	cryptoEngine, err := crypto.NewEngine("test-master-key-12345678901234567890")
	assert.NoError(t, err)

	server := createTestServer(mockDB, cryptoEngine)

	tests := []struct {
		name        string
		keyID       string
		aesKeyB64   string
		description string
		expectErr   bool
	}{
		{
			name:        "valid key storage",
			keyID:       "cccccccccccc",
			aesKeyB64:   "MTIzNDU2Nzg5MDEyMzQ1Ng", // 16 bytes base64 encoded
			description: "Test key",
			expectErr:   false,
		},
		{
			name:        "invalid key ID length",
			keyID:       "ccc",
			aesKeyB64:   "MTIzNDU2Nzg5MDEyMzQ1Ng",
			description: "Test key",
			expectErr:   true,
		},
		{
			name:        "invalid modhex in key ID",
			keyID:       "ccccccccccXX",
			aesKeyB64:   "MTIzNDU2Nzg5MDEyMzQ1Ng",
			description: "Test key",
			expectErr:   true,
		},
		{
			name:        "invalid base64 AES key",
			keyID:       "cccccccccccc",
			aesKeyB64:   "invalid-base64!",
			description: "Test key",
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset database state for test isolation
			err := mockDB.Reset()
			assert.NoError(t, err)

			err = server.StoreKey(tt.keyID, tt.aesKeyB64, tt.description)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify the key was actually stored correctly
				storedKey, err := mockDB.GetKey(tt.keyID)
				assert.NoError(t, err)
				assert.Equal(t, tt.keyID, storedKey.KeyID)
				assert.Equal(t, tt.description, storedKey.Description)
				assert.NotEmpty(t, storedKey.AESKeyEncrypted)
			}
		})
	}
}

func TestDecryptOTP_InvalidFormat(t *testing.T) {
	mockDB, err := database.NewMockDB()
	assert.NoError(t, err)
	defer mockDB.Close()

	cryptoEngine, err := crypto.NewEngine("test-master-key-12345678901234567890")
	assert.NoError(t, err)

	server := createTestServer(mockDB, cryptoEngine)

	response, err := server.DecryptOTP("invalid-otp")
	assert.NoError(t, err)
	assert.Equal(t, "ERROR", response.Status)
	assert.Equal(t, "INVALID_OTP", response.ErrorCode)
}

func TestDecryptOTP_KeyNotFound(t *testing.T) {
	mockDB, err := database.NewMockDB()
	assert.NoError(t, err)
	defer mockDB.Close()

	cryptoEngine, err := crypto.NewEngine("test-master-key-12345678901234567890")
	assert.NoError(t, err)

	server := createTestServer(mockDB, cryptoEngine)

	// No need to mock - empty database will naturally return no rows
	response, err := server.DecryptOTP("ccccccccccccdefghijklnrtuvcbdefghijklnrtuvic")
	assert.NoError(t, err)
	assert.Equal(t, "ERROR", response.Status)
	assert.Equal(t, "KEY_NOT_FOUND", response.ErrorCode)
}

func TestListKeys(t *testing.T) {
	mockDB, err := database.NewMockDB()
	assert.NoError(t, err)
	defer mockDB.Close()

	cryptoEngine, err := crypto.NewEngine("test-master-key-12345678901234567890")
	assert.NoError(t, err)

	server := createTestServer(mockDB, cryptoEngine)

	// Store test keys in the database
	err = server.StoreKey("cccccccccccc", "MTIzNDU2Nzg5MDEyMzQ1Ng", "Test key 1")
	assert.NoError(t, err)
	err = server.StoreKey("dddddddddddd", "MTIzNDU2Nzg5MDEyMzQ1Ng", "Test key 2")
	assert.NoError(t, err)

	keys, err := server.ListKeys()
	assert.NoError(t, err)
	assert.Len(t, keys, 2)

	// Verify key details (order might vary, so check by ID)
	keyMap := make(map[string]*database.YubikeyKey)
	for _, key := range keys {
		keyMap[key.KeyID] = key
	}

	assert.Contains(t, keyMap, "cccccccccccc")
	assert.Equal(t, "Test key 1", keyMap["cccccccccccc"].Description)
	assert.Contains(t, keyMap, "dddddddddddd")
	assert.Equal(t, "Test key 2", keyMap["dddddddddddd"].Description)
}

func TestDeleteKey(t *testing.T) {
	mockDB, err := database.NewMockDB()
	assert.NoError(t, err)
	defer mockDB.Close()

	cryptoEngine, err := crypto.NewEngine("test-master-key-12345678901234567890")
	assert.NoError(t, err)

	server := createTestServer(mockDB, cryptoEngine)

	// First store a key to delete
	err = server.StoreKey("cccccccccccc", "MTIzNDU2Nzg5MDEyMzQ1Ng", "Test key")
	assert.NoError(t, err)

	// Verify key exists
	keys, err := server.ListKeys()
	assert.NoError(t, err)
	assert.Len(t, keys, 1)

	// Delete the key
	err = server.DeleteKey("cccccccccccc")
	assert.NoError(t, err)

	// Verify key is no longer listed (soft delete)
	keys, err = server.ListKeys()
	assert.NoError(t, err)
	assert.Len(t, keys, 0)
}

func TestDeleteKey_NonExistentKey(t *testing.T) {
	server := setupTestServer(t)

	// SQLite soft-delete via UPDATE does not error on non-existent keys
	err := server.DeleteKey("cccccccccccc")
	assert.NoError(t, err)
}

func TestListKeys_EmptyDatabase(t *testing.T) {
	server := setupTestServer(t)

	keys, err := server.ListKeys()
	assert.NoError(t, err)
	assert.Empty(t, keys)
}

func TestDecryptOTP_DecryptionFailed(t *testing.T) {
	server := setupTestServer(t)

	// Store a key so GetKey succeeds
	err := server.StoreKey("cccccccccccc", "MTIzNDU2Nzg5MDEyMzQ1Ng", "Test key")
	assert.NoError(t, err)

	// Valid 44-char modhex OTP: 12 key ID + 32 encrypted data
	// The encrypted data is arbitrary, so CRC check will fail
	validOTP := "ccccccccccccjktuvurlnlnvghubeukgkejrliudllkv"
	assert.Len(t, validOTP, 44)

	response, err := server.DecryptOTP(validOTP)
	assert.NoError(t, err)
	assert.Equal(t, "ERROR", response.Status)
	assert.Equal(t, "DECRYPTION_FAILED", response.ErrorCode)
	assert.Equal(t, "Decryption failed", response.Message)
}

func TestDecryptOTP_AllErrorPaths(t *testing.T) {
	// Valid 44-char modhex OTPs for testing
	const validOTPWithKeyCC = "ccccccccccccjktuvurlnlnvghubeukgkejrliudllkv"
	const validOTPWithKeyDD = "ddddddddddddjktuvurlnlnvghubeukgkejrliudllkv"

	t.Run("invalid OTP format", func(t *testing.T) {
		server := setupTestServer(t)

		response, err := server.DecryptOTP("short")
		assert.NoError(t, err)
		assert.Equal(t, "ERROR", response.Status)
		assert.Equal(t, "INVALID_OTP", response.ErrorCode)
	})

	t.Run("key not found", func(t *testing.T) {
		server := setupTestServer(t)

		response, err := server.DecryptOTP(validOTPWithKeyCC)
		assert.NoError(t, err)
		assert.Equal(t, "ERROR", response.Status)
		assert.Equal(t, "KEY_NOT_FOUND", response.ErrorCode)
	})

	t.Run("decryption failed with stored key", func(t *testing.T) {
		server := setupTestServer(t)

		err := server.StoreKey("cccccccccccc", "MTIzNDU2Nzg5MDEyMzQ1Ng", "Test key")
		assert.NoError(t, err)

		response, err := server.DecryptOTP(validOTPWithKeyCC)
		assert.NoError(t, err)
		assert.Equal(t, "ERROR", response.Status)
		assert.Equal(t, "DECRYPTION_FAILED", response.ErrorCode)
	})

	t.Run("empty OTP", func(t *testing.T) {
		server := setupTestServer(t)

		response, err := server.DecryptOTP("")
		assert.NoError(t, err)
		assert.Equal(t, "ERROR", response.Status)
		assert.Equal(t, "INVALID_OTP", response.ErrorCode)
	})

	t.Run("invalid modhex characters", func(t *testing.T) {
		server := setupTestServer(t)

		response, err := server.DecryptOTP("XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXZZZZ")
		assert.NoError(t, err)
		assert.Equal(t, "ERROR", response.Status)
		assert.Equal(t, "INVALID_OTP", response.ErrorCode)
	})

	t.Run("key not found for different key ID", func(t *testing.T) {
		server := setupTestServer(t)

		// Store key with "cccccccccccc" but OTP uses "dddddddddddd"
		err := server.StoreKey("cccccccccccc", "MTIzNDU2Nzg5MDEyMzQ1Ng", "Test key")
		assert.NoError(t, err)

		response, err := server.DecryptOTP(validOTPWithKeyDD)
		assert.NoError(t, err)
		assert.Equal(t, "ERROR", response.Status)
		assert.Equal(t, "KEY_NOT_FOUND", response.ErrorCode)
	})
}

func TestStoreKey_AdditionalCases(t *testing.T) {
	t.Run("store with hex AES key", func(t *testing.T) {
		server := setupTestServer(t)

		err := server.StoreKey("cccccccccccc", "31323334353637383930313233343536", "Test hex key")
		assert.NoError(t, err)

		storedKey, err := server.ListKeys()
		assert.NoError(t, err)
		assert.Len(t, storedKey, 1)
		assert.Equal(t, "Test hex key", storedKey[0].Description)
	})

	t.Run("store with empty description", func(t *testing.T) {
		server := setupTestServer(t)

		err := server.StoreKey("cccccccccccc", "MTIzNDU2Nzg5MDEyMzQ1Ng", "")
		assert.NoError(t, err)
	})

	t.Run("store duplicate key ID", func(t *testing.T) {
		server := setupTestServer(t)

		err := server.StoreKey("cccccccccccc", "MTIzNDU2Nzg5MDEyMzQ1Ng", "First")
		assert.NoError(t, err)

		err = server.StoreKey("cccccccccccc", "MTIzNDU2Nzg5MDEyMzQ1Ng", "Second")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
	})

	t.Run("store with empty key ID", func(t *testing.T) {
		server := setupTestServer(t)

		err := server.StoreKey("", "MTIzNDU2Nzg5MDEyMzQ1Ng", "Test")
		assert.Error(t, err)
	})

	t.Run("store with empty AES key", func(t *testing.T) {
		server := setupTestServer(t)

		err := server.StoreKey("cccccccccccc", "", "Test")
		assert.Error(t, err)
	})
}

func TestListKeys_DatabaseError(t *testing.T) {
	server := setupTestServer(t)

	// Close the DB to force errors
	server.db.Close()

	keys, err := server.ListKeys()
	assert.Error(t, err)
	assert.Nil(t, keys)
}

func TestDeleteKey_DatabaseError(t *testing.T) {
	server := setupTestServer(t)

	// Close the DB to force errors
	server.db.Close()

	err := server.DeleteKey("cccccccccccc")
	assert.Error(t, err)
}

func TestDeleteKey_AfterStore(t *testing.T) {
	server := setupTestServer(t)

	err := server.StoreKey("cccccccccccc", "MTIzNDU2Nzg5MDEyMzQ1Ng", "Test")
	assert.NoError(t, err)

	// Delete the key
	err = server.DeleteKey("cccccccccccc")
	assert.NoError(t, err)

	// Verify it's gone from list
	keys, err := server.ListKeys()
	assert.NoError(t, err)
	assert.Empty(t, keys)

	// Also verify GetKey fails (soft deleted) using a valid 44-char OTP
	response, err := server.DecryptOTP("ccccccccccccjktuvurlnlnvghubeukgkejrliudllkv")
	assert.NoError(t, err)
	assert.Equal(t, "KEY_NOT_FOUND", response.ErrorCode)
}

func TestDecryptOTP_FullSuccess(t *testing.T) {
	server := setupTestServer(t)

	aesKey := []byte("1234567890123456")
	yk, err := yksoft.NewSoftwareYubikey(&yksoft.Config{
		KeyID:     "cccccccccccc",
		PrivateID: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
		AESKey:    aesKey,
	})
	assert.NoError(t, err)

	// Store key using base64 encoding of the same AES key
	aesKeyB64 := base64.RawURLEncoding.EncodeToString(aesKey)
	err = server.StoreKey(yk.GetKeyID(), aesKeyB64, "Test key")
	assert.NoError(t, err)

	otpResult, err := yk.GenerateOTP()
	assert.NoError(t, err)

	response, err := server.DecryptOTP(otpResult.OTP)
	assert.NoError(t, err)
	assert.Equal(t, "OK", response.Status)
	assert.Equal(t, "cccccccccccc", response.KeyID)
	assert.NotZero(t, response.DecryptedAt)
}

func TestDecryptOTP_ReplayDetected(t *testing.T) {
	server := setupTestServer(t)

	aesKey := []byte("1234567890123456")
	yk, err := yksoft.NewSoftwareYubikey(&yksoft.Config{
		KeyID:     "cccccccccccc",
		PrivateID: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
		AESKey:    aesKey,
	})
	assert.NoError(t, err)

	aesKeyB64 := base64.RawURLEncoding.EncodeToString(aesKey)
	err = server.StoreKey(yk.GetKeyID(), aesKeyB64, "Test key")
	assert.NoError(t, err)

	otpResult, err := yk.GenerateOTP()
	assert.NoError(t, err)

	// First use succeeds
	response, err := server.DecryptOTP(otpResult.OTP)
	assert.NoError(t, err)
	assert.Equal(t, "OK", response.Status)

	// Replay is detected
	response, err = server.DecryptOTP(otpResult.OTP)
	assert.NoError(t, err)
	assert.Equal(t, "ERROR", response.Status)
	assert.Equal(t, "REPLAY_DETECTED", response.ErrorCode)
}

func TestDecryptOTP_UpdatesKeyUsage(t *testing.T) {
	server := setupTestServer(t)

	aesKey := []byte("1234567890123456")
	yk, err := yksoft.NewSoftwareYubikey(&yksoft.Config{
		KeyID:     "cccccccccccc",
		PrivateID: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
		AESKey:    aesKey,
	})
	assert.NoError(t, err)

	aesKeyB64 := base64.RawURLEncoding.EncodeToString(aesKey)
	err = server.StoreKey(yk.GetKeyID(), aesKeyB64, "Test key")
	assert.NoError(t, err)

	otpResult, err := yk.GenerateOTP()
	assert.NoError(t, err)

	response, err := server.DecryptOTP(otpResult.OTP)
	assert.NoError(t, err)
	assert.Equal(t, "OK", response.Status)

	// Verify key usage was updated
	keys, err := server.ListKeys()
	assert.NoError(t, err)
	assert.Len(t, keys, 1)
	assert.Equal(t, 1, keys[0].UsageCount)
	assert.NotNil(t, keys[0].LastUsedAt)
}
