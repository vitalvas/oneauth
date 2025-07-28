package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/vitalvas/oneauth/internal/ksm/config"
)

type YubikeyKey struct {
	KeyID           string     `db:"key_id"`
	AESKeyEncrypted string     `db:"aes_key_encrypted"`
	Description     string     `db:"description"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
	LastUsedAt      *time.Time `db:"last_used_at"`
	UsageCount      int        `db:"usage_count"`
	Active          bool       `db:"active"`
}

type YubikeyCounter struct {
	KeyID         string    `db:"key_id"`
	Counter       int       `db:"counter"`
	SessionUse    int       `db:"session_use"`
	TimestampHigh int       `db:"timestamp_high"`
	TimestampLow  int       `db:"timestamp_low"`
	CreatedAt     time.Time `db:"created_at"`
}

type DB interface {
	Close() error

	StoreKey(key *YubikeyKey) error
	GetKey(keyID string) (*YubikeyKey, error)
	ListKeys() ([]*YubikeyKey, error)
	DeleteKey(keyID string) error
	UpdateKeyUsage(keyID string) error

	ValidateCounter(keyID string, counter, sessionUse int) error
	StoreCounter(counter *YubikeyCounter) error

	HealthCheck() error
}

func New(dbConfig *config.DatabaseConfig) (DB, error) {
	switch dbConfig.Type {
	case "postgres":
		if dbConfig.PostgreSQL == nil {
			return nil, fmt.Errorf("PostgreSQL configuration is required")
		}
		return NewPostgreSQL(dbConfig.PostgreSQL)
	case "sqlite":
		if dbConfig.SQLite == nil {
			return nil, fmt.Errorf("SQLite configuration is required")
		}
		return NewSQLite(dbConfig.SQLite)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbConfig.Type)
	}
}

func scanYubikeyKey(rows *sql.Rows) (*YubikeyKey, error) {
	key := &YubikeyKey{}
	err := rows.Scan(
		&key.KeyID,
		&key.AESKeyEncrypted,
		&key.Description,
		&key.CreatedAt,
		&key.UpdatedAt,
		&key.LastUsedAt,
		&key.UsageCount,
		&key.Active,
	)
	return key, err
}
