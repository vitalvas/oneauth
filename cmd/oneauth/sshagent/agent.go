package sshagent

import (
	"io"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/vitalvas/oneauth/internal/netutil"
	"github.com/vitalvas/oneauth/internal/yubikey"
	"golang.org/x/crypto/ssh/agent"
)

var _ agent.ExtendedAgent = &SSHAgent{}

type SSHAgent struct {
	yk   *yubikey.Yubikey
	lock sync.Mutex

	actions       Actions
	log           *logrus.Entry
	agentListener net.Listener
}

type Actions struct {
	BeforeSignHook string
}

func New(serial uint32, log *logrus.Logger, action Actions) (*SSHAgent, error) {
	yk, err := yubikey.OpenBySerial(serial)
	if err != nil {
		return nil, err
	}

	contextLogger := log.WithFields(logrus.Fields{
		"yubikey": serial,
	})

	return &SSHAgent{
		actions: action,
		yk:      yk,
		log:     contextLogger,
	}, nil
}

func (a *SSHAgent) Close() error {
	return nil
}

func (a *SSHAgent) handleConn(conn net.Conn) {
	defer conn.Close()

	creds, err := netutil.UnixSocketCreds(conn)
	if err != nil {
		a.log.Warnln("failed to get unix socket creds:", err)
		return
	}

	if err := netutil.CheckCreds(&creds); err != nil {
		a.log.Warnln(err)
		return
	}

	if err := agent.ServeAgent(a, conn); err != nil && err != io.EOF {
		a.log.Println("Agent client connection ended with error:", err)
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
