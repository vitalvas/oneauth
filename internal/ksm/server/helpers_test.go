package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendJSONError(t *testing.T) {
	server := &Server{}

	rr := httptest.NewRecorder()
	server.sendJSONError(rr, http.StatusBadRequest, "INVALID_REQUEST", "The request is invalid")

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "ERROR", response["status"])
	assert.Equal(t, "INVALID_REQUEST", response["error_code"])
	assert.Equal(t, "The request is invalid", response["message"])
}

func TestSendJSONResponse(t *testing.T) {
	server := &Server{}

	testData := map[string]interface{}{
		"status": "success",
		"data":   "test data",
		"count":  42,
	}

	rr := httptest.NewRecorder()
	server.sendJSONResponse(rr, http.StatusOK, testData)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "test data", response["data"])
	assert.Equal(t, float64(42), response["count"]) // JSON numbers are float64
}

func TestSendTEXTResponse(t *testing.T) {
	server := &Server{}

	rr := httptest.NewRecorder()
	server.sendTEXTResponse(rr, http.StatusOK, "Hello, World!")

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
	assert.Equal(t, "Hello, World!", rr.Body.String())
}

func TestMapErrorCodeToHTTPStatus(t *testing.T) {
	server := &Server{}

	tests := []struct {
		name           string
		errorCode      string
		expectedStatus int
	}{
		{
			name:           "invalid OTP format",
			errorCode:      "INVALID_OTP_FORMAT",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid key ID",
			errorCode:      "INVALID_KEY_ID",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing OTP",
			errorCode:      "MISSING_OTP",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "key not found",
			errorCode:      "KEY_NOT_FOUND",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "replay attack",
			errorCode:      "REPLAY_ATTACK",
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "decryption failed",
			errorCode:      "DECRYPTION_FAILED",
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "OTP decryption failed",
			errorCode:      "OTP_DECRYPTION_FAILED",
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "unknown error",
			errorCode:      "UNKNOWN_ERROR",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "empty error code",
			errorCode:      "",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := server.mapErrorCodeToHTTPStatus(tt.errorCode)
			assert.Equal(t, tt.expectedStatus, status)
		})
	}
}

func TestHelperFunctionsIntegration(t *testing.T) {
	server := &Server{}

	tests := []struct {
		name         string
		function     string
		expectedCode int
		expectedType string
	}{
		{
			name:         "JSON error response",
			function:     "sendJSONError",
			expectedCode: http.StatusBadRequest,
			expectedType: "application/json",
		},
		{
			name:         "JSON success response",
			function:     "sendJSONResponse",
			expectedCode: http.StatusOK,
			expectedType: "application/json",
		},
		{
			name:         "text response",
			function:     "sendTEXTResponse",
			expectedCode: http.StatusOK,
			expectedType: "text/plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			switch tt.function {
			case "sendJSONError":
				server.sendJSONError(rr, tt.expectedCode, "TEST_ERROR", "Test message")
			case "sendJSONResponse":
				server.sendJSONResponse(rr, tt.expectedCode, map[string]string{"test": "data"})
			case "sendTEXTResponse":
				server.sendTEXTResponse(rr, tt.expectedCode, "Test text")
			}

			assert.Equal(t, tt.expectedCode, rr.Code)
			assert.Equal(t, tt.expectedType, rr.Header().Get("Content-Type"))
			assert.NotEmpty(t, rr.Body.String())
		})
	}
}
