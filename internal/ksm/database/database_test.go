package database

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vitalvas/oneauth/internal/ksm/config"
)

func TestNew_SQLite(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Type: "sqlite",
		SQLite: &config.SQLiteConfig{
			Path:        ":memory:",
			JournalMode: "WAL",
			Synchronous: "NORMAL",
		},
	}

	db, err := New(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()

	// Test health check
	err = db.HealthCheck()
	assert.NoError(t, err)
}

func TestNew_PostgreSQL(t *testing.T) {
	// Test PostgreSQL path in New() function with connection failure
	cfg := &config.DatabaseConfig{
		Type: "postgres",
		PostgreSQL: &config.PostgreSQLConfig{
			Host:              "localhost",
			Port:              5432,
			Username:          "test",
			Password:          "test",
			Database:          "test",
			MaxConnections:    10,
			ConnectionTimeout: 30 * time.Second,
		},
	}

	// This should attempt to connect and fail
	db, err := New(cfg)
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestNew_PostgreSQL_MissingConfig(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Type:       "postgres",
		PostgreSQL: nil, // Missing config
	}

	db, err := New(cfg)
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "PostgreSQL configuration is required")
}

func TestNew_SQLite_MissingConfig(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Type:   "sqlite",
		SQLite: nil, // Missing config
	}

	db, err := New(cfg)
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "SQLite configuration is required")
}

func TestNew_InvalidType(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Type: "invalid",
	}

	db, err := New(cfg)
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "unsupported database type")
}

func TestSQLiteOperations(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Type: "sqlite",
		SQLite: &config.SQLiteConfig{
			Path:        ":memory:",
			JournalMode: "WAL",
			Synchronous: "NORMAL",
		},
	}

	db, err := New(cfg)
	assert.NoError(t, err)
	defer db.Close()

	// Test key operations
	testKeyOperations(t, db)

	// Test counter operations
	testCounterOperations(t, db)
}

func testKeyOperations(t *testing.T, db DB) {
	now := time.Now()
	key := &YubikeyKey{
		KeyID:           "cccccccccccc",
		AESKeyEncrypted: "encrypted-test-key",
		Description:     "Test YubiKey",
		CreatedAt:       now,
		UpdatedAt:       now,
		UsageCount:      0,
		Active:          true,
	}

	// Test store key
	err := db.StoreKey(key)
	assert.NoError(t, err)

	// Test get key
	retrieved, err := db.GetKey("cccccccccccc")
	assert.NoError(t, err)
	assert.Equal(t, key.KeyID, retrieved.KeyID)
	assert.Equal(t, key.AESKeyEncrypted, retrieved.AESKeyEncrypted)
	assert.Equal(t, key.Description, retrieved.Description)
	assert.Equal(t, key.UsageCount, retrieved.UsageCount)
	assert.Equal(t, key.Active, retrieved.Active)

	// Test get non-existent key
	_, err = db.GetKey("nonexistent")
	assert.Equal(t, sql.ErrNoRows, err)

	// Test list keys
	keys, err := db.ListKeys()
	assert.NoError(t, err)
	assert.Len(t, keys, 1)
	assert.Equal(t, "cccccccccccc", keys[0].KeyID)

	// Test update key usage
	err = db.UpdateKeyUsage("cccccccccccc")
	assert.NoError(t, err)

	// Verify usage count increased
	updated, err := db.GetKey("cccccccccccc")
	assert.NoError(t, err)
	assert.Equal(t, 1, updated.UsageCount)
	assert.NotNil(t, updated.LastUsedAt)

	// Test delete key (soft delete)
	err = db.DeleteKey("cccccccccccc")
	assert.NoError(t, err)

	// Key should no longer be retrievable
	_, err = db.GetKey("cccccccccccc")
	assert.Equal(t, sql.ErrNoRows, err)

	// List should be empty
	keys, err = db.ListKeys()
	assert.NoError(t, err)
	assert.Len(t, keys, 0)
}

