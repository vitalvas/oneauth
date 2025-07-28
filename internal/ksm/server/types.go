package server

import "time"

// DecryptResponse represents the response from OTP decryption
type DecryptResponse struct {
	Status        string    `json:"status"`
	KeyID         string    `json:"key_id,omitempty"`
	Counter       int       `json:"counter,omitempty"`
	TimestampLow  int       `json:"timestamp_low,omitempty"`
	TimestampHigh int       `json:"timestamp_high,omitempty"`
	SessionUse    int       `json:"session_use,omitempty"`
	DecryptedAt   time.Time `json:"decrypted_at,omitempty"`
	ErrorCode     string    `json:"error_code,omitempty"`
	Message       string    `json:"message,omitempty"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status   string         `json:"status"`
	Database DatabaseHealth `json:"database"`
}

// DatabaseHealth represents the database health status
type DatabaseHealth struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}
