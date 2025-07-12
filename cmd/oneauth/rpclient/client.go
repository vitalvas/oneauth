package rpclient

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
)

type Client struct {
	socketPath string
	client     *rpc.Client
}

func New(socketPath string) (*Client, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, err
	}

	client := rpc.NewClientWithCodec(jsonrpc.NewClientCodec(conn))

	return &Client{
		socketPath: socketPath,
		client:     client,
	}, nil
}

func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

func IsClient(socketPath string) bool {
	if info, err := os.Stat(socketPath); err != nil {
		return false
	} else if info.Mode()&os.ModeSocket == 0 {
		return false
	}

	_, err := net.Dial("unix", socketPath)
	return err == nil
}
