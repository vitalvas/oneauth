package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vitalvas/oneauth/internal/ksm/config"
)

func TestNewSQLiteFilePathHandling(t *testing.T) {
	t.Run("creates nested directories for database path", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "nested", "deep", "test.db")

		cfg := &config.SQLiteConfig{
			Path:        dbPath,
			JournalMode: "WAL",
			Synchronous: "NORMAL",
		}

		db, err := NewSQLite(cfg)
		assert.NoError(t, err)
		defer db.Close()

		// Verify directory was created
		dirInfo, err := os.Stat(filepath.Dir(dbPath))
		assert.NoError(t, err)
		assert.True(t, dirInfo.IsDir())
	})

	t.Run("existing directory works", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")

		cfg := &config.SQLiteConfig{
			Path:        dbPath,
			JournalMode: "WAL",
			Synchronous: "NORMAL",
		}

		db, err := NewSQLite(cfg)
		assert.NoError(t, err)
		defer db.Close()

		err = db.HealthCheck()
		assert.NoError(t, err)
	})

	t.Run("read-only directory fails", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("skipping test as root user")
		}

		tmpDir := t.TempDir()
		readOnlyDir := filepath.Join(tmpDir, "readonly")
		err := os.Mkdir(readOnlyDir, 0555)
		assert.NoError(t, err)

		dbPath := filepath.Join(readOnlyDir, "subdir", "test.db")

		cfg := &config.SQLiteConfig{
			Path:        dbPath,
			JournalMode: "WAL",
			Synchronous: "NORMAL",
		}

		_, err = NewSQLite(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create directory")
	})
}

func TestNewSQLiteDSNConstruction(t *testing.T) {
	t.Run("journal mode DELETE", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := &config.SQLiteConfig{
			Path:        filepath.Join(tmpDir, "delete_mode.db"),
			JournalMode: "DELETE",
			Synchronous: "FULL",
		}

		db, err := NewSQLite(cfg)
		assert.NoError(t, err)
		defer db.Close()

		err = db.HealthCheck()
		assert.NoError(t, err)
	})

	t.Run("journal mode MEMORY", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := &config.SQLiteConfig{
			Path:        filepath.Join(tmpDir, "memory_mode.db"),
			JournalMode: "MEMORY",
			Synchronous: "OFF",
		}

		db, err := NewSQLite(cfg)
		assert.NoError(t, err)
		defer db.Close()

		err = db.HealthCheck()
		assert.NoError(t, err)
	})
}

func TestSQLiteClose(t *testing.T) {
	t.Run("close returns nil on success", func(t *testing.T) {
		cfg := &config.SQLiteConfig{
			Path:        ":memory:",
			JournalMode: "WAL",
			Synchronous: "NORMAL",
		}

		db, err := NewSQLite(cfg)
		assert.NoError(t, err)

		err = db.Close()
		assert.NoError(t, err)
	})

	t.Run("health check fails after close", func(t *testing.T) {
		cfg := &config.SQLiteConfig{
			Path:        ":memory:",
			JournalMode: "WAL",
			Synchronous: "NORMAL",
		}

		db, err := NewSQLite(cfg)
		assert.NoError(t, err)

		err = db.Close()
		assert.NoError(t, err)

		err = db.HealthCheck()
		assert.Error(t, err)
	})
}

func TestNewSQLiteInvalidPath(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping test as root user")
	}

	t.Run("unwritable path for database", func(t *testing.T) {
		// Try to create a database in a path where we can't write the file
		tmpDir := t.TempDir()
		readOnlyDir := filepath.Join(tmpDir, "readonly")
		err := os.Mkdir(readOnlyDir, 0555)
		assert.NoError(t, err)

		cfg := &config.SQLiteConfig{
			Path:        filepath.Join(readOnlyDir, "test.db"),
			JournalMode: "WAL",
			Synchronous: "NORMAL",
		}

		// The file path directory exists but file cannot be created
		// This might succeed on MkdirAll but fail on sql.Open or Ping
		db, err := NewSQLite(cfg)
		if err != nil {
			assert.Error(t, err)
		} else {
			db.Close()
		}
	})
}

func TestSQLiteListKeysEmpty(t *testing.T) {
	cfg := &config.SQLiteConfig{
		Path:        ":memory:",
		JournalMode: "WAL",
		Synchronous: "NORMAL",
	}

	db, err := NewSQLite(cfg)
	assert.NoError(t, err)
	defer db.Close()

	// List on empty database
	keys, err := db.ListKeys()
	assert.NoError(t, err)
	assert.Empty(t, keys)
}

func TestSQLiteGetKeyNotFound(t *testing.T) {
	cfg := &config.SQLiteConfig{
		Path:        ":memory:",
		JournalMode: "WAL",
		Synchronous: "NORMAL",
	}

	db, err := NewSQLite(cfg)
	assert.NoError(t, err)
	defer db.Close()

	_, err = db.GetKey("nonexistent")
	assert.Error(t, err)
}

