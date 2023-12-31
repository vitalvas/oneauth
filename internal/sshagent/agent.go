package sshagent

import (
	"io"
	"log"
	"net"
	"sync"

	"github.com/vitalvas/oneauth/internal/netutil"
	"github.com/vitalvas/oneauth/internal/yubikey"
	"golang.org/x/crypto/ssh/agent"
)

var _ agent.ExtendedAgent = &SSHAgent{}

type SSHAgent struct {
	yk   *yubikey.Yubikey
	lock sync.Mutex

	agentListener net.Listener
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

func (a *SSHAgent) handleConn(conn net.Conn) {
	defer conn.Close()

	creds, err := netutil.UnixSocketCreds(conn)
	if err != nil {
		log.Println("failed to get unix socket creds:", err)
		return
	}

	if err := netutil.CheckCreds(&creds); err != nil {
		log.Println(err)
		return
	}

	if err := agent.ServeAgent(a, conn); err != nil && err != io.EOF {
		log.Println("Agent client connection ended with error:", err)
	}
}

func (a *SSHAgent) Shutdown() error {
	if err := a.Close(); err != nil {
		return err
	}

	if a.yk != nil {
		if err := a.yk.Close(); err != nil {
			return err
		}
	}

	if a.agentListener != nil {
		if err := a.agentListener.Close(); err != nil {
			return err
		}
	}

	return nil
}
