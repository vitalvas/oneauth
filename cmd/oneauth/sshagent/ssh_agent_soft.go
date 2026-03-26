package sshagent

import (
	"fmt"

	"github.com/vitalvas/oneauth/internal/agentkey"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func (a *SSHAgent) Add(newKey agent.AddedKey) error {
	if a.lockPassphrase != nil {
		return ErrAgentLocked
	}

	key, err := agentkey.NewKey(newKey)
	if err != nil {
		return fmt.Errorf("Add: %w", err)
	}

	a.softKeys.Add(key)

	return nil
}

func (a *SSHAgent) Remove(reqKey ssh.PublicKey) error {
	if a.lockPassphrase != nil {
		return ErrAgentLocked
	}

	fp := ssh.FingerprintSHA256(reqKey)
	a.softKeys.Remove(fp)

	return fmt.Errorf("Remove: %w", ErrOperationUnsupported)
}
