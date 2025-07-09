package sshagent

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vitalvas/oneauth/internal/mock"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func TestSSHAgent_SoftKeys(t *testing.T) {
	sshAgent := &SSHAgent{
		softKeys: mock.NewKeystore(),
	}

	// Generate test key
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	addedKey := agent.AddedKey{
		PrivateKey: key,
		Comment:    "test key",
	}

	t.Run("AddKey", func(t *testing.T) {
		err := sshAgent.Add(addedKey)
		assert.NoError(t, err)
	})

	t.Run("AddKeyWithLockingBehavior", func(t *testing.T) {
		// Lock agent
		err := sshAgent.Lock([]byte("test-passphrase"))
		assert.NoError(t, err)

		// Adding key should fail when locked
		err = sshAgent.Add(addedKey)
		assert.Error(t, err)
		assert.Equal(t, ErrAgentLocked, err)

		// Unlock agent
		err = sshAgent.Unlock([]byte("test-passphrase"))
		assert.NoError(t, err)

		// Adding key should succeed when unlocked
		err = sshAgent.Add(addedKey)
		assert.NoError(t, err)
	})

	t.Run("RemoveKey", func(t *testing.T) {
		pubkey, err := ssh.NewPublicKey(&key.PublicKey)
		require.NoError(t, err)

		err = sshAgent.Remove(pubkey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "operation unsupported")
	})

	t.Run("AddKeyWithCertificate", func(t *testing.T) {
		// Generate CA key
		caKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		// Generate user key
		userKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		// Create certificate
		userPubKey, err := ssh.NewPublicKey(&userKey.PublicKey)
		require.NoError(t, err)

		cert := &ssh.Certificate{
			Key:         userPubKey,
			Serial:      1,
			CertType:    ssh.UserCert,
			KeyId:       "test-user",
			ValidAfter:  uint64(time.Now().Unix()),
			ValidBefore: uint64(time.Now().Add(time.Hour).Unix()),
		}

		// Sign certificate
		signer, err := ssh.NewSignerFromKey(caKey)
		require.NoError(t, err)

		err = cert.SignCert(rand.Reader, signer)
		require.NoError(t, err)

		// Add key with certificate
		addedKeyWithCert := agent.AddedKey{
			PrivateKey:  userKey,
			Certificate: cert,
			Comment:     "test key with cert",
		}

		err = sshAgent.Add(addedKeyWithCert)
		assert.NoError(t, err)
	})

	t.Run("EdgeCases", func(t *testing.T) {
		// Test with invalid key
		invalidKey := agent.AddedKey{
			PrivateKey: "invalid",
			Comment:    "invalid key",
		}

		err := sshAgent.Add(invalidKey)
		assert.Error(t, err)

		// Test with empty comment
		emptyCommentKey := agent.AddedKey{
			PrivateKey: key,
			Comment:    "",
		}

		err = sshAgent.Add(emptyCommentKey)
		assert.NoError(t, err)
	})
}
