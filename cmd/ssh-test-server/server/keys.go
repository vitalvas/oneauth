package server

import (
	"errors"

	"golang.org/x/crypto/ssh"
)

func (s *Server) sshPasswordCallback(_ ssh.ConnMetadata, _ []byte) (*ssh.Permissions, error) {
	return nil, errors.New("yubikey OTP authentication is not supported yet")
}

func (s *Server) sshPublicKeyCallback(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	checker := ssh.CertChecker{
		IsUserAuthority: func(auth ssh.PublicKey) bool {
			return false
		},
		IsRevoked: func(cert *ssh.Certificate) bool {
			return false
		},
		UserKeyFallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			return nil, errors.New("only SSH certificates are supported")
		},
	}

	return checker.Authenticate(conn, key)
}
