package sshagent

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/vitalvas/oneauth/internal/mock"
)

func TestSSHAgentLock(t *testing.T) {
	t.Run("LockWithValidPassphrase", func(t *testing.T) {
		a := &SSHAgent{
			log:      logrus.New().WithField("test", "lock"),
			softKeys: mock.NewKeystore(),
		}
		err := a.Lock([]byte("secret"))
		assert.NoError(t, err)
		assert.NotNil(t, a.lockPassphrase)
	})

	t.Run("LockAlreadyLocked", func(t *testing.T) {
		a := &SSHAgent{
			log:      logrus.New().WithField("test", "lock"),
			softKeys: mock.NewKeystore(),
		}
		assert.NoError(t, a.Lock([]byte("secret")))
		err := a.Lock([]byte("another"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "agent locked")
	})

	t.Run("LockWithNilPassphrase", func(t *testing.T) {
		a := &SSHAgent{
			log:      logrus.New().WithField("test", "lock"),
			softKeys: mock.NewKeystore(),
		}
		err := a.Lock(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no private key")
	})

	t.Run("LockWithEmptyPassphrase", func(t *testing.T) {
		a := &SSHAgent{
			log:      logrus.New().WithField("test", "lock"),
			softKeys: mock.NewKeystore(),
		}
		err := a.Lock([]byte{})
		assert.NoError(t, err)
	})
}

func TestSSHAgentUnlock(t *testing.T) {
	t.Run("UnlockWithCorrectPassphrase", func(t *testing.T) {
		a := &SSHAgent{
			log:      logrus.New().WithField("test", "unlock"),
			softKeys: mock.NewKeystore(),
		}
		passphrase := []byte("secret")
		assert.NoError(t, a.Lock(passphrase))
		assert.NoError(t, a.Unlock(passphrase))
		assert.Nil(t, a.lockPassphrase)
	})

	t.Run("UnlockWithWrongPassphrase", func(t *testing.T) {
		a := &SSHAgent{
			log:      logrus.New().WithField("test", "unlock"),
			softKeys: mock.NewKeystore(),
		}
		assert.NoError(t, a.Lock([]byte("secret")))
		err := a.Unlock([]byte("wrong"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect passphrase")
	})

	t.Run("UnlockNotLocked", func(t *testing.T) {
		a := &SSHAgent{
			log:      logrus.New().WithField("test", "unlock"),
			softKeys: mock.NewKeystore(),
		}
		err := a.Unlock([]byte("secret"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can't unlock not locked agent")
	})
}
