package sshagent

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vitalvas/oneauth/cmd/oneauth/config"
	"github.com/vitalvas/oneauth/internal/mock"
	"github.com/vitalvas/oneauth/internal/yubikey"
	"golang.org/x/crypto/ssh"
	sshagent "golang.org/x/crypto/ssh/agent"
)

// Test core functionality
func TestSSHAgent(t *testing.T) {
	// Test creation with YubiKey if available
	cards, err := yubikey.Cards()
	if err == nil && len(cards) > 0 {
		cfg := &config.Config{
			Keyring: config.Keyring{BeforeSignHook: "echo test", KeepKeySeconds: 300},
		}
		agent, err := New(cards[0].Serial, logrus.New(), cfg)
		require.NoError(t, err)
		assert.Equal(t, "echo test", agent.actions.BeforeSignHook)
		agent.Close()
	}

	// Test invalid creation
	agent, err := New(999999, logrus.New(), &config.Config{})
	assert.Error(t, err)
	assert.Nil(t, agent)

	// Test basic operations
	testAgent := createTestAgent()
	assert.NoError(t, testAgent.Close())
	assert.NoError(t, testAgent.Shutdown())

	// Test SSH operations
	_, err = testAgent.List()
	if err != nil {
		assert.Contains(t, err.Error(), "no yubikey available")
	}
	assert.NoError(t, testAgent.RemoveAll())

	// Test signing
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	pubkey, err := ssh.NewPublicKey(&key.PublicKey)
	require.NoError(t, err)

	_, err = testAgent.Sign(pubkey, []byte("test"))
	assert.Error(t, err)

	// Test soft keys
	addedKey := sshagent.AddedKey{PrivateKey: key, Comment: "test"}
	assert.NoError(t, testAgent.Add(addedKey))

	err = testAgent.Remove(pubkey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation unsupported")

	// Test lock/unlock
	passphrase := []byte("test")
	assert.NoError(t, testAgent.Lock(passphrase))
	assert.NotNil(t, testAgent.lockPassphrase)

	err = testAgent.Add(sshagent.AddedKey{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent locked")

	assert.NoError(t, testAgent.Unlock(passphrase))
	assert.Nil(t, testAgent.lockPassphrase)

	// Test edge cases
	assert.Error(t, testAgent.Lock(nil))
	assert.NoError(t, testAgent.Lock([]byte{}))
	assert.NoError(t, testAgent.Unlock([]byte{}))
}

// Test listener and errors
func TestSSHAgentAdvanced(t *testing.T) {
	testAgent := createTestAgent()

	// Test socket listener
	tmpDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	socketPath := filepath.Join(tmpDir, "test.sock")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		testAgent.ListenAndServe(ctx, socketPath)
	}()

	// Wait for socket and test connection
	require.NoError(t, waitForSocket(socketPath, 100*time.Millisecond))
	conn, err := net.Dial("unix", socketPath)
	require.NoError(t, err)
	conn.Close()
	cancel()

	// Test invalid path
	newAgent := createTestAgent()
	err = newAgent.ListenAndServe(context.Background(), "/invalid/path")
	assert.Error(t, err)

	// Test error handling
	mockAgent := createTestAgent()
	mockListener := &TemporaryErrorListener{}
	mockAgent.setListener(mockListener)

	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = mockAgent.ListenAndServe(ctx, "mock")
	assert.NoError(t, err)
	assert.Greater(t, mockListener.AcceptCalls, 0)

	// Test errors
	assert.Contains(t, ErrOperationUnsupported.Error(), "operation unsupported")
	assert.Contains(t, ErrAgentLocked.Error(), "agent locked")
	assert.Equal(t, "pin not found", ErrPINNotFound.Error())

	// Test error wrapping
	wrappedErr := errors.New("wrapped")
	err = errors.Join(ErrPINNotFound, wrappedErr)
	assert.True(t, errors.Is(err, ErrPINNotFound))

	// Test temporary interface
	tempErr := &mock.TemporaryError{Message: "test"}
	assert.True(t, tempErr.Temporary())

	if err, ok := (error(tempErr)).(Temporary); ok {
		assert.True(t, err.Temporary())
	}

	// Test connection handling
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	done := make(chan bool)
	go func() {
		testAgent.handleConn(server)
		done <- true
	}()
	client.Close()

	select {
	case <-done:
	case <-time.After(50 * time.Millisecond):
		t.Error("timeout")
	}
}

// Helper functions
func createTestAgent() *SSHAgent {
	return &SSHAgent{
		log:      logrus.New().WithField("test", "agent"),
		softKeys: mock.NewKeystore(),
	}
}

func waitForSocket(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
		time.Sleep(2 * time.Millisecond)
	}
	return fmt.Errorf("timeout")
}

// Mock for testing
type TemporaryErrorListener struct {
	AcceptCalls int
}

func (t *TemporaryErrorListener) Accept() (net.Conn, error) {
	t.AcceptCalls++
	return nil, &mock.TemporaryError{Message: "temp"}
}

func (t *TemporaryErrorListener) Close() error { return nil }

func (t *TemporaryErrorListener) Addr() net.Addr { return &mock.Addr{} }
