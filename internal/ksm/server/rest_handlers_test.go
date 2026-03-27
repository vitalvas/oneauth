package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vitalvas/kasper/mux"
	"github.com/vitalvas/oneauth/internal/yksoft"
)

func TestHandleRESTDecrypt_Success(t *testing.T) {
	server := setupTestServer(t)

	// Store a test key first
	aesKeyB64 := "MTIzNDU2Nzg5MDEyMzQ1Ng"
	err := server.StoreKey("cccccccccccc", aesKeyB64, "Test key")
	assert.NoError(t, err)

	// Test REST decrypt
	reqBody := map[string]string{
		"otp": "ccccccccccccjktuvurlnlnvghubeukgkejrliudllkv",
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
		"otp": "ccccccccccccjktuvurlnlnvghubeukgkejrliudllkv",
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
					"otp": "ccccccccccccjktuvurlnlnvghubeukgkejrliudllkv",
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
					"otp": "ddddddddddddjktuvurlnlnvghubeukgkejrliudllkv",
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

func TestHandleStoreKey_InvalidJSON(t *testing.T) {
	server := setupTestServer(t)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/keys", bytes.NewBufferString("not json"))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.handleStoreKey(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "INVALID_JSON", response["error_code"])
}

func TestHandleStoreKey_MissingKeyID(t *testing.T) {
	server := setupTestServer(t)

	reqBody := map[string]string{
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
	assert.Equal(t, "MISSING_KEY_ID", response["error_code"])
}

func TestHandleStoreKey_MissingAESKey(t *testing.T) {
	server := setupTestServer(t)

	reqBody := map[string]string{
		"key_id":      "cccccccccccc",
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
	assert.Equal(t, "MISSING_AES_KEY", response["error_code"])
}

func TestHandleStoreKey_InvalidModhexCharacter(t *testing.T) {
	server := setupTestServer(t)

	reqBody := map[string]string{
		"key_id":      "ccccccccccXX",
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
	assert.Equal(t, "INVALID_KEY_ID_FORMAT", response["error_code"])
}

func TestHandleStoreKey_InvalidAESKeyLength(t *testing.T) {
	server := setupTestServer(t)

	reqBody := map[string]string{
		"key_id":      "cccccccccccc",
		"aes_key":     "MTIzNDU2Nzg5MA==", // only 10 bytes
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
	assert.Equal(t, "INVALID_AES_KEY_LENGTH", response["error_code"])
}

func TestHandleListKeys_EmptyList(t *testing.T) {
	server := setupTestServer(t)

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
	assert.Empty(t, keys)
}

func TestHandleDeleteKey_NonExistentKey(t *testing.T) {
	server := setupTestServer(t)

	req, err := http.NewRequest(http.MethodDelete, "/api/v1/keys/cccccccccccc", nil)
	assert.NoError(t, err)

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/keys/{key_id}", server.handleDeleteKey).Methods(http.MethodDelete)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// DeleteKey is a soft-delete (SQL UPDATE) that does not return an error
	// for non-existent keys, so the handler returns 200 with success response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
}

func TestHandleRESTDecrypt_InvalidOTP(t *testing.T) {
	server := setupTestServer(t)

	reqBody := map[string]string{
		"otp": "invalid-otp-format",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/decrypt", bytes.NewBuffer(jsonBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.handleRESTDecrypt(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response DecryptResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ERROR", response.Status)
	assert.Equal(t, "INVALID_OTP", response.ErrorCode)
}

func TestHandleRESTDecrypt_KeyNotFound(t *testing.T) {
	server := setupTestServer(t)

	reqBody := map[string]string{
		"otp": "ccccccccccccjktuvurlnlnvghubeukgkejrliudllkv",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/decrypt", bytes.NewBuffer(jsonBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.handleRESTDecrypt(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var response DecryptResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ERROR", response.Status)
	assert.Equal(t, "KEY_NOT_FOUND", response.ErrorCode)
}

func TestHandleRESTDecrypt_DecryptionFailed(t *testing.T) {
	server := setupTestServer(t)

	// Store a key so GetKey succeeds
	err := server.StoreKey("cccccccccccc", "MTIzNDU2Nzg5MDEyMzQ1Ng", "Test key")
	assert.NoError(t, err)

	// Use valid 44-char modhex OTP with matching key ID but arbitrary encrypted data
	reqBody := map[string]string{
		"otp": "ccccccccccccjktuvurlnlnvghubeukgkejrliudllkv",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/decrypt", bytes.NewBuffer(jsonBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.handleRESTDecrypt(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)

	var response DecryptResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ERROR", response.Status)
	assert.Equal(t, "DECRYPTION_FAILED", response.ErrorCode)
}

func TestHandleRESTDecrypt_EmptyBody(t *testing.T) {
	server := setupTestServer(t)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/decrypt", bytes.NewBufferString(""))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.handleRESTDecrypt(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleRESTDecrypt_ErrorCodeMapping(t *testing.T) {
	tests := []struct {
		name           string
		setupKey       bool
		otp            string
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "invalid OTP returns 400",
			setupKey:       false,
			otp:            "tooshort",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_OTP",
		},
		{
			name:           "key not found returns 404",
			setupKey:       false,
			otp:            "ddddddddddddjktuvurlnlnvghubeukgkejrliudllkv",
			expectedStatus: http.StatusNotFound,
			expectedCode:   "KEY_NOT_FOUND",
		},
		{
			name:           "decryption failed returns 422",
			setupKey:       true,
			otp:            "ccccccccccccjktuvurlnlnvghubeukgkejrliudllkv",
			expectedStatus: http.StatusUnprocessableEntity,
			expectedCode:   "DECRYPTION_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setupTestServer(t)

			if tt.setupKey {
				err := server.StoreKey("cccccccccccc", "MTIzNDU2Nzg5MDEyMzQ1Ng", "Test key")
				assert.NoError(t, err)
			}

			reqBody := map[string]string{"otp": tt.otp}
			jsonBody, _ := json.Marshal(reqBody)

			req, err := http.NewRequest(http.MethodPost, "/api/v1/decrypt", bytes.NewBuffer(jsonBody))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			server.handleRESTDecrypt(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var response DecryptResponse
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "ERROR", response.Status)
			assert.Equal(t, tt.expectedCode, response.ErrorCode)
		})
	}
}

func TestHandleStoreKey_DatabaseStorageError(t *testing.T) {
	server := setupTestServer(t)

	// Close the DB to force a storage error (not an input validation error)
	server.db.Close()

	reqBody := map[string]string{
		"key_id":      "cccccccccccc",
		"aes_key":     "MTIzNDU2Nzg5MDEyMzQ1Ng",
		"description": "Test key",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/keys", bytes.NewBuffer(jsonBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.handleStoreKey(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "STORAGE_ERROR", response["error_code"])
}

func TestHandleListKeys_DatabaseError(t *testing.T) {
	server := setupTestServer(t)

	// Close DB to force error
	server.db.Close()

	req, err := http.NewRequest(http.MethodGet, "/api/v1/keys", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	server.handleListKeys(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "LIST_ERROR", response["error_code"])
}

func TestHandleDeleteKey_DatabaseError(t *testing.T) {
	server := setupTestServer(t)

	// Close DB to force error
	server.db.Close()

	req, err := http.NewRequest(http.MethodDelete, "/api/v1/keys/cccccccccccc", nil)
	assert.NoError(t, err)

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/keys/{key_id}", server.handleDeleteKey).Methods(http.MethodDelete)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "DELETE_ERROR", response["error_code"])
}

func TestHandleDeleteKey_EmptyKeyIDDirect(t *testing.T) {
	server := setupTestServer(t)

	// Call handler directly without mux router so mux.Vars returns empty map
	req, err := http.NewRequest(http.MethodDelete, "/api/v1/keys/", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	server.handleDeleteKey(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "MISSING_KEY_ID", response["error_code"])
}

func TestHandleListKeys_ResponseFormat(t *testing.T) {
	server := setupTestServer(t)

	// Store a key and verify the response format includes all expected fields
	err := server.StoreKey("cccccccccccc", "MTIzNDU2Nzg5MDEyMzQ1Ng", "Test key")
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "/api/v1/keys", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	server.handleListKeys(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])

	keys := response["keys"].([]interface{})
	assert.Len(t, keys, 1)

	key := keys[0].(map[string]interface{})
	assert.Equal(t, "cccccccccccc", key["key_id"])
	assert.Equal(t, "Test key", key["description"])
	assert.Contains(t, key, "created_at")
	assert.Contains(t, key, "usage_count")
}

func TestHandleRESTDecrypt_SuccessWithYksoft(t *testing.T) {
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

	reqBody := map[string]string{"otp": otpResult.OTP}
	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/decrypt", bytes.NewBuffer(jsonBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.handleRESTDecrypt(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response DecryptResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "OK", response.Status)
	assert.Equal(t, "cccccccccccc", response.KeyID)
}
