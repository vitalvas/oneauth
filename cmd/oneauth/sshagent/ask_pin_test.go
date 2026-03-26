package sshagent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrPINNotFound(t *testing.T) {
	t.Run("ErrorMessage", func(t *testing.T) {
		assert.Equal(t, "pin not found", ErrPINNotFound.Error())
	})

	t.Run("ErrorIsNotNil", func(t *testing.T) {
		assert.NotNil(t, ErrPINNotFound)
	})
}
