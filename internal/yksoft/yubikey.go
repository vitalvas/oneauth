package yksoft

import (
	"crypto/aes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/vitalvas/oneauth/internal/ykshared"
)

// NewSoftwareYubikey creates a new software-based Yubikey implementation
func NewSoftwareYubikey(config *Config) (*SoftwareYubikey, error) {
	if config == nil {
		config = &Config{}
	}

	yk := &SoftwareYubikey{
		Counter:    1, // Start with counter 1
		SessionUse: 0, // Start with session use 0
		Created:    time.Now(),
	}

	// Generate or use provided KeyID
	if config.KeyID != "" {
		if err := ykshared.ValidateKeyIDFormat(config.KeyID); err != nil {
			return nil, fmt.Errorf("invalid KeyID: %w", err)
		}
		yk.KeyID = config.KeyID
	} else {
		keyID, err := ykshared.GenerateKeyID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate KeyID: %w", err)
		}
		yk.KeyID = keyID
	}

	// Generate or use provided PrivateID
	if config.PrivateID != nil {
		if len(config.PrivateID) != 6 {
			return nil, fmt.Errorf("PrivateID must be exactly 6 bytes")
		}
		yk.PrivateID = make([]byte, 6)
		copy(yk.PrivateID, config.PrivateID)
	} else {
		yk.PrivateID = make([]byte, 6)
		if _, err := rand.Read(yk.PrivateID); err != nil {
			return nil, fmt.Errorf("failed to generate PrivateID: %w", err)
		}
	}

	// Generate or use provided AES key
	if config.AESKey != nil {
		if len(config.AESKey) != 16 {
			return nil, fmt.Errorf("AES key must be exactly 16 bytes")
		}
		yk.AESKey = make([]byte, 16)
		copy(yk.AESKey, config.AESKey)
	} else {
		yk.AESKey = make([]byte, 16)
		if _, err := rand.Read(yk.AESKey); err != nil {
			return nil, fmt.Errorf("failed to generate AES key: %w", err)
		}
	}

	// Initialize timestamp and random seed
	yk.Timestamp = uint32(time.Now().Unix() & 0xFFFFFF) // 24-bit timestamp
	randomSeed := make([]byte, 2)
	if _, err := rand.Read(randomSeed); err != nil {
		return nil, fmt.Errorf("failed to generate random seed: %w", err)
	}
	yk.RandomSeed = binary.LittleEndian.Uint16(randomSeed)

	return yk, nil
}

// GenerateOTP generates a new OTP token
func (yk *SoftwareYubikey) GenerateOTP() (*OTPResult, error) {
	// Increment session use counter
	yk.SessionUse++

	// Update timestamp (simulate internal timer)
	yk.Timestamp = uint32(time.Now().Unix() & 0xFFFFFF)

	// Generate random data
	randomBytes := make([]byte, 2)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random data: %w", err)
	}
	randomData := binary.LittleEndian.Uint16(randomBytes)

	// Create OTP data structure
	otpData := otpData{
		Counter:       yk.Counter,
		TimestampLow:  uint16(yk.Timestamp & 0xFFFF),
		TimestampHigh: uint8((yk.Timestamp >> 16) & 0xFF),
		SessionUse:    yk.SessionUse,
		RandomData:    randomData,
	}

	// Copy private ID
	copy(otpData.PrivateID[:], yk.PrivateID)

	// Convert to bytes for CRC calculation
	otpBytes := make([]byte, 14) // 6 + 2 + 2 + 1 + 1 + 2 = 14 bytes before CRC
	copy(otpBytes[0:6], otpData.PrivateID[:])
	binary.LittleEndian.PutUint16(otpBytes[6:8], otpData.Counter)
	binary.LittleEndian.PutUint16(otpBytes[8:10], otpData.TimestampLow)
	otpBytes[10] = otpData.TimestampHigh
	otpBytes[11] = otpData.SessionUse
	binary.LittleEndian.PutUint16(otpBytes[12:14], otpData.RandomData)

	// Calculate CRC
	crc := ykshared.CalculateCRC16(otpBytes)
	otpData.CRC = crc

	// Create full 16-byte data structure with CRC
	fullData := make([]byte, 16)
	copy(fullData[0:14], otpBytes)
	binary.LittleEndian.PutUint16(fullData[14:16], crc)

	// Encrypt with AES
	block, err := aes.NewCipher(yk.AESKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	encrypted := make([]byte, 16)
	block.Encrypt(encrypted, fullData)

	// Create complete OTP: KeyID (12 modhex chars) + encrypted part (32 hex chars as modhex)
	// The OTP format is: 12 modhex chars (KeyID) + 32 modhex chars (encrypted data)
	// Total: 44 modhex characters

	// Convert encrypted data to modhex
	encryptedModhex := ykshared.BytesToModhex(encrypted)

	// Combine KeyID and encrypted data
	otpModhex := yk.KeyID + encryptedModhex

	return &OTPResult{
		OTP:        otpModhex,
		Counter:    yk.Counter,
		SessionUse: yk.SessionUse,
		Timestamp:  yk.Timestamp,
		CRC:        crc,
	}, nil
}

// IncrementCounter simulates a power cycle (increments session counter, resets session use)
func (yk *SoftwareYubikey) IncrementCounter() {
	yk.Counter++
	yk.SessionUse = 0
}

// GetKeyID returns the public key identifier
func (yk *SoftwareYubikey) GetKeyID() string {
	return yk.KeyID
}

// GetAESKey returns a copy of the AES key (for KSM storage)
func (yk *SoftwareYubikey) GetAESKey() []byte {
	key := make([]byte, 16)
	copy(key, yk.AESKey)
	return key
}

// GetPrivateID returns a copy of the private identifier
func (yk *SoftwareYubikey) GetPrivateID() []byte {
	privateID := make([]byte, 6)
	copy(privateID, yk.PrivateID)
	return privateID
}
