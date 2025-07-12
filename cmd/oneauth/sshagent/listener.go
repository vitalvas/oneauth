package sshagent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

func (a *SSHAgent) ListenAndServe(ctx context.Context, socketPath string) error {
	defer func() {
		if _, err := os.Stat(socketPath); err == nil {
			os.Remove(socketPath)
		}
	}()

	a.log.Println("listening ssh-agent on", socketPath)

	// If no listener is set (normal case), create one
	if a.getListener() == nil {
		// Remove any existing file at the socket path
		if _, err := os.Stat(socketPath); err == nil {
			os.Remove(socketPath)
		}

		var err error
		listener, err := net.Listen("unix", socketPath)
		if err != nil {
			return fmt.Errorf("failed to listen: %w", err)
		}

		a.setListener(listener)

		if err := os.Chmod(socketPath, 0600); err != nil {
			return fmt.Errorf("failed to chmod: %w", err)
		}

		defer func() {
			if l := a.getListener(); l != nil {
				l.Close()
			}
		}()
	}

	// Start a goroutine to close the listener when context is cancelled
	go func() {
		<-ctx.Done()
		if l := a.getListener(); l != nil {
			l.Close()
		}
	}()

	for {
		listener := a.getListener()
		if listener == nil {
			return fmt.Errorf("listener is nil")
		}
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil

			default:
			}

			if errors.Is(err, net.ErrClosed) || errors.Is(err, io.ErrClosedPipe) || errors.Is(err, io.EOF) {
				return nil
			}

			if err, ok := err.(Temporary); ok && err.Temporary() {
				a.log.Printf("temporary accept error: %v", err)
				// Use context-aware sleep instead of blocking sleep
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(100 * time.Millisecond):
					continue
				}
			}

			return fmt.Errorf("failed to accept: %w", err)
		}

		go a.handleConn(conn)
	}
}