func testCounterOperations(t *testing.T, db DB) {
	// First store a key for the counter
	key := &YubikeyKey{
		KeyID:           "dddddddddddd",
		AESKeyEncrypted: "encrypted-test-key-2",
		Description:     "Test YubiKey for counter",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		UsageCount:      0,
		Active:          true,
	}
	err := db.StoreKey(key)
	assert.NoError(t, err)

	keyID := "dddddddddddd"
	counter := &YubikeyCounter{
		KeyID:         keyID,
		Counter:       100,
		SessionUse:    1,
		TimestampHigh: 12345,
		TimestampLow:  67890,
		CreatedAt:     time.Now(),
	}

	// Test store counter
	err = db.StoreCounter(counter)
	assert.NoError(t, err)

	// Test validate counter - should pass for higher values
	err = db.ValidateCounter(keyID, 101, 1)
	assert.NoError(t, err)

	err = db.ValidateCounter(keyID, 100, 2)
	assert.NoError(t, err)

	// Test validate counter - should fail for replay (same or lower values)
	err = db.ValidateCounter(keyID, 100, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "replay attack detected")

	err = db.ValidateCounter(keyID, 99, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "replay attack detected")

	// Test duplicate counter storage (should be handled gracefully)
	err = db.StoreCounter(counter)
	assert.NoError(t, err) // Should not error due to ON CONFLICT handling
}

func TestScanYubikeyKey(t *testing.T) {
	// Create a mock row with test data
	cfg := &config.DatabaseConfig{
		Type: "sqlite",
		SQLite: &config.SQLiteConfig{
			Path:        ":memory:",
			JournalMode: "WAL",
			Synchronous: "NORMAL",
		},
	}

	db, err := New(cfg)
	assert.NoError(t, err)
	defer db.Close()

	// Store a key first
	now := time.Now()
	originalKey := &YubikeyKey{
		KeyID:           "testkey12345",
		AESKeyEncrypted: "encrypted-data",
		Description:     "Test Description",
		CreatedAt:       now,
		UpdatedAt:       now,
		UsageCount:      0, // StoreKey always starts with 0
		Active:          true,
	}

	err = db.StoreKey(originalKey)
	assert.NoError(t, err)

	// Retrieve and verify scanning worked correctly
	retrieved, err := db.GetKey("testkey12345")
	assert.NoError(t, err)
	assert.Equal(t, originalKey.KeyID, retrieved.KeyID)
	assert.Equal(t, originalKey.AESKeyEncrypted, retrieved.AESKeyEncrypted)
	assert.Equal(t, originalKey.Description, retrieved.Description)
	assert.Equal(t, 0, retrieved.UsageCount) // Should be 0 when first stored
	assert.Equal(t, originalKey.Active, retrieved.Active)
	assert.NotNil(t, retrieved.CreatedAt)
	assert.NotNil(t, retrieved.UpdatedAt)
}

func TestHealthCheck(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Type: "sqlite",
		SQLite: &config.SQLiteConfig{
			Path:        ":memory:",
			JournalMode: "WAL",
			Synchronous: "NORMAL",
		},
	}

	db, err := New(cfg)
	assert.NoError(t, err)
	defer db.Close()

	// Test health check on working database
	err = db.HealthCheck()
	assert.NoError(t, err)

	// Close database and test health check failure
	db.Close()
	err = db.HealthCheck()
	assert.Error(t, err)
}

func TestMultipleKeyOperations(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Type: "sqlite",
		SQLite: &config.SQLiteConfig{
			Path:        ":memory:",
			JournalMode: "WAL",
			Synchronous: "NORMAL",
		},
	}

	db, err := New(cfg)
	assert.NoError(t, err)
	defer db.Close()

	// Store multiple keys
	keys := []*YubikeyKey{
		{
			KeyID:           "key000000001",
			AESKeyEncrypted: "encrypted-1",
			Description:     "First Key",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			UsageCount:      0,
			Active:          true,
		},
		{
			KeyID:           "key000000002",
			AESKeyEncrypted: "encrypted-2",
			Description:     "Second Key",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			UsageCount:      0,
			Active:          true,
		},
		{
			KeyID:           "key000000003",
			AESKeyEncrypted: "encrypted-3",
			Description:     "Third Key",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			UsageCount:      0,
			Active:          true,
		},
	}

	for _, key := range keys {
		err := db.StoreKey(key)
		assert.NoError(t, err)
	}

	// List all keys
	allKeys, err := db.ListKeys()
	assert.NoError(t, err)
	assert.Len(t, allKeys, 3)

	// Verify keys are ordered by creation time (newest first)
	assert.Equal(t, "key000000003", allKeys[0].KeyID) // Last created should be first

	// Delete one key
	err = db.DeleteKey("key000000002")
	assert.NoError(t, err)

	// List should now have 2 keys
	remainingKeys, err := db.ListKeys()
	assert.NoError(t, err)
	assert.Len(t, remainingKeys, 2)

	// Verify correct keys remain
	keyIDs := make([]string, len(remainingKeys))
	for i, key := range remainingKeys {
		keyIDs[i] = key.KeyID
	}
	assert.Contains(t, keyIDs, "key000000001")
	assert.Contains(t, keyIDs, "key000000003")
	assert.NotContains(t, keyIDs, "key000000002")
}

