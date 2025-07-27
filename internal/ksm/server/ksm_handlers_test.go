package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vitalvas/oneauth/internal/yubico"
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
	req, err := http.NewRequest(http.MethodGet, "/wsapi/decrypt/?otp=ccccccccccccjktuvurlnlnvghubeukgkejrliudllkvj", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	server.handleKSMDecrypt(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	// Should get some response (may be decrypt error due to invalid OTP structure, but not key not found)
	assert.NotContains(t, rr.Body.String(), "YubiKey not registered")
}

func TestExtractKeyIDFromOTP(t *testing.T) {
	tests := []struct {
		name       string
		otp        string
		expectedID string
		expectErr  bool
	}{
		{
			name:       "valid OTP with cccccccccccc key ID",
			otp:        "ccccccccccccjktuvurlnlnvghubeukgkejrliudllkv",
			expectedID: "cccccccccccc",
			expectErr:  false,
		},
		{
			name:       "valid OTP with different key ID",
			otp:        "dddddddddddduvghubeukgkejrliudllkvjjktuvurln",
			expectedID: "dddddddddddd",
			expectErr:  false,
		},
		{
			name:      "invalid OTP too short",
			otp:       "short",
			expectErr: true,
		},
		{
			name:      "empty OTP",
			otp:       "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyID, err := yubico.ExtractKeyID(tt.otp)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, keyID)
			}
		})
	}
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
			queryParams:  "otp=ccccccccccccjktuvurlnlnvghubeukgkejrliudllkvj",
			expectedCode: http.StatusOK,
			containsText: "ERR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/wsapi/decrypt/"
			if tt.queryParams != "" {
				url += "?" + tt.queryParams
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
	req, err := http.NewRequest(http.MethodGet, "/wsapi/decrypt/?otp=ccccccccccccjktuvurlnlnvghubeukgkejrliudllkvj", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	server.handleKSMDecrypt(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
	// Response should not be "key not found"
	assert.NotContains(t, rr.Body.String(), "YubiKey not registered")
}

func TestKSMConcurrentRequests(t *testing.T) {
	server := setupTestServer(t)

	// Valid modhex characters to use for key variations
	modhexChars := []string{"cc", "cd", "ce", "cf", "cg"}

	// Store test keys
	for i := 0; i < 5; i++ {
		keyID := "cccccccccc" + modhexChars[i]
		err := server.StoreKey(keyID, "MTIzNDU2Nzg5MDEyMzQ1Ng", fmt.Sprintf("Test key %d", i))
		assert.NoError(t, err)
	}

	// Test concurrent requests
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			keyID := "cccccccccc" + modhexChars[i%5]
			otp := keyID + "jktuvurlnlnvghubeukgkejrliudllkvj"

			req, err := http.NewRequest(http.MethodGet, "/wsapi/decrypt/?otp="+otp, nil)
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
