package sshagent

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock implementation of Temporary interface
type MockTemporaryError struct {
	message   string
	temporary bool
}

func (e *MockTemporaryError) Error() string {
	return e.message
}

func (e *MockTemporaryError) Temporary() bool {
	return e.temporary
}

func TestTemporaryInterface(t *testing.T) {
	t.Run("TemporaryError", func(t *testing.T) {
		err := &MockTemporaryError{
			message:   "temporary error",
			temporary: true,
		}
		
		// Test that it implements Temporary
		assert.Implements(t, (*Temporary)(nil), err)
		
		// Test that it implements error
		assert.Implements(t, (*error)(nil), err)
		
		// Test functionality
		assert.Equal(t, "temporary error", err.Error())
		assert.True(t, err.Temporary())
	})
	
	t.Run("NonTemporaryError", func(t *testing.T) {
		err := &MockTemporaryError{
			message:   "permanent error",
			temporary: false,
		}
		
		assert.Implements(t, (*Temporary)(nil), err)
		assert.Equal(t, "permanent error", err.Error())
		assert.False(t, err.Temporary())
	})
}

func TestTemporaryTypeAssertion(t *testing.T) {
	t.Run("TemporaryError", func(t *testing.T) {
		var err error = &MockTemporaryError{
			message:   "test error",
			temporary: true,
		}
		
		if tempErr, ok := err.(Temporary); ok {
			assert.True(t, tempErr.Temporary())
		} else {
			t.Error("Expected error to implement Temporary interface")
		}
	})
	
	t.Run("NonTemporaryError", func(t *testing.T) {
		err := fmt.Errorf("regular error")
		
		if tempErr, ok := err.(Temporary); ok {
			t.Errorf("Expected error to NOT implement Temporary interface, but got: %v", tempErr)
		}
	})
}

func TestTemporaryInterfaceUsage(t *testing.T) {
	errors := []error{
		&MockTemporaryError{message: "temp1", temporary: true},
		&MockTemporaryError{message: "temp2", temporary: false},
		fmt.Errorf("regular error"),
	}
	
	temporaryCount := 0
	for _, err := range errors {
		if tempErr, ok := err.(Temporary); ok && tempErr.Temporary() {
			temporaryCount++
		}
	}
	
	assert.Equal(t, 1, temporaryCount)
}

func TestTemporaryInterface_EdgeCases(t *testing.T) {
	t.Run("NilError", func(t *testing.T) {
		var err error
		
		if tempErr, ok := err.(Temporary); ok {
			t.Errorf("Expected nil error to NOT implement Temporary interface, but got: %v", tempErr)
		}
	})
	
	t.Run("EmptyMessage", func(t *testing.T) {
		err := &MockTemporaryError{
			message:   "",
			temporary: true,
		}
		
		assert.Equal(t, "", err.Error())
		assert.True(t, err.Temporary())
	})
}

// Test helper function for checking temporary errors
func isTemporaryError(err error) bool {
	if tempErr, ok := err.(Temporary); ok {
		return tempErr.Temporary()
	}
	return false
}

func TestTemporaryHelperFunction(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "TemporaryError",
			err:      &MockTemporaryError{message: "temp", temporary: true},
			expected: true,
		},
		{
			name:     "NonTemporaryError",
			err:      &MockTemporaryError{message: "perm", temporary: false},
			expected: false,
		},
		{
			name:     "RegularError",
			err:      fmt.Errorf("regular"),
			expected: false,
		},
		{
			name:     "NilError",
			err:      nil,
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTemporaryError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}