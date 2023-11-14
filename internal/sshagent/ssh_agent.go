package sshagent

import (
	"fmt"

	"github.com/vitalvas/oneauth/internal/yubikey"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func (a *SSHAgent) List() ([]*agent.Key, error) {

	keys := make([]*agent.Key, 0, len(yubikey.AllSSHSlots))

	for _, slot := range yubikey.AllSSHSlots {

		certPublicKey, err := a.yk.GetCertPublicKey(slot.PIVSlot)
		if err != nil {
			return nil, fmt.Errorf("failed to get public key: %w", err)
		}

		pk, err := ssh.NewPublicKey(certPublicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create ssh public key: %w", err)
		}

		keys = append(keys, &agent.Key{
			Format: pk.Type(),
			Blob:   pk.Marshal(),
		})
	}

	return keys, nil
}
