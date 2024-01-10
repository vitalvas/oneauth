package sshagent

import (
	"fmt"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func (a *SSHAgent) Add(_ agent.AddedKey) error {
	return fmt.Errorf("Add: %w", ErrOperationUnsupported)
}

func (a *SSHAgent) Lock(_ []byte) error {
	return fmt.Errorf("Lock: %w", ErrOperationUnsupported)
}

func (a *SSHAgent) Unlock(_ []byte) error {
	return fmt.Errorf("Unlock: %w", ErrOperationUnsupported)
}

func (a *SSHAgent) Remove(_ ssh.PublicKey) error {
	return fmt.Errorf("Remove: %w", ErrOperationUnsupported)
}

func (a *SSHAgent) RemoveAll() error {
	return a.Close()
}

func (a *SSHAgent) Signers() ([]ssh.Signer, error) {
	return nil, fmt.Errorf("Signers: %w", ErrOperationUnsupported)
}

func (a *SSHAgent) Extension(_ string, _ []byte) ([]byte, error) {
	return nil, agent.ErrExtensionUnsupported
}
