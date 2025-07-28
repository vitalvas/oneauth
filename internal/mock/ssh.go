package mock

import (
	"io"

	"golang.org/x/crypto/ssh"
)

// NewChannel is a mock implementation of ssh.NewChannel for testing
type NewChannel struct {
	channelType   string
	data          []byte
	accepted      bool
	rejected      bool
	rejectReason  ssh.RejectionReason
	rejectMessage []byte
	acceptError   error
	conn          ssh.Channel
	requests      <-chan *ssh.Request
}

// NewSSHChannel creates a new mock SSH channel
func NewSSHChannel(channelType string) *NewChannel {
	return &NewChannel{
		channelType: channelType,
		data:        []byte{},
	}
}

// WithConn sets the connection for the mock channel
func (m *NewChannel) WithConn(conn ssh.Channel) *NewChannel {
	m.conn = conn
	return m
}

// WithRequests sets the request channel for the mock channel
func (m *NewChannel) WithRequests(requests <-chan *ssh.Request) *NewChannel {
	m.requests = requests
	return m
}

// WithAcceptError sets an error to be returned on Accept()
func (m *NewChannel) WithAcceptError(err error) *NewChannel {
	m.acceptError = err
	return m
}

// WithExtraData sets the extra data for the channel
func (m *NewChannel) WithExtraData(data []byte) *NewChannel {
	m.data = data
	return m
}

func (m *NewChannel) Accept() (ssh.Channel, <-chan *ssh.Request, error) {
	if m.acceptError != nil {
		return nil, nil, m.acceptError
	}
	m.accepted = true
	return m.conn, m.requests, nil
}

func (m *NewChannel) Reject(reason ssh.RejectionReason, message string) error {
	m.rejected = true
	m.rejectReason = reason
	m.rejectMessage = []byte(message)
	return nil
}

func (m *NewChannel) ChannelType() string {
	return m.channelType
}

func (m *NewChannel) ExtraData() []byte {
	return m.data
}

// IsAccepted returns whether the channel was accepted
func (m *NewChannel) IsAccepted() bool {
	return m.accepted
}

// IsRejected returns whether the channel was rejected
func (m *NewChannel) IsRejected() bool {
	return m.rejected
}

// RejectReason returns the rejection reason
func (m *NewChannel) RejectReason() ssh.RejectionReason {
	return m.rejectReason
}

// RejectMessage returns the rejection message
func (m *NewChannel) RejectMessage() string {
	return string(m.rejectMessage)
}

// ChannelConn is a mock implementation of ssh.Channel for testing
type ChannelConn struct {
	writeData []byte
	readData  []byte
	readPos   int
	closed    bool
}

// NewChannelConn creates a new mock channel connection
func NewChannelConn() *ChannelConn {
	return &ChannelConn{
		writeData: make([]byte, 0),
	}
}

// WithReadData sets the data to be returned on Read calls
func (m *ChannelConn) WithReadData(data []byte) *ChannelConn {
	m.readData = data
	m.readPos = 0
	return m
}

func (m *ChannelConn) Read(b []byte) (n int, err error) {
	if m.closed {
		return 0, io.EOF
	}
	if m.readPos >= len(m.readData) {
		return 0, io.EOF
	}

	n = copy(b, m.readData[m.readPos:])
	m.readPos += n
	return n, nil
}

func (m *ChannelConn) Write(b []byte) (n int, err error) {
	if m.closed {
		return 0, io.ErrClosedPipe
	}
	m.writeData = append(m.writeData, b...)
	return len(b), nil
}

func (m *ChannelConn) Close() error {
	m.closed = true
	return nil
}

func (m *ChannelConn) CloseWrite() error {
	return nil
}

func (m *ChannelConn) SendRequest(_ string, _ bool, _ []byte) (bool, error) {
	return true, nil
}

func (m *ChannelConn) Stderr() io.ReadWriter {
	return m
}

// GetWrittenData returns all data written to the connection
func (m *ChannelConn) GetWrittenData() []byte {
	return m.writeData
}

// IsClosed returns whether the connection is closed
func (m *ChannelConn) IsClosed() bool {
	return m.closed
}

// Helper functions for creating channels

// MakeNewChannelSlice creates a channel from a slice of NewChannel mocks
func MakeNewChannelSlice(channels []*NewChannel) <-chan ssh.NewChannel {
	ch := make(chan ssh.NewChannel, len(channels))
	for _, channel := range channels {
		ch <- channel
	}
	close(ch)
	return ch
}

// MakeRequestSlice creates a channel from a slice of ssh.Request
func MakeRequestSlice(requests []*ssh.Request) <-chan *ssh.Request {
	ch := make(chan *ssh.Request, len(requests))
	for _, req := range requests {
		ch <- req
	}
	close(ch)
	return ch
}
