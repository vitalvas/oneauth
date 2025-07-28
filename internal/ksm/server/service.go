package server

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/vitalvas/oneauth/internal/ksm/database"
	"github.com/vitalvas/oneauth/internal/yubico"
)

func (s *Server) parseAESKey(aesKeyStr string) ([]byte, error) {
	// Remove any whitespace
	aesKeyStr = strings.TrimSpace(aesKeyStr)

	// Check if it's hex format (32 hex characters for 16 bytes)
	if len(aesKeyStr) == 32 {
		// Validate all characters are hex
		isValidHex := true
		for _, c := range aesKeyStr {
			if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
				isValidHex = false
				break
			}
		}

		if isValidHex {
			// Try to decode as hex
			aesKey, err := hex.DecodeString(aesKeyStr)
			if err != nil {
				return nil, fmt.Errorf("invalid hex encoding: %w", err)
			}
			if len(aesKey) != 16 {
				return nil, fmt.Errorf("AES key must be exactly 16 bytes")
			}
			return aesKey, nil
		}
		// 32 characters but not valid hex
		return nil, fmt.Errorf("invalid hex or base64 encoding: contains invalid hex characters")
	}

	// Try to decode as base64
	aesKey, err := base64.RawURLEncoding.DecodeString(aesKeyStr)
	if err != nil {
		// Try standard base64 encoding as fallback
		aesKey, err = base64.StdEncoding.DecodeString(aesKeyStr)
		if err != nil {
			return nil, fmt.Errorf("invalid hex or base64 encoding: %w", err)
		}
	}

	if len(aesKey) != 16 {
		return nil, fmt.Errorf("AES key must be exactly 16 bytes")
	}

	return aesKey, nil
}

func (s *Server) ValidateOTPFormat(otp string) error {
	_, err := yubico.ValidateOTP(otp)
	if err != nil {
		return fmt.Errorf("invalid OTP: %w", err)
	}
	return nil
}

func (s *Server) DecryptOTP(otp string) (*DecryptResponse, error) {
	// Validate OTP format
	if err := s.ValidateOTPFormat(otp); err != nil {
		return &DecryptResponse{
			Status:    "ERROR",
			ErrorCode: "INVALID_OTP",
			Message:   "Invalid OTP format",
		}, nil
	}

	// Extract and validate key ID
	keyID, err := yubico.ExtractKeyID(otp)
	if err != nil {
		return &DecryptResponse{
			Status:    "ERROR",
			ErrorCode: "INVALID_KEY_ID",
			Message:   "Invalid key ID",
		}, nil
	}

	// Get encrypted AES key from database
	keyRecord, err := s.db.GetKey(keyID)
	if err != nil {
		return &DecryptResponse{
			Status:    "ERROR",
			ErrorCode: "KEY_NOT_FOUND",
			Message:   "Key not found",
		}, nil
	}

	// Decrypt the AES key using row-level encryption
	aesKey, err := s.crypto.DecryptAESKey(keyID, keyRecord.AESKeyEncrypted)
	if err != nil {
		return &DecryptResponse{
			Status:    "ERROR",
			ErrorCode: "DECRYPTION_FAILED",
			Message:   "Decryption failed",
		}, nil
	}
	defer clear(aesKey)

	// Decrypt the OTP using the AES key
	otpData, err := s.crypto.DecryptYubikeyOTP(otp, aesKey)
	if err != nil {
		return &DecryptResponse{
			Status:    "ERROR",
			ErrorCode: "DECRYPTION_FAILED",
			Message:   "Decryption failed",
		}, nil
	}

	// Validate counter for replay protection
	if err := s.db.ValidateCounter(keyID, otpData.Counter, otpData.SessionUse); err != nil {
		return &DecryptResponse{
			Status:    "ERROR",
			ErrorCode: "REPLAY_DETECTED",
			Message:   "Replay attack detected",
		}, nil
	}

	// Store the counter to prevent replay
	counterRecord := &database.YubikeyCounter{
		KeyID:         keyID,
		Counter:       otpData.Counter,
		SessionUse:    otpData.SessionUse,
		TimestampHigh: otpData.TimestampHigh,
		TimestampLow:  otpData.TimestampLow,
		CreatedAt:     time.Now(),
	}

	if err := s.db.StoreCounter(counterRecord); err != nil {
		return &DecryptResponse{
			Status:    "ERROR",
			ErrorCode: "STORAGE_FAILED",
			Message:   "Counter storage failed",
		}, nil
	}

	// Success - create response
	return &DecryptResponse{
		Status:        "OK",
		KeyID:         keyID,
		Counter:       otpData.Counter,
		TimestampLow:  otpData.TimestampLow,
		TimestampHigh: otpData.TimestampHigh,
		SessionUse:    otpData.SessionUse,
		DecryptedAt:   time.Now(),
	}, nil
}

func (s *Server) StoreKey(keyID, aesKeyStr, description string) error {
	// Validate key ID format using yubico package
	if err := yubico.ValidateKeyIDFormat(keyID); err != nil {
		return fmt.Errorf("invalid key ID format: %w", err)
	}

	// Parse AES key from hex or base64 format
	aesKey, err := s.parseAESKey(aesKeyStr)
	if err != nil {
		return err
	}
	defer clear(aesKey)

	// Encrypt AES key using row-level encryption
	encryptedKey, err := s.crypto.EncryptAESKey(keyID, aesKey)
	if err != nil {
		return fmt.Errorf("AES key encryption failed: %w", err)
	}

	// Store in database
	keyRecord := &database.YubikeyKey{
		KeyID:           keyID,
		AESKeyEncrypted: encryptedKey,
		Description:     description,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		UsageCount:      0,
		Active:          true,
	}

	if err := s.db.StoreKey(keyRecord); err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	return nil
}

func (s *Server) ListKeys() ([]*database.YubikeyKey, error) {
	keys, err := s.db.ListKeys()
	if err != nil {
		return nil, err
	}

	return keys, nil
}

func (s *Server) DeleteKey(keyID string) error {
	if err := s.db.DeleteKey(keyID); err != nil {
		return err
	}

	return nil
}
