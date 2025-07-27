package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	"github.com/vitalvas/oneauth/internal/ksm/config"
)

type PostgreSQL struct {
	db *sql.DB
}

func NewPostgreSQL(config *config.PostgreSQLConfig) (*PostgreSQL, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(config.MaxConnections)
	db.SetConnMaxLifetime(config.ConnectionTimeout)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	pg := &PostgreSQL{db: db}
	if err := pg.createTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return pg, nil
}

func (pg *PostgreSQL) Close() error {
	return pg.db.Close()
}

func (pg *PostgreSQL) createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS yubikey_keys (
		key_id VARCHAR(12) PRIMARY KEY,
		aes_key_encrypted TEXT NOT NULL,
		description TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		last_used_at TIMESTAMP WITH TIME ZONE,
		usage_count INTEGER DEFAULT 0,
		active BOOLEAN DEFAULT TRUE
	);

	CREATE INDEX IF NOT EXISTS idx_yubikey_keys_active ON yubikey_keys (active);
	CREATE INDEX IF NOT EXISTS idx_yubikey_keys_last_used ON yubikey_keys (last_used_at);

	CREATE TABLE IF NOT EXISTS yubikey_counters (
		key_id VARCHAR(12) REFERENCES yubikey_keys(key_id) ON DELETE CASCADE,
		counter INTEGER NOT NULL,
		session_use INTEGER NOT NULL,
		timestamp_high INTEGER NOT NULL,
		timestamp_low INTEGER NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		
		PRIMARY KEY (key_id, counter, session_use)
	);

	CREATE INDEX IF NOT EXISTS idx_yubikey_counters_key_counter ON yubikey_counters (key_id, counter DESC);
	`

	_, err := pg.db.Exec(schema)
	return err
}

func (pg *PostgreSQL) StoreKey(key *YubikeyKey) error {
	query := `
		INSERT INTO yubikey_keys (key_id, aes_key_encrypted, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	now := time.Now()
	_, err := pg.db.Exec(query, key.KeyID, key.AESKeyEncrypted, key.Description, now, now)
	return err
}

func (pg *PostgreSQL) GetKey(keyID string) (*YubikeyKey, error) {
	query := `
		SELECT key_id, aes_key_encrypted, description, created_at, updated_at, last_used_at, usage_count, active
		FROM yubikey_keys
		WHERE key_id = $1 AND active = TRUE
	`

	rows, err := pg.db.Query(query, keyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	return scanYubikeyKey(rows)
}

func (pg *PostgreSQL) ListKeys() ([]*YubikeyKey, error) {
	query := `
		SELECT key_id, aes_key_encrypted, description, created_at, updated_at, last_used_at, usage_count, active
		FROM yubikey_keys
		WHERE active = TRUE
		ORDER BY created_at DESC
	`

	rows, err := pg.db.Query(query)
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

func (pg *PostgreSQL) DeleteKey(keyID string) error {
	query := `UPDATE yubikey_keys SET active = FALSE WHERE key_id = $1`
	_, err := pg.db.Exec(query, keyID)
	return err
}

func (pg *PostgreSQL) UpdateKeyUsage(keyID string) error {
	query := `
		UPDATE yubikey_keys 
		SET last_used_at = NOW(), usage_count = usage_count + 1, updated_at = NOW()
		WHERE key_id = $1
	`
	_, err := pg.db.Exec(query, keyID)
	return err
}

func (pg *PostgreSQL) ValidateCounter(keyID string, counter, sessionUse int) error {
	query := `
		SELECT COUNT(*) FROM yubikey_counters
		WHERE key_id = $1 AND counter >= $2 AND session_use >= $3
	`

	var count int
	err := pg.db.QueryRow(query, keyID, counter, sessionUse).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("replay attack detected")
	}

	return nil
}

func (pg *PostgreSQL) StoreCounter(counter *YubikeyCounter) error {
	query := `
		INSERT INTO yubikey_counters (key_id, counter, session_use, timestamp_high, timestamp_low, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (key_id, counter, session_use) DO NOTHING
	`

	_, err := pg.db.Exec(query,
		counter.KeyID,
		counter.Counter,
		counter.SessionUse,
		counter.TimestampHigh,
		counter.TimestampLow,
		counter.CreatedAt,
	)
	return err
}

func (pg *PostgreSQL) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return pg.db.PingContext(ctx)
}
