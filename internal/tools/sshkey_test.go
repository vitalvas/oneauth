package tools

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"golang.org/x/crypto/ssh"
)

// Sample  key for testing purposes
const (
	samplePublicKey            = "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEgIBOKTbUjfgXzCNhgrth1lu/5RG5Quplafmzei+clUW107vE02RB3ddxGM94gLB5rEpWpcWYPQUZchTZ7r+w0="
	samplePublicKeyFingerprint = "SHA256:S20wyk0CtWZRweGzGtmW/G0PdXGK/ZS1YxOJKHwoFT0"
)

func TestGetSSHPublicKey(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate ecdsa key : %v", err)
	}

	tests := []struct {
		name        string
		inputKey    crypto.PublicKey
		expectedErr bool
	}{
		{
			name:        "ValidPublicKey",
			inputKey:    &key.PublicKey,
			expectedErr: false,
		},
		{
			name:        "InvalidPublicKey",
			inputKey:    nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetSSHPublicKey(tt.inputKey)

			if !tt.expectedErr && err != nil {
				t.Fatalf("Expected no error, but got: %v", err)
			}

			if tt.expectedErr && err == nil {
				t.Fatalf("Expected error: %v, but got no error", tt.expectedErr)
			}

			if !tt.expectedErr && err != nil {
				t.Fatalf("Expected error: %v, but got error: %v", tt.expectedErr, err)
			}

			if !tt.expectedErr && len(result) == 0 {
				t.Fatal("Expected non-empty SSH public key")
			}
		})
	}
}

func TestSSHFingerprint(t *testing.T) {
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(samplePublicKey)) //nolint:dogsled
	if err != nil {
		t.Fatal(err)
	}

	result := SSHFingerprint(pubKey)

	if result != samplePublicKeyFingerprint {
		t.Errorf("Expected fingerprint: %s, Got: %s", samplePublicKeyFingerprint, result)
	}
}
