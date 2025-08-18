package rpclient

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("ValidSocket", func(t *testing.T) {
		validSocketPath := filepath.Join(t.TempDir(), "valid.sock")
		listener, err := net.Listen("unix", validSocketPath)
		if err != nil {
			assert.NoError(t, err)
			return
		}
		defer os.Remove(validSocketPath)
		defer listener.Close()

		// Start a simple RPC server
		go func() {
			for {
				conn, err := listener.Accept()
				if err != nil {
					return
				}
				rpc.ServeCodec(jsonrpc.NewServerCodec(conn))
			}
		}()

		client, err := New(validSocketPath)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		if client != nil {
			client.Close()
		}
	})

	t.Run("InvalidSocket", func(t *testing.T) {
		invalidSocketPath := filepath.Join(t.TempDir(), "invalid.sock")
		file, err := os.Create(invalidSocketPath)
		if err != nil {
			assert.NoError(t, err)
			return
		}
		defer os.Remove(invalidSocketPath)
		defer file.Close()

		client, err := New(invalidSocketPath)
		assert.Error(t, err)
		assert.Nil(t, client)
	})

	t.Run("NonExistentSocket", func(t *testing.T) {
		nonExistentSocketPath := filepath.Join(t.TempDir(), "non_existent_socket.sock")

		client, err := New(nonExistentSocketPath)
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}

func TestIsClient(t *testing.T) {
	t.Run("ValidSocket", func(t *testing.T) {
		validSocketPath := filepath.Join(t.TempDir(), "valid.sock")
		listener, err := net.Listen("unix", validSocketPath)
		if err != nil {
			assert.NoError(t, err)
			return
		}
		defer os.Remove(validSocketPath)
		defer listener.Close()

		assert.True(t, IsClient(validSocketPath))
	})

	t.Run("InvalidSocket", func(t *testing.T) {
		invalidSocketPath := filepath.Join(t.TempDir(), "invalid.sock")
		file, err := os.Create(invalidSocketPath)
		if err != nil {
			assert.NoError(t, err)
			return
		}
		defer os.Remove(invalidSocketPath)
		defer file.Close()

		assert.False(t, IsClient(invalidSocketPath))
	})

	t.Run("NonExistentSocket", func(t *testing.T) {
		nonExistentSocketPath := filepath.Join(t.TempDir(), "non_existent_socket.sock")
		assert.False(t, IsClient(nonExistentSocketPath))
	})

	t.Run("CloseClient", func(t *testing.T) {
		client := &Client{}
		err := client.Close()
		assert.NoError(t, err)
	})
}
