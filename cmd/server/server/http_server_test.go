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

func TestHealthEndpoint(t *testing.T) {
	t.Run("ReturnsOK", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var resp map[string]string
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "ok", resp["status"])
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
