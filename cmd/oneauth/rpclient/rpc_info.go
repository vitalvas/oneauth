package rpclient

import (
	"github.com/vitalvas/oneauth/cmd/oneauth/rpcserver"
)

func (c *Client) GetInfo() (*rpcserver.InfoReply, error) {
	args := &rpcserver.InfoArgs{}
	reply := &rpcserver.InfoReply{}

	err := c.client.Call("AgentService.Info", args, reply)
	if err != nil {
		return nil, err
	}

	return reply, nil
}
