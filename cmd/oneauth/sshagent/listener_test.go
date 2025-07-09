package sshagent

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vitalvas/oneauth/internal/mock"
)

func TestSSHAgent_ListenAndServe(t *testing.T) {
	t.Run("ValidSocket", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "sshagent_test")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		socketPath := filepath.Join(tmpDir, "test.sock")
		agent := createTestAgent()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		errChan := make(chan error, 1)
		go func() {
			errChan <- agent.ListenAndServe(ctx, socketPath)
		}()

		// Wait for socket creation
		require.NoError(t, waitForSocket(socketPath, 100*time.Millisecond))

		// Verify socket permissions
		info, err := os.Stat(socketPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

		// Test connection
		conn, err := net.Dial("unix", socketPath)
		require.NoError(t, err)
		conn.Close()

		// Cleanup
		cancel()
		select {
		case err := <-errChan:
			assert.NoError(t, err)
		case <-time.After(200 * time.Millisecond):
			t.Fatal("Server did not stop")
		}
	})

	t.Run("InvalidPath", func(t *testing.T) {
		agent := createTestAgent()
		err := agent.ListenAndServe(context.Background(), "/invalid/path/socket")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to listen")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "sshagent_test")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		socketPath := filepath.Join(tmpDir, "test.sock")
		agent := createTestAgent()

		ctx, cancel := context.WithCancel(context.Background())
		errChan := make(chan error, 1)

		go func() {
			errChan <- agent.ListenAndServe(ctx, socketPath)
		}()

		// Wait for server to start
		require.NoError(t, waitForSocket(socketPath, 100*time.Millisecond))

		// Cancel context
		cancel()

		select {
		case err := <-errChan:
			assert.NoError(t, err)
		case <-time.After(200 * time.Millisecond):
			t.Fatal("Server did not stop")
		}
	})
}

func TestSSHAgent_ListenAndServe_ErrorHandling(t *testing.T) {
	t.Run("TemporaryError", func(t *testing.T) {
		agent := createTestAgent()
		mockListener := &TemporaryErrorListener{AcceptCalls: 0}
		agent.setListener(mockListener)

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		err := agent.ListenAndServe(ctx, "mock-socket")
		assert.NoError(t, err)
		assert.Greater(t, mockListener.AcceptCalls, 1)
	})

	t.Run("ClosedListener", func(t *testing.T) {
		agent := createTestAgent()
		mockListener := &mock.ClosedListener{}
		agent.setListener(mockListener)

		err := agent.ListenAndServe(context.Background(), "mock-socket")
		assert.NoError(t, err)
	})

	t.Run("EOFListener", func(t *testing.T) {
		agent := createTestAgent()
		mockListener := &mock.EOFListener{}
		agent.setListener(mockListener)

		err := agent.ListenAndServe(context.Background(), "mock-socket")
		assert.NoError(t, err)
	})
}

// Helper functions
func createTestAgent() *SSHAgent {
	return &SSHAgent{
		log:      logrus.New().WithField("test", "listen"),
		softKeys: mock.NewKeystore(),
	}
}

func waitForSocket(socketPath string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(socketPath); err == nil {
			return nil
		}
		time.Sleep(2 * time.Millisecond)
	}
	return fmt.Errorf("socket not ready within timeout")
}

// TemporaryErrorListener for testing
type TemporaryErrorListener struct {
	AcceptCalls int
}

func (t *TemporaryErrorListener) Accept() (net.Conn, error) {
	t.AcceptCalls++
	return nil, &mock.TemporaryError{Message: "temporary error"}
}

func (t *TemporaryErrorListener) Close() error {
	return nil
}

func (t *TemporaryErrorListener) Addr() net.Addr {
	return &mock.Addr{}
}
