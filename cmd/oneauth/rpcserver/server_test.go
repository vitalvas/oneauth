package rpcserver

import (
	"net/http"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/vitalvas/oneauth/cmd/oneauth/sshagent"
)

func TestNew(t *testing.T) {
	t.Run("CreateRPCServer", func(t *testing.T) {
		// Create mock SSH agent
		sshAgent := &sshagent.SSHAgent{}
		log := logrus.New()
		
		// Create RPC server
		rpcServer := New(sshAgent, log)
		
		// Verify initialization
		assert.NotNil(t, rpcServer)
		assert.Equal(t, sshAgent, rpcServer.SSHAgent)
		assert.Equal(t, log, rpcServer.log)
		assert.Nil(t, rpcServer.GetServer())
	})
}

func TestRPCServerType(t *testing.T) {
	t.Run("TypeVerification", func(t *testing.T) {
		// Create minimal RPC server
		rpcServer := &RPCServer{}
		
		// Verify type
		assert.IsType(t, &RPCServer{}, rpcServer)
		
		// Verify fields exist
		assert.NotNil(t, &rpcServer.SSHAgent)
		assert.NotNil(t, rpcServer.GetServer)
		assert.NotNil(t, &rpcServer.log)
	})
}

func TestRPCServerFields(t *testing.T) {
	t.Run("FieldAccess", func(t *testing.T) {
		sshAgent := &sshagent.SSHAgent{}
		log := logrus.New()
		
		rpcServer := New(sshAgent, log)
		
		// Test field access
		assert.Equal(t, sshAgent, rpcServer.SSHAgent)
		assert.Equal(t, log, rpcServer.log)
		
		// Server should initially be nil
		assert.Nil(t, rpcServer.GetServer())
	})
}

func TestShutdown(t *testing.T) {
	t.Run("ShutdownWithNilServer", func(t *testing.T) {
		rpcServer := &RPCServer{}
		
		// Should not panic when server is nil
		assert.NotPanics(t, func() {
			rpcServer.Shutdown()
		})
	})
	
	t.Run("ShutdownWithServer", func(t *testing.T) {
		rpcServer := &RPCServer{
			server: &http.Server{},
		}
		
		// Should not panic when server exists
		assert.NotPanics(t, func() {
			rpcServer.Shutdown()
		})
	})
}

func TestShutdownTimeout(t *testing.T) {
	t.Run("TimeoutContextCreation", func(t *testing.T) {
		rpcServer := &RPCServer{}
		
		// Test that shutdown creates proper context
		start := time.Now()
		rpcServer.Shutdown()
		elapsed := time.Since(start)
		
		// Should complete quickly when server is nil
		assert.Less(t, elapsed, 100*time.Millisecond)
	})
}

func TestRPCServerConstruction(t *testing.T) {
	t.Run("WithAllParameters", func(t *testing.T) {
		sshAgent := &sshagent.SSHAgent{}
		log := logrus.New()
		
		rpcServer := New(sshAgent, log)
		
		// Verify all parameters are set
		assert.NotNil(t, rpcServer.SSHAgent)
		assert.NotNil(t, rpcServer.log)
		assert.Nil(t, rpcServer.GetServer())
	})
	
	t.Run("WithNilParameters", func(t *testing.T) {
		// Test with nil parameters
		rpcServer := New(nil, nil)
		
		// Should still create server
		assert.NotNil(t, rpcServer)
		assert.Nil(t, rpcServer.SSHAgent)
		assert.Nil(t, rpcServer.log)
	})
}

func TestRPCServerMethods(t *testing.T) {
	t.Run("MethodExistence", func(t *testing.T) {
		sshAgent := &sshagent.SSHAgent{}
		log := logrus.New()
		
		rpcServer := New(sshAgent, log)
		
		// Test that methods exist and can be called
		assert.NotPanics(t, func() {
			rpcServer.Shutdown()
		})
	})
}

func TestRPCServerShutdownBehavior(t *testing.T) {
	t.Run("ShutdownCancellation", func(t *testing.T) {
		// Create a server with a mock HTTP server
		mockServer := &http.Server{}
		rpcServer := &RPCServer{
			server: mockServer,
		}
		
		// Test shutdown doesn't hang
		done := make(chan bool)
		go func() {
			rpcServer.Shutdown()
			done <- true
		}()
		
		select {
		case <-done:
			// Success - shutdown completed
		case <-time.After(10 * time.Second):
			t.Fatal("Shutdown took too long")
		}
	})
}

func TestRPCServerEdgeCases(t *testing.T) {
	t.Run("EmptyStruct", func(t *testing.T) {
		// Test with empty struct
		rpcServer := &RPCServer{}
		
		// Should not panic
		assert.NotPanics(t, func() {
			rpcServer.Shutdown()
		})
	})
	
	t.Run("RepeatedShutdown", func(t *testing.T) {
		rpcServer := &RPCServer{
			server: &http.Server{},
		}
		
		// Multiple shutdowns should not panic
		assert.NotPanics(t, func() {
			rpcServer.Shutdown()
			rpcServer.Shutdown()
			rpcServer.Shutdown()
		})
	})
}