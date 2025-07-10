package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vitalvas/oneauth/internal/mock"
	"golang.org/x/crypto/ssh"
)

// Test core server functionality
func TestServer(t *testing.T) {
	t.Run("BasicOperations", func(t *testing.T) {
		srv := &Server{}
		assert.NotNil(t, srv)
		assert.Nil(t, srv.serverURL)
		assert.Nil(t, srv.sshConfig)

		// Test URL parsing
		srv.serverURL, _ = url.Parse("http://test.example.com:9000")
		assert.Equal(t, "http", srv.serverURL.Scheme)
		assert.Equal(t, "test.example.com:9000", srv.serverURL.Host)

		// Test invalid URL
		_, err := url.Parse("://invalid-url") //nolint:staticcheck // Testing invalid URL parsing
		assert.Error(t, err)
	})
}

// Test SSH authentication
func TestSSHAuthentication(t *testing.T) {
	srv := &Server{
		serverURL: &url.URL{Scheme: "http", Host: "localhost:8080"},
	}

	mockConn := &mockConnMetadata{
		user:       "testuser",
		remoteAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345},
		sessionID:  []byte("test-session"),
	}

	t.Run("PasswordAuth_Success", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			response := YubikeyOTPVerifyResponse{OTP: "testotp", Serial: 12345}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer mockServer.Close()

		mockURL, _ := url.Parse(mockServer.URL)
		srv.serverURL = mockURL

		perms, err := srv.sshPasswordCallback(mockConn, []byte("testotp"))

		require.NoError(t, err)
		assert.NotNil(t, perms)
		assert.Equal(t, "yubikey-otp", perms.Extensions["auth-type"])
		assert.Equal(t, "12345", perms.Extensions["yubikey-serial"])
	})

	t.Run("PasswordAuth_Failure", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer mockServer.Close()

		mockURL, _ := url.Parse(mockServer.URL)
		srv.serverURL = mockURL

		perms, err := srv.sshPasswordCallback(mockConn, []byte("invalid"))
		assert.Error(t, err)
		assert.Nil(t, perms)
	})

	t.Run("PublicKeyAuth", func(t *testing.T) {
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)

		publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
		require.NoError(t, err)

		// Should reject regular public keys (only certificates supported)
		perms, err := srv.sshPublicKeyCallback(mockConn, publicKey)
		assert.Error(t, err)
		assert.Nil(t, perms)
		assert.Contains(t, err.Error(), "only SSH certificates are supported")
	})
}

// Test key generation
func TestKeyGeneration(t *testing.T) {
	t.Run("GenerateHostKey", func(t *testing.T) {
		signer, err := generatePrivateHostKey()

		require.NoError(t, err)
		assert.NotNil(t, signer)
		assert.Equal(t, "ecdsa-sha2-nistp256", signer.PublicKey().Type())

		// Test signing capability
		data := []byte("test data")
		signature, err := signer.Sign(rand.Reader, data)
		require.NoError(t, err)
		assert.NotNil(t, signature)
	})

	t.Run("MultipleKeys", func(t *testing.T) {
		key1, err := generatePrivateHostKey()
		require.NoError(t, err)

		key2, err := generatePrivateHostKey()
		require.NoError(t, err)

		// Keys should be different
		assert.NotEqual(t, key1.PublicKey().Marshal(), key2.PublicKey().Marshal())
	})
}

// Test SSH channel handling
func TestChannelHandling(t *testing.T) {
	srv := &Server{
		sshConfig: &ssh.ServerConfig{ServerVersion: "SSH-2.0-Test"},
	}

	hostKey, err := generatePrivateHostKey()
	require.NoError(t, err)
	srv.sshConfig.AddHostKey(hostKey)

	t.Run("ChannelTypes", func(t *testing.T) {
		testCases := []struct {
			channelType  string
			expectReject bool
		}{
			{"session", false},
			{"direct-tcpip", true},
			{"unknown", true},
		}

		for _, tc := range testCases {
			mockChannel := mock.NewSSHChannel(tc.channelType).
				WithConn(mock.NewChannelConn()).
				WithRequests(mock.MakeRequestSlice([]*ssh.Request{}))

			srv.handleChannels(mock.MakeNewChannelSlice([]*mock.NewChannel{mockChannel}))

			if tc.expectReject {
				assert.True(t, mockChannel.IsRejected())
				assert.Equal(t, ssh.UnknownChannelType, mockChannel.RejectReason())
			} else {
				assert.True(t, mockChannel.IsAccepted())
			}
		}
	})

	t.Run("SessionAcceptFailure", func(t *testing.T) {
		mockChannel := mock.NewSSHChannel("session").WithAcceptError(fmt.Errorf("accept failed"))
		err := srv.handleChannelSession(mockChannel)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not accept channel")
	})

	t.Run("RequestHandling", func(_ *testing.T) {
		mockReq := &ssh.Request{Type: "keepalive", WantReply: false, Payload: []byte("test")}
		srv.handleRequests(mock.MakeRequestSlice([]*ssh.Request{mockReq}))
		// Should handle gracefully without error
	})
}

