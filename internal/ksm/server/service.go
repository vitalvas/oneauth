package server

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/vitalvas/oneauth/internal/ksm/database"
	"github.com/vitalvas/oneauth/internal/yubico"
)

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

func (s *Server) StoreKey(keyID, aesKeyB64, description string) error {
	// Validate key ID format using yubico package
	if err := yubico.ValidateKeyIDFormat(keyID); err != nil {
		return fmt.Errorf("invalid key ID format: %w", err)
	}

	// Decode AES key from base64
	aesKey, err := base64.RawURLEncoding.DecodeString(aesKeyB64)
	if err != nil {
		return fmt.Errorf("invalid base64 encoding: %w", err)
	}
	defer clear(aesKey)

	if len(aesKey) != 16 {
		return fmt.Errorf("AES key must be exactly 16 bytes")
	}

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
