package rpclient

import (
	"context"
	"net"
	"net/http"
	"net/url"
)

type Client struct {
	client     *http.Client
	socketPath string
	baseURL    *url.URL
}

func New(socketPath string) *Client {
	transport := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
	}

	client := &http.Client{
		Transport: transport,
	}

	return &Client{
		client:     client,
		socketPath: socketPath,
		baseURL: &url.URL{
			Scheme: "unix",
			Host:   "localhost",
		},
	}
}
