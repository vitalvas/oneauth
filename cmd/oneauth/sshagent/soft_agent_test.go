package sshagent

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	sshagent "golang.org/x/crypto/ssh/agent"
)

func TestSoftAgent(t *testing.T) {
	// Test creation
	agent := NewSoftAgent("test", 300, logrus.New())
	assert.NotNil(t, agent)
	assert.Equal(t, "test", agent.name)

	// Test Close
	assert.NoError(t, agent.Close())

	// Test Shutdown
	assert.NoError(t, agent.Shutdown())
}

func TestSoftAgentKeyOperations(t *testing.T) {
	agent := NewSoftAgent("test", 300, logrus.New())
	defer agent.Close()

	// Generate test key
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	pubkey, err := ssh.NewPublicKey(&key.PublicKey)
	require.NoError(t, err)

	// Test List - empty
	keys, err := agent.List()
	require.NoError(t, err)
	assert.Empty(t, keys)

	// Test Add
	addedKey := sshagent.AddedKey{PrivateKey: key, Comment: "test-key"}
	assert.NoError(t, agent.Add(addedKey))

	// Test List - with key
	keys, err = agent.List()
	require.NoError(t, err)
	assert.Len(t, keys, 1)
	assert.Equal(t, "test-key", keys[0].Comment)

	// Test Sign
	sig, err := agent.Sign(pubkey, []byte("test data"))
	require.NoError(t, err)
	assert.NotNil(t, sig)

	// Test SignWithFlags
	sig, err = agent.SignWithFlags(pubkey, []byte("test data"), 0)
	require.NoError(t, err)
	assert.NotNil(t, sig)

	// Test Sign with unknown key
	unknownKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	unknownPubkey, _ := ssh.NewPublicKey(&unknownKey.PublicKey)
	_, err = agent.Sign(unknownPubkey, []byte("test"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown key")

	// Test Remove
	err = agent.Remove(pubkey)
	assert.NoError(t, err)

	// Verify key was removed
	keys, err = agent.List()
	require.NoError(t, err)
	assert.Empty(t, keys)

	// Test RemoveAll
	agent.Add(sshagent.AddedKey{PrivateKey: key, Comment: "key1"})
	assert.NoError(t, agent.RemoveAll())
	keys, err = agent.List()
	require.NoError(t, err)
	assert.Empty(t, keys)
}

func TestSoftAgentLockUnlock(t *testing.T) {
	agent := NewSoftAgent("test", 300, logrus.New())
	defer agent.Close()

	passphrase := []byte("test-passphrase")

	// Test Lock
	assert.NoError(t, agent.Lock(passphrase))
	assert.NotNil(t, agent.lockPassphrase)

	// Test double lock fails
	err := agent.Lock(passphrase)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent locked")

	// Test operations while locked
	_, err = agent.List()
	assert.Error(t, err)
	assert.Equal(t, ErrAgentLocked, err)

	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	err = agent.Add(sshagent.AddedKey{PrivateKey: key})
	assert.Error(t, err)
	assert.Equal(t, ErrAgentLocked, err)

	pubkey, _ := ssh.NewPublicKey(&key.PublicKey)
	_, err = agent.Sign(pubkey, []byte("test"))
	assert.Error(t, err)
	assert.Equal(t, ErrAgentLocked, err)

	err = agent.Remove(pubkey)
	assert.Error(t, err)
	assert.Equal(t, ErrAgentLocked, err)

	// Test Unlock with wrong passphrase
	err = agent.Unlock([]byte("wrong"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "incorrect passphrase")

	// Test Unlock with correct passphrase
	assert.NoError(t, agent.Unlock(passphrase))
	assert.Nil(t, agent.lockPassphrase)

	// Test unlock when not locked
	err = agent.Unlock(passphrase)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "can't unlock not locked agent")

	// Test Lock with nil passphrase
	err = agent.Lock(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no private key")

	// Test Lock with empty passphrase
	assert.NoError(t, agent.Lock([]byte{}))
	assert.NoError(t, agent.Unlock([]byte{}))
}

func TestSoftAgentUnsupportedOperations(t *testing.T) {
	agent := NewSoftAgent("test", 300, logrus.New())
	defer agent.Close()

	// Test Signers - unsupported
	signers, err := agent.Signers()
	assert.Error(t, err)
	assert.Nil(t, signers)
	assert.Contains(t, err.Error(), "operation unsupported")

	// Test Extension - unsupported
	result, err := agent.Extension("test", []byte("data"))
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, sshagent.ErrExtensionUnsupported, err)
}

func TestSoftAgentListener(t *testing.T) {
	agent := NewSoftAgent("test", 300, logrus.New())

	// Create temp directory for socket
	tmpDir, err := os.MkdirTemp("", "soft-agent-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	socketPath := filepath.Join(tmpDir, "test.sock")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start listener in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- agent.ListenAndServe(ctx, socketPath)
	}()

	// Wait for socket to be created
	require.NoError(t, waitForSocket(socketPath, 100*time.Millisecond))

	// Test connection
	conn, err := net.Dial("unix", socketPath)
	require.NoError(t, err)
	conn.Close()

	// Verify socket permissions
	info, err := os.Stat(socketPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	// Shutdown
	cancel()

	select {
	case err := <-errChan:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("listener did not shutdown in time")
	}

	// Verify socket was cleaned up
	_, err = os.Stat(socketPath)
	assert.True(t, os.IsNotExist(err))
}

func TestSoftAgentListenerInvalidPath(t *testing.T) {
	agent := NewSoftAgent("test", 300, logrus.New())

	err := agent.ListenAndServe(context.Background(), "/invalid/path/test.sock")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to listen")
}

func TestSoftAgentListenerTemporaryError(t *testing.T) {
	agent := NewSoftAgent("test", 300, logrus.New())

	mockListener := &TemporaryErrorListener{}
	agent.setListener(mockListener)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := agent.ListenAndServe(ctx, "mock")
	assert.NoError(t, err)
	assert.Greater(t, mockListener.AcceptCalls, 0)
}

func TestSoftAgentHandleConn(t *testing.T) {
	agent := NewSoftAgent("test", 300, logrus.New())

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	done := make(chan bool)
	go func() {
		agent.handleConn(server)
		done <- true
	}()
	client.Close()

	select {
	case <-done:
		// Success
	case <-time.After(50 * time.Millisecond):
		t.Error("handleConn did not complete in time")
	}
}

func TestSoftAgentExistingSocket(t *testing.T) {
	agent := NewSoftAgent("test", 300, logrus.New())

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "soft-agent-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	socketPath := filepath.Join(tmpDir, "test.sock")

	// Create existing regular file
	f, err := os.Create(socketPath)
	require.NoError(t, err)
	f.Close()

	// Verify it's a regular file before starting agent
	info, err := os.Stat(socketPath)
	require.NoError(t, err)
	assert.True(t, info.Mode().IsRegular())

	ctx, cancel := context.WithCancel(context.Background())

	// Start listener - should remove existing file and create socket
	go func() {
		agent.ListenAndServe(ctx, socketPath)
	}()

	// Wait for socket to be connectable (not just file to exist)
	require.NoError(t, waitForConnectableSocket(socketPath, 500*time.Millisecond))

	cancel()
	agent.Shutdown()
}

func waitForConnectableSocket(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.Dial("unix", path)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(5 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for socket to be connectable")
}

func TestSoftAgentIntegration(t *testing.T) {
	// Integration test using actual SSH agent client
	agent := NewSoftAgent("integration-test", 300, logrus.New())

	tmpDir, err := os.MkdirTemp("", "soft-agent-integration")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	socketPath := filepath.Join(tmpDir, "agent.sock")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		agent.ListenAndServe(ctx, socketPath)
	}()

	require.NoError(t, waitForSocket(socketPath, 100*time.Millisecond))

	// Connect using SSH agent client
	conn, err := net.Dial("unix", socketPath)
	require.NoError(t, err)
	defer conn.Close()

	client := sshagent.NewClient(conn)

	// List keys - should be empty
	keys, err := client.List()
	require.NoError(t, err)
	assert.Empty(t, keys)

	// Add a key
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	err = client.Add(sshagent.AddedKey{
		PrivateKey: key,
		Comment:    "integration-test-key",
	})
	require.NoError(t, err)

	// List keys - should have one
	keys, err = client.List()
	require.NoError(t, err)
	assert.Len(t, keys, 1)
	assert.Equal(t, "integration-test-key", keys[0].Comment)

	// Sign data
	pubkey, err := ssh.NewPublicKey(&key.PublicKey)
	require.NoError(t, err)

	sig, err := client.Sign(pubkey, []byte("test data to sign"))
	require.NoError(t, err)
	assert.NotNil(t, sig)

	// Remove key
	err = client.Remove(pubkey)
	require.NoError(t, err)

	// Verify removed
	keys, err = client.List()
	require.NoError(t, err)
	assert.Empty(t, keys)

	// Test lock/unlock
	err = client.Lock([]byte("passphrase"))
	require.NoError(t, err)

	// Should fail while locked
	_, err = client.List()
	assert.Error(t, err)

	err = client.Unlock([]byte("passphrase"))
	require.NoError(t, err)

	// Should work after unlock
	keys, err = client.List()
	require.NoError(t, err)
	assert.Empty(t, keys)
}

func TestSoftAgentMultipleKeys(t *testing.T) {
	agent := NewSoftAgent("multi-key-test", 300, logrus.New())
	defer agent.Close()

	// Add multiple keys
	for i := 0; i < 5; i++ {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		err = agent.Add(sshagent.AddedKey{
			PrivateKey: key,
			Comment:    "test-key-" + string(rune('a'+i)),
		})
		require.NoError(t, err)
	}

	// List should show all keys
	keys, err := agent.List()
	require.NoError(t, err)
	assert.Len(t, keys, 5)

	// RemoveAll should clear all
	assert.NoError(t, agent.RemoveAll())

	keys, err = agent.List()
	require.NoError(t, err)
	assert.Empty(t, keys)
}
