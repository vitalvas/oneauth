package rpclient

import (
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsClient(t *testing.T) {
	t.Run("ValidSocket", func(t *testing.T) {
		validSocketPath := filepath.Join(t.TempDir(), "valid.sock") // must be not longer than 107 characters
		listener, err := net.Listen("unix", validSocketPath)
		if err != nil {
			assert.Nil(t, err)
			return
		}
		defer os.Remove(validSocketPath)
		defer listener.Close()

		if !IsClient(validSocketPath) {
			t.Errorf("Expected IsClient to return true for a valid socket")
		}
	})

	t.Run("InvalidSocket", func(t *testing.T) {
		invalidSocketPath := filepath.Join(t.TempDir(), "invalid.sock")
		file, err := os.Create(invalidSocketPath)
		if err != nil {
			assert.Nil(t, err)
			return
		}
		defer os.Remove(invalidSocketPath)
		defer file.Close()

		if IsClient(invalidSocketPath) {
			t.Errorf("Expected IsClient to return false for an invalid socket")
		}
	})

	t.Run("NonExistentSocket", func(t *testing.T) {
		nonExistentSocketPath := filepath.Join(t.TempDir(), "non_existent_socket.sock")

		if IsClient(nonExistentSocketPath) {
			t.Errorf("Expected IsClient to return false for a non-existent socket")
		}
	})
}
