package sshagent

import (
	"fmt"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func (a *SSHAgent) Signers() ([]ssh.Signer, error) {
	return nil, fmt.Errorf("Signers: %w", ErrOperationUnsupported)
}

func (a *SSHAgent) Extension(_ string, _ []byte) ([]byte, error) {
	return nil, agent.ErrExtensionUnsupported
}
