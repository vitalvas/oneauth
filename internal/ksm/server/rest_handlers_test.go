package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestHandleRESTDecrypt_Success(t *testing.T) {
	server := setupTestServer(t)

	// Store a test key first
	aesKeyB64 := "MTIzNDU2Nzg5MDEyMzQ1Ng"
	err := server.StoreKey("cccccccccccc", aesKeyB64, "Test key")
	assert.NoError(t, err)

	// Test REST decrypt
	reqBody := map[string]string{
		"otp": "ccccccccccccjktuvurlnlnvghubeukgkejrliudllkvj",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/decrypt", bytes.NewBuffer(jsonBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.handleRESTDecrypt(rr, req)

	// Should not be key not found error
	assert.NotEqual(t, http.StatusNotFound, rr.Code)

	var response DecryptResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEqual(t, "KEY_NOT_FOUND", response.ErrorCode)
}

func TestHandleRESTDecrypt_InvalidJSON(t *testing.T) {
	server := setupTestServer(t)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/decrypt", bytes.NewBufferString("invalid json"))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.handleRESTDecrypt(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ERROR", response["status"])
	assert.Equal(t, "INVALID_JSON", response["error_code"])
}

func TestHandleRESTDecrypt_MissingOTP(t *testing.T) {
	server := setupTestServer(t)

	reqBody := map[string]string{}
	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/decrypt", bytes.NewBuffer(jsonBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.handleRESTDecrypt(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ERROR", response["status"])
	assert.Equal(t, "MISSING_OTP", response["error_code"])
}

func TestHandleStoreKey_Success(t *testing.T) {
	server := setupTestServer(t)

	reqBody := map[string]string{
		"key_id":      "dddddddddddd",
		"aes_key":     "MTIzNDU2Nzg5MDEyMzQ1Ng",
		"description": "Test key via REST",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/keys", bytes.NewBuffer(jsonBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.handleStoreKey(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "dddddddddddd", response["key_id"])
}

func TestHandleStoreKey_InvalidKeyID(t *testing.T) {
	server := setupTestServer(t)

	reqBody := map[string]string{
		"key_id":      "invalid",
		"aes_key":     "MTIzNDU2Nzg5MDEyMzQ1Ng",
		"description": "Test key",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/keys", bytes.NewBuffer(jsonBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.handleStoreKey(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ERROR", response["status"])
	assert.Equal(t, "INVALID_KEY_ID_LENGTH", response["error_code"])
}

func TestHandleStoreKey_AESKeyFormats(t *testing.T) {
	tests := []struct {
		name           string
		keyID          string
		aesKey         string
		description    string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Valid hex format",
			keyID:          "eeeeeeeeeeee",
			aesKey:         "31323334353637383930313233343536", // "1234567890123456" in hex
			description:    "Test key with hex format",
			expectedStatus: http.StatusCreated,
			expectedError:  "",
		},
		{
			name:           "Valid hex uppercase",
			keyID:          "eeeeeeeeeeed",
			aesKey:         "31323334353637383930313233343536",
			description:    "Test key with hex uppercase",
			expectedStatus: http.StatusCreated,
			expectedError:  "",
		},
		{
			name:           "Invalid hex characters",
			keyID:          "ffffffffffff",
			aesKey:         "3132333435363738393031323334353Z", // Invalid hex characters - 32 chars with 'Z'
			description:    "Test key with invalid hex",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_AES_KEY_FORMAT",
		},
		{
			name:           "Wrong hex length - too short",
			keyID:          "gggggggggggg",
			aesKey:         "3132333435363738393031323334", // 14 hex bytes instead of 16
			description:    "Test key with wrong hex length",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_AES_KEY_LENGTH",
		},
		{
			name:           "Wrong hex length - too long",
			keyID:          "hhhhhhhhhhhg",
			aesKey:         "313233343536373839303132333435363738", // 18 hex bytes instead of 16
			description:    "Test key with wrong hex length",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_AES_KEY_LENGTH",
		},
		{
			name:           "Valid standard base64",
			keyID:          "hhhhhhhhhhhh",
			aesKey:         "MTIzNDU2Nzg5MDEyMzQ1Ng==", // Standard base64 with padding
			description:    "Test key with standard base64",
			expectedStatus: http.StatusCreated,
			expectedError:  "",
		},
		{
			name:           "Valid URL base64",
			keyID:          "hhhhhhhhhhhj",
			aesKey:         "MTIzNDU2Nzg5MDEyMzQ1Ng", // URL base64 without padding
			description:    "Test key with URL base64",
			expectedStatus: http.StatusCreated,
			expectedError:  "",
		},
		{
			name:           "Invalid base64",
			keyID:          "jjjjjjjjjjjj",
			aesKey:         "MTIzNDU2Nzg5MDEyMzQ1Ng==!", // Invalid base64 character
			description:    "Test key with invalid base64",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_AES_KEY_FORMAT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setupTestServer(t)

			reqBody := map[string]string{
				"key_id":      tt.keyID,
				"aes_key":     tt.aesKey,
				"description": tt.description,
			}
			jsonBody, _ := json.Marshal(reqBody)

			req, err := http.NewRequest(http.MethodPost, "/api/v1/keys", bytes.NewBuffer(jsonBody))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			server.handleStoreKey(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var response map[string]interface{}
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedError == "" {
				assert.Equal(t, "success", response["status"])
				assert.Equal(t, tt.keyID, response["key_id"])
			} else {
				assert.Equal(t, "ERROR", response["status"])
				assert.Equal(t, tt.expectedError, response["error_code"])
			}
		})
	}
}

