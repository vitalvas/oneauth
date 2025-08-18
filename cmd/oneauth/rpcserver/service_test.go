package rpcserver

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAgentServiceInfo(t *testing.T) {
	t.Run("InfoMethod", func(t *testing.T) {
		// Create a mock RPC server with start time
		server := &RPCServer{
			startTime: time.Now().Add(-1 * time.Hour), // 1 hour ago
		}
		service := &AgentService{server: server}
		args := &InfoArgs{}
		reply := &InfoReply{}

		err := service.Info(args, reply)
		assert.NoError(t, err)
		assert.Equal(t, os.Getpid(), reply.Pid)
		// Version might be empty in tests, but should not be nil
		assert.NotNil(t, reply.Version)
		assert.NotEmpty(t, reply.Uptime)
		assert.Contains(t, reply.Uptime, "h") // Should contain hour indicator
	})

	t.Run("InfoReplyStructure", func(t *testing.T) {
		reply := &InfoReply{
			Pid:     123,
			Version: "1.0.0",
			Uptime:  "1h30m0s",
			Keys:    []InfoKey{},
		}
		assert.Equal(t, 123, reply.Pid)
		assert.Equal(t, "1.0.0", reply.Version)
		assert.Equal(t, "1h30m0s", reply.Uptime)
		assert.NotNil(t, reply.Keys)
	})

	t.Run("InfoArgsStructure", func(t *testing.T) {
		args := &InfoArgs{}
		assert.NotNil(t, args)
	})
}
