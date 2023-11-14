package tools

import (
	"crypto"

	"golang.org/x/crypto/ssh"
)

func GetSSHPublicKey(key crypto.PublicKey) ([]byte, error) {
	sshKey, err := ssh.NewPublicKey(key)
	if err != nil {
		return nil, err
	}

	return ssh.MarshalAuthorizedKey(sshKey), nil
}
