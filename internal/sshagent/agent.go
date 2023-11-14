package sshagent

import (
	"io"
	"log"
	"net"
	"sync"

	"github.com/vitalvas/oneauth/internal/yubikey"
	"golang.org/x/crypto/ssh/agent"
)

var _ agent.ExtendedAgent = &SSHAgent{}

type SSHAgent struct {
	yk   *yubikey.Yubikey
	lock sync.Mutex
}

func New(serial uint32) (*SSHAgent, error) {
	yk, err := yubikey.OpenBySerial(serial)
	if err != nil {
		return nil, err
	}

	return &SSHAgent{
		yk: yk,
	}, nil
}

func (a *SSHAgent) Close() error {
	return nil
}

func (a *SSHAgent) HandleConn(conn net.Conn) {
	if err := agent.ServeAgent(a, conn); err != nil && err != io.EOF {
		log.Println("Agent client connection ended with error:", err)
	}
}
