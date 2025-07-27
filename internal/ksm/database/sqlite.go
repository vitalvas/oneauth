package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/vitalvas/oneauth/internal/ksm/config"
	_ "modernc.org/sqlite" // SQLite driver
)

type SQLite struct {
	db *sql.DB
}

func NewSQLite(config *config.SQLiteConfig) (*SQLite, error) {
	dir := filepath.Dir(config.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	dsn := fmt.Sprintf("%s?_journal_mode=%s&_synchronous=%s",
		config.Path,
		config.JournalMode,
		config.Synchronous,
	)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	sqlite := &SQLite{db: db}
	if err := sqlite.createTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return sqlite, nil
}

func (s *SQLite) Close() error {
	return s.db.Close()
}

func (s *SQLite) createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS yubikey_keys (
		key_id TEXT PRIMARY KEY,
		aes_key_encrypted TEXT NOT NULL,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_used_at DATETIME,
		usage_count INTEGER DEFAULT 0,
		active BOOLEAN DEFAULT 1
	);

	CREATE INDEX IF NOT EXISTS idx_yubikey_keys_active ON yubikey_keys (active);
	CREATE INDEX IF NOT EXISTS idx_yubikey_keys_last_used ON yubikey_keys (last_used_at);

	CREATE TABLE IF NOT EXISTS yubikey_counters (
		key_id TEXT REFERENCES yubikey_keys(key_id) ON DELETE CASCADE,
		counter INTEGER NOT NULL,
		session_use INTEGER NOT NULL,
		timestamp_high INTEGER NOT NULL,
		timestamp_low INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		
		PRIMARY KEY (key_id, counter, session_use)
	);

	CREATE INDEX IF NOT EXISTS idx_yubikey_counters_key_counter ON yubikey_counters (key_id, counter DESC);
	`

	_, err := s.db.Exec(schema)
	return err
}

func (s *SQLite) StoreKey(key *YubikeyKey) error {
	query := `
		INSERT INTO yubikey_keys (key_id, aes_key_encrypted, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`

	now := time.Now()
	_, err := s.db.Exec(query, key.KeyID, key.AESKeyEncrypted, key.Description, now, now)
	return err
}

func (s *SQLite) GetKey(keyID string) (*YubikeyKey, error) {
	query := `
		SELECT key_id, aes_key_encrypted, description, created_at, updated_at, last_used_at, usage_count, active
		FROM yubikey_keys
		WHERE key_id = ? AND active = 1
	`

	rows, err := s.db.Query(query, keyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	return scanYubikeyKey(rows)
}

func (s *SQLite) ListKeys() ([]*YubikeyKey, error) {
	query := `
		SELECT key_id, aes_key_encrypted, description, created_at, updated_at, last_used_at, usage_count, active
		FROM yubikey_keys
		WHERE active = 1
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*YubikeyKey
	for rows.Next() {
		key, err := scanYubikeyKey(rows)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}

	return keys, rows.Err()
}

func (s *SQLite) DeleteKey(keyID string) error {
	query := `UPDATE yubikey_keys SET active = 0 WHERE key_id = ?`
	_, err := s.db.Exec(query, keyID)
	return err
}

func (s *SQLite) UpdateKeyUsage(keyID string) error {
	query := `
		UPDATE yubikey_keys 
		SET last_used_at = CURRENT_TIMESTAMP, usage_count = usage_count + 1, updated_at = CURRENT_TIMESTAMP
		WHERE key_id = ?
	`
	_, err := s.db.Exec(query, keyID)
	return err
}

func (s *SQLite) ValidateCounter(keyID string, counter, sessionUse int) error {
	query := `
		SELECT COUNT(*) FROM yubikey_counters
		WHERE key_id = ? AND counter >= ? AND session_use >= ?
	`

	var count int
	err := s.db.QueryRow(query, keyID, counter, sessionUse).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("replay attack detected")
	}

	return nil
}

func (s *SQLite) StoreCounter(counter *YubikeyCounter) error {
	query := `
		INSERT OR IGNORE INTO yubikey_counters (key_id, counter, session_use, timestamp_high, timestamp_low, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		counter.KeyID,
		counter.Counter,
		counter.SessionUse,
		counter.TimestampHigh,
		counter.TimestampLow,
		counter.CreatedAt,
	)
	return err
}

func (s *SQLite) HealthCheck() error {
	return s.db.Ping()
}