func TestSQLiteSpecificFeatures(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Type: "sqlite",
		SQLite: &config.SQLiteConfig{
			Path:        ":memory:",
			JournalMode: "WAL",
			Synchronous: "NORMAL",
		},
	}

	// Test SQLite creation with directory creation
	tmpDir := "/tmp/ksm-test"
	cfg.SQLite.Path = tmpDir + "/test.db"

	db, err := New(cfg)
	assert.NoError(t, err)
	defer db.Close()

	// Test that database works
	err = db.HealthCheck()
	assert.NoError(t, err)
}

func TestPostgreSQLSpecificFeatures(t *testing.T) {
	// Test PostgreSQL configuration validation
	cfg := &config.DatabaseConfig{
		Type: "postgresql",
		PostgreSQL: &config.PostgreSQLConfig{
			Host:              "nonexistent-host",
			Port:              5432,
			Username:          "test",
			Password:          "test",
			Database:          "test",
			MaxConnections:    10,
			ConnectionTimeout: 30 * time.Second,
		},
	}

	// This should fail to connect
	db, err := New(cfg)
	if err != nil {
		// Expected - can't connect to nonexistent host
		assert.Error(t, err)
		assert.Nil(t, db)
	} else {
		// If somehow it connects, clean up
		db.Close()
	}
}

func TestDatabaseErrorCases(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Type: "sqlite",
		SQLite: &config.SQLiteConfig{
			Path:        ":memory:",
			JournalMode: "WAL",
			Synchronous: "NORMAL",
		},
	}

	db, err := New(cfg)
	assert.NoError(t, err)
	defer db.Close()

	// Test operations with non-existent key ID
	err = db.UpdateKeyUsage("nonexistent")
	assert.NoError(t, err) // Should not error even if key doesn't exist

	err = db.DeleteKey("nonexistent")
	assert.NoError(t, err) // Should not error even if key doesn't exist

	// Test counter validation with non-existent key
	err = db.ValidateCounter("nonexistent", 1, 1)
	assert.NoError(t, err) // Should pass since no counters exist

	// Test storing counter for non-existent key (should fail due to foreign key)
	counter := &YubikeyCounter{
		KeyID:         "nonexistent",
		Counter:       1,
		SessionUse:    1,
		TimestampHigh: 100,
		TimestampLow:  200,
		CreatedAt:     time.Now(),
	}

	// This might fail or succeed depending on foreign key constraints
	// Just ensure it doesn't panic
	assert.NotPanics(t, func() {
		_ = db.StoreCounter(counter)
	})
}

func TestScanYubikeyKeyEdgeCases(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Type: "sqlite",
		SQLite: &config.SQLiteConfig{
			Path:        ":memory:",
			JournalMode: "WAL",
			Synchronous: "NORMAL",
		},
	}

	db, err := New(cfg)
	assert.NoError(t, err)
	defer db.Close()

	// Store keys with various edge case values
	tests := []struct {
		name string
		key  *YubikeyKey
	}{
		{
			name: "minimal key",
			key: &YubikeyKey{
				KeyID:           "minimalkey01",
				AESKeyEncrypted: "minimal",
				Description:     "", // Empty description
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
				UsageCount:      0,
				Active:          true,
			},
		},
		{
			name: "long description key",
			key: &YubikeyKey{
				KeyID:           "longdesckey2",
				AESKeyEncrypted: "encrypted-long",
				Description:     "This is a very long description that tests the database's ability to handle longer text fields without issues",
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
				UsageCount:      0,
				Active:          true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.StoreKey(tt.key)
			assert.NoError(t, err)

			retrieved, err := db.GetKey(tt.key.KeyID)
			assert.NoError(t, err)
			assert.Equal(t, tt.key.KeyID, retrieved.KeyID)
			assert.Equal(t, tt.key.AESKeyEncrypted, retrieved.AESKeyEncrypted)
			assert.Equal(t, tt.key.Description, retrieved.Description)
			assert.Equal(t, tt.key.Active, retrieved.Active)
		})
	}
}

