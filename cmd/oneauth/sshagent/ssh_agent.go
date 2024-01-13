package sshagent

import (
	"crypto/rand"
	"fmt"
	"os"
	"time"

	"github.com/go-piv/piv-go/piv"
	"github.com/vitalvas/oneauth/internal/tools"
	"github.com/vitalvas/oneauth/internal/yubikey"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func (a *SSHAgent) List() ([]*agent.Key, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	keys := make([]*agent.Key, 0, len(yubikey.AllSSHSlots))

	activeSlots, err := a.yk.GetActiveSlots(yubikey.AllSSHSlots...)
	if err != nil {
		return nil, fmt.Errorf("failed to get active slots: %w", err)
	}

	for _, slot := range activeSlots {
		certPublicKey, err := a.yk.GetCertPublicKey(slot.PIVSlot)
		if err != nil {
			return nil, fmt.Errorf("failed to get public key: %w", err)
		}

		pk, err := ssh.NewPublicKey(certPublicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create ssh public key: %w", err)
		}

		keys = append(keys, &agent.Key{
			Format:  pk.Type(),
			Blob:    pk.Marshal(),
			Comment: fmt.Sprintf("YubiKey #%d PIV Slot %s", a.yk.Serial, slot.PIVSlot.String()),
		})
	}

	return keys, nil
}

func (a *SSHAgent) Sign(reqKey ssh.PublicKey, data []byte) (*ssh.Signature, error) {
	return a.SignWithFlags(reqKey, data, 0)
}

func (a *SSHAgent) SignWithFlags(reqKey ssh.PublicKey, data []byte, flags agent.SignatureFlags) (*ssh.Signature, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	fp := tools.SSHFingerprint(reqKey)

	keys, err := a.yk.ListKeys(yubikey.AllSlots...)
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}

	for _, key := range keys {
		sshPublicKey, err := ssh.NewPublicKey(key.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create ssh public key for sing: %w", err)
		}

		if fp != tools.SSHFingerprint(sshPublicKey) {
			continue
		}

		sig, err := a.sshSign(key, data, flags)
		if err != nil {
			return nil, fmt.Errorf("failed to sign: %w", err)
		}

		return sig, nil
	}

	return nil, fmt.Errorf("unknown key %s", fp)
}

func (a *SSHAgent) sshSign(key yubikey.Cert, data []byte, _ agent.SignatureFlags) (*ssh.Signature, error) {
	_, skip := os.LookupEnv("I_AM_A_REALLY_STUPID_PERSON_WHO_IGNORES_SECURITY_ADVICE")
	if !skip {
		if !key.NotBefore.IsZero() && key.NotBefore.After(time.Now()) {
			return nil, fmt.Errorf("key not yet valid")
		}

		if !key.NotAfter.IsZero() && key.NotAfter.Before(time.Now()) {
			return nil, fmt.Errorf("key expired")
		}
	}

	priv, err := a.yk.PrivateKey(key.Slot.PIVSlot, key.PublicKey, piv.KeyAuth{
		PINPrompt: a.askPINPrompt,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %w", err)
	}

	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer: %w", err)
	}

	return signer.Sign(rand.Reader, data)
}
