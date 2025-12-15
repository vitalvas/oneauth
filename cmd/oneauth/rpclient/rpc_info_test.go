package rpclient

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vitalvas/oneauth/cmd/oneauth/rpcserver"
)

type MockAgentService struct{}

func (m *MockAgentService) Info(_ *rpcserver.InfoArgs, reply *rpcserver.InfoReply) error {
	reply.Pid = 12345
	reply.Version = "1.0.0-abcd1234"
	reply.Uptime = "1h30m0s"
	reply.Keys = []rpcserver.InfoKey{
		{
			Name:    "YubiKey 5 NFC",
			Serial:  "12345",
			Version: "5.4.3",
		},
	}
	return nil
}

func TestGetInfo(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		// Create temporary socket
		socketPath := filepath.Join(t.TempDir(), "test.sock")
		listener, err := net.Listen("unix", socketPath)
		assert.NoError(t, err)
		defer os.Remove(socketPath)
		defer listener.Close()

		// Start RPC server
		rpcServer := rpc.NewServer()
		rpcServer.RegisterName("AgentService", &MockAgentService{})

		go func() {
			for {
				conn, err := listener.Accept()
				if err != nil {
					return
				}
				go rpcServer.ServeCodec(jsonrpc.NewServerCodec(conn))
			}
		}()

		// Create client
		client, err := New(socketPath)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		defer client.Close()

		// Test GetInfo
		info, err := client.GetInfo()
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, 12345, info.Pid)
		assert.Equal(t, "1.0.0-abcd1234", info.Version)
		assert.Equal(t, "1h30m0s", info.Uptime)
		assert.Len(t, info.Keys, 1)
		assert.Equal(t, "YubiKey 5 NFC", info.Keys[0].Name)
		assert.Equal(t, "12345", info.Keys[0].Serial)
		assert.Equal(t, "5.4.3", info.Keys[0].Version)
	})

	t.Run("connection error", func(t *testing.T) {
		nonExistentSocketPath := filepath.Join(t.TempDir(), "non_existent.sock")

		client, err := New(nonExistentSocketPath)
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}
