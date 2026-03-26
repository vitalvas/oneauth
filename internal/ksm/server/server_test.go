package server

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vitalvas/oneauth/internal/logger"
)

func TestNew(t *testing.T) {
	t.Run("missing config file returns error", func(t *testing.T) {
		srv, err := New("/nonexistent/path/config.yaml")
		assert.Error(t, err)
		assert.Nil(t, srv)
		assert.Contains(t, err.Error(), "failed to load configuration")
	})

	t.Run("empty config path returns validation error", func(t *testing.T) {
		// Empty path means config loads from env only; without required env vars it should fail validation
		srv, err := New("")
		assert.Error(t, err)
		assert.Nil(t, srv)
	})
}

func TestServerClose(t *testing.T) {
	t.Run("close with nil db returns nil", func(t *testing.T) {
		srv := &Server{
			db: nil,
		}

		err := srv.Close()
		assert.NoError(t, err)
	})
}

func TestServerStop(t *testing.T) {
	t.Run("stop with nil httpServer returns nil", func(t *testing.T) {
		srv := &Server{
			httpServer: nil,
			logger:     logger.New(""),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := srv.Stop(ctx)
		assert.NoError(t, err)
	})
}

func TestServerClose_WithDB(t *testing.T) {
	srv := setupTestServer(t)

	err := srv.Close()
	assert.NoError(t, err)
}

func TestServerStartAndStop(t *testing.T) {
	srv := setupTestServer(t)
	srv.config.Server.Address = "localhost:0"

	errChan := make(chan error, 1)
	go func() {
		errChan <- srv.Start()
	}()

	// Give server time to start
	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := srv.Stop(ctx)
	assert.NoError(t, err)

	err = srv.Close()
	assert.NoError(t, err)
}

func TestRun(t *testing.T) {
	t.Run("run with invalid config path returns error", func(t *testing.T) {
		err := Run("/nonexistent/config.yaml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load configuration")
	})

	t.Run("run with empty config returns validation error", func(t *testing.T) {
		err := Run("")
		assert.Error(t, err)
	})
}

func TestNew_ValidConfig(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "ksm-config-*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(`{
		"server": {"address": "localhost:0"},
		"database": {"type": "sqlite", "sqlite": {"path": ":memory:", "journal_mode": "WAL", "synchronous": "NORMAL"}},
		"security": {"master_key": "test-master-key-for-testing-1234567890"},
		"logging": {"level": "error", "format": "text"}
	}`)
	assert.NoError(t, err)
	tmpFile.Close()

	srv, err := New(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, srv)
	defer srv.Close()
}

func TestNew_InvalidMasterKey(t *testing.T) {
	// Test config loading with a temp file that has empty master key
	tmpFile, err := os.CreateTemp("", "ksm-config-*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(`{
		"server": {"address": "localhost:0"},
		"database": {"type": "sqlite", "sqlite": {"path": ":memory:", "journal_mode": "WAL", "synchronous": "NORMAL"}},
		"security": {"master_key": ""},
		"logging": {"level": "error", "format": "text"}
	}`)
	assert.NoError(t, err)
	tmpFile.Close()

	srv, err := New(tmpFile.Name())
	assert.Error(t, err)
	assert.Nil(t, srv)
}
