package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vitalvas/oneauth/internal/mock"
	"golang.org/x/crypto/ssh"
)

func TestGeneratePrivateHostKey(t *testing.T) {
	t.Run("GeneratesValidSigner", func(t *testing.T) {
		signer, err := generatePrivateHostKey()
		require.NoError(t, err)
		assert.NotNil(t, signer)
		assert.Equal(t, "ecdsa-sha2-nistp256", signer.PublicKey().Type())
	})

	t.Run("GeneratesUniqueKeys", func(t *testing.T) {
		key1, err := generatePrivateHostKey()
		require.NoError(t, err)

		key2, err := generatePrivateHostKey()
		require.NoError(t, err)

		assert.NotEqual(t, key1.PublicKey().Marshal(), key2.PublicKey().Marshal())
	})

	t.Run("CanSign", func(t *testing.T) {
		signer, err := generatePrivateHostKey()
		require.NoError(t, err)

		sig, err := signer.Sign(rand.Reader, []byte("test data"))
		require.NoError(t, err)
		assert.NotNil(t, sig)
	})
}

func TestHandleConn(t *testing.T) {
	t.Run("InvalidConnection", func(t *testing.T) {
		srv := &Server{
			sshConfig: &ssh.ServerConfig{
				ServerVersion: "SSH-2.0-Test",
				NoClientAuth:  true,
			},
		}

		hostKey, err := generatePrivateHostKey()
		require.NoError(t, err)
		srv.sshConfig.AddHostKey(hostKey)

		serverConn, clientConn := net.Pipe()
		defer serverConn.Close()

		clientConn.Close()
		srv.handleConn(serverConn)
		// Should handle gracefully without panicking
	})

	t.Run("ConcurrentConnections", func(t *testing.T) {
		srv := &Server{
			sshConfig: &ssh.ServerConfig{
				ServerVersion: "SSH-2.0-Test",
				NoClientAuth:  true,
			},
		}

		hostKey, err := generatePrivateHostKey()
		require.NoError(t, err)
		srv.sshConfig.AddHostKey(hostKey)

		var wg sync.WaitGroup
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				serverConn, clientConn := net.Pipe()
				defer serverConn.Close()
				go srv.handleConn(serverConn)
				time.Sleep(5 * time.Millisecond)
				clientConn.Close()
			}()
		}
		wg.Wait()
	})
}

func TestHandleChannelSession(t *testing.T) {
	srv := &Server{
		sshConfig: &ssh.ServerConfig{ServerVersion: "SSH-2.0-Test"},
	}

	hostKey, err := generatePrivateHostKey()
	require.NoError(t, err)
	srv.sshConfig.AddHostKey(hostKey)

	t.Run("AcceptError", func(t *testing.T) {
		mockChannel := mock.NewSSHChannel("session").WithAcceptError(fmt.Errorf("accept error"))
		err := srv.handleChannelSession(mockChannel)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not accept channel")
	})

	t.Run("SessionWithNoRequests", func(t *testing.T) {
		conn := mock.NewChannelConn()
		mockChannel := mock.NewSSHChannel("session").
			WithConn(conn).
			WithRequests(mock.MakeRequestSlice([]*ssh.Request{}))

		err := srv.handleChannelSession(mockChannel)
		assert.NoError(t, err)
	})
}

func TestSSHPublicKeyCallback(t *testing.T) {
	t.Run("RejectsRegularPublicKey", func(t *testing.T) {
		srv := &Server{}
		mockConn := &mockConnMeta{
			user:       "testuser",
			remoteAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345},
			sessionID:  []byte("session"),
		}

		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)

		publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
		require.NoError(t, err)

		perms, err := srv.sshPublicKeyCallback(mockConn, publicKey)
		assert.Error(t, err)
		assert.Nil(t, perms)
		assert.Contains(t, err.Error(), "only SSH certificates are supported")
	})
}

type mockConnMeta struct {
	user       string
	remoteAddr net.Addr
	sessionID  []byte
}

func (m *mockConnMeta) User() string          { return m.user }
func (m *mockConnMeta) SessionID() []byte     { return m.sessionID }
func (m *mockConnMeta) ClientVersion() []byte { return []byte("SSH-2.0-Test") }
func (m *mockConnMeta) ServerVersion() []byte { return []byte("SSH-2.0-OneAuth") }
func (m *mockConnMeta) RemoteAddr() net.Addr  { return m.remoteAddr }
func (m *mockConnMeta) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 2022}
}
