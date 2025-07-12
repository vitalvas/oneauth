package rpcserver

import (
	"context"
	"errors"
	"net"
	"net/rpc/jsonrpc"
	"os"
	"strings"
)

func (s *RPCServer) ListenAndServe(ctx context.Context, socketPath string) error {
	defer func() {
		if _, err := os.Stat(socketPath); err == nil {
			os.Remove(socketPath)
		}
	}()

	s.log.Println("listening json-rpc on", socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}

	defer listener.Close()

	if err := os.Chmod(socketPath, 0600); err != nil {
		return err
	}

	s.mu.Lock()
	s.listener = listener
	s.mu.Unlock()

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if isNetworkClosedError(err) || errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}

		go s.rpcServer.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}

func isNetworkClosedError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "use of closed network connection")
}
