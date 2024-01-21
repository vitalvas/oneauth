package sshagent

import (
	"errors"

	"github.com/vitalvas/oneauth/internal/keyring"
)

var (
	ErrPINNotFound = errors.New("pin not found")
)

func (a *SSHAgent) askPINPrompt() (string, error) {
	pin, err := keyring.Get(keyring.GetYubikeyAccount(a.yk.Serial, "pin"))

	if err == nil {
		a.log.Println("used PIN from keyring")

		return pin, nil
	}

	if err != keyring.ErrNotFound {
		return "", err
	}

	// TODO: prompt for PIN

	return "", ErrPINNotFound
}
