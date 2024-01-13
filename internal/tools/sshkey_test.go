package tools

import (
	"testing"

	"golang.org/x/crypto/ssh"
)

// Sample  key for testing purposes
const (
	samplePublicKey            = "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEgIBOKTbUjfgXzCNhgrth1lu/5RG5Quplafmzei+clUW107vE02RB3ddxGM94gLB5rEpWpcWYPQUZchTZ7r+w0="
	samplePublicKeyFingerprint = "SHA256:S20wyk0CtWZRweGzGtmW/G0PdXGK/ZS1YxOJKHwoFT0"
)

func TestSSHFingerprint(t *testing.T) {
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(samplePublicKey))
	if err != nil {
		t.Fatal(err)
	}

	result := SSHFingerprint(pubKey)

	if result != samplePublicKeyFingerprint {
		t.Errorf("Expected fingerprint: %s, Got: %s", samplePublicKeyFingerprint, result)
	}
}
