package rpcserver

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vitalvas/oneauth/cmd/oneauth/sshagent"
)

type RPCServer struct {
	SSHAgent *sshagent.SSHAgent
	server   *http.Server
	log      *logrus.Logger
	mu       sync.RWMutex
}

func New(sshAgent *sshagent.SSHAgent, log *logrus.Logger) *RPCServer {
	return &RPCServer{
		SSHAgent: sshAgent,
		log:      log,
	}
}

func (s *RPCServer) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.mu.RLock()
	server := s.server
	s.mu.RUnlock()

	if server != nil {
		server.Shutdown(ctx)
	}
}

// GetServer returns the HTTP server instance (for testing)
func (s *RPCServer) GetServer() *http.Server {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.server
}
