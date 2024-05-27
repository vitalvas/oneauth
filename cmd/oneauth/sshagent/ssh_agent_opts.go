package sshagent

import (
	"crypto/subtle"
	"errors"
	"fmt"

	"github.com/vitalvas/oneauth/internal/tools"
)

func (a *SSHAgent) Lock(passphrase []byte) error {
	if a.lockPassphrase != nil {
		return fmt.Errorf("Lock: %w", ErrAgentLocked)
	}

	if passphrase == nil {
		return fmt.Errorf("Lock: %w", ErrNoPrivateKey)
	}

	a.lockPassphrase = tools.EncodePassphrase(passphrase)

	return nil
}

func (a *SSHAgent) Unlock(passphrase []byte) error {
	if a.lockPassphrase == nil {
		return errors.New("can't unlock not locked agent")
	}

	passphraseEncoded := tools.EncodePassphrase(passphrase)
	if subtle.ConstantTimeCompare(passphraseEncoded, a.lockPassphrase) != 1 {
		return errors.New("incorrect passphrase")
	}

	a.lockPassphrase = nil

	return nil
}
