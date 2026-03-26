package sshagent

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListenAndServe(t *testing.T) {
	t.Run("InvalidPath", func(t *testing.T) {
		agent := createTestAgent()
		err := agent.ListenAndServe(context.Background(), "/invalid/path/test.sock")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to listen")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		agent := createTestAgent()

		tmpDir, err := os.MkdirTemp("", "listener-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		socketPath := filepath.Join(tmpDir, "test.sock")
		ctx, cancel := context.WithCancel(context.Background())

		errChan := make(chan error, 1)
		go func() {
			errChan <- agent.ListenAndServe(ctx, socketPath)
		}()

		require.NoError(t, waitForSocket(socketPath, 200*time.Millisecond))
		cancel()

		select {
		case err := <-errChan:
			assert.NoError(t, err)
		case <-time.After(time.Second):
			t.Fatal("listener did not shutdown in time")
		}
	})

	t.Run("SocketCleanup", func(t *testing.T) {
		agent := createTestAgent()

		tmpDir, err := os.MkdirTemp("", "listener-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		socketPath := filepath.Join(tmpDir, "test.sock")
		ctx, cancel := context.WithCancel(context.Background())

		errChan := make(chan error, 1)
		go func() {
			errChan <- agent.ListenAndServe(ctx, socketPath)
		}()

		require.NoError(t, waitForSocket(socketPath, 200*time.Millisecond))
		cancel()

		select {
		case <-errChan:
		case <-time.After(time.Second):
			t.Fatal("listener did not shutdown in time")
		}

		_, err = os.Stat(socketPath)
		assert.True(t, os.IsNotExist(err), "socket file should be cleaned up")
	})
}
