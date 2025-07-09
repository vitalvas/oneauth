package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/internal/yubico"
)

func (s *Server) runHTTPServer(_ *cli.Context) error {
	if s.yubico == nil {
		yAuth, err := yubico.NewYubiAuth(s.config.Yubico.ClientID, s.config.Yubico.ClientSecret)
		if err != nil {
			return fmt.Errorf("failed to create YubiAuth: %w", err)
		}

		s.yubico = yAuth
	}

	r := mux.NewRouter()

	// Health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}).Methods(http.MethodGet)

	// Well-known endpoints
	wellKnown := r.PathPrefix("/.well-known").Subrouter()
	wellKnown.HandleFunc("/security.txt", s.wellKnownSecurityTxt).Methods(http.MethodGet)
	wellKnown.HandleFunc("/oneauth-server.json", s.wellKnownOneAuth).Methods(http.MethodGet)

	// API v1 endpoints
	v1 := r.PathPrefix("/api/v1").Subrouter()
	v1.HandleFunc("/yubikey/otp/verify", s.yubikeyOTPVerify).Methods(http.MethodGet, http.MethodGet)

	return http.ListenAndServe(":8080", r)
}