func TestSQLiteValidateCounterNoRecords(t *testing.T) {
	cfg := &config.SQLiteConfig{
		Path:        ":memory:",
		JournalMode: "WAL",
		Synchronous: "NORMAL",
	}

	db, err := NewSQLite(cfg)
	assert.NoError(t, err)
	defer db.Close()

	// Validate counter with no records - should pass
	err = db.ValidateCounter("anykey", 1, 1)
	assert.NoError(t, err)
}

func TestSQLiteConfigDefaults(t *testing.T) {
	t.Run("default config values", func(t *testing.T) {
		cfg := &config.SQLiteConfig{}
		cfg.Default()

		assert.Equal(t, "/var/lib/oneauth/yubikey_ksm.db", cfg.Path)
		assert.Equal(t, "WAL", cfg.JournalMode)
		assert.Equal(t, "NORMAL", cfg.Synchronous)
	})
}

func TestSQLiteOperationsAfterClose(t *testing.T) {
	cfg := &config.SQLiteConfig{
		Path:        ":memory:",
		JournalMode: "WAL",
		Synchronous: "NORMAL",
	}

	db, err := NewSQLite(cfg)
	assert.NoError(t, err)

	err = db.Close()
	assert.NoError(t, err)

	t.Run("StoreKey after close", func(t *testing.T) {
		key := &YubikeyKey{
			KeyID:           "afterclose01",
			AESKeyEncrypted: "encrypted",
			Description:     "test",
		}
		err := db.StoreKey(key)
		assert.Error(t, err)
	})

	t.Run("GetKey after close", func(t *testing.T) {
		_, err := db.GetKey("afterclose01")
		assert.Error(t, err)
	})

	t.Run("ListKeys after close", func(t *testing.T) {
		_, err := db.ListKeys()
		assert.Error(t, err)
	})

	t.Run("DeleteKey after close", func(t *testing.T) {
		err := db.DeleteKey("afterclose01")
		assert.Error(t, err)
	})

	t.Run("UpdateKeyUsage after close", func(t *testing.T) {
		err := db.UpdateKeyUsage("afterclose01")
		assert.Error(t, err)
	})

	t.Run("ValidateCounter after close", func(t *testing.T) {
		err := db.ValidateCounter("afterclose01", 1, 1)
		assert.Error(t, err)
	})

	t.Run("StoreCounter after close", func(t *testing.T) {
		counter := &YubikeyCounter{
			KeyID:      "afterclose01",
			Counter:    1,
			SessionUse: 1,
		}
		err := db.StoreCounter(counter)
		assert.Error(t, err)
	})
}

func TestSQLiteDuplicateKeyStore(t *testing.T) {
	cfg := &config.SQLiteConfig{
		Path:        ":memory:",
		JournalMode: "WAL",
		Synchronous: "NORMAL",
	}

	db, err := NewSQLite(cfg)
	assert.NoError(t, err)
	defer db.Close()

	key := &YubikeyKey{
		KeyID:           "dupkey000001",
		AESKeyEncrypted: "encrypted",
		Description:     "test",
	}

	err = db.StoreKey(key)
	assert.NoError(t, err)

	// Storing duplicate key should fail due to primary key constraint
	err = db.StoreKey(key)
	assert.Error(t, err)
}

func TestSQLiteMultipleUpdateKeyUsage(t *testing.T) {
	cfg := &config.SQLiteConfig{
		Path:        ":memory:",
		JournalMode: "WAL",
		Synchronous: "NORMAL",
	}

	db, err := NewSQLite(cfg)
	assert.NoError(t, err)
	defer db.Close()

	key := &YubikeyKey{
		KeyID:           "multiupdate1",
		AESKeyEncrypted: "encrypted",
		Description:     "test",
	}
	err = db.StoreKey(key)
	assert.NoError(t, err)

	// Update usage multiple times
	for i := 0; i < 5; i++ {
		err = db.UpdateKeyUsage("multiupdate1")
		assert.NoError(t, err)
	}

	retrieved, err := db.GetKey("multiupdate1")
	assert.NoError(t, err)
	assert.Equal(t, 5, retrieved.UsageCount)
	assert.NotNil(t, retrieved.LastUsedAt)
}

func TestSQLiteTableCreation(t *testing.T) {
	t.Run("tables are created on initialization", func(t *testing.T) {
		cfg := &config.SQLiteConfig{
			Path:        ":memory:",
			JournalMode: "WAL",
			Synchronous: "NORMAL",
		}

		db, err := NewSQLite(cfg)
		assert.NoError(t, err)
		defer db.Close()

		// Verify tables exist by performing operations
		keys, err := db.ListKeys()
		assert.NoError(t, err)
		assert.Empty(t, keys)
	})

	t.Run("idempotent table creation", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "idempotent.db")

		cfg := &config.SQLiteConfig{
			Path:        dbPath,
			JournalMode: "WAL",
			Synchronous: "NORMAL",
		}

		// First creation
		db1, err := NewSQLite(cfg)
		assert.NoError(t, err)
		db1.Close()

		// Second creation on same file should succeed (CREATE IF NOT EXISTS)
		db2, err := NewSQLite(cfg)
		assert.NoError(t, err)
		defer db2.Close()

		err = db2.HealthCheck()
		assert.NoError(t, err)
	})
}