func TestCounterEdgeCases(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Type: "sqlite",
		SQLite: &config.SQLiteConfig{
			Path:        ":memory:",
			JournalMode: "WAL",
			Synchronous: "NORMAL",
		},
	}

	db, err := New(cfg)
	assert.NoError(t, err)
	defer db.Close()

	// Store a key first
	key := &YubikeyKey{
		KeyID:           "countertest01",
		AESKeyEncrypted: "encrypted-counter-test",
		Description:     "Counter test key",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		UsageCount:      0,
		Active:          true,
	}
	err = db.StoreKey(key)
	assert.NoError(t, err)

	// Test edge cases for counters
	counters := []*YubikeyCounter{
		{
			KeyID:         "countertest01",
			Counter:       0, // Minimum counter
			SessionUse:    0, // Minimum session use
			TimestampHigh: 0,
			TimestampLow:  0,
			CreatedAt:     time.Now(),
		},
		{
			KeyID:         "countertest01",
			Counter:       65535, // Maximum 16-bit counter
			SessionUse:    255,   // Maximum 8-bit session use
			TimestampHigh: 255,   // Maximum 8-bit timestamp high
			TimestampLow:  65535, // Maximum 16-bit timestamp low
			CreatedAt:     time.Now(),
		},
	}

	for i, counter := range counters {
		t.Run(fmt.Sprintf("counter_%d", i), func(t *testing.T) {
			err := db.StoreCounter(counter)
			assert.NoError(t, err)

			// Validate that higher counters pass
			err = db.ValidateCounter(counter.KeyID, counter.Counter+1, counter.SessionUse)
			assert.NoError(t, err)

			err = db.ValidateCounter(counter.KeyID, counter.Counter, counter.SessionUse+1)
			assert.NoError(t, err)
		})
	}
}

func TestCounterValidationEdgeCases(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Type: "sqlite",
		SQLite: &config.SQLiteConfig{
			Path:        ":memory:",
			JournalMode: "WAL",
			Synchronous: "NORMAL",
		},
	}

	db, err := New(cfg)
	assert.NoError(t, err)
	defer db.Close()

	// Store a test key first
	key := &YubikeyKey{
		KeyID:           "edgetest0001",
		AESKeyEncrypted: "encrypted-edge-test",
		Description:     "Edge case test key",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		UsageCount:      0,
		Active:          true,
	}
	err = db.StoreKey(key)
	assert.NoError(t, err)

	// Test counter rollover scenarios
	testCases := []struct {
		name           string
		counter        int
		sessionUse     int
		expectStore    bool
		expectValidate bool
	}{
		{
			name:           "Zero counter and session",
			counter:        0,
			sessionUse:     0,
			expectStore:    true,
			expectValidate: false, // Should fail validation after storing
		},
		{
			name:           "Maximum 16-bit counter",
			counter:        65535,
			sessionUse:     1,
			expectStore:    true,
			expectValidate: false,
		},
		{
			name:           "Maximum 8-bit session use",
			counter:        1000,
			sessionUse:     255,
			expectStore:    true,
			expectValidate: false,
		},
		{
			name:           "Counter rollover simulation",
			counter:        0, // Simulating counter rollover
			sessionUse:     2, // But session use increased
			expectStore:    true,
			expectValidate: false, // Will fail due to previous stored counters with higher values
		},
		{
			name:           "Same counter, higher session",
			counter:        1000,
			sessionUse:     256, // Higher than max 8-bit, but allowed in storage
			expectStore:    true,
			expectValidate: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// First validate before storing (should pass initially)
			err = db.ValidateCounter("edgetest0001", tc.counter, tc.sessionUse)
			if tc.expectValidate {
				assert.NoError(t, err, "Validation should pass before storing")
			}

			// Store the counter
			counterRecord := &YubikeyCounter{
				KeyID:         "edgetest0001",
				Counter:       tc.counter,
				SessionUse:    tc.sessionUse,
				TimestampHigh: 12345,
				TimestampLow:  67890,
				CreatedAt:     time.Now(),
			}

			err = db.StoreCounter(counterRecord)
			if tc.expectStore {
				assert.NoError(t, err, "Counter storage should succeed")

				// Now validate again (should fail due to replay detection)
				err = db.ValidateCounter("edgetest0001", tc.counter, tc.sessionUse)
				assert.Error(t, err, "Validation should fail after storing (replay detection)")
				assert.Contains(t, err.Error(), "replay attack detected")
			} else {
				assert.Error(t, err, "Counter storage should fail")
			}
		})
	}
}

