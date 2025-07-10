package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vitalvas/oneauth/internal/yubico"
)

type YubikeyOTPVerifyRequest struct {
	Username string `json:"username"`
	OTP      string `json:"otp"`
}

type YubikeyOTPVerifyResponse struct {
	Username string `json:"username"`
	OTP      string `json:"otp"`
	Serial   int64  `json:"serial"`
}

func (s *Server) yubikeyOTPVerify(w http.ResponseWriter, r *http.Request) {
	var request YubikeyOTPVerifyRequest

	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		// Parse query parameters
		query := r.URL.Query()
		request.Username = query.Get("username")
		request.OTP = query.Get("otp")

	case http.MethodPost:
		// Parse JSON body
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	// Validate required fields
	if request.Username == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "username is required"})
		return
	}

	if request.OTP == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "otp is required"})
		return
	}

	valid, err := s.yubico.Verify(request.OTP)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if valid.Status != yubico.StatusOK {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("invalid OTP: %s", valid.Status)})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(YubikeyOTPVerifyResponse{
		Username: request.Username,
		OTP:      request.OTP,
		Serial:   valid.Serial,
	})
}
