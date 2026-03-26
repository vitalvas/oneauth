package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestYubikeyOTPVerifyResponseStruct(t *testing.T) {
	t.Run("JSONSerialization", func(t *testing.T) {
		resp := YubikeyOTPVerifyResponse{
			OTP:    "test-otp",
			Serial: 12345,
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)

		var parsed YubikeyOTPVerifyResponse
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)
		assert.Equal(t, "test-otp", parsed.OTP)
		assert.Equal(t, 12345, parsed.Serial)
	})

	t.Run("JSONTags", func(t *testing.T) {
		resp := YubikeyOTPVerifyResponse{
			OTP:    "otp123",
			Serial: 99,
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"otp"`)
		assert.Contains(t, string(data), `"serial"`)
	})
}

func TestSSHPublicKeyCallbackWithCert(t *testing.T) {
	t.Run("CertificateAuth", func(t *testing.T) {
		srv := &Server{}
		mockConn := &mockConnMeta{
			user:       "testuser",
			remoteAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345},
			sessionID:  []byte("session"),
		}

		// Create a certificate
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)

		publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
		require.NoError(t, err)

		// Create a certificate authority key
		caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)

		caSigner, err := ssh.NewSignerFromKey(caKey)
		require.NoError(t, err)

		cert := &ssh.Certificate{
			Key:             publicKey,
			CertType:        ssh.UserCert,
			ValidPrincipals: []string{"testuser"},
			ValidBefore:     ssh.CertTimeInfinity,
		}
		err = cert.SignCert(rand.Reader, caSigner)
		require.NoError(t, err)

		// The checker's IsUserAuthority returns false, so this should fail
		perms, err := srv.sshPublicKeyCallback(mockConn, cert)
		assert.Error(t, err)
		assert.Nil(t, perms)
	})
}

func TestSSHPasswordCallbackServerStatus(t *testing.T) {
	t.Run("ServerInternalError", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer mockServer.Close()

		mockURL, err := url.Parse(mockServer.URL)
		require.NoError(t, err)

		srv := &Server{serverURL: mockURL}
		mockConn := &mockConnMeta{
			user:       "testuser",
			remoteAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345},
			sessionID:  []byte("test-session"),
		}

		perms, err := srv.sshPasswordCallback(mockConn, []byte("otp"))
		assert.Error(t, err)
		assert.Nil(t, perms)
	})

	t.Run("ServerRedirect", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusFound)
		}))
		defer mockServer.Close()

		mockURL, err := url.Parse(mockServer.URL)
		require.NoError(t, err)

		srv := &Server{serverURL: mockURL}
		mockConn := &mockConnMeta{
			user:       "testuser",
			remoteAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345},
			sessionID:  []byte("test-session"),
		}

		perms, err := srv.sshPasswordCallback(mockConn, []byte("otp"))
		assert.Error(t, err)
		assert.Nil(t, perms)
	})
}

func TestSSHPasswordCallback(t *testing.T) {
	t.Run("SuccessfulVerification", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			resp := YubikeyOTPVerifyResponse{OTP: "testotp", Serial: 54321}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer mockServer.Close()

		mockURL, err := url.Parse(mockServer.URL)
		require.NoError(t, err)

		srv := &Server{serverURL: mockURL}
		mockConn := &mockConnMeta{
			user:       "testuser",
			remoteAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345},
			sessionID:  []byte("test-session"),
		}

		perms, err := srv.sshPasswordCallback(mockConn, []byte("testotp"))
		require.NoError(t, err)
		assert.NotNil(t, perms)
		assert.Equal(t, "yubikey-otp", perms.Extensions["auth-type"])
		assert.Equal(t, "54321", perms.Extensions["yubikey-serial"])
	})

	t.Run("UnauthorizedResponse", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("unauthorized"))
		}))
		defer mockServer.Close()

		mockURL, err := url.Parse(mockServer.URL)
		require.NoError(t, err)

		srv := &Server{serverURL: mockURL}
		mockConn := &mockConnMeta{
			user:       "testuser",
			remoteAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345},
			sessionID:  []byte("test-session"),
		}

		perms, err := srv.sshPasswordCallback(mockConn, []byte("invalid"))
		assert.Error(t, err)
		assert.Nil(t, perms)
	})

	t.Run("NetworkError", func(t *testing.T) {
		mockURL, err := url.Parse("http://nonexistent.invalid:9999")
		require.NoError(t, err)

		srv := &Server{serverURL: mockURL}
		mockConn := &mockConnMeta{
			user:       "testuser",
			remoteAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345},
			sessionID:  []byte("test-session"),
		}

		perms, err := srv.sshPasswordCallback(mockConn, []byte("otp"))
		assert.Error(t, err)
		assert.Nil(t, perms)
	})

	t.Run("InvalidJSONResponse", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("not json"))
		}))
		defer mockServer.Close()

		mockURL, err := url.Parse(mockServer.URL)
		require.NoError(t, err)

		srv := &Server{serverURL: mockURL}
		mockConn := &mockConnMeta{
			user:       "testuser",
			remoteAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345},
			sessionID:  []byte("test-session"),
		}

		perms, err := srv.sshPasswordCallback(mockConn, []byte("otp"))
		assert.Error(t, err)
		assert.Nil(t, perms)
	})
}
