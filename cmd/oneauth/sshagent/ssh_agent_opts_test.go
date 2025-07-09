package sshagent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vitalvas/oneauth/internal/mock"
)

func TestSSHAgent_LockUnlock(t *testing.T) {
	agent := &SSHAgent{
		softKeys: mock.NewKeystore(),
	}

	passphrase := []byte("test-passphrase")

	t.Run("Lock", func(t *testing.T) {
		err := agent.Lock(passphrase)
		assert.NoError(t, err)
		assert.NotNil(t, agent.lockPassphrase)
	})

	t.Run("LockAlreadyLocked", func(t *testing.T) {
		err := agent.Lock(passphrase)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "agent locked")
	})

	t.Run("UnlockWithCorrectPassphrase", func(t *testing.T) {
		err := agent.Unlock(passphrase)
		assert.NoError(t, err)
		assert.Nil(t, agent.lockPassphrase)
	})

	t.Run("UnlockWithIncorrectPassphrase", func(t *testing.T) {
		// Lock again
		agent.Lock(passphrase)

		err := agent.Unlock([]byte("wrong-passphrase"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect passphrase")

		// Unlock for cleanup
		agent.Unlock(passphrase)
	})

	t.Run("UnlockNotLocked", func(t *testing.T) {
		err := agent.Unlock(passphrase)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not locked")
	})

	t.Run("EdgeCases", func(t *testing.T) {
		// Test with nil passphrase - should fail
		err := agent.Lock(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no private key")

		// Test with empty passphrase
		err = agent.Lock([]byte{})
		assert.NoError(t, err)

		err = agent.Unlock([]byte{})
		assert.NoError(t, err)

		// Test with long passphrase
		longPassphrase := make([]byte, 1024)
		for i := range longPassphrase {
			longPassphrase[i] = byte(i % 256)
		}

		err = agent.Lock(longPassphrase)
		assert.NoError(t, err)

		err = agent.Unlock(longPassphrase)
		assert.NoError(t, err)
	})
}
