package mock

import (
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListener(t *testing.T) {
	t.Run("Accept", func(t *testing.T) {
		listener := &Listener{}

		// First few calls should return temporary error
		for i := 0; i < 2; i++ {
			conn, err := listener.Accept()
			assert.Nil(t, conn)
			assert.Error(t, err)

			tempErr, ok := err.(*TemporaryError)
			assert.True(t, ok)
			assert.True(t, tempErr.Temporary())
		}

		// Third call should return permanent error
		conn, err := listener.Accept()
		assert.Nil(t, conn)
		assert.Error(t, err)
		assert.IsType(t, &PermanentError{}, err)

		assert.Equal(t, 3, listener.AcceptCalls)
	})

	t.Run("Close", func(t *testing.T) {
		listener := &Listener{}
		assert.NoError(t, listener.Close())
	})

	t.Run("Addr", func(t *testing.T) {
		listener := &Listener{}
		addr := listener.Addr()
		assert.NotNil(t, addr)
		assert.Equal(t, "unix", addr.Network())
		assert.Equal(t, "mock-addr", addr.String())
	})
}

func TestClosedListener(t *testing.T) {
	t.Run("Accept", func(t *testing.T) {
		listener := &ClosedListener{}
		conn, err := listener.Accept()
		assert.Nil(t, conn)
		assert.Equal(t, net.ErrClosed, err)
	})

	t.Run("Close", func(t *testing.T) {
		listener := &ClosedListener{}
		assert.NoError(t, listener.Close())
	})

	t.Run("Addr", func(t *testing.T) {
		listener := &ClosedListener{}
		addr := listener.Addr()
		assert.NotNil(t, addr)
		assert.Equal(t, "unix", addr.Network())
		assert.Equal(t, "mock-addr", addr.String())
	})
}

func TestEOFListener(t *testing.T) {
	t.Run("Accept", func(t *testing.T) {
		listener := &EOFListener{}
		conn, err := listener.Accept()
		assert.Nil(t, conn)
		assert.Equal(t, io.EOF, err)
	})

	t.Run("Close", func(t *testing.T) {
		listener := &EOFListener{}
		assert.NoError(t, listener.Close())
	})

	t.Run("Addr", func(t *testing.T) {
		listener := &EOFListener{}
		addr := listener.Addr()
		assert.NotNil(t, addr)
		assert.Equal(t, "unix", addr.Network())
		assert.Equal(t, "mock-addr", addr.String())
	})
}

func TestTemporaryError(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		err := &TemporaryError{Message: "test error"}
		assert.Equal(t, "test error", err.Error())
		assert.True(t, err.Temporary())
	})
}

func TestPermanentError(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		err := &PermanentError{Message: "test error"}
		assert.Equal(t, "test error", err.Error())
	})
}

func TestAddr(t *testing.T) {
	t.Run("Network", func(t *testing.T) {
		addr := &Addr{}
		assert.Equal(t, "unix", addr.Network())
		assert.Equal(t, "mock-addr", addr.String())
	})
}
