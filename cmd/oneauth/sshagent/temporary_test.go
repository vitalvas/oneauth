package sshagent

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vitalvas/oneauth/internal/mock"
)

func TestTemporaryInterface(t *testing.T) {
	t.Run("TemporaryError", func(t *testing.T) {
		tempErr := &mock.TemporaryError{Message: "test error"}

		assert.True(t, tempErr.Temporary())
		assert.Equal(t, "test error", tempErr.Error())

		// Test type assertion
		if err, ok := (error(tempErr)).(Temporary); ok {
			assert.True(t, err.Temporary())
		} else {
			t.Error("Expected Temporary interface")
		}
	})

	t.Run("NonTemporaryError", func(t *testing.T) {
		regularErr := errors.New("regular error")

		// Should not implement Temporary interface
		if _, ok := regularErr.(Temporary); ok {
			t.Error("Regular error should not implement Temporary")
		}
	})

	t.Run("EdgeCases", func(t *testing.T) {
		// Test empty message
		tempErr := &mock.TemporaryError{Message: ""}
		assert.True(t, tempErr.Temporary())
		assert.Equal(t, "", tempErr.Error())
	})
}
