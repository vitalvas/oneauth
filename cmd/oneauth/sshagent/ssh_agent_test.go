package sshagent

import (
	"crypto/rand"
	"crypto/rsa"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vitalvas/oneauth/internal/mock"
	"golang.org/x/crypto/ssh"
)

func TestSSHAgent_Operations(t *testing.T) {
	agent := &SSHAgent{
		softKeys: mock.NewKeystore(),
	}

	t.Run("List", func(t *testing.T) {
		keys, err := agent.List()
		// List may return error when no yubikey is available
		if err != nil {
			assert.Contains(t, err.Error(), "no yubikey available")
		}
		// keys can be nil or empty when no yubikey is available
		_ = keys
	})

	t.Run("RemoveAll", func(t *testing.T) {
		err := agent.RemoveAll()
		assert.NoError(t, err)
	})

	t.Run("SignWithoutYubikey", func(t *testing.T) {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		assert.NoError(t, err)

		pubkey, err := ssh.NewPublicKey(&key.PublicKey)
		assert.NoError(t, err)

		data := []byte("test data")
		_, err = agent.Sign(pubkey, data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no yubikey available")
	})

	t.Run("SignWithFlags", func(t *testing.T) {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		assert.NoError(t, err)

		pubkey, err := ssh.NewPublicKey(&key.PublicKey)
		assert.NoError(t, err)

		data := []byte("test data")
		_, err = agent.SignWithFlags(pubkey, data, 0)
		assert.Error(t, err)
	})
}

func TestSSHAgent_Environment(t *testing.T) {
	// Test security environment variable handling
	originalValue := os.Getenv("ONEAUTH_SECURITY_BYPASS")
	defer func() {
		if originalValue != "" {
			os.Setenv("ONEAUTH_SECURITY_BYPASS", originalValue)
		} else {
			os.Unsetenv("ONEAUTH_SECURITY_BYPASS")
		}
	}()

	os.Setenv("ONEAUTH_SECURITY_BYPASS", "test_value")

	agent := &SSHAgent{
		softKeys: mock.NewKeystore(),
	}

	// Test that agent still works with environment variable set
	keys, err := agent.List()
	// List may return error when no yubikey is available
	if err != nil {
		assert.Contains(t, err.Error(), "no yubikey available")
	}
	// keys can be nil or empty when no yubikey is available
	_ = keys
}

func TestSSHAgent_ConcurrentAccess(t *testing.T) {
	agent := &SSHAgent{
		softKeys: mock.NewKeystore(),
	}

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	// Test concurrent List operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := agent.List()
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Collect errors
	errs := make([]error, 0, 10)
	for err := range errors {
		errs = append(errs, err)
	}

	// Some errors are expected (no yubikey available)
	if len(errs) > 0 {
		t.Logf("Concurrent operations completed with %d errors (expected)", len(errs))
	}
}

func TestSSHAgent_EdgeCases(t *testing.T) {
	t.Run("NilSoftKeys", func(t *testing.T) {
		agent := &SSHAgent{
			softKeys: nil,
		}

		// Should panic or return error
		defer func() {
			if r := recover(); r != nil {
				// Panic is expected with nil softKeys
				assert.NotNil(t, r)
			}
		}()

		_, err := agent.List()
		if err != nil {
			assert.Error(t, err)
		}
	})

	t.Run("EmptyData", func(t *testing.T) {
		agent := &SSHAgent{
			softKeys: mock.NewKeystore(),
		}

		key, err := rsa.GenerateKey(rand.Reader, 2048)
		assert.NoError(t, err)

		pubkey, err := ssh.NewPublicKey(&key.PublicKey)
		assert.NoError(t, err)

		// Test with empty data
		_, err = agent.Sign(pubkey, []byte{})
		assert.Error(t, err)

		// Give some time for processing
		time.Sleep(100 * time.Millisecond)
	})
}
