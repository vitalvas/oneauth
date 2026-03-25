package server

import (
	"fmt"
	"net/http"
)

func (s *Server) handleKSMDecrypt(w http.ResponseWriter, r *http.Request) {
	otp := r.URL.Query().Get("otp")
	if otp == "" {
		s.sendTEXTResponse(w, http.StatusBadRequest, "ERR Missing OTP parameter") // Only missing parameter gets 400
		return
	}

	// Use the service layer to decrypt the OTP
	response, err := s.DecryptOTP(otp)
	if err != nil {
		s.logger.WithField("error", err.Error()).Error("Failed to decrypt OTP")
		s.sendTEXTResponse(w, http.StatusInternalServerError, "ERR Internal server error")
		return
	}

	// Format response in traditional KSM format
	var ksmResponse string
	if response.Status != "OK" {
		ksmResponse = fmt.Sprintf("ERR %s", response.Message)
	} else {
		ksmResponse = fmt.Sprintf("OK counter=%04x low=%04x high=%02x use=%02x",
			response.Counter,
			response.TimestampLow,
			response.TimestampHigh,
			response.SessionUse,
		)
	}

	// Traditional KSM protocol returns 200 OK even for errors
	s.sendTEXTResponse(w, http.StatusOK, ksmResponse)
}
