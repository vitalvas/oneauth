package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vitalvas/oneauth/internal/ksm/config"
	"github.com/vitalvas/oneauth/internal/ksm/crypto"
	"github.com/vitalvas/oneauth/internal/ksm/database"
	"github.com/vitalvas/oneauth/internal/logger"
)

func setupTestServer(t *testing.T) *Server {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "sqlite",
			SQLite: &config.SQLiteConfig{
				Path:        ":memory:",
				JournalMode: "WAL",
				Synchronous: "NORMAL",
			},
		},
		Security: config.SecurityConfig{
			MasterKey: "test-master-key-12345678901234567890",
		},
		Logging: config.LoggingConfig{
			Level:  "error",
			Format: "text",
		},
	}

	db, err := database.New(&cfg.Database)
	assert.NoError(t, err)

	cryptoEngine, err := crypto.NewEngine(cfg.Security.MasterKey)
	assert.NoError(t, err)

	log := logger.New("")
	return &Server{
		config: cfg,
		db:     db,
		crypto: cryptoEngine,
		logger: log,
	}
}

func createTestServer(mockDB database.DB, cryptoEngine *crypto.Engine) *Server {
	cfg := &config.Config{}
	log := logger.New("")

	return &Server{
		config: cfg,
		db:     mockDB,
		crypto: cryptoEngine,
		logger: log,
	}
}
