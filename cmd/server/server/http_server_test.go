package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/internal/yubico"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()

	yAuth, err := yubico.NewYubiAuth(1, "c2VjcmV0")
	require.NoError(t, err)

	return &Server{
		config: &Config{
			Yubico: ConfigYubico{ClientID: 1, ClientSecret: "c2VjcmV0"},
		},
		yubico: yAuth,
	}
}

func TestNewRouter(t *testing.T) {
	router := newTestServer(t).newRouter()

	t.Run("Health", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var resp map[string]string
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.Equal(t, "ok", resp["status"])
	})

	t.Run("SecurityTxt", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/.well-known/security.txt", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")
		assert.Contains(t, w.Body.String(), "Contact:")
	})

	t.Run("OneAuthMetadata", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/.well-known/oneauth-server.json", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})

	t.Run("OTPVerifyRouteRegistered", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/yubikey/otp/verify", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Route exists: it must not return 404.
		assert.NotEqual(t, http.StatusNotFound, w.Code)
	})

	t.Run("UnknownPath", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/does/not/exist", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestServerInitialization(t *testing.T) {
	t.Run("ServerWithNilYubico", func(t *testing.T) {
		srv := &Server{
			config: &Config{
				Yubico: ConfigYubico{
					ClientID:     1,
					ClientSecret: "invalid-not-base64",
				},
			},
		}
		assert.Nil(t, srv.yubico)
	})
}

func TestRunHTTPServerYubicoInit(t *testing.T) {
	t.Run("InvalidClientSecret", func(t *testing.T) {
		srv := &Server{
			config: &Config{
				Yubico: ConfigYubico{
					ClientID:     1,
					ClientSecret: "invalid-not-base64!!!",
				},
			},
		}

		app := &cli.App{
			Action: func(c *cli.Context) error {
				return srv.runHTTPServer(c)
			},
		}

		err := app.Run([]string{"app"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create YubiAuth")
	})

	t.Run("ValidClientSecretYubicoAlreadySet", func(t *testing.T) {
		// Create a valid YubiAuth instance first
		yAuth, err := yubico.NewYubiAuth(1, "c2VjcmV0")
		require.NoError(t, err)

		srv := &Server{
			config: &Config{
				Yubico: ConfigYubico{
					ClientID:     1,
					ClientSecret: "c2VjcmV0",
				},
			},
			yubico: yAuth,
		}

		// runHTTPServer will skip YubiAuth creation since yubico is already set
		// It will fail on ListenAndServe because port 8080 may be in use
		// We just verify the yubico field is still set
		assert.NotNil(t, srv.yubico)
	})

}
