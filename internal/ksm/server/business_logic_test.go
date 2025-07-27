package server

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vitalvas/oneauth/internal/ksm/crypto"
	"github.com/vitalvas/oneauth/internal/ksm/database"
)

func TestValidateOTPFormat(t *testing.T) {
	mockDB := &database.MockDB{}
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
	mockDB := &database.MockDB{}
	cryptoEngine, err := crypto.NewEngine("test-master-key-12345678901234567890")
	assert.NoError(t, err)

	server := createTestServer(mockDB, cryptoEngine)

	tests := []struct {
		name        string
		keyID       string
		aesKeyB64   string
		description string
		expectErr   bool
		setupMock   func()
	}{
		{
			name:        "valid key storage",
			keyID:       "cccccccccccc",
			aesKeyB64:   "MTIzNDU2Nzg5MDEyMzQ1Ng", // 16 bytes base64 encoded
			description: "Test key",
			expectErr:   false,
			setupMock: func() {
				mockDB.On("StoreKey", mock.AnythingOfType("*database.YubikeyKey")).Return(nil).Once()
			},
		},
		{
			name:        "invalid key ID length",
			keyID:       "ccc",
			aesKeyB64:   "MTIzNDU2Nzg5MDEyMzQ1Ng",
			description: "Test key",
			expectErr:   true,
			setupMock:   func() {},
		},
		{
			name:        "invalid modhex in key ID",
			keyID:       "ccccccccccXX",
			aesKeyB64:   "MTIzNDU2Nzg5MDEyMzQ1Ng",
			description: "Test key",
			expectErr:   true,
			setupMock:   func() {},
		},
		{
			name:        "invalid base64 AES key",
			keyID:       "cccccccccccc",
			aesKeyB64:   "invalid-base64!",
			description: "Test key",
			expectErr:   true,
			setupMock:   func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockDB.ExpectedCalls = nil
			tt.setupMock()

			err := server.StoreKey(tt.keyID, tt.aesKeyB64, tt.description)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestDecryptOTP_InvalidFormat(t *testing.T) {
	mockDB := &database.MockDB{}
	cryptoEngine, err := crypto.NewEngine("test-master-key-12345678901234567890")
	assert.NoError(t, err)

	server := createTestServer(mockDB, cryptoEngine)

	response, err := server.DecryptOTP("invalid-otp")
	assert.NoError(t, err)
	assert.Equal(t, "ERROR", response.Status)
	assert.Equal(t, "INVALID_OTP", response.ErrorCode)
}

func TestDecryptOTP_KeyNotFound(t *testing.T) {
	mockDB := &database.MockDB{}
	cryptoEngine, err := crypto.NewEngine("test-master-key-12345678901234567890")
	assert.NoError(t, err)

	server := createTestServer(mockDB, cryptoEngine)

	mockDB.On("GetKey", "cccccccccccc").Return(nil, sql.ErrNoRows).Once()

	response, err := server.DecryptOTP("ccccccccccccdefghijklnrtuvcbdefghijklnrtuvic")
	assert.NoError(t, err)
	assert.Equal(t, "ERROR", response.Status)
	assert.Equal(t, "KEY_NOT_FOUND", response.ErrorCode)

	mockDB.AssertExpectations(t)
}

func TestListKeys(t *testing.T) {
	mockDB := &database.MockDB{}
	cryptoEngine, err := crypto.NewEngine("test-master-key-12345678901234567890")
	assert.NoError(t, err)

	server := createTestServer(mockDB, cryptoEngine)

	expectedKeys := []*database.YubikeyKey{
		{
			KeyID:       "cccccccccccc",
			Description: "Test key 1",
			CreatedAt:   time.Now(),
		},
		{
			KeyID:       "dddddddddddd",
			Description: "Test key 2",
			CreatedAt:   time.Now(),
		},
	}

	mockDB.On("ListKeys").Return(expectedKeys, nil).Once()

	keys, err := server.ListKeys()
	assert.NoError(t, err)
	assert.Equal(t, expectedKeys, keys)

	mockDB.AssertExpectations(t)
}

func TestDeleteKey(t *testing.T) {
	mockDB := &database.MockDB{}
	cryptoEngine, err := crypto.NewEngine("test-master-key-12345678901234567890")
	assert.NoError(t, err)

	server := createTestServer(mockDB, cryptoEngine)

	mockDB.On("DeleteKey", "cccccccccccc").Return(nil).Once()

	err = server.DeleteKey("cccccccccccc")
	assert.NoError(t, err)

	mockDB.AssertExpectations(t)
}
