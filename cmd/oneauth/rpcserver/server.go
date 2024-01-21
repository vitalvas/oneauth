package rpcserver

import (
	"context"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vitalvas/oneauth/cmd/oneauth/sshagent"
)

type RPCServer struct {
	SSHAgent *sshagent.SSHAgent
	server   *http.Server
	log      *logrus.Logger
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

	if s.server != nil {
		s.server.Shutdown(ctx)
	}
}