// Test connection handling
func TestConnectionHandling(t *testing.T) {
	srv := &Server{
		sshConfig: &ssh.ServerConfig{
			ServerVersion: "SSH-2.0-Test",
			NoClientAuth:  true,
		},
	}

	hostKey, err := generatePrivateHostKey()
	require.NoError(t, err)
	srv.sshConfig.AddHostKey(hostKey)

	t.Run("InvalidConnection", func(_ *testing.T) {
		server, client := net.Pipe()
		defer server.Close()
		defer client.Close()

		client.Close() // Close immediately to cause handshake failure
		srv.handleConn(server)
		// Should handle gracefully without panicking
	})

	t.Run("ConcurrentConnections", func(_ *testing.T) {
		var wg sync.WaitGroup

		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				server, client := net.Pipe()
				defer server.Close()
				defer client.Close()

				go srv.handleConn(server)
				time.Sleep(5 * time.Millisecond)
				client.Close()
			}()
		}

		wg.Wait()
		// Test passes if no deadlocks or panics occur
	})
}

// Test error scenarios
func TestErrorHandling(t *testing.T) {
	srv := &Server{}
	mockConn := &mockConnMetadata{
		user:       "testuser",
		remoteAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345},
		sessionID:  []byte("test-session"),
	}

	t.Run("NetworkError", func(t *testing.T) {
		srv.serverURL = &url.URL{Scheme: "http", Host: "nonexistent.invalid:9999"}
		perms, err := srv.sshPasswordCallback(mockConn, []byte("testotp"))
		assert.Error(t, err)
		assert.Nil(t, perms)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("invalid json"))
		}))
		defer mockServer.Close()

		mockURL, _ := url.Parse(mockServer.URL)
		srv.serverURL = mockURL

		perms, err := srv.sshPasswordCallback(mockConn, []byte("testotp"))
		assert.Error(t, err)
		assert.Nil(t, perms)
	})

	t.Run("HTTPRedirect", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusFound)
		}))
		defer mockServer.Close()

		mockURL, _ := url.Parse(mockServer.URL)
		srv.serverURL = mockURL

		perms, err := srv.sshPasswordCallback(mockConn, []byte("testotp"))
		assert.Error(t, err)
		assert.Nil(t, perms)
	})
}

// Test utilities and edge cases
func TestUtilities(t *testing.T) {
	t.Run("YubikeyOTPResponse", func(t *testing.T) {
		response := YubikeyOTPVerifyResponse{OTP: "test-otp", Serial: 12345}

		data, err := json.Marshal(response)
		require.NoError(t, err)

		var parsed YubikeyOTPVerifyResponse
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)

		assert.Equal(t, "test-otp", parsed.OTP)
		assert.Equal(t, 12345, parsed.Serial)
	})

	t.Run("BannerCallback", func(t *testing.T) {
		bannerCallback := func(conn ssh.ConnMetadata) string {
			remote, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
			return fmt.Sprintf("Welcome %s from %s!\n", conn.User(), remote)
		}

		mockConn := &mockConnMetadata{
			user:       "testuser",
			remoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.100"), Port: 54321},
		}

		banner := bannerCallback(mockConn)
		assert.Contains(t, banner, "Welcome testuser from 192.168.1.100!")
	})

	t.Run("URLHandling", func(t *testing.T) {
		baseURL, err := url.Parse("http://example.com:8080/base")
		require.NoError(t, err)

		relativeURL, err := url.Parse("/api/v1/yubikey/otp/verify")
		require.NoError(t, err)

		resolved := baseURL.ResolveReference(relativeURL)
		assert.Equal(t, "http://example.com:8080/api/v1/yubikey/otp/verify", resolved.String())
	})
}

// Mock implementations for testing
type mockConnMetadata struct {
	user       string
	remoteAddr net.Addr
	sessionID  []byte
}

func (m *mockConnMetadata) User() string          { return m.user }
func (m *mockConnMetadata) SessionID() []byte     { return m.sessionID }
func (m *mockConnMetadata) ClientVersion() []byte { return []byte("SSH-2.0-Test") }
func (m *mockConnMetadata) ServerVersion() []byte { return []byte("SSH-2.0-OneAuth") }
func (m *mockConnMetadata) RemoteAddr() net.Addr  { return m.remoteAddr }
func (m *mockConnMetadata) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 2022}
}
