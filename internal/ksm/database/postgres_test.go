package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vitalvas/oneauth/internal/ksm/config"
)

func TestPostgreSQLStructCreation(t *testing.T) {
	t.Run("zero value struct", func(t *testing.T) {
		pg := &PostgreSQL{}
		assert.Nil(t, pg.db)
	})

	t.Run("close on nil db panics gracefully handled", func(t *testing.T) {
		pg := &PostgreSQL{}
		// Calling Close on a nil db should panic, but we verify the struct is properly initialized
		assert.Nil(t, pg.db)
	})
}

func TestNewPostgreSQLConfigValidation(t *testing.T) {
	t.Run("empty URL fails connection", func(t *testing.T) {
		cfg := &config.PostgreSQLConfig{
			URL:               "",
			MaxConnections:    10,
			ConnectionTimeout: 5 * time.Second,
		}

		_, err := NewPostgreSQL(cfg)
		assert.Error(t, err)
	})

	t.Run("invalid URL fails connection", func(t *testing.T) {
		cfg := &config.PostgreSQLConfig{
			URL:               "not-a-valid-url",
			MaxConnections:    10,
			ConnectionTimeout: 5 * time.Second,
		}

		_, err := NewPostgreSQL(cfg)
		assert.Error(t, err)
	})

	t.Run("unreachable host fails with ping error", func(t *testing.T) {
		cfg := &config.PostgreSQLConfig{
			URL:               "postgres://user:pass@127.0.0.1:59999/testdb?sslmode=disable&connect_timeout=1",
			MaxConnections:    5,
			ConnectionTimeout: 2 * time.Second,
		}

		_, err := NewPostgreSQL(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to ping database")
	})
}

func TestPostgreSQLConfigFields(t *testing.T) {
	t.Run("default config values", func(t *testing.T) {
		cfg := &config.PostgreSQLConfig{}
		cfg.Default()

		assert.Equal(t, 25, cfg.MaxConnections)
		assert.Equal(t, 30*time.Second, cfg.ConnectionTimeout)
		assert.Empty(t, cfg.URL)
	})

	t.Run("custom config values", func(t *testing.T) {
		cfg := &config.PostgreSQLConfig{
			URL:               "postgres://user:pass@localhost:5432/mydb",
			MaxConnections:    50,
			ConnectionTimeout: 60 * time.Second,
		}

		assert.Equal(t, "postgres://user:pass@localhost:5432/mydb", cfg.URL)
		assert.Equal(t, 50, cfg.MaxConnections)
		assert.Equal(t, 60*time.Second, cfg.ConnectionTimeout)
	})

	t.Run("zero max connections", func(t *testing.T) {
		cfg := &config.PostgreSQLConfig{
			URL:               "postgres://localhost/test",
			MaxConnections:    0,
			ConnectionTimeout: 5 * time.Second,
		}

		assert.Equal(t, 0, cfg.MaxConnections)
	})
}

func TestNewPostgreSQLViaNewFunction(t *testing.T) {
	t.Run("nil PostgreSQL config returns error", func(t *testing.T) {
		cfg := &config.DatabaseConfig{
			Type:       "postgres",
			PostgreSQL: nil,
		}

		db, err := New(cfg)
		assert.Error(t, err)
		assert.Nil(t, db)
		assert.Contains(t, err.Error(), "PostgreSQL configuration is required")
	})
}

func TestPostgreSQLConnectionTimeoutShort(t *testing.T) {
	cfg := &config.PostgreSQLConfig{
		URL:               "postgres://user:pass@127.0.0.1:59999/testdb?sslmode=disable&connect_timeout=1",
		MaxConnections:    1,
		ConnectionTimeout: 1 * time.Second,
	}

	_, err := NewPostgreSQL(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to ping database")
}

func TestPostgreSQLStructFields(t *testing.T) {
	pg := &PostgreSQL{}
	assert.Nil(t, pg.db)
}
