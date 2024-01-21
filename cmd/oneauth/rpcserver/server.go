package rpcserver

import (
	"context"
	"net/http"

	"github.com/vitalvas/oneauth/cmd/oneauth/sshagent"
)

type RPCServer struct {
	SSHAgent *sshagent.SSHAgent
	server   *http.Server
}

func New(sshAgent *sshagent.SSHAgent) *RPCServer {
	return &RPCServer{
		SSHAgent: sshAgent,
	}
}

func (s *RPCServer) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5)
	defer cancel()

	if s.server != nil {
		s.server.Shutdown(ctx)
	}
}
