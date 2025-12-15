package rpcserver

import (
	"fmt"
	"os"
	"time"

	"github.com/vitalvas/oneauth/internal/buildinfo"
	"github.com/vitalvas/oneauth/internal/yubikey"
)

// AgentService provides JSON-RPC methods for the SSH agent
type AgentService struct {
	server *RPCServer
}

// InfoArgs represents arguments for the Info method
type InfoArgs struct{}

// InfoKey represents YubiKey information
type InfoKey struct {
	Name    string `json:"name"`
	Serial  string `json:"serial"`
	Version string `json:"version"`
}

// InfoReply represents the response from the Info method
type InfoReply struct {
	Pid     int       `json:"pid"`
	Keys    []InfoKey `json:"keys"`
	Version string    `json:"version"`
	Uptime  string    `json:"uptime"`
}

// Info returns information about the running agent
func (s *AgentService) Info(_ *InfoArgs, reply *InfoReply) error {
	reply.Pid = os.Getpid()
	reply.Version = buildinfo.FormattedVersion()
	
	uptime := time.Since(s.server.startTime)
	reply.Uptime = uptime.Truncate(time.Second).String()
	
	cards, err := yubikey.Cards()
	if err != nil {
		reply.Keys = []InfoKey{}
	} else {
		for _, card := range cards {
			reply.Keys = append(reply.Keys, InfoKey{
				Name:    card.String(),
				Serial:  fmt.Sprintf("%d", card.Serial),
				Version: card.Version,
			})
		}
	}
	
	return nil
}
