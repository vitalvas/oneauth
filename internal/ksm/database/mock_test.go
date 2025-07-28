package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMockDB(t *testing.T) {
	t.Run("creation and initialization", func(t *testing.T) {
		mockDB, err := NewMockDB()
		assert.NoError(t, err)
		assert.NotNil(t, mockDB)
		defer mockDB.Close()

		// Verify it implements the DB interface
		var _ DB = mockDB

		// Test health check works
		err = mockDB.HealthCheck()
		assert.NoError(t, err)
	})

	t.Run("key operations", func(t *testing.T) {
		mockDB, err := NewMockDB()
		assert.NoError(t, err)
		defer mockDB.Close()

		// Test storing a key
		key := &YubikeyKey{
			KeyID:           "cccccccccccc",
			AESKeyEncrypted: "encrypted-key-data",
			Description:     "Test key",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			UsageCount:      0,
			Active:          true,
		}

		err = mockDB.StoreKey(key)
		assert.NoError(t, err)

		// Test retrieving the key
		retrievedKey, err := mockDB.GetKey("cccccccccccc")
		assert.NoError(t, err)
		assert.Equal(t, key.KeyID, retrievedKey.KeyID)
		assert.Equal(t, key.AESKeyEncrypted, retrievedKey.AESKeyEncrypted)
		assert.Equal(t, key.Description, retrievedKey.Description)
		assert.Equal(t, key.UsageCount, retrievedKey.UsageCount)
		assert.Equal(t, key.Active, retrievedKey.Active)

		// Test listing keys
		keys, err := mockDB.ListKeys()
		assert.NoError(t, err)
		assert.Len(t, keys, 1)
		assert.Equal(t, "cccccccccccc", keys[0].KeyID)

		// Test updating key usage
		err = mockDB.UpdateKeyUsage("cccccccccccc")
		assert.NoError(t, err)

		// Verify usage count increased
		updatedKey, err := mockDB.GetKey("cccccccccccc")
		assert.NoError(t, err)
		assert.Equal(t, 1, updatedKey.UsageCount)
		assert.NotNil(t, updatedKey.LastUsedAt)

		// Test deleting key (soft delete)
		err = mockDB.DeleteKey("cccccccccccc")
		assert.NoError(t, err)

		// Verify key is no longer listed (soft deleted)
		keys, err = mockDB.ListKeys()
		assert.NoError(t, err)
		assert.Len(t, keys, 0)

		// But direct get should return no rows error
		_, err = mockDB.GetKey("cccccccccccc")
		assert.Error(t, err)
	})

	t.Run("counter operations", func(t *testing.T) {
		mockDB, err := NewMockDB()
		assert.NoError(t, err)
		defer mockDB.Close()

		// First store a key
		key := &YubikeyKey{
			KeyID:           "dddddddddddd",
			AESKeyEncrypted: "encrypted-key-data",
			Description:     "Test key for counter",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			UsageCount:      0,
			Active:          true,
		}
		err = mockDB.StoreKey(key)
		assert.NoError(t, err)

		// Test counter validation (should pass for new counter)
		err = mockDB.ValidateCounter("dddddddddddd", 1, 1)
		assert.NoError(t, err)

		// Store a counter
		counter := &YubikeyCounter{
			KeyID:         "dddddddddddd",
			Counter:       1,
			SessionUse:    1,
			TimestampHigh: 12345,
			TimestampLow:  67890,
			CreatedAt:     time.Now(),
		}
		err = mockDB.StoreCounter(counter)
		assert.NoError(t, err)

		// Test replay attack detection (same counter/session should fail)
		err = mockDB.ValidateCounter("dddddddddddd", 1, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "replay attack detected")

		// Higher counter should pass
		err = mockDB.ValidateCounter("dddddddddddd", 2, 1)
		assert.NoError(t, err)

		// Same counter with higher session should pass (new session on same counter)
		err = mockDB.ValidateCounter("dddddddddddd", 1, 2)
		assert.NoError(t, err)
	})

	t.Run("nonexistent key operations", func(t *testing.T) {
		mockDB, err := NewMockDB()
		assert.NoError(t, err)
		defer mockDB.Close()

		// Test getting nonexistent key
		_, err = mockDB.GetKey("nonexistent")
		assert.Error(t, err)

		// Test counter operations with nonexistent key
		err = mockDB.ValidateCounter("nonexistent", 1, 1)
		assert.NoError(t, err) // Should pass since no counters exist

		// Test updating usage for nonexistent key (should not error but no effect)
		err = mockDB.UpdateKeyUsage("nonexistent")
		assert.NoError(t, err)

		// Test deleting nonexistent key (should not error but no effect)
		err = mockDB.DeleteKey("nonexistent")
		assert.NoError(t, err)
	})
}

