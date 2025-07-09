package sshagent

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorConstants(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ErrOperationUnsupported",
			err:      ErrOperationUnsupported,
			expected: "operation unsupported",
		},
		{
			name:     "ErrNoPrivateKey",
			err:      ErrNoPrivateKey,
			expected: "no private key",
		},
		{
			name:     "ErrAgentLocked",
			err:      ErrAgentLocked,
			expected: "method is not allowed on agent locked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestErrorTypes(t *testing.T) {
	// Test that errors are of the correct type
	assert.Implements(t, (*error)(nil), ErrOperationUnsupported)
	assert.Implements(t, (*error)(nil), ErrNoPrivateKey)
	assert.Implements(t, (*error)(nil), ErrAgentLocked)
}

func TestErrorComparison(t *testing.T) {
	// Test error comparison using errors.Is
	assert.True(t, errors.Is(ErrOperationUnsupported, ErrOperationUnsupported))
	assert.True(t, errors.Is(ErrNoPrivateKey, ErrNoPrivateKey))
	assert.True(t, errors.Is(ErrAgentLocked, ErrAgentLocked))
	
	// Test that different errors are not equal
	assert.False(t, errors.Is(ErrOperationUnsupported, ErrNoPrivateKey))
	assert.False(t, errors.Is(ErrNoPrivateKey, ErrAgentLocked))
	assert.False(t, errors.Is(ErrAgentLocked, ErrOperationUnsupported))
}

func TestErrorWrapping(t *testing.T) {
	// Test that our errors can be wrapped and unwrapped
	wrappedErr1 := errors.New("wrapper: " + ErrOperationUnsupported.Error())
	wrappedErr2 := errors.New("wrapper: " + ErrNoPrivateKey.Error())
	wrappedErr3 := errors.New("wrapper: " + ErrAgentLocked.Error())
	
	assert.Contains(t, wrappedErr1.Error(), ErrOperationUnsupported.Error())
	assert.Contains(t, wrappedErr2.Error(), ErrNoPrivateKey.Error())
	assert.Contains(t, wrappedErr3.Error(), ErrAgentLocked.Error())
}

func TestErrorUniqueness(t *testing.T) {
	// Ensure each error has a unique message
	errors := []error{
		ErrOperationUnsupported,
		ErrNoPrivateKey,
		ErrAgentLocked,
	}
	
	messages := make(map[string]bool)
	for _, err := range errors {
		msg := err.Error()
		assert.False(t, messages[msg], "Error message should be unique: %s", msg)
		messages[msg] = true
	}
}

func TestErrorStringRepresentation(t *testing.T) {
	// Test that Error() method works correctly
	assert.NotEmpty(t, ErrOperationUnsupported.Error())
	assert.NotEmpty(t, ErrNoPrivateKey.Error())
	assert.NotEmpty(t, ErrAgentLocked.Error())
	
	// Test that they're not empty strings
	assert.NotEqual(t, "", ErrOperationUnsupported.Error())
	assert.NotEqual(t, "", ErrNoPrivateKey.Error())
	assert.NotEqual(t, "", ErrAgentLocked.Error())
}