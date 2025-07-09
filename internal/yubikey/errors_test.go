package yubikey

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
			name:     "ErrCardReaderUnavailable",
			err:      ErrCardReaderUnavailable,
			expected: "the specified reader is not currently available for use",
		},
		{
			name:     "ErrYubikeyNotOpen",
			err:      ErrYubikeyNotOpen,
			expected: "yubikey not opened",
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
	assert.Implements(t, (*error)(nil), ErrCardReaderUnavailable)
	assert.Implements(t, (*error)(nil), ErrYubikeyNotOpen)
}

func TestErrorComparison(t *testing.T) {
	// Test error comparison using errors.Is
	assert.True(t, errors.Is(ErrCardReaderUnavailable, ErrCardReaderUnavailable))
	assert.True(t, errors.Is(ErrYubikeyNotOpen, ErrYubikeyNotOpen))

	// Test that different errors are not equal
	assert.False(t, errors.Is(ErrCardReaderUnavailable, ErrYubikeyNotOpen))
	assert.False(t, errors.Is(ErrYubikeyNotOpen, ErrCardReaderUnavailable))
}

func TestErrorWrapping(t *testing.T) {
	// Test that our errors can be wrapped and unwrapped
	wrappedErr1 := errors.New("wrapper: " + ErrCardReaderUnavailable.Error())
	wrappedErr2 := errors.New("wrapper: " + ErrYubikeyNotOpen.Error())

	assert.Contains(t, wrappedErr1.Error(), ErrCardReaderUnavailable.Error())
	assert.Contains(t, wrappedErr2.Error(), ErrYubikeyNotOpen.Error())
}

func TestErrorUniqueness(t *testing.T) {
	// Ensure each error has a unique message
	errors := []error{
		ErrCardReaderUnavailable,
		ErrYubikeyNotOpen,
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
	assert.NotEmpty(t, ErrCardReaderUnavailable.Error())
	assert.NotEmpty(t, ErrYubikeyNotOpen.Error())

	// Test that they're not empty strings
	assert.NotEqual(t, "", ErrCardReaderUnavailable.Error())
	assert.NotEqual(t, "", ErrYubikeyNotOpen.Error())
}
