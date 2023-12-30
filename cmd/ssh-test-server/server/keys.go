package server

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"golang.org/x/crypto/ssh"
)

type YubikeyOTPVerifyResponse struct {
	OTP    string `json:"otp"`
	Serial int    `json:"serial"`
}

func (s *Server) sshPasswordCallback(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	relativeURL, err := url.Parse("/api/v1/yubikey/otp/verify")
	if err != nil {
		log.Println("failed to parse relative URL:", err)
		return nil, err
	}

	validate := s.serverURL.ResolveReference(relativeURL)

	host, _, _ := net.SplitHostPort(conn.RemoteAddr().String())

	sessionID := hex.EncodeToString(conn.SessionID())

	args := url.Values{
		"username":       []string{conn.User()},
		"otp":            []string{string(password)},
		"client-address": []string{host},
		"session-id":     []string{sessionID},
	}

	validate.RawQuery = args.Encode()

	resp, err := http.Get(validate.String())
	if err != nil {
		log.Println("failed to validate YubiKey OTP:", err)
		return nil, fmt.Errorf("failed to validate YubiKey OTP: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			bodyBytes, err2 := io.ReadAll(resp.Body)
			if err != nil {
				log.Println("failed to read YubiKey OTP validation response:", err2)
			} else {
				log.Println("failed to validate YubiKey OTP:", string(bodyBytes))
			}
		} else {
			log.Println("failed to validate YubiKey OTP:", resp.Status)
		}
		return nil, errors.New("failed to validate YubiKey OTP")
	}

	var verifyResponse YubikeyOTPVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&verifyResponse); err != nil {
		log.Println("failed to decode YubiKey OTP validation response:", err)
		return nil, fmt.Errorf("failed to decode YubiKey OTP validation response: %w", err)
	}

	return &ssh.Permissions{
		Extensions: map[string]string{
			"auth-type":      "yubikey-otp",
			"yubikey-serial": strconv.Itoa(verifyResponse.Serial),
		},
	}, nil
}

func (s *Server) sshPublicKeyCallback(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	checker := ssh.CertChecker{
		IsUserAuthority: func(auth ssh.PublicKey) bool {
			return false
		},
		IsRevoked: func(cert *ssh.Certificate) bool {
			return false
		},
		UserKeyFallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			return nil, errors.New("only SSH certificates are supported")
		},
	}

	return checker.Authenticate(conn, key)
}
