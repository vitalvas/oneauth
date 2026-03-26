package tools

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
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
