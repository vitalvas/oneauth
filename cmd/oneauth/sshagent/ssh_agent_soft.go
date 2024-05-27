package sshagent

import (
	"fmt"

	"github.com/vitalvas/oneauth/internal/agentkey"
	"github.com/vitalvas/oneauth/internal/tools"
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

	fp := tools.SSHFingerprint(reqKey)
	a.softKeys.Remove(fp)

	return fmt.Errorf("Remove: %w", ErrOperationUnsupported)
}
