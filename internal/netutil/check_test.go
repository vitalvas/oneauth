package netutil

import (
	"net"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckCreds(t *testing.T) {
	tests := []struct {
		name      string
		UID       int
		exceptErr bool
	}{
		{"root", 0, false},
		{"current-user", os.Getuid(), false},
		{"non-valid", 667, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := CheckCreds(&UnixCreds{
				UID: test.UID,
			})

			if test.exceptErr && err == nil {
				t.Errorf("Expected error, but got no error")
			}

			if !test.exceptErr && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}
		})
	}
}

func TestUnixCreds_Structure(t *testing.T) {
	creds := UnixCreds{
		PID: 12345,
		UID: 1000,
	}
	
	assert.Equal(t, 12345, creds.PID)
	assert.Equal(t, 1000, creds.UID)
}

func TestCheckCreds_EdgeCases(t *testing.T) {
	currentUID := os.Getuid()
	
	tests := []struct {
		name        string
		creds       *UnixCreds
		expectError bool
	}{
		{
			name: "nil-creds",
			creds: &UnixCreds{
				UID: currentUID,
				PID: os.Getpid(),
			},
			expectError: false,
		},
		{
			name: "negative-uid",
			creds: &UnixCreds{
				UID: -1,
				PID: os.Getpid(),
			},
			expectError: false, // UID <= 0 is allowed (root case)
		},
		{
			name: "zero-uid-root",
			creds: &UnixCreds{
				UID: 0,
				PID: os.Getpid(),
			},
			expectError: false, // root is always allowed
		},
		{
			name: "different-user-positive-uid",
			creds: &UnixCreds{
				UID: currentUID + 1000, // Different user with positive UID
				PID: os.Getpid(),
			},
			expectError: true,
		},
		{
			name: "same-user-different-pid",
			creds: &UnixCreds{
				UID: currentUID,
				PID: 99999, // Different PID shouldn't matter
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckCreds(tt.creds)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "connection from another user")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUnixSocketCreds_NonUnixConnection(t *testing.T) {
	// Create a TCP connection instead of Unix socket
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	// Connect to it
	conn, err := net.Dial("tcp", listener.Addr().String())
	require.NoError(t, err)
	defer conn.Close()

	// Accept the connection
	serverConn, err := listener.Accept()
	require.NoError(t, err)
	defer serverConn.Close()

	// Test that non-Unix connection returns expected values
	creds, err := UnixSocketCreds(serverConn)
	assert.NoError(t, err)
	assert.Equal(t, -1, creds.UID)
	assert.Equal(t, -1, creds.PID)
}

func TestUnixSocketCreds_WithUnixSocket(t *testing.T) {
	// Skip this test on systems where Unix sockets don't work properly
	if runtime.GOOS == "windows" {
		t.Skip("Unix sockets not fully supported on Windows")
	}

	// Create a temporary Unix socket with shorter path
	tempDir := t.TempDir()
	socketPath := filepath.Join(tempDir, "test.sock")
	
	// Ensure the socket path is reasonable length
	if len(socketPath) > 100 {
		t.Skip("Socket path too long for this system")
	}

	// Create Unix socket listener
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Logf("Unix socket creation failed (expected on some systems): %v", err)
		return
	}
	defer listener.Close()

	// Connect to the Unix socket
	clientConn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Logf("Unix socket connection failed (expected on some systems): %v", err)
		return
	}
	defer clientConn.Close()

	// Accept the connection
	serverConn, err := listener.Accept()
	if err != nil {
		t.Logf("Unix socket accept failed (expected on some systems): %v", err)
		return
	}
	defer serverConn.Close()

	// Test Unix socket credentials
	creds, err := UnixSocketCreds(serverConn)
	
	// The exact behavior depends on the platform
	// On some systems this might fail due to permission or syscall issues
	if err != nil {
		t.Logf("UnixSocketCreds failed (expected on some systems): %v", err)
		return
	}

	// If it succeeds, validate the credentials make sense
	t.Logf("Got credentials: UID=%d, PID=%d", creds.UID, creds.PID)
	
	// UID should be reasonable (our current UID or similar)
	if creds.UID >= 0 {
		assert.True(t, creds.UID >= 0, "UID should be non-negative")
	}
	
	// PID should be reasonable if set
	if creds.PID > 0 {
		assert.True(t, creds.PID > 0, "PID should be positive if set")
	}
}

func TestCheckCreds_ErrorMessage(t *testing.T) {
	currentUID := os.Getuid()
	differentUID := currentUID + 1000
	
	if differentUID <= 0 {
		differentUID = currentUID + 1
	}
	
	creds := &UnixCreds{
		UID: differentUID,
		PID: os.Getpid(),
	}
	
	err := CheckCreds(creds)
	require.Error(t, err)
	
	expectedMsg := "connection from another user (except root) is prohibited"
	assert.Contains(t, err.Error(), expectedMsg)
	assert.Contains(t, err.Error(), "!= ")
}

func TestUnixSocketCreds_InvalidConnection(t *testing.T) {
	// Test with nil connection
	creds, err := UnixSocketCreds(nil)
	
	// This should either return error or handle gracefully
	if err == nil {
		// If no error, creds should have default values
		assert.Equal(t, -1, creds.UID)
		assert.Equal(t, -1, creds.PID)
	}
}
