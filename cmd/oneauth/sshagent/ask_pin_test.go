package sshagent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrPINNotFound(t *testing.T) {
	assert.Equal(t, "pin not found", ErrPINNotFound.Error())
	assert.Implements(t, (*error)(nil), ErrPINNotFound)
}

func TestErrPINNotFoundUniqueness(t *testing.T) {
	// Test that ErrPINNotFound is different from other errors
	assert.NotEqual(t, ErrPINNotFound, ErrOperationUnsupported)
	assert.NotEqual(t, ErrPINNotFound, ErrNoPrivateKey)
	assert.NotEqual(t, ErrPINNotFound, ErrAgentLocked)
}

func TestErrPINNotFoundString(t *testing.T) {
	assert.NotEmpty(t, ErrPINNotFound.Error())
	assert.NotEqual(t, "", ErrPINNotFound.Error())
}

func TestAllSSHAgentErrors(t *testing.T) {
	// Test all errors defined in the package
	errors := []error{
		ErrOperationUnsupported,
		ErrNoPrivateKey,
		ErrAgentLocked,
		ErrPINNotFound,
	}

	// Check that all errors have non-empty messages
	for _, err := range errors {
		assert.NotEmpty(t, err.Error())
		assert.Implements(t, (*error)(nil), err)
	}

	// Check that all error messages are unique
	messages := make(map[string]bool)
	for _, err := range errors {
		msg := err.Error()
		assert.False(t, messages[msg], "Error message should be unique: %s", msg)
		messages[msg] = true
	}
}

// Test that the ErrPINNotFound is properly categorized
func TestErrPINNotFoundCategorization(t *testing.T) {
	// Test that ErrPINNotFound is not a temporary error
	if tempErr, ok := ErrPINNotFound.(Temporary); ok {
		assert.False(t, tempErr.Temporary(), "ErrPINNotFound should not be temporary")
	}

	// Test that it's a proper error
	assert.Error(t, ErrPINNotFound)
}
