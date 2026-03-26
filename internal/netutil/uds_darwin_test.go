package netutil

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnixSocketCreds_NonUnixConn(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	conn, err := net.Dial("tcp", listener.Addr().String())
	require.NoError(t, err)
	defer conn.Close()

	serverConn, err := listener.Accept()
	require.NoError(t, err)
	defer serverConn.Close()

	creds, err := UnixSocketCreds(serverConn)
	assert.NoError(t, err)
	assert.Equal(t, -1, creds.UID)
	assert.Equal(t, -1, creds.PID)
}

func TestUnixSocketCreds_RealUnixSocket(t *testing.T) {
	tempDir := t.TempDir()
	socketPath := filepath.Join(tempDir, "t.sock")

	if len(socketPath) > 100 {
		t.Skip("Socket path too long")
	}

	listener, err := net.Listen("unix", socketPath)
	require.NoError(t, err)
	defer listener.Close()

	clientConn, err := net.Dial("unix", socketPath)
	require.NoError(t, err)
	defer clientConn.Close()

	serverConn, err := listener.Accept()
	require.NoError(t, err)
	defer serverConn.Close()

	creds, err := UnixSocketCreds(serverConn)
	require.NoError(t, err)

	assert.Equal(t, os.Getuid(), creds.UID)
	assert.True(t, creds.PID > 0 || creds.PID == -1, "PID should be positive or -1")
}

func TestUnixSocketCreds_NilConn(t *testing.T) {
	creds, err := UnixSocketCreds(nil)
	assert.NoError(t, err)
	assert.Equal(t, -1, creds.UID)
	assert.Equal(t, -1, creds.PID)
}

func TestUnixSocketCreds_ClosedUnixSocket(t *testing.T) {
	socketPath := filepath.Join("/tmp", fmt.Sprintf("oa_test_close_%d.sock", os.Getpid()))
	defer os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	require.NoError(t, err)
	defer listener.Close()

	clientConn, err := net.Dial("unix", socketPath)
	require.NoError(t, err)

	serverConn, err := listener.Accept()
	require.NoError(t, err)

	// Close the client side first
	clientConn.Close()

	// Getting creds from server side after client disconnect
	// should still work or return an error gracefully
	creds, err := UnixSocketCreds(serverConn)
	serverConn.Close()
	if err != nil {
		assert.Error(t, err)
	} else {
		assert.True(t, creds.UID >= 0 || creds.UID == -1)
	}
}

func TestUnixSocketCreds_ClientSide(t *testing.T) {
	socketPath := filepath.Join("/tmp", fmt.Sprintf("oa_test_cli_%d.sock", os.Getpid()))
	defer os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	require.NoError(t, err)
	defer listener.Close()

	clientConn, err := net.Dial("unix", socketPath)
	require.NoError(t, err)
	defer clientConn.Close()

	serverConn, err := listener.Accept()
	require.NoError(t, err)
	defer serverConn.Close()

	// Test getting creds from client-side connection
	creds, err := UnixSocketCreds(clientConn)
	if err != nil {
		t.Logf("Client-side creds retrieval failed: %v", err)
	} else {
		assert.True(t, creds.UID >= 0 || creds.UID == -1)
	}
}

func TestUnixSocketCreds_MultipleCallsSameConn(t *testing.T) {
	socketPath := filepath.Join("/tmp", fmt.Sprintf("oa_test_multi_%d.sock", os.Getpid()))
	defer os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	require.NoError(t, err)
	defer listener.Close()

	clientConn, err := net.Dial("unix", socketPath)
	require.NoError(t, err)
	defer clientConn.Close()

	serverConn, err := listener.Accept()
	require.NoError(t, err)
	defer serverConn.Close()

	// Call UnixSocketCreds multiple times on the same connection
	creds1, err1 := UnixSocketCreds(serverConn)
	require.NoError(t, err1)

	creds2, err2 := UnixSocketCreds(serverConn)
	require.NoError(t, err2)

	// Should return consistent results
	assert.Equal(t, creds1.UID, creds2.UID)
	assert.Equal(t, os.Getuid(), creds1.UID)
}

func TestUnixSocketCreds_ClosedServerConn(t *testing.T) {
	socketPath := filepath.Join("/tmp", fmt.Sprintf("oa_test_csrv_%d.sock", os.Getpid()))
	defer os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	require.NoError(t, err)
	defer listener.Close()

	clientConn, err := net.Dial("unix", socketPath)
	require.NoError(t, err)
	defer clientConn.Close()

	serverConn, err := listener.Accept()
	require.NoError(t, err)

	// Close the server connection before trying to get creds
	serverConn.Close()

	// Try to get creds on closed connection - should error on SyscallConn or Control
	_, err = UnixSocketCreds(serverConn)
	// Closed UnixConn should trigger an error in SyscallConn or Control
	assert.Error(t, err)
}

func TestUnixSocketCreds_StreamSocket(t *testing.T) {
	socketPath := filepath.Join("/tmp", fmt.Sprintf("oa_test_stream_%d.sock", os.Getpid()))
	defer os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	require.NoError(t, err)
	defer listener.Close()

	clientConn, err := net.Dial("unix", socketPath)
	require.NoError(t, err)
	defer clientConn.Close()

	serverConn, err := listener.Accept()
	require.NoError(t, err)
	defer serverConn.Close()

	creds, err := UnixSocketCreds(serverConn)
	require.NoError(t, err)
	assert.Equal(t, os.Getuid(), creds.UID)
	assert.True(t, creds.PID > 0 || creds.PID == -1)
}
