package sshagent

import "errors"

var (
	ErrOperationUnsupported = errors.New("operation unsupported")
	ErrNoPrivateKey         = errors.New("no private key")
)
