package netutil

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnixSocketCreds_NonUnixConn(t *testing.T) {
	// TCP connection should return -1/-1
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	defer ln.Close()

	go func() {
		conn, _ := ln.Accept()
		if conn != nil {
			conn.Close()
		}
	}()

	conn, err := net.Dial("tcp", ln.Addr().String())
	assert.NoError(t, err)
	defer conn.Close()

	creds, err := UnixSocketCreds(conn)
	assert.NoError(t, err)
	assert.Equal(t, -1, creds.UID)
	assert.Equal(t, -1, creds.PID)
}