func TestHandleListKeys_Success(t *testing.T) {
	server := setupTestServer(t)

	// Store some test keys
	server.StoreKey("cccccccccccc", "MTIzNDU2Nzg5MDEyMzQ1Ng", "Test key 1")
	server.StoreKey("dddddddddddd", "MTIzNDU2Nzg5MDEyMzQ1Ng", "Test key 2")

	req, err := http.NewRequest(http.MethodGet, "/api/v1/keys", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	server.handleListKeys(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])

	keys, ok := response["keys"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(keys))
}

func TestHandleDeleteKey_Success(t *testing.T) {
	server := setupTestServer(t)

	// Store a test key first
	server.StoreKey("cccccccccccc", "MTIzNDU2Nzg5MDEyMzQ1Ng", "Test key")

	req, err := http.NewRequest(http.MethodDelete, "/api/v1/keys/cccccccccccc", nil)
	assert.NoError(t, err)

	// Set up router with path variables
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/keys/{key_id}", server.handleDeleteKey).Methods(http.MethodDelete)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
}

func TestHandleDeleteKey_MissingKeyID(t *testing.T) {
	server := setupTestServer(t)

	req, err := http.NewRequest(http.MethodDelete, "/api/v1/keys/", nil)
	assert.NoError(t, err)

	// Set up router with path variables
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/keys/{key_id}", server.handleDeleteKey).Methods(http.MethodDelete)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should return 404 as the route won't match
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestRESTAPIEndToEnd(t *testing.T) {
	server := setupTestServer(t)

	// 1. Store a key
	storeReq := map[string]string{
		"key_id":      "cccccccccccc",
		"aes_key":     "MTIzNDU2Nzg5MDEyMzQ1Ng",
		"description": "End-to-end test key",
	}
	jsonBody, _ := json.Marshal(storeReq)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/keys", bytes.NewBuffer(jsonBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.handleStoreKey(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)

	// 2. List keys to verify it was stored
	req, err = http.NewRequest(http.MethodGet, "/api/v1/keys", nil)
	assert.NoError(t, err)

	rr = httptest.NewRecorder()
	server.handleListKeys(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var listResponse map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &listResponse)
	assert.NoError(t, err)
	keys := listResponse["keys"].([]interface{})
	assert.Len(t, keys, 1)

	// 3. Try to decrypt with the stored key
	decryptReq := map[string]string{
		"otp": "ccccccccccccjktuvurlnlnvghubeukgkejrliudllkvj",
	}
	jsonBody, _ = json.Marshal(decryptReq)

	req, err = http.NewRequest(http.MethodPost, "/api/v1/decrypt", bytes.NewBuffer(jsonBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr = httptest.NewRecorder()
	server.handleRESTDecrypt(rr, req)

	// Should not be key not found
	assert.NotEqual(t, http.StatusNotFound, rr.Code)
}

func TestCompleteRESTAPIWorkflow(t *testing.T) {
	server := setupTestServer(t)

	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{
			name: "Store First Key",
			run: func(t *testing.T) {
				reqBody := map[string]string{
					"key_id":      "cccccccccccc",
					"aes_key":     "MTIzNDU2Nzg5MDEyMzQ1Ng",
					"description": "First test key",
				}
				jsonBody, _ := json.Marshal(reqBody)

				req, err := http.NewRequest(http.MethodPost, "/api/v1/keys", bytes.NewBuffer(jsonBody))
				assert.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")

				rr := httptest.NewRecorder()
				server.handleStoreKey(rr, req)

				assert.Equal(t, http.StatusCreated, rr.Code)
			},
		},
		{
			name: "Store Second Key",
			run: func(t *testing.T) {
				reqBody := map[string]string{
					"key_id":      "dddddddddddd",
					"aes_key":     "MTIzNDU2Nzg5MDEyMzQ1Ng",
					"description": "Second test key",
				}
				jsonBody, _ := json.Marshal(reqBody)

				req, err := http.NewRequest(http.MethodPost, "/api/v1/keys", bytes.NewBuffer(jsonBody))
				assert.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")

				rr := httptest.NewRecorder()
				server.handleStoreKey(rr, req)

				assert.Equal(t, http.StatusCreated, rr.Code)
			},
		},
		{
			name: "List Stored Keys",
			run: func(t *testing.T) {
				req, err := http.NewRequest(http.MethodGet, "/api/v1/keys", nil)
				assert.NoError(t, err)

				rr := httptest.NewRecorder()
				server.handleListKeys(rr, req)

				assert.Equal(t, http.StatusOK, rr.Code)

				var response map[string]interface{}
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)

				keys := response["keys"].([]interface{})
				assert.Equal(t, 2, len(keys))
			},
		},
		{
			name: "OTP with first key",
			run: func(t *testing.T) {
				reqBody := map[string]string{
					"otp": "ccccccccccccjktuvurlnlnvghubeukgkejrliudllkvj",
				}
				jsonBody, _ := json.Marshal(reqBody)

				req, err := http.NewRequest(http.MethodPost, "/api/v1/decrypt", bytes.NewBuffer(jsonBody))
				assert.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")

				rr := httptest.NewRecorder()
				server.handleRESTDecrypt(rr, req)

				var response DecryptResponse
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEqual(t, "KEY_NOT_FOUND", response.ErrorCode)
			},
		},
		{
			name: "OTP with second key",
			run: func(t *testing.T) {
				reqBody := map[string]string{
					"otp": "dddddddddddduvghubeukgkejrliudllkvjjktuvurlnln",
				}
				jsonBody, _ := json.Marshal(reqBody)

				req, err := http.NewRequest(http.MethodPost, "/api/v1/decrypt", bytes.NewBuffer(jsonBody))
				assert.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")

				rr := httptest.NewRecorder()
				server.handleRESTDecrypt(rr, req)

				var response DecryptResponse
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEqual(t, "KEY_NOT_FOUND", response.ErrorCode)
			},
		},
		{
			name: "Delete first key",
			run: func(t *testing.T) {
				req, err := http.NewRequest(http.MethodDelete, "/api/v1/keys/cccccccccccc", nil)
				assert.NoError(t, err)

				router := mux.NewRouter()
				router.HandleFunc("/api/v1/keys/{key_id}", server.handleDeleteKey).Methods(http.MethodDelete)

				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)

				assert.Equal(t, http.StatusOK, rr.Code)
			},
		},
		{
			name: "Verify key was deleted",
			run: func(t *testing.T) {
				req, err := http.NewRequest(http.MethodGet, "/api/v1/keys", nil)
				assert.NoError(t, err)

				rr := httptest.NewRecorder()
				server.handleListKeys(rr, req)

				assert.Equal(t, http.StatusOK, rr.Code)

				var response map[string]interface{}
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)

				keys := response["keys"].([]interface{})
				assert.Equal(t, 1, len(keys))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
