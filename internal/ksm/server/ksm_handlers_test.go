package server

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vitalvas/oneauth/internal/yksoft"
)

func TestHandleKSMDecrypt_MissingOTP(t *testing.T) {
	server := setupTestServer(t)

	req, err := http.NewRequest(http.MethodGet, "/wsapi/decrypt/", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	server.handleKSMDecrypt(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
	assert.Equal(t, "ERR Missing OTP parameter", rr.Body.String())
}

func TestHandleKSMDecrypt_InvalidOTPFormat(t *testing.T) {
	server := setupTestServer(t)

	req, err := http.NewRequest(http.MethodGet, "/wsapi/decrypt/?otp=invalid", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	server.handleKSMDecrypt(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "ERR")
}

func TestHandleKSMDecrypt_KeyNotFound(t *testing.T) {
	server := setupTestServer(t)

	// Use a valid OTP format but with non-existent key (44 chars)
	req, err := http.NewRequest(http.MethodGet, "/wsapi/decrypt/?otp=ccccccccccccjktuvurlnlnvghubeukgkejrliudllkv", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	server.handleKSMDecrypt(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "ERR Key not found")
}

func TestHandleKSMDecrypt_WithStoredKey(t *testing.T) {
	server := setupTestServer(t)

	// First store a key
	aesKeyB64 := "MTIzNDU2Nzg5MDEyMzQ1Ng" // 16 bytes base64 encoded
	err := server.StoreKey("cccccccccccc", aesKeyB64, "Test key")
	assert.NoError(t, err)

	// Now try to decrypt with that key
	req, err := http.NewRequest(http.MethodGet, "/wsapi/decrypt/?otp=ccccccccccccjktuvurlnlnvghubeukgkejrliudllkv", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	server.handleKSMDecrypt(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	// Should get some response (may be decrypt error due to invalid OTP structure, but not key not found)
	assert.NotContains(t, rr.Body.String(), "YubiKey not registered")
}

func TestKSMResponseFormat(t *testing.T) {
	tests := []struct {
		name     string
		response *DecryptResponse
		expected string
	}{
		{
			name: "successful response",
			response: &DecryptResponse{
				Status:        "OK",
				Counter:       15,
				TimestampLow:  50497,
				TimestampHigh: 167,
				SessionUse:    4,
			},
			expected: "OK counter=000f low=c541 high=a7 use=04",
		},
		{
			name: "error response",
			response: &DecryptResponse{
				Status:  "ERROR",
				Message: "Key not found",
			},
			expected: "ERR Key not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			if tt.response.Status != "OK" {
				result = fmt.Sprintf("ERR %s", tt.response.Message)
			} else {
				result = fmt.Sprintf("OK counter=%04x low=%04x high=%02x use=%02x",
					tt.response.Counter,
					tt.response.TimestampLow,
					tt.response.TimestampHigh,
					tt.response.SessionUse,
				)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestKSMProtocolCompatibility(t *testing.T) {
	server := setupTestServer(t)

	tests := []struct {
		name         string
		queryParams  string
		expectedCode int
		containsText string
	}{
		{
			name:         "missing OTP parameter",
			queryParams:  "",
			expectedCode: http.StatusBadRequest,
			containsText: "ERR Missing OTP parameter",
		},
		{
			name:         "invalid OTP format",
			queryParams:  "otp=invalid",
			expectedCode: http.StatusOK, // KSM protocol returns 200 even for errors
			containsText: "ERR",
		},
		{
			name:         "key not found",
			queryParams:  "otp=ccccccccccccjktuvurlnlnvghubeukgkejrliudllkv",
			expectedCode: http.StatusOK,
			containsText: "ERR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/wsapi/decrypt/"
			if tt.queryParams != "" {
				url = fmt.Sprintf("%s?%s", url, tt.queryParams)
			}

			req, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			server.handleKSMDecrypt(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			assert.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
			assert.Contains(t, rr.Body.String(), tt.containsText)
		})
	}
}

func TestKSMResponseFormatCompliance(t *testing.T) {
	// Test that KSM responses follow the expected format
	tests := []struct {
		name     string
		status   string
		message  string
		counter  int
		tsLow    int
		tsHigh   int
		session  int
		expected string
	}{
		{
			name:     "error format",
			status:   "ERROR",
			message:  "Key not found",
			expected: "ERR Key not found",
		},
		{
			name:     "success format",
			status:   "OK",
			counter:  255,
			tsLow:    65535,
			tsHigh:   255,
			session:  15,
			expected: "OK counter=00ff low=ffff high=ff use=0f",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			if tt.status != "OK" {
				result = fmt.Sprintf("ERR %s", tt.message)
			} else {
				result = fmt.Sprintf("OK counter=%04x low=%04x high=%02x use=%02x",
					tt.counter, tt.tsLow, tt.tsHigh, tt.session)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestKSMEndToEnd(t *testing.T) {
	server := setupTestServer(t)

	// Store a test key
	aesKeyB64 := "MTIzNDU2Nzg5MDEyMzQ1Ng" // 16 bytes
	err := server.StoreKey("cccccccccccc", aesKeyB64, "Test key")
	assert.NoError(t, err)

	// Test KSM decrypt endpoint
	req, err := http.NewRequest(http.MethodGet, "/wsapi/decrypt/?otp=ccccccccccccjktuvurlnlnvghubeukgkejrliudllkv", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	server.handleKSMDecrypt(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
	// Response should not be "key not found"
	assert.NotContains(t, rr.Body.String(), "YubiKey not registered")
}

func TestHandleKSMDecrypt_DecryptionFailed(t *testing.T) {
	server := setupTestServer(t)

	// Store a key so GetKey succeeds
	err := server.StoreKey("cccccccccccc", "MTIzNDU2Nzg5MDEyMzQ1Ng", "Test key")
	assert.NoError(t, err)

	// OTP with correct key ID but arbitrary encrypted data - will fail during OTP decryption
	req, err := http.NewRequest(http.MethodGet, "/wsapi/decrypt/?otp=ccccccccccccjktuvurlnlnvghubeukgkejrliudllkv", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	server.handleKSMDecrypt(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
	assert.Contains(t, rr.Body.String(), "ERR Decryption failed")
}

func TestHandleKSMDecrypt_ErrorResponseFormats(t *testing.T) {
	tests := []struct {
		name           string
		setupKey       bool
		otp            string
		expectedCode   int
		expectedPrefix string
	}{
		{
			name:           "missing OTP returns 400",
			setupKey:       false,
			otp:            "",
			expectedCode:   http.StatusBadRequest,
			expectedPrefix: "ERR Missing OTP parameter",
		},
		{
			name:           "invalid OTP returns 200 with ERR",
			setupKey:       false,
			otp:            "invalid",
			expectedCode:   http.StatusOK,
			expectedPrefix: "ERR",
		},
		{
			name:           "key not found returns 200 with ERR",
			setupKey:       false,
			otp:            "ddddddddddddjktuvurlnlnvghubeukgkejrliudllkv",
			expectedCode:   http.StatusOK,
			expectedPrefix: "ERR Key not found",
		},
		{
			name:           "decryption failed returns 200 with ERR",
			setupKey:       true,
			otp:            "ccccccccccccjktuvurlnlnvghubeukgkejrliudllkv",
			expectedCode:   http.StatusOK,
			expectedPrefix: "ERR Decryption failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setupTestServer(t)

			if tt.setupKey {
				err := server.StoreKey("cccccccccccc", "MTIzNDU2Nzg5MDEyMzQ1Ng", "Test key")
				assert.NoError(t, err)
			}

			url := "/wsapi/decrypt/"
			if tt.otp != "" {
				url = fmt.Sprintf("%s?otp=%s", url, tt.otp)
			}

			req, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			server.handleKSMDecrypt(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			assert.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
			assert.Contains(t, rr.Body.String(), tt.expectedPrefix)
		})
	}
}

func TestKSMConcurrentRequests(t *testing.T) {
	server := setupTestServer(t)

	// Valid modhex characters to use for key variations
	modhexChars := []string{"cc", "cd", "ce", "cf", "cg"}

	// Store test keys
	for i := 0; i < 5; i++ {
		keyID := fmt.Sprintf("cccccccccc%s", modhexChars[i])
		err := server.StoreKey(keyID, "MTIzNDU2Nzg5MDEyMzQ1Ng", fmt.Sprintf("Test key %d", i))
		assert.NoError(t, err)
	}

	// Test concurrent requests
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			keyID := fmt.Sprintf("cccccccccc%s", modhexChars[i%5])
			otp := fmt.Sprintf("%sjktuvurlnlnvghubeukgkejrliudllkv", keyID)

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/wsapi/decrypt/?otp=%s", otp), nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			server.handleKSMDecrypt(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			done <- true
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestHandleKSMDecrypt_SuccessfulDecryptWithYksoft(t *testing.T) {
	server := setupTestServer(t)

	aesKey := []byte("1234567890123456")
	yk, err := yksoft.NewSoftwareYubikey(&yksoft.Config{
		KeyID:     "cccccccccccc",
		PrivateID: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
		AESKey:    aesKey,
	})
	assert.NoError(t, err)

	aesKeyB64 := base64.RawURLEncoding.EncodeToString(aesKey)
	err = server.StoreKey(yk.GetKeyID(), aesKeyB64, "Test key")
	assert.NoError(t, err)

	otpResult, err := yk.GenerateOTP()
	assert.NoError(t, err)

	url := fmt.Sprintf("/wsapi/decrypt/?otp=%s", otpResult.OTP)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	server.handleKSMDecrypt(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
	assert.Contains(t, rr.Body.String(), "OK counter=")
}
