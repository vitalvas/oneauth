package sshagent

import "errors"

var (
	ErrOperationUnsupported = errors.New("operation unsupported")
	ErrNoPrivateKey         = errors.New("no private key")

	ErrAgentLocked = errors.New("method is not allowed on agent locked")
)
