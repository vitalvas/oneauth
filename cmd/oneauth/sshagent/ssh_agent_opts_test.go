package sshagent

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock SSHAgent for testing
type MockSSHAgent struct {
	lockPassphrase []byte
}

func (a *MockSSHAgent) Lock(passphrase []byte) error {
	if a.lockPassphrase != nil {
		return errors.New("Lock: method is not allowed on agent locked")
	}

	if passphrase == nil {
		return errors.New("Lock: no private key")
	}

	a.lockPassphrase = make([]byte, len(passphrase))
	copy(a.lockPassphrase, passphrase)

	return nil
}

func (a *MockSSHAgent) Unlock(passphrase []byte) error {
	if a.lockPassphrase == nil {
		return errors.New("can't unlock not locked agent")
	}

	// Simple comparison for testing
	if string(passphrase) != string(a.lockPassphrase) {
		return errors.New("incorrect passphrase")
	}

	a.lockPassphrase = nil

	return nil
}

func TestSSHAgentLock(t *testing.T) {
	t.Run("LockWithValidPassphrase", func(t *testing.T) {
		agent := &MockSSHAgent{}
		passphrase := []byte("test-passphrase")
		
		err := agent.Lock(passphrase)
		assert.NoError(t, err)
		assert.NotNil(t, agent.lockPassphrase)
	})

	t.Run("LockWithNilPassphrase", func(t *testing.T) {
		agent := &MockSSHAgent{}
		
		err := agent.Lock(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no private key")
		assert.Nil(t, agent.lockPassphrase)
	})

	t.Run("LockAlreadyLocked", func(t *testing.T) {
		agent := &MockSSHAgent{}
		passphrase := []byte("test-passphrase")
		
		// Lock first time
		err := agent.Lock(passphrase)
		require.NoError(t, err)
		
		// Try to lock again
		err = agent.Lock(passphrase)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "method is not allowed on agent locked")
	})

	t.Run("LockWithEmptyPassphrase", func(t *testing.T) {
		agent := &MockSSHAgent{}
		passphrase := []byte("")
		
		err := agent.Lock(passphrase)
		assert.NoError(t, err)
		assert.NotNil(t, agent.lockPassphrase)
	})
}

func TestSSHAgentUnlock(t *testing.T) {
	t.Run("UnlockWithCorrectPassphrase", func(t *testing.T) {
		agent := &MockSSHAgent{}
		passphrase := []byte("test-passphrase")
		
		// Lock first
		err := agent.Lock(passphrase)
		require.NoError(t, err)
		
		// Unlock with correct passphrase
		err = agent.Unlock(passphrase)
		assert.NoError(t, err)
		assert.Nil(t, agent.lockPassphrase)
	})

	t.Run("UnlockWithIncorrectPassphrase", func(t *testing.T) {
		agent := &MockSSHAgent{}
		passphrase := []byte("test-passphrase")
		wrongPassphrase := []byte("wrong-passphrase")
		
		// Lock first
		err := agent.Lock(passphrase)
		require.NoError(t, err)
		
		// Unlock with wrong passphrase
		err = agent.Unlock(wrongPassphrase)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect passphrase")
		assert.NotNil(t, agent.lockPassphrase)
	})

	t.Run("UnlockNotLocked", func(t *testing.T) {
		agent := &MockSSHAgent{}
		passphrase := []byte("test-passphrase")
		
		// Try to unlock without locking first
		err := agent.Unlock(passphrase)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can't unlock not locked agent")
	})

	t.Run("UnlockWithNilPassphrase", func(t *testing.T) {
		agent := &MockSSHAgent{}
		passphrase := []byte("test-passphrase")
		
		// Lock first
		err := agent.Lock(passphrase)
		require.NoError(t, err)
		
		// Unlock with nil passphrase
		err = agent.Unlock(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect passphrase")
		assert.NotNil(t, agent.lockPassphrase)
	})
}

func TestSSHAgentLockUnlockCycle(t *testing.T) {
	t.Run("MultipleLockUnlockCycles", func(t *testing.T) {
		agent := &MockSSHAgent{}
		passphrase := []byte("test-passphrase")
		
		for i := 0; i < 3; i++ {
			// Lock
			err := agent.Lock(passphrase)
			require.NoError(t, err)
			assert.NotNil(t, agent.lockPassphrase)
			
			// Unlock
			err = agent.Unlock(passphrase)
			require.NoError(t, err)
			assert.Nil(t, agent.lockPassphrase)
		}
	})

	t.Run("LockUnlockWithDifferentPassphrases", func(t *testing.T) {
		agent := &MockSSHAgent{}
		
		passphrases := [][]byte{
			[]byte("passphrase1"),
			[]byte("passphrase2"),
			[]byte("passphrase3"),
		}
		
		for _, passphrase := range passphrases {
			// Lock with current passphrase
			err := agent.Lock(passphrase)
			require.NoError(t, err)
			
			// Unlock with same passphrase
			err = agent.Unlock(passphrase)
			require.NoError(t, err)
		}
	})
}

func TestSSHAgentLockState(t *testing.T) {
	t.Run("CheckLockState", func(t *testing.T) {
		agent := &MockSSHAgent{}
		passphrase := []byte("test-passphrase")
		
		// Initially unlocked
		assert.Nil(t, agent.lockPassphrase)
		
		// Lock
		err := agent.Lock(passphrase)
		require.NoError(t, err)
		assert.NotNil(t, agent.lockPassphrase)
		
		// Unlock
		err = agent.Unlock(passphrase)
		require.NoError(t, err)
		assert.Nil(t, agent.lockPassphrase)
	})

	t.Run("PassphraseNotExposed", func(t *testing.T) {
		agent := &MockSSHAgent{}
		passphrase := []byte("test-passphrase")
		
		// Lock
		err := agent.Lock(passphrase)
		require.NoError(t, err)
		
		// Modify original passphrase
		passphrase[0] = 'X'
		
		// Should still be able to unlock with original passphrase
		err = agent.Unlock([]byte("test-passphrase"))
		assert.NoError(t, err)
	})
}

func TestSSHAgentEdgeCases(t *testing.T) {
	t.Run("LockWithLongPassphrase", func(t *testing.T) {
		agent := &MockSSHAgent{}
		// Create a very long passphrase
		longPassphrase := make([]byte, 1000)
		for i := range longPassphrase {
			longPassphrase[i] = byte('a' + (i % 26))
		}
		
		err := agent.Lock(longPassphrase)
		assert.NoError(t, err)
		
		err = agent.Unlock(longPassphrase)
		assert.NoError(t, err)
	})

	t.Run("LockWithBinaryPassphrase", func(t *testing.T) {
		agent := &MockSSHAgent{}
		// Create a binary passphrase with null bytes
		binaryPassphrase := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}
		
		err := agent.Lock(binaryPassphrase)
		assert.NoError(t, err)
		
		err = agent.Unlock(binaryPassphrase)
		assert.NoError(t, err)
	})
}