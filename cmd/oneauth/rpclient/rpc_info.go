package rpclient

import (
	"encoding/json"
	"net/url"

	"github.com/vitalvas/oneauth/cmd/oneauth/rpcserver"
)

func (c *Client) GetInfo() (*rpcserver.Info, error) {
	reqURL := c.baseURL.ResolveReference(&url.URL{Path: "/info"})

	resp, err := c.client.Get(reqURL.String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var info rpcserver.Info

	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	return &info, nil
}
