package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vitalvas/oneauth/internal/yubico"
)

func TestYubikeyOTPVerifyRequest(t *testing.T) {
	t.Run("JSONSerialization", func(t *testing.T) {
		req := YubikeyOTPVerifyRequest{
			Username: "testuser",
			OTP:      "testotp123",
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var parsed YubikeyOTPVerifyRequest
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)
		assert.Equal(t, "testuser", parsed.Username)
		assert.Equal(t, "testotp123", parsed.OTP)
	})
}

func TestYubikeyOTPVerifyResponse(t *testing.T) {
	t.Run("JSONSerialization", func(t *testing.T) {
		resp := YubikeyOTPVerifyResponse{
			Username: "testuser",
			OTP:      "testotp",
			Serial:   12345,
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)

		var parsed YubikeyOTPVerifyResponse
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)
		assert.Equal(t, "testuser", parsed.Username)
		assert.Equal(t, "testotp", parsed.OTP)
		assert.Equal(t, int64(12345), parsed.Serial)
	})
}

func TestYubikeyOTPVerifyHandler(t *testing.T) {
	srv := &Server{config: &Config{}}

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/yubikey/otp/verify", nil)
		w := httptest.NewRecorder()

		srv.yubikeyOTPVerify(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("GetMissingUsername", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/yubikey/otp/verify?otp=test", nil)
		w := httptest.NewRecorder()

		srv.yubikeyOTPVerify(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]string
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "username is required", resp["error"])
	})

	t.Run("GetMissingOTP", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/yubikey/otp/verify?username=user", nil)
		w := httptest.NewRecorder()

		srv.yubikeyOTPVerify(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]string
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "otp is required", resp["error"])
	})

	t.Run("PostInvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/yubikey/otp/verify", bytes.NewBufferString("invalid"))
		w := httptest.NewRecorder()

		srv.yubikeyOTPVerify(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("PostMissingFields", func(t *testing.T) {
		body, _ := json.Marshal(YubikeyOTPVerifyRequest{})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/yubikey/otp/verify", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		srv.yubikeyOTPVerify(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ContentTypeIsJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/yubikey/otp/verify", nil)
		w := httptest.NewRecorder()

		srv.yubikeyOTPVerify(w, req)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})
}

func TestYubikeyOTPVerifyHandlerWithMockYubico(t *testing.T) {
	validOTP := "cccccbhuinjdrvtgbgrbrcikvrtvulvltkdufcrngunn"

	t.Run("GetWithValidOTP_VerifyReturnsError", func(t *testing.T) {
		yAuth, err := yubico.NewYubiAuth(1, "c2VjcmV0")
		require.NoError(t, err)

		srv := &Server{
			config: &Config{},
			yubico: yAuth,
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/yubikey/otp/verify?username=testuser&otp="+validOTP, nil)
		w := httptest.NewRecorder()

		srv.yubikeyOTPVerify(w, req)
		// The Yubico verify may return 500 (error) or 401 (invalid status)
		// depending on whether the real servers are reachable
		assert.True(t, w.Code == http.StatusInternalServerError || w.Code == http.StatusUnauthorized)
	})

	t.Run("PostWithValidJSON", func(t *testing.T) {
		yAuth, err := yubico.NewYubiAuth(1, "c2VjcmV0")
		require.NoError(t, err)

		srv := &Server{
			config: &Config{},
			yubico: yAuth,
		}

		body, err := json.Marshal(YubikeyOTPVerifyRequest{
			Username: "testuser",
			OTP:      validOTP,
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/yubikey/otp/verify", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		srv.yubikeyOTPVerify(w, req)
		// The Yubico verify may return 500 (error) or 401 (invalid status)
		assert.True(t, w.Code == http.StatusInternalServerError || w.Code == http.StatusUnauthorized)
	})

	t.Run("PostMissingUsername", func(t *testing.T) {
		yAuth, err := yubico.NewYubiAuth(1, "c2VjcmV0")
		require.NoError(t, err)

		srv := &Server{
			config: &Config{},
			yubico: yAuth,
		}

		body, err := json.Marshal(YubikeyOTPVerifyRequest{
			OTP: "testotp",
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/yubikey/otp/verify", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		srv.yubikeyOTPVerify(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]string
		err = json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "username is required", resp["error"])
	})

	t.Run("PostMissingOTP", func(t *testing.T) {
		yAuth, err := yubico.NewYubiAuth(1, "c2VjcmV0")
		require.NoError(t, err)

		srv := &Server{
			config: &Config{},
			yubico: yAuth,
		}

		body, err := json.Marshal(YubikeyOTPVerifyRequest{
			Username: "testuser",
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/yubikey/otp/verify", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		srv.yubikeyOTPVerify(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]string
		err = json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "otp is required", resp["error"])
	})

	t.Run("GetWithInvalidOTP_VerifyFails", func(t *testing.T) {
		yAuth, err := yubico.NewYubiAuth(1, "c2VjcmV0")
		require.NoError(t, err)

		srv := &Server{
			config: &Config{},
			yubico: yAuth,
		}

		// Short OTP that will fail validation in Verify
		req := httptest.NewRequest(http.MethodGet, "/api/v1/yubikey/otp/verify?username=testuser&otp=short", nil)
		w := httptest.NewRecorder()

		srv.yubikeyOTPVerify(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var resp map[string]string
		err = json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Contains(t, resp["error"], "failed to validate otp")
	})

	t.Run("PutMethodNotAllowed", func(t *testing.T) {
		srv := &Server{config: &Config{}}

		req := httptest.NewRequest(http.MethodPut, "/api/v1/yubikey/otp/verify", nil)
		w := httptest.NewRecorder()

		srv.yubikeyOTPVerify(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

		var resp map[string]string
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "method not allowed", resp["error"])
	})
}
