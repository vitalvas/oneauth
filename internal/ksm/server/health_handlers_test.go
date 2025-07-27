package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vitalvas/oneauth/internal/ksm/crypto"
	"github.com/vitalvas/oneauth/internal/ksm/database"
)

func TestHandleHealth_Healthy(t *testing.T) {
	mockDB := &database.MockDB{}
	cryptoEngine, err := crypto.NewEngine("test-master-key-12345678901234567890")
	assert.NoError(t, err)

	server := createTestServer(mockDB, cryptoEngine)

	// Mock successful health check
	mockDB.On("HealthCheck").Return(nil).Once()

	req, err := http.NewRequest(http.MethodGet, "/health", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	server.handleHealth(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response HealthResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "healthy", response.Status)
	assert.Equal(t, "connected", response.Database.Status)
	assert.Empty(t, response.Database.Error)

	mockDB.AssertExpectations(t)
}

func TestHandleHealth_DatabaseError(t *testing.T) {
	mockDB := &database.MockDB{}
	cryptoEngine, err := crypto.NewEngine("test-master-key-12345678901234567890")
	assert.NoError(t, err)

	server := createTestServer(mockDB, cryptoEngine)

	// Mock database health check failure
	dbError := errors.New("database connection failed")
	mockDB.On("HealthCheck").Return(dbError).Once()

	req, err := http.NewRequest(http.MethodGet, "/health", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	server.handleHealth(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response HealthResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "unhealthy", response.Status)
	assert.Equal(t, "error", response.Database.Status)
	assert.Equal(t, "database connection failed", response.Database.Error)

	mockDB.AssertExpectations(t)
}

func TestHandleHealth_IntegrationTest(t *testing.T) {
	// Use real database for integration test
	server := setupTestServer(t)

	req, err := http.NewRequest(http.MethodGet, "/health", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	server.handleHealth(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response HealthResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "healthy", response.Status)
	assert.Equal(t, "connected", response.Database.Status)
	assert.Empty(t, response.Database.Error)
}
