package mock

import (
	"io"
	"net"
)

// Listener is a mock implementation of net.Listener for testing
type Listener struct {
	AcceptCalls int
}

func (m *Listener) Accept() (net.Conn, error) {
	m.AcceptCalls++

	// Return temporary error for first few calls
	if m.AcceptCalls < 3 {
		return nil, &TemporaryError{Message: "temporary error"}
	}

	// Return permanent error to stop the loop
	return nil, &PermanentError{Message: "permanent error"}
}

func (m *Listener) Close() error {
	return nil
}

func (m *Listener) Addr() net.Addr {
	return &Addr{}
}

// TemporaryError is a mock error that implements the Temporary() method
type TemporaryError struct {
	Message string
}

func (e *TemporaryError) Error() string {
	return e.Message
}

func (e *TemporaryError) Temporary() bool {
	return true
}

// PermanentError is a mock error for permanent failures
type PermanentError struct {
	Message string
}

func (e *PermanentError) Error() string {
	return e.Message
}

// Addr is a mock implementation of net.Addr
type Addr struct{}

func (a *Addr) Network() string {
	return "unix"
}

func (a *Addr) String() string {
	return "mock-addr"
}

// ClosedListener is a mock listener that returns net.ErrClosed
type ClosedListener struct{}

func (m *ClosedListener) Accept() (net.Conn, error) {
	return nil, net.ErrClosed
}

func (m *ClosedListener) Close() error {
	return nil
}

func (m *ClosedListener) Addr() net.Addr {
	return &Addr{}
}

// EOFListener is a mock listener that returns io.EOF
type EOFListener struct{}

func (m *EOFListener) Accept() (net.Conn, error) {
	return nil, io.EOF
}

func (m *EOFListener) Close() error {
	return nil
}

func (m *EOFListener) Addr() net.Addr {
	return &Addr{}
}
