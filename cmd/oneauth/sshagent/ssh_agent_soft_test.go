package sshagent

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vitalvas/oneauth/internal/mock"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func TestSSHAgentAdd(t *testing.T) {
	t.Run("AddWhenLocked", func(t *testing.T) {
		a := &SSHAgent{
			log:      logrus.New().WithField("test", "add"),
			softKeys: mock.NewKeystore(),
		}
		assert.NoError(t, a.Lock([]byte("pass")))
		err := a.Add(agent.AddedKey{})
		assert.Error(t, err)
		assert.Equal(t, ErrAgentLocked, err)
	})

	t.Run("AddValidKey", func(t *testing.T) {
		a := &SSHAgent{
			log:      logrus.New().WithField("test", "add"),
			softKeys: mock.NewKeystore(),
		}
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		err = a.Add(agent.AddedKey{PrivateKey: key, Comment: "test-key"})
		assert.NoError(t, err)
	})

	t.Run("AddInvalidKey", func(t *testing.T) {
		a := &SSHAgent{
			log:      logrus.New().WithField("test", "add"),
			softKeys: mock.NewKeystore(),
		}
		err := a.Add(agent.AddedKey{PrivateKey: nil})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Add:")
	})
}

func TestSSHAgentRemove(t *testing.T) {
	t.Run("RemoveWhenLocked", func(t *testing.T) {
		a := &SSHAgent{
			log:      logrus.New().WithField("test", "remove"),
			softKeys: mock.NewKeystore(),
		}
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)
		pubkey, err := ssh.NewPublicKey(&key.PublicKey)
		require.NoError(t, err)

		assert.NoError(t, a.Lock([]byte("pass")))
		err = a.Remove(pubkey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "agent locked")
	})

	t.Run("RemoveReturnsUnsupported", func(t *testing.T) {
		a := &SSHAgent{
			log:      logrus.New().WithField("test", "remove"),
			softKeys: mock.NewKeystore(),
		}
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)
		pubkey, err := ssh.NewPublicKey(&key.PublicKey)
		require.NoError(t, err)

		err = a.Remove(pubkey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "operation unsupported")
	})
}