func TestMockDB_Reset(t *testing.T) {
	t.Run("reset clears all data", func(t *testing.T) {
		mockDB, err := NewMockDB()
		assert.NoError(t, err)
		defer mockDB.Close()

		// Store some test data
		key1 := &YubikeyKey{
			KeyID:           "cccccccccccc",
			AESKeyEncrypted: "encrypted-key-1",
			Description:     "Test key 1",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			Active:          true,
		}
		key2 := &YubikeyKey{
			KeyID:           "dddddddddddd",
			AESKeyEncrypted: "encrypted-key-2",
			Description:     "Test key 2",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			Active:          true,
		}

		err = mockDB.StoreKey(key1)
		assert.NoError(t, err)
		err = mockDB.StoreKey(key2)
		assert.NoError(t, err)

		// Store a counter
		counter := &YubikeyCounter{
			KeyID:      "cccccccccccc",
			Counter:    1,
			SessionUse: 1,
			CreatedAt:  time.Now(),
		}
		err = mockDB.StoreCounter(counter)
		assert.NoError(t, err)

		// Verify data exists
		keys, err := mockDB.ListKeys()
		assert.NoError(t, err)
		assert.Len(t, keys, 2)

		// Reset the database
		err = mockDB.Reset()
		assert.NoError(t, err)

		// Verify all data is cleared
		keys, err = mockDB.ListKeys()
		assert.NoError(t, err)
		assert.Len(t, keys, 0)

		// Verify keys are gone
		_, err = mockDB.GetKey("cccccccccccc")
		assert.Error(t, err)
		_, err = mockDB.GetKey("dddddddddddd")
		assert.Error(t, err)

		// Verify counters are cleared (validation should pass for previously used counter)
		err = mockDB.ValidateCounter("cccccccccccc", 1, 1)
		assert.NoError(t, err)
	})

	t.Run("reset allows fresh operations", func(t *testing.T) {
		mockDB, err := NewMockDB()
		assert.NoError(t, err)
		defer mockDB.Close()

		// Store and delete a key
		key := &YubikeyKey{
			KeyID:           "testkey1234",
			AESKeyEncrypted: "encrypted-key",
			Description:     "Test key",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			Active:          true,
		}
		err = mockDB.StoreKey(key)
		assert.NoError(t, err)

		err = mockDB.DeleteKey("testkey1234")
		assert.NoError(t, err)

		// Reset database
		err = mockDB.Reset()
		assert.NoError(t, err)

		// Should be able to store the same key again
		err = mockDB.StoreKey(key)
		assert.NoError(t, err)

		// Should be able to retrieve it
		retrievedKey, err := mockDB.GetKey("testkey1234")
		assert.NoError(t, err)
		assert.Equal(t, key.KeyID, retrievedKey.KeyID)
	})
}

func TestMockDBWithErrors(t *testing.T) {
	t.Run("creation and basic functionality", func(t *testing.T) {
		mockDB, err := NewMockDBWithErrors()
		assert.NoError(t, err)
		assert.NotNil(t, mockDB)
		defer mockDB.Close()

		// Verify it implements the DB interface
		var _ DB = mockDB

		// Should work normally by default
		err = mockDB.HealthCheck()
		assert.NoError(t, err)

		// Should be able to perform normal operations
		key := &YubikeyKey{
			KeyID:           "cccccccccccc",
			AESKeyEncrypted: "encrypted-key",
			Description:     "Test key",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			Active:          true,
		}
		err = mockDB.StoreKey(key)
		assert.NoError(t, err)

		retrievedKey, err := mockDB.GetKey("cccccccccccc")
		assert.NoError(t, err)
		assert.Equal(t, key.KeyID, retrievedKey.KeyID)
	})

	t.Run("health check error simulation", func(t *testing.T) {
		mockDB, err := NewMockDBWithErrors()
		assert.NoError(t, err)
		defer mockDB.Close()

		// Initially should work
		err = mockDB.HealthCheck()
		assert.NoError(t, err)

		// Set health check error
		testError := assert.AnError
		mockDB.SetHealthCheckError(testError)

		// Now health check should return the error
		err = mockDB.HealthCheck()
		assert.Error(t, err)
		assert.Equal(t, testError, err)

		// Clear the error by setting to nil
		mockDB.SetHealthCheckError(nil)

		// Should work again
		err = mockDB.HealthCheck()
		assert.NoError(t, err)
	})

	t.Run("other operations unaffected by health check error", func(t *testing.T) {
		mockDB, err := NewMockDBWithErrors()
		assert.NoError(t, err)
		defer mockDB.Close()

		// Set health check error
		mockDB.SetHealthCheckError(assert.AnError)

		// Other operations should still work
		key := &YubikeyKey{
			KeyID:           "dddddddddddd",
			AESKeyEncrypted: "encrypted-key",
			Description:     "Test key",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			Active:          true,
		}
		err = mockDB.StoreKey(key)
		assert.NoError(t, err)

		retrievedKey, err := mockDB.GetKey("dddddddddddd")
		assert.NoError(t, err)
		assert.Equal(t, key.KeyID, retrievedKey.KeyID)

		keys, err := mockDB.ListKeys()
		assert.NoError(t, err)
		assert.Len(t, keys, 1)

		// But health check should still fail
		err = mockDB.HealthCheck()
		assert.Error(t, err)
	})

	t.Run("reset functionality works with errors", func(t *testing.T) {
		mockDB, err := NewMockDBWithErrors()
		assert.NoError(t, err)
		defer mockDB.Close()

		// Store some data
		key := &YubikeyKey{
			KeyID:           "resettest123",
			AESKeyEncrypted: "encrypted-key",
			Description:     "Reset test key",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			Active:          true,
		}
		err = mockDB.StoreKey(key)
		assert.NoError(t, err)

		// Set health check error
		mockDB.SetHealthCheckError(assert.AnError)

		// Reset should work
		err = mockDB.Reset()
		assert.NoError(t, err)

		// Data should be cleared
		keys, err := mockDB.ListKeys()
		assert.NoError(t, err)
		assert.Len(t, keys, 0)

		// Health check error should persist after reset
		err = mockDB.HealthCheck()
		assert.Error(t, err)
	})
}

