package keyring

import (
	"errors"
	"time"

	"github.com/zalando/go-keyring"
)

// origin: https://github.com/cli/cli/blob/06c36a74c2d94083111c3a585dc7b7323102602b/internal/keyring/keyring.go

var (
	ErrNotFound            = errors.New("secret not found in keyring")
	ErrTimeoutSetSecret    = errors.New("timeout while trying to set secret in keyring")
	ErrTimeoutGetSecret    = errors.New("timeout while trying to get secret from keyring")
	ErrTimeoutDeleteSecret = errors.New("timeout while trying to delete secret from keyring")

	opsTimeout = 3 * time.Second
)

const (
	serviceName = "dev.vitalvas.oneauth"
)

// Set secret in keyring for user.
func Set(user, secret string) error {
	ch := make(chan error, 1)

	go func() {
		defer close(ch)
		ch <- keyring.Set(serviceName, user, secret)
	}()

	select {
	case err := <-ch:
		return err

	case <-time.After(opsTimeout):
		return ErrTimeoutSetSecret
	}
}

// Get secret from keyring given service and user name.
func Get(user string) (string, error) {
	ch := make(chan struct {
		val string
		err error
	}, 1)

	go func() {
		defer close(ch)
		val, err := keyring.Get(serviceName, user)
		ch <- struct {
			val string
			err error
		}{val, err}
	}()

	select {
	case res := <-ch:
		if errors.Is(res.err, keyring.ErrNotFound) {
			return "", ErrNotFound
		}
		return res.val, res.err

	case <-time.After(opsTimeout):
		return "", ErrTimeoutGetSecret
	}
}

// Delete secret from keyring.
func Delete(user string) error {
	ch := make(chan error, 1)

	go func() {
		defer close(ch)
		ch <- keyring.Delete(serviceName, user)
	}()

	select {
	case err := <-ch:
		return err

	case <-time.After(opsTimeout):
		return ErrTimeoutDeleteSecret
	}
}
