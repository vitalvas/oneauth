package crypto

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
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
		longKey := string(make([]byte, 1000)) // 1000 null bytes
		for i := range longKey {
			longKey = longKey[:i] + "a" + longKey[i+1:]
		}

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
	longKeyID := string(make([]byte, 1000))
	for i := range longKeyID {
		longKeyID = longKeyID[:i] + "c" + longKeyID[i+1:]
	}

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
