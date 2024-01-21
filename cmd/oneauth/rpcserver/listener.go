package rpcserver

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

func (s *RPCServer) ListenAndServe(_ context.Context, socketPath string) error {
	defer func() {
		if _, err := os.Stat(socketPath); err == nil {
			os.Remove(socketPath)
		}
	}()

	log.Println("listening rpc on", socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}

	defer listener.Close()

	if err := os.Chmod(socketPath, 0600); err != nil {
		return err
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello world from oneauth agent")
	})

	s.server = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 2 * time.Second,
	}

	if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}
