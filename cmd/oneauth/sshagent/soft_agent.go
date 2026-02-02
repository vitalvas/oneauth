package sshagent

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vitalvas/oneauth/internal/agentkey"
	"github.com/vitalvas/oneauth/internal/keystore"
	"github.com/vitalvas/oneauth/internal/netutil"
	"github.com/vitalvas/oneauth/internal/tools"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

var _ agent.ExtendedAgent = &SoftAgent{}

// SoftAgent is a lightweight SSH agent that only supports software keys (no YubiKey)
type SoftAgent struct {
	name string
	lock sync.Mutex
	log  *logrus.Entry

	agentListener  net.Listener
	lockPassphrase []byte
	softKeys       *keystore.Store
}

// NewSoftAgent creates a new soft-key-only SSH agent
func NewSoftAgent(name string, keepKeySeconds int64, log *logrus.Logger) *SoftAgent {
	contextLogger := log.WithFields(logrus.Fields{
		"agent": name,
	})

	return &SoftAgent{
		name:     name,
		log:      contextLogger,
		softKeys: keystore.New(keepKeySeconds),
	}
}

func (a *SoftAgent) Close() error {
	if a.softKeys != nil {
		a.softKeys.RemoveAll()
	}
	return nil
}

func (a *SoftAgent) getListener() net.Listener {
	a.lock.Lock()
	defer a.lock.Unlock()
	return a.agentListener
}

func (a *SoftAgent) setListener(listener net.Listener) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.agentListener = listener
}

func (a *SoftAgent) handleConn(conn net.Conn) {
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

func (a *SoftAgent) ListenAndServe(ctx context.Context, socketPath string) error {
	defer func() {
		if _, err := os.Stat(socketPath); err == nil {
			os.Remove(socketPath)
		}
	}()

	a.log.Println("listening ssh-agent on", socketPath)

	if a.getListener() == nil {
		if _, err := os.Stat(socketPath); err == nil {
			os.Remove(socketPath)
		}

		listener, err := net.Listen("unix", socketPath)
		if err != nil {
			return fmt.Errorf("failed to listen: %w", err)
		}

		a.setListener(listener)

		if err := os.Chmod(socketPath, 0600); err != nil {
			return fmt.Errorf("failed to chmod: %w", err)
		}

		defer func() {
			if l := a.getListener(); l != nil {
				l.Close()
			}
		}()
	}

	go func() {
		<-ctx.Done()
		if l := a.getListener(); l != nil {
			l.Close()
		}
	}()

	for {
		listener := a.getListener()
		if listener == nil {
			return fmt.Errorf("listener is nil")
		}

		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
			}

			if errors.Is(err, net.ErrClosed) || errors.Is(err, io.ErrClosedPipe) || errors.Is(err, io.EOF) {
				return nil
			}

			if err, ok := err.(Temporary); ok && err.Temporary() {
				a.log.Printf("temporary accept error: %v", err)
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(100 * time.Millisecond):
					continue
				}
			}

			return fmt.Errorf("failed to accept: %w", err)
		}

		go a.handleConn(conn)
	}
}

func (a *SoftAgent) Shutdown() error {
	if err := a.Close(); err != nil {
		return err
	}

	if listener := a.getListener(); listener != nil {
		if err := listener.Close(); err != nil {
			return err
		}
	}

	return nil
}

// List returns all keys in the soft key store
func (a *SoftAgent) List() ([]*agent.Key, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	if a.lockPassphrase != nil {
		return nil, ErrAgentLocked
	}

	keys := make([]*agent.Key, 0, a.softKeys.Len())

	for _, key := range a.softKeys.List() {
		keys = append(keys, key.AgentKey())
	}

	return keys, nil
}

func (a *SoftAgent) Sign(reqKey ssh.PublicKey, data []byte) (*ssh.Signature, error) {
	return a.SignWithFlags(reqKey, data, 0)
}

func (a *SoftAgent) SignWithFlags(reqKey ssh.PublicKey, data []byte, flags agent.SignatureFlags) (*ssh.Signature, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	if a.lockPassphrase != nil {
		return nil, ErrAgentLocked
	}

	fp := tools.SSHFingerprint(reqKey)
	dataHash := tools.FastHash(data)

	a.log.Println("request to sign payload:", dataHash)

	if key, ok := a.softKeys.Get(fp); ok {
		sig, err := key.Sign(data, flags)
		if err != nil {
			return nil, err
		}

		a.log.Println("signed payload:", dataHash)
		return sig, nil
	}

	return nil, fmt.Errorf("unknown key %s", fp)
}

func (a *SoftAgent) Add(newKey agent.AddedKey) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	if a.lockPassphrase != nil {
		return ErrAgentLocked
	}

	key, err := agentkey.NewKey(newKey)
	if err != nil {
		return fmt.Errorf("Add: %w", err)
	}

	a.softKeys.Add(key)

	return nil
}

func (a *SoftAgent) Remove(reqKey ssh.PublicKey) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	if a.lockPassphrase != nil {
		return ErrAgentLocked
	}

	fp := tools.SSHFingerprint(reqKey)
	a.softKeys.Remove(fp)

	return nil
}

func (a *SoftAgent) RemoveAll() error {
	a.lock.Lock()
	defer a.lock.Unlock()

	return a.Close()
}

func (a *SoftAgent) Lock(passphrase []byte) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	if a.lockPassphrase != nil {
		return fmt.Errorf("Lock: %w", ErrAgentLocked)
	}

	if passphrase == nil {
		return fmt.Errorf("Lock: %w", ErrNoPrivateKey)
	}

	a.lockPassphrase = tools.EncodePassphrase(passphrase)

	return nil
}

func (a *SoftAgent) Unlock(passphrase []byte) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	if a.lockPassphrase == nil {
		return errors.New("can't unlock not locked agent")
	}

	passphraseEncoded := tools.EncodePassphrase(passphrase)
	if subtle.ConstantTimeCompare(passphraseEncoded, a.lockPassphrase) != 1 {
		return errors.New("incorrect passphrase")
	}

	a.lockPassphrase = nil

	return nil
}

func (a *SoftAgent) Signers() ([]ssh.Signer, error) {
	return nil, fmt.Errorf("Signers: %w", ErrOperationUnsupported)
}

func (a *SoftAgent) Extension(_ string, _ []byte) ([]byte, error) {
	return nil, agent.ErrExtensionUnsupported
}
