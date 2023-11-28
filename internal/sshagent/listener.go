package sshagent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
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

	log.Println("listening ssh-agent on", socketPath)

	var err error
	a.agentListener, err = net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	if err := os.Chmod(socketPath, 0600); err != nil {
		return fmt.Errorf("failed to chmod: %w", err)
	}

	defer a.agentListener.Close()

	for {
		conn, err := a.agentListener.Accept()
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
				log.Printf("temporary accept error: %v", err)
				time.Sleep(time.Second)
				continue
			}

			return fmt.Errorf("failed to accept: %w", err)
		}

		go a.handleConn(conn)
	}
}
