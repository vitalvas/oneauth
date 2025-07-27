package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func (s *Server) handleRESTDecrypt(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OTP string `json:"otp"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.WithField("error", err.Error()).Warn("Invalid JSON in REST decrypt request")
		s.sendJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Request body must be valid JSON")
		return
	}

	if req.OTP == "" {
		s.sendJSONError(w, http.StatusBadRequest, "MISSING_OTP", "OTP field is required")
		return
	}

	// Use the service layer to decrypt the OTP
	response, err := s.DecryptOTP(req.OTP)
	if err != nil {
		s.logger.WithField("error", err.Error()).Error("Failed to decrypt OTP in REST handler")

		s.sendJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	// Send JSON response
	if response.Status == "OK" {
		s.logger.WithFields(logrus.Fields{
			"key_id":      response.KeyID,
			"counter":     response.Counter,
			"session_use": response.SessionUse,
		}).Info("REST OTP decryption successful")
		s.sendJSONResponse(w, http.StatusOK, response)
	} else {
		// Map error codes to appropriate HTTP status codes
		statusCode := s.mapErrorCodeToHTTPStatus(response.ErrorCode)
		s.logger.WithFields(logrus.Fields{
			"error_code":  response.ErrorCode,
			"message":     response.Message,
			"http_status": statusCode,
		}).Warn("REST OTP decryption failed")
		s.sendJSONResponse(w, statusCode, response)
	}
}

func (s *Server) handleStoreKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		KeyID       string `json:"key_id"`
		AESKey      string `json:"aes_key"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.WithField("error", err.Error()).Warn("Invalid JSON in store key request")
		s.sendJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Request body must be valid JSON")
		return
	}

	if req.KeyID == "" {
		s.sendJSONError(w, http.StatusBadRequest, "MISSING_KEY_ID", "key_id field is required")
		return
	}

	if req.AESKey == "" {
		s.sendJSONError(w, http.StatusBadRequest, "MISSING_AES_KEY", "aes_key field is required")
		return
	}

	// Use the service layer to store the key
	if err := s.StoreKey(req.KeyID, req.AESKey, req.Description); err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":  err.Error(),
			"key_id": req.KeyID,
		}).Error("Failed to store key")

		// Map specific errors to appropriate responses
		switch {
		case strings.Contains(err.Error(), "invalid modhex length"):
			s.sendJSONError(w, http.StatusBadRequest, "INVALID_KEY_ID_LENGTH", err.Error())
		case strings.Contains(err.Error(), "invalid modhex character"):
			s.sendJSONError(w, http.StatusBadRequest, "INVALID_KEY_ID_FORMAT", err.Error())
		case strings.Contains(err.Error(), "AES key must be exactly 16 bytes"):
			s.sendJSONError(w, http.StatusBadRequest, "INVALID_AES_KEY_LENGTH", err.Error())
		case strings.Contains(err.Error(), "invalid base64 encoding"):
			s.sendJSONError(w, http.StatusBadRequest, "INVALID_BASE64", err.Error())
		default:
			s.sendJSONError(w, http.StatusInternalServerError, "STORAGE_ERROR", "Failed to store key")
		}
		return
	}

	s.logger.WithField("key_id", req.KeyID).Info("Key stored successfully")

	// Send success response
	s.sendJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"status":  "success",
		"message": "Key stored successfully",
		"key_id":  req.KeyID,
	})
}

func (s *Server) handleListKeys(w http.ResponseWriter, _ *http.Request) {
	// Use the service layer to list keys
	keys, err := s.ListKeys()
	if err != nil {
		s.logger.WithField("error", err.Error()).Error("Failed to list keys")
		s.sendJSONError(w, http.StatusInternalServerError, "LIST_ERROR", "Failed to retrieve keys")
		return
	}

	// Convert to API response format
	apiKeys := make([]map[string]interface{}, len(keys))
	for i, key := range keys {
		apiKeys[i] = map[string]interface{}{
			"key_id":      key.KeyID,
			"description": key.Description,
			"created_at":  key.CreatedAt,
			"last_used":   key.LastUsedAt,
			"usage_count": key.UsageCount,
		}
	}

	s.sendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"keys":   apiKeys,
	})
}

func (s *Server) handleDeleteKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyID := vars["key_id"]

	if keyID == "" {
		s.sendJSONError(w, http.StatusBadRequest, "MISSING_KEY_ID", "key_id parameter is required")
		return
	}

	// Use the service layer to delete the key
	if err := s.DeleteKey(keyID); err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":  err.Error(),
			"key_id": keyID,
		}).Error("Failed to delete key")
		s.sendJSONError(w, http.StatusInternalServerError, "DELETE_ERROR", "Failed to delete key")
		return
	}

	s.logger.WithField("key_id", keyID).Info("Key deleted successfully")

	s.sendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Key deleted successfully",
		"key_id":  keyID,
	})
}