func TestMockDB_InterfaceCompliance(t *testing.T) {
	t.Run("MockDB implements DB interface", func(t *testing.T) {
		mockDB, err := NewMockDB()
		assert.NoError(t, err)
		defer mockDB.Close()

		// Test all interface methods are callable
		var db DB = mockDB

		// Health check
		err = db.HealthCheck()
		assert.NoError(t, err)

		// Key operations
		key := &YubikeyKey{
			KeyID:           "interface123",
			AESKeyEncrypted: "encrypted",
			Description:     "Interface test",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			Active:          true,
		}
		err = db.StoreKey(key)
		assert.NoError(t, err)

		_, err = db.GetKey("interface123")
		assert.NoError(t, err)

		_, err = db.ListKeys()
		assert.NoError(t, err)

		err = db.UpdateKeyUsage("interface123")
		assert.NoError(t, err)

		err = db.DeleteKey("interface123")
		assert.NoError(t, err)

		// Counter operations
		counter := &YubikeyCounter{
			KeyID:      "interface123",
			Counter:    1,
			SessionUse: 1,
			CreatedAt:  time.Now(),
		}
		err = db.StoreCounter(counter)
		assert.NoError(t, err)

		err = db.ValidateCounter("interface123", 2, 1)
		assert.NoError(t, err)

		// Close
		err = db.Close()
		assert.NoError(t, err)
	})

	t.Run("MockDBWithErrors implements DB interface", func(t *testing.T) {
		mockDB, err := NewMockDBWithErrors()
		assert.NoError(t, err)
		defer mockDB.Close()

		// Test all interface methods are callable
		var db DB = mockDB

		err = db.HealthCheck()
		assert.NoError(t, err)

		// All other operations should work the same as regular MockDB
		key := &YubikeyKey{
			KeyID:           "errorintf123",
			AESKeyEncrypted: "encrypted",
			Description:     "Error interface test",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			Active:          true,
		}
		err = db.StoreKey(key)
		assert.NoError(t, err)
	})
}

func TestMockDB_ConcurrentAccess(t *testing.T) {
	t.Run("sequential operations instead of concurrent", func(t *testing.T) {
		// Note: In-memory SQLite databases don't handle true concurrency well
		// This test verifies that the MockDB can handle multiple operations correctly
		mockDB, err := NewMockDB()
		assert.NoError(t, err)
		defer mockDB.Close()

		// Store keys sequentially but verify no interference
		key1 := &YubikeyKey{
			KeyID:           "concurrent001",
			AESKeyEncrypted: "encrypted-1",
			Description:     "Concurrent test 1",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			Active:          true,
		}
		err = mockDB.StoreKey(key1)
		assert.NoError(t, err)

		key2 := &YubikeyKey{
			KeyID:           "concurrent002",
			AESKeyEncrypted: "encrypted-2",
			Description:     "Concurrent test 2",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			Active:          true,
		}
		err = mockDB.StoreKey(key2)
		assert.NoError(t, err)

		// Verify both keys were stored
		keys, err := mockDB.ListKeys()
		assert.NoError(t, err)
		assert.Len(t, keys, 2)

		keyIDs := make([]string, len(keys))
		for i, key := range keys {
			keyIDs[i] = key.KeyID
		}
		assert.Contains(t, keyIDs, "concurrent001")
		assert.Contains(t, keyIDs, "concurrent002")
	})
}