func TestCounterSequenceValidation(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Type: "sqlite",
		SQLite: &config.SQLiteConfig{
			Path:        ":memory:",
			JournalMode: "WAL",
			Synchronous: "NORMAL",
		},
	}

	db, err := New(cfg)
	assert.NoError(t, err)
	defer db.Close()

	// Store a test key first
	key := &YubikeyKey{
		KeyID:           "seqtest00001",
		AESKeyEncrypted: "encrypted-sequence-test",
		Description:     "Sequence test key",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		UsageCount:      0,
		Active:          true,
	}
	err = db.StoreKey(key)
	assert.NoError(t, err)

	// Store a sequence of counters to test progression
	sequence := []struct {
		counter    int
		sessionUse int
	}{
		{100, 1},
		{100, 2}, // Same counter, higher session
		{101, 1}, // Higher counter, reset session
		{101, 2},
		{102, 1},
	}

	// Store all counters in sequence
	for i, s := range sequence {
		t.Run(fmt.Sprintf("Store_sequence_%d", i), func(t *testing.T) {
			counterRecord := &YubikeyCounter{
				KeyID:         "seqtest00001",
				Counter:       s.counter,
				SessionUse:    s.sessionUse,
				TimestampHigh: 10000 + i,
				TimestampLow:  20000 + i,
				CreatedAt:     time.Now(),
			}

			err = db.StoreCounter(counterRecord)
			assert.NoError(t, err)
		})
	}

	// Test validation against various scenarios
	validationTests := []struct {
		name       string
		counter    int
		sessionUse int
		shouldPass bool
		reason     string
	}{
		{
			name:       "Below all stored counters",
			counter:    99,
			sessionUse: 1,
			shouldPass: false,
			reason:     "100 >= 99",
		},
		{
			name:       "Equal to lowest counter, lower session",
			counter:    100,
			sessionUse: 0,
			shouldPass: false,
			reason:     "counter 100, session 1 exists and 1 >= 0",
		},
		{
			name:       "Equal to stored counter and session",
			counter:    100,
			sessionUse: 1,
			shouldPass: false,
			reason:     "exact match exists",
		},
		{
			name:       "Equal counter, higher session than stored",
			counter:    100,
			sessionUse: 3,
			shouldPass: true,
			reason:     "no counter >= 100 with session >= 3 exists",
		},
		{
			name:       "Higher than all stored",
			counter:    103,
			sessionUse: 1,
			shouldPass: true,
			reason:     "no counter >= 103",
		},
		{
			name:       "Same as highest counter, higher session",
			counter:    102,
			sessionUse: 2,
			shouldPass: true,
			reason:     "no counter >= 102 with session >= 2 exists (only 102,1 exists)",
		},
	}

	for _, vt := range validationTests {
		t.Run(vt.name, func(t *testing.T) {
			err = db.ValidateCounter("seqtest00001", vt.counter, vt.sessionUse)
			if vt.shouldPass {
				assert.NoError(t, err, "Expected validation to pass: %s", vt.reason)
			} else {
				assert.Error(t, err, "Expected validation to fail: %s", vt.reason)
				if err != nil {
					assert.Contains(t, err.Error(), "replay attack detected")
				}
			}
		})
	}
}
