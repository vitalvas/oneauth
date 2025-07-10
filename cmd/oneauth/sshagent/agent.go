package sshagent

import (
	"io"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/vitalvas/oneauth/cmd/oneauth/config"
	"github.com/vitalvas/oneauth/internal/keystore"
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

	lockPassphrase []byte

	softKeys *keystore.Store
}

type Actions struct {
	BeforeSignHook string
}

func New(serial uint32, log *logrus.Logger, config *config.Config) (*SSHAgent, error) {
	yk, err := yubikey.OpenBySerial(serial)
	if err != nil {
		return nil, err
	}

	contextLogger := log.WithFields(logrus.Fields{
		"yubikey": serial,
	})

	return &SSHAgent{
		actions: Actions{
			BeforeSignHook: config.Keyring.BeforeSignHook,
		},
		yk:  yk,
		log: contextLogger,

		softKeys: keystore.New(config.Keyring.KeepKeySeconds),
	}, nil
}

func (a *SSHAgent) Close() error {
	if a.softKeys != nil {
		a.softKeys.RemoveAll()
	}

	return nil
}

// getListener safely returns the current listener
func (a *SSHAgent) getListener() net.Listener {
	a.lock.Lock()
	defer a.lock.Unlock()
	return a.agentListener
}

// setListener safely sets the listener
func (a *SSHAgent) setListener(listener net.Listener) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.agentListener = listener
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

	if listener := a.getListener(); listener != nil {
		if err := listener.Close(); err != nil {
			return err
		}
	}

	return nil
}
