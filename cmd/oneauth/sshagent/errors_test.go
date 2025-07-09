package sshagent

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSSHAgentErrors(t *testing.T) {
	t.Run("ErrorConstants", func(t *testing.T) {
		assert.Contains(t, ErrOperationUnsupported.Error(), "operation unsupported")
		assert.Contains(t, ErrNoPrivateKey.Error(), "no private key")
		assert.Contains(t, ErrAgentLocked.Error(), "agent locked")
	})

	t.Run("PINNotFoundError", func(t *testing.T) {
		err := ErrPINNotFound
		assert.Error(t, err)
		assert.Equal(t, "pin not found", err.Error())
		assert.True(t, errors.Is(err, ErrPINNotFound))
		assert.NotEqual(t, err, ErrAgentLocked)
	})

	t.Run("ErrorWrapping", func(t *testing.T) {
		wrappedErr := errors.New("wrapped error")

		// Test that our errors can be wrapped
		err := errors.Join(ErrPINNotFound, wrappedErr)
		assert.True(t, errors.Is(err, ErrPINNotFound))
		assert.True(t, errors.Is(err, wrappedErr))
	})
}
