package server

import (
	"encoding/json"
	"net/http"
)

func (s *Server) sendJSONError(w http.ResponseWriter, statusCode int, errorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "ERROR",
		"error_code": errorCode,
		"message":    message,
	})
}

func (s *Server) sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) sendTEXTResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}

func (s *Server) mapErrorCodeToHTTPStatus(errorCode string) int {
	switch errorCode {
	case "INVALID_OTP_FORMAT", "INVALID_KEY_ID", "MISSING_OTP":
		return http.StatusBadRequest
	case "KEY_NOT_FOUND":
		return http.StatusNotFound
	case "REPLAY_ATTACK":
		return http.StatusConflict
	case "DECRYPTION_FAILED", "OTP_DECRYPTION_FAILED":
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}
