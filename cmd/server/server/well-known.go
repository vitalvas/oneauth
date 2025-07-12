package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// generateSecurityTxt creates a properly formatted security.txt content
func (s *Server) generateSecurityTxt() string {
	getSecurityTxtExpiration := func() string {
		expiration := time.Now().AddDate(0, 6, 0) // Add 6 months
		return expiration.Format(time.RFC3339)
	}

	fields := map[string]string{
		"Contact":             "https://github.com/vitalvas/oneauth/issues",
		"Expires":             getSecurityTxtExpiration(),
		"Preferred-Languages": "en",
		"Canonical":           "https://oneauth.vitalvas.dev/.well-known/security.txt",
		"Policy":              "https://github.com/vitalvas/oneauth/blob/master/SECURITY.md",
		"Hiring":              "https://github.com/vitalvas/oneauth",
	}

	lines := make([]string, 0, len(fields))
	for field, value := range fields {
		lines = append(lines, fmt.Sprintf("%s: %s", field, value))
	}

	return strings.Join(lines, "\n") + "\n"
}

func (s *Server) wellKnownSecurityTxt(w http.ResponseWriter, _ *http.Request) {
	content := s.generateSecurityTxt()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 24 hours
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(content))
}

func (s *Server) wellKnownOneAuth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{})
}
