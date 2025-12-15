package rpcserver

import (
	"net/rpc"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vitalvas/oneauth/cmd/oneauth/sshagent"
)

type RPCServer struct {
	SSHAgent  *sshagent.SSHAgent
	rpcServer *rpc.Server
	log       *logrus.Logger
	mu        sync.RWMutex
	listener  interface{ Close() error }
	startTime time.Time
}

func New(sshAgent *sshagent.SSHAgent, log *logrus.Logger) *RPCServer {
	rpcServer := rpc.NewServer()
	s := &RPCServer{
		SSHAgent:  sshAgent,
		rpcServer: rpcServer,
		log:       log,
		startTime: time.Now(),
	}

	rpcServer.Register(&AgentService{server: s})

	return s
}

func (s *RPCServer) Shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener != nil {
		s.listener.Close()
		s.listener = nil
	}
}

func (s *RPCServer) GetRPCServer() *rpc.Server {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rpcServer
}
