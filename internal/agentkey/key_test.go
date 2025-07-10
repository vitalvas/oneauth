package agentkey

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func createTestAddedKey(t *testing.T, comment string) agent.AddedKey {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	return agent.AddedKey{
		PrivateKey: privateKey,
		Comment:    comment,
	}
}

func TestNewKey(t *testing.T) {
	tests := []struct {
		name    string
		comment string
	}{
		{
			name:    "With comment",
			comment: "test-key-comment",
		},
		{
			name:    "Without comment",
			comment: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addedKey := createTestAddedKey(t, tt.comment)

			key, err := NewKey(addedKey)
			require.NoError(t, err)
			assert.NotNil(t, key)

			// Check fingerprint is set
			assert.NotEmpty(t, key.Fingerprint())

			// Check name is set correctly
			if tt.comment != "" {
				assert.Equal(t, tt.comment, key.name)
			} else {
				assert.Equal(t, key.Fingerprint(), key.name)
			}

			// Check that lastUsed is recent
			assert.True(t, time.Since(key.LastUsed()) < time.Second)

			// Check agentKey is set
			agentKey := key.AgentKey()
			assert.NotNil(t, agentKey)
			assert.Equal(t, tt.comment, agentKey.Comment)
			assert.NotEmpty(t, agentKey.Blob)
			assert.NotEmpty(t, agentKey.Format)
		})
	}
}

func TestNewKey_InvalidPrivateKey(t *testing.T) {
	addedKey := agent.AddedKey{
		PrivateKey: "invalid-key",
		Comment:    "test",
	}

	key, err := NewKey(addedKey)
	assert.Error(t, err)
	assert.Nil(t, key)
}

func TestKey_Fingerprint(t *testing.T) {
	addedKey := createTestAddedKey(t, "test")
	key, err := NewKey(addedKey)
	require.NoError(t, err)

	fingerprint1 := key.Fingerprint()
	fingerprint2 := key.Fingerprint()

	assert.NotEmpty(t, fingerprint1)
	assert.Equal(t, fingerprint1, fingerprint2) // Should be consistent
}

func TestKey_LastUsed(t *testing.T) {
	addedKey := createTestAddedKey(t, "test")
	key, err := NewKey(addedKey)
	require.NoError(t, err)

	initialTime := key.LastUsed()
	assert.True(t, time.Since(initialTime) < time.Second)

	// Sleep a bit and then sign something to update lastUsed
	time.Sleep(10 * time.Millisecond)

	data := []byte("test data")
	_, err = key.Sign(data, 0)
	require.NoError(t, err)

	// LastUsed should not change for default signature (flags == 0)
	assert.Equal(t, initialTime.Unix(), key.LastUsed().Unix())
}

func TestKey_AgentKey(t *testing.T) {
	comment := "test-comment"
	addedKey := createTestAddedKey(t, comment)
	key, err := NewKey(addedKey)
	require.NoError(t, err)

	agentKey := key.AgentKey()
	assert.NotNil(t, agentKey)
	assert.Equal(t, comment, agentKey.Comment)
	assert.NotEmpty(t, agentKey.Blob)
	assert.NotEmpty(t, agentKey.Format)
}

func TestKey_Sign_DefaultFlags(t *testing.T) {
	addedKey := createTestAddedKey(t, "test")
	key, err := NewKey(addedKey)
	require.NoError(t, err)

	data := []byte("test data to sign")
	signature, err := key.Sign(data, 0)

	assert.NoError(t, err)
	assert.NotNil(t, signature)
	assert.NotEmpty(t, signature.Blob)
	assert.NotEmpty(t, signature.Format)
}

func TestKey_Sign_RSASha256(t *testing.T) {
	addedKey := createTestAddedKey(t, "test")
	key, err := NewKey(addedKey)
	require.NoError(t, err)

	data := []byte("test data to sign")
	initialLastUsed := key.LastUsed()

	// Wait a moment to ensure time difference
	time.Sleep(10 * time.Millisecond)

	signature, err := key.Sign(data, agent.SignatureFlagRsaSha256)

	assert.NoError(t, err)
	assert.NotNil(t, signature)
	assert.Equal(t, ssh.KeyAlgoRSASHA256, signature.Format)

	// LastUsed should be updated for non-default signatures
	assert.True(t, key.LastUsed().After(initialLastUsed))
}

func TestKey_Sign_RSASha512(t *testing.T) {
	addedKey := createTestAddedKey(t, "test")
	key, err := NewKey(addedKey)
	require.NoError(t, err)

	data := []byte("test data to sign")
	signature, err := key.Sign(data, agent.SignatureFlagRsaSha512)

	assert.NoError(t, err)
	assert.NotNil(t, signature)
	assert.Equal(t, ssh.KeyAlgoRSASHA512, signature.Format)
}

func TestKey_Sign_UnsupportedFlags(t *testing.T) {
	addedKey := createTestAddedKey(t, "test")
	key, err := NewKey(addedKey)
	require.NoError(t, err)

	data := []byte("test data to sign")
	signature, err := key.Sign(data, agent.SignatureFlags(999)) // Invalid flags

	assert.Error(t, err)
	assert.Nil(t, signature)
	assert.Contains(t, err.Error(), "unsupported signature flags")
}

func TestKey_ConsistentFingerprints(t *testing.T) {
	// Create two keys from the same private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	addedKey1 := agent.AddedKey{
		PrivateKey: privateKey,
		Comment:    "comment1",
	}
	addedKey2 := agent.AddedKey{
		PrivateKey: privateKey,
		Comment:    "comment2",
	}

	key1, err := NewKey(addedKey1)
	require.NoError(t, err)

	key2, err := NewKey(addedKey2)
	require.NoError(t, err)

	// Same private key should produce same fingerprint
	assert.Equal(t, key1.Fingerprint(), key2.Fingerprint())
}

func TestKey_DifferentFingerprints(t *testing.T) {
	addedKey1 := createTestAddedKey(t, "key1")
	addedKey2 := createTestAddedKey(t, "key2")

	key1, err := NewKey(addedKey1)
	require.NoError(t, err)

	key2, err := NewKey(addedKey2)
	require.NoError(t, err)

	// Different private keys should produce different fingerprints
	assert.NotEqual(t, key1.Fingerprint(), key2.Fingerprint())
}
