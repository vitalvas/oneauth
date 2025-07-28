package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vitalvas/oneauth/internal/ksm/crypto"
	"github.com/vitalvas/oneauth/internal/ksm/database"
)

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
