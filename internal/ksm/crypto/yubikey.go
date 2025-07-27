package crypto

import (
	"crypto/aes"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/vitalvas/oneauth/internal/yubico"
)

// YubiKey OTP data structure after decryption
type OTPData struct {
	Counter       int
	TimestampLow  int
	TimestampHigh int
	SessionUse    int
	RandomData    int
	CRC           uint16
}

// DecryptYubikeyOTP decrypts a YubiKey OTP using the provided AES key
func (e *Engine) DecryptYubikeyOTP(otp string, aesKey []byte) (*OTPData, error) {
	// Convert modhex OTP to hex using yubico package
	hexOTP, err := yubico.ModhexToHex(otp)
	if err != nil {
		return nil, fmt.Errorf("failed to convert modhex to hex: %w", err)
	}
	if len(hexOTP) != 44 {
		return nil, fmt.Errorf("invalid OTP length after modhex conversion")
	}

	// Convert hex string to bytes
	otpBytes, err := hex.DecodeString(hexOTP)
	if err != nil {
		return nil, fmt.Errorf("failed to convert hex to bytes: %w", err)
	}

	// Extract encrypted part (skip key ID, take last 32 hex chars = 16 bytes)
	if len(otpBytes) != 32 {
		return nil, fmt.Errorf("invalid OTP byte length")
	}

	encryptedPart := otpBytes[16:] // Skip first 16 bytes (key ID part)

	// Decrypt using AES-128 ECB mode
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	if len(encryptedPart) != 16 {
		return nil, fmt.Errorf("encrypted part must be 16 bytes")
	}

	decrypted := make([]byte, 16)
	block.Decrypt(decrypted, encryptedPart)

	// Parse decrypted data
	otpData := &OTPData{
		Counter:       int(binary.LittleEndian.Uint16(decrypted[8:10])),
		TimestampLow:  int(binary.LittleEndian.Uint16(decrypted[10:12])),
		TimestampHigh: int(decrypted[12]),
		SessionUse:    int(decrypted[13]),
		RandomData:    int(binary.LittleEndian.Uint16(decrypted[14:16])),
		CRC:           binary.LittleEndian.Uint16(decrypted[14:16]),
	}

	// Verify CRC
	if !verifyCRC(decrypted[:14], otpData.CRC) {
		return nil, fmt.Errorf("CRC verification failed")
	}

	return otpData, nil
}

// CRC-16 calculation for YubiKey (simplified version)
func verifyCRC(data []byte, expectedCRC uint16) bool {
	const poly = 0x1021
	crc := uint16(0xFFFF)

	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ poly
			} else {
				crc <<= 1
			}
		}
	}

	return crc == expectedCRC
}
