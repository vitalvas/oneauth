package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSecurityTxt(t *testing.T) {
	srv := &Server{config: &Config{}}

	t.Run("ContainsRequiredFields", func(t *testing.T) {
		content := srv.generateSecurityTxt()

		requiredFields := []string{"Contact:", "Expires:", "Preferred-Languages:", "Canonical:", "Policy:", "Hiring:"}
		for _, field := range requiredFields {
			assert.True(t, strings.Contains(content, field), "expected %q in security.txt", field)
		}
	})

	t.Run("ContainsCorrectURLs", func(t *testing.T) {
		content := srv.generateSecurityTxt()

		assert.Contains(t, content, "https://github.com/vitalvas/oneauth/issues")
		assert.Contains(t, content, "https://oneauth.vitalvas.dev/.well-known/security.txt")
		assert.Contains(t, content, "https://github.com/vitalvas/oneauth/blob/master/SECURITY.md")
	})

	t.Run("EndsWithNewline", func(t *testing.T) {
		content := srv.generateSecurityTxt()
		assert.True(t, strings.HasSuffix(content, "\n"))
	})
}

func TestWellKnownSecurityTxt(t *testing.T) {
	srv := &Server{config: &Config{}}

	t.Run("StatusAndHeaders", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/.well-known/security.txt", nil)
		w := httptest.NewRecorder()

		srv.wellKnownSecurityTxt(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))
		assert.Equal(t, "public, max-age=86400", w.Header().Get("Cache-Control"))
		assert.NotEmpty(t, w.Body.String())
	})
}

func TestWellKnownOneAuth(t *testing.T) {
	srv := &Server{config: &Config{}}

	t.Run("StatusAndHeaders", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/.well-known/oneauth-server.json", nil)
		w := httptest.NewRecorder()

		srv.wellKnownOneAuth(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})

	t.Run("ReturnsEmptyObject", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/.well-known/oneauth-server.json", nil)
		w := httptest.NewRecorder()

		srv.wellKnownOneAuth(w, req)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Empty(t, resp)
	})
}
