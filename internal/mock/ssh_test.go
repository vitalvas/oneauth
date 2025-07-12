package mock

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
)

func TestNewChannel(t *testing.T) {
	t.Run("Creation", func(t *testing.T) {
		channel := NewSSHChannel("session")
		assert.Equal(t, "session", channel.ChannelType())
		assert.False(t, channel.IsAccepted())
		assert.False(t, channel.IsRejected())
	})

	t.Run("WithMethods", func(t *testing.T) {
		conn := NewChannelConn()
		requests := MakeRequestSlice([]*ssh.Request{})
		data := []byte("test data")

		channel := NewSSHChannel("session").
			WithConn(conn).
			WithRequests(requests).
			WithExtraData(data)

		assert.Equal(t, "session", channel.ChannelType())
		assert.Equal(t, data, channel.ExtraData())
	})

	t.Run("Accept", func(t *testing.T) {
		conn := NewChannelConn()
		requests := MakeRequestSlice([]*ssh.Request{})

		channel := NewSSHChannel("session").
			WithConn(conn).
			WithRequests(requests)

		returnedConn, returnedReqs, err := channel.Accept()
		assert.NoError(t, err)
		assert.Equal(t, conn, returnedConn)
		assert.Equal(t, requests, returnedReqs)
		assert.True(t, channel.IsAccepted())
	})

	t.Run("AcceptError", func(t *testing.T) {
		expectedErr := fmt.Errorf("accept failed")
		channel := NewSSHChannel("session").WithAcceptError(expectedErr)

		_, _, err := channel.Accept()
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.False(t, channel.IsAccepted())
	})

	t.Run("Reject", func(t *testing.T) {
		channel := NewSSHChannel("session")

		err := channel.Reject(ssh.UnknownChannelType, "test rejection")
		assert.NoError(t, err)
		assert.True(t, channel.IsRejected())
		assert.Equal(t, ssh.UnknownChannelType, channel.RejectReason())
		assert.Equal(t, "test rejection", channel.RejectMessage())
	})
}

func TestChannelConn(t *testing.T) {
	t.Run("Creation", func(t *testing.T) {
		conn := NewChannelConn()
		assert.NotNil(t, conn)
		assert.False(t, conn.IsClosed())
		assert.Empty(t, conn.GetWrittenData())
	})

	t.Run("Write", func(t *testing.T) {
		conn := NewChannelConn()
		data := []byte("test data")

		n, err := conn.Write(data)
		assert.NoError(t, err)
		assert.Equal(t, len(data), n)
		assert.Equal(t, data, conn.GetWrittenData())
	})

	t.Run("Read", func(t *testing.T) {
		conn := NewChannelConn().WithReadData([]byte("test data"))

		buf := make([]byte, 5)
		n, err := conn.Read(buf)
		assert.NoError(t, err)
		assert.Equal(t, 5, n)
		assert.Equal(t, []byte("test "), buf)
	})

	t.Run("Close", func(t *testing.T) {
		conn := NewChannelConn()

		err := conn.Close()
		assert.NoError(t, err)
		assert.True(t, conn.IsClosed())

		// Writing to closed connection should fail
		_, err = conn.Write([]byte("test"))
		assert.Error(t, err)
	})

	t.Run("SendRequest", func(t *testing.T) {
		conn := NewChannelConn()

		ok, err := conn.SendRequest("test", true, []byte("payload"))
		assert.NoError(t, err)
		assert.True(t, ok)
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("MakeNewChannelSlice", func(t *testing.T) {
		channels := []*NewChannel{
			NewSSHChannel("session"),
			NewSSHChannel("direct-tcpip"),
		}

		ch := MakeNewChannelSlice(channels)

		// Read from channel
		newChan1 := <-ch
		assert.Equal(t, "session", newChan1.ChannelType())

		newChan2 := <-ch
		assert.Equal(t, "direct-tcpip", newChan2.ChannelType())

		// Channel should be closed
		_, ok := <-ch
		assert.False(t, ok)
	})

	t.Run("MakeRequestSlice", func(t *testing.T) {
		requests := []*ssh.Request{
			{Type: "shell", WantReply: true},
			{Type: "exec", WantReply: false},
		}

		ch := MakeRequestSlice(requests)

		// Read from channel
		req1 := <-ch
		assert.Equal(t, "shell", req1.Type)
		assert.True(t, req1.WantReply)

		req2 := <-ch
		assert.Equal(t, "exec", req2.Type)
		assert.False(t, req2.WantReply)

		// Channel should be closed
		_, ok := <-ch
		assert.False(t, ok)
	})
}
