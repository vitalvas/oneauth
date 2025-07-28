package yksoft

import (
	"time"
)

// SoftwareYubikey represents a software-based Yubikey implementation
type SoftwareYubikey struct {
	KeyID      string    // 12-character modhex key identifier (public ID)
	PrivateID  []byte    // 6-byte private identifier
	AESKey     []byte    // 16-byte AES key for encryption
	Counter    uint16    // Session counter (increments on power-up)
	SessionUse uint8     // Usage counter within current session
	Timestamp  uint32    // Internal timestamp (24-bit, low 16 + high 8)
	RandomSeed uint16    // Random seed for generating random data
	Created    time.Time // When this software key was created
}

// OTPResult represents the result of OTP generation
type OTPResult struct {
	OTP        string // The generated 44-character modhex OTP
	Counter    uint16 // Session counter used
	SessionUse uint8  // Session use counter used
	Timestamp  uint32 // Timestamp used
	CRC        uint16 // CRC checksum
}

// Configuration for creating software Yubikeys
type Config struct {
	// If not provided, will be generated randomly
	KeyID     string // 12-character modhex key identifier
	PrivateID []byte // 6-byte private identifier (if nil, will be generated)
	AESKey    []byte // 16-byte AES key (if nil, will be generated)
}

// Internal OTP data structure (matches ksm/crypto package)
type otpData struct {
	PrivateID     [6]byte // Private identifier
	Counter       uint16  // Session counter
	TimestampLow  uint16  // Timestamp low 16 bits
	TimestampHigh uint8   // Timestamp high 8 bits
	SessionUse    uint8   // Session use counter
	RandomData    uint16  // Random data
	CRC           uint16  // CRC checksum
}
