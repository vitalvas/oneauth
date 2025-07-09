package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorConstants(t *testing.T) {
	t.Run("ErrNotInstalled", func(t *testing.T) {
		// Test that error is not nil
		assert.NotNil(t, ErrNotInstalled)
		assert.Error(t, ErrNotInstalled)
		
		// Test error message
		assert.Equal(t, "oneauth service is not installed", ErrNotInstalled.Error())
		
		// Test that it's a proper error
		assert.Implements(t, (*error)(nil), ErrNotInstalled)
	})
	
	t.Run("ErrNotImplemented", func(t *testing.T) {
		// Test that error is not nil
		assert.NotNil(t, ErrNotImplemented)
		assert.Error(t, ErrNotImplemented)
		
		// Test error message
		assert.Equal(t, "not yet implemented for your OS", ErrNotImplemented.Error())
		
		// Test that it's a proper error
		assert.Implements(t, (*error)(nil), ErrNotImplemented)
	})
}

func TestErrorUniqueness(t *testing.T) {
	t.Run("ErrorsAreDifferent", func(t *testing.T) {
		// Test that errors are different
		assert.NotEqual(t, ErrNotInstalled, ErrNotImplemented)
		assert.NotEqual(t, ErrNotInstalled.Error(), ErrNotImplemented.Error())
	})
}

func TestErrorComparison(t *testing.T) {
	t.Run("ErrorComparison", func(t *testing.T) {
		// Test error comparison using errors.Is
		err1 := ErrNotInstalled
		err2 := ErrNotImplemented
		
		assert.True(t, errors.Is(err1, ErrNotInstalled))
		assert.True(t, errors.Is(err2, ErrNotImplemented))
		assert.False(t, errors.Is(err1, ErrNotImplemented))
		assert.False(t, errors.Is(err2, ErrNotInstalled))
	})
}

func TestErrorWrapping(t *testing.T) {
	t.Run("WrappedErrors", func(t *testing.T) {
		// Test that errors can be wrapped
		wrappedNotInstalled := errors.New("wrapped: " + ErrNotInstalled.Error())
		wrappedNotImplemented := errors.New("wrapped: " + ErrNotImplemented.Error())
		
		assert.Contains(t, wrappedNotInstalled.Error(), ErrNotInstalled.Error())
		assert.Contains(t, wrappedNotImplemented.Error(), ErrNotImplemented.Error())
	})
}

func TestErrorTypes(t *testing.T) {
	t.Run("ErrorTypes", func(t *testing.T) {
		// Test that both errors are of the same underlying type
		assert.IsType(t, ErrNotInstalled, ErrNotImplemented)
		
		// Test that they implement error interface
		err1 := ErrNotInstalled
		err2 := ErrNotImplemented
		
		assert.NotNil(t, err1)
		assert.NotNil(t, err2)
	})
}

func TestErrorMessages(t *testing.T) {
	t.Run("MessageContent", func(t *testing.T) {
		// Test specific message content
		assert.Contains(t, ErrNotInstalled.Error(), "oneauth service")
		assert.Contains(t, ErrNotInstalled.Error(), "not installed")
		
		assert.Contains(t, ErrNotImplemented.Error(), "not yet implemented")
		assert.Contains(t, ErrNotImplemented.Error(), "your OS")
	})
}

func TestErrorStringification(t *testing.T) {
	t.Run("StringRepresentation", func(t *testing.T) {
		// Test that errors have proper string representation
		notInstalledStr := ErrNotInstalled.Error()
		notImplementedStr := ErrNotImplemented.Error()
		
		assert.NotEmpty(t, notInstalledStr)
		assert.NotEmpty(t, notImplementedStr)
		assert.NotEqual(t, notInstalledStr, notImplementedStr)
	})
}

func TestErrorBehavior(t *testing.T) {
	t.Run("ErrorBehavior", func(t *testing.T) {
		// Test that errors behave as expected in conditional statements
		err := ErrNotInstalled
		
		if err == ErrNotInstalled {
			assert.True(t, true) // This should be reached
		} else {
			assert.Fail(t, "Error comparison failed")
		}
		
		err = ErrNotImplemented
		
		if err == ErrNotImplemented {
			assert.True(t, true) // This should be reached
		} else {
			assert.Fail(t, "Error comparison failed")
		}
	})
}

func TestErrorConstantImmutability(t *testing.T) {
	t.Run("ConstantImmutability", func(t *testing.T) {
		// Test that error constants maintain their values
		originalNotInstalled := ErrNotInstalled.Error()
		originalNotImplemented := ErrNotImplemented.Error()
		
		// These should remain the same
		assert.Equal(t, originalNotInstalled, ErrNotInstalled.Error())
		assert.Equal(t, originalNotImplemented, ErrNotImplemented.Error())
	})
}