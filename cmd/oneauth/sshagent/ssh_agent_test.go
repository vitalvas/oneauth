package sshagent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vitalvas/oneauth/internal/mock"
	"golang.org/x/crypto/ssh/agent"
)

func TestSSHAgentImplementsExtendedAgent(t *testing.T) {
	t.Run("InterfaceCompliance", func(_ *testing.T) {
		var _ agent.ExtendedAgent = &SSHAgent{}
	})
}

func TestSSHAgentClose(t *testing.T) {
	t.Run("CloseWithSoftKeys", func(t *testing.T) {
		a := &SSHAgent{
			softKeys: mock.NewKeystore(),
		}
		assert.NoError(t, a.Close())
	})

	t.Run("CloseWithNilSoftKeys", func(t *testing.T) {
		a := &SSHAgent{}
		assert.NoError(t, a.Close())
	})
}

func TestSSHAgentRemoveAll(t *testing.T) {
	t.Run("DelegatesToClose", func(t *testing.T) {
		a := &SSHAgent{
			softKeys: mock.NewKeystore(),
		}
		assert.NoError(t, a.RemoveAll())
	})
}

func TestSSHAgentActions(t *testing.T) {
	t.Run("ActionsStruct", func(t *testing.T) {
		actions := Actions{
			BeforeSignHook: "echo test",
		}
		assert.Equal(t, "echo test", actions.BeforeSignHook)
	})

	t.Run("EmptyActions", func(t *testing.T) {
		actions := Actions{}
		assert.Empty(t, actions.BeforeSignHook)
	})
}
