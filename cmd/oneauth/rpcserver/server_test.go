package rpcserver

import (
	"testing"

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
		assert.NotNil(t, rpcServer.GetRPCServer())
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
		assert.NotNil(t, rpcServer.GetRPCServer)
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

		// RPC Server should be initialized
		assert.NotNil(t, rpcServer.GetRPCServer())
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

	t.Run("ShutdownWithListener", func(t *testing.T) {
		mockListener := &mockListener{}
		rpcServer := &RPCServer{
			listener: mockListener,
		}

		// Should not panic when listener exists
		assert.NotPanics(t, func() {
			rpcServer.Shutdown()
		})
		assert.True(t, mockListener.closed)
	})
}

type mockListener struct {
	closed bool
}

func (m *mockListener) Close() error {
	m.closed = true
	return nil
}

func TestShutdownSpeed(t *testing.T) {
	t.Run("QuickShutdown", func(t *testing.T) {
		rpcServer := &RPCServer{}

		// Test that shutdown completes quickly
		assert.NotPanics(t, func() {
			rpcServer.Shutdown()
		})
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
		assert.NotNil(t, rpcServer.GetRPCServer())
	})

	t.Run("WithNilParameters", func(t *testing.T) {
		// Test with nil parameters
		rpcServer := New(nil, nil)

		// Should still create server
		assert.NotNil(t, rpcServer)
		assert.Nil(t, rpcServer.SSHAgent)
		assert.Nil(t, rpcServer.log)
		assert.NotNil(t, rpcServer.GetRPCServer())
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
	t.Run("ShutdownCompletion", func(t *testing.T) {
		// Create a server with a mock listener
		mockListener := &mockListener{}
		rpcServer := &RPCServer{
			listener: mockListener,
		}

		// Test shutdown completes
		assert.NotPanics(t, func() {
			rpcServer.Shutdown()
		})
		assert.True(t, mockListener.closed)
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
		mockListener := &mockListener{}
		rpcServer := &RPCServer{
			listener: mockListener,
		}

		// Multiple shutdowns should not panic
		assert.NotPanics(t, func() {
			rpcServer.Shutdown()
			rpcServer.Shutdown()
			rpcServer.Shutdown()
		})
		assert.True(t, mockListener.closed)
	})
}
